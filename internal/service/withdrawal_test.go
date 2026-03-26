package service

import (
	"context"
	"testing"
	"time"

	"gophermart/internal/model"
	"gophermart/internal/repository/mocks"
	"go.uber.org/mock/gomock"
)

func TestWithdrawalService_Withdraw(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	orderRepo := repository_mocks.NewMockBalanceRepository(ctrl)
	withRepo := repository_mocks.NewMockWithdrawalRepository(ctrl)
	svc := NewWithdrawalService(orderRepo, withRepo)

	ctx := context.Background()
	orderRepo.EXPECT().GetTotalAccrual(ctx, "user1").Return(500.0, nil)
	withRepo.EXPECT().GetTotalWithdrawn(ctx, "user1").Return(0.0, nil)
	withRepo.EXPECT().
		Create(ctx, "user1", "123", 100.0).
		Return(&model.Withdrawal{ID: 1, UserID: "user1", OrderNumber: "123", Sum: 100, ProcessedAt: time.Now()}, nil)

	err := svc.Withdraw(ctx, "user1", "123", 100.0)
	if err != nil {
		t.Fatalf("Withdraw: %v", err)
	}
}

func TestWithdrawalService_Withdraw_InsufficientFunds(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	orderRepo := repository_mocks.NewMockBalanceRepository(ctrl)
	withRepo := repository_mocks.NewMockWithdrawalRepository(ctrl)
	svc := NewWithdrawalService(orderRepo, withRepo)

	ctx := context.Background()
	orderRepo.EXPECT().GetTotalAccrual(ctx, "user1").Return(50.0, nil)
	withRepo.EXPECT().GetTotalWithdrawn(ctx, "user1").Return(0.0, nil)

	err := svc.Withdraw(ctx, "user1", "123", 100.0)
	if err != ErrInsufficientFunds {
		t.Errorf("got err %v", err)
	}
}
