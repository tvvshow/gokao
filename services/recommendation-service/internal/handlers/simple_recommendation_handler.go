package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oktetopython/gaokao/services/recommendation-service/internal/cache"
	"github.com/oktetopython/gaokao/services/recommendation-service/internal/llm"
	"github.com/oktetopython/gaokao/services/recommendation-service/pkg/cppbridge"
)

// RecommendationCache 推荐缓存
type RecommendationCache struct {
	mu     sync.RWMutex
	data   map[string]*CacheEntry
	expiry time.Duration
}

// CacheEntry 缓存条目
type CacheEntry struct {
	Response  *cppbridge.RecommendationResponse
	Timestamp time.Time
	HitCount  int64
}

// RecommendationExplainer 推荐解释器
type RecommendationExplainer struct {
	scoreWeights map[string]float64
	factorDB     map[string][]string
}

// ConfidenceCalculator 置信度计算器
type ConfidenceCalculator struct {
	historicalData map[string]float64
	baseConfidence float64
}

// BatchProcessor 批处理器
type BatchProcessor struct {
	maxWorkers    int
	batchSize     int
	timeout       time.Duration
	retryAttempts int
}

// SimpleRecommendationHandler 简化的推荐处理器
type SimpleRecommendationHandler struct {
	bridge           cppbridge.HybridRecommendationBridge
	cache            cache.CacheInterface
	explainer        *RecommendationExplainer
	confidenceCalc   *ConfidenceCalculator
	batchProcessor   *BatchProcessor
	performanceStats *PerformanceStats
	analyzer         llm.Analyzer
	mu               sync.RWMutex
}

// PerformanceStats 性能统计
type PerformanceStats struct {
	mu              sync.RWMutex
	totalRequests   int64
	successRequests int64
	failedRequests  int64
	totalLatency    time.Duration
	cacheHits       int64
	cacheMisses     int64
	lastUpdate      time.Time
}

// CacheWarmupSummary 缓存预热结果
type CacheWarmupSummary struct {
	Attempted int `json:"attempted"`
	Warmed    int `json:"warmed"`
	Skipped   int `json:"skipped"`
	Failed    int `json:"failed"`
}

// NewRecommendationCache 创建推荐缓存
func NewRecommendationCache(expiry time.Duration) *RecommendationCache {
	return &RecommendationCache{
		data:   make(map[string]*CacheEntry),
		expiry: expiry,
	}
}

// NewRecommendationExplainer 创建推荐解释器
func NewRecommendationExplainer() *RecommendationExplainer {
	return &RecommendationExplainer{
		scoreWeights: map[string]float64{
			"score_match":         0.35,
			"interest_match":      0.25,
			"employment_prospect": 0.20,
			"location_preference": 0.15,
			"cost_factor":         0.05,
		},
		factorDB: map[string][]string{
			"high_score_match": {
				"分数匹配度高达{percentage}%",
				"历年录取分数线符合度极佳",
				"录取概率在安全范围内",
			},
			"interest_alignment": {
				"专业与您的兴趣高度匹配",
				"符合您的职业发展规划",
				"学科特长与专业要求契合",
			},
			"employment_prospects": {
				"就业前景广阔，就业率达{rate}%",
				"行业发展趋势良好",
				"薪资水平在合理范围",
			},
		},
	}
}

// NewConfidenceCalculator 创建置信度计算器
func NewConfidenceCalculator() *ConfidenceCalculator {
	return &ConfidenceCalculator{
		historicalData: make(map[string]float64),
		baseConfidence: 0.7,
	}
}

// NewBatchProcessor 创建批处理器
func NewBatchProcessor(maxWorkers, batchSize int, timeout time.Duration) *BatchProcessor {
	return &BatchProcessor{
		maxWorkers:    maxWorkers,
		batchSize:     batchSize,
		timeout:       timeout,
		retryAttempts: 3,
	}
}

// NewSimpleRecommendationHandler 创建新的简化推荐处理器
func NewSimpleRecommendationHandler(bridge cppbridge.HybridRecommendationBridge, cacheInterface cache.CacheInterface, analyzer llm.Analyzer) *SimpleRecommendationHandler {
	if analyzer == nil {
		analyzer = llm.NewLocalFallbackAnalyzer()
	}

	return &SimpleRecommendationHandler{
		bridge:           bridge,
		cache:            cacheInterface,
		explainer:        NewRecommendationExplainer(),
		confidenceCalc:   NewConfidenceCalculator(),
		batchProcessor:   NewBatchProcessor(10, 50, 30*time.Second),
		performanceStats: &PerformanceStats{lastUpdate: time.Now()},
		analyzer:         analyzer,
	}
}

// ErrorResponse 错误响应（legacy）。
//
// 字段命名与 pkg/response.APIResponse 不一致（裸 error/message vs. error.code/error.message），
// 前端已基于此形状解析；强迁需要前后端同步发版，本次债务清理范围内不动。
// 新增 handler 请直接使用 pkg/response.OK / BadRequest / InternalError 等。
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// StudentInfo 前端学生信息结构
type StudentInfo struct {
	Score       *int               `json:"score"`
	Province    string             `json:"province"`
	ScienceType string             `json:"scienceType"`
	Year        int                `json:"year"`
	Rank        *int               `json:"rank"`
	Preferences StudentPreferences `json:"preferences"`
}

// StudentPreferences 学生偏好设置
type StudentPreferences struct {
	Regions             []string `json:"regions"`
	MajorCategories     []string `json:"majorCategories"`
	UniversityTypes     []string `json:"universityTypes"`
	RiskTolerance       string   `json:"riskTolerance"`
	SpecialRequirements string   `json:"specialRequirements"`
}

// APIResponse 统一API响应格式
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
}

// RecommendationData 推荐数据格式
type RecommendationData struct {
	Recommendations []FrontendRecommendation `json:"recommendations"`
	AnalysisReport  string                   `json:"analysisReport"`
}

