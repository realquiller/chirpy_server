-- name: GetUser :one
SELECT users.*
FROM users
WHERE users.email = $1;
