package config

import (
	"os"
	"testing"
)

func TestLoadSupportsServerEnvAliases(t *testing.T) {
	os.Setenv("SERVER_PORT", ":8080")
	os.Setenv("SERVER_MODE", "release")
	defer func() {
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("SERVER_MODE")
	}()

	cfg := Load()
	if cfg.Port != "8080" {
		t.Fatalf("expected normalized port 8080, got %s", cfg.Port)
	}
	if cfg.Environment != "release" {
		t.Fatalf("expected release, got %s", cfg.Environment)
	}
}

func TestLoadFallsBackToLegacyEnv(t *testing.T) {
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("SERVER_MODE")
	os.Setenv("PORT", "9090")
	os.Setenv("GIN_MODE", "debug")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("GIN_MODE")
	}()

	cfg := Load()
	if cfg.Port != "9090" {
		t.Fatalf("expected port 9090, got %s", cfg.Port)
	}
	if cfg.Environment != "debug" {
		t.Fatalf("expected debug, got %s", cfg.Environment)
	}
}

func TestLoadUsesUnifiedDefaultPort(t *testing.T) {
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("SERVER_MODE")
	os.Unsetenv("PORT")
	os.Unsetenv("GIN_MODE")

	cfg := Load()
	if cfg.Port != "8083" {
		t.Fatalf("expected default port 8083, got %s", cfg.Port)
	}
	if cfg.Environment != "debug" {
		t.Fatalf("expected default mode debug, got %s", cfg.Environment)
	}
}
