# httpcache

An HTTP transport for Go standard library's HTTP Client that caches HTTP Responses in a local SQLite database to speed up subsequent requests.

## Motivation

I was working on an internal tool at work that needed to run an expensive query over HTTP multiple times. Since the HTTP response didn't change between runs, I wrote `httpcache` to cap the maximum latency of making multiple HTTP requests to the latency of the initial cold request.

## Implementation

When an HTTP request is made, its signature is first queried against the cache (SQLite). If found, the corresponding HTTP response is returned unmodified.

If not found (_i.e. cold request_), the underlying `http.RoundTripper` is first called to execute the HTTP request and the underlying `http.Response.Body` replaced with a custom `io.ReadCloser` that buffers the response while it is being read and caches the response when it is closed.

The implementation currently uses a very basic hashing mechanism to create a signature of an outbound HTTP request. While most parts of an HTTP request are easy to serialize as text, the one component that is not trivial to serialize are the HTTP Headers. Since HTTP headers are expressed as an unordered map, the keys are currently sorted before concatenating each header using a pre-defined delimiter (`|`), and then computing its `sha256` hash to generate the unique signature. The signature of an HTTP request has the following form when stored in the cache -

```text
{HTTP Method}:{URL}:{HTTP Header Hash}
```

## Todo

- [ ] Custom caching strategy based on HTTP response status codes
- [ ] Bring your own HTTP request hasher
- [ ] Cache Invalidation support
- [ ] Improve test coverage
- [ ] Add build/contribution guide to readme
