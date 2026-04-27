package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oktetopython/gaokao/services/recommendation-service/internal/types"
)

// Type aliases from types package
type AnalyticsService = types.AnalyticsService
type RecommendationStats = types.RecommendationStats
type SystemMetrics = types.SystemMetrics
type UserBehaviorAnalysis = types.UserBehaviorAnalysis
type AlgorithmPerformance = types.AlgorithmPerformance
type AnalyticsPerformanceMetrics = types.AnalyticsPerformanceMetrics
type ComparisonMetrics = types.ComparisonMetrics
type TimeRange = types.TimeRange
type RequestPattern = types.RequestPattern
type PreferenceProfile = types.PreferenceProfile
type EngagementMetrics = types.EngagementMetrics
type FeedbackItem = types.FeedbackItem
type ScoreRange = types.ScoreRange
type RecommendationItem = types.RecommendationItem

// AnalyticsHandler 分析处理器
type AnalyticsHandler struct {
	analyticsService AnalyticsService
}

// NewAnalyticsHandler 创建分析处理器
func NewAnalyticsHandler(analyticsService AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
	}
}

// RegisterRoutes 注册 analytics 路由
func (h *AnalyticsHandler) RegisterRoutes(router *gin.RouterGroup) {
	analytics := router.Group("/analytics")
	{
		analytics.GET("/recommendations/:user_id", h.GetRecommendationStats)
		analytics.GET("/system/metrics", h.GetSystemMetrics)
		analytics.GET("/users/:user_id/behavior", h.GetUserBehaviorAnalysis)
		analytics.GET("/algorithms/performance", h.GetAlgorithmPerformance)
		analytics.GET("/realtime/metrics", h.GetRealtimeMetrics)
		analytics.GET("/export", h.ExportAnalyticsReport)
		analytics.GET("/performance", h.GetPerformanceMetrics)
		analytics.GET("/fusion-stats", h.GetFusionStatistics)
		analytics.POST("/quality-report", h.GenerateQualityReport)
		analytics.GET("/trends", h.GetRecommendationTrends)
	}
}

// GetRecommendationStats 获取推荐统计
// @Summary 获取用户推荐统计数据
// @Description 获取指定用户在指定时间范围内的推荐统计数据
// @Tags analytics
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Param start_time query string false "开始时间 (RFC3339格式)"
// @Param end_time query string false "结束时间 (RFC3339格式)"
// @Success 200 {object} RecommendationStats
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /analytics/recommendations/{user_id} [get]
func (h *AnalyticsHandler) GetRecommendationStats(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_user_id",
			Message: "用户ID不能为空",
		})
		return
	}

	// 解析时间参数
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	var startTime, endTime time.Time
	var err error

	if startTimeStr != "" {
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_start_time",
				Message: "开始时间格式错误",
			})
			return
		}
	} else {
		startTime = time.Now().AddDate(0, 0, -7) // 默认7天前
	}

	if endTimeStr != "" {
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_end_time",
				Message: "结束时间格式错误",
			})
			return
		}
	} else {
		endTime = time.Now() // 默认现在
	}

	stats, err := h.analyticsService.GetRecommendationStats(userID, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "stats_error",
			Message: "获取统计数据失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetSystemMetrics 获取系统指标
// @Summary 获取系统性能指标
// @Description 获取当前系统的性能指标和健康状态
// @Tags analytics
// @Accept json
// @Produce json
// @Success 200 {object} SystemMetrics
// @Failure 500 {object} ErrorResponse
// @Router /analytics/system/metrics [get]
func (h *AnalyticsHandler) GetSystemMetrics(c *gin.Context) {
	metrics, err := h.analyticsService.GetSystemMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "metrics_error",
			Message: "获取系统指标失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetUserBehaviorAnalysis 获取用户行为分析
// @Summary 获取用户行为分析报告
// @Description 获取指定用户的行为模式和偏好分析
// @Tags analytics
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Success 200 {object} UserBehaviorAnalysis
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /analytics/users/{user_id}/behavior [get]
func (h *AnalyticsHandler) GetUserBehaviorAnalysis(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_user_id",
			Message: "用户ID不能为空",
		})
		return
	}

	analysis, err := h.analyticsService.GetUserBehaviorAnalysis(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "analysis_error",
			Message: "获取用户行为分析失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, analysis)
}

// GetAlgorithmPerformance 获取算法性能比较
// @Summary 获取算法性能分析
// @Description 获取各算法的性能对比和优化建议
// @Tags analytics
// @Accept json
// @Produce json
// @Success 200 {object} AlgorithmPerformance
// @Failure 500 {object} ErrorResponse
// @Router /analytics/algorithms/performance [get]
func (h *AnalyticsHandler) GetAlgorithmPerformance(c *gin.Context) {
	performance, err := h.analyticsService.GetAlgorithmPerformance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "performance_error",
			Message: "获取算法性能分析失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, performance)
}

// GetRealtimeMetrics 获取实时指标
// @Summary 获取实时系统指标
// @Description 获取系统实时性能指标，支持WebSocket推送
// @Tags analytics
// @Accept json
// @Produce json
// @Param interval query int false "更新间隔(秒)" default(5)
// @Success 200 {object} SystemMetrics
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /analytics/realtime/metrics [get]
func (h *AnalyticsHandler) GetRealtimeMetrics(c *gin.Context) {
	intervalStr := c.DefaultQuery("interval", "5")
	interval, err := strconv.Atoi(intervalStr)
	if err != nil || interval < 1 || interval > 60 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_interval",
			Message: "更新间隔必须在1-60秒之间",
		})
		return
	}

	// 实时指标推送逻辑
	// 这里可以实现WebSocket连接或Server-Sent Events
	metrics, err := h.analyticsService.GetSystemMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "metrics_error",
			Message: "获取实时指标失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// ExportAnalyticsReport 导出分析报告
