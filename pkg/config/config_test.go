package config

import (
	"os"
	"testing"
	"time"
)

func TestNormalizePort(t *testing.T) {
	cases := map[string]string{
		"8080":   "8080",
		":8080":  "8080",
		" 8080 ": "8080",
		"":       "",
	}
	for in, want := range cases {
		if got := NormalizePort(in); got != want {
			t.Errorf("NormalizePort(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFirstNonEmpty(t *testing.T) {
	t.Setenv("X_A", "")
	t.Setenv("X_B", "  ")
	t.Setenv("X_C", "value")
	if got := FirstNonEmpty("X_A", "X_B", "X_C", "X_D"); got != "value" {
		t.Errorf("FirstNonEmpty = %q, want %q", got, "value")
	}
	os.Unsetenv("X_C")
	if got := FirstNonEmpty("X_A", "X_B"); got != "" {
		t.Errorf("FirstNonEmpty empty case = %q, want empty", got)
	}
}

func TestGetEnvAsInt(t *testing.T) {
	t.Setenv("X_INT_OK", "42")
	t.Setenv("X_INT_BAD", "not-a-number")
	if got := GetEnvAsInt("X_INT_OK", 1); got != 42 {
		t.Errorf("ok case = %d", got)
	}
	if got := GetEnvAsInt("X_INT_BAD", 7); got != 7 {
		t.Errorf("bad case fallback = %d, want 7", got)
	}
	if got := GetEnvAsInt("X_INT_MISSING", 9); got != 9 {
		t.Errorf("missing case fallback = %d, want 9", got)
	}
}

func TestGetEnvAsBool(t *testing.T) {
	t.Setenv("X_BOOL_TRUE", "true")
	t.Setenv("X_BOOL_FALSE", "false")
	t.Setenv("X_BOOL_BAD", "yesplease")
	if !GetEnvAsBool("X_BOOL_TRUE", false) {
		t.Error("true case")
	}
	if GetEnvAsBool("X_BOOL_FALSE", true) {
		t.Error("false case")
	}
	if !GetEnvAsBool("X_BOOL_BAD", true) {
		t.Error("bad case fallback")
	}
}

func TestGetEnvAsDuration(t *testing.T) {
	t.Setenv("X_DUR_GO", "5m")
	t.Setenv("X_DUR_SECS", "30")
	t.Setenv("X_DUR_BAD", "garbage")

	if got := GetEnvAsDuration("X_DUR_GO", "1m"); got != 5*time.Minute {
		t.Errorf("go-duration case = %v", got)
	}
	if got := GetEnvAsDuration("X_DUR_SECS", "1m"); got != 30*time.Second {
		t.Errorf("seconds case = %v", got)
	}
	if got := GetEnvAsDuration("X_DUR_BAD", "2m"); got != 2*time.Minute {
		t.Errorf("bad-value default-fallback = %v", got)
	}
	if got := GetEnvAsDuration("X_DUR_MISSING", ""); got != 15*time.Minute {
		t.Errorf("ultimate-fallback = %v", got)
	}
}

func TestLoadServer(t *testing.T) {
	t.Setenv("SERVER_PORT", ":9090")
	t.Setenv("GIN_MODE", "release")
	cfg := LoadServer("8080", "")
	if cfg.Port != "9090" {
		t.Errorf("Port = %q", cfg.Port)
	}
	if cfg.Environment != "release" {
		t.Errorf("Environment = %q", cfg.Environment)
	}
}

func TestLoadServer_Defaults(t *testing.T) {
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("PORT")
	os.Unsetenv("SERVER_MODE")
	os.Unsetenv("GIN_MODE")
	os.Unsetenv("ENABLE_SWAGGER")
	cfg := LoadServer("8083", "")
	if cfg.Port != "8083" {
		t.Errorf("default Port = %q", cfg.Port)
	}
	if cfg.Environment != "debug" {
		t.Errorf("default Environment = %q", cfg.Environment)
	}
	if !cfg.EnableSwagger {
		t.Error("default EnableSwagger should be true")
	}
}

func TestLoadDatabase_Defaults(t *testing.T) {
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("DB_MAX_OPEN_CONNS")
	os.Unsetenv("DB_MAX_IDLE_CONNS")
	cfg := LoadDatabase("postgres://x")
	if cfg.DatabaseURL != "postgres://x" {
		t.Errorf("default DSN = %q", cfg.DatabaseURL)
	}
	if cfg.MaxOpenConns != 25 || cfg.MaxIdleConns != 5 {
		t.Errorf("default pool = %d/%d", cfg.MaxOpenConns, cfg.MaxIdleConns)
	}
}

func TestLoadRedis_Defaults(t *testing.T) {
	os.Unsetenv("REDIS_URL")
	os.Unsetenv("REDIS_PASSWORD")
	os.Unsetenv("REDIS_DB")
	cfg := LoadRedis("")
	if cfg.RedisURL != "localhost:6379" {
		t.Errorf("default Redis URL = %q", cfg.RedisURL)
	}
}
