-- name: CreateUserStats :exec
INSERT INTO user_stats (user_id)
VALUES (?);

-- name: GetUserStats :one
SELECT
    COUNT(*) AS total_games,
    COUNT(*) FILTER (WHERE placement = 1) AS wins,
    COUNT(*) FILTER (WHERE placement != 1) AS losses
FROM game_participants
WHERE user_id = ?;
