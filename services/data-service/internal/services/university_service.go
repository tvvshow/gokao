package services

import (
	"context"
	"data-service/internal/database"
	"data-service/internal/models"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/olivere/elastic/v7"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// UniversityService 院校服务
type UniversityService struct {
	db     *database.DB
	logger *logrus.Logger
}

// NewUniversityService 创建院校服务实例
func NewUniversityService(db *database.DB, logger *logrus.Logger) *UniversityService {
	return &UniversityService{
		db:     db,
		logger: logger,
	}
}

// UniversityQueryParams 院校查询参数
type UniversityQueryParams struct {
	// 基本查询
	ID       string `form:"id"`
	Code     string `form:"code"`
	Name     string `form:"name"`
	Keyword  string `form:"keyword"`
	
	// 分类筛选
	Type     string `form:"type"`     // undergraduate, graduate, vocational
	Level    string `form:"level"`    // 985, 211, double_first_class, ordinary
	Nature   string `form:"nature"`   // public, private, joint_venture
	Category string `form:"category"`
	
	// 地理位置
	Province string `form:"province"`
	City     string `form:"city"`
	
	// 排名筛选
	MinRank int `form:"min_rank"`
	MaxRank int `form:"max_rank"`
	
	// 状态筛选
	IsActive      *bool `form:"is_active"`
	IsRecruiting  *bool `form:"is_recruiting"`
	
	// 排序选项
	SortBy    string `form:"sort_by"`    // name, rank, score, created_at
	SortOrder string `form:"sort_order"` // asc, desc
	
	// 分页参数
	Page     int `form:"page,default=1"`
	PageSize int `form:"page_size,default=20"`
	
	// 关联数据
	IncludeMajors bool `form:"include_majors"`
}

// UniversityListResponse 院校列表响应
type UniversityListResponse struct {
	Universities []models.University `json:"universities"`
	Total        int64               `json:"total"`
	Page         int                 `json:"page"`
	PageSize     int                 `json:"page_size"`
	TotalPages   int                 `json:"total_pages"`
}

// GetUniversityByID 根据ID获取院校详情
func (s *UniversityService) GetUniversityByID(ctx context.Context, id string) (*models.University, error) {
	startTime := time.Now()
	
	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("university:id:%s", id)
	if s.db.Redis != nil && s.db.Config.CacheEnabled {
		cached, err := s.db.Redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var university models.University
			if err := json.Unmarshal([]byte(cached), &university); err == nil {
				s.logger.Debugf("从缓存获取院校: %s", id)
				
				// 记录缓存命中性能
				s.logQueryPerformance("GetUniversityByID", "cache_hit", time.Since(startTime), 1)
				return &university, nil
			}
		} else if err != redis.Nil {
			s.logger.Warnf("获取缓存失败: %v", err)
		}
	}

	// 从数据库查询
	var university models.University
	query := s.db.PostgreSQL.Where("id = ?", id)
	
	if err := query.First(&university).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("院校不存在")
		}
		return nil, fmt.Errorf("查询院校失败: %w", err)
	}

	// 缓存结果
	if s.db.Redis != nil && s.db.Config.CacheEnabled {
		data, _ := json.Marshal(university)
		s.db.Redis.Set(ctx, cacheKey, data, s.db.Config.CacheDefaultTTL).Err()
	}

	// 记录数据库查询性能
	s.logQueryPerformance("GetUniversityByID", "database", time.Since(startTime), 1)

	return &university, nil
}

