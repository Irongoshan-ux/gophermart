-- name: CreateWithdrawal :one
INSERT INTO withdrawals (user_id, order_number, sum, processed_at)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, order_number, sum, processed_at;

-- name: GetWithdrawalsByUserID :many
SELECT id, user_id, order_number, sum, processed_at
FROM withdrawals
WHERE user_id = $1
ORDER BY processed_at DESC;

-- name: GetTotalWithdrawnByUserID :one
SELECT COALESCE(SUM(sum), 0)::double precision AS total
FROM withdrawals
WHERE user_id = $1;
