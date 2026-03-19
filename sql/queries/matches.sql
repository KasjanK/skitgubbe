-- name: CreateGame :exec
INSERT INTO games (id, started_at, game_mode)
VALUES (?, ?, ?);

-- name: GetMatchHistoryForUser :many
SELECT
    gp.game_id,
    g.started_at,
    g.game_mode,
    gp.placement
FROM game_participants gp
JOIN games g ON g.id = gp.game_id
WHERE gp.user_id = ?
ORDER BY g.started_at DESC
LIMIT ?;

-- name: AddGameParticipant :exec
INSERT INTO game_participants (game_id, user_id, placement)
VALUES (?, ?, ?);
