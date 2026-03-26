package service

import (
	"context"

	"gophermart/internal/repository"
)

type BalanceService struct {
	orders     repository.BalanceRepository
	withdrawals repository.WithdrawalRepository
}

func NewBalanceService(orders repository.BalanceRepository, withdrawals repository.WithdrawalRepository) *BalanceService {
	return &BalanceService{
		orders:      orders,
		withdrawals: withdrawals,
	}
}

type Balance struct {
	Current   float64
	Withdrawn float64
}

func (s *BalanceService) GetBalance(ctx context.Context, userID string) (Balance, error) {
	accrued, err := s.orders.GetTotalAccrual(ctx, userID)
	if err != nil {
		return Balance{}, err
	}
	withdrawn, err := s.withdrawals.GetTotalWithdrawn(ctx, userID)
	if err != nil {
		return Balance{}, err
	}
	return Balance{
		Current:   accrued - withdrawn,
		Withdrawn: withdrawn,
	}, nil
}
