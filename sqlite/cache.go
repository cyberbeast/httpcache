package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cyberbeast/httpcache"
	"github.com/cyberbeast/httpcache/internal/sqlite"
)

type cache struct{ queries *sqlite.Queries }

func (c *cache) CacheResponse(ctx context.Context, arg httpcache.Params) (httpcache.Response, error) {
	return wrapSQLiteResponse(c.queries.CacheResponse(ctx, sqlite.CacheResponseParams{
		ReqHash:    arg.ReqHash,
		Body:       sql.NullString{String: arg.Body, Valid: true},
		Headers:    sql.NullString{String: arg.Headers, Valid: true},
		StatusCode: sql.NullInt64{Int64: int64(arg.StatusCode), Valid: true},
	}))
}

func (c *cache) DeleteAllResponses(ctx context.Context) error {
	return c.queries.DeleteAllResponses(ctx)
}

func (c *cache) GetResponse(ctx context.Context, reqHash string) (httpcache.Response, error) {
	return wrapSQLiteResponse(c.queries.GetResponse(ctx, reqHash))
}

func wrapSQLiteResponse(res sqlite.Response, err error) (httpcache.Response, error) {
	return httpcache.Response{
		ReqHash:    res.ReqHash,
		Body:       res.Body.String,
		Headers:    res.Headers.String,
		StatusCode: int(res.StatusCode.Int64),
		UpdatedAt:  res.UpdatedAt.String,
	}, err
}

type DB struct{ *sql.DB }

func (db DB) Init(ctx context.Context) (httpcache.ResponseCacher, error) {
	if _, err := db.ExecContext(ctx, sqlite.Schema); err != nil {
		return nil, fmt.Errorf("creating db schema: %w", err)
	}

	return &cache{queries: sqlite.New(db.DB)}, nil
}
