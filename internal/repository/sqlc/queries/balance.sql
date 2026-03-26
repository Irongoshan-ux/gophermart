-- name: GetTotalAccrualByUserID :one
SELECT COALESCE(SUM(accrual), 0)::double precision AS total
FROM orders
WHERE user_id = $1 AND status = 'PROCESSED';
