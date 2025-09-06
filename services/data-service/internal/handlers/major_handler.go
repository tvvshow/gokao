package handlers

import (
	"data-service/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// MajorHandler 专业处理器
type MajorHandler struct {
	majorService *services.MajorService
	logger       *logrus.Logger
}

// NewMajorHandler 创建专业处理器实例
func NewMajorHandler(majorService *services.MajorService, logger *logrus.Logger) *MajorHandler {
	return &MajorHandler{
		majorService: majorService,
		logger:       logger,
	}
}

// GetMajorByID 根据ID获取专业详情
// @Summary 根据ID获取专业详情
// @Description 根据专业ID获取详细信息
// @Tags Major
// @Accept json
// @Produce json
// @Param id path string true "专业ID"
// @Success 200 {object} APIResponse{data=models.Major} "成功"
// @Failure 400 {object} APIResponse "请求参数错误"
// @Failure 404 {object} APIResponse "专业不存在"
// @Failure 500 {object} APIResponse "服务器错误"
// @Router /api/v1/majors/{id} [get]
func (h *MajorHandler) GetMajorByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, NewErrorResponse("专业ID不能为空"))
		return
	}

	major, err := h.majorService.GetMajorByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Errorf("获取专业详情失败: %v", err)
		if err.Error() == "专业不存在" {
			c.JSON(http.StatusNotFound, NewErrorResponse("专业不存在"))
		} else {
			c.JSON(http.StatusInternalServerError, NewErrorResponse("获取专业详情失败"))
		}
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse(major))
}

// ListMajors 获取专业列表
// @Summary 获取专业列表
// @Description 根据筛选条件获取专业列表
// @Tags Major
// @Accept json
// @Produce json
// @Param university_id query string false "院校ID"
// @Param name query string false "专业名称"
// @Param keyword query string false "关键词搜索"
// @Param category query string false "专业类别"
// @Param discipline query string false "学科"
// @Param sub_discipline query string false "子学科"
// @Param degree_type query string false "学位类型" Enums(bachelor, master, doctoral)
// @Param min_employment_rate query number false "最低就业率"
// @Param max_employment_rate query number false "最高就业率"
// @Param min_salary query number false "最低薪资"
// @Param max_salary query number false "最高薪资"
// @Param min_popularity query number false "最低热度"
// @Param max_popularity query number false "最高热度"
// @Param is_active query bool false "是否激活"
// @Param is_recruiting query bool false "是否招生"
// @Param sort_by query string false "排序字段" Enums(name, popularity, employment_rate, salary)
// @Param sort_order query string false "排序方向" Enums(asc, desc)
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Param include_university query bool false "是否包含院校信息"
// @Param include_admission_data query bool false "是否包含录取数据"
// @Success 200 {object} APIResponse{data=services.MajorListResponse} "成功"
// @Failure 400 {object} APIResponse "请求参数错误"
// @Failure 500 {object} APIResponse "服务器错误"
// @Router /api/v1/majors [get]
func (h *MajorHandler) ListMajors(c *gin.Context) {
	var params services.MajorQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		h.logger.Errorf("参数绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, NewErrorResponse("请求参数错误"))
		return
	}

	// 验证排序参数
	if params.SortBy != "" && !isValidSortField(params.SortBy, []string{"name", "popularity", "employment_rate", "salary"}) {
		c.JSON(http.StatusBadRequest, NewErrorResponse("无效的排序字段"))
		return
	}

	response, err := h.majorService.ListMajors(c.Request.Context(), params)
	if err != nil {
		h.logger.Errorf("获取专业列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("获取专业列表失败"))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse(response))
}

// SearchMajors 搜索专业
// @Summary 搜索专业
// @Description 使用关键词搜索专业
// @Tags Major
// @Accept json
// @Produce json
// @Param q query string true "搜索关键词"
// @Param university_id query string false "院校ID"
// @Param category query string false "专业类别"
// @Param discipline query string false "学科"
// @Param degree_type query string false "学位类型" Enums(bachelor, master, doctoral)
// @Param min_employment_rate query number false "最低就业率"
// @Param max_employment_rate query number false "最高就业率"
// @Param min_salary query number false "最低薪资"
// @Param max_salary query number false "最高薪资"
// @Param sort_by query string false "排序字段" Enums(name, popularity, employment_rate, salary, relevance)
// @Param sort_order query string false "排序方向" Enums(asc, desc)
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Param include_university query bool false "是否包含院校信息"
// @Param include_admission_data query bool false "是否包含录取数据"
// @Success 200 {object} APIResponse{data=services.MajorListResponse} "成功"
// @Failure 400 {object} APIResponse "请求参数错误"
// @Failure 500 {object} APIResponse "服务器错误"
// @Router /api/v1/majors/search [get]
func (h *MajorHandler) SearchMajors(c *gin.Context) {
	keyword := c.Query("q")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, NewErrorResponse("搜索关键词不能为空"))
		return
	}

	var params services.MajorQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		h.logger.Errorf("参数绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, NewErrorResponse("请求参数错误"))
		return
	}

	response, err := h.majorService.SearchMajors(c.Request.Context(), keyword, params)
	if err != nil {
		h.logger.Errorf("搜索专业失败: %v", err)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("搜索专业失败"))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse(response))
}

// GetMajorCategories 获取专业类别列表
// @Summary 获取专业类别列表
// @Description 获取所有专业类别
// @Tags Major
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=[]string} "成功"
// @Failure 500 {object} APIResponse "服务器错误"
// @Router /api/v1/majors/categories [get]
func (h *MajorHandler) GetMajorCategories(c *gin.Context) {
	categories, err := h.majorService.GetMajorCategories(c.Request.Context())
	if err != nil {
		h.logger.Errorf("获取专业类别失败: %v", err)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("获取专业类别失败"))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse(categories))
}

// GetMajorDisciplines 获取学科列表
// @Summary 获取学科列表
// @Description 获取所有学科
// @Tags Major
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=[]string} "成功"
// @Failure 500 {object} APIResponse "服务器错误"
// @Router /api/v1/majors/disciplines [get]
func (h *MajorHandler) GetMajorDisciplines(c *gin.Context) {
	disciplines, err := h.majorService.GetMajorDisciplines(c.Request.Context())
	if err != nil {
		h.logger.Errorf("获取学科列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("获取学科列表失败"))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse(disciplines))
}

// GetMajorStatistics 获取专业统计信息
// @Summary 获取专业统计信息
// @Description 获取专业的统计信息
// @Tags Major
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=map[string]interface{}} "成功"
// @Failure 500 {object} APIResponse "服务器错误"
// @Router /api/v1/majors/statistics [get]
func (h *MajorHandler) GetMajorStatistics(c *gin.Context) {
	stats, err := h.majorService.GetMajorStatistics(c.Request.Context())
	if err != nil {
		h.logger.Errorf("获取专业统计失败: %v", err)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("获取统计信息失败"))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse(stats))
}

// GetDegreeTypes 获取学位类型列表
// @Summary 获取学位类型列表
// @Description 获取所有学位类型
// @Tags Major
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=map[string]string} "成功"
// @Router /api/v1/majors/degree-types [get]
func (h *MajorHandler) GetDegreeTypes(c *gin.Context) {
	degreeTypes := map[string]string{
		"bachelor":  "学士",
		"master":    "硕士",
		"doctoral":  "博士",
	}

	c.JSON(http.StatusOK, NewSuccessResponse(degreeTypes))
}