package httpcache

import (
	"cmp"
	"context"
	"io"
	"net/http"
	"strings"
)

type Source interface {
	Init(ctx context.Context) (Querier, error)
}

type Querier interface {
	GetResponse(ctx context.Context, reqHash string) (Response, error)
	CacheResponse(ctx context.Context, arg Params) (Response, error)
	DeleteAllResponses(ctx context.Context) error
}

func NewTransport(ctx context.Context, src Source, rt http.RoundTripper) (*cachedTransport, error) {
	store, err := src.Init(ctx)
	if err != nil {
		return nil, err
	}

	return &cachedTransport{
		rt:        cmp.Or(rt, http.DefaultTransport),
		queries:   store,
		reqHashFn: simpleRequestHash,
	}, nil
}

type Params struct {
	ReqHash    string
	Body       string
	Headers    string
	StatusCode int
}

type Response struct {
	ReqHash    string
	Body       string
	Headers    string
	StatusCode int
	UpdatedAt  string
}

type cachedTransport struct {
	queries   Querier
	rt        http.RoundTripper
	reqHashFn RequestHashFn
}

func (ct cachedTransport) InvalidateAllResponses(ctx context.Context) error {
	return ct.queries.DeleteAllResponses(ctx)
}

func (ct cachedTransport) cachedRoundTrip(req *http.Request) *http.Response {
	res, err := ct.queries.GetResponse(req.Context(), ct.reqHashFn(req))
	if err != nil {
		return nil
	}

	return &http.Response{
		Body:       io.NopCloser(strings.NewReader(res.Body)),
		StatusCode: res.StatusCode,
		Status:     http.StatusText(res.StatusCode),
		// Header:     res.Headers.String,
	}
}

func (ct cachedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if res := ct.cachedRoundTrip(req); res != nil {
		return res, nil
	}

	res, err := ct.rt.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	res.Body = newCachedReadCloser(ct.reqHashFn(req), ct.queries, res)

	return res, nil
}
