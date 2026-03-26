package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"gophermart/internal/repository"
	"gophermart/internal/repository/mocks"
	"gophermart/internal/service"
	"github.com/rs/zerolog"
	"go.uber.org/mock/gomock"
)

func TestNewServer_ServesRoutes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	userRepo := repository_mocks.NewMockUserRepository(ctrl)
	orderRepo := repository_mocks.NewMockOrderRepository(ctrl)
	withRepo := repository_mocks.NewMockWithdrawalRepository(ctrl)
	balRepo := repository_mocks.NewMockBalanceRepository(ctrl)

	userSvc := service.NewUserService(userRepo)
	orderSvc := service.NewOrderService(orderRepo, nil)
	balSvc := service.NewBalanceService(balRepo, withRepo)
	withSvc := service.NewWithdrawalService(balRepo, withRepo)

	handler := NewServer(userSvc, orderSvc, balSvc, withSvc, zerolog.Nop(), "secret")
	srv := httptest.NewServer(handler)
	defer srv.Close()

	// Unauthorized request to protected route
	resp, err := http.Get(srv.URL + "/api/user/balance")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("GET /api/user/balance without auth: status %d", resp.StatusCode)
	}

	// Register returns 400 for invalid body
	resp2, err := http.Post(srv.URL+"/api/user/register", "application/json", nil)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	resp2.Body.Close()
	if resp2.StatusCode != http.StatusBadRequest {
		t.Errorf("POST /api/user/register empty body: status %d", resp2.StatusCode)
	}
}

func TestNewServer_Register_Conflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	userRepo := repository_mocks.NewMockUserRepository(ctrl)
	userRepo.EXPECT().
		Create(gomock.Any(), "u", gomock.Any()).
		Return(nil, repository.ErrConflict)

	userSvc := service.NewUserService(userRepo)
	orderRepo := repository_mocks.NewMockOrderRepository(ctrl)
	withRepo := repository_mocks.NewMockWithdrawalRepository(ctrl)
	balRepo := repository_mocks.NewMockBalanceRepository(ctrl)
	orderSvc := service.NewOrderService(orderRepo, nil)
	balSvc := service.NewBalanceService(balRepo, withRepo)
	withSvc := service.NewWithdrawalService(balRepo, withRepo)

	handler := NewServer(userSvc, orderSvc, balSvc, withSvc, zerolog.Nop(), "secret")
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader([]byte(`{"login":"u","password":"p"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Errorf("status %d", w.Code)
	}
}
