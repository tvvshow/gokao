package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/tvvshow/gokao/services/data-service/internal/database"
	"github.com/tvvshow/gokao/services/data-service/internal/models"
	"strings"
	"time"

	"github.com/olivere/elastic/v7"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// MajorService 专业服务
type MajorService struct {
	db     *database.DB
	logger *logrus.Logger
}

// NewMajorService 创建专业服务实例
func NewMajorService(db *database.DB, logger *logrus.Logger) *MajorService {
	return &MajorService{
		db:     db,
		logger: logger,
	}
}

// MajorQueryParams 专业查询参数
type MajorQueryParams struct {
	// 基本查询
	ID           string `form:"id"`
	UniversityID string `form:"university_id"`
	Code         string `form:"code"`
	Name         string `form:"name"`
	Keyword      string `form:"keyword"`

	// 分类筛选
	Category      string `form:"category"`
	Discipline    string `form:"discipline"`
	SubDiscipline string `form:"sub_discipline"`
	DegreeType    string `form:"degree_type"`

	// 就业相关筛选
	MinEmploymentRate float64 `form:"min_employment_rate"`
	MaxEmploymentRate float64 `form:"max_employment_rate"`
	MinSalary         float64 `form:"min_salary"`
	MaxSalary         float64 `form:"max_salary"`

	// 热度筛选
	MinPopularity float64 `form:"min_popularity"`
	MaxPopularity float64 `form:"max_popularity"`

	// 状态筛选
	IsActive     *bool `form:"is_active"`
	IsRecruiting *bool `form:"is_recruiting"`

	// 排序选项
	SortBy    string `form:"sort_by"`    // name, popularity, employment_rate, salary
	SortOrder string `form:"sort_order"` // asc, desc

	// 分页参数
	Page     int `form:"page,default=1"`
	PageSize int `form:"page_size,default=20"`

	// 关联数据
	IncludeUniversity    bool `form:"include_university"`
	IncludeAdmissionData bool `form:"include_admission_data"`
}

