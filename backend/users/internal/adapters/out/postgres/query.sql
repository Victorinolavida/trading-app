
-- name: SaveUser :exec
INSERT INTO users(
    id,email,first_name, last_name, password_hash, created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7);

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: UpdateUser :exec
UPDATE users
    set first_name = $1,
    last_name = $2,
    updated_at = $3
WHERE id = $4;


