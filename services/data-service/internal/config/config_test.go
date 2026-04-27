package config

import (
	"os"
	"testing"
)

func TestLoadSupportsServerAliases(t *testing.T) {
	os.Setenv("SERVER_PORT", ":8082")
	os.Setenv("SERVER_MODE", "release")
	defer func() {
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("SERVER_MODE")
	}()
	cfg := Load()
	if cfg.Port != "8082" {
		t.Fatalf("expected 8082, got %s", cfg.Port)
	}
	if cfg.Environment != "release" {
		t.Fatalf("expected release, got %s", cfg.Environment)
	}
}

func TestLoadUsesUnifiedDefaultPort(t *testing.T) {
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("SERVER_MODE")
	os.Unsetenv("PORT")
	os.Unsetenv("GIN_MODE")

	cfg := Load()
	if cfg.Port != "8082" {
		t.Fatalf("expected default port 8082, got %s", cfg.Port)
	}
	if cfg.Environment != "debug" {
		t.Fatalf("expected default mode debug, got %s", cfg.Environment)
	}
}
