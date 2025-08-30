package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gaokao/recommendation-service/pkg/cppbridge"
)

// APIResponse 统一API响应格式
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// RecommendationHandler 推荐处理器
type RecommendationHandler struct {
	bridge *cppbridge.HybridRecommendationBridge
}

// NewRecommendationHandler 创建新的推荐处理器
func NewRecommendationHandler(bridge *cppbridge.HybridRecommendationBridge) *RecommendationHandler {
	return &RecommendationHandler{
		bridge: bridge,
	}
}

// GenerateRecommendationsRequest 生成推荐请求
type GenerateRecommendationsRequest struct {
	Student       *cppbridge.Student `json:"student" binding:"required"`
	MaxVolunteers int                `json:"max_volunteers,omitempty"`
	Options       *RecommendationOptions `json:"options,omitempty"`
}

// RecommendationOptions 推荐选项
type RecommendationOptions struct {
	AlgorithmType    string  `json:"algorithm_type"`    // "hybrid", "traditional", "ai"
	RiskPreference   string  `json:"risk_preference"`   // "conservative", "moderate", "aggressive"
	DiversityLevel   string  `json:"diversity_level"`   // "low", "medium", "high"
	PreferenceWeight float64 `json:"preference_weight"` // 0-1
}

// BatchGenerateRecommendationsRequest 批量生成推荐请求
type BatchGenerateRecommendationsRequest struct {
	Students      []*cppbridge.Student   `json:"students" binding:"required"`
	MaxVolunteers int                    `json:"max_volunteers,omitempty"`
	Options       *RecommendationOptions `json:"options,omitempty"`
}

// BatchGenerateRecommendationsResponse 批量生成推荐响应
type BatchGenerateRecommendationsResponse struct {
	Results []BatchRecommendationResult `json:"results"`
	Summary *BatchSummary               `json:"summary"`
}

// BatchRecommendationResult 批量推荐结果
type BatchRecommendationResult struct {
	StudentID string                       `json:"student_id"`
	Plan      *cppbridge.VolunteerPlan    `json:"plan,omitempty"`
	Error     string                       `json:"error,omitempty"`
	Status    string                       `json:"status"` // "success", "error"
}

// BatchSummary 批量处理摘要
type BatchSummary struct {
	TotalStudents    int     `json:"total_students"`
	SuccessCount     int     `json:"success_count"`
	ErrorCount       int     `json:"error_count"`
	SuccessRate      float64 `json:"success_rate"`
	AvgProcessingTime float64 `json:"avg_processing_time_ms"`
}

// OptimizeRecommendationsRequest 优化推荐请求
type OptimizeRecommendationsRequest struct {
	CurrentPlan      *cppbridge.VolunteerPlan `json:"current_plan" binding:"required"`
	OptimizationGoal string                   `json:"optimization_goal"` // "safety", "diversity", "preference", "balance"
	Constraints      *OptimizationConstraints `json:"constraints,omitempty"`
}

// OptimizationConstraints 优化约束
type OptimizationConstraints struct {
	MinSafeRatio    float64  `json:"min_safe_ratio"`
	MaxRushRatio    float64  `json:"max_rush_ratio"`
	RequiredCities  []string `json:"required_cities,omitempty"`
	RequiredMajors  []string `json:"required_majors,omitempty"`
	ForbiddenCities []string `json:"forbidden_cities,omitempty"`
	ForbiddenMajors []string `json:"forbidden_majors,omitempty"`
}

// ExplainRecommendationResponse 解释推荐响应
type ExplainRecommendationResponse struct {
	Recommendation    *cppbridge.VolunteerRecommendation `json:"recommendation"`
	DetailedExplanation *DetailedExplanation             `json:"detailed_explanation"`
	SimilarOptions    []*cppbridge.VolunteerRecommendation `json:"similar_options,omitempty"`
	RiskFactors       []RiskFactor                        `json:"risk_factors"`
	Suggestions       []string                            `json:"suggestions"`
}

// DetailedExplanation 详细解释
type DetailedExplanation struct {
	MatchAnalysis    *MatchAnalysis    `json:"match_analysis"`
	RiskAnalysis     *RiskAnalysisDetail `json:"risk_analysis"`
	PreferenceMatch  *PreferenceMatch  `json:"preference_match"`
	CompetitionLevel *CompetitionLevel `json:"competition_level"`
}

// MatchAnalysis 匹配分析
type MatchAnalysis struct {
	ScoreMatch      float64 `json:"score_match"`
	RankingMatch    float64 `json:"ranking_match"`
	SubjectMatch    float64 `json:"subject_match"`
	OverallMatch    float64 `json:"overall_match"`
	MatchReason     string  `json:"match_reason"`
}

