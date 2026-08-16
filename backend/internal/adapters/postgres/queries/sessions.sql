-- name: SaveSession :exec
INSERT INTO sessions (
    id,
    user_id,
    secret_hash,
    user_agent,
    client_ip,
    last_active_at,
    expires_at,
    created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
ON CONFLICT (id) DO UPDATE SET
    secret_hash = EXCLUDED.secret_hash,
    user_agent = EXCLUDED.user_agent,
    client_ip = EXCLUDED.client_ip,
    expires_at = EXCLUDED.expires_at;

-- name: FindValid :one
SELECT
    sqlc.embed(s),
    sqlc.embed(u)
FROM sessions s
INNER JOIN users u ON u.id = s.user_id
WHERE s.id = $1
  AND s.expires_at > now()
  AND u.deleted_at IS NULL;

-- name: FindByID :one
SELECT *
FROM sessions s
WHERE s.id = $1
    AND s.expires_at > now();

-- name: ListActiveByUserID :many
SELECT *
FROM sessions s
WHERE s.user_id = $1
    AND s.expires_at > now();

-- name: Delete :exec
DELETE FROM sessions
WHERE id = $1;
