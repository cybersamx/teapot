package sqlstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	mathrand "math/rand"
	"regexp"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type txHandlerFunc func(tx *sqlx.Tx) error

var (
	ErrSQLTx    = errors.New("failed to start a sql transaction")
	ErrSQLBuild = errors.New("failed to build sql query")
)

func isDriverSupported(driverName string) bool {
	return driverName == driverPostgres ||
		driverName == driverMySQL ||
		driverName == driverSQLite
}

func transact(ctx context.Context, db *sqlx.DB, fn txHandlerFunc) (rerr error) {
	if fn == nil {
		return fmt.Errorf("transact - nil txHandlerFunc; %w", ErrSQLTx)
	}

	var err error

	tx, err := db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable, // Strict for now.
		ReadOnly:  false,
	})
	if err != nil {
		return fmt.Errorf("transact, begin tx - rootErr=%v; %w", err, ErrSQLTx)
	}
	defer func() {
		if err == nil {
			return
		}

		if err := tx.Rollback(); err != nil {
			rerr = fmt.Errorf("transact, rollback - rootErr=%v; %w", err, ErrSQLTx)
			return
		}
	}()

	if err = fn(tx); err != nil {
		return fmt.Errorf("transact, exec - %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("transact, commit - rootErr=%v; %w", err, ErrSQLTx)
	}

	return nil
}

func isValidTableName(table string) bool {
	pattern := `^[\w.]+$`
	ok, err := regexp.MatchString(pattern, table)
	if err != nil || !ok {
		return false
	}

	return true
}

func hasRecord(ctx context.Context, db *sqlx.DB, stmt squirrel.SelectBuilder) (bool, error) {
	query, args, err := stmt.ToSql()
	if err != nil {
		return false, fmt.Errorf("hasRecord - rootErr=%v; %w", err, ErrSQLBuild)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return false, fmt.Errorf("hasReord - rootErr=%v; %w", err, ErrSQLExecute)
	}
	defer rows.Close()

	return rows.Next(), nil
}

func randIntRange(min, max int) int {
	return mathrand.Intn(max-min) + min
}
