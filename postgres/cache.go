package postgres

import (
	"context"

	"github.com/cyberbeast/httpcache"
	"github.com/cyberbeast/httpcache/internal/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type cache struct{ queries *postgres.Queries }

func (c *cache) CacheResponse(ctx context.Context, arg httpcache.Params) (httpcache.Response, error) {
	return wrapPostgresResponse(c.queries.CacheResponse(ctx, postgres.CacheResponseParams{
		ReqHash:    arg.ReqHash,
		Body:       pgtype.Text{String: arg.Body, Valid: true},
		Headers:    pgtype.Text{String: arg.Headers, Valid: true},
		StatusCode: pgtype.Int4{Int32: int32(arg.StatusCode), Valid: true},
	}))
}

func (c *cache) DeleteAllResponses(ctx context.Context) error {
	return c.queries.DeleteAllResponses(ctx)
}

func (c *cache) GetResponse(ctx context.Context, reqHash string) (httpcache.Response, error) {
	return wrapPostgresResponse(c.queries.GetResponse(ctx, reqHash))
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

	return &cache{queries: postgres.New(c.Conn)}, nil
}
