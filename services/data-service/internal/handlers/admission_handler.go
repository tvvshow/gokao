package handlers

import (
	"data-service/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AdmissionHandler 录取数据处理器
type AdmissionHandler struct {
	admissionService *services.AdmissionService
	logger           *logrus.Logger
}

// NewAdmissionHandler 创建录取数据处理器实例
func NewAdmissionHandler(admissionService *services.AdmissionService, logger *logrus.Logger) *AdmissionHandler {
	return &AdmissionHandler{
		admissionService: admissionService,
		logger:           logger,
	}
}

// ListAdmissionData 获取录取数据列表
// @Summary 获取录取数据列表
// @Description 根据筛选条件获取录取数据列表
// @Tags Admission
// @Accept json
// @Produce json
// @Param university_id query string false "院校ID"
// @Param major_id query string false "专业ID"
// @Param year query int false "年份"
// @Param min_year query int false "最小年份"
// @Param max_year query int false "最大年份"
// @Param province query string false "省份"
// @Param batch query string false "录取批次" Enums(early_admission, first_batch, second_batch, third_batch, specialized)
// @Param category query string false "科类" Enums(science, liberal_arts, comprehensive)
// @Param min_score query number false "最低分数"
// @Param max_score query number false "最高分数"
// @Param min_rank query int false "最小排名"
// @Param max_rank query int false "最大排名"
// @Param difficulty query string false "录取难度" Enums(very_easy, easy, medium, hard, very_hard)
// @Param sort_by query string false "排序字段" Enums(year, score, rank)
// @Param sort_order query string false "排序方向" Enums(asc, desc)
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Param include_university query bool false "是否包含院校信息"
// @Param include_major query bool false "是否包含专业信息"
// @Success 200 {object} APIResponse{data=services.AdmissionListResponse} "成功"
// @Failure 400 {object} APIResponse "请求参数错误"
// @Failure 500 {object} APIResponse "服务器错误"
// @Router /api/v1/admission/data [get]
func (h *AdmissionHandler) ListAdmissionData(c *gin.Context) {
	var params services.AdmissionQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		h.logger.Errorf("参数绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, NewErrorResponse("请求参数错误"))
		return
	}

	// 验证排序参数
	if params.SortBy != "" && !isValidSortField(params.SortBy, []string{"year", "score", "rank"}) {
		c.JSON(http.StatusBadRequest, NewErrorResponse("无效的排序字段"))
		return
	}

	response, err := h.admissionService.ListAdmissionData(c.Request.Context(), params)
	if err != nil {
		h.logger.Errorf("获取录取数据列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("获取录取数据列表失败"))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse(response))
}

// AnalyzeAdmissionData 分析录取数据趋势
// @Summary 分析录取数据趋势
// @Description 分析指定院校/专业的录取数据趋势
// @Tags Admission
// @Accept json
// @Produce json
// @Param university_id query string true "院校ID"
// @Param major_id query string false "专业ID"
// @Param province query string true "省份"
// @Param category query string true "科类" Enums(science, liberal_arts, comprehensive)
// @Success 200 {object} APIResponse{data=services.AdmissionAnalysis} "成功"
// @Failure 400 {object} APIResponse "请求参数错误"
// @Failure 500 {object} APIResponse "服务器错误"
// @Router /api/v1/admission/analyze [get]
func (h *AdmissionHandler) AnalyzeAdmissionData(c *gin.Context) {
	universityID := c.Query("university_id")
	if universityID == "" {
		c.JSON(http.StatusBadRequest, NewErrorResponse("院校ID不能为空"))
		return
	}

	province := c.Query("province")
	if province == "" {
		c.JSON(http.StatusBadRequest, NewErrorResponse("省份不能为空"))
		return
	}

	category := c.Query("category")
	if category == "" {
		c.JSON(http.StatusBadRequest, NewErrorResponse("科类不能为空"))
		return
	}

	majorID := c.Query("major_id")

	analysis, err := h.admissionService.AnalyzeAdmissionData(c.Request.Context(), universityID, majorID, province, category)
	if err != nil {
		h.logger.Errorf("分析录取数据失败: %v", err)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("分析录取数据失败"))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse(analysis))
}

// PredictAdmission 预测录取概率
// @Summary 预测录取概率
// @Description 根据分数预测录取概率
// @Tags Admission
// @Accept json
// @Produce json
// @Param request body services.PredictionRequest true "预测请求"
// @Success 200 {object} APIResponse{data=services.PredictionResponse} "成功"
// @Failure 400 {object} APIResponse "请求参数错误"
// @Failure 500 {object} APIResponse "服务器错误"
// @Router /api/v1/admission/predict [post]
func (h *AdmissionHandler) PredictAdmission(c *gin.Context) {
	var req services.PredictionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("参数绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, NewErrorResponse("请求参数错误"))
		return
	}

	// 基本验证
	if req.UniversityID == "" {
		c.JSON(http.StatusBadRequest, NewErrorResponse("院校ID不能为空"))
		return
	}
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

	prediction, err := h.admissionService.PredictAdmission(c.Request.Context(), req)
	if err != nil {
		h.logger.Errorf("预测录取概率失败: %v", err)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("预测录取概率失败"))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse(prediction))
}

// GetAdmissionStatistics 获取录取数据统计
// @Summary 获取录取数据统计
// @Description 获取录取数据的统计信息
// @Tags Admission
// @Accept json
// @Produce json
// @Param year query int false "年份"
// @Success 200 {object} APIResponse{data=map[string]interface{}} "成功"
// @Failure 500 {object} APIResponse "服务器错误"
// @Router /api/v1/admission/statistics [get]
func (h *AdmissionHandler) GetAdmissionStatistics(c *gin.Context) {
	year := 0
	if yearStr := c.Query("year"); yearStr != "" {
		if parsedYear, err := strconv.Atoi(yearStr); err == nil {
			year = parsedYear
		}
	}

	stats, err := h.admissionService.GetAdmissionStatistics(c.Request.Context(), year)
	if err != nil {
		h.logger.Errorf("获取录取数据统计失败: %v", err)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("获取统计信息失败"))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse(stats))
}

