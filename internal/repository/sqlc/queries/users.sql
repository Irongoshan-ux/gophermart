-- name: CreateUser :one
INSERT INTO users (login, password_hash, created_at)
VALUES ($1, $2, $3)
RETURNING id, login, password_hash, created_at;

-- name: GetUserByLogin :one
SELECT id, login, password_hash, created_at
FROM users
WHERE login = $1;

-- name: GetUserByID :one
SELECT id, login, password_hash, created_at
FROM users
WHERE id = $1;
