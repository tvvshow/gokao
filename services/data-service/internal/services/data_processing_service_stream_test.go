package services

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/tvvshow/gokao/services/data-service/internal/config"
	"github.com/tvvshow/gokao/services/data-service/internal/database"
	"github.com/tvvshow/gokao/services/data-service/internal/models"
)

// newProcessingTestDB 复用 admission_service_test 的 sqlite 套路，
// 但需要 AutoMigrate University/Major/AdmissionData 三表配套。
func newProcessingTestDB(t *testing.T) *database.DB {
	t.Helper()
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := gormDB.AutoMigrate(&models.University{}, &models.Major{}, &models.AdmissionData{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return &database.DB{
		PostgreSQL: gormDB,
		Config:     &config.Config{CacheEnabled: false},
		Logger:     logrus.New(),
	}
}

// TestProcessUniversityDataStream_HappyPath 验证流式 decode + upsert 全流程：
// 首次解析创建，第二次同输入触发 update 分支（按 code 匹配 existing）。
func TestProcessUniversityDataStream_HappyPath(t *testing.T) {
	db := newProcessingTestDB(t)
	svc := NewDataProcessingService(db, logrus.New())

	payload := `[
		{"code": "10001", "name": "北京大学", "type": "undergraduate"},
		{"code": "10002", "name": "清华大学", "type": "undergraduate"}
	]`
	if err := svc.ProcessUniversityDataStream(strings.NewReader(payload)); err != nil {
		t.Fatalf("first import: %v", err)
	}

	var count int64
	db.PostgreSQL.Model(&models.University{}).Count(&count)
	if count != 2 {
		t.Fatalf("after first import: want 2 rows, got %d", count)
	}

	// 第二次同样 code，验证 upsert update 路径生效（不会插重复）。
	updated := `[
		{"code": "10001", "name": "北京大学（更新）", "type": "undergraduate"}
	]`
	if err := svc.ProcessUniversityDataStream(strings.NewReader(updated)); err != nil {
		t.Fatalf("upsert import: %v", err)
	}
	db.PostgreSQL.Model(&models.University{}).Count(&count)
	if count != 2 {
		t.Fatalf("after upsert: want still 2 rows (no duplicate), got %d", count)
	}
	var pku models.University
	db.PostgreSQL.Where("code = ?", "10001").First(&pku)
	if pku.Name != "北京大学（更新）" {
		t.Fatalf("upsert update failed: name=%q", pku.Name)
	}
}

// TestProcessUniversityDataStream_RejectsNonArray 验证显式校验顶层必须是数组，
// 不是数组时立即返错而不是行为不定地继续。
func TestProcessUniversityDataStream_RejectsNonArray(t *testing.T) {
	db := newProcessingTestDB(t)
	svc := NewDataProcessingService(db, logrus.New())

	payload := `{"code": "10001", "name": "北京大学"}`
	err := svc.ProcessUniversityDataStream(strings.NewReader(payload))
	if err == nil {
		t.Fatal("expected error for non-array JSON, got nil")
	}
	if !strings.Contains(err.Error(), "数组") {
		t.Fatalf("error should mention array expectation, got: %v", err)
	}

	var count int64
	db.PostgreSQL.Model(&models.University{}).Count(&count)
	if count != 0 {
		t.Fatalf("nothing should be inserted on parse failure, got %d rows", count)
	}
}

// TestProcessUniversityDataStream_RollbackOnPartialFailure 验证流式解析途中
// 单条 record 反序列化失败时整个事务回滚，已 decode 的前几条不会半残留库。
func TestProcessUniversityDataStream_RollbackOnPartialFailure(t *testing.T) {
	db := newProcessingTestDB(t)
	svc := NewDataProcessingService(db, logrus.New())

	// 第 3 条是非法 JSON（code 字段类型错误：数字而非字符串）。
	// 但 GORM/sqlite 对 string 字段写入 number 实际能容忍，所以构造一个
	// 更硬的解析失败：直接拼接非法的 JSON token。
	payload := `[
		{"code": "10001", "name": "北京大学"},
		{"code": "10002", "name": "清华大学"},
		not-json-here
	]`
	err := svc.ProcessUniversityDataStream(strings.NewReader(payload))
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}

	var count int64
	db.PostgreSQL.Model(&models.University{}).Count(&count)
	if count != 0 {
		t.Fatalf("transaction should rollback all, got %d rows persisted", count)
	}
}

// TestProcessUniversityDataStream_EmptyArray 验证空数组场景下事务正常提交，无副作用。
func TestProcessUniversityDataStream_EmptyArray(t *testing.T) {
	db := newProcessingTestDB(t)
	svc := NewDataProcessingService(db, logrus.New())

	if err := svc.ProcessUniversityDataStream(strings.NewReader(`[]`)); err != nil {
		t.Fatalf("empty array should succeed: %v", err)
	}

	var count int64
	db.PostgreSQL.Model(&models.University{}).Count(&count)
	if count != 0 {
		t.Fatalf("empty input should leave table empty, got %d", count)
	}
}

// TestProcessMajorDataStream_HappyPath 覆盖专业流式入口的基础语义。
func TestProcessMajorDataStream_HappyPath(t *testing.T) {
	db := newProcessingTestDB(t)
	svc := NewDataProcessingService(db, logrus.New())

	// 先插一个 university 以便 major.university_id 有意义。
	uni := models.University{ID: uuid.New(), Code: "10001", Name: "北京大学"}
	db.PostgreSQL.Create(&uni)

	payload := `[
		{"code": "080901", "name": "计算机科学与技术", "university_id": "` + uni.ID.String() + `"}
	]`
	if err := svc.ProcessMajorDataStream(strings.NewReader(payload)); err != nil {
		t.Fatalf("major stream: %v", err)
	}

	var count int64
	db.PostgreSQL.Model(&models.Major{}).Count(&count)
	if count != 1 {
		t.Fatalf("want 1 major, got %d", count)
	}
}

// TestProcessAdmissionDataStream_HappyPath 覆盖录取流式入口。
func TestProcessAdmissionDataStream_HappyPath(t *testing.T) {
	db := newProcessingTestDB(t)
	svc := NewDataProcessingService(db, logrus.New())

	uniID := uuid.New()
	db.PostgreSQL.Create(&models.University{ID: uniID, Code: "10001", Name: "北京大学"})

	payload := `[
		{"university_id": "` + uniID.String() + `", "year": 2024, "province": "北京", "category": "science", "min_score": 600, "avg_score": 640}
	]`
	if err := svc.ProcessAdmissionDataStream(strings.NewReader(payload)); err != nil {
		t.Fatalf("admission stream: %v", err)
	}

	var count int64
	db.PostgreSQL.Model(&models.AdmissionData{}).Count(&count)
	if count != 1 {
		t.Fatalf("want 1 admission, got %d", count)
	}
}

// TestProcessUniversityData_LegacyByteEntry 验证旧 []byte 入口仍工作（向后兼容）。
func TestProcessUniversityData_LegacyByteEntry(t *testing.T) {
	db := newProcessingTestDB(t)
	svc := NewDataProcessingService(db, logrus.New())

	data := []byte(`[{"code": "10001", "name": "北京大学"}]`)
	if err := svc.ProcessUniversityData(data); err != nil {
		t.Fatalf("legacy entry: %v", err)
	}

	var count int64
	db.PostgreSQL.Model(&models.University{}).Count(&count)
	if count != 1 {
		t.Fatalf("legacy entry: want 1 row, got %d", count)
	}
}
