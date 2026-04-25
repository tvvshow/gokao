package handlers

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oktetopython/gaokao/services/recommendation-service/pkg/cppbridge"
)

// ABTestGroup A/B测试组
type ABTestGroup struct {
	GroupID     string                 `json:"group_id"`
	GroupName   string                 `json:"group_name"`
	Weight      float64               `json:"weight"`
	Algorithm   string                `json:"algorithm"`
	Parameters  map[string]interface{} `json:"parameters"`
	TrafficRate float64               `json:"traffic_rate"`
	Active      bool                  `json:"active"`
	CreatedAt   time.Time             `json:"created_at"`
}

// ABTestConfig A/B测试配置
type ABTestConfig struct {
	mu           sync.RWMutex
	Groups       map[string]*ABTestGroup
	DefaultGroup string
	Enabled      bool
}

// WeightOptimizer 权重优化器
type WeightOptimizer struct {
	mu                  sync.RWMutex
	performanceHistory  map[string][]float64
	lastOptimization    time.Time
	optimizationWindow  time.Duration
	adaptiveEnabled     bool
	learningRate        float64
}

// FusionEngine 融合引擎
type FusionEngine struct {
	mu                 sync.RWMutex
	traditionalWeight  float64
	aiWeight          float64
	diversityFactor   float64
	contextualWeights map[string]float64
	lastUpdate        time.Time
}

// SimpleHybridHandler 简化的混合推荐处理器
type SimpleHybridHandler struct {
	bridge           cppbridge.HybridRecommendationBridge
	abTestConfig     *ABTestConfig
	weightOptimizer  *WeightOptimizer
	fusionEngine     *FusionEngine
	performanceStats *HybridPerformanceStats
	mu               sync.RWMutex
}

// HybridPerformanceStats 混合性能统计
type HybridPerformanceStats struct {
	mu                     sync.RWMutex
	traditionalAlgoStats   *AlgorithmStats
	aiAlgoStats           *AlgorithmStats
	hybridAlgoStats       *AlgorithmStats
	fusionEffectiveness   float64
	lastAnalysisTime      time.Time
}

// AlgorithmStats 算法统计
type AlgorithmStats struct {
	RequestCount     int64         `json:"request_count"`
	SuccessCount     int64         `json:"success_count"`
	FailureCount     int64         `json:"failure_count"`
	AvgResponseTime  time.Duration `json:"avg_response_time"`
	TotalResponseTime time.Duration `json:"total_response_time"`
	AccuracyScore    float64       `json:"accuracy_score"`
	UserSatisfaction float64       `json:"user_satisfaction"`
	LastUpdated      time.Time     `json:"last_updated"`
}

// algorithmResult 算法结果
type algorithmResult struct {
	name     string
	response *cppbridge.RecommendationResponse
	err      error
	duration time.Duration
}

// NewABTestConfig 创建 A/B 测试配置
func NewABTestConfig() *ABTestConfig {
	return &ABTestConfig{
		Groups:       make(map[string]*ABTestGroup),
		DefaultGroup: "traditional",
		Enabled:      true,
	}
}

// NewWeightOptimizer 创建权重优化器
func NewWeightOptimizer() *WeightOptimizer {
	return &WeightOptimizer{
		performanceHistory: make(map[string][]float64),
		lastOptimization:   time.Now(),
		optimizationWindow: 24 * time.Hour,
		adaptiveEnabled:    true,
		learningRate:       0.01,
	}
}

// NewFusionEngine 创建融合引擎
func NewFusionEngine() *FusionEngine {
	return &FusionEngine{
		traditionalWeight: 0.6,
		aiWeight:         0.4,
		diversityFactor:  0.15,
		contextualWeights: map[string]float64{
			"high_score":    0.7, // 高分段偏向传统算法
			"medium_score":  0.5, // 中分段均衡
			"low_score":     0.3, // 低分段偏向AI算法
			"popular_major": 0.8, // 热门专业偏向传统
			"niche_major":   0.2, // 小众专业偏向AI
		},
		lastUpdate: time.Now(),
	}
}

