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
