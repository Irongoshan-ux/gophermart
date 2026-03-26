package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gophermart/internal/auth"
	"gophermart/internal/model"
	handler_mocks "gophermart/internal/handler/mocks"
	"gophermart/internal/repository"
	"gophermart/internal/service"
	"github.com/rs/zerolog"
	"go.uber.org/mock/gomock"
)

func TestHandler_Register_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	users := handler_mocks.NewMockUserService(ctrl)
	orders := handler_mocks.NewMockOrderService(ctrl)
	balance := handler_mocks.NewMockBalanceService(ctrl)
	withdrawals := handler_mocks.NewMockWithdrawalService(ctrl)

	h := NewHandler(users, orders, balance, withdrawals, zerolog.Nop(), "secret")

	users.EXPECT().
		Register(gomock.Any(), "u1", "p1").
		Return(&model.User{ID: "id1", Login: "u1", CreatedAt: time.Now()}, nil)

	body, _ := json.Marshal(map[string]string{"login": "u1", "password": "p1"})
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Register(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status %d", w.Code)
	}
}

func TestHandler_Register_Conflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	users := handler_mocks.NewMockUserService(ctrl)
	h := NewHandler(users, nil, nil, nil, zerolog.Nop(), "secret")

	users.EXPECT().
		Register(gomock.Any(), "u1", "p1").
		Return(nil, repository.ErrConflict)

	body, _ := json.Marshal(map[string]string{"login": "u1", "password": "p1"})
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Register(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("status %d", w.Code)
	}
}

func TestHandler_Login_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	users := handler_mocks.NewMockUserService(ctrl)
	h := NewHandler(users, nil, nil, nil, zerolog.Nop(), "secret")

	users.EXPECT().
		Login(gomock.Any(), "u1", "wrong").
		Return(nil, repository.ErrNotFound)

	body, _ := json.Marshal(map[string]string{"login": "u1", "password": "wrong"})
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status %d", w.Code)
	}
}

func TestHandler_PostOrder_RequiresAuth(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	h := NewHandler(nil, nil, nil, nil, zerolog.Nop(), "secret")
	r := h.Router()

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte("12345678903")))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status %d", w.Code)
	}
}

func TestHandler_PostOrder_ValidLuhn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	orders := handler_mocks.NewMockOrderService(ctrl)
	h := NewHandler(nil, orders, nil, nil, zerolog.Nop(), "secret")
	r := h.Router()

	orders.EXPECT().
		AddOrder(gomock.Any(), "user1", "12345678903").
		Return(false, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte("12345678903")))
	req = req.WithContext(auth.WithUserID(context.Background(), "user1"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("status %d", w.Code)
	}
}

func TestHandler_PostOrder_InvalidLuhn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	h := NewHandler(nil, nil, nil, nil, zerolog.Nop(), "secret")
	r := h.Router()

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewReader([]byte("12345678902")))
	req = req.WithContext(auth.WithUserID(context.Background(), "user1"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status %d", w.Code)
	}
}

func TestHandler_GetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	balance := handler_mocks.NewMockBalanceService(ctrl)
	h := NewHandler(nil, nil, balance, nil, zerolog.Nop(), "secret")
	r := h.Router()

	balance.EXPECT().
		GetBalance(gomock.Any(), "user1").
		Return(service.Balance{Current: 100.5, Withdrawn: 50}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	req = req.WithContext(auth.WithUserID(context.Background(), "user1"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status %d", w.Code)
	}
}

func TestHandler_GetOrders_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	orders := handler_mocks.NewMockOrderService(ctrl)
	h := NewHandler(nil, orders, nil, nil, zerolog.Nop(), "secret")
	r := h.Router()

	orders.EXPECT().GetOrdersByUser(gomock.Any(), "user1").Return(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	req = req.WithContext(auth.WithUserID(context.Background(), "user1"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status %d", w.Code)
	}
}

func TestHandler_Register_BadRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	h := NewHandler(nil, nil, nil, nil, zerolog.Nop(), "secret")
	r := h.Router()

	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status %d", w.Code)
	}
}

func TestHandler_GetWithdrawals_NoContent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	withdrawals := handler_mocks.NewMockWithdrawalService(ctrl)
	h := NewHandler(nil, nil, nil, withdrawals, zerolog.Nop(), "secret")
	r := h.Router()

	withdrawals.EXPECT().GetWithdrawals(gomock.Any(), "user1").Return(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	req = req.WithContext(auth.WithUserID(context.Background(), "user1"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status %d", w.Code)
	}
}

func TestHandler_Withdraw_InsufficientFunds(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	withdrawals := handler_mocks.NewMockWithdrawalService(ctrl)
	h := NewHandler(nil, nil, nil, withdrawals, zerolog.Nop(), "secret")
	r := h.Router()

	withdrawals.EXPECT().
		Withdraw(gomock.Any(), "user1", "2377225624", 751.0).
		Return(service.ErrInsufficientFunds)

	body, _ := json.Marshal(map[string]interface{}{"order": "2377225624", "sum": 751})
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(auth.WithUserID(context.Background(), "user1"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusPaymentRequired {
		t.Errorf("status %d", w.Code)
	}
}
