package httpx

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cybersamx/teapot/common"
	"github.com/cybersamx/teapot/model"
	"github.com/cybersamx/teapot/store/sqlstore"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type configOverride func(cfg *model.Config)

func newTestContext(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()

	return context.WithCancel(context.Background())
}

func newTestServer(t *testing.T, ctx context.Context, opts ...configOverride) *Server {
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

	// Server.
	server := New(ss, logger, &cfg)

	t.Cleanup(func() {
		err = ss.Close()
		require.NoError(t, err)
		err = server.httpd.Close()
		require.NoError(t, err)
		logger.Exit(0)
	})

	return server
}

func TestUseTracing(t *testing.T) {
	ctx, cancel := newTestContext(t)
	defer cancel()

	server := newTestServer(t, ctx)

	server.initRequestID()

	// We can call any endpoint to test tracing.
	record := httptest.NewRecorder()
	server.router.ServeHTTP(record, httptest.NewRequest(http.MethodGet, "/health", nil))
	requestIDs := record.Header().Values(HeaderXRequestID)
	assert.Len(t, requestIDs, 1)
	assert.NotEmpty(t, requestIDs[0])
}
