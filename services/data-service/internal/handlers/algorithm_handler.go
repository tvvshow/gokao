package handlers

import (
	"github.com/oktetopython/gaokao/services/data-service/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AlgorithmHandler 算法处理器
type AlgorithmHandler struct {
	algorithmService *services.AlgorithmService
	logger           *logrus.Logger
}

// NewAlgorithmHandler 创建算法处理器实例
func NewAlgorithmHandler(algorithmService *services.AlgorithmService, logger *logrus.Logger) *AlgorithmHandler {
	return &AlgorithmHandler{
		algorithmService: algorithmService,
		logger:           logger,
	}
}

// MatchVolunteers 志愿匹配
// @Summary 志愿匹配
// @Description 基于考生信息和偏好进行志愿匹配推荐
// @Tags Algorithm
// @Accept json
// @Produce json
// @Param request body services.VolunteerMatchRequest true "志愿匹配请求"
// @Success 200 {object} APIResponse{data=services.VolunteerMatchResponse} "成功"
// @Failure 400 {object} APIResponse "请求参数错误"
// @Failure 500 {object} APIResponse "服务器错误"
// @Router /api/v1/algorithm/match [post]
func (h *AlgorithmHandler) MatchVolunteers(c *gin.Context) {
	var req services.VolunteerMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("参数绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, NewErrorResponse("请求参数错误"))
		return
	}

	// 基本验证
	if req.Province == "" {
		c.JSON(http.StatusBadRequest, NewErrorResponse("省份不能为空"))
		return
	}
	if req.Category == "" {
		c.JSON(http.StatusBadRequest, NewErrorResponse("科类不能为空"))
		return
	}
	if req.Score <= 0 || req.Score > 1000 {
		c.JSON(http.StatusBadRequest, NewErrorResponse("分数必须在0-1000之间"))
		return
	}

	// 验证风险偏好
	validRiskTolerance := map[string]bool{
		"conservative": true,
		"moderate":     true,
		"aggressive":   true,
	}
	if req.RiskTolerance != "" && !validRiskTolerance[req.RiskTolerance] {
		c.JSON(http.StatusBadRequest, NewErrorResponse("无效的风险偏好"))
		return
	}

	response, err := h.algorithmService.MatchVolunteers(c.Request.Context(), req)
	if err != nil {
		h.logger.Errorf("志愿匹配失败: %v", err)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("志愿匹配失败"))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse(response))
}

// GetRiskToleranceOptions 获取风险偏好选项
// @Summary 获取风险偏好选项
// @Description 获取所有风险偏好选项
// @Tags Algorithm
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=map[string]string} "成功"
// @Router /api/v1/algorithm/risk-tolerance [get]
func (h *AlgorithmHandler) GetRiskToleranceOptions(c *gin.Context) {
	options := map[string]string{
		"conservative": "保守型",
		"moderate":     "稳健型",
		"aggressive":   "激进型",
	}

	c.JSON(http.StatusOK, NewSuccessResponse(options))
}

// GetRecommendTypes 获取推荐类型
// @Summary 获取推荐类型
// @Description 获取所有推荐类型
// @Tags Algorithm
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=map[string]string} "成功"
// @Router /api/v1/algorithm/recommend-types [get]
func (h *AlgorithmHandler) GetRecommendTypes(c *gin.Context) {
	types := map[string]string{
		"safe":     "保底志愿",
		"moderate": "稳妥志愿",
		"reach":    "冲刺志愿",
	}

	c.JSON(http.StatusOK, NewSuccessResponse(types))
}
