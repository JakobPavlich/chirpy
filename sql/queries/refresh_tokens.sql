-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, expires_at, revoked_at, user_id)
VALUES (
  $1,
  NOW(),
  NOW(),
  $2,
  NULL,
  $3
)
RETURNING *;

-- -- -- name: ResetDB :exec
-- -- DELETE FROM chirps;
--
-- -- name: GetChirps :many
-- SELECT * FROM chirps
-- ORDER BY created_at;
--
-- -- name: GetChirpByID :one
-- SELECT
--   *
-- FROM
--   chirps
-- WHERE
--   id = $1;
