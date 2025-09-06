package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/oktetopython/gaokao/recommendation-service/internal/services"
)

// WeightHandler 权重配置处理器
type WeightHandler struct {
	weightService *services.WeightService
	logger        *logrus.Logger
}

// NewWeightHandler 创建权重处理器
func NewWeightHandler(weightService *services.WeightService, logger *logrus.Logger) *WeightHandler {
	return &WeightHandler{
		weightService: weightService,
		logger:        logger,
	}
}

// RegisterRoutes 注册路由
func (h *WeightHandler) RegisterRoutes(router *gin.RouterGroup) {
	weightGroup := router.Group("/weights")
	{
		weightGroup.GET("", h.GetWeights)
		weightGroup.GET("/:key", h.GetWeightConfig)
		weightGroup.POST("/:key", h.SetWeightConfig)
		weightGroup.DELETE("/:key", h.DeleteWeightConfig)
		weightGroup.POST("/:key/reset", h.ResetWeightConfig)
		weightGroup.GET("/:key/export", h.ExportWeightConfig)
		weightGroup.POST("/:key/import", h.ImportWeightConfig)
		weightGroup.GET("/stats", h.GetWeightStats)
		weightGroup.POST("/presets", h.CreatePresetWeights)
	}
}

// GetWeights 获取所有权重配置
// @Summary 获取所有权重配置
// @Description 获取系统默认和所有自定义权重配置
// @Tags weights
// @Produce json
// @Success 200 {object} map[string]services.WeightConfig
// @Router /weights [get]
func (h *WeightHandler) GetWeights(c *gin.Context) {
	configs := h.weightService.ListWeights()
	c.JSON(http.StatusOK, configs)
}

// GetWeightConfig 获取特定权重配置
// @Summary 获取特定权重配置
// @Description 根据key获取权重配置，如果不存在则返回默认配置
// @Tags weights
// @Produce json
// @Param key path string true "配置键名"
// @Success 200 {object} services.WeightConfig
// @Router /weights/{key} [get]
func (h *WeightHandler) GetWeightConfig(c *gin.Context) {
	key := c.Param("key")
	config := h.weightService.GetWeights(key)
	c.JSON(http.StatusOK, config)
}

// SetWeightConfig 设置权重配置
// @Summary 设置权重配置
// @Description 创建或更新权重配置
// @Tags weights
// @Accept json
// @Produce json
// @Param key path string true "配置键名"
// @Param config body services.WeightConfig true "权重配置"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /weights/{key} [post]
func (h *WeightHandler) SetWeightConfig(c *gin.Context) {
	key := c.Param("key")
	
	var config services.WeightConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		h.logger.Warnf("解析权重配置失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "无效的权重配置格式",
		})
		return
	}
	
	if err := h.weightService.SetWeights(key, &config); err != nil {
		h.logger.Warnf("设置权重配置失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_weights",
			"message": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "权重配置已更新",
		"key":     key,
		"config":  config,
	})
}

// DeleteWeightConfig 删除权重配置
// @Summary 删除权重配置
// @Description 删除特定的权重配置
// @Tags weights
// @Produce json
// @Param key path string true "配置键名"
// @Success 200 {object} map[string]interface{}
// @Router /weights/{key} [delete]
func (h *WeightHandler) DeleteWeightConfig(c *gin.Context) {
	key := c.Param("key")
	
	h.weightService.DeleteWeights(key)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "权重配置已删除",
		"key":     key,
	})
}

// ResetWeightConfig 重置权重配置
// @Summary 重置权重配置
// @Description 重置为默认权重配置
// @Tags weights
// @Produce json
// @Param key path string true "配置键名"
// @Success 200 {object} map[string]interface{}
// @Router /weights/{key}/reset [post]
func (h *WeightHandler) ResetWeightConfig(c *gin.Context) {
	key := c.Param("key")
	
	h.weightService.ResetToDefault(key)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "权重配置已重置为默认",
		"key":     key,
	})
}