// GetUniversitiesWithMajors 批量获取院校及其专业信息，避免N+1查询
func (s *UniversityService) GetUniversitiesWithMajors(ctx context.Context, universityIDs []string) ([]models.University, error) {
	if len(universityIDs) == 0 {
		return []models.University{}, nil
	}

	// 生成缓存键
	cacheKey := fmt.Sprintf("universities:with_majors:%s", strings.Join(universityIDs, ","))
	
	// 尝试从缓存获取
	if s.db.Redis != nil && s.db.Config.CacheEnabled {
		cached, err := s.db.Redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var universities []models.University
			if err := json.Unmarshal([]byte(cached), &universities); err == nil {
				s.logger.Debugf("从缓存获取院校及专业信息: %d 所院校", len(universityIDs))
				return universities, nil
			}
		} else if err != redis.Nil {
			s.logger.Warnf("获取缓存失败: %v", err)
		}
	}

	// 使用预加载一次性获取院校及其专业信息
	var universities []models.University
	query := s.db.PostgreSQL.
		Where("id IN ?", universityIDs).
		Preload("Majors", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_active = ?", true).Order("popularity_score DESC, name ASC")
		}).
		Preload("Majors.AdmissionData", func(db *gorm.DB) *gorm.DB {
			return db.Where("year = ?", time.Now().Year()-1). // 获取去年的录取数据
				Order("province ASC, batch ASC")
		})

	if err := query.Find(&universities).Error; err != nil {
		return nil, fmt.Errorf("查询院校及专业信息失败: %w", err)
	}

	// 缓存结果
	if s.db.Redis != nil && s.db.Config.CacheEnabled {
		data, err := json.Marshal(universities)
		if err == nil {
			s.db.Redis.Set(ctx, cacheKey, data, 5*time.Minute).Err()
		}
	}

	s.logger.Debugf("成功获取 %d 所院校的专业信息", len(universities))
	
	// 记录批量查询性能
	startTime := time.Now()
	defer func() {
		s.logQueryPerformance("GetUniversitiesWithMajors", "database", time.Since(startTime), len(universities))
	}()
	
	return universities, nil
}

// logQueryPerformance 记录查询性能指标
func (s *UniversityService) logQueryPerformance(method, source string, duration time.Duration, count int) {
	if s.logger == nil {
		return
	}

	// 记录性能指标
	fields := logrus.Fields{
		"component":   "database",
		"method":      method,
		"source":      source,
		"duration_ms": duration.Milliseconds(),
		"count":       count,
		"throughput":  float64(count) / duration.Seconds(),
	}

	// 根据性能分级记录
	if duration > 100*time.Millisecond {
		s.logger.WithFields(fields).Warn("Slow database query detected")
	} else if duration > 50*time.Millisecond {
		s.logger.WithFields(fields).Info("Database query performance")
	} else {
		s.logger.WithFields(fields).Debug("Database query executed")
	}

	// 记录到性能统计
	if s.db.Redis != nil {
		ctx := context.Background()
		statsKey := fmt.Sprintf("stats:query:%s:%s", method, source)
		s.db.Redis.HIncrBy(ctx, statsKey, "count", int64(count)).Err()
		s.db.Redis.HIncrBy(ctx, statsKey, "total_duration_ms", duration.Milliseconds()).Err()
		s.db.Redis.Expire(ctx, statsKey, 24*time.Hour).Err()
	}
}

// GetUniversityByCode 根据代码获取院校详情
func (s *UniversityService) GetUniversityByCode(ctx context.Context, code string) (*models.University, error) {
	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("university:code:%s", code)
	if s.db.Redis != nil && s.db.Config.CacheEnabled {
		cached, err := s.db.Redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var university models.University
			if err := json.Unmarshal([]byte(cached), &university); err == nil {
				s.logger.Debugf("从缓存获取院校: %s", code)
				return &university, nil
			}
		} else if err != redis.Nil {
			s.logger.Warnf("获取缓存失败: %v", err)
		}
	}

	var university models.University
	if err := s.db.PostgreSQL.Where("code = ?", code).First(&university).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("院校不存在")
		}
		return nil, fmt.Errorf("查询院校失败: %w", err)
	}

	// 缓存结果
	if s.db.Redis != nil && s.db.Config.CacheEnabled {
		data, _ := json.Marshal(university)
		s.db.Redis.Set(ctx, cacheKey, data, s.db.Config.CacheDefaultTTL).Err()
	}

	return &university, nil
}

