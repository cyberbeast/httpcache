package httpcache

import (
	"net/http"
	"net/url"
	"testing"
)

func mustParse(t *testing.T, urlstr string) *url.URL {
	t.Helper()
	parsed, err := url.ParseRequestURI(urlstr)
	if err != nil {
		t.Fatalf("unable to parse url string: %v", err)
	}

	return parsed
}

func TestSimpleRequestHash(t *testing.T) {
	req := &http.Request{
		Method: http.MethodGet,
		URL:    mustParse(t, "http://www.google.com"),
		Header: http.Header{"key": {"value"}},
	}

	if got, want := simpleRequestHash(req), "GET:253d142703041dd25197550a0fc11d6ac03befc1e64a1320009f1edf400c39ad:7cb0ba540850f2f8b7f62da704748662704cfa62c97c36bf251f26d656610656"; got != want {
		t.Fatalf("expected %s; got %s", want, got)
	}
}
