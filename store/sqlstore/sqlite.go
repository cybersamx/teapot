package sqlstore

import (
	"context"
	"fmt"
	"sync"

	"github.com/Masterminds/squirrel"
	"github.com/cybersamx/teapot/model"
	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

var _ wrapper = (*sqliteWrapper)(nil)

type sqliteWrapper struct {
	mu sync.RWMutex
}

func (sw *sqliteWrapper) stmtBuilder() squirrel.StatementBuilderType {
	return squirrel.StatementBuilder
}

func (sw *sqliteWrapper) preConnect(db *sqlx.DB, cfg *model.Config) {
	// Implement file operations.
}

func (sw *sqliteWrapper) postConnect(db *sqlx.DB, cfg *model.Config) {
	// With sqlite3, these are always set to 1, overriding the values from the config, as the database
	// doesn't support concurrent mutable operations.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
}

func (sw *sqliteWrapper) transact(ctx context.Context, db *sqlx.DB, fn txHandlerFunc) error {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	return transact(ctx, db, fn)
}

func (sw *sqliteWrapper) isTableExists(ctx context.Context, db *sqlx.DB, table string) bool {
	var count int

	err := db.Get(&count, `SELECT COUNT(*) FROM sqlite_master WHERE name = ? AND type = 'table'`, table)
	if err != nil {
		return false
	}

	return count > 0
}

func (sw *sqliteWrapper) tableNames(ctx context.Context, db *sqlx.DB) ([]string, error) {
	var tables []string
	err := db.Select(&tables, `SELECT name FROM sqlite_master WHERE type = 'table' AND type NOT LIKE 'sqlite_%'`)

	return tables, err
}

func (sw *sqliteWrapper) clearTable(ctx context.Context, db *sqlx.DB, table string) error {
	// Simple validation.
	if !isValidTableName(table) {
		return fmt.Errorf("(*sqliteWrapper).clearTable - sql injection or invalid table=%s; %w", table, ErrInvalidTableName)
	}

	return transact(ctx, db, func(tx *sqlx.Tx) error {
		_, err := tx.ExecContext(ctx, "DELETE FROM "+table)
		return err
	})
}

func (sw *sqliteWrapper) isDuplicateErr(err error) bool {
	sqliteErr, ok := err.(sqlite3.Error)
	if !ok {
		return false
	}

	return sqliteErr.ExtendedCode == sqlite3.ErrConstraintPrimaryKey ||
		sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique
}

func newSQLiteWrapper() *sqliteWrapper {
	return &sqliteWrapper{}
}

// NewSQLiteStore creates a sqlite3 datastore based on the configuration.
func NewSQLiteStore(logger *logrus.Logger) *SQLStore {
	return &SQLStore{
		logger:  logger,
		wrapper: newSQLiteWrapper(),
	}
}
