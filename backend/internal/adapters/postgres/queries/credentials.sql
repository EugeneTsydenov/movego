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
