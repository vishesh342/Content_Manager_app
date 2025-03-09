
-- name: CreatePlatform :one
INSERT INTO platforms (
platform_name, api_endpoint, created_at
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: GetPlatform :one
SELECT * FROM platforms
WHERE platform_name = $1 LIMIT 1;

-- name: UpdatePlatform :exec
UPDATE platforms
SET  api_endpoint= $2
WHERE platform_name = $1;

-- name: DeletePlatform :exec
DELETE FROM platforms
WHERE platform_name = $1;