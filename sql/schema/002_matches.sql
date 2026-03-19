-- +goose Up
CREATE TABLE games (
    id TEXT PRIMARY KEY,
    started_at TIMESTAMP NOT NULL,
    game_mode TEXT
);

CREATE TABLE game_participants (
    game_id TEXT NOT NULL REFERENCES games(id),
    user_id TEXT NOT NULL REFERENCES users(id),
    placement INTEGER NOT NULL,
    PRIMARY KEY (game_id, user_id)
);

-- +goose Down
DROP TABLE games;
DROP TABLE game_participants;