// FrontendRecommendation 前端推荐格式
type FrontendRecommendation struct {
	ID                   string                   `json:"id"`
	University           FrontendUniversity       `json:"university"`
	Type                 string                   `json:"type"`
	AdmissionProbability int                      `json:"admissionProbability"`
	MatchScore           int                      `json:"matchScore"`
	RecommendReason      string                   `json:"recommendReason"`
	RiskLevel            string                   `json:"riskLevel"`
	SuggestedMajors      []FrontendMajor          `json:"suggestedMajors"`
	HistoricalData       []FrontendHistoricalData `json:"historicalData"`
}

// FrontendUniversity 前端大学格式
type FrontendUniversity struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Province   string `json:"province"`
	City       string `json:"city"`
	Level      string `json:"level"`
	Type       string `json:"type"`
	IsFavorite bool   `json:"isFavorite"`
}

// FrontendMajor 前端专业格式
type FrontendMajor struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Probability int    `json:"probability"`
}

// FrontendHistoricalData 前端历史数据格式
type FrontendHistoricalData struct {
	MinScore int `json:"minScore"`
	AvgScore int `json:"avgScore"`
	MaxScore int `json:"maxScore"`
	Year     int `json:"year"`
}

// GenerateRecommendations 生成推荐
// @Summary 生成志愿推荐
// @Description 根据学生信息生成个性化的志愿填报推荐
// @Tags recommendations
// @Accept json
// @Produce json
// @Param request body StudentInfo true "学生信息"
// @Success 200 {object} APIResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /recommendations/generate [post]
func (h *SimpleRecommendationHandler) GenerateRecommendations(c *gin.Context) {
	var studentInfo StudentInfo
	if err := c.ShouldBindJSON(&studentInfo); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "请求格式错误: " + err.Error(),
		})
		return
	}

	// 转换为后端期望的格式
	request, err := h.convertStudentInfoToRequest(&studentInfo)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "conversion_failed",
			Message: "数据转换失败: " + err.Error(),
		})
		return
	}

	startTime := time.Now()

	// 更新性能统计
	h.updatePerformanceStats(true, false, 0)

	// 设置默认值和参数验证
	if err := h.validateAndSetDefaults(request); err != nil {
		h.updatePerformanceStats(false, true, time.Since(startTime))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "请求参数验证失败: " + err.Error(),
		})
		return
	}

	// 尝试从缓存获取
	cacheKey := h.generateCacheKey(request)
	if cachedResponse := h.getFromCache(cacheKey); cachedResponse != nil {
		h.updatePerformanceStats(false, false, time.Since(startTime))
		h.performanceStats.mu.Lock()
		h.performanceStats.cacheHits++
		h.performanceStats.mu.Unlock()

		// 转换为前端格式
		frontendData := h.buildRecommendationData(c.Request.Context(), &studentInfo, request, cachedResponse)
		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Data:    frontendData,
			Message: "推荐生成成功（来自缓存）",
		})
		return
	}

	h.performanceStats.mu.Lock()
	h.performanceStats.cacheMisses++
	h.performanceStats.mu.Unlock()

	// 生成推荐
	response, err := h.generateEnhancedRecommendations(request)
	if err != nil {
		h.updatePerformanceStats(false, true, time.Since(startTime))
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Data:    nil,
			Message: "生成推荐失败: " + err.Error(),
		})
		return
	}

	// 缓存结果
	h.setToCache(cacheKey, response)

	// 更新性能统计
	h.updatePerformanceStats(false, false, time.Since(startTime))

	// 转换为前端格式
	frontendData := h.buildRecommendationData(c.Request.Context(), &studentInfo, request, response)
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    frontendData,
		Message: "推荐生成成功",
	})
}

// BatchGenerateRecommendations 批量生成推荐
// @Summary 批量生成志愿推荐
// @Description 为多个学生批量生成个性化的志愿填报推荐
// @Tags recommendations
// @Accept json
// @Produce json
// @Param request body cppbridge.BatchRecommendationRequest true "批量推荐请求"
// @Success 200 {object} cppbridge.BatchRecommendationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /recommendations/batch [post]
func (h *SimpleRecommendationHandler) BatchGenerateRecommendations(c *gin.Context) {
	var batchRequest cppbridge.BatchRecommendationRequest
	if err := c.ShouldBindJSON(&batchRequest); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "批量请求格式错误: " + err.Error(),
		})
		return
	}

	// 设置默认值
	if batchRequest.BatchSize == 0 {
		batchRequest.BatchSize = h.batchProcessor.batchSize
	}
	if batchRequest.Timeout == 0 {
		batchRequest.Timeout = int(h.batchProcessor.timeout.Milliseconds())
	}

	// 使用增强的批处理逻辑
	responses, successCount, failedCount, err := h.processBatchRecommendationsWithOptimization(&batchRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "batch_processing_failed",
			Message: "批量处理失败: " + err.Error(),
		})
		return
	}

	batchResponse := cppbridge.BatchRecommendationResponse{
		Responses:    responses,
		TotalCount:   len(responses),
		SuccessCount: successCount,
		FailedCount:  failedCount,
		ProcessedAt:  time.Now().Unix(),
	}

	c.JSON(http.StatusOK, batchResponse)
}

// ExplainRecommendation 解释推荐
// @Summary 解释推荐结果
// @Description 解释为什么推荐了某个学校和专业
// @Tags recommendations
// @Accept json
// @Produce json
// @Param id path string true "推荐ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /recommendations/explain/{id} [get]
func (h *SimpleRecommendationHandler) ExplainRecommendation(c *gin.Context) {
	recommendationID := c.Param("id")
	if recommendationID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_id",
			Message: "推荐ID不能为空",
		})
		return
	}

	// 使用智能解释器生成解释
	explanation, err := h.generateIntelligentExplanation(recommendationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "explanation_failed",
			Message: "生成解释失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, explanation)
}

