package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/tvvshow/gokao/services/data-service/internal/database"
	"github.com/tvvshow/gokao/services/data-service/internal/models"
	"strings"
	"time"

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
	ID      string `form:"id"`
	Code    string `form:"code"`
	Name    string `form:"name"`
	Keyword string `form:"keyword"`

	// 分类筛选
	Type     string `form:"type"`   // undergraduate, graduate, vocational
	Level    string `form:"level"`  // 985, 211, double_first_class, ordinary
	Nature   string `form:"nature"` // public, private, joint_venture
	Category string `form:"category"`

	// 地理位置
	Province string `form:"province"`
	City     string `form:"city"`

	// 排名筛选
	MinRank int `form:"min_rank"`
	MaxRank int `form:"max_rank"`

	// 状态筛选
	IsActive     *bool `form:"is_active"`
	IsRecruiting *bool `form:"is_recruiting"`

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
	// 参数校验和默认值
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		params.PageSize = 20
	}

	// 构建查询
	query := s.db.PostgreSQL.Model(&models.University{})

	// 关键词搜索
	if params.Keyword != "" {
		keyword := "%" + strings.ToLower(params.Keyword) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(code) LIKE ?", keyword, keyword)
	}

	// 名称筛选
	if params.Name != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(params.Name)+"%")
	}

	// 省份筛选
	if params.Province != "" {
		query = query.Where("province = ?", params.Province)
	}

	// 城市筛选
	if params.City != "" {
		query = query.Where("city = ?", params.City)
	}

	// 层次筛选
	if params.Level != "" {
		query = query.Where("level = ?", params.Level)
	}

	// 类型筛选
	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
	}

	// 性质筛选
	if params.Nature != "" {
		query = query.Where("nature = ?", params.Nature)
	}

	// 类别筛选
	if params.Category != "" {
		query = query.Where("category = ?", params.Category)
	}

	// 排名范围筛选
	if params.MinRank > 0 {
		query = query.Where("national_rank >= ?", params.MinRank)
	}
	if params.MaxRank > 0 {
		query = query.Where("national_rank <= ?", params.MaxRank)
	}

	// 状态筛选
	if params.IsActive != nil {
		query = query.Where("is_active = ?", *params.IsActive)
	}
	if params.IsRecruiting != nil {
		query = query.Where("is_recruiting = ?", *params.IsRecruiting)
	}

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("统计院校数量失败: %w", err)
	}

	// 排序（白名单防 SQL 注入）
	allowedSortFields := map[string]string{
		"":             "national_rank",
		"ranking":      "national_rank",
		"national_rank": "national_rank",
		"name":         "name",
		"province":     "province",
		"type":         "type",
		"level":        "level",
		"founded_year": "founded_year",
	}
	sortCol, ok := allowedSortFields[params.SortBy]
	if !ok {
		sortCol = "national_rank"
	}
	sortOrder := "ASC"
	if params.SortOrder == "desc" {
		sortOrder = "DESC"
	}
	query = query.Order(fmt.Sprintf("%s %s", sortCol, sortOrder))

	// 分页
	offset := (params.Page - 1) * params.PageSize
	query = query.Offset(offset).Limit(params.PageSize)

	// 关联数据
	if params.IncludeMajors {
		query = query.Preload("Majors")
	}

	// 执行查询
	var universities []models.University
	if err := query.Find(&universities).Error; err != nil {
		return nil, fmt.Errorf("查询院校列表失败: %w", err)
	}

	// 计算总页数
	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	return &UniversityListResponse{
		Universities: universities,
		Total:        total,
		Page:         params.Page,
		PageSize:     params.PageSize,
		TotalPages:   totalPages,
	}, nil
}

// SearchUniversities 搜索院校
func (s *UniversityService) SearchUniversities(ctx context.Context, keyword string, params UniversityQueryParams) (*UniversityListResponse, error) {
	// 设置关键词到参数
	params.Keyword = keyword

	// 复用 ListUniversities 逻辑
	return s.ListUniversities(ctx, params)
}

// UniversityStatistics 院校统计信息
type UniversityStatistics struct {
	Total         int64            `json:"total"`
	By985         int64            `json:"by_985"`
	By211         int64            `json:"by_211"`
	ByDoubleFirst int64            `json:"by_double_first_class"`
	ByProvince    map[string]int64 `json:"by_province"`
	ByType        map[string]int64 `json:"by_type"`
	ByNature      map[string]int64 `json:"by_nature"`
}

// GetUniversityStatistics 获取院校统计信息（4 条聚合查询替代原来 7 条独立查询）
func (s *UniversityService) GetUniversityStatistics(ctx context.Context) (*UniversityStatistics, error) {
	stats := &UniversityStatistics{
		ByProvince: make(map[string]int64),
		ByType:     make(map[string]int64),
		ByNature:   make(map[string]int64),
	}

	db := s.db.PostgreSQL.Model(&models.University{})

	// 1. 单条聚合查询获取总数 + 各级别数量（CASE WHEN 兼容 SQLite + PostgreSQL）
	var counts struct {
		Total         int64
		By985         int64
		By211         int64
		ByDoubleFirst int64
	}
	if err := s.db.PostgreSQL.Raw(
		"SELECT COUNT(*) as Total, SUM(CASE WHEN level = '985' THEN 1 ELSE 0 END) as By985, SUM(CASE WHEN level = '211' THEN 1 ELSE 0 END) as By211, SUM(CASE WHEN level = 'double_first_class' THEN 1 ELSE 0 END) as ByDoubleFirst FROM universities WHERE deleted_at IS NULL",
	).Scan(&counts).Error; err != nil {
		return nil, fmt.Errorf("统计院校数量失败: %w", err)
	}
	stats.Total = counts.Total
	stats.By985 = counts.By985
	stats.By211 = counts.By211
	stats.ByDoubleFirst = counts.ByDoubleFirst

	// 2-4. 三条 GROUP BY 查询（不同维度无法合并）
	type groupCount struct {
		Key   string
		Count int64
	}

	var provinceCounts []groupCount
	if err := db.Select("province as key, COUNT(*) as count").Group("province").Scan(&provinceCounts).Error; err != nil {
		return nil, fmt.Errorf("按省份统计失败: %w", err)
	}
	for _, gc := range provinceCounts {
		stats.ByProvince[gc.Key] = gc.Count
	}

	var typeCounts []groupCount
	if err := db.Select("type as key, COUNT(*) as count").Group("type").Scan(&typeCounts).Error; err != nil {
		return nil, fmt.Errorf("按类型统计失败: %w", err)
	}
	for _, gc := range typeCounts {
		stats.ByType[gc.Key] = gc.Count
	}

	var natureCounts []groupCount
	if err := db.Select("nature as key, COUNT(*) as count").Group("nature").Scan(&natureCounts).Error; err != nil {
		return nil, fmt.Errorf("按性质统计失败: %w", err)
	}
	for _, gc := range natureCounts {
		stats.ByNature[gc.Key] = gc.Count
	}

	return stats, nil
}
