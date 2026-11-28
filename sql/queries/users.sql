-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
  gen_random_uuid(),
  NOW(),
  NOW(),
  $1,
  $2
)
RETURNING *;

-- name: ResetDB :exec
DELETE FROM users;

-- name: GetUserByEmail :one
SELECT
  *
FROM
  users
WHERE
  email = $1;

-- name: UpdateUserEmailAndPassword :one
UPDATE users
SET email = $2, hashed_password = $3, updated_at = $4 /*updated_at drugače kot v naslednjem primeru*/
WHERE id = $1
RETURNING *;

-- name: ChirpyRedUpgrade :exec
UPDATE users
SET is_chirpy_red = TRUE, updated_at = NOW()
WHERE id = $1;
