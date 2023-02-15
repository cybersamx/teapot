package sqlstore

import (
	"context"
	"embed"
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	dbOpTimeout = 5 * time.Second
)

var (
	newBareStoreFns = map[string]newTestStoreFunc{
		"mysql":    newBareMySQLStore,
		"postgres": newBarePostgresStore,
		"sqlite":   newBareSQLiteStore,
	}
)

// newTestStoreFunc represents the function signature of a test store constructor.
type newTestStoreFunc func(t *testing.T) *SQLStore

// closeFunc is a function return by a sqlstore test constructor to allow the caller to close the store.
type closeFunc func()

func newTestContext(t *testing.T) context.Context {
	t.Helper()
	return context.TODO()
}

func newTestTimeoutContext(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), dbOpTimeout)
}

// migrateDown run the down migration scripts so that the database may start from a clean slate.
func migrateDown(t *testing.T, sqlstore *SQLStore) {
	t.Helper()

	ctx := newTestContext(t)

	mig := NewMigrator(sqlstore)
	err := mig.down(ctx, sqlstore.cfg.Store.Driver)
	require.NoError(t, err)
}

// initStoreWithMigrations initializes the connected sqlstore by running 0 to many migration scripts. The scripts are
// the names of the embedded migration scripts minus the `_up.sql` or `_down.sql` suffixes. If there's no script passed,
// then the sqlstore is responsible for setting up and tearing down tables.
func initStoreWithMigrations(t *testing.T, sqlstore *SQLStore, scripts ...string) {
	t.Helper()

	if len(scripts) == 0 {
		return
	}

	var upFS, downFS *embed.FS
	var dirname string

	switch sqlstore.db.DriverName() {
	case driverMySQL:
		upFS = &mysqlUpScripts
		downFS = &mysqlDownScripts
		dirname = "mysql"
	case driverPostgres:
		upFS = &postgresUpScripts
		downFS = &postgresDownScripts
		dirname = "postgres"
	case driverSQLite:
		upFS = &sqliteUpScripts
		downFS = &sqliteDownScripts
		dirname = "sqlite"
	default:
		t.Fatalf("Driver %s isn't supported", sqlstore.db.DriverName())
	}

	mig := NewMigrator(sqlstore)
	t.Cleanup(func() {
		if sqlstore != nil {
			for _, migFile := range scripts {
				migFile = fmt.Sprintf("%s/%s/%s_down.sql", migrationsDir, dirname, migFile)
				err := mig.execFile(newTestContext(t), downFS, migFile)
				require.NoError(t, err)
			}
		}
	})

	for _, migFile := range scripts {
		migFile = fmt.Sprintf("%s/%s/%s_up.sql", migrationsDir, dirname, migFile)
		err := mig.execFile(newTestContext(t), upFS, migFile)
		require.NoError(t, err)
	}
}

func testSQLStoreClear(t *testing.T, newTestStoreFn newTestStoreFunc) {
	t.Helper()

	sqlstore := newTestStoreFn(t)
	t.Cleanup(func() {
		if sqlstore != nil {
			ctx := newTestContext(t)
			_, err := sqlstore.db.ExecContext(ctx, "DROP TABLE clear")
			require.NoError(t, err)
		}
	})

	ctx := newTestContext(t)
	_, err := sqlstore.db.ExecContext(ctx, "CREATE TABLE clear (id VARCHAR(8), username VARCHAR(16))")
	require.NoError(t, err)

	terr := sqlstore.wrapper.transact(ctx, sqlstore.db, func(tx *sqlx.Tx) error {
		result, eerr := tx.ExecContext(ctx,
			"INSERT INTO clear (id, username) VALUES ('123', 'dave')",
		)
		assert.NoError(t, eerr)

		n, eerr := result.RowsAffected()
		assert.NoError(t, eerr)
		assert.Greater(t, n, int64(0))

		return nil
	})
	require.NoError(t, terr)

	err = sqlstore.Clear(ctx)
	assert.NoError(t, err)

	var count int
	err = sqlstore.db.Get(&count, `SELECT COUNT(*) FROM clear`)
	assert.NoError(t, err)
	assert.Zero(t, count)
}

func TestSQLStore_Clear(t *testing.T) {
	for platform, fn := range newBareStoreFns {
		t.Run("With "+platform, func(t *testing.T) {
			testSQLStoreClear(t, fn)
		})
	}
}
