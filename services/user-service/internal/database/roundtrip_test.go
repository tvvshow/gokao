//go:build migration_roundtrip

// migration_roundtrip 标签下的测试需要 postgres 实例（CI 在 services.postgres 容器侧提供）。
// 默认 `go test ./...` 不会触发，本地无须依赖外部 DB。
package database

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

// TestMigrationRoundTrip 真 SQL 验证：goose Up 到最新 → Reset 回 0 → 校验版本号在两端正确。
// user-service 有 00001_init.sql（schema）+ 00002_seed.sql（默认权限/角色），两段都跑。
func TestMigrationRoundTrip(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set — CI postgres container should export it")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("db.Ping: %v", err)
	}

	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("goose.SetDialect: %v", err)
	}

	_ = goose.Reset(db, "migrations")

	if err := goose.Up(db, "migrations"); err != nil {
		t.Fatalf("goose.Up failed — schema/seed SQL has a real bug: %v", err)
	}

	v, err := goose.GetDBVersion(db)
	if err != nil {
		t.Fatalf("GetDBVersion after Up: %v", err)
	}
	if v < 2 {
		t.Fatalf("expected version >= 2 (init + seed), got %d", v)
	}
	t.Logf("user-service: migrated up to version %d", v)

	// 验证 seed 实际写入：6 角色 + 28 权限 + 角色权限映射 >0 行
	for _, q := range []struct {
		name    string
		query   string
		atLeast int
	}{
		{"roles", "SELECT COUNT(*) FROM roles WHERE is_system = true", 5},
		{"permissions", "SELECT COUNT(*) FROM permissions", 27},
		{"role_permissions", "SELECT COUNT(*) FROM role_permissions", 1},
	} {
		var n int
		if err := db.QueryRow(q.query).Scan(&n); err != nil {
			t.Fatalf("count %s: %v", q.name, err)
		}
		if n < q.atLeast {
			t.Errorf("seed row count for %s = %d, want >= %d", q.name, n, q.atLeast)
		}
	}

	if err := goose.Reset(db, "migrations"); err != nil {
		t.Fatalf("goose.Reset failed — Down path has a real bug: %v", err)
	}

	v2, err := goose.GetDBVersion(db)
	if err != nil {
		t.Fatalf("GetDBVersion after Reset: %v", err)
	}
	if v2 != 0 {
		t.Fatalf("expected version 0 after Reset, got %d", v2)
	}
}