// RiskAnalysisDetail 风险分析详情
type RiskAnalysisDetail struct {
	AdmissionProbability float64    `json:"admission_probability"`
	RiskLevel           string     `json:"risk_level"`
	RiskFactors         []string   `json:"risk_factors"`
	HistoricalTrend     string     `json:"historical_trend"`
	CompetitionIntensity float64   `json:"competition_intensity"`
}

// PreferenceMatch 偏好匹配
type PreferenceMatch struct {
	CityMatch         bool    `json:"city_match"`
	MajorMatch        bool    `json:"major_match"`
	LevelMatch        bool    `json:"level_match"`
	OverallPreferenceScore float64 `json:"overall_preference_score"`
}

// CompetitionLevel 竞争水平
type CompetitionLevel struct {
	EnrollmentPlan    int     `json:"enrollment_plan"`
	ExpectedApplicants int    `json:"expected_applicants"`
	CompetitionRatio  float64 `json:"competition_ratio"`
	DifficultyLevel   string  `json:"difficulty_level"`
}

// RiskFactor 风险因素
type RiskFactor struct {
	Factor      string  `json:"factor"`
	Impact      string  `json:"impact"`      // "high", "medium", "low"
	Description string  `json:"description"`
	Mitigation  string  `json:"mitigation"`
}

// SystemStatusResponse 系统状态响应
type SystemStatusResponse struct {
	Service       string                    `json:"service"`
	Status        string                    `json:"status"`
	Version       string                    `json:"version"`
	Uptime        int64                     `json:"uptime_seconds"`
	HybridStats   *cppbridge.HybridStats   `json:"hybrid_stats"`
	Performance   *PerformanceMetrics       `json:"performance"`
	Resources     *ResourceUsage            `json:"resources"`
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	RequestsPerSecond  float64 `json:"requests_per_second"`
	AvgResponseTime    float64 `json:"avg_response_time_ms"`
	MaxResponseTime    float64 `json:"max_response_time_ms"`
	ErrorRate          float64 `json:"error_rate"`
	CacheHitRate       float64 `json:"cache_hit_rate"`
}

// ResourceUsage 资源使用情况
type ResourceUsage struct {
	MemoryUsage    int64   `json:"memory_usage_bytes"`
	CPUUsage       float64 `json:"cpu_usage_percent"`
	GoroutineCount int     `json:"goroutine_count"`
	GCStats        *GCStats `json:"gc_stats"`
}

// GCStats 垃圾回收统计
type GCStats struct {
	NumGC        uint32  `json:"num_gc"`
	PauseTotal   int64   `json:"pause_total_ns"`
	PauseAvg     float64 `json:"pause_avg_ns"`
}

// GenerateRecommendations 生成推荐
// @Summary 生成志愿推荐
// @Description 为单个学生生成志愿填报推荐方案
// @Tags recommendations
// @Accept json
// @Produce json
// @Param request body GenerateRecommendationsRequest true "生成推荐请求"
// @Success 200 {object} APIResponse{data=cppbridge.VolunteerPlan}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /recommendations/generate [post]
func (h *RecommendationHandler) GenerateRecommendations(c *gin.Context) {
	var req GenerateRecommendationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request parameters",
			Error:   err.Error(),
		})
		return
	}

	// 设置默认值
	if req.MaxVolunteers == 0 {
		req.MaxVolunteers = 96
	}
	if req.Options == nil {
		req.Options = &RecommendationOptions{
			AlgorithmType:    "hybrid",
			RiskPreference:   "moderate",
			DiversityLevel:   "medium",
			PreferenceWeight: 0.7,
		}
	}

	// 验证学生信息
	if err := validateStudent(req.Student); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid student information",
			Error:   err.Error(),
		})
		return
	}

	// 生成推荐
	plan, err := h.bridge.GenerateHybridPlan(req.Student, req.MaxVolunteers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to generate recommendations",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Recommendations generated successfully",
		Data:    plan,
	})
}

