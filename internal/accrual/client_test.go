package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GetOrder_200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/orders/123" {
			t.Errorf("path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(OrderInfo{Order: "123", Status: "PROCESSED", Accrual: float64Ptr(500)})
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	info, err := client.GetOrder(context.Background(), "123")
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if info == nil || info.Order != "123" || info.Status != "PROCESSED" || info.Accrual == nil || *info.Accrual != 500 {
		t.Errorf("got %+v", info)
	}
}

func TestClient_GetOrder_204(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	info, err := client.GetOrder(context.Background(), "999")
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if info != nil {
		t.Errorf("expected nil, got %+v", info)
	}
}

func TestClient_GetOrder_429(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	_, err := client.GetOrder(context.Background(), "123")
	var rl *RateLimitError
	if err == nil || !errors.As(err, &rl) {
		t.Errorf("expected RateLimitError, got %v", err)
	}
	if rl != nil && rl.RetryAfter != 30 {
		t.Errorf("RetryAfter = %d", rl.RetryAfter)
	}
}

func float64Ptr(f float64) *float64 { return &f }