// ListUniversities 获取院校列表
func (s *UniversityService) ListUniversities(ctx context.Context, params UniversityQueryParams) (*UniversityListResponse, error) {
	startTime := time.Now()
	
	// 验证分页参数
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > s.db.Config.MaxPageSize {
		params.PageSize = s.db.Config.DefaultPageSize
	}

	// 生成缓存键
	cacheKey := s.generateCacheKey("universities:list", params)
	
	// 尝试从缓存获取
	if s.db.Redis != nil && s.db.Config.CacheEnabled {
		cached, err := s.db.Redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var response UniversityListResponse
			if err := json.Unmarshal([]byte(cached), &response); err == nil {
				s.logger.Debugf("从缓存获取院校列表")
				
				// 记录缓存命中性能
				s.logQueryPerformance("ListUniversities", "cache_hit", time.Since(startTime), len(response.Universities))
				return &response, nil
			}
		} else if err != redis.Nil {
			s.logger.Warnf("获取缓存失败: %v", err)
		}
	}

	// 构建查询
	query := s.db.PostgreSQL.Model(&models.University{})
	
	// 应用筛选条件
	s.applyUniversityFilters(query, params)
	
	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("计算总数失败: %w", err)
	}

	// 应用排序
	s.applyUniversitySort(query, params)
	
	// 应用分页
	offset := (params.Page - 1) * params.PageSize
	query = query.Offset(offset).Limit(params.PageSize)
	
	// 预加载关联数据
	if params.IncludeMajors {
		query = query.Preload("Majors", "is_active = ?", true)
	}

	// 执行查询
	var universities []models.University
	if err := query.Find(&universities).Error; err != nil {
		return nil, fmt.Errorf("查询院校列表失败: %w", err)
	}

	// 构建响应
	totalPages := int((total + int64(params.PageSize) - 1) / int64(params.PageSize))
	response := &UniversityListResponse{
		Universities: universities,
		Total:        total,
		Page:         params.Page,
		PageSize:     params.PageSize,
		TotalPages:   totalPages,
	}

	// 缓存结果
	if s.db.Redis != nil && s.db.Config.CacheEnabled {
		data, _ := json.Marshal(response)
		s.db.Redis.Set(ctx, cacheKey, data, s.db.Config.CacheDefaultTTL).Err()
	}

	// 记录数据库查询性能
	s.logQueryPerformance("ListUniversities", "database", time.Since(startTime), len(universities))

	return response, nil
}

// SearchUniversities 搜索院校（使用Elasticsearch）
func (s *UniversityService) SearchUniversities(ctx context.Context, keyword string, params UniversityQueryParams) (*UniversityListResponse, error) {
	if s.db.Elasticsearch == nil {
		// 如果Elasticsearch不可用，回退到数据库搜索
		return s.searchUniversitiesDB(ctx, keyword, params)
	}

	// 验证分页参数
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > s.db.Config.MaxPageSize {
		params.PageSize = s.db.Config.DefaultPageSize
	}

	// 构建Elasticsearch查询
	searchService := s.db.Elasticsearch.Search().Index("universities")
	
	// 构建查询条件
	boolQuery := elastic.NewBoolQuery()
	
	// 主查询 - 多字段搜索
	if keyword != "" {
		multiMatchQuery := elastic.NewMultiMatchQuery(keyword, "name^3", "alias^2", "description").
			Type("best_fields").
			Fuzziness("AUTO")
		boolQuery.Must(multiMatchQuery)
	}
	
	// 应用过滤条件
	s.applyElasticsearchFilters(boolQuery, params)
	
	searchService.Query(boolQuery)
	
	// 应用排序
	s.applyElasticsearchSort(searchService, params)
	
	// 应用分页
	from := (params.Page - 1) * params.PageSize
	searchService.From(from).Size(params.PageSize)
	
	// 执行搜索
	searchResult, err := searchService.Do(ctx)
	if err != nil {
		s.logger.Errorf("Elasticsearch搜索失败: %v", err)
		// 回退到数据库搜索
		return s.searchUniversitiesDB(ctx, keyword, params)
	}

	// 提取大学ID
	var universityIDs []string
	for _, hit := range searchResult.Hits.Hits {
		var source map[string]interface{}
		if err := json.Unmarshal(hit.Source, &source); err == nil {
			if id, ok := source["id"].(string); ok {
				universityIDs = append(universityIDs, id)
			}
		}
	}

	// 从数据库获取完整数据
	var universities []models.University
	if len(universityIDs) > 0 {
		query := s.db.PostgreSQL.Where("id IN ?", universityIDs)
		if params.IncludeMajors {
			query = query.Preload("Majors", "is_active = ?", true)
		}
		
		if err := query.Find(&universities).Error; err != nil {
			return nil, fmt.Errorf("查询院校详情失败: %w", err)
		}
	}

	// 按搜索结果排序
	orderedUniversities := s.orderUniversitiesByIDs(universities, universityIDs)
	
	totalPages := int((searchResult.Hits.TotalHits.Value + int64(params.PageSize) - 1) / int64(params.PageSize))
	
	return &UniversityListResponse{
		Universities: orderedUniversities,
		Total:        searchResult.Hits.TotalHits.Value,
		Page:         params.Page,
		PageSize:     params.PageSize,
		TotalPages:   totalPages,
	}, nil
}

