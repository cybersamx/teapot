package sqlstore

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/jmoiron/sqlx"
)

const (
	migrationsDir = "migrations"
)

//go:embed migrations/sqlite/*up.sql
var sqliteUpScripts embed.FS

//go:embed migrations/sqlite/*down.sql
var sqliteDownScripts embed.FS

//go:embed migrations/mysql/*up.sql
var mysqlUpScripts embed.FS

//go:embed migrations/mysql/*down.sql
var mysqlDownScripts embed.FS

//go:embed migrations/postgres/*up.sql
var postgresUpScripts embed.FS

//go:embed migrations/postgres/*down.sql
var postgresDownScripts embed.FS

var (
	ErrMigration = errors.New("failed to run migration")
	ErrReadDir   = errors.New("failed to read directory")
)

type Migrator struct {
	sqlstore *SQLStore
}

func NewMigrator(sqlStore *SQLStore) *Migrator {
	return &Migrator{
		sqlstore: sqlStore,
	}
}

func (m *Migrator) execFile(ctx context.Context, fs *embed.FS, filename string) error {
	buf, err := fs.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("(*Migrator).execFile - migrationScript=%s; rootErr=%v; %w",
			filename, err, ErrMigration)
	}

	return m.sqlstore.wrapper.transact(ctx, m.sqlstore.db, func(tx *sqlx.Tx) error {
		stmt := string(buf)
		_, terr := tx.ExecContext(ctx, stmt)
		if terr != nil {
			return fmt.Errorf("(*Migrator).execFile - migrationScript=%s; rootErr=%v; %w", filename, terr, ErrSQLTx)
		}

		return nil
	})
}

func (m *Migrator) execDir(ctx context.Context, fs *embed.FS, dirname string) error {
	files, err := fs.ReadDir(dirname)
	if err != nil {
		return fmt.Errorf("(*Migrator).execDir - dirname=%s; %w", dirname, ErrReadDir)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	for _, file := range files {
		filename := filepath.Join(dirname, file.Name())
		m.sqlstore.logger.WithFields(map[string]any{"file": filename}).Info("Running migration script")

		if err := m.execFile(ctx, fs, filename); err != nil {
			return err
		}
	}

	return nil
}

func (m *Migrator) up(ctx context.Context, driver string) error {
	var fs *embed.FS
	var dirname string

	switch driver {
	case driverMySQL:
		fs = &mysqlUpScripts
		dirname = "mysql"
	case driverPostgres:
		fs = &postgresUpScripts
		dirname = "postgres"
	case driverSQLite:
		fs = &sqliteUpScripts
		dirname = "sqlite"
	default:
		return fmt.Errorf("(*Migrator).up - driver=%s: %w", driver, ErrDBNotSupported)
	}

	dirname = filepath.Join(migrationsDir, dirname)

	return m.execDir(ctx, fs, dirname)
}

func (m *Migrator) down(ctx context.Context, driver string) error {
	var fs *embed.FS
	var dirname string

	switch driver {
	case driverMySQL:
		fs = &mysqlDownScripts
		dirname = "mysql"
	case driverPostgres:
		fs = &postgresDownScripts
		dirname = "postgres"
	case driverSQLite:
		fs = &sqliteDownScripts
		dirname = "sqlite"
	default:
		return fmt.Errorf("(*Migrator).down - driver=%s: %w", driver, ErrDBNotSupported)
	}

	dirname = filepath.Join(migrationsDir, dirname)

	return m.execDir(ctx, fs, dirname)
}
