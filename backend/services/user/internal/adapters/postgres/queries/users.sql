-- name: SaveUser :exec
INSERT INTO users (
    id,
    email,
    tag,
    display_name,
    role,
    created_at,
    updated_at,
    deleted_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
ON CONFLICT (id) DO UPDATE SET
    email = EXCLUDED.email,
    tag = EXCLUDED.tag,
    display_name = EXCLUDED.display_name,
    role = EXCLUDED.role,
    updated_at = EXCLUDED.updated_at,
    deleted_at = EXCLUDED.deleted_at;
