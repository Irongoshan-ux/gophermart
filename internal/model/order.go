package model

import "time"

type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

type Order struct {
	ID         int64
	UserID     string
	Number     string
	Status     OrderStatus
	Accrual    *float64
	UploadedAt time.Time
}