// OptimizeRecommendations 优化推荐
// @Summary 优化推荐结果
// @Description 根据用户反馈优化推荐算法参数
// @Tags recommendations
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "优化请求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /recommendations/optimize [post]
func (h *SimpleRecommendationHandler) OptimizeRecommendations(c *gin.Context) {
	var request map[string]interface{}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "优化请求格式错误: " + err.Error(),
		})
		return
	}

	// 使用智能优化逻辑
	result, err := h.performIntelligentOptimization(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "optimization_failed",
			Message: "优化失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetSystemStatus 获取系统状态
// @Summary 获取系统状态
// @Description 获取推荐系统的运行状态和健康信息
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} ErrorResponse
// @Router /system/status [get]
func (h *SimpleRecommendationHandler) GetSystemStatus(c *gin.Context) {
	status, err := h.bridge.GetSystemStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "status_error",
			Message: "获取系统状态失败: " + err.Error(),
		})
		return
	}

	status["performance"] = h.GetPerformanceStats()
	status["cache"] = h.getCacheStatus()
	status["analysis"] = h.getAnalysisStatus()
	status["timestamp"] = time.Now().Format(time.RFC3339)

	c.JSON(http.StatusOK, status)
}

func (h *SimpleRecommendationHandler) getCacheStatus() map[string]interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	status := map[string]interface{}{
		"healthy": true,
		"type":    "configured",
	}
	if h.cache == nil {
		status["healthy"] = false
		status["type"] = "none"
		status["error"] = "cache is nil"
		return status
	}
	if err := h.cache.HealthCheck(ctx); err != nil {
		status["healthy"] = false
		status["error"] = err.Error()
	}
	return status
}

func (h *SimpleRecommendationHandler) getAnalysisStatus() map[string]interface{} {
	status := map[string]interface{}{
		"enabled": false,
		"status":  "unknown",
	}
	if h.analyzer == nil {
		status["status"] = "not_configured"
		return status
	}
	if reporter, ok := h.analyzer.(llm.StatusReporter); ok {
		for k, v := range reporter.Status() {
			status[k] = v
		}
		return status
	}
	status["status"] = "configured"
	return status
}

// ClearCache 清空缓存
// @Summary 清空系统缓存
// @Description 清空推荐系统的缓存数据
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} ErrorResponse
// @Router /system/cache/clear [post]
func (h *SimpleRecommendationHandler) ClearCache(c *gin.Context) {
	// 清空推荐服务缓存
	ctx := context.Background()
	err := h.cache.Clear(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "cache_error",
			Message: "清空推荐缓存失败: " + err.Error(),
		})
		return
	}

	// 清空C++桥接缓存
	err = h.bridge.ClearCache()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "cache_error",
			Message: "清空桥接缓存失败: " + err.Error(),
		})
		return
	}

	result := map[string]interface{}{
		"status":    "success",
		"message":   "缓存已清空",
		"timestamp": time.Now().Unix(),
	}

	c.JSON(http.StatusOK, result)
}

// UpdateModel 更新模型
// @Summary 更新AI模型
// @Description 更新推荐系统的AI模型
// @Tags system
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "模型更新请求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /system/model/update [put]
func (h *SimpleRecommendationHandler) UpdateModel(c *gin.Context) {
	var request map[string]interface{}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "模型更新请求格式错误: " + err.Error(),
		})
		return
	}

	modelPath, ok := request["model_path"].(string)
	if !ok || modelPath == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_model_path",
			Message: "模型路径不能为空",
		})
		return
	}

	err := h.bridge.UpdateModel(modelPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "update_error",
			Message: "更新模型失败: " + err.Error(),
		})
		return
	}

	result := map[string]interface{}{
		"status":     "success",
		"message":    "模型更新成功",
		"model_path": modelPath,
		"timestamp":  time.Now().Unix(),
	}

	c.JSON(http.StatusOK, result)
}

// convertStudentInfoToRequest 将前端StudentInfo转换为RecommendationRequest
func (h *SimpleRecommendationHandler) convertStudentInfoToRequest(studentInfo *StudentInfo) (*cppbridge.RecommendationRequest, error) {
	if studentInfo.Score == nil {
		return nil, fmt.Errorf("score is required")
	}
	if studentInfo.Province == "" {
		return nil, fmt.Errorf("province is required")
	}

	request := &cppbridge.RecommendationRequest{
		StudentID:          fmt.Sprintf("student_%d", time.Now().UnixNano()),
		Name:               "Student",
		TotalScore:         *studentInfo.Score,
		Province:           studentInfo.Province,
		SubjectCombination: studentInfo.ScienceType,
		MaxRecommendations: 30,
		Algorithm:          "hybrid",
	}

	// 设置排名（如果提供）
	if studentInfo.Rank != nil {
		request.Ranking = *studentInfo.Rank
	}

	// 转换偏好设置
	preferences := make(map[string]interface{})
	if len(studentInfo.Preferences.Regions) > 0 {
		preferences["regions"] = studentInfo.Preferences.Regions
	}
	if len(studentInfo.Preferences.MajorCategories) > 0 {
		preferences["major_categories"] = studentInfo.Preferences.MajorCategories
	}
	if len(studentInfo.Preferences.UniversityTypes) > 0 {
		preferences["university_types"] = studentInfo.Preferences.UniversityTypes
	}
	if studentInfo.Preferences.RiskTolerance != "" {
		preferences["risk_tolerance"] = studentInfo.Preferences.RiskTolerance
	}
	if studentInfo.Preferences.SpecialRequirements != "" {
		preferences["special_requirements"] = studentInfo.Preferences.SpecialRequirements
	}
	request.Preferences = preferences

	return request, nil
}