// GetUniversityStatistics 获取院校统计信息
func (s *UniversityService) GetUniversityStatistics(ctx context.Context) (map[string]interface{}, error) {
	cacheKey := "university:statistics"
	
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
	s.db.PostgreSQL.Model(&models.University{}).Where("is_active = ?", true).Count(&total)
	stats["total"] = total
	
	// 按类型统计
	var typeStats []struct {
		Type  string `json:"type"`
		Count int64  `json:"count"`
	}
	s.db.PostgreSQL.Model(&models.University{}).
		Select("type, count(*) as count").
		Where("is_active = ?", true).
		Group("type").
		Scan(&typeStats)
	stats["by_type"] = typeStats
	
	// 按层次统计
	var levelStats []struct {
		Level string `json:"level"`
		Count int64  `json:"count"`
	}
	s.db.PostgreSQL.Model(&models.University{}).
		Select("level, count(*) as count").
		Where("is_active = ?", true).
		Group("level").
		Scan(&levelStats)
	stats["by_level"] = levelStats
	
	// 按省份统计
	var provinceStats []struct {
		Province string `json:"province"`
		Count    int64  `json:"count"`
	}
	s.db.PostgreSQL.Model(&models.University{}).
		Select("province, count(*) as count").
		Where("is_active = ?", true).
		Group("province").
		Order("count DESC").
		Limit(20).
		Scan(&provinceStats)
	stats["by_province"] = provinceStats

	// 缓存结果
	if s.db.Redis != nil && s.db.Config.CacheEnabled {
		data, _ := json.Marshal(stats)
		s.db.Redis.Set(ctx, cacheKey, data, 10*time.Minute).Err()
	}

	return stats, nil
}

// 辅助方法

