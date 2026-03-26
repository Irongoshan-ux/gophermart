package service

import (
	"context"
	"testing"

	"gophermart/internal/repository/mocks"
	"go.uber.org/mock/gomock"
)

func TestBalanceService_GetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	orderRepo := repository_mocks.NewMockBalanceRepository(ctrl)
	withRepo := repository_mocks.NewMockWithdrawalRepository(ctrl)
	svc := NewBalanceService(orderRepo, withRepo)

	ctx := context.Background()
	orderRepo.EXPECT().GetTotalAccrual(ctx, "user1").Return(1000.0, nil)
	withRepo.EXPECT().GetTotalWithdrawn(ctx, "user1").Return(200.0, nil)

	bal, err := svc.GetBalance(ctx, "user1")
	if err != nil {
		t.Fatalf("GetBalance: %v", err)
	}
	if bal.Current != 800.0 || bal.Withdrawn != 200.0 {
		t.Errorf("got %+v", bal)
	}
}
