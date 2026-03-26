package model

import "time"

type Withdrawal struct {
	ID          int64
	UserID      string
	OrderNumber string
	Sum         float64
	ProcessedAt time.Time
}
