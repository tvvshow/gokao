package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/tvvshow/gokao/services/data-service/internal/database"
	"github.com/tvvshow/gokao/services/data-service/internal/models"
	"time"

	"github.com/olivere/elastic/v7"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// SearchService 搜索服务
type SearchService struct {
	db     *database.DB
	logger *logrus.Logger
}

// NewSearchService 创建搜索服务实例
func NewSearchService(db *database.DB, logger *logrus.Logger) *SearchService {
	return &SearchService{
		db:     db,
		logger: logger,
	}
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Keyword   string   `json:"keyword" form:"keyword" validate:"required"`
	Types     []string `json:"types" form:"types"` // university, major
	Province  string   `json:"province" form:"province"`
	Category  string   `json:"category" form:"category"`
	Page      int      `json:"page" form:"page,default=1"`
	PageSize  int      `json:"page_size" form:"page_size,default=20"`
	SortBy    string   `json:"sort_by" form:"sort_by"`       // relevance, popularity, name
	SortOrder string   `json:"sort_order" form:"sort_order"` // asc, desc
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Universities []models.University `json:"universities,omitempty"`
	Majors       []models.Major      `json:"majors,omitempty"`
	Total        int64               `json:"total"`
	Page         int                 `json:"page"`
	PageSize     int                 `json:"page_size"`
	TotalPages   int                 `json:"total_pages"`
	Suggestions  []string            `json:"suggestions,omitempty"`
	HotKeywords  []string            `json:"hot_keywords,omitempty"`
}

// AutoCompleteRequest 自动补全请求
type AutoCompleteRequest struct {
	Keyword string `json:"keyword" form:"keyword" validate:"required"`
	Type    string `json:"type" form:"type"` // university, major, all
	Limit   int    `json:"limit" form:"limit,default=10"`
}

// AutoCompleteResponse 自动补全响应
type AutoCompleteResponse struct {
	Suggestions []AutoCompleteSuggestion `json:"suggestions"`
}

// AutoCompleteSuggestion 自动补全建议
type AutoCompleteSuggestion struct {
	Text     string `json:"text"`
	Type     string `json:"type"` // university, major
	Category string `json:"category,omitempty"`
	ID       string `json:"id,omitempty"`
}

// HotSearchResponse 热搜响应
type HotSearchResponse struct {
	Keywords []HotKeyword `json:"keywords"`
	Date     time.Time    `json:"date"`
}

// HotKeyword 热搜关键词
type HotKeyword struct {
	Keyword     string `json:"keyword"`
	SearchCount uint64 `json:"search_count"`
	Category    string `json:"category,omitempty"`
	Trend       string `json:"trend"` // up, down, stable, new
}

// GlobalSearch 全局搜索
func (s *SearchService) GlobalSearch(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	// 验证分页参数
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > s.db.Config.MaxPageSize {
		req.PageSize = s.db.Config.DefaultPageSize
	}

	// 记录搜索日志
	s.recordSearch(ctx, req.Keyword, req.Province, req.Category)

	// 如果有Elasticsearch，使用ES搜索
	if s.db.Elasticsearch != nil {
		return s.searchWithElasticsearch(ctx, req)
	}

	// 否则使用数据库搜索
	return s.searchWithDatabase(ctx, req)
}

