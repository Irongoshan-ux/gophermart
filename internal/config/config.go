package config

import (
	"flag"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	RunAddress          string `env:"RUN_ADDRESS"`
	DatabaseURI         string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	JWTSecret           string `env:"JWT_SECRET"`
}

func Load() (*Config, error) {
	var cfg Config
	flag.StringVar(&cfg.RunAddress, "a", "localhost:8080", "HTTP server address (RUN_ADDRESS)")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "PostgreSQL connection URI (DATABASE_URI)")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "http://localhost:8081", "Accrual system base URL (ACCRUAL_SYSTEM_ADDRESS)")
	flag.StringVar(&cfg.JWTSecret, "s", "secret", "JWT signing secret (JWT_SECRET)")
	flag.Parse()
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