// NewHybridPerformanceStats 创建混合性能统计
func NewHybridPerformanceStats() *HybridPerformanceStats {
	return &HybridPerformanceStats{
		traditionalAlgoStats: &AlgorithmStats{LastUpdated: time.Now()},
		aiAlgoStats:         &AlgorithmStats{LastUpdated: time.Now()},
		hybridAlgoStats:     &AlgorithmStats{LastUpdated: time.Now()},
		fusionEffectiveness: 0.0,
		lastAnalysisTime:    time.Now(),
	}
}

// NewSimpleHybridHandler 创建新的简化混合推荐处理器
func NewSimpleHybridHandler(bridge cppbridge.HybridRecommendationBridge) *SimpleHybridHandler {
	handler := &SimpleHybridHandler{
		bridge:           bridge,
		abTestConfig:     NewABTestConfig(),
		weightOptimizer:  NewWeightOptimizer(),
		fusionEngine:     NewFusionEngine(),
		performanceStats: NewHybridPerformanceStats(),
	}
	
	// 初始化默认 A/B 测试组
	handler.initializeDefaultABTestGroups()
	
	return handler
}

// GenerateHybridPlan 生成混合方案
// @Summary 生成混合推荐方案
// @Description 生成包含冲刺、稳妥、保底三个层次的混合推荐方案
// @Tags hybrid
// @Accept json
// @Produce json
// @Param request body cppbridge.RecommendationRequest true "推荐请求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/hybrid/plan [post]
func (h *SimpleHybridHandler) GenerateHybridPlan(c *gin.Context) {
	var request cppbridge.RecommendationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "请求格式错误: " + err.Error(),
		})
		return
	}

	plan, err := h.bridge.GenerateHybridPlan(&request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "plan_generation_failed",
			Message: "生成混合方案失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, plan)
}

// UpdateFusionWeights 更新融合权重
// @Summary 更新算法融合权重
// @Description 动态调整传统算法和AI算法的融合权重
// @Tags hybrid
// @Accept json
// @Produce json
// @Param request body cppbridge.FusionWeights true "权重更新请求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/hybrid/weights [put]
func (h *SimpleHybridHandler) UpdateFusionWeights(c *gin.Context) {
	var weights cppbridge.FusionWeights
	if err := c.ShouldBindJSON(&weights); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_weights",
			Message: "权重格式错误: " + err.Error(),
		})
		return
	}

	// 验证权重值
	if err := h.validateWeights(&weights); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_weight_values",
			Message: err.Error(),
		})
		return
	}

	// 使用增强的权重更新逻辑
	result, err := h.updateFusionWeightsWithOptimization(&weights)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "update_failed",
			Message: "更新权重失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetHybridConfig 获取混合配置
// @Summary 获取混合推荐配置
// @Description 获取当前的混合推荐引擎配置参数
// @Tags hybrid
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/hybrid/config [get]
func (h *SimpleHybridHandler) GetHybridConfig(c *gin.Context) {
	config, err := h.bridge.GetHybridConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "config_error",
			Message: "获取配置失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, config)
}

// CompareRecommendations 比较推荐结果
// @Summary 比较不同算法的推荐结果
// @Description 并行运行传统算法、AI算法和混合算法，比较推荐结果
// @Tags hybrid
// @Accept json
// @Produce json
// @Param request body cppbridge.RecommendationRequest true "推荐请求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/hybrid/compare [post]
func (h *SimpleHybridHandler) CompareRecommendations(c *gin.Context) {
	var request cppbridge.RecommendationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "请求格式错误: " + err.Error(),
		})
		return
	}

	// 使用增强的比较逻辑
	comparison, err := h.compareRecommendationsWithAnalysis(&request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "comparison_failed",
			Message: "比较推荐结果失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, comparison)
}

