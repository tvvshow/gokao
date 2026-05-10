package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/tvvshow/gokao/pkg/testutil"
)

// APIIntegrationTestSuite API集成测试套件
type APIIntegrationTestSuite struct {
	suite.Suite
	env        *testutil.TestEnvironment
	server     *httptest.Server
	httpClient *http.Client
	baseURL    string
}

// SetupSuite 测试套件设置
func (suite *APIIntegrationTestSuite) SetupSuite() {
	// 设置测试环境
	suite.env = testutil.SetupTestEnvironment(suite.T(), &testutil.TestConfig{
		DatabaseURL:     "postgres://postgres:postgres@localhost:5432/gaokao_test?sslmode=disable",
		RedisURL:        "localhost:6379",
		TestTimeout:     30 * time.Second,
		EnableMock:      false, // 使用真实数据库进行集成测试
		CleanupDatabase: true,
	})

	// 创建HTTP测试服务器
	// 这里需要根据实际的路由设置来创建服务器
	suite.server = httptest.NewServer(suite.createTestRouter())
	suite.baseURL = suite.server.URL

	// 创建HTTP客户端
	suite.httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}

	// 加载测试数据
	suite.loadTestData()
}

// TearDownSuite 测试套件清理
func (suite *APIIntegrationTestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}
	if suite.env != nil {
		suite.env.Teardown()
	}
}

// SetupTest 每个测试前的设置
func (suite *APIIntegrationTestSuite) SetupTest() {
	// 确保数据库是干净的
	if suite.env != nil && suite.env.Config.CleanupDatabase {
		suite.env.CleanupDatabase(suite.T())
	}
}

// createTestRouter 创建测试路由
// 这里需要根据实际的路由设置来实现
func (suite *APIIntegrationTestSuite) createTestRouter() http.Handler {
	// 返回一个简单的测试处理器
	// 在实际项目中，这里应该返回真正的路由处理器
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/health":
			suite.handleHealthCheck(w, r)
		case "/api/universities":
			suite.handleUniversities(w, r)
		case "/api/majors":
			suite.handleMajors(w, r)
		case "/api/recommendations":
			suite.handleRecommendations(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// loadTestData 加载测试数据
func (suite *APIIntegrationTestSuite) loadTestData() {
	// 这里应该加载测试数据到数据库
	// 例如：创建测试用户、大学、专业等
}

// TestHealthCheck 测试健康检查接口
func (suite *APIIntegrationTestSuite) TestHealthCheck() {
	req, err := http.NewRequest("GET", suite.baseURL+"/api/health", nil)
	require.NoError(suite.T(), err)

	resp, err := suite.httpClient.Do(req)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "healthy", result["status"])
}

// TestUniversitiesAPI 测试大学API
func (suite *APIIntegrationTestSuite) TestUniversitiesAPI() {
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectData     bool
	}{
		{
			name:           "GET universities list",
			method:         "GET",
			path:           "/api/universities",
			expectedStatus: http.StatusOK,
			expectData:     true,
		},
		{
			name:           "GET university by id",
			method:         "GET",
			path:           "/api/universities/1",
			expectedStatus: http.StatusOK,
			expectData:     true,
		},
		{
			name:           "GET non-existent university",
			method:         "GET",
			path:           "/api/universities/9999",
			expectedStatus: http.StatusNotFound,
			expectData:     false,
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, suite.baseURL+tt.path, nil)
			require.NoError(t, err)

			resp, err := suite.httpClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectData && resp.StatusCode == http.StatusOK {
				var data interface{}
				err = json.NewDecoder(resp.Body).Decode(&data)
				require.NoError(t, err)
				assert.NotNil(t, data)
			}
		})
	}
}

// TestRecommendationsAPI 测试推荐API
func (suite *APIIntegrationTestSuite) TestRecommendationsAPI() {
	// 创建测试用户数据
	userData := map[string]interface{}{
		"province":    "北京",
		"score":       650,
		"ranking":     1000,
		"preferences": []string{"计算机", "电子信息"},
	}

	body, err := json.Marshal(userData)
	require.NoError(suite.T(), err)

	req, err := http.NewRequest("POST", suite.baseURL+"/api/recommendations", bytes.NewReader(body))
	require.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.httpClient.Do(req)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(suite.T(), err)

	assert.Contains(suite.T(), result, "recommendations")
	assert.Contains(suite.T(), result, "algorithm")
	assert.Contains(suite.T(), result, "confidence")
}