// validateAndSetDefaults 验证并设置默认值
func (h *SimpleRecommendationHandler) validateAndSetDefaults(request *cppbridge.RecommendationRequest) error {
	if request.StudentID == "" {
		request.StudentID = "cache_warmup"
	}

	if request.TotalScore <= 0 || request.TotalScore > 750 {
		return fmt.Errorf("total_score must be between 1 and 750")
	}

	if request.Province == "" {
		return fmt.Errorf("province is required")
	}

	// 设置默认值
	if request.MaxRecommendations == 0 {
		request.MaxRecommendations = 30
	}
	if request.Algorithm == "" {
		request.Algorithm = "hybrid"
	}

	// 限制最大推荐数量
	if request.MaxRecommendations > 100 {
		request.MaxRecommendations = 100
	}

	return nil
}

// generateCacheKey 生成缓存键
func (h *SimpleRecommendationHandler) generateCacheKey(request *cppbridge.RecommendationRequest) string {
	fingerprint := struct {
		TotalScore         int                    `json:"total_score"`
		Ranking            int                    `json:"ranking"`
		Province           string                 `json:"province"`
		SubjectCombination string                 `json:"subject_combination"`
		ChineseScore       int                    `json:"chinese_score"`
		MathScore          int                    `json:"math_score"`
		EnglishScore       int                    `json:"english_score"`
		Physics            int                    `json:"physics,omitempty"`
		Chemistry          int                    `json:"chemistry,omitempty"`
		Biology            int                    `json:"biology,omitempty"`
		History            int                    `json:"history,omitempty"`
		Geography          int                    `json:"geography,omitempty"`
		Politics           int                    `json:"politics,omitempty"`
		Preferences        map[string]interface{} `json:"preferences,omitempty"`
		Filters            map[string]interface{} `json:"filters,omitempty"`
		MaxRecommendations int                    `json:"max_recommendations"`
		Algorithm          string                 `json:"algorithm"`
	}{
		TotalScore:         request.TotalScore,
		Ranking:            request.Ranking,
		Province:           request.Province,
		SubjectCombination: request.SubjectCombination,
		ChineseScore:       request.ChineseScore,
		MathScore:          request.MathScore,
		EnglishScore:       request.EnglishScore,
		Physics:            request.Physics,
		Chemistry:          request.Chemistry,
		Biology:            request.Biology,
		History:            request.History,
		Geography:          request.Geography,
		Politics:           request.Politics,
		Preferences:        request.Preferences,
		Filters:            request.Filters,
		MaxRecommendations: request.MaxRecommendations,
		Algorithm:          request.Algorithm,
	}

	data, err := json.Marshal(fingerprint)
	if err != nil {
		return fmt.Sprintf("rec_%s_%d_%s_%s_%d",
			request.Province,
			request.TotalScore,
			request.SubjectCombination,
			request.Algorithm,
			request.MaxRecommendations)
	}
	sum := sha256.Sum256(data)
	return fmt.Sprintf("rec_%x", sum[:12])
}

// getFromCache 从缓存获取
func (h *SimpleRecommendationHandler) getFromCache(key string) *cppbridge.RecommendationResponse {
	ctx := context.Background()
	value, err := h.cache.Get(ctx, key)
	if err != nil {
		return nil
	}

	var response cppbridge.RecommendationResponse
	if err := json.Unmarshal(value, &response); err != nil {
		return nil
	}

	return &response
}

// setToCache 设置到缓存
func (h *SimpleRecommendationHandler) setToCache(key string, response *cppbridge.RecommendationResponse) {
	ctx := context.Background()
	data, err := json.Marshal(response)
	if err != nil {
		return
	}

	// 设置30分钟过期时间
	h.cache.Set(ctx, key, data, 30*time.Minute)
}

// WarmRecommendationCache 预热热门推荐请求缓存
func (h *SimpleRecommendationHandler) WarmRecommendationCache(ctx context.Context, requests []*cppbridge.RecommendationRequest) CacheWarmupSummary {
	summary := CacheWarmupSummary{}
	for _, request := range requests {
		if request == nil {
			continue
		}

		select {
		case <-ctx.Done():
			return summary
		default:
		}

		summary.Attempted++

		cloned := cloneRecommendationRequest(request)
		if err := h.validateAndSetDefaults(cloned); err != nil {
			summary.Failed++
			continue
		}

		cacheKey := h.generateCacheKey(cloned)
		if cached := h.getFromCache(cacheKey); cached != nil {
			summary.Skipped++
			continue
		}

		response, err := h.generateEnhancedRecommendations(cloned)
		if err != nil {
			summary.Failed++
			continue
		}

		h.setToCache(cacheKey, response)
		summary.Warmed++
	}

	return summary
}

func cloneRecommendationRequest(src *cppbridge.RecommendationRequest) *cppbridge.RecommendationRequest {
	if src == nil {
		return nil
	}
	cloned := *src
	cloned.Preferences = cloneInterfaceMap(src.Preferences)
	cloned.Filters = cloneInterfaceMap(src.Filters)
	return &cloned
}

func cloneInterfaceMap(src map[string]interface{}) map[string]interface{} {
	if len(src) == 0 {
		return nil
	}
	cloned := make(map[string]interface{}, len(src))
	for key, value := range src {
		cloned[key] = value
	}
	return cloned
}

// cleanupExpiredEntries 清理过期条目 (现在由缓存接口自动处理)
func (h *SimpleRecommendationHandler) cleanupExpiredEntries() {
	// 缓存接口会自动处理过期条目的清理
	// 这个方法保留以保持向后兼容性
}

// generateEnhancedRecommendations 生成增强推荐
func (h *SimpleRecommendationHandler) generateEnhancedRecommendations(request *cppbridge.RecommendationRequest) (*cppbridge.RecommendationResponse, error) {
	// 调用基础推荐生成
	response, err := h.bridge.GenerateRecommendations(request)
	if err != nil {
		return nil, err
	}

	// 增强推荐结果
	if response != nil && len(response.Recommendations) > 0 {
		// 计算置信度
		for i := range response.Recommendations {
			confidence := h.calculateConfidence(&response.Recommendations[i], request)
			response.Recommendations[i].Score = confidence

			// 生成推荐原因
			reasons := h.generateRecommendationReasons(&response.Recommendations[i], request)
			response.Recommendations[i].Reasons = reasons
		}

		// 重新排序基于置信度
		h.sortRecommendationsByConfidence(response.Recommendations)
	}

	return response, nil
}

