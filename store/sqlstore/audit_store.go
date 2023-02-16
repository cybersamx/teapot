package sqlstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/cybersamx/teapot/model"
	"github.com/cybersamx/teapot/store"
	"github.com/jmoiron/sqlx"
)

type AuditSQLStore struct {
	parent *SQLStore
}

func newAuditSQLStore(parent *SQLStore) *AuditSQLStore {
	return &AuditSQLStore{
		parent: parent,
	}
}

// wrapper is a convenient method getting the associated wrapper for operations to the underlying database.
func (as *AuditSQLStore) wrapper() wrapper {
	return as.parent.wrapper
}

func (as *AuditSQLStore) Clear(ctx context.Context) error {
	return as.wrapper().clearTable(ctx, as.parent.db, "audits")
}

func (as *AuditSQLStore) Get(ctx context.Context, requestID string) (*model.Audit, error) {
	builder := squirrel.StatementBuilder.
		Select(
			"request_id", "created_at", "client_agent",
			"client_address", "status_code", "error", "event",
		).
		From("audits").
		Where("request_id = ?", requestID).
		OrderBy("created_at ASC")

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("(*AuditSQLStore).Get - id=%s; root_err=%v; %w",
			requestID, err, ErrSQLBuild)
	}

	var audit model.Audit
	err = as.parent.db.GetContext(ctx, &audit, query, args...)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, fmt.Errorf("(*AuditSQLStore).Get - id=%s; %w",
			requestID, store.ErrNoRows)
	case err != nil:
		return nil, fmt.Errorf("(*AuditSQLStore).Get - id=%s; root_err=%v: %w",
			requestID, err, store.ErrInternal)
	}

	return &audit, nil
}

func (as *AuditSQLStore) Insert(ctx context.Context, audit *model.Audit) (*model.Audit, error) {
	audit.PreSave()

	terr := as.wrapper().transact(ctx, as.parent.db, func(tx *sqlx.Tx) error {
		stmt := squirrel.StatementBuilder.
			Insert("audits").
			Columns(
				"request_id", "created_at", "client_agent",
				"client_address", "status_code", "error", "event",
			).
			Values(
				audit.RequestID, audit.CreatedAt, audit.ClientAgent,
				audit.ClientAddress, audit.StatusCode, audit.Error, audit.Event,
			)

		query, args, err := stmt.ToSql()
		if err != nil {
			return fmt.Errorf("(*AuditSQLStore).Insert - id=%s; root_err=%v; %w",
				audit.RequestID, err, ErrSQLBuild)
		}

		result, err := tx.ExecContext(ctx, query, args...)
		switch {
		case as.wrapper().isDuplicateErr(err):
			return fmt.Errorf("(*AuditSQLStore).Insert request_id=%s; %w",
				audit.RequestID, ErrSQLDuplicate)
		case err != nil:
			return fmt.Errorf("(*AuditSQLStore).Insert - root_err=%v; %w", err, ErrSQLExecute)
		}

		n, err := result.RowsAffected()
		if err != nil {
			as.parent.logger.WithError(err).
				Warnf("database driver %v doesn't support RowsAffected",
					as.parent.db.DriverName())
		}
		if n == 0 {
			return fmt.Errorf("(*AuditSQLStore).Insert - id=%s; %w",
				audit.RequestID, store.ErrNoRows)
		}

		return nil
	})

	if terr != nil {
		return nil, terr
	}

	return audit, nil
}