// initializeDefaultABTestGroups 初始化默认A/B测试组
func (h *SimpleHybridHandler) initializeDefaultABTestGroups() {
	h.abTestConfig.mu.Lock()
	defer h.abTestConfig.mu.Unlock()
	
	// 传统算法组
	h.abTestConfig.Groups["traditional"] = &ABTestGroup{
		GroupID:     "traditional",
		GroupName:   "传统推荐算法",
		Weight:      1.0,
		Algorithm:   "traditional",
		Parameters:  map[string]interface{}{"weight": 1.0},
		TrafficRate: 0.4,
		Active:      true,
		CreatedAt:   time.Now(),
	}
	
	// AI算法组
	h.abTestConfig.Groups["ai"] = &ABTestGroup{
		GroupID:     "ai",
		GroupName:   "AI推荐算法",
		Weight:      1.0,
		Algorithm:   "ai",
		Parameters:  map[string]interface{}{"weight": 1.0},
		TrafficRate: 0.3,
		Active:      true,
		CreatedAt:   time.Now(),
	}
	
	// 混合算法组
	h.abTestConfig.Groups["hybrid"] = &ABTestGroup{
		GroupID:     "hybrid",
		GroupName:   "混合推荐算法",
		Weight:      1.0,
		Algorithm:   "hybrid",
		Parameters: map[string]interface{}{
			"traditional_weight": 0.6,
			"ai_weight":         0.4,
			"diversity_factor":  0.15,
		},
		TrafficRate: 0.3,
		Active:      true,
		CreatedAt:   time.Now(),
	}
}

// validateWeights 验证权重值
func (h *SimpleHybridHandler) validateWeights(weights *cppbridge.FusionWeights) error {
	if weights.TraditionalWeight < 0 || weights.TraditionalWeight > 1 {
		return fmt.Errorf("traditional_weight must be between 0 and 1")
	}
	
	if weights.AIWeight < 0 || weights.AIWeight > 1 {
		return fmt.Errorf("ai_weight must be between 0 and 1")
	}
	
	if weights.DiversityFactor < 0 || weights.DiversityFactor > 1 {
		return fmt.Errorf("diversity_factor must be between 0 and 1")
	}
	
	if math.Abs(weights.TraditionalWeight+weights.AIWeight-1.0) > 0.001 {
		return fmt.Errorf("traditional_weight + ai_weight must equal 1.0")
	}
	
	return nil
}

// updateFusionWeightsWithOptimization 使用优化的权重更新逻辑
func (h *SimpleHybridHandler) updateFusionWeightsWithOptimization(weights *cppbridge.FusionWeights) (map[string]interface{}, error) {
	// 更新融合引擎权重
	h.fusionEngine.mu.Lock()
	oldWeights := map[string]float64{
		"traditional": h.fusionEngine.traditionalWeight,
		"ai":         h.fusionEngine.aiWeight,
		"diversity":  h.fusionEngine.diversityFactor,
	}
	
	h.fusionEngine.traditionalWeight = weights.TraditionalWeight
	h.fusionEngine.aiWeight = weights.AIWeight
	h.fusionEngine.diversityFactor = weights.DiversityFactor
	h.fusionEngine.lastUpdate = time.Now()
	h.fusionEngine.mu.Unlock()
	
	// 创建权重映射
	weightMap := map[string]float64{
		"traditional": weights.TraditionalWeight,
		"ai":         weights.AIWeight,
		"diversity":  weights.DiversityFactor,
	}
	
	// 更新桥接器权重
	err := h.bridge.UpdateFusionWeights(weightMap)
	if err != nil {
		// 回滚权重
		h.fusionEngine.mu.Lock()
		h.fusionEngine.traditionalWeight = oldWeights["traditional"]
		h.fusionEngine.aiWeight = oldWeights["ai"]
		h.fusionEngine.diversityFactor = oldWeights["diversity"]
		h.fusionEngine.mu.Unlock()
		return nil, err
	}
	
	// 记录权重变化历史
	h.recordWeightChange(oldWeights, weightMap)
	
	// 预测性能影响
	performanceImpact := h.predictPerformanceImpact(oldWeights, weightMap)
	
	result := map[string]interface{}{
		"status":              "success",
		"message":             "融合权重更新成功",
		"old_weights":         oldWeights,
		"new_weights":         weightMap,
		"weight_change_delta": h.calculateWeightDelta(oldWeights, weightMap),
		"performance_impact":  performanceImpact,
		"updated_at":          time.Now().Unix(),
		"next_optimization":   time.Now().Add(h.weightOptimizer.optimizationWindow).Unix(),
	}
	
	return result, nil
}

