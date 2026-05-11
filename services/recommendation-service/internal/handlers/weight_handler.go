package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/tvvshow/gokao/pkg/response"
	"github.com/tvvshow/gokao/services/recommendation-service/internal/services"
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
func (h *WeightHandler) GetWeights(c *gin.Context) {
	configs := h.weightService.ListWeights()
	response.OK(c, configs)
}

// GetWeightConfig 获取特定权重配置
func (h *WeightHandler) GetWeightConfig(c *gin.Context) {
	key := c.Param("key")
	config := h.weightService.GetWeights(key)
	response.OK(c, config)
}

// SetWeightConfig 设置权重配置
func (h *WeightHandler) SetWeightConfig(c *gin.Context) {
	key := c.Param("key")

	var config services.WeightConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		h.logger.Warnf("解析权重配置失败: %v", err)
		response.BadRequest(c, "invalid_request", "无效的权重配置格式", nil)
		return
	}

	if err := h.weightService.SetWeights(key, &config); err != nil {
		h.logger.Warnf("设置权重配置失败: %v", err)
		response.BadRequest(c, "invalid_weights", err.Error(), nil)
		return
	}

	response.OKWithMessage(c, gin.H{
		"key":    key,
		"config": config,
	}, "权重配置已更新")
}

// DeleteWeightConfig 删除权重配置
func (h *WeightHandler) DeleteWeightConfig(c *gin.Context) {
	key := c.Param("key")

	h.weightService.DeleteWeights(key)

	response.OKWithMessage(c, gin.H{"key": key}, "权重配置已删除")
}

// ResetWeightConfig 重置权重配置
func (h *WeightHandler) ResetWeightConfig(c *gin.Context) {
	key := c.Param("key")

	h.weightService.ResetToDefault(key)

	response.OKWithMessage(c, gin.H{"key": key}, "权重配置已重置为默认")
}

// ExportWeightConfig 导出权重配置
// 该路由强制返回原始 JSON 字节（作为可下载附件），不进入 pkg/response 包装。
func (h *WeightHandler) ExportWeightConfig(c *gin.Context) {
	key := c.Param("key")

	data, err := h.weightService.ExportWeights(key)
	if err != nil {
		h.logger.Warnf("导出权重配置失败: %v", err)
		response.InternalError(c, "export_failed", "导出权重配置失败", nil)
		return
	}

	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=\"weight_config_"+key+".json\"")
	c.String(http.StatusOK, string(data))
}

// ImportWeightConfig 导入权重配置
func (h *WeightHandler) ImportWeightConfig(c *gin.Context) {
	key := c.Param("key")

	file, err := c.FormFile("file")
	if err != nil {
		h.logger.Warnf("获取上传文件失败: %v", err)
		response.BadRequest(c, "no_file", "请选择要上传的文件", nil)
		return
	}

	if !strings.HasSuffix(file.Filename, ".json") {
		response.BadRequest(c, "invalid_file_type", "只支持JSON格式文件", nil)
		return
	}

	f, err := file.Open()
	if err != nil {
		h.logger.Warnf("打开文件失败: %v", err)
		response.InternalError(c, "file_open_error", "无法打开文件", nil)
		return
	}
	defer f.Close()

	data := make([]byte, file.Size)
	if _, err := f.Read(data); err != nil {
		h.logger.Warnf("读取文件失败: %v", err)
		response.InternalError(c, "file_read_error", "无法读取文件", nil)
		return
	}

	if err := h.weightService.ImportWeights(key, data); err != nil {
		h.logger.Warnf("导入权重配置失败: %v", err)
		response.BadRequest(c, "import_failed", err.Error(), nil)
		return
	}

	response.OKWithMessage(c, gin.H{"key": key}, "权重配置导入成功")
}

// GetWeightStats 获取权重统计信息
func (h *WeightHandler) GetWeightStats(c *gin.Context) {
	stats := h.weightService.GetWeightStats()
	response.OK(c, stats)
}

// CreatePresetWeights 创建预设权重配置
func (h *WeightHandler) CreatePresetWeights(c *gin.Context) {
	h.weightService.CreatePresetWeights()

	response.OKWithMessage(c, gin.H{
		"presets": []string{"score_focused", "location_focused", "employment_focused"},
	}, "预设权重配置已创建")
}

// WeightConfigRequest 权重配置请求（用于文档）
type WeightConfigRequest struct {
	Name                  string  `json:"name" example:"自定义配置"`
	Description           string  `json:"description" example:"我的自定义权重配置"`
	ScoreMatchWeight      float64 `json:"score_match_weight" example:"0.35"`
	LocationWeight        float64 `json:"location_weight" example:"0.15"`
	InterestWeight        float64 `json:"interest_weight" example:"0.15"`
	EmploymentWeight      float64 `json:"employment_weight" example:"0.15"`
	UniversityRankWeight  float64 `json:"university_rank_weight" example:"0.10"`
	CompetitionWeight     float64 `json:"competition_weight" example:"0.10"`
	EnableAdaptiveWeights bool    `json:"enable_adaptive_weights" example:"true"`
	MinWeightThreshold    float64 `json:"min_weight_threshold" example:"0.05"`
	MaxWeightThreshold    float64 `json:"max_weight_threshold" example:"0.40"`
}
