package handlers

import (
	"github.com/tvvshow/gokao/services/data-service/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// UniversityHandler 院校处理器
type UniversityHandler struct {
	universityService *services.UniversityService
	logger            *logrus.Logger
}

// NewUniversityHandler 创建院校处理器实例
func NewUniversityHandler(universityService *services.UniversityService, logger *logrus.Logger) *UniversityHandler {
	return &UniversityHandler{
		universityService: universityService,
		logger:            logger,
	}
}

// GetUniversityByID 根据ID获取院校详情
// @Summary 根据ID获取院校详情
// @Description 根据院校ID获取详细信息
// @Tags University
// @Accept json
// @Produce json
// @Param id path string true "院校ID"
// @Success 200 {object} APIResponse{data=models.University} "成功"
// @Failure 400 {object} APIResponse "请求参数错误"
// @Failure 404 {object} APIResponse "院校不存在"
// @Failure 500 {object} APIResponse "服务器错误"
// @Router /api/v1/universities/{id} [get]
func (h *UniversityHandler) GetUniversityByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, NewErrorResponse("院校ID不能为空"))
		return
	}

	university, err := h.universityService.GetUniversityByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Errorf("获取院校详情失败: %v", err)
		if err.Error() == "院校不存在" {
			c.JSON(http.StatusNotFound, NewErrorResponse("院校不存在"))
		} else {
			c.JSON(http.StatusInternalServerError, NewErrorResponse("获取院校详情失败"))
		}
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse(university))
}

// GetUniversityByCode 根据代码获取院校详情
// @Summary 根据代码获取院校详情
// @Description 根据院校代码获取详细信息
// @Tags University
// @Accept json
// @Produce json
// @Param code path string true "院校代码"
// @Success 200 {object} APIResponse{data=models.University} "成功"
// @Failure 400 {object} APIResponse "请求参数错误"
// @Failure 404 {object} APIResponse "院校不存在"
// @Failure 500 {object} APIResponse "服务器错误"
// @Router /api/v1/universities/code/{code} [get]
func (h *UniversityHandler) GetUniversityByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, NewErrorResponse("院校代码不能为空"))
		return
	}

	university, err := h.universityService.GetUniversityByCode(c.Request.Context(), code)
	if err != nil {
		h.logger.Errorf("获取院校详情失败: %v", err)
		if err.Error() == "院校不存在" {
			c.JSON(http.StatusNotFound, NewErrorResponse("院校不存在"))
		} else {
			c.JSON(http.StatusInternalServerError, NewErrorResponse("获取院校详情失败"))
		}
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse(university))
}

// ListUniversities 获取院校列表
// @Summary 获取院校列表
// @Description 根据筛选条件获取院校列表
// @Tags University
// @Accept json
// @Produce json
// @Param name query string false "院校名称"
// @Param keyword query string false "关键词搜索"
// @Param type query string false "院校类型" Enums(undergraduate, graduate, vocational)
// @Param level query string false "院校层次" Enums(985, 211, double_first_class, ordinary)
// @Param nature query string false "院校性质" Enums(public, private, joint_venture)
// @Param category query string false "院校类别"
// @Param province query string false "省份"
// @Param city query string false "城市"
// @Param min_rank query int false "最小排名"
// @Param max_rank query int false "最大排名"
// @Param is_active query bool false "是否激活"
// @Param is_recruiting query bool false "是否招生"
// @Param sort_by query string false "排序字段" Enums(name, rank, score, created_at)
// @Param sort_order query string false "排序方向" Enums(asc, desc)
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Param include_majors query bool false "是否包含专业信息"
// @Success 200 {object} APIResponse{data=services.UniversityListResponse} "成功"
// @Failure 400 {object} APIResponse "请求参数错误"
// @Failure 500 {object} APIResponse "服务器错误"
// @Router /api/v1/universities [get]
func (h *UniversityHandler) ListUniversities(c *gin.Context) {
	var params services.UniversityQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		h.logger.Errorf("参数绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, NewErrorResponse("请求参数错误"))
		return
	}

	// 验证排序参数
	if params.SortBy != "" && !isValidSortField(params.SortBy, []string{"name", "rank", "score", "created_at"}) {
		c.JSON(http.StatusBadRequest, NewErrorResponse("无效的排序字段"))
		return
	}

	response, err := h.universityService.ListUniversities(c.Request.Context(), params)
	if err != nil {
		h.logger.Errorf("获取院校列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("获取院校列表失败"))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse(response))
}

// SearchUniversities 搜索院校
// @Summary 搜索院校
// @Description 使用关键词搜索院校
// @Tags University
// @Accept json
// @Produce json
// @Param q query string true "搜索关键词"
// @Param type query string false "院校类型" Enums(undergraduate, graduate, vocational)
// @Param level query string false "院校层次" Enums(985, 211, double_first_class, ordinary)
// @Param nature query string false "院校性质" Enums(public, private, joint_venture)
// @Param province query string false "省份"
// @Param city query string false "城市"
// @Param min_rank query int false "最小排名"
// @Param max_rank query int false "最大排名"
// @Param sort_by query string false "排序字段" Enums(name, rank, score, relevance)
// @Param sort_order query string false "排序方向" Enums(asc, desc)
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Param include_majors query bool false "是否包含专业信息"
// @Success 200 {object} APIResponse{data=services.UniversityListResponse} "成功"
// @Failure 400 {object} APIResponse "请求参数错误"
// @Failure 500 {object} APIResponse "服务器错误"
// @Router /api/v1/universities/search [get]
func (h *UniversityHandler) SearchUniversities(c *gin.Context) {
	keyword := c.Query("q")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, NewErrorResponse("搜索关键词不能为空"))
		return
	}

	var params services.UniversityQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		h.logger.Errorf("参数绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, NewErrorResponse("请求参数错误"))
		return
	}

	response, err := h.universityService.SearchUniversities(c.Request.Context(), keyword, params)
	if err != nil {
		h.logger.Errorf("搜索院校失败: %v", err)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("搜索院校失败"))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse(response))
}