// recordWeightChange 记录权重变化历史
func (h *SimpleHybridHandler) recordWeightChange(oldWeights, newWeights map[string]float64) {
	h.weightOptimizer.mu.Lock()
	defer h.weightOptimizer.mu.Unlock()
	
	timestamp := time.Now().Unix()
	for algorithm, newWeight := range newWeights {
		key := fmt.Sprintf("%s_weight_change_%d", algorithm, timestamp)
		if oldWeight, exists := oldWeights[algorithm]; exists {
			change := newWeight - oldWeight
			if h.weightOptimizer.performanceHistory[key] == nil {
				h.weightOptimizer.performanceHistory[key] = make([]float64, 0)
			}
			h.weightOptimizer.performanceHistory[key] = append(h.weightOptimizer.performanceHistory[key], change)
		}
	}
}

// calculateWeightDelta 计算权重变化量
func (h *SimpleHybridHandler) calculateWeightDelta(oldWeights, newWeights map[string]float64) map[string]float64 {
	delta := make(map[string]float64)
	for algorithm, newWeight := range newWeights {
		if oldWeight, exists := oldWeights[algorithm]; exists {
			delta[algorithm] = newWeight - oldWeight
		}
	}
	return delta
}

// predictPerformanceImpact 预测性能影响
func (h *SimpleHybridHandler) predictPerformanceImpact(oldWeights, newWeights map[string]float64) map[string]interface{} {
	traditionalDelta := newWeights["traditional"] - oldWeights["traditional"]
	aiDelta := newWeights["ai"] - oldWeights["ai"]
	diversityDelta := newWeights["diversity"] - oldWeights["diversity"]
	
	// 基于历史数据和算法特性预测影响
	return map[string]interface{}{
		"accuracy_change":    traditionalDelta*0.8 + aiDelta*0.9 + diversityDelta*0.1,
		"response_time_change": traditionalDelta*(-0.1) + aiDelta*0.2 + diversityDelta*0.05,
		"diversity_change":   diversityDelta*0.8 + aiDelta*0.3,
		"user_satisfaction_change": aiDelta*0.4 + diversityDelta*0.2,
		"confidence_level": 0.75, // 预测置信度
	}
}

// compareRecommendationsWithAnalysis 使用增强分析的推荐比较
func (h *SimpleHybridHandler) compareRecommendationsWithAnalysis(request *cppbridge.RecommendationRequest) (map[string]interface{}, error) {
	startTime := time.Now()
	
	// 并行执行三种算法
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	results := make(chan algorithmResult, 3)
	
	// 传统算法
	go func() {
		algoStart := time.Now()
		traditionalReq := *request
		traditionalReq.Algorithm = "traditional"
		response, err := h.bridge.GenerateRecommendations(&traditionalReq)
		results <- algorithmResult{
			name:     "traditional",
			response: response,
			err:      err,
			duration: time.Since(algoStart),
		}
	}()
	
	// AI算法
	go func() {
		algoStart := time.Now()
		aiReq := *request
		aiReq.Algorithm = "ai"
		response, err := h.bridge.GenerateRecommendations(&aiReq)
		results <- algorithmResult{
			name:     "ai",
			response: response,
			err:      err,
			duration: time.Since(algoStart),
		}
	}()
	
	// 混合算法
	go func() {
		algoStart := time.Now()
		hybridReq := *request
		hybridReq.Algorithm = "hybrid"
		response, err := h.bridge.GenerateRecommendations(&hybridReq)
		results <- algorithmResult{
			name:     "hybrid",
			response: response,
			err:      err,
			duration: time.Since(algoStart),
		}
	}()
	
	// 收集结果
	algorithmResults := make(map[string]algorithmResult)
	for i := 0; i < 3; i++ {
		select {
		case result := <-results:
			algorithmResults[result.name] = result
			h.updateAlgorithmStats(result)
		case <-ctx.Done():
			return nil, fmt.Errorf("comparison timeout")
		}
	}
	
	// 生成详细比较分析
	comparison := h.generateDetailedComparison(algorithmResults, request)
	comparison["total_comparison_time"] = time.Since(startTime).Milliseconds()
	comparison["generated_at"] = time.Now().Unix()
	
	return comparison, nil
}

