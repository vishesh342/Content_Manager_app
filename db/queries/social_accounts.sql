-- name: CreateAccount :one
INSERT INTO social_accounts (
username, platform_username, access_token, refresh_token, expires_at
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetAccount :one
SELECT * FROM social_accounts
WHERE username = $1 LIMIT 1;

-- name: UpdateAccount :exec
UPDATE social_accounts
SET  access_token= $3,
     refresh_token = $4,
     expires_at = $5,
     updated_at = $6
WHERE username = $1 AND platform_username = $2;

-- name: DeleteAccount :exec
DELETE FROM social_accounts
WHERE username = $1 AND platform_username = $2;