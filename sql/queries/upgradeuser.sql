-- name: UpgradeUser :exec
UPDATE users
SET is_chirpy_red = true, updated_at = NOW()
WHERE users.id = $1;
