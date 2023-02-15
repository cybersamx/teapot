package sqlstore

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/cybersamx/teapot/model"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

const (
	mysqlErrCodeUniqueViolation = 1062
)

var _ wrapper = (*mysqlWrapper)(nil)

type mysqlWrapper struct{}

func (mw *mysqlWrapper) stmtBuilder() squirrel.StatementBuilderType {
	return squirrel.StatementBuilder
}

func (mw *mysqlWrapper) preConnect(db *sqlx.DB, cfg *model.Config) {}

func (mw *mysqlWrapper) postConnect(db *sqlx.DB, cfg *model.Config) {
	db.SetMaxOpenConns(cfg.Store.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Store.MaxIdleConns)

	// When ConnMaxLifetime isn't set, the app may get sporadic error with bad connection and unexpected eof.
	// This is resolved with setting the right max lifetime.
	// See https://github.blog/2020-05-20-three-bugs-in-the-go-mysql-driver/
	db.SetConnMaxLifetime(cfg.Store.ConnMaxLifetime)
}

func (mw *mysqlWrapper) transact(ctx context.Context, db *sqlx.DB, fn txHandlerFunc) (rerr error) {
	return transact(ctx, db, fn)
}

func (mw *mysqlWrapper) isTableExists(ctx context.Context, db *sqlx.DB, table string) bool {
	var count int

	err := db.Get(&count,
		`SELECT COUNT(*) FROM information_schema.tables WHERE table_name = ? AND table_schema = DATABASE()`,
		table,
	)
	if err != nil {
		return false
	}

	return count > 0
}

func (mw *mysqlWrapper) tableNames(ctx context.Context, db *sqlx.DB) ([]string, error) {
	var tables []string
	err := db.Select(&tables, `SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE()`)

	return tables, err
}

func (mw *mysqlWrapper) clearTable(ctx context.Context, db *sqlx.DB, table string) error {
	// Simple validation.
	if !isValidTableName(table) {
		return fmt.Errorf("(*mysqlWrapper).clearTable - sql injection or invalid table=%s; %w", table, ErrInvalidTableName)
	}

	return transact(ctx, db, func(tx *sqlx.Tx) error {
		_, err := tx.ExecContext(ctx, "TRUNCATE "+table)
		return err
	})
}

func (mw *mysqlWrapper) isDuplicateErr(err error) bool {
	pqErr, ok := err.(*mysql.MySQLError)
	if !ok {
		return false
	}

	return pqErr.Number == mysqlErrCodeUniqueViolation
}

func newMySQLWrapper() *mysqlWrapper {
	return &mysqlWrapper{}
}

// NewMySQLStore creates a mysql datastore.
func NewMySQLStore(logger *logrus.Logger) *SQLStore {
	return &SQLStore{
		logger:  logger,
		wrapper: newMySQLWrapper(),
	}
}