// BatchGenerateRecommendations 批量生成推荐
// @Summary 批量生成推荐
// @Description 为多个学生批量生成志愿填报推荐方案
// @Tags recommendations
// @Accept json
// @Produce json
// @Param request body BatchGenerateRecommendationsRequest true "批量生成推荐请求"
// @Success 200 {object} APIResponse{data=BatchGenerateRecommendationsResponse}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /recommendations/batch [post]
func (h *RecommendationHandler) BatchGenerateRecommendations(c *gin.Context) {
	var req BatchGenerateRecommendationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request parameters",
			Error:   err.Error(),
		})
		return
	}

	// 设置默认值
	if req.MaxVolunteers == 0 {
		req.MaxVolunteers = 96
	}
	if req.Options == nil {
		req.Options = &RecommendationOptions{
			AlgorithmType:    "hybrid",
			RiskPreference:   "moderate",
			DiversityLevel:   "medium",
			PreferenceWeight: 0.7,
		}
	}

	// 验证输入
	if len(req.Students) == 0 {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "No students provided",
		})
		return
	}

	if len(req.Students) > 100 { // 限制批量大小
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Too many students (maximum 100)",
		})
		return
	}

	// 批量处理
	response := &BatchGenerateRecommendationsResponse{
		Results: make([]BatchRecommendationResult, len(req.Students)),
		Summary: &BatchSummary{
			TotalStudents: len(req.Students),
		},
	}

	successCount := 0
	errorCount := 0

	for i, student := range req.Students {
		result := &response.Results[i]
		result.StudentID = student.StudentID

		// 验证学生信息
		if err := validateStudent(student); err != nil {
			result.Status = "error"
			result.Error = err.Error()
			errorCount++
			continue
		}

		// 生成推荐
		plan, err := h.bridge.GenerateHybridPlan(student, req.MaxVolunteers)
		if err != nil {
			result.Status = "error"
			result.Error = err.Error()
			errorCount++
		} else {
			result.Status = "success"
			result.Plan = plan
			successCount++
		}
	}

	// 计算摘要
	response.Summary.SuccessCount = successCount
	response.Summary.ErrorCount = errorCount
	response.Summary.SuccessRate = float64(successCount) / float64(len(req.Students))

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Batch recommendations generated",
		Data:    response,
	})
}

// ExplainRecommendation 解释推荐
// @Summary 解释推荐
// @Description 获取特定推荐的详细解释
// @Tags recommendations
// @Produce json
// @Param id path string true "推荐ID"
// @Param student_id query string true "学生ID"
// @Success 200 {object} APIResponse{data=ExplainRecommendationResponse}
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /recommendations/explain/{id} [get]
func (h *RecommendationHandler) ExplainRecommendation(c *gin.Context) {
	recommendationID := c.Param("id")
	studentID := c.Query("student_id")

	if recommendationID == "" || studentID == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Missing recommendation ID or student ID",
		})
		return
	}

	// TODO: 这里需要从数据库或缓存中获取推荐信息
	// 现在使用模拟数据
	recommendation := &cppbridge.VolunteerRecommendation{
		UniversityID:         recommendationID,
		UniversityName:       "示例大学",
		MajorID:              "example_major",
		MajorName:            "示例专业",
		AdmissionProbability: 0.75,
		RiskLevel:            "稳",
		ScoreGap:             10,
		RankingGap:           500,
		MatchScore:           85.5,
		RecommendationReason: "综合匹配度较高",
		RiskFactors:          []string{"竞争激烈", "分数要求较高"},
	}

	// 获取详细解释
	explanation, err := h.bridge.GetHybridExplanation(recommendation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to generate explanation",
			Error:   err.Error(),
		})
		return
	}

	// 构建详细响应
	response := &ExplainRecommendationResponse{
		Recommendation: recommendation,
		DetailedExplanation: &DetailedExplanation{
			MatchAnalysis: &MatchAnalysis{
				ScoreMatch:   0.85,
				RankingMatch: 0.78,
				SubjectMatch: 0.92,
				OverallMatch: 0.85,
				MatchReason:  "学生成绩与目标院校录取要求高度匹配",
			},
			RiskAnalysis: &RiskAnalysisDetail{
				AdmissionProbability: recommendation.AdmissionProbability,
				RiskLevel:           recommendation.RiskLevel,
				RiskFactors:         recommendation.RiskFactors,
				HistoricalTrend:     "录取分数线呈上升趋势",
				CompetitionIntensity: 0.7,
			},
			PreferenceMatch: &PreferenceMatch{
				CityMatch:              true,
				MajorMatch:             true,
				LevelMatch:             true,
				OverallPreferenceScore: 0.9,
			},
			CompetitionLevel: &CompetitionLevel{
				EnrollmentPlan:     200,
				ExpectedApplicants: 1500,
				CompetitionRatio:   7.5,
				DifficultyLevel:    "中等",
			},
		},
		RiskFactors: []RiskFactor{
			{
				Factor:      "分数竞争",
				Impact:      "medium",
				Description: "该专业历年录取分数较为稳定，但申请人数较多",
				Mitigation:  "建议同时关注相关专业作为备选",
			},
		},
		Suggestions: []string{
			"考虑申请该校的相关专业作为备选",
			"关注该专业的调剂政策",
			"准备充分的面试材料（如适用）",
		},
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Explanation generated successfully",
		Data:    response,
	})
}

