-- name: GetResponse :one
SELECT * FROM responses WHERE req_hash = $1 LIMIT 1;

-- name: CacheResponse :one
INSERT INTO responses (
  req_hash, body, headers, status_code
)
VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: DeleteAllResponses :exec
DELETE FROM responses;