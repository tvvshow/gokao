package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/tvvshow/gokao/services/data-service/internal/config"
	"github.com/tvvshow/gokao/services/data-service/internal/database"
	"github.com/tvvshow/gokao/services/data-service/internal/models"
)

// newUniversityTestDB 与 processing/admission test 共用同款 sqlite 套路，
// 但每个测试拥有独立 :memory: 库（不带 cache=shared）避免跨测试串扰。
// 必须 SetMaxOpenConns(1) —— sqlite ":memory:" 每条物理连接是独立内存空间，
// 池里多个连接会看不到彼此的表。GetUniversityStatistics 的 errgroup 3 路
// 并发查询如果开多连接，第 2/3 个会落到没有 universities 表的空连接。
func newUniversityTestDB(t *testing.T) *database.DB {
	t.Helper()
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		t.Fatalf("get sqlDB: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := gormDB.AutoMigrate(&models.University{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return &database.DB{
		PostgreSQL: gormDB,
		Config:     &config.Config{CacheEnabled: false},
		Logger:     logrus.New(),
	}
}

// TestGetUniversityStatistics_ParallelDimensions 验证 errgroup 并发 3 路
// GROUP BY 结果正确，3 个独立 map 在 -race 下不产生竞争。
// 必须用 go test -race ./... 跑才能检出 builder 共享 race —— 当前实现给
// 每个 goroutine 新建 Session，应通过。
func TestGetUniversityStatistics_ParallelDimensions(t *testing.T) {
	db := newUniversityTestDB(t)
	seedUniversityStats(t, db.PostgreSQL)
	svc := NewUniversityService(db, logrus.New())

	stats, err := svc.GetUniversityStatistics(context.Background())
	if err != nil {
		t.Fatalf("GetUniversityStatistics: %v", err)
	}

	if stats.Total != 6 {
		t.Fatalf("total: want 6, got %d", stats.Total)
	}
	if stats.By985 != 2 {
		t.Fatalf("by_985: want 2, got %d", stats.By985)
	}
	if stats.By211 != 1 {
		t.Fatalf("by_211: want 1, got %d", stats.By211)
	}
	if stats.ByDoubleFirst != 1 {
		t.Fatalf("by_double_first_class: want 1, got %d", stats.ByDoubleFirst)
	}

	if stats.ByProvince["北京"] != 3 || stats.ByProvince["上海"] != 2 || stats.ByProvince["广东"] != 1 {
		t.Fatalf("by_province wrong: %+v", stats.ByProvince)
	}
	if stats.ByType["undergraduate"] != 5 || stats.ByType["associate"] != 1 {
		t.Fatalf("by_type wrong: %+v", stats.ByType)
	}
	if stats.ByNature["public"] != 4 || stats.ByNature["private"] != 2 {
		t.Fatalf("by_nature wrong: %+v", stats.ByNature)
	}
}

// TestGetUniversityStatistics_EmptyTable 验证空库下三路 GROUP BY 返回空 map 而不报错。
func TestGetUniversityStatistics_EmptyTable(t *testing.T) {
	db := newUniversityTestDB(t)
	svc := NewUniversityService(db, logrus.New())

	stats, err := svc.GetUniversityStatistics(context.Background())
	if err != nil {
		t.Fatalf("empty table: %v", err)
	}
	if stats.Total != 0 {
		t.Fatalf("empty total: want 0, got %d", stats.Total)
	}
	if len(stats.ByProvince) != 0 || len(stats.ByType) != 0 || len(stats.ByNature) != 0 {
		t.Fatalf("empty buckets expected: %+v", stats)
	}
}

func seedUniversityStats(t *testing.T, db *gorm.DB) {
	t.Helper()
	rows := []models.University{
		{ID: uuid.New(), Code: "10001", Name: "北京大学", Province: "北京", Type: "undergraduate", Nature: "public", Level: "985"},
		{ID: uuid.New(), Code: "10002", Name: "清华大学", Province: "北京", Type: "undergraduate", Nature: "public", Level: "985"},
		{ID: uuid.New(), Code: "10003", Name: "北京交通大学", Province: "北京", Type: "undergraduate", Nature: "public", Level: "211"},
		{ID: uuid.New(), Code: "10248", Name: "上海交通大学", Province: "上海", Type: "undergraduate", Nature: "public", Level: "double_first_class"},
		{ID: uuid.New(), Code: "10271", Name: "上海外国语大学", Province: "上海", Type: "undergraduate", Nature: "private", Level: ""},
		{ID: uuid.New(), Code: "12121", Name: "广州番禺职业技术学院", Province: "广东", Type: "associate", Nature: "private", Level: ""},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}
}
