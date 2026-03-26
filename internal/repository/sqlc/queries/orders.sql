-- name: CreateOrder :one
INSERT INTO orders (user_id, number, status, accrual, uploaded_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, number, status, accrual, uploaded_at;

-- name: GetOrderByNumber :one
SELECT id, user_id, number, status, accrual, uploaded_at
FROM orders
WHERE number = $1;

-- name: GetOrdersByUserID :many
SELECT id, user_id, number, status, accrual, uploaded_at
FROM orders
WHERE user_id = $1
ORDER BY uploaded_at DESC;

-- name: GetOrdersPendingAccrual :many
SELECT id, user_id, number, status, accrual, uploaded_at
FROM orders
WHERE status IN ('NEW', 'PROCESSING')
ORDER BY uploaded_at ASC;

-- name: UpdateOrderAccrual :exec
UPDATE orders
SET status = $2, accrual = $3
WHERE id = $1;
