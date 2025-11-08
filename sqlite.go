package httpcache

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cyberbeast/httpcache/internal/sqlite"
)

type sqliteStore struct{ queries *sqlite.Queries }

func (s *sqliteStore) CacheResponse(ctx context.Context, arg Params) (Response, error) {
	return wrapSQLiteResponse(s.queries.CacheResponse(ctx, sqlite.CacheResponseParams{
		ReqHash:    arg.ReqHash,
		Body:       sql.NullString{String: arg.Body, Valid: true},
		Headers:    sql.NullString{String: arg.Headers, Valid: true},
		StatusCode: sql.NullInt64{Int64: int64(arg.StatusCode), Valid: true},
	}))
}

func (s *sqliteStore) DeleteAllResponses(ctx context.Context) error {
	return s.queries.DeleteAllResponses(ctx)
}

func (s *sqliteStore) GetResponse(ctx context.Context, reqHash string) (Response, error) {
	return wrapSQLiteResponse(s.queries.GetResponse(ctx, reqHash))
}

func wrapSQLiteResponse(res sqlite.Response, err error) (Response, error) {
	return Response{
		ReqHash:    res.ReqHash,
		Body:       res.Body.String,
		Headers:    res.Headers.String,
		StatusCode: int(res.StatusCode.Int64),
		UpdatedAt:  res.UpdatedAt.String,
	}, err
}

type SQLiteSource struct{ *sql.DB }

func (s SQLiteSource) Init(ctx context.Context) (Querier, error) {
	if _, err := s.ExecContext(ctx, sqlite.Schema); err != nil {
		return nil, fmt.Errorf("creating db schema: %w", err)
	}

	return &sqliteStore{queries: sqlite.New(s)}, nil
}
