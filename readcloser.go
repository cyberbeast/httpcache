package httpcache

import (
	"bytes"
	"context"
	"database/sql"
	"io"
	"net/http"
)

type cachedReadCloser struct {
	ctx      context.Context
	original io.ReadCloser
	buffer   *bytes.Buffer
	cache    *Queries
	data     func() CacheResponseParams
	tee      io.Reader
}

func (b *cachedReadCloser) Read(p []byte) (int, error) { return b.tee.Read(p) }
func (b *cachedReadCloser) Close() error {
	_, err := b.cache.CacheResponse(b.ctx, b.data())
	if err != nil {
		return err
	}

	return b.original.Close()
}

func newCachedReadCloser(hash string, cache *Queries, resp *http.Response) (*cachedReadCloser) {
	buffer := &bytes.Buffer{}
	return &cachedReadCloser{
		ctx:      resp.Request.Context(),
		original: resp.Body,
		buffer:   buffer,
		cache:    cache,
		data: func() CacheResponseParams {
			return CacheResponseParams{
				ReqHash:    hash,
				Body:       sql.NullString{String: buffer.String(), Valid: true},
				Headers:    sql.NullString{String: "", Valid: true},
				StatusCode: sql.NullInt64{Int64: int64(resp.StatusCode), Valid: true},
			}
		},
		tee: io.TeeReader(resp.Body, buffer),
	}
}
