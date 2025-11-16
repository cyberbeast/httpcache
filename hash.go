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
	return fmt.Sprintf("%s:%s:%s", req.Method, sha256str([]byte(req.URL.String())), hash(req.Header))
}

func sha256str(key []byte) string {
	hash := sha256.Sum256(key)
	return hex.EncodeToString(hash[:])
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

	return sha256str([]byte(sb.String()))
}
