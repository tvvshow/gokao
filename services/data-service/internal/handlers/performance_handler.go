package handlers

import (
	"data-service/internal/database"
	"data-service/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// PerformanceHandler 性能监控处理器
type PerformanceHandler struct {
	performanceService *services.PerformanceService
	cacheService       *services.CacheService
	database          *database.DB
	logger             *logrus.Logger
}

// NewPerformanceHandler 创建性能监控处理器实例
func NewPerformanceHandler(performanceService *services.PerformanceService, cacheService *services.CacheService, database *database.DB, logger *logrus.Logger) *PerformanceHandler {
	return &PerformanceHandler{
		performanceService: performanceService,
		cacheService:       cacheService,
		database:          database,
		logger:             logger,
	}
}

// GetMetrics 获取性能指标
// @Summary 获取性能指标
// @Description 获取详细的性能指标数据
// @Tags Performance
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=services.PerformanceMetrics} "成功"
// @Failure 500 {object} APIResponse "服务器错误"
// @Router /api/v1/performance/metrics [get]
func (h *PerformanceHandler) GetMetrics(c *gin.Context) {
	metrics := h.performanceService.GetMetrics()
	c.JSON(http.StatusOK, NewSuccessResponse(metrics))
}

// GetSummary 获取性能摘要
// @Summary 获取性能摘要
// @Description 获取性能指标的摘要信息
// @Tags Performance
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=map[string]interface{}} "成功"
// @Failure 500 {object} APIResponse "服务器错误"
// @Router /api/v1/performance/summary [get]
func (h *PerformanceHandler) GetSummary(c *gin.Context) {
	summary := h.performanceService.GetSummary()
	c.JSON(http.StatusOK, NewSuccessResponse(summary))
}

// ResetMetrics 重置性能指标
// @Summary 重置性能指标
// @Description 重置所有性能指标数据
// @Tags Performance
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse "成功"
// @Failure 500 {object} APIResponse "服务器错误"
// @Router /api/v1/performance/reset [post]
func (h *PerformanceHandler) ResetMetrics(c *gin.Context) {
	h.performanceService.ResetMetrics()
	c.JSON(http.StatusOK, NewSuccessResponse("性能指标已重置"))
}

// GetCacheStats 获取缓存统计
// @Summary 获取缓存统计
// @Description 获取Redis缓存的统计信息
// @Tags Performance
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=map[string]interface{}} "成功"
// @Failure 500 {object} APIResponse "服务器错误"
// @Router /api/v1/performance/cache-stats [get]
func (h *PerformanceHandler) GetCacheStats(c *gin.Context) {
	stats, err := h.cacheService.GetCacheStats(c.Request.Context())
	if err != nil {
		h.logger.Errorf("获取缓存统计失败: %v", err)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("获取缓存统计失败"))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse(stats))
}

// ClearCache 清空缓存
// @Summary 清空缓存
// @Description 清空Redis中的所有缓存数据
// @Tags Performance
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse "成功"
// @Failure 500 {object} APIResponse "服务器错误"
// @Router /api/v1/performance/clear-cache [post]
func (h *PerformanceHandler) ClearCache(c *gin.Context) {
	err := h.cacheService.ClearAllCache(c.Request.Context())
	if err != nil {
		h.logger.Errorf("清空缓存失败: %v", err)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("清空缓存失败"))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("缓存已清空"))
}

// RefreshCache 刷新指定类型的缓存
// @Summary 刷新缓存
// @Description 刷新指定类型的缓存数据
// @Tags Performance
// @Accept json
// @Produce json
// @Param type query string true "缓存类型" Enums(university, major, admission, search, statistics)
// @Success 200 {object} APIResponse "成功"
// @Failure 400 {object} APIResponse "请求参数错误"
// @Failure 500 {object} APIResponse "服务器错误"
// @Router /api/v1/performance/refresh-cache [post]
func (h *PerformanceHandler) RefreshCache(c *gin.Context) {
	cacheType := c.Query("type")
	if cacheType == "" {
		c.JSON(http.StatusBadRequest, NewErrorResponse("缓存类型不能为空"))
		return
	}

	// 验证缓存类型
	validTypes := map[string]bool{
		"university": true,
		"major":      true,
		"admission":  true,
		"search":     true,
		"statistics": true,
		"hot_searches": true,
		"autocomplete": true,
	}
	
	if !validTypes[cacheType] {
		c.JSON(http.StatusBadRequest, NewErrorResponse("无效的缓存类型"))
		return
	}

	err := h.cacheService.RefreshCache(c.Request.Context(), cacheType)
	if err != nil {
		h.logger.Errorf("刷新缓存失败: %v", err)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("刷新缓存失败"))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("缓存已刷新"))
}

// WarmupCache 缓存预热
// @Summary 缓存预热
// @Description 预热常用的缓存数据
// @Tags Performance
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse "成功"
// @Failure 500 {object} APIResponse "服务器错误"
// @Router /api/v1/performance/warmup-cache [post]
func (h *PerformanceHandler) WarmupCache(c *gin.Context) {
	err := h.cacheService.WarmupCache(c.Request.Context())
	if err != nil {
		h.logger.Errorf("缓存预热失败: %v", err)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("缓存预热失败"))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse("缓存预热已启动"))
}

// GetDBPoolStats 获取数据库连接池统计
// @Summary 获取数据库连接池统计
// @Description 获取PostgreSQL数据库连接池的详细统计信息
// @Tags Performance
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=map[string]interface{}} "成功"
// @Failure 500 {object} APIResponse "服务器错误"
// @Router /api/v1/performance/db-pool-stats [get]
func (h *PerformanceHandler) GetDBPoolStats(c *gin.Context) {
	if h.database == nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse("数据库未初始化"))
		return
	}

	stats := h.database.GetConnectionPoolStats()
	c.JSON(http.StatusOK, NewSuccessResponse(stats))
}