// AutoComplete 自动补全
func (s *SearchService) AutoComplete(ctx context.Context, req AutoCompleteRequest) (*AutoCompleteResponse, error) {
	if req.Limit < 1 || req.Limit > 20 {
		req.Limit = 10
	}

	// 生成缓存键
	cacheKey := fmt.Sprintf("autocomplete:%s:%s:%d", req.Type, req.Keyword, req.Limit)

	// 尝试从缓存获取
	if s.db.Redis != nil && s.db.Config.CacheEnabled {
		cached, err := s.db.Redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var response AutoCompleteResponse
			if err := json.Unmarshal([]byte(cached), &response); err == nil {
				return &response, nil
			}
		} else if err != redis.Nil {
			s.logger.Warnf("获取缓存失败: %v", err)
		}
	}

	var suggestions []AutoCompleteSuggestion

	// 搜索院校
	if req.Type == "" || req.Type == "all" || req.Type == "university" {
		universitySuggestions := s.getUniversitySuggestions(ctx, req.Keyword, req.Limit)
		suggestions = append(suggestions, universitySuggestions...)
	}

	// 搜索专业
	if req.Type == "" || req.Type == "all" || req.Type == "major" {
		majorSuggestions := s.getMajorSuggestions(ctx, req.Keyword, req.Limit)
		suggestions = append(suggestions, majorSuggestions...)
	}

	// 限制总数
	if len(suggestions) > req.Limit {
		suggestions = suggestions[:req.Limit]
	}

	response := &AutoCompleteResponse{
		Suggestions: suggestions,
	}

	// 缓存结果
	if s.db.Redis != nil && s.db.Config.CacheEnabled {
		data, _ := json.Marshal(response)
		s.db.Redis.Set(ctx, cacheKey, data, 5*time.Minute).Err()
	}

	return response, nil
}

// GetHotSearches 获取热搜关键词
func (s *SearchService) GetHotSearches(ctx context.Context, category string, limit int) (*HotSearchResponse, error) {
	if limit < 1 || limit > 50 {
		limit = 10
	}

	// 生成缓存键
	cacheKey := fmt.Sprintf("hot_searches:%s:%d", category, limit)

	// 尝试从缓存获取
	if s.db.Redis != nil && s.db.Config.CacheEnabled {
		cached, err := s.db.Redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var response HotSearchResponse
			if err := json.Unmarshal([]byte(cached), &response); err == nil {
				return &response, nil
			}
		}
	}

	// 从数据库获取热搜数据
	var hotSearches []models.HotSearch
	query := s.db.PostgreSQL.Model(&models.HotSearch{}).
		Where("date >= ?", time.Now().AddDate(0, 0, -7)) // 近7天数据

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if err := query.Order("search_count DESC").Limit(limit).Find(&hotSearches).Error; err != nil {
		return nil, fmt.Errorf("查询热搜数据失败: %w", err)
	}

	// 构建响应
	keywords := make([]HotKeyword, len(hotSearches))
	for i, hs := range hotSearches {
		keywords[i] = HotKeyword{
			Keyword:     hs.Keyword,
			SearchCount: hs.SearchCount,
			Category:    hs.Category,
			Trend:       s.calculateTrend(ctx, hs.Keyword), // 计算趋势
		}
	}

	response := &HotSearchResponse{
		Keywords: keywords,
		Date:     time.Now(),
	}

	// 缓存结果
	if s.db.Redis != nil && s.db.Config.CacheEnabled {
		data, _ := json.Marshal(response)
		s.db.Redis.Set(ctx, cacheKey, data, 10*time.Minute).Err()
	}

	return response, nil
}

// GetSearchSuggestions 获取搜索建议
func (s *SearchService) GetSearchSuggestions(ctx context.Context, keyword string) ([]string, error) {
	// 基于历史搜索数据生成建议
	var suggestions []string

	// 从热搜中获取相似关键词
	var hotKeywords []string
	s.db.PostgreSQL.Model(&models.HotSearch{}).
		Select("keyword").
		Where("LOWER(keyword) LIKE LOWER(?)", "%"+keyword+"%").
		Where("date >= ?", time.Now().AddDate(0, 0, -30)).
		Order("search_count DESC").
		Limit(5).
		Pluck("keyword", &hotKeywords)

	suggestions = append(suggestions, hotKeywords...)

	// 从院校名称中获取建议
	var universityNames []string
	s.db.PostgreSQL.Model(&models.University{}).
		Select("name").
		Where("LOWER(name) LIKE LOWER(?) OR LOWER(alias) LIKE LOWER(?)", "%"+keyword+"%", "%"+keyword+"%").
		Where("is_active = ?", true).
		Limit(3).
		Pluck("name", &universityNames)

	suggestions = append(suggestions, universityNames...)

	// 从专业名称中获取建议
	var majorNames []string
	s.db.PostgreSQL.Model(&models.Major{}).
		Select("name").
		Where("LOWER(name) LIKE LOWER(?)", "%"+keyword+"%").
		Where("is_active = ?", true).
		Limit(3).
		Pluck("name", &majorNames)

	suggestions = append(suggestions, majorNames...)

	// 去重并限制数量
	uniqueSuggestions := s.removeDuplicates(suggestions)
	if len(uniqueSuggestions) > 10 {
		uniqueSuggestions = uniqueSuggestions[:10]
	}

	return uniqueSuggestions, nil
}