// GetUniversityStatistics 获取院校统计信息
// @Summary 获取院校统计信息
// @Description 获取院校的统计信息
// @Tags University
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=map[string]interface{}} "成功"
// @Failure 500 {object} APIResponse "服务器错误"
// @Router /api/v1/universities/statistics [get]
func (h *UniversityHandler) GetUniversityStatistics(c *gin.Context) {
	stats, err := h.universityService.GetUniversityStatistics(c.Request.Context())
	if err != nil {
		h.logger.Errorf("获取院校统计失败: %v", err)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("获取统计信息失败"))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse(stats))
}

// GetUniversityProvinces 获取所有省份列表
// @Summary 获取所有省份列表
// @Description 获取有院校的所有省份列表
// @Tags University
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=[]string} "成功"
// @Failure 500 {object} APIResponse "服务器错误"
// @Router /api/v1/universities/provinces [get]
func (h *UniversityHandler) GetUniversityProvinces(c *gin.Context) {
	// 这里可以从缓存或数据库获取省份列表
	provinces := []string{
		"北京市", "天津市", "河北省", "山西省", "内蒙古自治区",
		"辽宁省", "吉林省", "黑龙江省", "上海市", "江苏省",
		"浙江省", "安徽省", "福建省", "江西省", "山东省",
		"河南省", "湖北省", "湖南省", "广东省", "广西壮族自治区",
		"海南省", "重庆市", "四川省", "贵州省", "云南省",
		"西藏自治区", "陕西省", "甘肃省", "青海省", "宁夏回族自治区",
		"新疆维吾尔自治区", "台湾省", "香港特别行政区", "澳门特别行政区",
	}

	c.JSON(http.StatusOK, NewSuccessResponse(provinces))
}

// GetUniversityTypes 获取院校类型列表
// @Summary 获取院校类型列表
// @Description 获取所有院校类型
// @Tags University
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=map[string]string} "成功"
// @Router /api/v1/universities/types [get]
func (h *UniversityHandler) GetUniversityTypes(c *gin.Context) {
	types := map[string]string{
		"undergraduate": "本科院校",
		"graduate":      "研究生院校",
		"vocational":    "高职院校",
	}

	c.JSON(http.StatusOK, NewSuccessResponse(types))
}

// GetUniversityLevels 获取院校层次列表
// @Summary 获取院校层次列表
// @Description 获取所有院校层次
// @Tags University
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=map[string]string} "成功"
// @Router /api/v1/universities/levels [get]
func (h *UniversityHandler) GetUniversityLevels(c *gin.Context) {
	levels := map[string]string{
		"985":                "985工程",
		"211":                "211工程",
		"double_first_class": "双一流",
		"ordinary":           "普通院校",
	}

	c.JSON(http.StatusOK, NewSuccessResponse(levels))
}

// 辅助函数

// isValidSortField 检查排序字段是否有效
func isValidSortField(field string, validFields []string) bool {
	for _, validField := range validFields {
		if field == validField {
			return true
		}
	}
	return false
}
