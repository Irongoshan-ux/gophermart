package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgerrcode"

	"gophermart/internal/model"
	"gophermart/internal/repository/sqlc/db"
)

const pgUniqueUsersLogin = "users_login_key"
const pgUniqueOrdersNumber = "orders_number_key"

func uuidFromString(s string) (pgtype.UUID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, err
	}
	var p pgtype.UUID
	copy(p.Bytes[:], u[:])
	p.Valid = true
	return p, nil
}

func uuidToString(p pgtype.UUID) string {
	if !p.Valid {
		return ""
	}
	return uuid.UUID(p.Bytes).String()
}

func float8Ptr(f pgtype.Float8) *float64 {
	if !f.Valid {
		return nil
	}
	return &f.Float64
}

type PostgresUserRepository struct{ q *db.Queries }
type PostgresOrderRepository struct{ q *db.Queries }
type PostgresWithdrawalRepository struct{ q *db.Queries }
type PostgresBalanceRepository struct{ q *db.Queries }

func NewPostgresRepositories(pool *pgxpool.Pool) (
	UserRepository, OrderRepository, WithdrawalRepository, BalanceRepository,
) {
	q := db.New(pool)
	return &PostgresUserRepository{q},
		&PostgresOrderRepository{q},
		&PostgresWithdrawalRepository{q},
		&PostgresBalanceRepository{q}
}

func (r *PostgresUserRepository) Create(ctx context.Context, login, passwordHash string) (*model.User, error) {
	u, err := r.q.CreateUser(ctx, db.CreateUserParams{
		Login:        login,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	})
	if err != nil {
		var e *pgconn.PgError
		if errors.As(err, &e) && e.Code == pgerrcode.UniqueViolation && e.ConstraintName == pgUniqueUsersLogin {
			return nil, ErrConflict
		}
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &model.User{
		ID:           uuidToString(u.ID),
		Login:        u.Login,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
	}, nil
}

func (r *PostgresUserRepository) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	u, err := r.q.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by login: %w", err)
	}
	return &model.User{
		ID:           uuidToString(u.ID),
		Login:        u.Login,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
	}, nil
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	pid, err := uuidFromString(id)
	if err != nil {
		return nil, ErrNotFound
	}
	u, err := r.q.GetUserByID(ctx, pid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &model.User{
		ID:           uuidToString(u.ID),
		Login:        u.Login,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
	}, nil
}

func (r *PostgresOrderRepository) Create(ctx context.Context, userID, number, status string, accrual *float64) (*model.Order, error) {
	uid, err := uuidFromString(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	var acc pgtype.Float8
	if accrual != nil {
		acc.Valid = true
		acc.Float64 = *accrual
	}
	o, err := r.q.CreateOrder(ctx, db.CreateOrderParams{
		UserID:     uid,
		Number:     number,
		Status:     status,
		Accrual:    acc,
		UploadedAt: time.Now(),
	})
	if err != nil {
		var e *pgconn.PgError
		if errors.As(err, &e) && e.Code == pgerrcode.UniqueViolation && e.ConstraintName == pgUniqueOrdersNumber {
			return nil, ErrConflict
		}
		return nil, fmt.Errorf("create order: %w", err)
	}
	return &model.Order{
		ID:         o.ID,
		UserID:     uuidToString(o.UserID),
		Number:     o.Number,
		Status:     model.OrderStatus(o.Status),
		Accrual:    float8Ptr(o.Accrual),
		UploadedAt: o.UploadedAt,
	}, nil
}

func (r *PostgresOrderRepository) GetByNumber(ctx context.Context, number string) (*model.Order, error) {
	o, err := r.q.GetOrderByNumber(ctx, number)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get order by number: %w", err)
	}
	return &model.Order{
		ID:         o.ID,
		UserID:     uuidToString(o.UserID),
		Number:     o.Number,
		Status:     model.OrderStatus(o.Status),
		Accrual:    float8Ptr(o.Accrual),
		UploadedAt: o.UploadedAt,
	}, nil
}

