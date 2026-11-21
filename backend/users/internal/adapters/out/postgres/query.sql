
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

-- name: StorePasswordResetToken :exec
INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
ON CONFLICT (user_id) DO UPDATE
SET token_hash = EXCLUDED.token_hash, expires_at = EXCLUDED.expires_at;

-- name: ConsumePasswordResetToken :one
DELETE FROM password_reset_tokens
WHERE token_hash = $1 AND expires_at > now()
RETURNING user_id;

-- name: UpdateUserPassword :exec
UPDATE users SET password_hash = $1, updated_at = now()
WHERE id = $2;