// ExportWeightConfig 导出权重配置
// @Summary 导出权重配置
// @Description 导出权重配置为JSON格式
// @Tags weights
// @Produce application/json
// @Param key path string true "配置键名"
// @Success 200 {string} string "权重配置JSON"
// @Router /weights/{key}/export [get]
func (h *WeightHandler) ExportWeightConfig(c *gin.Context) {
	key := c.Param("key")
	
	data, err := h.weightService.ExportWeights(key)
	if err != nil {
		h.logger.Warnf("导出权重配置失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "export_failed",
			"message": "导出权重配置失败",
		})
		return
	}
	
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=\"weight_config_"+key+".json\"")
	c.String(http.StatusOK, string(data))
}

// ImportWeightConfig 导入权重配置
// @Summary 导入权重配置
// @Description 从JSON文件导入权重配置
// @Tags weights
// @Accept multipart/form-data
// @Produce json
// @Param key path string true "配置键名"
// @Param file formData file true "权重配置JSON文件"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /weights/{key}/import [post]
func (h *WeightHandler) ImportWeightConfig(c *gin.Context) {
	key := c.Param("key")
	
	file, err := c.FormFile("file")
	if err != nil {
		h.logger.Warnf("获取上传文件失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "no_file",
			"message": "请选择要上传的文件",
		})
		return
	}
	
	// 检查文件类型
	if !strings.HasSuffix(file.Filename, ".json") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_file_type",
			"message": "只支持JSON格式文件",
		})
		return
	}
	
	// 读取文件内容
	f, err := file.Open()
	if err != nil {
		h.logger.Warnf("打开文件失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "file_open_error",
			"message": "无法打开文件",
		})
		return
	}
	defer f.Close()
	
	data := make([]byte, file.Size)
	_, err = f.Read(data)
	if err != nil {
		h.logger.Warnf("读取文件失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "file_read_error",
			"message": "无法读取文件",
		})
		return
	}
	
	// 导入配置
	if err := h.weightService.ImportWeights(key, data); err != nil {
		h.logger.Warnf("导入权重配置失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "import_failed",
			"message": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "权重配置导入成功",
		"key":     key,
	})
}

// GetWeightStats 获取权重统计信息
// @Summary 获取权重统计信息
// @Description 获取权重配置的统计信息
// @Tags weights
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /weights/stats [get]
func (h *WeightHandler) GetWeightStats(c *gin.Context) {
	stats := h.weightService.GetWeightStats()
	c.JSON(http.StatusOK, stats)
}

// CreatePresetWeights 创建预设权重配置
// @Summary 创建预设权重配置
// @Description 创建系统预设的权重配置
// @Tags weights
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /weights/presets [post]
func (h *WeightHandler) CreatePresetWeights(c *gin.Context) {
	h.weightService.CreatePresetWeights()
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "预设权重配置已创建",
		"presets": []string{"score_focused", "location_focused", "employment_focused"},
	})
}

// WeightConfigRequest 权重配置请求（用于文档）
type WeightConfigRequest struct {
	Name        string  `json:"name" example:"自定义配置"`
	Description string  `json:"description" example:"我的自定义权重配置"`
	ScoreMatchWeight    float64 `json:"score_match_weight" example:"0.35"`
	LocationWeight      float64 `json:"location_weight" example:"0.15"`
	InterestWeight      float64 `json:"interest_weight" example:"0.15"`
	EmploymentWeight    float64 `json:"employment_weight" example:"0.15"`
	UniversityRankWeight float64 `json:"university_rank_weight" example:"0.10"`
	CompetitionWeight   float64 `json:"competition_weight" example:"0.10"`
	EnableAdaptiveWeights bool   `json:"enable_adaptive_weights" example:"true"`
	MinWeightThreshold    float64 `json:"min_weight_threshold" example:"0.05"`
	MaxWeightThreshold    float64 `json:"max_weight_threshold" example:"0.40"`
}