// GetBatches 获取录取批次列表
// @Summary 获取录取批次列表
// @Description 获取所有录取批次
// @Tags Admission
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=map[string]string} "成功"
// @Router /api/v1/admission/batches [get]
func (h *AdmissionHandler) GetBatches(c *gin.Context) {
	batches := map[string]string{
		"early_admission": "提前批",
		"first_batch":     "本科一批",
		"second_batch":    "本科二批",
		"third_batch":     "本科三批",
		"specialized":     "专科批",
	}

	c.JSON(http.StatusOK, NewSuccessResponse(batches))
}

// GetCategories 获取科类列表
// @Summary 获取科类列表
// @Description 获取所有科类
// @Tags Admission
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=map[string]string} "成功"
// @Router /api/v1/admission/categories [get]
func (h *AdmissionHandler) GetCategories(c *gin.Context) {
	categories := map[string]string{
		"science":       "理科",
		"liberal_arts":  "文科",
		"comprehensive": "综合",
	}

	c.JSON(http.StatusOK, NewSuccessResponse(categories))
}

// GetDifficulties 获取难度级别列表
// @Summary 获取难度级别列表
// @Description 获取所有录取难度级别
// @Tags Admission
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=map[string]string} "成功"
// @Router /api/v1/admission/difficulties [get]
func (h *AdmissionHandler) GetDifficulties(c *gin.Context) {
	difficulties := map[string]string{
		"very_easy": "很容易",
		"easy":      "容易",
		"medium":    "中等",
		"hard":      "困难",
		"very_hard": "很困难",
	}

	c.JSON(http.StatusOK, NewSuccessResponse(difficulties))
}