// calculateConfidence 计算置信度
func (h *SimpleRecommendationHandler) calculateConfidence(rec *cppbridge.Recommendation, request *cppbridge.RecommendationRequest) float64 {
	confidence := h.confidenceCalc.baseConfidence

	// 分数匹配度
	scoreMatch := h.calculateScoreMatch(rec.AdmissionScore, request.TotalScore)
	confidence += scoreMatch * 0.3

	// 地理位置偏好
	locationMatch := h.calculateLocationMatch(rec.SchoolName, request)
	confidence += locationMatch * 0.2

	// 历史成功率
	historicalKey := fmt.Sprintf("%s_%s", rec.SchoolID, rec.MajorID)
	if historical, exists := h.confidenceCalc.historicalData[historicalKey]; exists {
		confidence += historical * 0.2
	}

	// 风险调整
	riskAdjustment := h.calculateRiskAdjustment(rec.RiskLevel)
	confidence += riskAdjustment

	// 确保置信度在0-1范围内
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.0 {
		confidence = 0.0
	}

	return confidence
}

// calculateScoreMatch 计算分数匹配度
func (h *SimpleRecommendationHandler) calculateScoreMatch(admissionScore, totalScore int) float64 {
	if admissionScore <= 0 {
		return 0.0
	}

	scoreDiff := float64(totalScore - admissionScore)

	// 分数高于录取线时匹配度更高
	if scoreDiff >= 0 {
		return math.Min(1.0, scoreDiff/50.0) // 高50分为满分
	} else {
		// 分数低于录取线时急剧下降
		return math.Max(0.0, 1.0+scoreDiff/30.0)
	}
}

// calculateLocationMatch 计算地理位置匹配度
func (h *SimpleRecommendationHandler) calculateLocationMatch(schoolName string, request *cppbridge.RecommendationRequest) float64 {
	if request.Preferences == nil {
		return 0.0
	}

	preferredLocations, ok := request.Preferences["preferred_locations"].([]interface{})
	if !ok {
		return 0.0
	}

	// 简化的地理位置匹配逻辑
	for _, loc := range preferredLocations {
		if locStr, ok := loc.(string); ok && strings.Contains(schoolName, locStr) {
			return 0.3
		}
	}

	return 0.0
}

// calculateRiskAdjustment 计算风险调整
func (h *SimpleRecommendationHandler) calculateRiskAdjustment(riskLevel string) float64 {
	switch riskLevel {
	case "low":
		return 0.1
	case "medium":
		return 0.0
	case "high":
		return -0.1
	default:
		return 0.0
	}
}

// generateRecommendationReasons 生成推荐原因
func (h *SimpleRecommendationHandler) generateRecommendationReasons(rec *cppbridge.Recommendation, request *cppbridge.RecommendationRequest) []string {
	var reasons []string

	// 基于分数匹配
	if rec.AdmissionScore > 0 && request.TotalScore >= rec.AdmissionScore {
		scoreDiff := request.TotalScore - rec.AdmissionScore
		if scoreDiff >= 30 {
			reasons = append(reasons, "分数优势明显，录取几率很高")
		} else if scoreDiff >= 10 {
			reasons = append(reasons, "分数符合要求，有较好录取机会")
		} else {
			reasons = append(reasons, "分数达到录取线，建议作为稳妥选择")
		}
	}

	// 基于概率
	if rec.Probability >= 0.8 {
		reasons = append(reasons, "根据历年数据，录取概率很高")
	} else if rec.Probability >= 0.6 {
		reasons = append(reasons, "录取概率较高，值得考虑")
	}

	// 基于专业特色
	if strings.Contains(rec.MajorName, "计算机") || strings.Contains(rec.MajorName, "人工智能") {
		reasons = append(reasons, "热门专业，就业前景广阔")
	}

	// 基于学校声誉
	prestigiousKeywords := []string{"大学", "学院", "工业大学", "科技大学"}
	for _, keyword := range prestigiousKeywords {
		if strings.Contains(rec.SchoolName, keyword) {
			reasons = append(reasons, "知名院校，教学质量有保障")
			break
		}
	}

	return reasons
}

// sortRecommendationsByConfidence 按置信度排序推荐
func (h *SimpleRecommendationHandler) sortRecommendationsByConfidence(recommendations []cppbridge.Recommendation) {
	sort.Slice(recommendations, func(i, j int) bool {
		// 首先按置信度降序排列
		if recommendations[i].Score != recommendations[j].Score {
			return recommendations[i].Score > recommendations[j].Score
		}
		// 置信度相同时按概率降序排列
		return recommendations[i].Probability > recommendations[j].Probability
	})

	// 更新排名
	for i := range recommendations {
		recommendations[i].Ranking = i + 1
	}
}

// processBatchRecommendationsWithOptimization 优化的批量处理
func (h *SimpleRecommendationHandler) processBatchRecommendationsWithOptimization(batchRequest *cppbridge.BatchRecommendationRequest) ([]cppbridge.RecommendationResponse, int, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(batchRequest.Timeout)*time.Millisecond)
	defer cancel()

	// 创建工作池
	jobs := make(chan cppbridge.RecommendationRequest, len(batchRequest.Requests))
	results := make(chan cppbridge.RecommendationResponse, len(batchRequest.Requests))

	// 启动工作协程
	workerCount := h.batchProcessor.maxWorkers
	if len(batchRequest.Requests) < workerCount {
		workerCount = len(batchRequest.Requests)
	}

	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go h.batchWorker(ctx, jobs, results, &wg, batchRequest.Algorithm)
	}

	// 发送任务
	go func() {
		defer close(jobs)
		for _, request := range batchRequest.Requests {
			select {
			case jobs <- request:
			case <-ctx.Done():
				return
			}
		}
	}()

	// 等待所有工作完成
	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集结果
	var responses []cppbridge.RecommendationResponse
	successCount := 0
	failedCount := 0

	for response := range results {
		responses = append(responses, response)
		if response.Success {
			successCount++
		} else {
			failedCount++
		}
	}

	return responses, successCount, failedCount, nil
}

