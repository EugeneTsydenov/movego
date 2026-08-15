-- name: SaveSession :exec
INSERT INTO sessions (
    id,
    user_id,
    secret_hash,
    user_agent,
    client_ip,
    expires_at,
    created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
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

-- name: Delete :exec
DELETE FROM sessions
WHERE id = $1;
