package postgres

import (
	"context"

	"github.com/cyberbeast/httpcache"
	"github.com/cyberbeast/httpcache/internal/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type store struct{ queries *postgres.Queries }

func (s *store) CacheResponse(ctx context.Context, arg httpcache.Params) (httpcache.Response, error) {
	return wrapPostgresResponse(s.queries.CacheResponse(ctx, postgres.CacheResponseParams{
		ReqHash:    arg.ReqHash,
		Body:       pgtype.Text{String: arg.Body, Valid: true},
		Headers:    pgtype.Text{String: arg.Headers, Valid: true},
		StatusCode: pgtype.Int4{Int32: int32(arg.StatusCode), Valid: true},
	}))
}

func (s *store) DeleteAllResponses(ctx context.Context) error {
	return s.queries.DeleteAllResponses(ctx)
}

func (s *store) GetResponse(ctx context.Context, reqHash string) (httpcache.Response, error) {
	return wrapPostgresResponse(s.queries.GetResponse(ctx, reqHash))
}

func wrapPostgresResponse(res postgres.Response, err error) (httpcache.Response, error) {
	return httpcache.Response{
		ReqHash:    res.ReqHash,
		Body:       res.Body.String,
		Headers:    res.Headers.String,
		StatusCode: int(res.StatusCode.Int32),
		UpdatedAt:  res.UpdatedAt.Time.String(),
	}, err
}

type Connection struct{ *pgx.Conn }

func (c Connection) Init(ctx context.Context) (httpcache.ResponseCacher, error) {
	if _, err := c.Exec(ctx, postgres.Schema); err != nil {
		return nil, err
	}

	return &store{queries: postgres.New(c.Conn)}, nil
}
