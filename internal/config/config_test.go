package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Reset flag state
	os.Clearenv()
	for _, k := range []string{"RUN_ADDRESS", "DATABASE_URI", "ACCRUAL_SYSTEM_ADDRESS", "JWT_SECRET"} {
		os.Unsetenv(k)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.RunAddress != "localhost:8080" {
		t.Errorf("RunAddress = %q", cfg.RunAddress)
	}
	if cfg.AccrualSystemAddress != "http://localhost:8081" {
		t.Errorf("AccrualSystemAddress = %q", cfg.AccrualSystemAddress)
	}
}
