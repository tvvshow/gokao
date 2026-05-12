package database

import (
	"io/fs"
	"strings"
	"testing"
)

// TestEmbedMigrationsPresent 验证 baseline migration 文件被 embed 进二进制，
// 且包含预期的核心 schema 关键词。真 SQL round-trip 在 CI 跑 postgres 容器侧验证。
func TestEmbedMigrationsPresent(t *testing.T) {
	entries, err := fs.ReadDir(embedMigrations, "migrations")
	if err != nil {
		t.Fatalf("read embedded migrations: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("no migration files embedded — embed directive likely broken")
	}

	// 至少必须存在 baseline init
	var found00001 bool
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "00001_") && strings.HasSuffix(e.Name(), ".sql") {
			found00001 = true
		}
	}
	if !found00001 {
		t.Fatalf("baseline 00001_*.sql not embedded; entries: %v", entries)
	}

	// 内容关键词校验：6 张核心表 + 默认套餐 seed
	body, err := fs.ReadFile(embedMigrations, "migrations/00001_init.sql")
	if err != nil {
		t.Fatalf("read 00001_init.sql: %v", err)
	}
	content := string(body)

	wantTables := []string{
		"CREATE TABLE IF NOT EXISTS payment_orders",
		"CREATE TABLE IF NOT EXISTS refund_records",
		"CREATE TABLE IF NOT EXISTS membership_plans",
		"CREATE TABLE IF NOT EXISTS user_memberships",
		"CREATE TABLE IF NOT EXISTS payment_callbacks",
		"CREATE TABLE IF NOT EXISTS license_info",
	}
	for _, w := range wantTables {
		if !strings.Contains(content, w) {
			t.Errorf("missing schema clause: %s", w)
		}
	}

	// goose 标记必须配对：Up + Down 两段
	if !strings.Contains(content, "-- +goose Up") {
		t.Error("missing -- +goose Up marker")
	}
	if !strings.Contains(content, "-- +goose Down") {
		t.Error("missing -- +goose Down marker — rollback path broken")
	}

	// 默认套餐 seed
	for _, plan := range []string{"'basic'", "'premium'", "'ultimate'"} {
		if !strings.Contains(content, plan) {
			t.Errorf("missing default plan seed: %s", plan)
		}
	}
}
