-- name: GetByID :one
SELECT * FROM securities 
WHERE public_id = $1;

-- name: GetByISIN :one
SELECT * FROM securities 
WHERE isin = $1;

-- name: GetEnableList :many
SELECT * FROM securities
WHERE enable = true;

-- name: GetByIDs :many
SELECT * FROM securities
WHERE public_id = ANY($1::uuid[])
ORDER BY issue_date DESC;

-- name: GetByISINs :many
SELECT * FROM securities
WHERE isin = ANY($1::text[])
ORDER BY issue_date DESC;

-- name: GetBySeries :many
SELECT * FROM securities
WHERE serie = ANY($1::text[])
ORDER BY issue_date DESC;

-- name: GetByTypes :many
SELECT * FROM securities
WHERE type = ANY($1::text[])
ORDER BY issue_date DESC;