// 私有方法

// searchWithElasticsearch 使用Elasticsearch搜索
func (s *SearchService) searchWithElasticsearch(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	response := &SearchResponse{
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	// 搜索院校
	if len(req.Types) == 0 || s.containsType(req.Types, "university") {
		universities, total, err := s.searchUniversitiesES(ctx, req)
		if err != nil {
			s.logger.Errorf("ES搜索院校失败: %v", err)
		} else {
			response.Universities = universities
			response.Total += total
		}
	}

	// 搜索专业
	if len(req.Types) == 0 || s.containsType(req.Types, "major") {
		majors, total, err := s.searchMajorsES(ctx, req)
		if err != nil {
			s.logger.Errorf("ES搜索专业失败: %v", err)
		} else {
			response.Majors = majors
			response.Total += total
		}
	}

	// 计算总页数
	response.TotalPages = int((response.Total + int64(req.PageSize) - 1) / int64(req.PageSize))

	// 获取搜索建议
	suggestions, _ := s.GetSearchSuggestions(ctx, req.Keyword)
	response.Suggestions = suggestions

	// 获取热搜关键词
	hotSearches, _ := s.GetHotSearches(ctx, "", 5)
	if hotSearches != nil {
		for _, hs := range hotSearches.Keywords {
			response.HotKeywords = append(response.HotKeywords, hs.Keyword)
		}
	}

	return response, nil
}

// searchWithDatabase 使用数据库搜索
func (s *SearchService) searchWithDatabase(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	response := &SearchResponse{
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	// 搜索院校
	if len(req.Types) == 0 || s.containsType(req.Types, "university") {
		universities, total, err := s.searchUniversitiesDB(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("搜索院校失败: %w", err)
		}
		response.Universities = universities
		response.Total += total
	}

	// 搜索专业
	if len(req.Types) == 0 || s.containsType(req.Types, "major") {
		majors, total, err := s.searchMajorsDB(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("搜索专业失败: %w", err)
		}
		response.Majors = majors
		response.Total += total
	}

	// 计算总页数
	response.TotalPages = int((response.Total + int64(req.PageSize) - 1) / int64(req.PageSize))

	return response, nil
}

// searchUniversitiesES 使用ES搜索院校
func (s *SearchService) searchUniversitiesES(ctx context.Context, req SearchRequest) ([]models.University, int64, error) {
	searchService := s.db.Elasticsearch.Search().Index("universities")

	// 构建查询
	boolQuery := elastic.NewBoolQuery()

	// 主查询
	multiMatchQuery := elastic.NewMultiMatchQuery(req.Keyword, "name^3", "alias^2", "description").
		Type("best_fields").
		Fuzziness("AUTO")
	boolQuery.Must(multiMatchQuery)

	// 过滤条件
	if req.Province != "" {
		boolQuery.Filter(elastic.NewTermQuery("province", req.Province))
	}

	searchService.Query(boolQuery)

	// 排序
	if req.SortBy == "name" {
		searchService.Sort("name.keyword", req.SortOrder == "asc")
	} else {
		searchService.Sort("_score", false).Sort("national_rank", true)
	}

	// 分页
	from := (req.Page - 1) * req.PageSize
	searchService.From(from).Size(req.PageSize)

	// 执行搜索
	searchResult, err := searchService.Do(ctx)
	if err != nil {
		return nil, 0, err
	}

	// 提取ID并从数据库获取完整数据
	var universityIDs []string
	for _, hit := range searchResult.Hits.Hits {
		var source map[string]interface{}
		if err := json.Unmarshal(hit.Source, &source); err == nil {
			if id, ok := source["id"].(string); ok {
				universityIDs = append(universityIDs, id)
			}
		}
	}

	var universities []models.University
	if len(universityIDs) > 0 {
		if err := s.db.PostgreSQL.Where("id IN ?", universityIDs).Find(&universities).Error; err != nil {
			return nil, 0, err
		}
	}

	return universities, searchResult.Hits.TotalHits.Value, nil
}

// searchMajorsES 使用ES搜索专业
func (s *SearchService) searchMajorsES(ctx context.Context, req SearchRequest) ([]models.Major, int64, error) {
	searchService := s.db.Elasticsearch.Search().Index("majors")

	// 构建查询
	boolQuery := elastic.NewBoolQuery()

	// 主查询
	multiMatchQuery := elastic.NewMultiMatchQuery(req.Keyword, "name^3", "description^2", "career_prospects").
		Type("best_fields").
		Fuzziness("AUTO")
	boolQuery.Must(multiMatchQuery)

	// 过滤条件
	if req.Category != "" {
		boolQuery.Filter(elastic.NewTermQuery("category", req.Category))
	}

	searchService.Query(boolQuery)

	// 排序
	if req.SortBy == "name" {
		searchService.Sort("name.keyword", req.SortOrder == "asc")
	} else {
		searchService.Sort("_score", false).Sort("popularity_score", false)
	}

	// 分页
	from := (req.Page - 1) * req.PageSize
	searchService.From(from).Size(req.PageSize)

	// 执行搜索
	searchResult, err := searchService.Do(ctx)
	if err != nil {
		return nil, 0, err
	}

	// 提取ID并从数据库获取完整数据
	var majorIDs []string
	for _, hit := range searchResult.Hits.Hits {
		var source map[string]interface{}
		if err := json.Unmarshal(hit.Source, &source); err == nil {
			if id, ok := source["id"].(string); ok {
				majorIDs = append(majorIDs, id)
			}
		}
	}

	var majors []models.Major
	if len(majorIDs) > 0 {
		if err := s.db.PostgreSQL.Preload("University").Where("id IN ?", majorIDs).Find(&majors).Error; err != nil {
			return nil, 0, err
		}
	}

	return majors, searchResult.Hits.TotalHits.Value, nil
}

// searchUniversitiesDB 使用数据库搜索院校
func (s *SearchService) searchUniversitiesDB(ctx context.Context, req SearchRequest) ([]models.University, int64, error) {
	query := s.db.PostgreSQL.Model(&models.University{})

	// 搜索条件
	keyword := "%" + req.Keyword + "%"
	query = query.Where("LOWER(name) LIKE LOWER(?) OR LOWER(alias) LIKE LOWER(?) OR LOWER(description) LIKE LOWER(?)", keyword, keyword, keyword)

	// 过滤条件
	if req.Province != "" {
		query = query.Where("province = ?", req.Province)
	}

	query = query.Where("is_active = ?", true)

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序
	if req.SortBy == "name" {
		order := "name ASC"
		if req.SortOrder == "desc" {
			order = "name DESC"
		}
		query = query.Order(order)
	} else {
		query = query.Order("national_rank ASC, name ASC")
	}

	// 分页
	offset := (req.Page - 1) * req.PageSize
	query = query.Offset(offset).Limit(req.PageSize)

	var universities []models.University
	if err := query.Find(&universities).Error; err != nil {
		return nil, 0, err
	}

	return universities, total, nil
}

// searchMajorsDB 使用数据库搜索专业
func (s *SearchService) searchMajorsDB(ctx context.Context, req SearchRequest) ([]models.Major, int64, error) {
	query := s.db.PostgreSQL.Model(&models.Major{})

	// 搜索条件
	keyword := "%" + req.Keyword + "%"
	query = query.Where("LOWER(name) LIKE LOWER(?) OR LOWER(description) LIKE LOWER(?) OR LOWER(career_prospects) LIKE LOWER(?)", keyword, keyword, keyword)

	// 过滤条件
	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}

	query = query.Where("is_active = ?", true)

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序
	if req.SortBy == "name" {
		order := "name ASC"
		if req.SortOrder == "desc" {
			order = "name DESC"
		}
		query = query.Order(order)
	} else {
		query = query.Order("popularity_score DESC, name ASC")
	}

	// 分页
	offset := (req.Page - 1) * req.PageSize
	query = query.Offset(offset).Limit(req.PageSize)

	var majors []models.Major
	if err := query.Preload("University").Find(&majors).Error; err != nil {
		return nil, 0, err
	}

	return majors, total, nil
}

// getUniversitySuggestions 获取院校建议
func (s *SearchService) getUniversitySuggestions(ctx context.Context, keyword string, limit int) []AutoCompleteSuggestion {
	var universities []models.University
	s.db.PostgreSQL.Model(&models.University{}).
		Select("id, name, type").
		Where("LOWER(name) LIKE LOWER(?) OR LOWER(alias) LIKE LOWER(?)", "%"+keyword+"%", "%"+keyword+"%").
		Where("is_active = ?", true).
		Order("national_rank ASC").
		Limit(limit).
		Find(&universities)

	var suggestions []AutoCompleteSuggestion
	for _, u := range universities {
		suggestions = append(suggestions, AutoCompleteSuggestion{
			Text:     u.Name,
			Type:     "university",
			Category: u.Type,
			ID:       u.ID.String(),
		})
	}

	return suggestions
}

// getMajorSuggestions 获取专业建议
func (s *SearchService) getMajorSuggestions(ctx context.Context, keyword string, limit int) []AutoCompleteSuggestion {
	var majors []models.Major
	s.db.PostgreSQL.Model(&models.Major{}).
		Select("id, name, category").
		Where("LOWER(name) LIKE LOWER(?)", "%"+keyword+"%").
		Where("is_active = ?", true).
		Order("popularity_score DESC").
		Limit(limit).
		Find(&majors)

	var suggestions []AutoCompleteSuggestion
	for _, m := range majors {
		suggestions = append(suggestions, AutoCompleteSuggestion{
			Text:     m.Name,
			Type:     "major",
			Category: m.Category,
			ID:       m.ID.String(),
		})
	}

	return suggestions
}

// recordSearch 记录搜索
func (s *SearchService) recordSearch(ctx context.Context, keyword, province, category string) {
	// 更新热搜数据
	go func() {
		today := time.Now().Truncate(24 * time.Hour)

		var hotSearch models.HotSearch
		result := s.db.PostgreSQL.Where("keyword = ? AND date = ? AND category = ?", keyword, today, category).First(&hotSearch)

		if result.Error == gorm.ErrRecordNotFound {
			// 创建新记录
			hotSearch = models.HotSearch{
				Keyword:     keyword,
				SearchCount: 1,
				Category:    category,
				Date:        today,
			}
			s.db.PostgreSQL.Create(&hotSearch)
		} else if result.Error == nil {
			// 更新计数
			s.db.PostgreSQL.Model(&hotSearch).Update("search_count", gorm.Expr("search_count + ?", 1))
		}
	}()
}

// calculateTrend 计算趋势
func (s *SearchService) calculateTrend(ctx context.Context, keyword string) string {
	// 简单的趋势计算，可以根据需要优化
	return "stable"
}

// containsType 检查类型是否包含
func (s *SearchService) containsType(types []string, targetType string) bool {
	for _, t := range types {
		if t == targetType {
			return true
		}
	}
	return false
}

// removeDuplicates 去重
func (s *SearchService) removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}
