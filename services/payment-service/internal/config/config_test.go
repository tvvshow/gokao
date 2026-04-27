package config

import (
	"os"
	"testing"
)

func TestLoadSupportsServerAliases(t *testing.T) {
	os.Setenv("SERVER_PORT", "8085")
	os.Setenv("SERVER_MODE", "release")
	defer func() {
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("SERVER_MODE")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if cfg.Server.Port != 8085 {
		t.Fatalf("expected 8085, got %d", cfg.Server.Port)
	}
	if cfg.Server.Mode != "release" {
		t.Fatalf("expected release, got %s", cfg.Server.Mode)
	}
}

func TestLoadUsesUnifiedDefaultPort(t *testing.T) {
	oldServerPort := os.Getenv("SERVER_PORT")
	oldPort := os.Getenv("PORT")
	oldMode := os.Getenv("SERVER_MODE")
	oldGinMode := os.Getenv("GIN_MODE")
	defer func() {
		_ = os.Setenv("SERVER_PORT", oldServerPort)
		_ = os.Setenv("PORT", oldPort)
		_ = os.Setenv("SERVER_MODE", oldMode)
		_ = os.Setenv("GIN_MODE", oldGinMode)
	}()

	_ = os.Unsetenv("SERVER_PORT")
	_ = os.Unsetenv("PORT")
	_ = os.Unsetenv("SERVER_MODE")
	_ = os.Unsetenv("GIN_MODE")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if cfg.Server.Port != 8085 {
		t.Fatalf("expected default port 8085, got %d", cfg.Server.Port)
	}
	if cfg.Server.Mode != "debug" {
		t.Fatalf("expected default mode debug, got %s", cfg.Server.Mode)
	}
}

func TestLoadSupportsDatabasePoolConfig(t *testing.T) {
	os.Setenv("DB_MAX_OPEN_CONNS", "40")
	os.Setenv("DB_MAX_IDLE_CONNS", "12")
	os.Setenv("DB_CONN_MAX_LIFETIME", "3600")
	os.Setenv("DB_CONN_MAX_IDLE_TIME", "1200")
	defer func() {
		os.Unsetenv("DB_MAX_OPEN_CONNS")
		os.Unsetenv("DB_MAX_IDLE_CONNS")
		os.Unsetenv("DB_CONN_MAX_LIFETIME")
		os.Unsetenv("DB_CONN_MAX_IDLE_TIME")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if cfg.Database.MaxOpenConns != 40 {
		t.Fatalf("expected MaxOpenConns 40, got %d", cfg.Database.MaxOpenConns)
	}
	if cfg.Database.MaxIdleConns != 12 {
		t.Fatalf("expected MaxIdleConns 12, got %d", cfg.Database.MaxIdleConns)
	}
	if cfg.Database.ConnMaxLifetime != 3600 {
		t.Fatalf("expected ConnMaxLifetime 3600, got %d", cfg.Database.ConnMaxLifetime)
	}
	if cfg.Database.ConnMaxIdleTime != 1200 {
		t.Fatalf("expected ConnMaxIdleTime 1200, got %d", cfg.Database.ConnMaxIdleTime)
	}
}
