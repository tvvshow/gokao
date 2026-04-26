package tests

import (
	"context"
	"github.com/oktetopython/gaokao/services/data-service/internal/config"
	"github.com/oktetopython/gaokao/services/data-service/internal/database"
	"github.com/oktetopython/gaokao/services/data-service/internal/models"
	"github.com/oktetopython/gaokao/services/data-service/internal/services"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// UniversityServiceTestSuite 院校服务测试套件
type UniversityServiceTestSuite struct {
	suite.Suite
	db      *gorm.DB
	service *services.UniversityService
	ctx     context.Context
}

// SetupSuite 设置测试套件
func (suite *UniversityServiceTestSuite) SetupSuite() {
	// 使用内存SQLite数据库进行测试
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	suite.Require().NoError(err)
	sqlDB, err := db.DB()
	suite.Require().NoError(err)
	sqlDB.SetMaxOpenConns(1)

	// 自动迁移
	err = db.AutoMigrate(&models.University{}, &models.Major{}, &models.AdmissionData{})
	suite.Require().NoError(err)

	suite.db = db
	suite.ctx = context.Background()

	// 创建测试用的数据库连接管理器
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel) // 减少测试日志输出

	cfg := &config.Config{
		CacheEnabled:    false, // 测试时禁用缓存
		MaxPageSize:     100,
		DefaultPageSize: 20,
	}

	dbManager := &database.DB{
		PostgreSQL: db,
		Config:     cfg,
		Logger:     logger,
	}

	// 创建服务实例
	suite.service = services.NewUniversityService(dbManager, logger)
}

// TearDownSuite 清理测试套件
func (suite *UniversityServiceTestSuite) TearDownSuite() {
	sqlDB, _ := suite.db.DB()
	sqlDB.Close()
}

// SetupTest 设置每个测试
func (suite *UniversityServiceTestSuite) SetupTest() {
	// 清理数据
	suite.db.Exec("DELETE FROM universities")
	suite.db.Exec("DELETE FROM majors")
	suite.db.Exec("DELETE FROM admission_data")
}

// TestGetUniversityByID 测试根据ID获取院校
func (suite *UniversityServiceTestSuite) TestGetUniversityByID() {
	// 创建测试数据
	university := models.University{
		ID:       uuid.New(),
		Code:     "10001",
		Name:     "北京大学",
		Type:     "undergraduate",
		Level:    "985",
		Nature:   "public",
		Province: "北京市",
		City:     "北京市",
		IsActive: true,
	}
	suite.db.Create(&university)

	// 测试获取存在的院校
	result, err := suite.service.GetUniversityByID(suite.ctx, university.ID.String())
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(university.Name, result.Name)
	suite.Equal(university.Code, result.Code)

	// 测试获取不存在的院校
	nonExistentID := uuid.New().String()
	result, err = suite.service.GetUniversityByID(suite.ctx, nonExistentID)
	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "院校不存在")
}

// TestGetUniversityByCode 测试根据代码获取院校
func (suite *UniversityServiceTestSuite) TestGetUniversityByCode() {
	// 创建测试数据
	university := models.University{
		ID:       uuid.New(),
		Code:     "10001",
		Name:     "北京大学",
		Type:     "undergraduate",
		Level:    "985",
		Nature:   "public",
		Province: "北京市",
		City:     "北京市",
		IsActive: true,
	}
	suite.db.Create(&university)

	// 测试获取存在的院校
	result, err := suite.service.GetUniversityByCode(suite.ctx, university.Code)
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(university.Name, result.Name)
	suite.Equal(university.Code, result.Code)

	// 测试获取不存在的院校
	result, err = suite.service.GetUniversityByCode(suite.ctx, "99999")
	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "院校不存在")
}

