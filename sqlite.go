package httpcache

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var ddl string

type SQLiteSource string

func (s SQLiteSource) name() string { return "sqlite" }

func (s SQLiteSource) filepath() string { return "file://" + string(s) }

func initSQLiteDB(ctx context.Context, src SQLiteSource) (*sql.DB, error) {
	db, err := sql.Open(src.name(), src.filepath())
	if err != nil {
		return nil, fmt.Errorf("opening db: %w", err)
	}

	if _, err := db.ExecContext(ctx, ddl); err != nil {
		return db, fmt.Errorf("creating db schema: %w", err)
	}

	return db, nil
}
