-- name: CreateUrl :one
INSERT INTO urls (id, original_url, short_code, created_at)
VALUES ($1, $2, $3, $4)
RETURNING short_code;

-- name: GetUrlByShortCode :one
SELECT * FROM urls WHERE short_code = $1;

-- name: GetUrlByOriginalUrl :one
SELECT * FROM urls WHERE original_url = $1;