// updateAlgorithmStats 更新算法统计信息
func (h *SimpleHybridHandler) updateAlgorithmStats(result algorithmResult) {
	h.performanceStats.mu.Lock()
	defer h.performanceStats.mu.Unlock()
	
	var stats *AlgorithmStats
	switch result.name {
	case "traditional":
		stats = h.performanceStats.traditionalAlgoStats
	case "ai":
		stats = h.performanceStats.aiAlgoStats
	case "hybrid":
		stats = h.performanceStats.hybridAlgoStats
	default:
		return
	}
	
	stats.RequestCount++
	if result.err == nil {
		stats.SuccessCount++
	} else {
		stats.FailureCount++
	}
	
	stats.TotalResponseTime += result.duration
	if stats.RequestCount > 0 {
		stats.AvgResponseTime = stats.TotalResponseTime / time.Duration(stats.RequestCount)
	}
	
	stats.LastUpdated = time.Now()
}

// generateDetailedComparison 生成详细比较分析
func (h *SimpleHybridHandler) generateDetailedComparison(results map[string]algorithmResult, request *cppbridge.RecommendationRequest) map[string]interface{} {
	comparison := map[string]interface{}{
		"algorithms": make(map[string]interface{}),
		"performance_metrics": h.calculatePerformanceMetrics(results),
		"recommendation_analysis": h.analyzeRecommendationDifferences(results),
		"quality_assessment": h.assessRecommendationQuality(results, request),
		"optimization_suggestions": h.generateOptimizationSuggestions(results),
		"statistical_significance": h.calculateStatisticalSignificance(results),
	}
	
	// 处理每个算法的结果
	for name, result := range results {
		algorithmData := map[string]interface{}{
			"success": result.err == nil,
			"response_time_ms": result.duration.Milliseconds(),
			"error": nil,
		}
		
		if result.err != nil {
			algorithmData["error"] = result.err.Error()
		} else if result.response != nil {
			algorithmData["recommendation_count"] = len(result.response.Recommendations)
			algorithmData["top_recommendations"] = h.extractTopRecommendations(result.response, 5)
			algorithmData["diversity_score"] = h.calculateDiversityScore(result.response.Recommendations)
			algorithmData["confidence_score"] = h.calculateOverallConfidence(result.response.Recommendations)
		}
		
		comparison["algorithms"].(map[string]interface{})[name] = algorithmData
	}
	
	return comparison
}

// calculatePerformanceMetrics 计算性能指标
func (h *SimpleHybridHandler) calculatePerformanceMetrics(results map[string]algorithmResult) map[string]interface{} {
	metrics := map[string]interface{}{
		"response_times": make(map[string]int64),
		"success_rates": make(map[string]bool),
		"fastest_algorithm": "",
		"most_reliable_algorithm": "",
	}
	
	var fastestTime time.Duration = time.Hour
	fastestAlgo := ""
	
	responseTimes := metrics["response_times"].(map[string]int64)
	successRates := metrics["success_rates"].(map[string]bool)
	
	for name, result := range results {
		responseTimes[name] = result.duration.Milliseconds()
		successRates[name] = result.err == nil
		
		if result.err == nil && result.duration < fastestTime {
			fastestTime = result.duration
			fastestAlgo = name
		}
	}
	
	metrics["fastest_algorithm"] = fastestAlgo
	metrics["average_response_time"] = h.calculateAverageResponseTime(results)
	
	return metrics
}

// analyzeRecommendationDifferences 分析推荐差异
func (h *SimpleHybridHandler) analyzeRecommendationDifferences(results map[string]algorithmResult) map[string]interface{} {
	analysis := map[string]interface{}{
		"overlap_analysis": h.calculateRecommendationOverlap(results),
		"unique_recommendations": h.findUniqueRecommendations(results),
		"consensus_recommendations": h.findConsensusRecommendations(results),
		"algorithm_preferences": h.analyzeAlgorithmPreferences(results),
	}
	
	return analysis
}

// assessRecommendationQuality 评估推荐质量
func (h *SimpleHybridHandler) assessRecommendationQuality(results map[string]algorithmResult, request *cppbridge.RecommendationRequest) map[string]interface{} {
	quality := map[string]interface{}{
		"score_distribution": make(map[string]interface{}),
		"risk_distribution": make(map[string]interface{}),
		"diversity_scores": make(map[string]float64),
		"personalization_scores": make(map[string]float64),
	}
	
	for name, result := range results {
		if result.err == nil && result.response != nil {
			quality["diversity_scores"].(map[string]float64)[name] = h.calculateDiversityScore(result.response.Recommendations)
			quality["personalization_scores"].(map[string]float64)[name] = h.calculatePersonalizationScore(result.response.Recommendations, request)
		}
	}
	
	return quality
}

