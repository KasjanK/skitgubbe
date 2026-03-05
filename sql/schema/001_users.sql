-- +goose Up
CREATE TABLE users (
    username TEXT NOT NULL UNIQUE,
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    email TEXT UNIQUE NOT NULL,
    hashed_password TEXT NOT NULL
);

-- +goose Down
DROP TABLE users;
