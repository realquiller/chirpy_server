-- name: GetChirp :one
SELECT chirps.*
FROM chirps
WHERE chirps.id = $1;
