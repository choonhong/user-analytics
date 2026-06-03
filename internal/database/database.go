package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/choonhong/user-analytics/ent"
	_ "modernc.org/sqlite"
)

func Open(ctx context.Context, dsn string) (*ent.Client, *sql.DB, error) {
	if err := ensureDBDir(dsn); err != nil {
		return nil, nil, fmt.Errorf("create database directory: %w", err)
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}

	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()

		return nil, nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := ent.NewClient(ent.Driver(drv))

	if err := client.Schema.Create(ctx); err != nil {
		_ = client.Close()
		_ = db.Close()

		return nil, nil, fmt.Errorf("migrate schema: %w", err)
	}

	return client, db, nil
}

func ensureDBDir(dsn string) error {
	if strings.HasPrefix(dsn, "file:") || strings.Contains(dsn, "mode=memory") {
		return nil
	}

	dir := filepath.Dir(dsn)
	if dir == "." || dir == "" {
		return nil
	}

	return os.MkdirAll(dir, 0o755)
}
