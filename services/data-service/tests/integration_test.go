package tests

import (
	"bytes"
	"encoding/json"
	"github.com/tvvshow/gokao/services/data-service/internal/config"
	"github.com/tvvshow/gokao/services/data-service/internal/database"
	"github.com/tvvshow/gokao/services/data-service/internal/handlers"
	"github.com/tvvshow/gokao/services/data-service/internal/middleware"
	"github.com/tvvshow/gokao/services/data-service/internal/models"
	"github.com/tvvshow/gokao/services/data-service/internal/services"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// IntegrationTestSuite 集成测试套件
type IntegrationTestSuite struct {
	suite.Suite
	router *gin.Engine
	db     *gorm.DB
	logger *logrus.Logger
}

// SetupSuite 设置测试套件
func (suite *IntegrationTestSuite) SetupSuite() {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 创建测试数据库
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	suite.Require().NoError(err)
	sqlDB, err := db.DB()
	suite.Require().NoError(err)
	sqlDB.SetMaxOpenConns(1)

	// 自动迁移
	err = db.AutoMigrate(
		&models.University{},
		&models.Major{},
		&models.AdmissionData{},
		&models.SearchIndex{},
		&models.AnalysisResult{},
		&models.HotSearch{},
		&models.DataStatistics{},
	)
	suite.Require().NoError(err)

	suite.db = db

	// 创建日志器
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	suite.logger = logger

	// 创建配置
	cfg := &config.Config{
		CacheEnabled:    false,
		MaxPageSize:     100,
		DefaultPageSize: 20,
		QueryTimeout:    30,
	}

	// 创建数据库管理器
	dbManager := &database.DB{
		PostgreSQL: db,
		Config:     cfg,
		Logger:     logger,
	}

	// 创建服务
	universityService := services.NewUniversityService(dbManager, logger)
	majorService := services.NewMajorService(dbManager, logger)
	admissionService := services.NewAdmissionService(dbManager, logger)
	searchService := services.NewSearchService(dbManager, logger)
	algorithmService := services.NewAlgorithmService(dbManager, logger)
	performanceService := services.NewPerformanceService(dbManager, logger)
	cacheService := services.NewCacheService(dbManager, logger)

	// 创建处理器
	universityHandler := handlers.NewUniversityHandler(universityService, logger)
	majorHandler := handlers.NewMajorHandler(majorService, logger)
	admissionHandler := handlers.NewAdmissionHandler(admissionService, logger)
	searchHandler := handlers.NewSearchHandler(searchService, logger)
	algorithmHandler := handlers.NewAlgorithmHandler(algorithmService, logger)
	performanceHandler := handlers.NewPerformanceHandler(performanceService, cacheService, dbManager, logger)

	// 创建路由
	router := gin.New()
	router.Use(middleware.Logger(logger))
	router.Use(middleware.Recovery(logger))
	router.Use(middleware.PerformanceMonitoring(performanceService))

	// 注册路由
	apiV1 := router.Group("/api/v1")
	{
		// 院校路由
		universities := apiV1.Group("/universities")
		{
			universities.GET("", universityHandler.ListUniversities)
			universities.GET("/search", universityHandler.SearchUniversities)
			universities.GET("/statistics", universityHandler.GetUniversityStatistics)
			universities.GET("/:id", universityHandler.GetUniversityByID)
			universities.GET("/code/:code", universityHandler.GetUniversityByCode)
		}

		// 专业路由
		majors := apiV1.Group("/majors")
		{
			majors.GET("", majorHandler.ListMajors)
			majors.GET("/search", majorHandler.SearchMajors)
			majors.GET("/:id", majorHandler.GetMajorByID)
		}

		// 录取数据路由
		admission := apiV1.Group("/admission")
		{
			admission.GET("/data", admissionHandler.ListAdmissionData)
			admission.GET("/analyze", admissionHandler.AnalyzeAdmissionData)
			admission.POST("/predict", admissionHandler.PredictAdmission)
		}

		// 搜索路由
		search := apiV1.Group("/search")
		{
			search.GET("", searchHandler.GlobalSearch)
			search.GET("/autocomplete", searchHandler.AutoComplete)
		}

		// 算法路由
		algorithm := apiV1.Group("/algorithm")
		{
			algorithm.POST("/match", algorithmHandler.MatchVolunteers)
		}

		// 性能监控路由
		performance := apiV1.Group("/performance")
		{
			performance.GET("/metrics", performanceHandler.GetMetrics)
			performance.GET("/summary", performanceHandler.GetSummary)
		}
	}

	suite.router = router

	// 插入测试数据
	suite.seedTestData()
}

// TearDownSuite 清理测试套件
func (suite *IntegrationTestSuite) TearDownSuite() {
	sqlDB, _ := suite.db.DB()
	sqlDB.Close()
}

// seedTestData 插入测试数据
func (suite *IntegrationTestSuite) seedTestData() {
	// 创建测试院校
	universities := []models.University{
		{
			ID:           uuid.New(),
			Code:         "10001",
			Name:         "北京大学",
			Type:         "undergraduate",
			Level:        "985",
			Nature:       "public",
			Province:     "北京市",
			City:         "北京市",
			NationalRank: 1,
			OverallScore: 95.5,
			IsActive:     true,
			IsRecruiting: true,
		},
		{
			ID:           uuid.New(),
			Code:         "10002",
			Name:         "清华大学",
			Type:         "undergraduate",
			Level:        "985",
			Nature:       "public",
			Province:     "北京市",
			City:         "北京市",
			NationalRank: 2,
			OverallScore: 95.0,
			IsActive:     true,
			IsRecruiting: true,
		},
	}

	for _, u := range universities {
		suite.db.Create(&u)
	}

	// 创建测试专业
	majors := []models.Major{
		{
			ID:              uuid.New(),
			UniversityID:    universities[0].ID,
			Code:            "080901",
			Name:            "计算机科学与技术",
			Category:        "工学",
			Discipline:      "计算机类",
			DegreeType:      "bachelor",
			EmploymentRate:  95.5,
			AverageSalary:   12000,
			PopularityScore: 85.0,
			IsActive:        true,
			IsRecruiting:    true,
		},
	}

	for _, m := range majors {
		suite.db.Create(&m)
	}

	// 创建测试录取数据
	admissionData := []models.AdmissionData{
		{
			ID:            uuid.New(),
			UniversityID:  universities[0].ID,
			MajorID:       &majors[0].ID,
			Year:          2023,
			Province:      "北京市",
			Batch:         "first_batch",
			Category:      "science",
			MinScore:      650.0,
			MaxScore:      690.0,
			AvgScore:      670.0,
			MinRank:       500,
			MaxRank:       100,
			AvgRank:       300,
			PlannedCount:  100,
			ActualCount:   100,
			AdmissionRate: 0.8,
		},
	}

	for _, ad := range admissionData {
		suite.db.Create(&ad)
	}
}

// TestUniversityAPI 测试院校API
func (suite *IntegrationTestSuite) TestUniversityAPI() {
	// 测试获取院校列表
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/universities", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), 200, w.Code)

	var response handlers.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)

	// 测试搜索院校
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/universities/search?q=北京", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), 200, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)

	// 测试获取院校统计
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/universities/statistics", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), 200, w.Code)
}