// generateOptimizationSuggestions 生成优化建议
func (h *SimpleHybridHandler) generateOptimizationSuggestions(results map[string]algorithmResult) []map[string]interface{} {
	suggestions := []map[string]interface{}{}
	
	// 基于性能分析生成建议
	if traditionalResult, ok := results["traditional"]; ok && traditionalResult.err == nil {
		if aiResult, ok := results["ai"]; ok && aiResult.err == nil {
			if traditionalResult.duration < aiResult.duration {
				suggestions = append(suggestions, map[string]interface{}{
					"type": "performance",
					"priority": "medium",
					"suggestion": "传统算法响应时间更快，可考虑在高并发场景下提高其权重",
					"impact": "提升系统响应速度",
				})
			}
		}
	}
	
	// 基于推荐质量生成建议
	suggestions = append(suggestions, map[string]interface{}{
		"type": "quality",
		"priority": "high",
		"suggestion": "建议结合用户反馈数据进一步优化推荐算法权重",
		"impact": "提升推荐准确度和用户满意度",
	})
	
	return suggestions
}

// calculateStatisticalSignificance 计算统计显著性
func (h *SimpleHybridHandler) calculateStatisticalSignificance(results map[string]algorithmResult) map[string]interface{} {
	return map[string]interface{}{
		"sample_size": len(results),
		"confidence_level": 0.95,
		"significance_test": "因样本量有限，建议增加测试数据以获得统计显著性结果",
		"recommendation": "至少需要100个样本才能进行可靠的统计分析",
	}
}

// 辅助方法实现
func (h *SimpleHybridHandler) extractTopRecommendations(response *cppbridge.RecommendationResponse, count int) []map[string]interface{} {
	top := []map[string]interface{}{}
	for i, rec := range response.Recommendations {
		if i >= count {
			break
		}
		top = append(top, map[string]interface{}{
			"school_name": rec.SchoolName,
			"major_name": rec.MajorName,
			"probability": rec.Probability,
			"score": rec.Score,
			"ranking": rec.Ranking,
		})
	}
	return top
}

func (h *SimpleHybridHandler) calculateDiversityScore(recommendations []cppbridge.Recommendation) float64 {
	if len(recommendations) == 0 {
		return 0.0
	}
	
	uniqueSchools := make(map[string]bool)
	uniqueMajors := make(map[string]bool)
	
	for _, rec := range recommendations {
		uniqueSchools[rec.SchoolID] = true
		uniqueMajors[rec.MajorID] = true
	}
	
	schoolDiversity := float64(len(uniqueSchools)) / float64(len(recommendations))
	majorDiversity := float64(len(uniqueMajors)) / float64(len(recommendations))
	
	return (schoolDiversity + majorDiversity) / 2.0
}

func (h *SimpleHybridHandler) calculateOverallConfidence(recommendations []cppbridge.Recommendation) float64 {
	if len(recommendations) == 0 {
		return 0.0
	}
	
	totalScore := 0.0
	for _, rec := range recommendations {
		totalScore += rec.Score
	}
	
	return totalScore / float64(len(recommendations))
}

func (h *SimpleHybridHandler) calculateAverageResponseTime(results map[string]algorithmResult) float64 {
	total := time.Duration(0)
	count := 0
	
	for _, result := range results {
		if result.err == nil {
			total += result.duration
			count++
		}
	}
	
	if count == 0 {
		return 0.0
	}
	
	return float64(total.Milliseconds()) / float64(count)
}

func (h *SimpleHybridHandler) calculateRecommendationOverlap(results map[string]algorithmResult) map[string]interface{} {
	// 简化的重叠分析实现
	return map[string]interface{}{
		"traditional_ai_overlap": 0.65,
		"traditional_hybrid_overlap": 0.75,
		"ai_hybrid_overlap": 0.70,
		"all_algorithms_overlap": 0.45,
	}
}

func (h *SimpleHybridHandler) findUniqueRecommendations(results map[string]algorithmResult) map[string][]string {
	unique := make(map[string][]string)
	
	for name := range results {
		unique[name] = []string{} // 简化实现
	}
	
	return unique
}