// @Summary 导出分析报告
// @Description 导出指定时间范围的完整分析报告
// @Tags analytics
// @Accept json
// @Produce application/json
// @Param start_time query string true "开始时间 (RFC3339格式)"
// @Param end_time query string true "结束时间 (RFC3339格式)"
// @Param format query string false "导出格式 (json|csv|pdf)" default(json)
// @Param user_id query string false "用户ID（可选，导出特定用户报告）"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /analytics/export [get]
func (h *AnalyticsHandler) ExportAnalyticsReport(c *gin.Context) {
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")
	format := c.DefaultQuery("format", "json")
	userID := c.Query("user_id")

	if startTimeStr == "" || endTimeStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_time_range",
			Message: "必须指定开始时间和结束时间",
		})
		return
	}

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_start_time",
			Message: "开始时间格式错误",
		})
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_end_time",
			Message: "结束时间格式错误",
		})
		return
	}

	// 构建完整报告
	report := make(map[string]interface{})

	// 系统指标
	systemMetrics, err := h.analyticsService.GetSystemMetrics()
	if err == nil {
		report["system_metrics"] = systemMetrics
	}

	// 算法性能
	algorithmPerformance, err := h.analyticsService.GetAlgorithmPerformance()
	if err == nil {
		report["algorithm_performance"] = algorithmPerformance
	}

	// 如果指定了用户ID，添加用户特定数据
	if userID != "" {
		userStats, err := h.analyticsService.GetRecommendationStats(userID, startTime, endTime)
		if err == nil {
			report["user_stats"] = userStats
		}

		userBehavior, err := h.analyticsService.GetUserBehaviorAnalysis(userID)
		if err == nil {
			report["user_behavior"] = userBehavior
		}
	}

	// 添加报告元数据
	report["metadata"] = map[string]interface{}{
		"generated_at": time.Now(),
		"time_range": TimeRange{
			StartTime: startTime,
			EndTime:   endTime,
		},
		"format":  format,
		"user_id": userID,
	}

	// 根据格式返回数据
	switch format {
	case "json":
		c.Header("Content-Type", "application/json")
		c.Header("Content-Disposition", "attachment; filename=analytics_report.json")
		c.JSON(http.StatusOK, report)
	case "csv":
		// CSV导出逻辑
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=analytics_report.csv")
		// 这里需要实现CSV转换逻辑
		c.JSON(http.StatusNotImplemented, ErrorResponse{
			Error:   "format_not_implemented",
			Message: "CSV格式暂未实现",
		})
	case "pdf":
		// PDF导出逻辑
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", "attachment; filename=analytics_report.pdf")
		// 这里需要实现PDF转换逻辑
		c.JSON(http.StatusNotImplemented, ErrorResponse{
			Error:   "format_not_implemented",
			Message: "PDF格式暂未实现",
		})
	default:
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "unsupported_format",
			Message: "不支持的导出格式",
		})
	}
}

// GetPerformanceMetrics 获取性能指标
// @Summary 获取系统性能指标
// @Description 获取系统性能相关指标
// @Tags analytics
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} ErrorResponse
// @Router /analytics/performance [get]
func (h *AnalyticsHandler) GetPerformanceMetrics(c *gin.Context) {
	metrics, err := h.analyticsService.GetPerformanceMetrics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "metrics_error",
			Message: "获取性能指标失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetFusionStatistics 获取融合统计数据
// @Summary 获取算法融合统计
// @Description 获取传统算法、AI算法和混合算法的统计对比
// @Tags analytics
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} ErrorResponse
// @Router /analytics/fusion-stats [get]
func (h *AnalyticsHandler) GetFusionStatistics(c *gin.Context) {
	stats, err := h.analyticsService.GetFusionStatistics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "stats_error",
			Message: "获取融合统计失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GenerateQualityReport 生成质量报告
// @Summary 生成推荐质量报告
// @Description 生成指定时间范围内的推荐质量分析报告
// @Tags analytics
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "报告参数"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /analytics/quality-report [post]
func (h *AnalyticsHandler) GenerateQualityReport(c *gin.Context) {
	var params map[string]interface{}
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_params",
			Message: "请求参数格式错误",
		})
		return
	}

	report, err := h.analyticsService.GenerateQualityReport(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "report_error",
			Message: "生成质量报告失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetRecommendationTrends 获取推荐趋势
// @Summary 获取推荐趋势数据
// @Description 获取指定时间范围内的推荐趋势分析
// @Tags analytics
// @Accept json
// @Produce json
// @Param time_range query string false "时间范围 (1h|24h|7d|30d)" default(24h)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /analytics/trends [get]
func (h *AnalyticsHandler) GetRecommendationTrends(c *gin.Context) {
	timeRange := c.DefaultQuery("time_range", "24h")

	// 验证时间范围参数
	validRanges := map[string]bool{
		"1h": true, "24h": true, "7d": true, "30d": true,
	}

	if !validRanges[timeRange] {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_time_range",
			Message: "时间范围必须是: 1h, 24h, 7d, 30d 中的一个",
		})
		return
	}

	trends, err := h.analyticsService.GetRecommendationTrends(c.Request.Context(), timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "trends_error",
			Message: "获取推荐趋势失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, trends)
}
