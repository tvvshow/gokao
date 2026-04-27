package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAnalyticsService 模拟分析服务
type MockAnalyticsService struct {
	mock.Mock
}

func (m *MockAnalyticsService) GetRecommendationStats(userID string, startTime, endTime time.Time) (*RecommendationStats, error) {
	args := m.Called(userID, startTime, endTime)
	return args.Get(0).(*RecommendationStats), args.Error(1)
}

func (m *MockAnalyticsService) GetSystemMetrics() (*SystemMetrics, error) {
	args := m.Called()
	return args.Get(0).(*SystemMetrics), args.Error(1)
}

func (m *MockAnalyticsService) GetUserBehaviorAnalysis(userID string) (*UserBehaviorAnalysis, error) {
	args := m.Called(userID)
	return args.Get(0).(*UserBehaviorAnalysis), args.Error(1)
}

func (m *MockAnalyticsService) GetAlgorithmPerformance() (*AlgorithmPerformance, error) {
	args := m.Called()
	return args.Get(0).(*AlgorithmPerformance), args.Error(1)
}

func (m *MockAnalyticsService) GetPerformanceMetrics(ctx context.Context) (map[string]interface{}, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockAnalyticsService) GetFusionStatistics(ctx context.Context) (map[string]interface{}, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockAnalyticsService) GenerateQualityReport(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockAnalyticsService) GetRecommendationTrends(ctx context.Context, timeRange string) (map[string]interface{}, error) {
	args := m.Called(ctx, timeRange)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func setupAnalyticsHandler() (*AnalyticsHandler, *MockAnalyticsService) {
	mockService := new(MockAnalyticsService)
	handler := NewAnalyticsHandler(mockService)
	return handler, mockService
}

func TestAnalyticsHandler_GetRecommendationStats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		startTime      string
		endTime        string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "成功获取统计数据",
			userID:         "test_user_123",
			startTime:      time.Now().Add(-7 * 24 * time.Hour).Format(time.RFC3339),
			endTime:        time.Now().Format(time.RFC3339),
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "用户ID为空",
			userID:         "",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "无效的开始时间",
			userID:         "test_user_123",
			startTime:      "invalid-time",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "无效的结束时间",
			userID:         "test_user_123",
			endTime:        "invalid-time",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := setupAnalyticsHandler()

			if !tt.expectError && tt.userID != "" {
				mockStats := &RecommendationStats{
					UserID:        tt.userID,
					TotalRequests: 100,
					SuccessRate:   0.95,
				}
				mockService.On("GetRecommendationStats", mock.AnythingOfType("string"), mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(mockStats, nil)
			}

			// 创建请求
			req := httptest.NewRequest("GET", "/analytics/recommendations/"+tt.userID, nil)
			q := req.URL.Query()
			if tt.startTime != "" {
				q.Add("start_time", tt.startTime)
			}
			if tt.endTime != "" {
				q.Add("end_time", tt.endTime)
			}
			req.URL.RawQuery = q.Encode()

			// 创建响应记录器
			w := httptest.NewRecorder()

			// 创建Gin上下文
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{{Key: "user_id", Value: tt.userID}}

			// 调用处理器
			handler.GetRecommendationStats(c)

			// 验证响应
			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectError {
				var response RecommendationStats
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.userID, response.UserID)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestAnalyticsHandler_GetSystemMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, mockService := setupAnalyticsHandler()

	// 设置模拟返回
	mockMetrics := &SystemMetrics{
		Timestamp:     time.Now(),
		CPUUsage:      65.2,
		MemoryUsage:   78.5,
		ServiceHealth: "healthy",
	}
	mockService.On("GetSystemMetrics").Return(mockMetrics, nil)

	// 创建请求
	req := httptest.NewRequest("GET", "/analytics/system/metrics", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// 调用处理器
	handler.GetSystemMetrics(c)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var response SystemMetrics
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, mockMetrics.CPUUsage, response.CPUUsage)
	assert.Equal(t, mockMetrics.ServiceHealth, response.ServiceHealth)

	mockService.AssertExpectations(t)
}

func TestAnalyticsHandler_GetUserBehaviorAnalysis(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "成功获取用户行为分析",
			userID:         "test_user_123",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "用户ID为空",
			userID:         "",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := setupAnalyticsHandler()

			if !tt.expectError {
				mockAnalysis := &UserBehaviorAnalysis{
					UserID: tt.userID,
					EngagementMetrics: EngagementMetrics{
						ClickThroughRate: 0.15,
						ViewTime:         125.5,
					},
				}
				mockService.On("GetUserBehaviorAnalysis", tt.userID).Return(mockAnalysis, nil)
			}

			// 创建请求
			req := httptest.NewRequest("GET", "/analytics/users/"+tt.userID+"/behavior", nil)
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{{Key: "user_id", Value: tt.userID}}

			// 调用处理器
			handler.GetUserBehaviorAnalysis(c)

			// 验证响应
			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectError {
				var response UserBehaviorAnalysis
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.userID, response.UserID)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestAnalyticsHandler_GetAlgorithmPerformance(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, mockService := setupAnalyticsHandler()

	// 设置模拟返回
	mockPerformance := &AlgorithmPerformance{
		HybridAlgorithm: AnalyticsPerformanceMetrics{
			AvgResponseTime: 95.5,
			SuccessRate:     0.95,
			Accuracy:        0.93,
		},
		ComparisonMetrics: ComparisonMetrics{
			BestPerforming:  "hybrid",
			PerformanceGain: 12.5,
		},
		LastUpdated: time.Now(),
	}
	mockService.On("GetAlgorithmPerformance").Return(mockPerformance, nil)

	// 创建请求
	req := httptest.NewRequest("GET", "/analytics/algorithms/performance", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// 调用处理器
	handler.GetAlgorithmPerformance(c)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var response AlgorithmPerformance
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, mockPerformance.ComparisonMetrics.BestPerforming, response.ComparisonMetrics.BestPerforming)

	mockService.AssertExpectations(t)
}

func TestAnalyticsHandler_GetPerformanceMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, mockService := setupAnalyticsHandler()

	// 设置模拟返回
	mockMetrics := map[string]interface{}{
		"cpu_usage":    65.2,
		"memory_usage": 78.5,
		"qps":          150.5,
	}
	mockService.On("GetPerformanceMetrics", mock.Anything).Return(mockMetrics, nil)

	// 创建请求
	req := httptest.NewRequest("GET", "/analytics/performance", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// 调用处理器
	handler.GetPerformanceMetrics(c)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 65.2, response["cpu_usage"])

	mockService.AssertExpectations(t)
}

func TestAnalyticsHandler_GenerateQualityReport(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, mockService := setupAnalyticsHandler()

	// 测试请求参数
	requestParams := map[string]interface{}{
		"start_time": time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
		"end_time":   time.Now().Format(time.RFC3339),
		"user_id":    "test_user",
	}

	// 设置模拟返回
	mockReport := map[string]interface{}{
		"report_id":    "qr_123456",
		"generated_at": time.Now(),
		"summary": map[string]interface{}{
			"total_recommendations": 1250,
			"success_rate":          0.94,
		},
	}
	mockService.On("GenerateQualityReport", mock.Anything, mock.Anything).Return(mockReport, nil)

	// 创建请求
	requestBody, _ := json.Marshal(requestParams)
	req := httptest.NewRequest("POST", "/analytics/quality-report", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// 调用处理器
	handler.GenerateQualityReport(c)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "qr_123456", response["report_id"])

	mockService.AssertExpectations(t)
}

func TestAnalyticsHandler_GetRecommendationTrends(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		timeRange      string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "有效时间范围 - 1h",
			timeRange:      "1h",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "有效时间范围 - 24h",
			timeRange:      "24h",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "默认时间范围",
			timeRange:      "",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "无效时间范围",
			timeRange:      "invalid",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := setupAnalyticsHandler()

			if !tt.expectError {
				expectedTimeRange := tt.timeRange
				if expectedTimeRange == "" {
					expectedTimeRange = "24h"
				}

				mockTrends := map[string]interface{}{
					"time_range": map[string]interface{}{
						"duration": expectedTimeRange,
					},
					"request_volume": []interface{}{},
				}
				mockService.On("GetRecommendationTrends", mock.Anything, expectedTimeRange).Return(mockTrends, nil)
			}

			// 创建请求
			url := "/analytics/trends"
			if tt.timeRange != "" {
				url += "?time_range=" + tt.timeRange
			}
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// 调用处理器
			handler.GetRecommendationTrends(c)

			// 验证响应
			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectError {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "time_range")
			}

			mockService.AssertExpectations(t)
		})
	}
}

// 基准测试
func BenchmarkAnalyticsHandler_GetSystemMetrics(b *testing.B) {
	gin.SetMode(gin.TestMode)
	handler, mockService := setupAnalyticsHandler()

	mockMetrics := &SystemMetrics{
		Timestamp:     time.Now(),
		CPUUsage:      65.2,
		MemoryUsage:   78.5,
		ServiceHealth: "healthy",
	}
	mockService.On("GetSystemMetrics").Return(mockMetrics, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/analytics/system/metrics", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler.GetSystemMetrics(c)
	}
}

func TestAnalyticsHandler_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, mockService := setupAnalyticsHandler()
	router := gin.New()
	api := router.Group("/api/v1")
	handler.RegisterRoutes(api)

	mockMetrics := &SystemMetrics{
		Timestamp:     time.Now(),
		CPUUsage:      45.0,
		MemoryUsage:   52.0,
		ServiceHealth: "healthy",
	}
	mockService.On("GetSystemMetrics").Return(mockMetrics, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/system/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}
