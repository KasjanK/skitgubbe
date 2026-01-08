-- name: CreateUser :one
INSERT INTO users (
    id, created_at, updated_at, email, hashed_password
) VALUES (
    ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, ?
)
RETURNING *;
