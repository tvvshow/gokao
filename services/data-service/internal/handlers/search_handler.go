package handlers

import (
	"data-service/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// SearchHandler 搜索处理器
type SearchHandler struct {
	searchService *services.SearchService
	logger        *logrus.Logger
}

// NewSearchHandler 创建搜索处理器实例
func NewSearchHandler(searchService *services.SearchService, logger *logrus.Logger) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
		logger:        logger,
	}
}

// GlobalSearch 全局搜索
// @Summary 全局搜索
// @Description 在院校和专业中进行全局搜索
// @Tags Search
// @Accept json
// @Produce json
// @Param keyword query string true "搜索关键词"
// @Param types query array false "搜索类型" collectionFormat(multi) Enums(university, major)
// @Param province query string false "省份"
// @Param category query string false "类别"
// @Param sort_by query string false "排序字段" Enums(relevance, popularity, name)
// @Param sort_order query string false "排序方向" Enums(asc, desc)
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Success 200 {object} APIResponse{data=services.SearchResponse} "成功"
// @Failure 400 {object} APIResponse "请求参数错误"
// @Failure 500 {object} APIResponse "服务器错误"
// @Router /api/v1/search [get]
func (h *SearchHandler) GlobalSearch(c *gin.Context) {
	var req services.SearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Errorf("参数绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, NewErrorResponse("请求参数错误"))
		return
	}

	// 验证必需参数
	if req.Keyword == "" {
		c.JSON(http.StatusBadRequest, NewErrorResponse("搜索关键词不能为空"))
		return
	}

	// 处理数组参数
	if types := c.QueryArray("types"); len(types) > 0 {
		req.Types = types
	}

	// 验证搜索类型
	validTypes := map[string]bool{"university": true, "major": true}
	for _, t := range req.Types {
		if !validTypes[t] {
			c.JSON(http.StatusBadRequest, NewErrorResponse("无效的搜索类型: "+t))
			return
		}
	}

	// 验证排序参数
	if req.SortBy != "" {
		validSortFields := map[string]bool{"relevance": true, "popularity": true, "name": true}
		if !validSortFields[req.SortBy] {
			c.JSON(http.StatusBadRequest, NewErrorResponse("无效的排序字段"))
			return
		}
	}

	response, err := h.searchService.GlobalSearch(c.Request.Context(), req)
	if err != nil {
		h.logger.Errorf("全局搜索失败: %v", err)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("搜索失败"))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse(response))
}

// AutoComplete 自动补全
// @Summary 搜索自动补全
// @Description 提供搜索关键词的自动补全建议
// @Tags Search
// @Accept json
// @Produce json
// @Param keyword query string true "搜索关键词"
// @Param type query string false "搜索类型" Enums(university, major, all)
// @Param limit query int false "限制数量" default(10)
// @Success 200 {object} APIResponse{data=services.AutoCompleteResponse} "成功"
// @Failure 400 {object} APIResponse "请求参数错误"
// @Failure 500 {object} APIResponse "服务器错误"
// @Router /api/v1/search/autocomplete [get]
func (h *SearchHandler) AutoComplete(c *gin.Context) {
	var req services.AutoCompleteRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Errorf("参数绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, NewErrorResponse("请求参数错误"))
		return
	}

	// 验证必需参数
	if req.Keyword == "" {
		c.JSON(http.StatusBadRequest, NewErrorResponse("搜索关键词不能为空"))
		return
	}

	// 验证搜索类型
	if req.Type != "" {
		validTypes := map[string]bool{"university": true, "major": true, "all": true}
		if !validTypes[req.Type] {
			c.JSON(http.StatusBadRequest, NewErrorResponse("无效的搜索类型"))
			return
		}
	}

	response, err := h.searchService.AutoComplete(c.Request.Context(), req)
	if err != nil {
		h.logger.Errorf("自动补全失败: %v", err)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("自动补全失败"))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse(response))
}

// GetHotSearches 获取热搜关键词
// @Summary 获取热搜关键词
// @Description 获取当前热门搜索关键词
// @Tags Search
// @Accept json
// @Produce json
// @Param category query string false "类别"
// @Param limit query int false "限制数量" default(10)
// @Success 200 {object} APIResponse{data=services.HotSearchResponse} "成功"
// @Failure 400 {object} APIResponse "请求参数错误"
// @Failure 500 {object} APIResponse "服务器错误"
// @Router /api/v1/search/hot [get]
func (h *SearchHandler) GetHotSearches(c *gin.Context) {
	category := c.Query("category")
	
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if limit > 50 {
		limit = 50
	}

	response, err := h.searchService.GetHotSearches(c.Request.Context(), category, limit)
	if err != nil {
		h.logger.Errorf("获取热搜失败: %v", err)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("获取热搜失败"))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse(response))
}

// GetSearchSuggestions 获取搜索建议
// @Summary 获取搜索建议
// @Description 根据关键词获取搜索建议
// @Tags Search
// @Accept json
// @Produce json
// @Param keyword query string true "搜索关键词"
// @Success 200 {object} APIResponse{data=[]string} "成功"
// @Failure 400 {object} APIResponse "请求参数错误"
// @Failure 500 {object} APIResponse "服务器错误"
// @Router /api/v1/search/suggestions [get]
func (h *SearchHandler) GetSearchSuggestions(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, NewErrorResponse("搜索关键词不能为空"))
		return
	}

	suggestions, err := h.searchService.GetSearchSuggestions(c.Request.Context(), keyword)
	if err != nil {
		h.logger.Errorf("获取搜索建议失败: %v", err)
		c.JSON(http.StatusInternalServerError, NewErrorResponse("获取搜索建议失败"))
		return
	}

	c.JSON(http.StatusOK, NewSuccessResponse(suggestions))
}