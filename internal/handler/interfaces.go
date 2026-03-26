package handler

import (
	"context"

	"gophermart/internal/model"
	"gophermart/internal/service"
)

type UserService interface {
	Register(ctx context.Context, login, password string) (*model.User, error)
	Login(ctx context.Context, login, password string) (*model.User, error)
}

type OrderService interface {
	AddOrder(ctx context.Context, userID, number string) (alreadyUploaded bool, err error)
	GetOrdersByUser(ctx context.Context, userID string) ([]*model.Order, error)
}

type BalanceService interface {
	GetBalance(ctx context.Context, userID string) (service.Balance, error)
}

type WithdrawalService interface {
	Withdraw(ctx context.Context, userID, orderNumber string, sum float64) error
	GetWithdrawals(ctx context.Context, userID string) ([]*model.Withdrawal, error)
}
