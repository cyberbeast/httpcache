-- name: GetResponse :one
SELECT * FROM responses WHERE req_hash = ? LIMIT 1;

-- name: CacheResponse :one
INSERT INTO responses (
  req_hash, body, headers, status_code
)
VALUES (
  ?, ?, ?, ?
)
RETURNING *;

-- name: DeleteAllResponses :exec
DELETE FROM responses;