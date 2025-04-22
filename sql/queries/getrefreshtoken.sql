-- name: GetRefreshToken :one
SELECT refresh_tokens.*
FROM refresh_tokens
WHERE refresh_tokens.token = $1;
