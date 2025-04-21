-- name: CreatePost :one
INSERT INTO posts (
    id, content, media_type, media_urns, scheduled_time, visibility, account_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetPost :one
SELECT * FROM posts
WHERE id = $1;

-- name: UpdatePost :one
UPDATE posts
SET content = $2,
    media_type = $3,
    media_urns = $4,
    scheduled_time = $5,
    visibility = $6,
    account_id = $7
WHERE id = $1
RETURNING *;

-- name: DeletePost :exec
DELETE FROM posts
WHERE id = $1;

-- name: ListPosts :many
SELECT * FROM posts
ORDER BY created_at DESC;

-- name: ListPostsPaginated :many
SELECT * FROM posts
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;
