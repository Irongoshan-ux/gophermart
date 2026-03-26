package service

import (
	"context"

	"gophermart/internal/accrual"
	"gophermart/internal/model"
	"gophermart/internal/repository"
)

type AccrualFetcher interface {
	GetOrder(ctx context.Context, number string) (*accrual.OrderInfo, error)
}

type OrderService struct {
	orders  repository.OrderRepository
	accrual AccrualFetcher
}

func NewOrderService(orders repository.OrderRepository, accrualClient AccrualFetcher) *OrderService {
	return &OrderService{orders: orders, accrual: accrualClient}
}

func mapAccrualStatus(accrualStatus string) (string, bool) {
	switch accrualStatus {
	case "PROCESSED":
		return "PROCESSED", true
	case "INVALID":
		return "INVALID", true
	case "PROCESSING", "REGISTERED":
		return "PROCESSING", true
	default:
		return "", false
	}
}

func (s *OrderService) fetchAndUpdateAccrual(ctx context.Context, order *model.Order) {
	if s.accrual == nil || (order.Status != model.OrderStatusNew && order.Status != model.OrderStatusProcessing) {
		return
	}
	info, err := s.accrual.GetOrder(ctx, order.Number)
	if err != nil || info == nil {
		return
	}
	status, ok := mapAccrualStatus(info.Status)
	if !ok {
		return
	}
	_ = s.orders.UpdateAccrual(ctx, order.ID, status, info.Accrual)
}

func (s *OrderService) AddOrder(ctx context.Context, userID, number string) (alreadyUploaded bool, err error) {
	existing, err := s.orders.GetByNumber(ctx, number)
	if err == nil {
		if existing.UserID == userID {
			return true, nil
		}
		return false, repository.ErrConflict
	}
	if err != repository.ErrNotFound {
		return false, err
	}
	order, err := s.orders.Create(ctx, userID, number, string(model.OrderStatusNew), nil)
	if err != nil {
		return false, err
	}
	s.fetchAndUpdateAccrual(ctx, order)
	return false, nil
}

func (s *OrderService) GetOrdersByUser(ctx context.Context, userID string) ([]*model.Order, error) {
	list, err := s.orders.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, o := range list {
		s.fetchAndUpdateAccrual(ctx, o)
	}
	return s.orders.GetByUserID(ctx, userID)
}