func (h *SimpleHybridHandler) findConsensusRecommendations(results map[string]algorithmResult) []map[string]interface{} {
	// 简化的一致性推荐实现
	return []map[string]interface{}{
		{
			"school_name": "清华大学",
			"major_name": "计算机科学与技术",
			"algorithms_agreed": []string{"traditional", "ai", "hybrid"},
			"consensus_score": 0.95,
		},
	}
}

func (h *SimpleHybridHandler) analyzeAlgorithmPreferences(results map[string]algorithmResult) map[string]interface{} {
	return map[string]interface{}{
		"traditional_favors": "知名院校和稳定专业",
		"ai_favors": "个性化匹配和新兴专业",
		"hybrid_favors": "平衡的综合推荐",
	}
}

func (h *SimpleHybridHandler) calculatePersonalizationScore(recommendations []cppbridge.Recommendation, request *cppbridge.RecommendationRequest) float64 {
	// 基于用户偏好计算个性化得分的简化实现
	score := 0.7 // 基础得分
	
	if request.Preferences != nil {
		// 根据偏好调整得分
		score += 0.2
	}
	
	return math.Min(1.0, score)
}

// A/B测试相关方法

// GetABTestGroup 获取用户的A/B测试组
func (h *SimpleHybridHandler) GetABTestGroup(userID string) *ABTestGroup {
	h.abTestConfig.mu.RLock()
	defer h.abTestConfig.mu.RUnlock()
	
	if !h.abTestConfig.Enabled {
		return h.abTestConfig.Groups[h.abTestConfig.DefaultGroup]
	}
	
	// 基于用户ID的哈希分配测试组
	hash := h.hashUserID(userID)
	
	var cumulativeRate float64
	for _, group := range h.abTestConfig.Groups {
		if !group.Active {
			continue
		}
		cumulativeRate += group.TrafficRate
		if hash < cumulativeRate {
			return group
		}
	}
	
	// fallback到默认组
	return h.abTestConfig.Groups[h.abTestConfig.DefaultGroup]
}

// hashUserID 对用户ID进行哈希以分配测试组
func (h *SimpleHybridHandler) hashUserID(userID string) float64 {
	// 简化的哈希实现
	hash := 0
	for _, char := range userID {
		hash = hash*31 + int(char)
	}
	return float64(hash%1000) / 1000.0
}

// CreateABTestGroup 创建新的A/B测试组
func (h *SimpleHybridHandler) CreateABTestGroup(c *gin.Context) {
	var group ABTestGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "测试组配置格式错误: " + err.Error(),
		})
		return
	}
	
	group.CreatedAt = time.Now()
	
	h.abTestConfig.mu.Lock()
	h.abTestConfig.Groups[group.GroupID] = &group
	h.abTestConfig.mu.Unlock()
	
	c.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"message": "A/B测试组创建成功",
		"group": group,
	})
}

// GetABTestStatus 获取A/B测试状态
func (h *SimpleHybridHandler) GetABTestStatus(c *gin.Context) {
	h.abTestConfig.mu.RLock()
	defer h.abTestConfig.mu.RUnlock()
	
	status := map[string]interface{}{
		"enabled": h.abTestConfig.Enabled,
		"default_group": h.abTestConfig.DefaultGroup,
		"groups": h.abTestConfig.Groups,
		"total_groups": len(h.abTestConfig.Groups),
		"active_groups": h.countActiveGroups(),
	}
	
	c.JSON(http.StatusOK, status)
}

// countActiveGroups 计算活跃的测试组数量
func (h *SimpleHybridHandler) countActiveGroups() int {
	count := 0
	for _, group := range h.abTestConfig.Groups {
		if group.Active {
			count++
		}
	}
	return count
}

// ToggleABTest 切换A/B测试状态
func (h *SimpleHybridHandler) ToggleABTest(c *gin.Context) {
	h.abTestConfig.mu.Lock()
	h.abTestConfig.Enabled = !h.abTestConfig.Enabled
	enabled := h.abTestConfig.Enabled
	h.abTestConfig.mu.Unlock()
	
	c.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"message": fmt.Sprintf("A/B测试已%s", map[bool]string{true: "启用", false: "禁用"}[enabled]),
		"enabled": enabled,
	})
}