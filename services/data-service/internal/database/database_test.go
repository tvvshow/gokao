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

	body, err := fs.ReadFile(embedMigrations, "migrations/00001_init.sql")
	if err != nil {
		t.Fatalf("read 00001_init.sql: %v", err)
	}
	content := string(body)

	// 9 张核心表必须全部出现在 schema clauses 中。
	wantTables := []string{
		"CREATE TABLE IF NOT EXISTS universities",
		"CREATE TABLE IF NOT EXISTS majors",
		"CREATE TABLE IF NOT EXISTS admission_data",
		"CREATE TABLE IF NOT EXISTS search_indices",
		"CREATE TABLE IF NOT EXISTS analysis_results",
		"CREATE TABLE IF NOT EXISTS hot_searches",
		"CREATE TABLE IF NOT EXISTS data_statistics",
		"CREATE TABLE IF NOT EXISTS university_statistics",
		"CREATE TABLE IF NOT EXISTS major_statistics",
	}
	for _, w := range wantTables {
		if !strings.Contains(content, w) {
			t.Errorf("missing schema clause: %s", w)
		}
	}

	// 扩展：pgcrypto (gen_random_uuid) 与 pg_trgm（搜索 GIN 索引）
	wantExtensions := []string{
		"CREATE EXTENSION IF NOT EXISTS pgcrypto",
		"CREATE EXTENSION IF NOT EXISTS pg_trgm",
	}
	for _, w := range wantExtensions {
		if !strings.Contains(content, w) {
			t.Errorf("missing extension clause: %s", w)
		}
	}

	// pg_trgm GIN 表达式索引：服务层 LOWER(col) LIKE '%kw%' 的关键索引，缺失则搜索退化。
	wantTrgmIndexes := []string{
		"idx_universities_name_trgm",
		"idx_universities_code_trgm",
		"idx_universities_alias_trgm",
		"idx_majors_name_trgm",
		"idx_hot_searches_keyword_trgm",
	}
	for _, w := range wantTrgmIndexes {
		if !strings.Contains(content, w) {
			t.Errorf("missing trgm index: %s", w)
		}
	}

	// 旧 createIndices() 的两处 column 拼错（universities.popularity_score、
	// admission_data.batch_type）必须修正：新 SQL 不能再出现非法列名。
	if strings.Contains(content, "universities(national_rank, popularity_score)") {
		t.Error("bug regression: universities has no popularity_score column")
	}
	if strings.Contains(content, "admission_data(year, batch_type)") {
		t.Error("bug regression: admission_data has no batch_type column")
	}

	// goose 标记必须配对：Up + Down 两段
	if !strings.Contains(content, "-- +goose Up") {
		t.Error("missing -- +goose Up marker")
	}
	if !strings.Contains(content, "-- +goose Down") {
		t.Error("missing -- +goose Down marker — rollback path broken")
	}

	// popularity_score 默认 seed
	if !strings.Contains(content, "popularity_score = 95") {
		t.Error("missing popularity seed: 95")
	}
	if !strings.Contains(content, "popularity_score = 70 WHERE popularity_score = 0") {
		t.Error("missing default fallback popularity seed")
	}
}