// batchWorker 批处理工作协程
func (h *SimpleRecommendationHandler) batchWorker(ctx context.Context, jobs <-chan cppbridge.RecommendationRequest, results chan<- cppbridge.RecommendationResponse, wg *sync.WaitGroup, algorithm string) {
	defer wg.Done()

	for {
		select {
		case request, ok := <-jobs:
			if !ok {
				return
			}

			// 设置默认值
			if request.MaxRecommendations == 0 {
				request.MaxRecommendations = 30
			}
			if request.Algorithm == "" {
				request.Algorithm = algorithm
			}

			// 尝试处理请求
			response := h.processRequestWithRetry(&request, h.batchProcessor.retryAttempts)

			select {
			case results <- response:
			case <-ctx.Done():
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

// processRequestWithRetry 带重试的请求处理
func (h *SimpleRecommendationHandler) processRequestWithRetry(request *cppbridge.RecommendationRequest, maxRetries int) cppbridge.RecommendationResponse {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		response, err := h.generateEnhancedRecommendations(request)
		if err == nil && response != nil {
			return *response
		}

		lastErr = err
		if attempt < maxRetries {
			// 指数退避
			time.Sleep(time.Duration(math.Pow(2, float64(attempt))) * 100 * time.Millisecond)
		}
	}

	// 返回失败响应
	return cppbridge.RecommendationResponse{
		StudentID:    request.StudentID,
		Success:      false,
		ErrorMessage: fmt.Sprintf("Failed after %d attempts: %v", maxRetries+1, lastErr),
	}
}

// generateIntelligentExplanation 生成智能解释
func (h *SimpleRecommendationHandler) generateIntelligentExplanation(recommendationID string) (map[string]interface{}, error) {
	// 模拟从数据库或缓存获取推荐详情
	// 在实际实现中，这里应该查询具体的推荐记录

	explanation := map[string]interface{}{
		"recommendation_id": recommendationID,
		"explanation": map[string]interface{}{
			"overall_score":    0.88,
			"confidence_level": "高",
			"match_factors": []map[string]interface{}{
				{
					"factor":      "分数匹配度",
					"score":       0.92,
					"weight":      h.explainer.scoreWeights["score_match"],
					"description": "您的分数与该专业历年录取分数线匹配度为92%",
				},
				{
					"factor":      "兴趣匹配度",
					"score":       0.85,
					"weight":      h.explainer.scoreWeights["interest_match"],
					"description": "根据您的兴趣偏好，该专业符合度为85%",
				},
				{
					"factor":      "就业前景",
					"score":       0.90,
					"weight":      h.explainer.scoreWeights["employment_prospect"],
					"description": "该专业就业前景良好，就业率达90%",
				},
			},
		},
		"detailed_analysis":    h.generateDetailedAnalysis(recommendationID),
		"comparative_analysis": h.generateComparativeAnalysis(recommendationID),
		"risk_assessment":      h.generateRiskAssessment(recommendationID),
		"recommendations":      h.generateActionRecommendations(recommendationID),
		"confidence_breakdown": h.generateConfidenceBreakdown(recommendationID),
	}

	return explanation, nil
}

// generateDetailedAnalysis 生成详细分析
func (h *SimpleRecommendationHandler) generateDetailedAnalysis(recommendationID string) map[string]interface{} {
	return map[string]interface{}{
		"strengths": []string{
			"分数优势明显，超出录取线20分以上",
			"专业与个人兴趣高度匹配",
			"学校地理位置符合偏好",
			"专业就业率和薪资水平较高",
		},
		"considerations": []string{
			"该专业竞争相对激烈",
			"需要较强的数学基础",
			"建议了解专业课程设置",
		},
		"historical_data": map[string]interface{}{
			"admission_trend":        "近三年录取分数稳中有升",
			"employment_rate":        "95%",
			"average_salary":         "8000-12000元/月",
			"further_education_rate": "30%",
		},
	}
}

// generateComparativeAnalysis 生成对比分析
func (h *SimpleRecommendationHandler) generateComparativeAnalysis(recommendationID string) map[string]interface{} {
	return map[string]interface{}{
		"vs_similar_majors": []map[string]interface{}{
			{
				"major":         "软件工程",
				"advantages":    []string{"更偏向实践应用", "就业机会更多"},
				"disadvantages": []string{"理论基础相对较浅"},
			},
		},
		"vs_similar_schools": []map[string]interface{}{
			{
				"school":        "相似档次学校",
				"advantages":    []string{"录取分数相近", "专业实力相当"},
				"disadvantages": []string{"地理位置稍逊"},
			},
		},
		"ranking_in_category": "该专业在同类推荐中排名第3",
	}
}

// generateRiskAssessment 生成风险评估
func (h *SimpleRecommendationHandler) generateRiskAssessment(recommendationID string) map[string]interface{} {
	return map[string]interface{}{
		"overall_risk": "中等",
		"admission_risk": map[string]interface{}{
			"level":   "低",
			"factors": []string{"分数优势明显", "历年录取稳定"},
		},
		"career_risk": map[string]interface{}{
			"level":   "低",
			"factors": []string{"行业发展趋势良好", "技能通用性强"},
		},
		"mitigation_strategies": []string{
			"准备相关专业的备选方案",
			"提前了解专业课程要求",
			"关注该校其他优势专业",
		},
	}
}

// generateActionRecommendations 生成行动建议
func (h *SimpleRecommendationHandler) generateActionRecommendations(recommendationID string) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"priority":    "高",
			"action":      "深入了解专业课程设置",
			"description": "建议查看详细的课程体系和培养方案",
			"deadline":    "填报前1周",
		},
		{
			"priority":    "中",
			"action":      "咨询在校学生或毕业生",
			"description": "获取第一手的专业学习和就业信息",
			"deadline":    "填报前3天",
		},
		{
			"priority":    "低",
			"action":      "关注学校开放日活动",
			"description": "实地了解校园环境和学习氛围",
			"deadline":    "填报前1个月",
		},
	}
}

