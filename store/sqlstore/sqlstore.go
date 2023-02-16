package sqlstore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/cybersamx/teapot/common"
	"github.com/cybersamx/teapot/model"
	"github.com/cybersamx/teapot/store"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

const (
	dbPingTimeout  = 10 * time.Second
	dbPingRetry    = 3
	driverPostgres = "pgx"
	driverMySQL    = "mysql"
	driverSQLite   = "sqlite3"
)

var (
	ErrSQLExecute       = errors.New("failed to execute sql query")
	ErrDBNotSupported   = errors.New("non-supported database driver")
	ErrDBOpen           = errors.New("failed to Connect the database")
	ErrInvalidTableName = errors.New("invalid table name")
	ErrSQLDuplicate     = errors.New("duplicate record found")
)

type SQLStore struct {
	cfg *model.Config

	db        *sqlx.DB
	logger    *logrus.Logger
	stmtCache *squirrel.StmtCache
	wrapper   wrapper

	audits *AuditSQLStore
}

func New(driver string, logger *logrus.Logger) *SQLStore {
	switch driver {
	case driverMySQL:
		return NewSQLiteStore(logger)
	case driverPostgres:
		return NewPostgresStore(logger)
	case driverSQLite:
		return NewSQLiteStore(logger)
	default:
		logrus.Panicf("Driver %s is not supported", driver)
		return nil
	}
}

func (ss *SQLStore) migration() error {
	ss.logger.Infof("Starting database migrations")
	mig := NewMigrator(ss)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := mig.up(ctx, ss.cfg.Store.Driver)
	if err != nil {
		return err
	}

	return nil
}

func (ss *SQLStore) Clear(ctx context.Context) error {
	tables, err := ss.wrapper.tableNames(ctx, ss.db)
	if err != nil {
		return err
	}

	for _, table := range tables {
		err := ss.wrapper.clearTable(ctx, ss.db, table)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ss *SQLStore) Close() error {
	if ss.db != nil {
		return ss.db.Close()
	}

	return nil
}

func (ss *SQLStore) Config() *model.Config {
	return ss.cfg
}

func (ss *SQLStore) Connect(ctx context.Context, cfg *model.Config) error {
	ss.cfg = cfg

	ss.logger.WithFields(map[string]any{
		"driver": ss.cfg.Store.Driver,
		"dsn":    common.MaskPassword(ss.cfg.Store.DSN),
	}).Infoln("Connecting to datastore")

	if !isDriverSupported(cfg.Store.Driver) {
		return fmt.Errorf("(*SQLStore).Connect - driver=%s; %w", cfg.Store.Driver, ErrDBNotSupported)
	}

	// Open the db connection.
	var err error
	ss.db, err = sqlx.Open(ss.cfg.Store.Driver, ss.cfg.Store.DSN)
	if err != nil {
		return fmt.Errorf("(*SQLStore).Connect - dsn=%s; root_err=%v; %w",
			common.MaskPassword(ss.cfg.Store.DSN), err, ErrDBOpen)
	}

	// PingContext the db connection.
	if err := ss.PingContext(ctx); err != nil {
		return err
	}

	ss.logger.Infoln("Successfully connected to datastore")

	// Post setup.
	ss.wrapper.postConnect(ss.db, ss.cfg)

	return nil
}

func (ss *SQLStore) InitDB(ctx context.Context) error {
	if err := ss.migration(); err != nil {
		return fmt.Errorf("(*SQLStore).InitDB - dsn=%s; root_err=%v; %w",
			common.MaskPassword(ss.cfg.Store.DSN), err, ErrMigration)
	}

	// Other store setups.
	ss.audits = newAuditSQLStore(ss)

	return nil
}

func (ss *SQLStore) PingContext(ctx context.Context) error {
	var err error

	for i := 0; i < dbPingRetry; i++ {
		err = func() error {
			ctx, cancel := context.WithTimeout(ctx, dbPingTimeout)
			defer cancel()

			if ss.db == nil {
				return errors.New("database db not initialized yet")
			}
			return ss.db.PingContext(ctx)
		}()

		if err == nil {
			return nil
		}

		ss.logger.Infof("attempt %d: failed to ping %s - retrying in %v",
			i+1, common.MaskPassword(ss.cfg.Store.DSN), dbPingTimeout,
		)

		time.Sleep(dbPingTimeout)
	}

	return fmt.Errorf("(*SQLStore).PingContext - dsn=%s; root_err=%v; %w",
		common.MaskPassword(ss.cfg.Store.DSN), err, ErrDBOpen)
}

func (ss *SQLStore) Audits() store.AuditStore {
	return ss.audits
}
