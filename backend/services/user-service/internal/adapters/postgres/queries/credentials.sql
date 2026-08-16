-- name: SaveCredential :exec
INSERT INTO credentials (
    id,
    user_id,
    password_hash,
    provider,
    provider_key
) VALUES (
    $1, $2, $3, $4, $5
)
ON CONFLICT (id) DO UPDATE SET
    password_hash = EXCLUDED.password_hash,
    provider = EXCLUDED.provider,
    provider_key = EXCLUDED.provider_key;

-- name: FindForAuth :one
SELECT
    sqlc.embed(c),
    sqlc.embed(u)
FROM credentials c
INNER JOIN users u ON u.id = c.user_id
WHERE u.email = $1
AND c.provider = $2
AND u.deleted_at IS NULL;
