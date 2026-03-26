package service

import (
	"context"
	"testing"
	"time"

	"gophermart/internal/model"
	"gophermart/internal/repository"
	"gophermart/internal/repository/mocks"
	"go.uber.org/mock/gomock"
)

func TestOrderService_AddOrder_New(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := repository_mocks.NewMockOrderRepository(ctrl)
	svc := NewOrderService(repo, nil)

	ctx := context.Background()
	repo.EXPECT().GetByNumber(ctx, "123").Return(nil, repository.ErrNotFound)
	repo.EXPECT().
		Create(ctx, "user1", "123", "NEW", nil).
		Return(&model.Order{ID: 1, UserID: "user1", Number: "123", Status: model.OrderStatusNew, UploadedAt: time.Now()}, nil)

	already, err := svc.AddOrder(ctx, "user1", "123")
	if err != nil {
		t.Fatalf("AddOrder: %v", err)
	}
	if already {
		t.Error("expected already=false")
	}
}

func TestOrderService_AddOrder_AlreadyByUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := repository_mocks.NewMockOrderRepository(ctrl)
	svc := NewOrderService(repo, nil)

	ctx := context.Background()
	repo.EXPECT().
		GetByNumber(ctx, "123").
		Return(&model.Order{ID: 1, UserID: "user1", Number: "123", Status: model.OrderStatusProcessed, UploadedAt: time.Now()}, nil)

	already, err := svc.AddOrder(ctx, "user1", "123")
	if err != nil {
		t.Fatalf("AddOrder: %v", err)
	}
	if !already {
		t.Error("expected already=true")
	}
}

func TestOrderService_AddOrder_Conflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := repository_mocks.NewMockOrderRepository(ctrl)
	svc := NewOrderService(repo, nil)

	ctx := context.Background()
	repo.EXPECT().
		GetByNumber(ctx, "123").
		Return(&model.Order{ID: 1, UserID: "other", Number: "123", UploadedAt: time.Now()}, nil)

	_, err := svc.AddOrder(ctx, "user1", "123")
	if err != repository.ErrConflict {
		t.Errorf("got err %v", err)
	}
}
