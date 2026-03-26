package repository

import (
	"context"

	"gophermart/internal/model"
)

var (
	ErrNotFound = errNotFound{}
	ErrConflict = errConflict{}
)

type errNotFound struct{}

func (errNotFound) Error() string { return "not found" }

type errConflict struct{}

func (errConflict) Error() string { return "conflict" }

type UserRepository interface {
	Create(ctx context.Context, login, passwordHash string) (*model.User, error)
	GetByLogin(ctx context.Context, login string) (*model.User, error)
	GetByID(ctx context.Context, id string) (*model.User, error)
}

type OrderRepository interface {
	Create(ctx context.Context, userID, number, status string, accrual *float64) (*model.Order, error)
	GetByNumber(ctx context.Context, number string) (*model.Order, error)
	GetByUserID(ctx context.Context, userID string) ([]*model.Order, error)
	GetPendingAccrual(ctx context.Context) ([]*model.Order, error)
	UpdateAccrual(ctx context.Context, orderID int64, status string, accrual *float64) error
}

type WithdrawalRepository interface {
	Create(ctx context.Context, userID, orderNumber string, sum float64) (*model.Withdrawal, error)
	GetByUserID(ctx context.Context, userID string) ([]*model.Withdrawal, error)
	GetTotalWithdrawn(ctx context.Context, userID string) (float64, error)
}

type BalanceRepository interface {
	GetTotalAccrual(ctx context.Context, userID string) (float64, error)
}
