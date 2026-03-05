-- name: CreateUser :one
INSERT INTO users (
    username, id, created_at, updated_at, email, hashed_password
) VALUES (
    ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, ?
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = ? LIMIT 1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = ? LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = ? LIMIT 1;
