# httpcache

[![Go](https://github.com/cyberbeast/httpcache/actions/workflows/go.yml/badge.svg)](https://github.com/cyberbeast/httpcache/actions/workflows/go.yml)
![coverage](https://raw.githubusercontent.com/cyberbeast/httpcache/badges/.badges/main/coverage.svg)

An HTTP transport for Go standard library's HTTP Client that caches HTTP Responses in a database to speed up subsequent requests. Currently supported:

1. `sqlite`
2. `postgres` (coming soon)

## Motivation

I was working on an unrelated tool that ran expensive queries over HTTP multiple times. Since the HTTP response didn't change between runs, I wrote `httpcache` to cap the maximum latency of making multiple HTTP requests to the latency of the initial cold request.

## Implementation

When an HTTP request is made, its signature is first queried against the cache (SQLite). If found, the corresponding HTTP response is returned unmodified.

If not found (_i.e. cold request_), the underlying `http.RoundTripper` is first called to execute the HTTP request and the underlying `http.Response.Body` is replaced with a custom `io.ReadCloser` that buffers the response while it is being read and caches the response when it is closed.

The implementation currently uses a very basic hashing mechanism to create a signature of an outbound HTTP request. While most parts of an HTTP request are easy to serialize as string, the one component that is not trivial to serialize are the HTTP Headers. Since HTTP headers are expressed as an unordered map, the keys are currently sorted before concatenating each header using a pre-defined delimiter (`|`), and then its `sha256` hash computed to generate the unique signature. This also has the benefit of not leaking Authentication headers as plaintext in the cache. The signature of an HTTP request has the following form when stored in the cache -

```text
{HTTP Method}:{URL}:{HTTP Header Hash}
```

## Todo

- [x] Basic Cache Invalidation support
- [x] Add `postgres` as an alternative source
- [ ] Improve test coverage
- [ ] Add `pgxpool` support for `postgres` source
- [ ] Custom caching strategy based on HTTP response status codes
- [ ] Explore parsing [Cache Control headers](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Cache-Control) for request-level control
- [ ] Bring your own HTTP request hasher
- [ ] Configurable Cache Invalidation support
- [ ] Add build/contribution guide to readme
- [ ] Support returning HTTP headers in cached response
- [ ] Improve documentation
