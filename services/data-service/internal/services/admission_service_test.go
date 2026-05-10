package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/tvvshow/gokao/services/data-service/internal/config"
	"github.com/tvvshow/gokao/services/data-service/internal/database"
	"github.com/tvvshow/gokao/services/data-service/internal/models"
)

// newAdmissionTestDB 构造一个基于 sqlite in-memory 的 *database.DB，
// 仅满足 AdmissionService 静态依赖（PostgreSQL + Config + Logger），不连真实 PG / Redis / ES。
// 用 `:memory:`（不带 cache=shared）以保证各测试拥有独立内存库。
func newAdmissionTestDB(t *testing.T) *database.DB {
	t.Helper()
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := gormDB.AutoMigrate(&models.AdmissionData{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return &database.DB{
		PostgreSQL: gormDB,
		Config: &config.Config{
			CacheEnabled: false,
		},
		Logger: logrus.New(),
	}
}

// seedAdmissionFixtures 写入跨省份、跨批次、跨分数的样本，覆盖 4 个聚合统计维度。
func seedAdmissionFixtures(t *testing.T, db *gorm.DB) {
	t.Helper()
	now := time.Now()
	uniID := uuid.New()
	majorID := uuid.New()
	rows := []models.AdmissionData{
		{ID: uuid.New(), UniversityID: uniID, MajorID: &majorID, Year: 2024, Province: "北京", Batch: "first_batch", Category: "science", MinScore: 600, MaxScore: 680, AvgScore: 640, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.New(), UniversityID: uniID, MajorID: &majorID, Year: 2024, Province: "北京", Batch: "first_batch", Category: "science", MinScore: 590, MaxScore: 670, AvgScore: 630, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.New(), UniversityID: uniID, MajorID: &majorID, Year: 2024, Province: "上海", Batch: "first_batch", Category: "science", MinScore: 610, MaxScore: 690, AvgScore: 650, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.New(), UniversityID: uniID, MajorID: &majorID, Year: 2024, Province: "广东", Batch: "second_batch", Category: "science", MinScore: 500, MaxScore: 580, AvgScore: 540, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.New(), UniversityID: uniID, MajorID: &majorID, Year: 2024, Province: "广东", Batch: "second_batch", Category: "science", MinScore: 510, MaxScore: 590, AvgScore: 550, CreatedAt: now, UpdatedAt: now},
		// 异年份样本，验证 year 过滤
		{ID: uuid.New(), UniversityID: uniID, MajorID: &majorID, Year: 2023, Province: "北京", Batch: "first_batch", Category: "science", MinScore: 580, MaxScore: 660, AvgScore: 620, CreatedAt: now, UpdatedAt: now},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}
}

// TestGetAdmissionStatistics_NoBuilderPollution 是这一改造的核心防御性测试：
// 旧实现复用同一 *gorm.DB 链式 builder 在 4 次终止操作间累计 Select/Group，
// 导致 by_batch 仍带前一次 Select(province) 条件，结果错位。
// 这里断言 4 个统计维度独立且正确。
func TestGetAdmissionStatistics_NoBuilderPollution(t *testing.T) {
	db := newAdmissionTestDB(t)
	seedAdmissionFixtures(t, db.PostgreSQL)
	svc := NewAdmissionService(db, logrus.New())

	stats, err := svc.GetAdmissionStatistics(context.Background(), 2024)
	if err != nil {
		t.Fatalf("GetAdmissionStatistics: %v", err)
	}

	if got := stats["total"].(int64); got != 5 {
		t.Fatalf("total: want 5 (year=2024), got %d", got)
	}

	provinceStats, ok := stats["by_province"].([]AdmissionProvinceStat)
	if !ok {
		t.Fatalf("by_province has unexpected type %T", stats["by_province"])
	}
	provinceMap := map[string]int64{}
	for _, p := range provinceStats {
		provinceMap[p.Province] = p.Count
	}
	if provinceMap["北京"] != 2 || provinceMap["上海"] != 1 || provinceMap["广东"] != 2 {
		t.Fatalf("by_province distribution wrong: %+v", provinceMap)
	}

	batchStats, ok := stats["by_batch"].([]AdmissionBatchStat)
	if !ok {
		t.Fatalf("by_batch has unexpected type %T", stats["by_batch"])
	}
	batchMap := map[string]int64{}
	for _, b := range batchStats {
		batchMap[b.Batch] = b.Count
	}
	// 关键断言：如果 builder 污染，by_batch 会继承前一次 GROUP BY province
	// 或留下 SELECT province，导致行数 != 2 / Batch 字段为空。
	if len(batchMap) != 2 || batchMap["first_batch"] != 3 || batchMap["second_batch"] != 2 {
		t.Fatalf("by_batch distribution wrong: %+v (builder pollution suspected)", batchMap)
	}

	scoreDist, ok := stats["score_distribution"].(AdmissionScoreDistribution)
	if !ok {
		t.Fatalf("score_distribution has unexpected type %T", stats["score_distribution"])
	}
	if scoreDist.MinScore != 500 {
		t.Fatalf("min_score: want 500, got %v", scoreDist.MinScore)
	}
	if scoreDist.MaxScore != 690 {
		t.Fatalf("max_score: want 690, got %v", scoreDist.MaxScore)
	}
	// AvgScore 是对 avg_score 列再求平均：(640+630+650+540+550)/5 = 602
	if scoreDist.AvgScore < 601 || scoreDist.AvgScore > 603 {
		t.Fatalf("avg_score: want ~602, got %v", scoreDist.AvgScore)
	}
}

// TestGetAdmissionStatistics_NoYearFilter 验证 year=0 时返回全年总和（不应被前次 Where(year=?) 污染）。
func TestGetAdmissionStatistics_NoYearFilter(t *testing.T) {
	db := newAdmissionTestDB(t)
	seedAdmissionFixtures(t, db.PostgreSQL)
	svc := NewAdmissionService(db, logrus.New())

	stats, err := svc.GetAdmissionStatistics(context.Background(), 0)
	if err != nil {
		t.Fatalf("GetAdmissionStatistics(0): %v", err)
	}
	if got := stats["total"].(int64); got != 6 {
		t.Fatalf("total (all years): want 6, got %d", got)
	}
}
