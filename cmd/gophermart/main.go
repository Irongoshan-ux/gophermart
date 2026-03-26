package main

import (
	"context"
	"net/http"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"gophermart/internal/accrual"
	"gophermart/internal/app"
	"gophermart/internal/config"
	"gophermart/internal/repository"
	"gophermart/internal/service"
)

func runMigrations(uri string, migrationsPath string) error {
	m, err := migrate.New("file://"+filepath.ToSlash(migrationsPath), uri)
	if err != nil {
		return err
	}
	defer m.Close()
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func main() {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("load config")
	}

	if cfg.DatabaseURI == "" {
		logger.Fatal().Msg("DATABASE_URI is required")
	}

	migrationsPath := "migrations"
	if p := os.Getenv("MIGRATIONS_PATH"); p != "" {
		migrationsPath = p
	}
	if err := runMigrations(cfg.DatabaseURI, migrationsPath); err != nil {
		logger.Fatal().Err(err).Msg("migrations")
	}

	pool, err := pgxpool.New(ctx, cfg.DatabaseURI)
	if err != nil {
		logger.Fatal().Err(err).Msg("pgxpool")
	}
	defer pool.Close()

	userRepo, orderRepo, withdrawalRepo, balanceRepo := repository.NewPostgresRepositories(pool)

	userSvc := service.NewUserService(userRepo)
	accrualClient := accrual.NewClient(cfg.AccrualSystemAddress)
	orderSvc := service.NewOrderService(orderRepo, accrualClient)
	balanceSvc := service.NewBalanceService(balanceRepo, withdrawalRepo)
	withdrawalSvc := service.NewWithdrawalService(balanceRepo, withdrawalRepo)

	handler := app.NewServer(userSvc, orderSvc, balanceSvc, withdrawalSvc, logger, cfg.JWTSecret)
	server := &http.Server{Addr: cfg.RunAddress, Handler: handler}

	logger.Info().Str("address", cfg.RunAddress).Msg("gophermart starting")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal().Err(err).Msg("server")
	}
}
