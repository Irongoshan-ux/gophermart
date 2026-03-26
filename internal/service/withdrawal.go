package service

import (
	"context"
	"errors"

	"gophermart/internal/model"
	"gophermart/internal/repository"
)

var ErrInsufficientFunds = errors.New("insufficient funds")

type WithdrawalService struct {
	orders      repository.BalanceRepository
	withdrawals repository.WithdrawalRepository
}

func NewWithdrawalService(orders repository.BalanceRepository, withdrawals repository.WithdrawalRepository) *WithdrawalService {
	return &WithdrawalService{
		orders:      orders,
		withdrawals: withdrawals,
	}
}

func (s *WithdrawalService) Withdraw(ctx context.Context, userID, orderNumber string, sum float64) error {
	accrued, err := s.orders.GetTotalAccrual(ctx, userID)
	if err != nil {
		return err
	}
	withdrawn, err := s.withdrawals.GetTotalWithdrawn(ctx, userID)
	if err != nil {
		return err
	}
	current := accrued - withdrawn
	if sum > current || sum <= 0 {
		return ErrInsufficientFunds
	}
	_, err = s.withdrawals.Create(ctx, userID, orderNumber, sum)
	return err
}

func (s *WithdrawalService) GetWithdrawals(ctx context.Context, userID string) ([]*model.Withdrawal, error) {
	return s.withdrawals.GetByUserID(ctx, userID)
}
