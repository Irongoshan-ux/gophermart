package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"gophermart/internal/auth"
	"gophermart/internal/repository"
	"gophermart/internal/service"
	"gophermart/internal/validation"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

type Handler struct {
	users      UserService
	orders     OrderService
	balance    BalanceService
	withdrawals WithdrawalService
	log        zerolog.Logger
	jwtSecret  string
}

func NewHandler(
	users UserService,
	orders OrderService,
	balance BalanceService,
	withdrawals WithdrawalService,
	log zerolog.Logger,
	jwtSecret string,
) *Handler {
	return &Handler{
		users:      users,
		orders:     orders,
		balance:    balance,
		withdrawals: withdrawals,
		log:        log,
		jwtSecret:  jwtSecret,
	}
}

func (h *Handler) Router() chi.Router {
	r := chi.NewRouter()

	r.Post("/api/user/register", h.Register)
	r.Post("/api/user/login", h.Login)

	r.Group(func(r chi.Router) {
		r.Use(requireAuth)
		r.Post("/api/user/orders", h.PostOrder)
		r.Get("/api/user/orders", h.GetOrders)
		r.Get("/api/user/balance", h.GetBalance)
		r.Post("/api/user/balance/withdraw", h.Withdraw)
		r.Get("/api/user/withdrawals", h.GetWithdrawals)
	})

	return r
}

func requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := auth.UserIDFromContext(r.Context()); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	req.Login = strings.TrimSpace(req.Login)
	if req.Login == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user, err := h.users.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		if err == repository.ErrConflict {
			w.WriteHeader(http.StatusConflict)
			return
		}
		h.log.Error().Err(err).Msg("register failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	token, err := auth.NewToken(h.jwtSecret, user.ID)
	if err != nil {
		h.log.Error().Err(err).Msg("create token")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	auth.SetTokenCookie(w, token, 3600*24*30) // 30 days
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	req.Login = strings.TrimSpace(req.Login)
	if req.Login == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user, err := h.users.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		if err == repository.ErrNotFound {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		h.log.Error().Err(err).Msg("login failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	token, err := auth.NewToken(h.jwtSecret, user.ID)
	if err != nil {
		h.log.Error().Err(err).Msg("create token")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	auth.SetTokenCookie(w, token, 3600*24*30)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) PostOrder(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	number := strings.TrimSpace(string(body))
	if number == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !validation.LuhnValid(number) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	already, err := h.orders.AddOrder(r.Context(), userID, number)
	if err != nil {
		if err == repository.ErrConflict {
			w.WriteHeader(http.StatusConflict)
			return
		}
		h.log.Error().Err(err).Msg("add order failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if already {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	orders, err := h.orders.GetOrdersByUser(r.Context(), userID)
	if err != nil {
		h.log.Error().Err(err).Msg("get orders failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	type item struct {
		Number     string   `json:"number"`
		Status     string   `json:"status"`
		Accrual    *float64 `json:"accrual,omitempty"`
		UploadedAt string   `json:"uploaded_at"`
	}
	resp := make([]item, len(orders))
	for i, o := range orders {
		resp[i] = item{
			Number:     o.Number,
			Status:     string(o.Status),
			Accrual:    o.Accrual,
			UploadedAt: o.UploadedAt.Format("2006-01-02T15:04:05-07:00"),
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	bal, err := h.balance.GetBalance(r.Context(), userID)
	if err != nil {
		h.log.Error().Err(err).Msg("get balance failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]float64{
		"current":   bal.Current,
		"withdrawn": bal.Withdrawn,
	})
}

func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	var req struct {
		Order string  `json:"order"`
		Sum   float64 `json:"sum"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	req.Order = strings.TrimSpace(req.Order)
	if req.Order == "" || req.Sum <= 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	if !validation.LuhnValid(req.Order) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	err := h.withdrawals.Withdraw(r.Context(), userID, req.Order, req.Sum)
	if err != nil {
		if err == service.ErrInsufficientFunds {
			w.WriteHeader(http.StatusPaymentRequired)
			return
		}
		h.log.Error().Err(err).Msg("withdraw failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	list, err := h.withdrawals.GetWithdrawals(r.Context(), userID)
	if err != nil {
		h.log.Error().Err(err).Msg("get withdrawals failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(list) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	type item struct {
		Order       string  `json:"order"`
		Sum         float64 `json:"sum"`
		ProcessedAt string  `json:"processed_at"`
	}
	resp := make([]item, len(list))
	for i, wd := range list {
		resp[i] = item{
			Order:       wd.OrderNumber,
			Sum:         wd.Sum,
			ProcessedAt: wd.ProcessedAt.Format("2006-01-02T15:04:05-07:00"),
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
