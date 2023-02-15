package sqlstore

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/cybersamx/teapot/model"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

const (
	pqErrCodeUniqueViolation = "23505"
)

var _ wrapper = (*postgresWrapper)(nil)

type postgresWrapper struct{}

func (pw *postgresWrapper) stmtBuilder() squirrel.StatementBuilderType {
	return squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
}

func (pw *postgresWrapper) preConnect(db *sqlx.DB, cfg *model.Config) {}

func (pw *postgresWrapper) postConnect(db *sqlx.DB, cfg *model.Config) {
	db.SetMaxOpenConns(cfg.Store.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Store.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Store.ConnMaxLifetime)
}

func (pw *postgresWrapper) isTableExists(ctx context.Context, db *sqlx.DB, table string) bool {
	var count int

	err := db.Get(&count,
		`SELECT COUNT(*) FROM information_schema.tables WHERE table_name = $1 AND table_schema = 'public' AND table_type = 'BASE TABLE'`,
		table,
	)
	if err != nil {
		return false
	}

	return count > 0
}

func (pw *postgresWrapper) transact(ctx context.Context, db *sqlx.DB, fn txHandlerFunc) (rerr error) {
	return transact(ctx, db, fn)
}

func (pw *postgresWrapper) tableNames(ctx context.Context, db *sqlx.DB) ([]string, error) {
	var tables []string
	err := db.Select(&tables, `SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE'`)

	return tables, err
}

func (pw *postgresWrapper) clearTable(ctx context.Context, db *sqlx.DB, table string) error {
	// Simple validation.
	if !isValidTableName(table) {
		return fmt.Errorf("(*postgresWrapper).clearTable - sql injection or invalid table=%s; %w", table, ErrInvalidTableName)
	}

	return transact(ctx, db, func(tx *sqlx.Tx) error {
		_, err := tx.ExecContext(ctx, "TRUNCATE "+table)
		return err
	})
}

func (pw *postgresWrapper) isDuplicateErr(err error) bool {
	pqErr, ok := err.(*pgconn.PgError)
	if !ok {
		return false
	}

	return pqErr.Code == pqErrCodeUniqueViolation
}

func newPostgresWrapper() *postgresWrapper {
	return &postgresWrapper{}
}

// NewPostgresStore creates a postgres datastore.
func NewPostgresStore(logger *logrus.Logger) *SQLStore {
	return &SQLStore{
		logger:  logger,
		wrapper: newPostgresWrapper(),
	}
}