// TestListUniversities 测试获取院校列表
func (suite *UniversityServiceTestSuite) TestListUniversities() {
	// 创建测试数据
	universities := []models.University{
		{
			ID:       uuid.New(),
			Code:     "10001",
			Name:     "北京大学",
			Type:     "undergraduate",
			Level:    "985",
			Province: "北京市",
			IsActive: true,
		},
		{
			ID:       uuid.New(),
			Code:     "10002",
			Name:     "清华大学",
			Type:     "undergraduate",
			Level:    "985",
			Province: "北京市",
			IsActive: true,
		},
		{
			ID:       uuid.New(),
			Code:     "10003",
			Name:     "复旦大学",
			Type:     "undergraduate",
			Level:    "985",
			Province: "上海市",
			IsActive: true,
		},
	}

	for _, u := range universities {
		suite.db.Create(&u)
	}

	// 测试无过滤条件的列表查询
	params := services.UniversityQueryParams{
		Page:     1,
		PageSize: 10,
	}
	result, err := suite.service.ListUniversities(suite.ctx, params)
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(int64(3), result.Total)
	suite.Len(result.Universities, 3)

	// 测试按省份过滤
	params.Province = "北京市"
	result, err = suite.service.ListUniversities(suite.ctx, params)
	suite.NoError(err)
	suite.Equal(int64(2), result.Total)
	suite.Len(result.Universities, 2)

	// 测试按层次过滤
	params.Province = ""
	params.Level = "985"
	result, err = suite.service.ListUniversities(suite.ctx, params)
	suite.NoError(err)
	suite.Equal(int64(3), result.Total)

	// 测试分页
	params.Level = ""
	params.PageSize = 2
	result, err = suite.service.ListUniversities(suite.ctx, params)
	suite.NoError(err)
	suite.Equal(int64(3), result.Total)
	suite.Len(result.Universities, 2)
	suite.Equal(2, result.TotalPages)
}

// TestGetUniversityStatistics 测试获取院校统计
func (suite *UniversityServiceTestSuite) TestGetUniversityStatistics() {
	// 创建测试数据
	universities := []models.University{
		{
			ID:       uuid.New(),
			Code:     "10001",
			Name:     "北京大学",
			Type:     "undergraduate",
			Level:    "985",
			Province: "北京市",
			IsActive: true,
		},
		{
			ID:       uuid.New(),
			Code:     "10002",
			Name:     "清华大学",
			Type:     "undergraduate",
			Level:    "985",
			Province: "北京市",
			IsActive: true,
		},
		{
			ID:       uuid.New(),
			Code:     "10003",
			Name:     "北京师范大学",
			Type:     "undergraduate",
			Level:    "211",
			Province: "北京市",
			IsActive: true,
		},
	}

	for _, u := range universities {
		suite.db.Create(&u)
	}

	// 测试统计功能
	stats, err := suite.service.GetUniversityStatistics(suite.ctx)
	suite.NoError(err)
	suite.NotNil(stats)

	// 验证总数
	suite.Equal(int64(3), stats.Total)

	// 验证层次统计
	suite.Equal(int64(2), stats.By985)
	suite.Equal(int64(1), stats.By211)

	// 验证按省份统计
	suite.NotEmpty(stats.ByProvince)
	suite.Equal(int64(3), stats.ByProvince["北京市"])
}

// TestUniversityQueryParamsValidation 测试查询参数验证
func (suite *UniversityServiceTestSuite) TestUniversityQueryParamsValidation() {
	// 测试分页参数自动修正
	params := services.UniversityQueryParams{
		Page:     0,  // 无效页码
		PageSize: -1, // 无效页大小
	}

	result, err := suite.service.ListUniversities(suite.ctx, params)
	suite.NoError(err)
	suite.Equal(1, result.Page)      // 应该被修正为1
	suite.Equal(20, result.PageSize) // 应该被修正为默认值
}

// 运行测试套件
func TestUniversityServiceSuite(t *testing.T) {
	suite.Run(t, new(UniversityServiceTestSuite))
}

// TestUniversityServiceBasic 基础功能测试
func TestUniversityServiceBasic(t *testing.T) {
	// 简单的单元测试示例
	assert := assert.New(t)

	// 测试UUID生成
	id := uuid.New()
	assert.NotEmpty(id.String())

	// 测试时间处理
	now := time.Now()
	assert.True(now.Before(time.Now().Add(time.Second)))
}
