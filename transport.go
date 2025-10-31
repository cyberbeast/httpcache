package httpcache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
)

type RequestHashFn func(req *http.Request) string

type cachedTransport struct {
	queries   *Queries
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
		Body:       io.NopCloser(strings.NewReader(res.Body.String)),
		StatusCode: int(res.StatusCode.Int64),
		Status:     http.StatusText(int(res.StatusCode.Int64)),
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

func NewTransport(ctx context.Context, src SQLiteSource, rt http.RoundTripper) (*cachedTransport, error) {
	db, err := initSQLiteDB(ctx, src)
	if err != nil {
		return nil, err
	}

	if rt == nil {
		rt = http.DefaultTransport
	}

	return &cachedTransport{
		rt:      rt,
		queries: New(db),
		reqHashFn: func(req *http.Request) string {
			return fmt.Sprintf("%s:%s:%s", req.Method, req.URL.String(), hash(req.Header))
		},
	}, nil
}

const delimiter = "|"

func hash(headers http.Header) string {
	keys := make([]string, 0, len(headers))

	for key := range headers {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	var sb strings.Builder
	for _, key := range keys {
		sb.WriteString(fmt.Sprintf("%s:%s%s", key, headers.Get(key), delimiter))
	}

	hash := sha256.Sum256([]byte(sb.String()))
	return hex.EncodeToString(hash[:])
}
