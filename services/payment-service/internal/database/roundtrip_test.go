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
// 与 TestEmbedMigrationsPresent（结构断言）的关系：前者跑 SQL 语法 + 索引创建路径，
// 后者只校 .sql 文件内容。两者互补。
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

	// 起始清场。Reset 在无版本表时返回 "no migrations" 类错误，可忽略。
	_ = goose.Reset(db, "migrations")

	if err := goose.Up(db, "migrations"); err != nil {
		t.Fatalf("goose.Up failed — schema SQL has a real bug: %v", err)
	}

	v, err := goose.GetDBVersion(db)
	if err != nil {
		t.Fatalf("GetDBVersion after Up: %v", err)
	}
	if v == 0 {
		t.Fatal("expected version > 0 after Up, got 0 (no migrations applied)")
	}
	t.Logf("payment-service: migrated up to version %d", v)

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
