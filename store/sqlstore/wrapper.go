package sqlstore

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/cybersamx/teapot/model"
	"github.com/jmoiron/sqlx"
)

// wrapper contains information about how we handle a specific sql database.
// CREDIT: Inspired by Dex's flavor struct, https://github.com/dexidp/dex/blob/master/storage/storage.go.
// Licensed under Apache 2.0.
type wrapper interface {
	stmtBuilder() squirrel.StatementBuilderType
	preConnect(db *sqlx.DB, cfg *model.Config)
	postConnect(db *sqlx.DB, cfg *model.Config)
	transact(ctx context.Context, db *sqlx.DB, fn txHandlerFunc) error
	isTableExists(ctx context.Context, db *sqlx.DB, tableName string) bool
	tableNames(ctx context.Context, db *sqlx.DB) ([]string, error)
	clearTable(ctx context.Context, db *sqlx.DB, table string) error

	isDuplicateErr(err error) bool // Check for error that caused by duplicate key insert.
}
