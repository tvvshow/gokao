package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oktetopython/gaokao/recommendation-service/pkg/cppbridge"
)

// SimpleRecommendationHandler 简化的推荐处理器
type SimpleRecommendationHandler struct {
	bridge cppbridge.HybridRecommendationBridge
}

// NewSimpleRecommendationHandler 创建新的简化推荐处理器
func NewSimpleRecommendationHandler(bridge cppbridge.HybridRecommendationBridge) *SimpleRecommendationHandler {
	return &SimpleRecommendationHandler{
		bridge: bridge,
	}
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// GenerateRecommendations 生成推荐
// @Summary 生成志愿推荐
// @Description 根据学生信息生成个性化的志愿填报推荐
// @Tags recommendations
// @Accept json
// @Produce json
// @Param request body cppbridge.RecommendationRequest true "推荐请求"
// @Success 200 {object} cppbridge.RecommendationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/recommendations/generate [post]
func (h *SimpleRecommendationHandler) GenerateRecommendations(c *gin.Context) {
	var request cppbridge.RecommendationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "请求格式错误: " + err.Error(),
		})
		return
	}

	// 设置默认值
	if request.MaxRecommendations == 0 {
		request.MaxRecommendations = 30
	}
	if request.Algorithm == "" {
		request.Algorithm = "hybrid"
	}

	response, err := h.bridge.GenerateRecommendations(&request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "generation_failed",
			Message: "生成推荐失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
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
// @Router /api/v1/recommendations/batch [post]
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
		batchRequest.BatchSize = 10
	}
	if batchRequest.Timeout == 0 {
		batchRequest.Timeout = 30000 // 30秒
	}

	var responses []cppbridge.RecommendationResponse
	successCount := 0
	failedCount := 0

	for _, request := range batchRequest.Requests {
		if request.MaxRecommendations == 0 {
			request.MaxRecommendations = 30
		}
		if request.Algorithm == "" {
			request.Algorithm = batchRequest.Algorithm
		}

		response, err := h.bridge.GenerateRecommendations(&request)
		if err != nil {
			failedCount++
			responses = append(responses, cppbridge.RecommendationResponse{
				StudentID:    request.StudentID,
				Success:      false,
				ErrorMessage: err.Error(),
			})
		} else {
			successCount++
			responses = append(responses, *response)
		}
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
// @Router /api/v1/recommendations/explain/{id} [get]
func (h *SimpleRecommendationHandler) ExplainRecommendation(c *gin.Context) {
	recommendationID := c.Param("id")
	if recommendationID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_id",
			Message: "推荐ID不能为空",
		})
		return
	}

	// 模拟解释逻辑
	explanation := map[string]interface{}{
		"recommendation_id": recommendationID,
		"explanation": map[string]interface{}{
			"score_match": "您的分数与该专业历年录取分数线匹配度为85%",
			"interest_match": "根据您的兴趣偏好，该专业符合度为90%",
			"employment_prospect": "该专业就业前景良好，就业率达95%",
			"location_preference": "该学校位于您的首选城市",
		},
		"factors": []string{
			"分数匹配度高",
			"专业发展前景好",
			"地理位置优越",
			"学校声誉良好",
		},
		"risks": []string{
			"竞争较为激烈",
			"专业要求较高",
		},
		"alternatives": []string{
			"可以考虑相关专业作为备选",
			"建议关注该校其他优势专业",
		},
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
// @Router /api/v1/recommendations/optimize [post]
func (h *SimpleRecommendationHandler) OptimizeRecommendations(c *gin.Context) {
	var request map[string]interface{}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "优化请求格式错误: " + err.Error(),
		})
		return
	}

	// 模拟优化逻辑
	result := map[string]interface{}{
		"optimization_id": "opt_" + strconv.FormatInt(time.Now().Unix(), 10),
		"status": "completed",
		"improvements": map[string]interface{}{
			"accuracy_increase": 0.05,
			"diversity_improvement": 0.03,
			"user_satisfaction": 0.08,
		},
		"new_weights": map[string]float64{
			"traditional": 0.65,
			"ai": 0.35,
		},
		"message": "推荐算法已根据反馈进行优化",
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
// @Router /api/v1/system/status [get]
func (h *SimpleRecommendationHandler) GetSystemStatus(c *gin.Context) {
	status, err := h.bridge.GetSystemStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "status_error",
			Message: "获取系统状态失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// ClearCache 清空缓存
// @Summary 清空系统缓存
// @Description 清空推荐系统的缓存数据
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/system/cache/clear [post]
func (h *SimpleRecommendationHandler) ClearCache(c *gin.Context) {
	err := h.bridge.ClearCache()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "cache_error",
			Message: "清空缓存失败: " + err.Error(),
		})
		return
	}

	result := map[string]interface{}{
		"status": "success",
		"message": "缓存已清空",
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
// @Router /api/v1/system/model/update [put]
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
		"status": "success",
		"message": "模型更新成功",
		"model_path": modelPath,
		"timestamp": time.Now().Unix(),
	}

	c.JSON(http.StatusOK, result)
}