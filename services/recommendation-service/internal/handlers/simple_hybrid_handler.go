package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oktetopython/gaokao/recommendation-service/pkg/cppbridge"
)

// SimpleHybridHandler 简化的混合推荐处理器
type SimpleHybridHandler struct {
	bridge cppbridge.HybridRecommendationBridge
}

// NewSimpleHybridHandler 创建新的简化混合推荐处理器
func NewSimpleHybridHandler(bridge cppbridge.HybridRecommendationBridge) *SimpleHybridHandler {
	return &SimpleHybridHandler{
		bridge: bridge,
	}
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
	if weights.TraditionalWeight < 0 || weights.TraditionalWeight > 1 ||
		weights.AIWeight < 0 || weights.AIWeight > 1 ||
		weights.TraditionalWeight+weights.AIWeight != 1.0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_weight_values",
			Message: "权重值必须在0-1之间，且两者之和必须为1",
		})
		return
	}

	weightMap := map[string]float64{
		"traditional": weights.TraditionalWeight,
		"ai":         weights.AIWeight,
		"diversity":  weights.DiversityFactor,
	}

	err := h.bridge.UpdateFusionWeights(weightMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "update_failed",
			Message: "更新权重失败: " + err.Error(),
		})
		return
	}

	result := map[string]interface{}{
		"status": "success",
		"message": "融合权重更新成功",
		"weights": weightMap,
		"updated_at": time.Now().Unix(),
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

	comparison, err := h.bridge.CompareRecommendations(&request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "comparison_failed",
			Message: "比较推荐结果失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, comparison)
}