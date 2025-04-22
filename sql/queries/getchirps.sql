-- name: GetChirps :many
SELECT chirps.*
FROM chirps
ORDER BY created_at ASC;