// TestMajorAPI 测试专业API
func (suite *IntegrationTestSuite) TestMajorAPI() {
	// 测试获取专业列表
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/majors", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), 200, w.Code)

	var response handlers.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)

	// 测试搜索专业
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/majors/search?q=计算机", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), 200, w.Code)
}

// TestAdmissionAPI 测试录取数据API
func (suite *IntegrationTestSuite) TestAdmissionAPI() {
	// 测试获取录取数据列表
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/admission/data", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), 200, w.Code)

	var response handlers.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)

	// 测试录取概率预测
	predictionReq := services.PredictionRequest{
		UniversityID: suite.getFirstUniversityID(),
		Province:     "北京市",
		Category:     "science",
		Score:        660.0,
	}

	reqBody, _ := json.Marshal(predictionReq)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/admission/predict", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), 200, w.Code)
}

// TestSearchAPI 测试搜索API
func (suite *IntegrationTestSuite) TestSearchAPI() {
	// 测试全局搜索
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/search?keyword=北京大学", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), 200, w.Code)

	var response handlers.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)

	// 测试自动补全
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/search/autocomplete?keyword=北京", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), 200, w.Code)
}

// TestAlgorithmAPI 测试算法API
func (suite *IntegrationTestSuite) TestAlgorithmAPI() {
	// 测试志愿匹配
	matchReq := services.VolunteerMatchRequest{
		Province:      "北京市",
		Category:      "science",
		Score:         650.0,
		RiskTolerance: "moderate",
		Preferences: services.VolunteerPreferences{
			SchoolWeight:     0.4,
			MajorWeight:      0.3,
			LocationWeight:   0.2,
			EmploymentWeight: 0.1,
		},
	}

	reqBody, _ := json.Marshal(matchReq)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/algorithm/match", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), 200, w.Code)

	var response handlers.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
}

// TestPerformanceAPI 测试性能监控API
func (suite *IntegrationTestSuite) TestPerformanceAPI() {
	// 测试获取性能指标
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/performance/metrics", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), 200, w.Code)

	var response handlers.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)

	// 测试获取性能摘要
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/performance/summary", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), 200, w.Code)
}

// TestErrorHandling 测试错误处理
func (suite *IntegrationTestSuite) TestErrorHandling() {
	// 测试不存在的院校ID
	w := httptest.NewRecorder()
	nonExistentID := uuid.New().String()
	req, _ := http.NewRequest("GET", "/api/v1/universities/"+nonExistentID, nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), 404, w.Code)

	var response handlers.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)

	// 测试无效的搜索参数
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/search", nil) // 缺少keyword参数
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), 400, w.Code)
}

// TestPaginationAndFiltering 测试分页和过滤
func (suite *IntegrationTestSuite) TestPaginationAndFiltering() {
	// 测试分页
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/universities?page=1&page_size=1", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), 200, w.Code)

	var response handlers.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)

	// 测试过滤
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/universities?province=北京市&level=985", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), 200, w.Code)
}

// 辅助方法
// getFirstUniversityID 显式按 Code 取北大（admission fixture 关联的院校），
// 而不是依赖 db.First()。后者按主键升序取首条，但 University.ID 是 uuid.New()
// 随机生成的，主键序与插入顺序无关 — 这会导致测试随机命中清华或北大，
// 而 admission_data 仅与北大关联，测试稳定性变成概率事件。
func (suite *IntegrationTestSuite) getFirstUniversityID() string {
	var university models.University
	suite.db.Where("code = ?", "10001").First(&university)
	return university.ID.String()
}

// 运行集成测试套件
func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
