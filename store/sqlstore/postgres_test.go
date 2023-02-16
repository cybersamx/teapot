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

func newPostgresStore(t *testing.T) *SQLStore {
	const envDNSName = "AX_TEST_POSTGRES_DSN"
	dsn := os.Getenv(envDNSName)
	if dsn == "" {
		t.Skipf("Skipping TestPostgres as env %s isn't set - this is ok", envDNSName)
	}

	cfg := model.Config{
		Store: model.StoreConfig{
			Driver: "pgx",
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

	sqlstore := NewPostgresStore(logger)
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

func newBarePostgresStore(t *testing.T) *SQLStore {
	sqlstore := newPostgresStore(t)

	ctx := newTestContext(t)
	err := sqlstore.Clear(ctx)
	require.NoError(t, err)

	return sqlstore
}

func newAuditPostgresStore(t *testing.T) *SQLStore {
	sqlstore := newPostgresStore(t)

	initStoreWithMigrations(t, sqlstore, "0001_create_audits")
	sqlstore.audits = newAuditSQLStore(sqlstore)

	ctx := newTestContext(t)
	err := sqlstore.Clear(ctx)
	require.NoError(t, err)
	err = sqlstore.audits.Clear(ctx)
	require.NoError(t, err)

	return sqlstore
}