// MajorListResponse 专业列表响应
type MajorListResponse struct {
	Majors     []models.Major `json:"majors"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// GetMajorByID 根据ID获取专业详情
func (s *MajorService) GetMajorByID(ctx context.Context, id string) (*models.Major, error) {
	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("major:id:%s", id)
	if s.db.Redis != nil && s.db.Config.CacheEnabled {
		cached, err := s.db.Redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var major models.Major
			if err := json.Unmarshal([]byte(cached), &major); err == nil {
				s.logger.Debugf("从缓存获取专业: %s", id)
				return &major, nil
			}
		} else if err != redis.Nil {
			s.logger.Warnf("获取缓存失败: %v", err)
		}
	}

	// 从数据库查询
	var major models.Major
	query := s.db.PostgreSQL.Preload("University").Where("id = ?", id)

	if err := query.First(&major).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("专业不存在")
		}
		return nil, fmt.Errorf("查询专业失败: %w", err)
	}

	// 缓存结果
	if s.db.Redis != nil && s.db.Config.CacheEnabled {
		data, _ := json.Marshal(major)
		s.db.Redis.Set(ctx, cacheKey, data, s.db.Config.CacheDefaultTTL).Err()
	}

	return &major, nil
}

// ListMajors 获取专业列表
func (s *MajorService) ListMajors(ctx context.Context, params MajorQueryParams) (*MajorListResponse, error) {
	// 验证分页参数
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > s.db.Config.MaxPageSize {
		params.PageSize = s.db.Config.DefaultPageSize
	}

	// 生成缓存键
	cacheKey := s.generateMajorCacheKey("majors:list", params)

	// 尝试从缓存获取
	if s.db.Redis != nil && s.db.Config.CacheEnabled {
		cached, err := s.db.Redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var response MajorListResponse
			if err := json.Unmarshal([]byte(cached), &response); err == nil {
				s.logger.Debugf("从缓存获取专业列表")
				return &response, nil
			}
		} else if err != redis.Nil {
			s.logger.Warnf("获取缓存失败: %v", err)
		}
	}

	// 构建查询
	query := s.db.PostgreSQL.Model(&models.Major{})

	// 应用筛选条件
	s.applyMajorFilters(query, params)

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("计算总数失败: %w", err)
	}

	// 应用排序
	s.applyMajorSort(query, params)

	// 应用分页
	offset := (params.Page - 1) * params.PageSize
	query = query.Offset(offset).Limit(params.PageSize)

	// 预加载关联数据
	if params.IncludeUniversity {
		query = query.Preload("University")
	}
	if params.IncludeAdmissionData {
		query = query.Preload("AdmissionData", func(db *gorm.DB) *gorm.DB {
			return db.Order("year DESC").Limit(5)
		})
	}

	// 执行查询
	var majors []models.Major
	if err := query.Find(&majors).Error; err != nil {
		return nil, fmt.Errorf("查询专业列表失败: %w", err)
	}

	// 构建响应
	totalPages := int((total + int64(params.PageSize) - 1) / int64(params.PageSize))
	response := &MajorListResponse{
		Majors:     majors,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}

	// 缓存结果
	if s.db.Redis != nil && s.db.Config.CacheEnabled {
		data, _ := json.Marshal(response)
		s.db.Redis.Set(ctx, cacheKey, data, s.db.Config.CacheDefaultTTL).Err()
	}

	return response, nil
}

// SearchMajors 搜索专业（使用Elasticsearch）
func (s *MajorService) SearchMajors(ctx context.Context, keyword string, params MajorQueryParams) (*MajorListResponse, error) {
	if s.db.Elasticsearch == nil {
		// 如果Elasticsearch不可用，回退到数据库搜索
		return s.searchMajorsDB(ctx, keyword, params)
	}

	// 验证分页参数
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > s.db.Config.MaxPageSize {
		params.PageSize = s.db.Config.DefaultPageSize
	}

	// 构建Elasticsearch查询
	searchService := s.db.Elasticsearch.Search().Index("majors")

	// 构建查询条件
	boolQuery := elastic.NewBoolQuery()

	// 主查询 - 多字段搜索
	if keyword != "" {
		multiMatchQuery := elastic.NewMultiMatchQuery(keyword, "name^3", "description^2", "career_prospects").
			Type("best_fields").
			Fuzziness("AUTO")
		boolQuery.Must(multiMatchQuery)
	}

	// 应用过滤条件
	s.applyMajorElasticsearchFilters(boolQuery, params)

	searchService.Query(boolQuery)

	// 应用排序
	s.applyMajorElasticsearchSort(searchService, params)

	// 应用分页
	from := (params.Page - 1) * params.PageSize
	searchService.From(from).Size(params.PageSize)

	// 执行搜索
	searchResult, err := searchService.Do(ctx)
	if err != nil {
		s.logger.Errorf("Elasticsearch搜索失败: %v", err)
		// 回退到数据库搜索
		return s.searchMajorsDB(ctx, keyword, params)
	}

	// 提取专业ID
	var majorIDs []string
	for _, hit := range searchResult.Hits.Hits {
		var source map[string]interface{}
		if err := json.Unmarshal(hit.Source, &source); err == nil {
			if id, ok := source["id"].(string); ok {
				majorIDs = append(majorIDs, id)
			}
		}
	}

	// 从数据库获取完整数据
	var majors []models.Major
	if len(majorIDs) > 0 {
		query := s.db.PostgreSQL.Where("id IN ?", majorIDs)
		if params.IncludeUniversity {
			query = query.Preload("University")
		}
		if params.IncludeAdmissionData {
			query = query.Preload("AdmissionData", func(db *gorm.DB) *gorm.DB {
				return db.Order("year DESC").Limit(5)
			})
		}

		if err := query.Find(&majors).Error; err != nil {
			return nil, fmt.Errorf("查询专业详情失败: %w", err)
		}
	}

	// 按搜索结果排序
	orderedMajors := s.orderMajorsByIDs(majors, majorIDs)

	totalPages := int((searchResult.Hits.TotalHits.Value + int64(params.PageSize) - 1) / int64(params.PageSize))

	return &MajorListResponse{
		Majors:     orderedMajors,
		Total:      searchResult.Hits.TotalHits.Value,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetMajorCategories 获取专业类别列表
func (s *MajorService) GetMajorCategories(ctx context.Context) ([]string, error) {
	cacheKey := "major:categories"

	// 尝试从缓存获取
	if s.db.Redis != nil && s.db.Config.CacheEnabled {
		cached, err := s.db.Redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var categories []string
			if err := json.Unmarshal([]byte(cached), &categories); err == nil {
				return categories, nil
			}
		}
	}

	var categories []string
	err := s.db.PostgreSQL.Model(&models.Major{}).
		Distinct("category").
		Where("category IS NOT NULL AND category != ''").
		Pluck("category", &categories).Error

	if err != nil {
		return nil, fmt.Errorf("获取专业类别失败: %w", err)
	}

	// 缓存结果
	if s.db.Redis != nil && s.db.Config.CacheEnabled {
		data, _ := json.Marshal(categories)
		s.db.Redis.Set(ctx, cacheKey, data, time.Hour).Err()
	}

	return categories, nil
}

// GetMajorDisciplines 获取学科列表
func (s *MajorService) GetMajorDisciplines(ctx context.Context) ([]string, error) {
	cacheKey := "major:disciplines"

	// 尝试从缓存获取
	if s.db.Redis != nil && s.db.Config.CacheEnabled {
		cached, err := s.db.Redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var disciplines []string
			if err := json.Unmarshal([]byte(cached), &disciplines); err == nil {
				return disciplines, nil
			}
		}
	}

	var disciplines []string
	err := s.db.PostgreSQL.Model(&models.Major{}).
		Distinct("discipline").
		Where("discipline IS NOT NULL AND discipline != ''").
		Pluck("discipline", &disciplines).Error

	if err != nil {
		return nil, fmt.Errorf("获取学科列表失败: %w", err)
	}

	// 缓存结果
	if s.db.Redis != nil && s.db.Config.CacheEnabled {
		data, _ := json.Marshal(disciplines)
		s.db.Redis.Set(ctx, cacheKey, data, time.Hour).Err()
	}

	return disciplines, nil
}

// GetMajorStatistics 获取专业统计信息
func (s *MajorService) GetMajorStatistics(ctx context.Context) (map[string]interface{}, error) {
	cacheKey := "major:statistics"

	// 尝试从缓存获取
	if s.db.Redis != nil && s.db.Config.CacheEnabled {
		cached, err := s.db.Redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var stats map[string]interface{}
			if err := json.Unmarshal([]byte(cached), &stats); err == nil {
				return stats, nil
			}
		}
	}

	stats := make(map[string]interface{})

	// 总数统计
	var total int64
	s.db.PostgreSQL.Model(&models.Major{}).Where("is_active = ?", true).Count(&total)
	stats["total"] = total

	// 按类别统计
	var categoryStats []struct {
		Category string `json:"category"`
		Count    int64  `json:"count"`
	}
	s.db.PostgreSQL.Model(&models.Major{}).
		Select("category, count(*) as count").
		Where("is_active = ?", true).
		Group("category").
		Order("count DESC").
		Limit(20).
		Scan(&categoryStats)
	stats["by_category"] = categoryStats

	// 按学位类型统计
	var degreeStats []struct {
		DegreeType string `json:"degree_type"`
		Count      int64  `json:"count"`
	}
	s.db.PostgreSQL.Model(&models.Major{}).
		Select("degree_type, count(*) as count").
		Where("is_active = ? AND degree_type IS NOT NULL", true).
		Group("degree_type").
		Scan(&degreeStats)
	stats["by_degree_type"] = degreeStats

	// 热门专业（按热度排序）
	var hotMajors []struct {
		Name            string  `json:"name"`
		PopularityScore float64 `json:"popularity_score"`
	}
	s.db.PostgreSQL.Model(&models.Major{}).
		Select("name, popularity_score").
		Where("is_active = ? AND popularity_score > 0", true).
		Order("popularity_score DESC").
		Limit(10).
		Scan(&hotMajors)
	stats["hot_majors"] = hotMajors

	// 缓存结果
	if s.db.Redis != nil && s.db.Config.CacheEnabled {
		data, _ := json.Marshal(stats)
		s.db.Redis.Set(ctx, cacheKey, data, 10*time.Minute).Err()
	}

	return stats, nil
}

// 辅助方法

// applyMajorFilters 应用专业筛选条件
func (s *MajorService) applyMajorFilters(query *gorm.DB, params MajorQueryParams) {
	if params.ID != "" {
		query.Where("id = ?", params.ID)
	}
	if params.UniversityID != "" {
		query.Where("university_id = ?", params.UniversityID)
	}
	if params.Code != "" {
		query.Where("code = ?", params.Code)
	}
	if params.Name != "" {
		query.Where("LOWER(name) LIKE LOWER(?)", "%"+params.Name+"%")
	}
	if params.Keyword != "" {
		keyword := "%" + params.Keyword + "%"
		query.Where("LOWER(name) LIKE LOWER(?) OR LOWER(description) LIKE LOWER(?) OR LOWER(career_prospects) LIKE LOWER(?)", keyword, keyword, keyword)
	}
	if params.Category != "" {
		query.Where("category = ?", params.Category)
	}
	if params.Discipline != "" {
		query.Where("discipline = ?", params.Discipline)
	}
	if params.SubDiscipline != "" {
		query.Where("sub_discipline = ?", params.SubDiscipline)
	}
	if params.DegreeType != "" {
		query.Where("degree_type = ?", params.DegreeType)
	}
	if params.MinEmploymentRate > 0 {
		query.Where("employment_rate >= ?", params.MinEmploymentRate)
	}
	if params.MaxEmploymentRate > 0 {
		query.Where("employment_rate <= ?", params.MaxEmploymentRate)
	}
	if params.MinSalary > 0 {
		query.Where("average_salary >= ?", params.MinSalary)
	}
	if params.MaxSalary > 0 {
		query.Where("average_salary <= ?", params.MaxSalary)
	}
	if params.MinPopularity > 0 {
		query.Where("popularity_score >= ?", params.MinPopularity)
	}
	if params.MaxPopularity > 0 {
		query.Where("popularity_score <= ?", params.MaxPopularity)
	}
	if params.IsActive != nil {
		query.Where("is_active = ?", *params.IsActive)
	}
	if params.IsRecruiting != nil {
		query.Where("is_recruiting = ?", *params.IsRecruiting)
	}
}

// applyMajorSort 应用排序
func (s *MajorService) applyMajorSort(query *gorm.DB, params MajorQueryParams) {
	sortBy := params.SortBy
	if sortBy == "" {
		sortBy = "popularity"
	}

	sortOrder := strings.ToUpper(params.SortOrder)
	if sortOrder != "ASC" {
		sortOrder = "DESC"
	}

	switch sortBy {
	case "name":
		query.Order("name " + (map[string]string{"DESC": "DESC", "ASC": "ASC"}[sortOrder]))
	case "popularity":
		query.Order("popularity_score " + sortOrder + ", name ASC")
	case "employment_rate":
		query.Order("employment_rate " + sortOrder + ", name ASC")
	case "salary":
		query.Order("average_salary " + sortOrder + ", name ASC")
	default:
		query.Order("popularity_score DESC, name ASC")
	}
}

// applyMajorElasticsearchFilters 应用Elasticsearch筛选条件
func (s *MajorService) applyMajorElasticsearchFilters(boolQuery *elastic.BoolQuery, params MajorQueryParams) {
	if params.UniversityID != "" {
		boolQuery.Filter(elastic.NewTermQuery("university_id", params.UniversityID))
	}
	if params.Category != "" {
		boolQuery.Filter(elastic.NewTermQuery("category", params.Category))
	}
	if params.Discipline != "" {
		boolQuery.Filter(elastic.NewTermQuery("discipline", params.Discipline))
	}
	if params.DegreeType != "" {
		boolQuery.Filter(elastic.NewTermQuery("degree_type", params.DegreeType))
	}
	if params.MinEmploymentRate > 0 || params.MaxEmploymentRate > 0 {
		rangeQuery := elastic.NewRangeQuery("employment_rate")
		if params.MinEmploymentRate > 0 {
			rangeQuery.Gte(params.MinEmploymentRate)
		}
		if params.MaxEmploymentRate > 0 {
			rangeQuery.Lte(params.MaxEmploymentRate)
		}
		boolQuery.Filter(rangeQuery)
	}
	if params.IsActive != nil {
		boolQuery.Filter(elastic.NewTermQuery("is_active", *params.IsActive))
	}
	if params.IsRecruiting != nil {
		boolQuery.Filter(elastic.NewTermQuery("is_recruiting", *params.IsRecruiting))
	}
}

// applyMajorElasticsearchSort 应用Elasticsearch排序
func (s *MajorService) applyMajorElasticsearchSort(searchService *elastic.SearchService, params MajorQueryParams) {
	sortBy := params.SortBy
	if sortBy == "" {
		sortBy = "popularity"
	}

	ascending := params.SortOrder == "asc"

	switch sortBy {
	case "name":
		searchService.Sort("name.keyword", ascending)
	case "popularity":
		searchService.Sort("popularity_score", !ascending) // 热度默认降序
	case "employment_rate":
		searchService.Sort("employment_rate", !ascending)
	case "salary":
		searchService.Sort("average_salary", !ascending)
	default:
		searchService.Sort("_score", false).Sort("popularity_score", false)
	}
}

// searchMajorsDB 使用数据库搜索（Elasticsearch不可用时的回退方案）
func (s *MajorService) searchMajorsDB(ctx context.Context, keyword string, params MajorQueryParams) (*MajorListResponse, error) {
	params.Keyword = keyword
	return s.ListMajors(ctx, params)
}

// orderMajorsByIDs 按ID顺序排列专业列表
func (s *MajorService) orderMajorsByIDs(majors []models.Major, ids []string) []models.Major {
	idToMajor := make(map[string]models.Major)
	for _, m := range majors {
		idToMajor[m.ID.String()] = m
	}

	var ordered []models.Major
	for _, id := range ids {
		if m, exists := idToMajor[id]; exists {
			ordered = append(ordered, m)
		}
	}

	return ordered
}

// generateMajorCacheKey 生成缓存键
func (s *MajorService) generateMajorCacheKey(prefix string, params MajorQueryParams) string {
	data, _ := json.Marshal(params)
	return fmt.Sprintf("%s:%x", prefix, data)
}