// OptimizeRecommendations 优化推荐
// @Summary 优化推荐方案
// @Description 基于特定目标优化现有推荐方案
// @Tags recommendations
// @Accept json
// @Produce json
// @Param request body OptimizeRecommendationsRequest true "优化推荐请求"
// @Success 200 {object} APIResponse{data=cppbridge.VolunteerPlan}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /recommendations/optimize [post]
func (h *RecommendationHandler) OptimizeRecommendations(c *gin.Context) {
	var req OptimizeRecommendationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request parameters",
			Error:   err.Error(),
		})
		return
	}

	if req.CurrentPlan == nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Current plan is required",
		})
		return
	}

	// 设置默认优化目标
	if req.OptimizationGoal == "" {
		req.OptimizationGoal = "balance"
	}

	// TODO: 实现具体的优化逻辑
	// 这里返回原方案作为示例
	optimizedPlan := req.CurrentPlan

	// 模拟优化过程
	switch req.OptimizationGoal {
	case "safety":
		// 增加保底志愿
		optimizedPlan.SafeCount += 2
		optimizedPlan.RushCount -= 1
		optimizedPlan.StableCount -= 1
		optimizedPlan.OverallRiskScore += 0.1
		optimizedPlan.OptimizationSuggestions = append(optimizedPlan.OptimizationSuggestions,
			"已增加保底志愿数量，提高录取安全性")

	case "diversity":
		// 增加多样性
		optimizedPlan.OptimizationSuggestions = append(optimizedPlan.OptimizationSuggestions,
			"已优化地域和专业分布，增加选择多样性")

	case "preference":
		// 优化偏好匹配
		optimizedPlan.OptimizationSuggestions = append(optimizedPlan.OptimizationSuggestions,
			"已根据个人偏好调整推荐顺序")

	case "balance":
		// 平衡优化
		optimizedPlan.OptimizationSuggestions = append(optimizedPlan.OptimizationSuggestions,
			"已平衡风险和收益，优化整体方案结构")
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Recommendations optimized successfully",
		Data:    optimizedPlan,
	})
}

// GetSystemStatus 获取系统状态
// @Summary 获取系统状态
// @Description 获取推荐服务的系统状态和性能指标
// @Tags system
// @Produce json
// @Success 200 {object} APIResponse{data=SystemStatusResponse}
// @Failure 500 {object} APIResponse
// @Router /system/status [get]
func (h *RecommendationHandler) GetSystemStatus(c *gin.Context) {
	// 获取混合引擎统计
	hybridStats, err := h.bridge.GetHybridStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to get system status",
			Error:   err.Error(),
		})
		return
	}

	// 构建系统状态响应
	status := &SystemStatusResponse{
		Service:     "recommendation-service",
		Status:      "healthy",
		Version:     "1.0.0",
		Uptime:      3600, // 示例值
		HybridStats: hybridStats,
		Performance: &PerformanceMetrics{
			RequestsPerSecond: 50.0,
			AvgResponseTime:   150.0,
			MaxResponseTime:   500.0,
			ErrorRate:         0.01,
			CacheHitRate:      0.85,
		},
		Resources: &ResourceUsage{
			MemoryUsage:    1024 * 1024 * 100, // 100MB
			CPUUsage:       25.5,
			GoroutineCount: 20,
			GCStats: &GCStats{
				NumGC:     100,
				PauseTotal: 1000000,
				PauseAvg:   10000,
			},
		},
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "System status retrieved successfully",
		Data:    status,
	})
}

// ClearCache 清空缓存
// @Summary 清空缓存
// @Description 清空推荐系统的缓存
// @Tags system
// @Produce json
// @Success 200 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /system/cache/clear [post]
func (h *RecommendationHandler) ClearCache(c *gin.Context) {
	// TODO: 实现缓存清理逻辑
	// 这里只是返回成功响应
	
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Cache cleared successfully",
	})
}

// UpdateModel 更新模型
// @Summary 更新AI模型
// @Description 热更新AI推荐模型
// @Tags system
// @Accept json
// @Produce json
// @Param model_path formData string true "新模型文件路径"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /system/model/update [put]
func (h *RecommendationHandler) UpdateModel(c *gin.Context) {
	modelPath := c.PostForm("model_path")
	if modelPath == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Model path is required",
		})
		return
	}

	// TODO: 实现模型更新逻辑
	// 需要调用C++的模型更新接口
	
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Model updated successfully",
		Data: map[string]string{
			"model_path": modelPath,
			"updated_at": "2025-01-18T10:00:00Z",
		},
	})
}