package httpcache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"slices"
	"strings"
)

type RequestHashFn func(req *http.Request) string

func simpleRequestHash(req *http.Request) string {
	return fmt.Sprintf("%s:%s:%s", req.Method, req.URL.String(), hash(req.Header))
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
