package database

import (
	"io/fs"
	"strings"
	"testing"
)

// TestEmbedMigrationsPresent 校验 baseline migration 文件被 embed 进二进制，
// 且包含所有期望的 schema 关键词。真 SQL round-trip 在 CI postgres 容器侧 goose up/down 验证。
func TestEmbedMigrationsPresent(t *testing.T) {
	entries, err := fs.ReadDir(embedMigrations, "migrations")
	if err != nil {
		t.Fatalf("read embedded migrations: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("no migration files embedded — embed directive likely broken")
	}

	// 至少必须存在 baseline init + seed 两份
	var found00001, found00002 bool
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "00001_") && strings.HasSuffix(e.Name(), ".sql") {
			found00001 = true
		}
		if strings.HasPrefix(e.Name(), "00002_") && strings.HasSuffix(e.Name(), ".sql") {
			found00002 = true
		}
	}
	if !found00001 {
		t.Errorf("baseline 00001_*.sql not embedded; entries: %v", entries)
	}
	if !found00002 {
		t.Errorf("seed 00002_*.sql not embedded; entries: %v", entries)
	}

	body, err := fs.ReadFile(embedMigrations, "migrations/00001_init.sql")
	if err != nil {
		t.Fatalf("read 00001_init.sql: %v", err)
	}
	initSQL := string(body)

	// 12 张核心表必须全部出现在 schema clauses 中。
	wantTables := []string{
		"CREATE TABLE IF NOT EXISTS users",
		"CREATE TABLE IF NOT EXISTS roles",
		"CREATE TABLE IF NOT EXISTS permissions",
		"CREATE TABLE IF NOT EXISTS user_roles",
		"CREATE TABLE IF NOT EXISTS role_permissions",
		"CREATE TABLE IF NOT EXISTS login_attempts",
		"CREATE TABLE IF NOT EXISTS audit_logs",
		"CREATE TABLE IF NOT EXISTS refresh_tokens",
		"CREATE TABLE IF NOT EXISTS device_fingerprints",
		"CREATE TABLE IF NOT EXISTS device_licenses",
		"CREATE TABLE IF NOT EXISTS membership_orders",
		"CREATE TABLE IF NOT EXISTS user_sessions",
	}
	for _, w := range wantTables {
		if !strings.Contains(initSQL, w) {
			t.Errorf("missing schema clause: %s", w)
		}
	}

	// 扩展：pgcrypto 必备（gen_random_uuid）；pg_trgm 保留接入点
	wantExtensions := []string{
		"CREATE EXTENSION IF NOT EXISTS pgcrypto",
		"CREATE EXTENSION IF NOT EXISTS pg_trgm",
	}
	for _, w := range wantExtensions {
		if !strings.Contains(initSQL, w) {
			t.Errorf("missing extension clause: %s", w)
		}
	}

	// UNIQUE 约束关键列：username/email/token 任一失守都直接破坏鉴权契约
	wantUniques := []string{
		"users_username_key UNIQUE (username)",
		"users_email_key    UNIQUE (email)",
		"refresh_tokens_token_key UNIQUE (token)",
		"device_fingerprints_device_id_key UNIQUE (device_id)",
		"membership_orders_order_no_key UNIQUE (order_no)",
		"user_sessions_session_token_key UNIQUE (session_token)",
	}
	for _, w := range wantUniques {
		if !strings.Contains(initSQL, w) {
			t.Errorf("missing UNIQUE constraint: %s", w)
		}
	}

	// goose 标记必须配对：Up + Down
	if !strings.Contains(initSQL, "-- +goose Up") {
		t.Error("missing -- +goose Up marker in 00001")
	}
	if !strings.Contains(initSQL, "-- +goose Down") {
		t.Error("missing -- +goose Down marker in 00001 — rollback path broken")
	}

	// seed 验证：6 角色 + 28 权限的关键 INSERT
	seedBody, err := fs.ReadFile(embedMigrations, "migrations/00002_seed.sql")
	if err != nil {
		t.Fatalf("read 00002_seed.sql: %v", err)
	}
	seedSQL := string(seedBody)

	wantRoles := []string{"'admin'", "'user'", "'basic'", "'premium'", "'enterprise'", "'moderator'"}
	for _, w := range wantRoles {
		if !strings.Contains(seedSQL, w) {
			t.Errorf("missing default role seed: %s", w)
		}
	}

	wantPerms := []string{"'admin:all'", "'user:read'", "'membership:upgrade'", "'audit:export'"}
	for _, w := range wantPerms {
		if !strings.Contains(seedSQL, w) {
			t.Errorf("missing default permission seed: %s", w)
		}
	}

	// 关键 idempotency 保证：所有 INSERT 必须带 ON CONFLICT DO NOTHING，否则二次启动会 PK 冲突
	if !strings.Contains(seedSQL, "ON CONFLICT (name) DO NOTHING") {
		t.Error("seed missing ON CONFLICT (name) DO NOTHING — re-run will crash on duplicate keys")
	}
	if !strings.Contains(seedSQL, "ON CONFLICT (role_id, permission_id) DO NOTHING") {
		t.Error("seed missing ON CONFLICT (role_id, permission_id) DO NOTHING — role-perm re-run will crash")
	}

	if !strings.Contains(seedSQL, "-- +goose Up") {
		t.Error("missing -- +goose Up marker in 00002")
	}
	if !strings.Contains(seedSQL, "-- +goose Down") {
		t.Error("missing -- +goose Down marker in 00002")
	}
}
