
-- name: CreateUser :one
INSERT INTO users (
username, email, hashed_password, created_at
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE username = $1 LIMIT 1;

-- name: UpdateUser :exec
UPDATE users
SET  hashed_password= $2,
    updated_at = $3
WHERE username = $1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE username = $1;