// applyUniversityFilters 应用院校筛选条件
// applyUniversityFilters 应用查询过滤器 - 优化版本，提升查询性能
func (s *UniversityService) applyUniversityFilters(query *gorm.DB, params UniversityQueryParams) {
	// 优先使用精确匹配的索引字段，提升查询效率
	if params.ID != "" {
		query.Where("id = ?", params.ID)
		return // ID查询直接返回，无需其他条件
	}
	if params.Code != "" {
		query.Where("code = ?", params.Code)
		return // Code查询直接返回，无需其他条件
	}
	
	// 地理位置过滤 - 优先使用province索引
	if params.Province != "" {
		query.Where("province = ?", params.Province)
	}
	if params.City != "" {
		query.Where("city = ?", params.City)
	}
	
	// 分类过滤 - 使用复合索引
	if params.Type != "" {
		query.Where("type = ?", params.Type)
	}
	if params.Level != "" {
		query.Where("level = ?", params.Level)
	}
	if params.Nature != "" {
		query.Where("nature = ?", params.Nature)
	}
	if params.Category != "" {
		query.Where("category = ?", params.Category)
	}
	
	// 排名范围查询 - 优化为单个BETWEEN查询
	if params.MinRank > 0 && params.MaxRank > 0 {
		query.Where("national_rank BETWEEN ? AND ?", params.MinRank, params.MaxRank)
	} else if params.MinRank > 0 {
		query.Where("national_rank >= ?", params.MinRank)
	} else if params.MaxRank > 0 {
		query.Where("national_rank <= ? AND national_rank > 0", params.MaxRank)
	}
	
	// 状态过滤
	if params.IsActive != nil {
		query.Where("is_active = ?", *params.IsActive)
	}
	if params.IsRecruiting != nil {
		query.Where("is_recruiting = ?", *params.IsRecruiting)
	}
	
	// 文本搜索 - 优化查询性能
	if params.Name != "" {
		// 优先使用精确匹配，利用索引
		query.Where("name = ?", params.Name)
	}
	if params.Keyword != "" {
		// 使用全文搜索索引，避免ILIKE全表扫描
		query.Where("to_tsvector('chinese', name || ' ' || COALESCE(alias, '') || ' ' || COALESCE(description, '')) @@ plainto_tsquery('chinese', ?)", params.Keyword)
	}
}

// applyUniversitySort 应用排序 - 优化版本，使用更高效的排序策略
func (s *UniversityService) applyUniversitySort(query *gorm.DB, params UniversityQueryParams) {
	sortBy := params.SortBy
	if sortBy == "" {
		sortBy = "national_rank"
	}
	
	sortOrder := strings.ToUpper(params.SortOrder)
	if sortOrder != "DESC" {
		sortOrder = "ASC"
	}
	
	// 优化排序逻辑，减少复杂的CASE WHEN语句
	switch sortBy {
	case "name":
		query.Order("name " + sortOrder)
	case "rank":
		// 使用NULLS LAST/FIRST代替CASE WHEN，性能更好
		if sortOrder == "ASC" {
			query.Order("national_rank ASC NULLS LAST, name ASC")
		} else {
			query.Order("national_rank DESC NULLS LAST, name ASC")
		}
	case "score":
		if sortOrder == "ASC" {
			query.Order("overall_score ASC NULLS LAST, name ASC")
		} else {
			query.Order("overall_score DESC NULLS LAST, name ASC")
		}
	case "created_at":
		query.Order("created_at " + sortOrder + ", name ASC")
	default:
		// 默认按排名排序，0值排名视为未排名，放在最后
		if sortOrder == "ASC" {
			// 使用WHERE子句分离有排名和无排名的记录，避免复杂排序
			query.Order("(national_rank = 0), national_rank ASC, name ASC")
		} else {
			query.Order("(national_rank = 0), national_rank DESC, name ASC")
		}
	}
}

