package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/pressly/goose/v3"
)

// Connect opens a SQLite database connection and runs migrations.
func Connect(ctx context.Context, dataDir string) (*sql.DB, error) {
	if dataDir == "" {
		return nil, fmt.Errorf("data.dir is not set")
	}
	dbPath := filepath.Join(dataDir, "floyd.db")

	db, err := openDB(dbPath)
	if err != nil {
		return nil, err
	}

	if err = db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Back-fill columns that were added to the initial migration after
	// some databases had already been created. This runs before goose
	// so the SQL migrations always see a consistent schema.
	if err := ensureColumns(ctx, db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ensure columns: %w", err)
	}

	goose.SetBaseFS(FS)

	if err := goose.SetDialect("sqlite3"); err != nil {
		slog.Error("Failed to set dialect", "error", err)
		return nil, fmt.Errorf("failed to set dialect: %w", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		slog.Error("Failed to apply migrations", "error", err)
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	return db, nil
}

// ensureColumns idempotently adds columns that may be missing from
// databases created before the column was part of the initial
// migration. SQLite does not support IF NOT EXISTS for ALTER TABLE
// ADD COLUMN, so we check pragma_table_info first.
//
// This also handles schema migrations from old versions (e.g., name -> title).
func ensureColumns(ctx context.Context, db *sql.DB) error {
	type col struct {
		Table  string
		Column string
		DDL    string
	}

	backfills := []col{
		{"sessions", "title", "ALTER TABLE sessions ADD COLUMN title TEXT"},
		{"sessions", "parent_session_id", "ALTER TABLE sessions ADD COLUMN parent_session_id TEXT"},
		{"sessions", "message_count", "ALTER TABLE sessions ADD COLUMN message_count INTEGER NOT NULL DEFAULT 0"},
		{"sessions", "prompt_tokens", "ALTER TABLE sessions ADD COLUMN prompt_tokens INTEGER NOT NULL DEFAULT 0"},
		{"sessions", "completion_tokens", "ALTER TABLE sessions ADD COLUMN completion_tokens INTEGER NOT NULL DEFAULT 0"},
		{"sessions", "cost", "ALTER TABLE sessions ADD COLUMN cost REAL NOT NULL DEFAULT 0.0"},
		{"sessions", "summary_message_id", "ALTER TABLE sessions ADD COLUMN summary_message_id TEXT"},
		{"sessions", "todos", "ALTER TABLE sessions ADD COLUMN todos TEXT"},
		{"sessions", "cache_read_tokens", "ALTER TABLE sessions ADD COLUMN cache_read_tokens INTEGER NOT NULL DEFAULT 0"},
	}

	for _, c := range backfills {
		// Check if table exists first
		var tableExists int
		err := db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?",
			c.Table,
		).Scan(&tableExists)
		if err != nil || tableExists == 0 {
			// Table doesn't exist yet, skip this backfill (will be created by migrations)
			continue
		}

		var count int
		err = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?",
			c.Table, c.Column,
		).Scan(&count)
		if err != nil {
			return fmt.Errorf("checking column %s.%s: %w", c.Table, c.Column, err)
		}
		if count == 0 {
			if _, err := db.ExecContext(ctx, c.DDL); err != nil {
				return fmt.Errorf("adding column %s.%s: %w", c.Table, c.Column, err)
			}
			slog.Info("Added missing column", "table", c.Table, "column", c.Column)
		}
	}

	// Handle name -> title migration for old databases
	// Check if sessions has 'name' column and 'title' is empty
	var hasName int
	err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM pragma_table_info('sessions') WHERE name = 'name'").Scan(&hasName)
	if err != nil {
		return fmt.Errorf("checking for name column: %w", err)
	}
	if hasName > 0 {
		// Migrate name to title where title is NULL
		result, err := db.ExecContext(ctx,
			"UPDATE sessions SET title = name WHERE title IS NULL OR title = ''")
		if err != nil {
			return fmt.Errorf("migrating name to title: %w", err)
		}
		rows, _ := result.RowsAffected()
		if rows > 0 {
			slog.Info("Migrated sessions from name to title", "count", rows)
		}
	}

	return nil
}