// TestAPIPerformance API性能测试
func (suite *APIIntegrationTestSuite) TestAPIPerformance() {
	if testing.Short() {
		suite.T().Skip("Skipping performance test in short mode")
	}

	start := time.Now()
	const numRequests = 100

	for i := 0; i < numRequests; i++ {
		req, err := http.NewRequest("GET", suite.baseURL+"/api/health", nil)
		require.NoError(suite.T(), err)

		resp, err := suite.httpClient.Do(req)
		require.NoError(suite.T(), err)
		resp.Body.Close()

		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	}

	duration := time.Since(start)
	avgLatency := duration / numRequests

	suite.T().Logf("Average latency for %d requests: %v", numRequests, avgLatency)
	assert.True(suite.T(), avgLatency < 100*time.Millisecond,
		"Average latency should be less than 100ms, got %v", avgLatency)
}

// TestAPIConcurrency API并发测试
func (suite *APIIntegrationTestSuite) TestAPIConcurrency() {
	if testing.Short() {
		suite.T().Skip("Skipping concurrency test in short mode")
	}

	const numConcurrent = 50
	var wg sync.WaitGroup
	errors := make(chan error, numConcurrent)

	start := time.Now()

	for i := 0; i < numConcurrent; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			req, err := http.NewRequest("GET", suite.baseURL+"/api/health", nil)
			if err != nil {
				errors <- err
				return
			}

			resp, err := suite.httpClient.Do(req)
			if err != nil {
				errors <- err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errors <- fmt.Errorf("unexpected status code: %d", resp.StatusCode)
				return
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	duration := time.Since(start)
	suite.T().Logf("Completed %d concurrent requests in %v", numConcurrent, duration)

	// 检查错误
	for err := range errors {
		if err != nil {
			suite.T().Errorf("Request failed: %v", err)
		}
	}
}

// 简单的处理器实现（用于测试）

func (suite *APIIntegrationTestSuite) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"services": map[string]string{
			"database": "healthy",
			"redis":    "healthy",
			"api":      "healthy",
		},
	})
}

func (suite *APIIntegrationTestSuite) handleUniversities(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 模拟大学数据
	universities := []map[string]interface{}{
		{
			"id":       1,
			"name":     "清华大学",
			"province": "北京",
			"ranking":  1,
		},
		{
			"id":       2,
			"name":     "北京大学",
			"province": "北京",
			"ranking":  2,
		},
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"universities": universities,
		"total":        len(universities),
		"page":         1,
		"pageSize":     10,
	})
}

func (suite *APIIntegrationTestSuite) handleMajors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 模拟专业数据
	majors := []map[string]interface{}{
		{
			"id":       1,
			"name":     "计算机科学与技术",
			"category": "工学",
		},
		{
			"id":       2,
			"name":     "电子信息工程",
			"category": "工学",
		},
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"majors": majors,
		"total":  len(majors),
	})
}

func (suite *APIIntegrationTestSuite) handleRecommendations(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Province    string   `json:"province"`
		Score       int      `json:"score"`
		Ranking     int      `json:"ranking"`
		Preferences []string `json:"preferences"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// 模拟推荐结果
	recommendations := []map[string]interface{}{
		{
			"university": "清华大学",
			"major":      "计算机科学与技术",
			"score":      680,
			"chance":     0.85,
		},
		{
			"university": "北京大学",
			"major":      "电子信息工程",
			"score":      670,
			"chance":     0.78,
		},
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"recommendations": recommendations,
		"algorithm":       "rule-based",
		"confidence":      0.92,
		"generated_at":    time.Now().Format(time.RFC3339),
	})
}

// RunIntegrationTests 运行集成测试
func RunIntegrationTests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite.Run(t, new(APIIntegrationTestSuite))
}