// applyElasticsearchFilters 应用Elasticsearch筛选条件
func (s *UniversityService) applyElasticsearchFilters(boolQuery *elastic.BoolQuery, params UniversityQueryParams) {
	if params.Type != "" {
		boolQuery.Filter(elastic.NewTermQuery("type", params.Type))
	}
	if params.Level != "" {
		boolQuery.Filter(elastic.NewTermQuery("level", params.Level))
	}
	if params.Nature != "" {
		boolQuery.Filter(elastic.NewTermQuery("nature", params.Nature))
	}
	if params.Province != "" {
		boolQuery.Filter(elastic.NewTermQuery("province", params.Province))
	}
	if params.City != "" {
		boolQuery.Filter(elastic.NewTermQuery("city", params.City))
	}
	if params.MinRank > 0 || params.MaxRank > 0 {
		rangeQuery := elastic.NewRangeQuery("national_rank")
		if params.MinRank > 0 {
			rangeQuery.Gte(params.MinRank)
		}
		if params.MaxRank > 0 {
			rangeQuery.Lte(params.MaxRank)
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

// applyElasticsearchSort 应用Elasticsearch排序
func (s *UniversityService) applyElasticsearchSort(searchService *elastic.SearchService, params UniversityQueryParams) {
	sortBy := params.SortBy
	if sortBy == "" {
		sortBy = "rank"
	}
	
	ascending := params.SortOrder != "desc"
	
	switch sortBy {
	case "name":
		searchService.Sort("name.keyword", ascending)
	case "rank":
		searchService.Sort("national_rank", ascending)
	case "score":
		searchService.Sort("overall_score", !ascending) // 分数默认降序
	default:
		searchService.Sort("_score", false).Sort("national_rank", true)
	}
}

// searchUniversitiesDB 使用数据库搜索（Elasticsearch不可用时的回退方案）
func (s *UniversityService) searchUniversitiesDB(ctx context.Context, keyword string, params UniversityQueryParams) (*UniversityListResponse, error) {
	params.Keyword = keyword
	return s.ListUniversities(ctx, params)
}

// orderUniversitiesByIDs 按ID顺序排列院校列表
func (s *UniversityService) orderUniversitiesByIDs(universities []models.University, ids []string) []models.University {
	idToUniversity := make(map[string]models.University)
	for _, u := range universities {
		idToUniversity[u.ID.String()] = u
	}
	
	var ordered []models.University
	for _, id := range ids {
		if u, exists := idToUniversity[id]; exists {
			ordered = append(ordered, u)
		}
	}
	
	return ordered
}

// generateCacheKey 生成缓存键 - 优化版本，避免JSON序列化开销
func (s *UniversityService) generateCacheKey(prefix string, params UniversityQueryParams) string {
	// 使用字符串拼接代替JSON序列化，提升性能
	var keyParts []string
	keyParts = append(keyParts, prefix)
	
	if params.ID != "" {
		keyParts = append(keyParts, "id:"+params.ID)
	}
	if params.Code != "" {
		keyParts = append(keyParts, "code:"+params.Code)
	}
	if params.Name != "" {
		keyParts = append(keyParts, "name:"+params.Name)
	}
	if params.Keyword != "" {
		keyParts = append(keyParts, "keyword:"+params.Keyword)
	}
	if params.Type != "" {
		keyParts = append(keyParts, "type:"+params.Type)
	}
	if params.Level != "" {
		keyParts = append(keyParts, "level:"+params.Level)
	}
	if params.Nature != "" {
		keyParts = append(keyParts, "nature:"+params.Nature)
	}
	if params.Category != "" {
		keyParts = append(keyParts, "category:"+params.Category)
	}
	if params.Province != "" {
		keyParts = append(keyParts, "province:"+params.Province)
	}
	if params.City != "" {
		keyParts = append(keyParts, "city:"+params.City)
	}
	if params.MinRank > 0 {
		keyParts = append(keyParts, fmt.Sprintf("minrank:%d", params.MinRank))
	}
	if params.MaxRank > 0 {
		keyParts = append(keyParts, fmt.Sprintf("maxrank:%d", params.MaxRank))
	}
	if params.IsActive != nil {
		keyParts = append(keyParts, fmt.Sprintf("active:%t", *params.IsActive))
	}
	if params.IsRecruiting != nil {
		keyParts = append(keyParts, fmt.Sprintf("recruiting:%t", *params.IsRecruiting))
	}
	
	keyParts = append(keyParts, fmt.Sprintf("sort:%s:%s", params.SortBy, params.SortOrder))
	keyParts = append(keyParts, fmt.Sprintf("page:%d:%d", params.Page, params.PageSize))
	keyParts = append(keyParts, fmt.Sprintf("majors:%t", params.IncludeMajors))
	
	return strings.Join(keyParts, "|")
}