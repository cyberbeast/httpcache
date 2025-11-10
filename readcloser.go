package httpcache

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

type cachedReadCloser struct {
	ctx      context.Context
	original io.ReadCloser
	buffer   *bytes.Buffer
	cache    ResponseCacher
	data     func() Params
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

func newCachedReadCloser(hash string, cache ResponseCacher, resp *http.Response) *cachedReadCloser {
	buffer := &bytes.Buffer{}
	return &cachedReadCloser{
		ctx:      resp.Request.Context(),
		original: resp.Body,
		buffer:   buffer,
		cache:    cache,
		data: func() Params {
			return Params{
				ReqHash:    hash,
				Body:       buffer.String(),
				Headers:    "",
				StatusCode: resp.StatusCode,
			}
		},
		tee: io.TeeReader(resp.Body, buffer),
	}
}
