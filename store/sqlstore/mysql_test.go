package sqlstore

import (
	"bytes"
	"os"
	"testing"

	"github.com/cybersamx/teapot/common"
	"github.com/cybersamx/teapot/model"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func newMySQLStore(t *testing.T) *SQLStore {
	const envDNSName = "AX_TEST_MYSQL_DSN"
	dsn := os.Getenv(envDNSName)
	if dsn == "" {
		t.Skipf("Skipping TestMySQL as env %s isn't set - this is ok", envDNSName)
	}

	cfg := model.Config{
		Store: model.StoreConfig{
			Driver: "mysql",
			DSN:    dsn,
		},
	}

	var logger *logrus.Logger
	buf := bytes.Buffer{}
	logger = common.NewTestLogger(&buf, func(code int) {
		require.Equal(t, 0, code)
		text := buf.String()
		require.Contains(t, text, cfg.Store.Driver)
	})

	sqlstore := NewMySQLStore(logger)
	t.Cleanup(func() {
		if sqlstore != nil {
			err := sqlstore.Close()
			require.NoError(t, err)

			sqlstore.logger.Exit(0)
		}
	})

	ctx := newTestContext(t)
	err := sqlstore.Connect(ctx, &cfg)
	require.NoError(t, err)

	return sqlstore
}

func newBareMySQLStore(t *testing.T) *SQLStore {
	sqlstore := newMySQLStore(t)

	ctx := newTestContext(t)
	err := sqlstore.Clear(ctx)
	require.NoError(t, err)

	return sqlstore
}
