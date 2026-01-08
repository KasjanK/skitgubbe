-- +goose Up
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    email TEXT UNIQUE NOT NULL,
    hashed_password TEXT NOT NULL
);

-- +goose Down
DROP TABLE users;