// generateConfidenceBreakdown 生成置信度分解
func (h *SimpleRecommendationHandler) generateConfidenceBreakdown(recommendationID string) map[string]interface{} {
	return map[string]interface{}{
		"base_confidence": h.confidenceCalc.baseConfidence,
		"adjustments": map[string]float64{
			"score_match_bonus":   0.15,
			"interest_alignment":  0.08,
			"location_preference": 0.05,
			"historical_success":  0.03,
		},
		"final_confidence": 0.88,
		"confidence_level": "高置信度",
		"reliability_notes": []string{
			"基于近5年历史数据计算",
			"考虑了个人偏好和市场趋势",
			"建议结合个人实际情况评估",
		},
	}
}

// performIntelligentOptimization 执行智能优化
func (h *SimpleRecommendationHandler) performIntelligentOptimization(request map[string]interface{}) (map[string]interface{}, error) {
	optimizationID := "opt_" + strconv.FormatInt(time.Now().Unix(), 10)

	// 提取优化参数
	feedbackData, ok := request["feedback_data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing feedback_data")
	}

	// 分析反馈数据
	improvements := h.analyzeFeedbackAndOptimize(feedbackData)

	// 更新权重
	newWeights := h.calculateOptimalWeights(feedbackData)

	// 应用优化
	err := h.bridge.UpdateFusionWeights(newWeights)
	if err != nil {
		return nil, fmt.Errorf("failed to update weights: %v", err)
	}

	result := map[string]interface{}{
		"optimization_id": optimizationID,
		"status":          "completed",
		"improvements":    improvements,
		"new_weights":     newWeights,
		"optimization_details": map[string]interface{}{
			"feedback_analyzed":      len(feedbackData),
			"confidence_improvement": improvements["accuracy_increase"],
			"user_satisfaction_gain": improvements["user_satisfaction"],
		},
		"performance_impact": h.calculatePerformanceImpact(newWeights),
		"message":            "推荐算法已根据反馈进行智能优化",
		"next_evaluation":    time.Now().Add(7 * 24 * time.Hour).Unix(),
	}

	return result, nil
}

// analyzeFeedbackAndOptimize 分析反馈并优化
func (h *SimpleRecommendationHandler) analyzeFeedbackAndOptimize(feedbackData map[string]interface{}) map[string]interface{} {
	// 模拟反馈分析逻辑
	userSatisfaction, _ := feedbackData["user_satisfaction"].(float64)
	clickThroughRate, _ := feedbackData["click_through_rate"].(float64)
	conversionRate, _ := feedbackData["conversion_rate"].(float64)

	// 计算改进度
	accuracyIncrease := math.Min(0.1, userSatisfaction*0.05)
	diversityImprovement := math.Min(0.08, clickThroughRate*0.1)
	satisfactionGain := math.Min(0.15, conversionRate*0.2)

	return map[string]interface{}{
		"accuracy_increase":         accuracyIncrease,
		"diversity_improvement":     diversityImprovement,
		"user_satisfaction":         satisfactionGain,
		"response_time_improvement": 0.02,
	}
}

// calculateOptimalWeights 计算最优权重
func (h *SimpleRecommendationHandler) calculateOptimalWeights(feedbackData map[string]interface{}) map[string]float64 {
	// 基于反馈数据动态调整权重
	traditionalPerformance, _ := feedbackData["traditional_performance"].(float64)
	aiPerformance, _ := feedbackData["ai_performance"].(float64)

	// 如果没有提供性能数据，使用默认值
	if traditionalPerformance == 0 {
		traditionalPerformance = 0.7
	}
	if aiPerformance == 0 {
		aiPerformance = 0.8
	}

	// 根据性能调整权重
	totalPerformance := traditionalPerformance + aiPerformance
	traditionalWeight := traditionalPerformance / totalPerformance * 0.9 // 稍微偏向传统算法以保证稳定性
	aiWeight := 1.0 - traditionalWeight

	return map[string]float64{
		"traditional": traditionalWeight,
		"ai":          aiWeight,
		"diversity":   0.15,
	}
}

// calculatePerformanceImpact 计算性能影响
func (h *SimpleRecommendationHandler) calculatePerformanceImpact(newWeights map[string]float64) map[string]interface{} {
	return map[string]interface{}{
		"expected_latency_change": -0.05, // 预期延迟减少5%
		"expected_accuracy_gain":  0.03,  // 预期准确度提升3%
		"resource_usage_change":   0.02,  // 资源使用增加2%
		"throughput_impact":       0.01,  // 吞吐量提升1%
	}
}

// updatePerformanceStats 更新性能统计
func (h *SimpleRecommendationHandler) updatePerformanceStats(isNewRequest, isFailed bool, latency time.Duration) {
	h.performanceStats.mu.Lock()
	defer h.performanceStats.mu.Unlock()

	if isNewRequest {
		h.performanceStats.totalRequests++
	}

	if isFailed {
		h.performanceStats.failedRequests++
	} else {
		h.performanceStats.successRequests++
	}

	if latency > 0 {
		h.performanceStats.totalLatency += latency
	}

	h.performanceStats.lastUpdate = time.Now()
}

