package httpcache

import (
	"context"

	"github.com/cyberbeast/httpcache/internal/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type pgstore struct{ queries *postgres.Queries }

func (s *pgstore) CacheResponse(ctx context.Context, arg Params) (Response, error) {
	return wrapPostgresResponse(s.queries.CacheResponse(ctx, postgres.CacheResponseParams{
		ReqHash:    arg.ReqHash,
		Body:       pgtype.Text{String: arg.Body, Valid: true},
		Headers:    pgtype.Text{String: arg.Headers, Valid: true},
		StatusCode: pgtype.Int4{Int32: int32(arg.StatusCode), Valid: true},
	}))
}

func (s *pgstore) DeleteAllResponses(ctx context.Context) error {
	return s.queries.DeleteAllResponses(ctx)
}

func (s *pgstore) GetResponse(ctx context.Context, reqHash string) (Response, error) {
	return wrapPostgresResponse(s.queries.GetResponse(ctx, reqHash))
}

func wrapPostgresResponse(res postgres.Response, err error) (Response, error) {
	return Response{
		ReqHash:    res.ReqHash,
		Body:       res.Body.String,
		Headers:    res.Headers.String,
		StatusCode: int(res.StatusCode.Int32),
		UpdatedAt:  res.UpdatedAt.Time.String(),
	}, err
}

type PostgresSource struct{ *pgx.Conn }

func (p PostgresSource) Init(ctx context.Context) (Querier, error) {
	if _, err := p.Exec(ctx, postgres.Schema); err != nil {
		return nil, err
	}

	return &pgstore{queries: postgres.New(p)}, nil
}