func (r *PostgresOrderRepository) GetByUserID(ctx context.Context, userID string) ([]*model.Order, error) {
	uid, err := uuidFromString(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	list, err := r.q.GetOrdersByUserID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("get orders by user: %w", err)
	}
	out := make([]*model.Order, len(list))
	for i := range list {
		o := &list[i]
		out[i] = &model.Order{
			ID:         o.ID,
			UserID:     uuidToString(o.UserID),
			Number:     o.Number,
			Status:     model.OrderStatus(o.Status),
			Accrual:    float8Ptr(o.Accrual),
			UploadedAt: o.UploadedAt,
		}
	}
	return out, nil
}

func (r *PostgresOrderRepository) GetPendingAccrual(ctx context.Context) ([]*model.Order, error) {
	list, err := r.q.GetOrdersPendingAccrual(ctx)
	if err != nil {
		return nil, fmt.Errorf("get pending orders: %w", err)
	}
	out := make([]*model.Order, len(list))
	for i := range list {
		o := &list[i]
		out[i] = &model.Order{
			ID:         o.ID,
			UserID:     uuidToString(o.UserID),
			Number:     o.Number,
			Status:     model.OrderStatus(o.Status),
			Accrual:    float8Ptr(o.Accrual),
			UploadedAt: o.UploadedAt,
		}
	}
	return out, nil
}

func (r *PostgresOrderRepository) UpdateAccrual(ctx context.Context, orderID int64, status string, accrual *float64) error {
	var acc pgtype.Float8
	if accrual != nil {
		acc.Valid = true
		acc.Float64 = *accrual
	}
	return r.q.UpdateOrderAccrual(ctx, db.UpdateOrderAccrualParams{
		ID: orderID, Status: status, Accrual: acc,
	})
}

func (r *PostgresWithdrawalRepository) Create(ctx context.Context, userID, orderNumber string, sum float64) (*model.Withdrawal, error) {
	uid, err := uuidFromString(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	w, err := r.q.CreateWithdrawal(ctx, db.CreateWithdrawalParams{
		UserID:      uid,
		OrderNumber: orderNumber,
		Sum:         sum,
		ProcessedAt: time.Now(),
	})
	if err != nil {
		return nil, fmt.Errorf("create withdrawal: %w", err)
	}
	return &model.Withdrawal{
		ID:          w.ID,
		UserID:      uuidToString(w.UserID),
		OrderNumber: w.OrderNumber,
		Sum:         w.Sum,
		ProcessedAt: w.ProcessedAt,
	}, nil
}

func (r *PostgresWithdrawalRepository) GetByUserID(ctx context.Context, userID string) ([]*model.Withdrawal, error) {
	uid, err := uuidFromString(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	list, err := r.q.GetWithdrawalsByUserID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("get withdrawals: %w", err)
	}
	out := make([]*model.Withdrawal, len(list))
	for i := range list {
		w := &list[i]
		out[i] = &model.Withdrawal{
			ID:          w.ID,
			UserID:      uuidToString(w.UserID),
			OrderNumber: w.OrderNumber,
			Sum:         w.Sum,
			ProcessedAt: w.ProcessedAt,
		}
	}
	return out, nil
}

func (r *PostgresWithdrawalRepository) GetTotalWithdrawn(ctx context.Context, userID string) (float64, error) {
	uid, err := uuidFromString(userID)
	if err != nil {
		return 0, fmt.Errorf("invalid user id: %w", err)
	}
	return r.q.GetTotalWithdrawnByUserID(ctx, uid)
}

func (r *PostgresBalanceRepository) GetTotalAccrual(ctx context.Context, userID string) (float64, error) {
	uid, err := uuidFromString(userID)
	if err != nil {
		return 0, fmt.Errorf("invalid user id: %w", err)
	}
	return r.q.GetTotalAccrualByUserID(ctx, uid)
}