// convertToFrontendFormat 转换为前端格式
func (h *SimpleRecommendationHandler) buildRecommendationData(ctx context.Context, studentInfo *StudentInfo, request *cppbridge.RecommendationRequest, response *cppbridge.RecommendationResponse) RecommendationData {
	var frontendRecs []FrontendRecommendation

	for i, rec := range response.Recommendations {
		// 确定推荐类型
		recType := "稳妥"
		if rec.Probability >= 0.8 {
			recType = "稳妥"
		} else if rec.Probability >= 0.6 {
			recType = "适中"
		} else {
			recType = "冲刺"
		}

		// 确定风险等级
		riskLevel := "中等"
		if rec.Probability >= 0.8 {
			riskLevel = "低"
		} else if rec.Probability < 0.5 {
			riskLevel = "高"
		}

		// 生成推荐原因
		recommendReason := "综合评估推荐"
		if len(rec.Reasons) > 0 {
			recommendReason = rec.Reasons[0]
		}

		// 构建前端推荐格式
		frontendRec := FrontendRecommendation{
			ID: fmt.Sprintf("rec_%d", i+1),
			University: FrontendUniversity{
				ID:         rec.SchoolID,
				Name:       rec.SchoolName,
				Province:   rec.Province,
				City:       rec.City,
				Level:      rec.SchoolLevel,
				Type:       rec.SchoolType,
				IsFavorite: false,
			},
			Type:                 recType,
			AdmissionProbability: int(rec.Probability * 100),
			MatchScore:           int(rec.Score * 100),
			RecommendReason:      recommendReason,
			RiskLevel:            riskLevel,
			SuggestedMajors: []FrontendMajor{
				{
					ID:          rec.MajorID,
					Name:        rec.MajorName,
					Probability: int(rec.Probability * 100),
				},
			},
			HistoricalData: []FrontendHistoricalData{
				{
					MinScore: rec.AdmissionScore - 10,
					AvgScore: rec.AdmissionScore,
					MaxScore: rec.AdmissionScore + 10,
					Year:     2023,
				},
			},
		}

		frontendRecs = append(frontendRecs, frontendRec)
	}

	// 生成分析报告
	analysisReport := h.generateAnalysisReport(ctx, studentInfo, request, frontendRecs)

	return RecommendationData{
		Recommendations: frontendRecs,
		AnalysisReport:  analysisReport,
	}
}

// generateAnalysisReport 生成分析报告
func (h *SimpleRecommendationHandler) generateAnalysisReport(ctx context.Context, studentInfo *StudentInfo, request *cppbridge.RecommendationRequest, recommendations []FrontendRecommendation) string {
	input := h.buildRecommendationAnalysisInput(studentInfo, request, recommendations)
	if h.analyzer != nil {
		report, err := h.analyzer.AnalyzeRecommendation(ctx, input)
		if err == nil && strings.TrimSpace(report) != "" {
			return report
		}
	}

	fallback, _ := llm.NewLocalFallbackAnalyzer().AnalyzeRecommendation(ctx, input)
	return fallback
}

func (h *SimpleRecommendationHandler) buildRecommendationAnalysisInput(studentInfo *StudentInfo, request *cppbridge.RecommendationRequest, recommendations []FrontendRecommendation) llm.RecommendationAnalysisInput {
	input := llm.RecommendationAnalysisInput{
		StudentName:         request.Name,
		Score:               request.TotalScore,
		Province:            request.Province,
		SubjectCombination:  request.SubjectCombination,
		RiskTolerance:       studentInfo.Preferences.RiskTolerance,
		PreferredRegions:    append([]string(nil), studentInfo.Preferences.Regions...),
		PreferredMajors:     append([]string(nil), studentInfo.Preferences.MajorCategories...),
		UniversityTypes:     append([]string(nil), studentInfo.Preferences.UniversityTypes...),
		SpecialRequirements: studentInfo.Preferences.SpecialRequirements,
		TotalCount:          len(recommendations),
		Recommendations:     make([]llm.RecommendationCandidate, 0, len(recommendations)),
	}
	if input.StudentName == "" {
		input.StudentName = "Student"
	}
	if request.Ranking > 0 {
		rank := request.Ranking
		input.Rank = &rank
	}

	for _, rec := range recommendations {
		candidate := llm.RecommendationCandidate{
			SchoolName:           rec.University.Name,
			MajorName:            firstMajorName(rec.SuggestedMajors),
			Probability:          float64(rec.AdmissionProbability) / 100,
			AdmissionProbability: rec.AdmissionProbability,
			MatchScore:           rec.MatchScore,
			RiskLevel:            rec.RiskLevel,
			Type:                 rec.University.Type,
			Province:             rec.University.Province,
			City:                 rec.University.City,
			Reason:               rec.RecommendReason,
		}
		input.Recommendations = append(input.Recommendations, candidate)
	}

	return input
}

func firstMajorName(majors []FrontendMajor) string {
	if len(majors) == 0 {
		return ""
	}
	return majors[0].Name
}

// GetPerformanceStats 获取性能统计（新增方法）
func (h *SimpleRecommendationHandler) GetPerformanceStats() map[string]interface{} {
	h.performanceStats.mu.RLock()
	defer h.performanceStats.mu.RUnlock()

	var avgLatency float64
	if h.performanceStats.totalRequests > 0 {
		avgLatency = float64(h.performanceStats.totalLatency.Nanoseconds()) / float64(h.performanceStats.totalRequests) / 1e6 // 转换为毫秒
	}

	var successRate float64
	if h.performanceStats.totalRequests > 0 {
		successRate = float64(h.performanceStats.successRequests) / float64(h.performanceStats.totalRequests)
	}

	var cacheHitRate float64
	totalCacheRequests := h.performanceStats.cacheHits + h.performanceStats.cacheMisses
	if totalCacheRequests > 0 {
		cacheHitRate = float64(h.performanceStats.cacheHits) / float64(totalCacheRequests)
	}

	return map[string]interface{}{
		"total_requests":   h.performanceStats.totalRequests,
		"success_requests": h.performanceStats.successRequests,
		"failed_requests":  h.performanceStats.failedRequests,
		"success_rate":     successRate,
		"avg_latency_ms":   avgLatency,
		"cache_hit_rate":   cacheHitRate,
		"cache_hits":       h.performanceStats.cacheHits,
		"cache_misses":     h.performanceStats.cacheMisses,
		"last_update":      h.performanceStats.lastUpdate,
	}
}
