package sqlstore

import (
	"bytes"
	"testing"

	"github.com/cybersamx/teapot/common"
	"github.com/cybersamx/teapot/model"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func newSQLiteStore(t *testing.T) *SQLStore {
	cfg := model.Config{
		Store: model.StoreConfig{
			Driver: "sqlite3",
			DSN:    ":memory:",
		},
	}

	var logger *logrus.Logger
	buf := bytes.Buffer{}
	logger = common.NewTestLogger(&buf, func(code int) {
		require.Equal(t, 0, code)
		text := buf.String()
		require.Contains(t, text, cfg.Store.Driver)
	})

	sqlstore := NewSQLiteStore(logger)
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

func newBareSQLiteStore(t *testing.T) *SQLStore {
	sqlstore := newSQLiteStore(t)

	ctx := newTestContext(t)
	err := sqlstore.Clear(ctx)
	require.NoError(t, err)

	return sqlstore
}
