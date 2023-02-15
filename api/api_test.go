package api

import (
	"bytes"
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cybersamx/teapot/common"
	"github.com/cybersamx/teapot/httpx"
	"github.com/cybersamx/teapot/model"
	"github.com/cybersamx/teapot/store/sqlstore"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAPIPath = "/api/v1"
)

type configOverride func(cfg *model.Config)
type beforeFunc func(t *testing.T, tc *testCaseRoute)
type afterFunc func(t *testing.T, w *httptest.ResponseRecorder, tc *testCaseRoute)

type testCaseRoute struct {
	description string
	method      string
	header      map[string]string
	path        string
	body        string
	wantCode    int
	wantErrMsg  string
	before      beforeFunc
	after       afterFunc
}

func newTestContext(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()

	return context.WithCancel(context.Background())
}

func newTestServer(t *testing.T, ctx context.Context, opts ...configOverride) *httpx.Server {
	// Logger.
	var logger *logrus.Logger
	buf := bytes.Buffer{}
	logger = common.NewTestLogger(&buf, func(code int) {
		require.Equal(t, 0, code)
	})

	// Config.
	cfg := model.Config{
		Mode: "production",
		HTTP: model.HTTPConfig{
			Address: ":7000",
		},
		Store: model.StoreConfig{
			Driver: "sqlite3",
			DSN:    ":memory:",
		},
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	// Store.
	ss := sqlstore.NewSQLiteStore(logger)
	err := ss.Connect(ctx, &cfg)
	require.NoError(t, err)
	err = ss.InitDB(ctx)
	require.NoError(t, err)

	// Server.
	srv := httpx.New(ss, logger, &cfg)

	t.Cleanup(func() {
		err = ss.Close()
		require.NoError(t, err)
		err = srv.HTTPServer().Close()
		require.NoError(t, err)
		logger.Exit(0)
	})

	return srv
}

func runTestCaseRoute(t *testing.T, srv *httpx.Server, tc *testCaseRoute) {
	t.Run(tc.description, func(t *testing.T) {
		if tc.before != nil {
			tc.before(t, tc)
		}

		record := httptest.NewRecorder()
		reader := strings.NewReader(tc.body)
		srv.Router().ServeHTTP(record, httptest.NewRequest(tc.method, tc.path, reader))
		assert.Equal(t, tc.wantCode, record.Code)

		if tc.wantErrMsg != "" {
			er, err := common.ParseJSON[ErrorResponse](record.Body)
			require.NoError(t, err)

			assert.Equal(t, tc.wantErrMsg, er.Message)
			assert.Equal(t, tc.wantCode, er.Code)
		}

		if tc.after != nil {
			tc.after(t, record, tc)
		}
	})
}
