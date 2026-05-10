package database

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/olivere/elastic/v7"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/tvvshow/gokao/services/data-service/internal/config"
	shareddb "github.com/tvvshow/gokao/pkg/database"
)

// DB 数据库连接管理器
type DB struct {
	PostgreSQL    *gorm.DB
	Redis         *redis.Client
	Elasticsearch *elastic.Client
	Config        *config.Config
	Logger        *logrus.Logger
}

// NewDB 创建新的数据库连接管理器
func NewDB(cfg *config.Config, log *logrus.Logger) (*DB, error) {
	db := &DB{
		Config: cfg,
		Logger: log,
	}

	if err := db.initPostgreSQL(); err != nil {
		return nil, fmt.Errorf("初始化PostgreSQL失败: %w", err)
	}

	if err := db.initRedis(); err != nil {
		return nil, fmt.Errorf("初始化Redis失败: %w", err)
	}

	// Elasticsearch 不是必需的，失败也继续
	if err := db.initElasticsearch(); err != nil {
		log.Warnf("初始化Elasticsearch失败: %v", err)
	}

	return db, nil
}

// initPostgreSQL 初始化 PostgreSQL 连接 + 自动迁移。
func (db *DB) initPostgreSQL() error {
	conn, err := shareddb.OpenGorm(db.Config.DatabaseConfig, shareddb.GormOpenOptions{
		Production: db.Config.Environment != "debug",
	})
	if err != nil {
		return err
	}

	// 验证连接（OpenGorm 内部已 ping，但这里再做一次以保留原日志语义）
	sqlDB, err := conn.DB()
	if err != nil {
		return fmt.Errorf("获取SQL DB实例失败: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("PostgreSQL连接测试失败: %w", err)
	}

	db.PostgreSQL = conn
	db.Logger.Info("PostgreSQL连接成功")

	if err := db.migrate(); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	return nil
}

// initRedis 初始化 Redis 连接（共享层应用统一池/超时常量）。
func (db *DB) initRedis() error {
	client, err := shareddb.OpenRedis(db.Config.RedisConfig, 0)
	if err != nil {
		return fmt.Errorf("Redis连接测试失败: %w", err)
	}
	db.Redis = client
	db.Logger.Info("Redis连接成功")
	return nil
}

// initElasticsearch 初始化Elasticsearch连接
func (db *DB) initElasticsearch() error {
	opts := []elastic.ClientOptionFunc{
		elastic.SetURL(db.Config.ElasticsearchURL),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false),
		elastic.SetRetrier(elastic.NewBackoffRetrier(elastic.NewExponentialBackoff(100*time.Millisecond, 2*time.Second))),
	}

	// 如果配置了用户名和密码
	if db.Config.ElasticsearchUsername != "" && db.Config.ElasticsearchPassword != "" {
		opts = append(opts, elastic.SetBasicAuth(db.Config.ElasticsearchUsername, db.Config.ElasticsearchPassword))
	}

	client, err := elastic.NewClient(opts...)
	if err != nil {
		return fmt.Errorf("创建Elasticsearch客户端失败: %w", err)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	info, code, err := client.Ping(db.Config.ElasticsearchURL).Do(ctx)
	if err != nil {
		return fmt.Errorf("Elasticsearch连接测试失败: %w", err)
	}

	if code >= 400 {
		return fmt.Errorf("Elasticsearch返回错误状态码: %d", code)
	}

	db.Elasticsearch = client
	db.Logger.Infof("Elasticsearch连接成功, 版本: %s", info.Version.Number)

	// 初始化索引
	if err := db.initElasticsearchIndices(); err != nil {
		return fmt.Errorf("初始化Elasticsearch索引失败: %w", err)
	}

	return nil
}

// migrate 执行数据库迁移
func (db *DB) migrate() error {
	db.Logger.Info("开始数据库迁移...")

	// 执行一个简单的查询来验证连接
	var result int
	if err := db.PostgreSQL.Raw("SELECT 1").Scan(&result).Error; err != nil {
		return fmt.Errorf("数据库连接验证失败: %w", err)
	}

	// NEW: 确保 pgcrypto 扩展可用以支持 gen_random_uuid()
	if err := db.PostgreSQL.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto;").Error; err != nil {
		db.Logger.Warnf("创建扩展 pgcrypto 失败（可能权限不足）：%v", err)
	}

	// pg_trgm 扩展为搜索路径上的 LOWER(col) LIKE '%xxx%' 提供 GIN 三元组索引支持。
	// 创建失败不阻塞启动，但搜索将退化为全表扫描，必须在日志中明显标记。
	if err := db.PostgreSQL.Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm;").Error; err != nil {
		db.Logger.Warnf("PERFORMANCE_DEGRADED: 创建扩展 pg_trgm 失败，搜索将退化为全表扫描（可能权限不足）：%v", err)
	}

	// 初始化迁移器
	migrator := NewMigrator(db.PostgreSQL)
	if err := migrator.SetupMigrationTable(); err != nil {
		return fmt.Errorf("设置迁移表失败: %w", err)
	}

	// 定义所有需要迁移的模型
	models := migratableModels()

	// 对所有模型执行迁移，确保字段更新
	for _, model := range models {
		// 获取模型对应的表名
		modelType := reflect.TypeOf(model)
		if modelType.Kind() == reflect.Ptr {
			modelType = modelType.Elem()
		}
		typeName := modelType.Name()
		
		db.Logger.Infof("正在迁移模型: %s", typeName)
		if err := db.PostgreSQL.AutoMigrate(model); err != nil {
			// 如果是约束相关错误，记录警告但继续执行
			if strings.Contains(err.Error(), "constraint") || strings.Contains(err.Error(), "does not exist") {
				db.Logger.Warnf("迁移模型 %s 时遇到约束错误，继续执行: %v", typeName, err)
				continue
			}
			return fmt.Errorf("迁移模型 %T 失败: %w", model, err)
		}
	}

	// NEW: 为 UUID 主键设置数据库默认值 gen_random_uuid()，避免插入失败
	uuidDefaultAlters := []string{
		"ALTER TABLE IF EXISTS universities ALTER COLUMN id SET DEFAULT gen_random_uuid();",
		"ALTER TABLE IF EXISTS majors ALTER COLUMN id SET DEFAULT gen_random_uuid();",
		"ALTER TABLE IF EXISTS admission_data ALTER COLUMN id SET DEFAULT gen_random_uuid();",
		"ALTER TABLE IF EXISTS search_indices ALTER COLUMN id SET DEFAULT gen_random_uuid();",
		"ALTER TABLE IF EXISTS analysis_results ALTER COLUMN id SET DEFAULT gen_random_uuid();",
	}
	for _, q := range uuidDefaultAlters {
		if err := db.PostgreSQL.Exec(q).Error; err != nil {
			db.Logger.Warnf("设置UUID默认值失败: %v | SQL: %s", err, q)
		}
	}

	// 手动添加 popularity_score 字段到 majors 表
	db.Logger.Info("检查并添加 popularity_score 字段...")
	if err := db.PostgreSQL.Exec("ALTER TABLE majors ADD COLUMN IF NOT EXISTS popularity_score INTEGER DEFAULT 0;").Error; err != nil {
		db.Logger.Warnf("添加 popularity_score 字段失败: %v", err)
	} else {
		db.Logger.Info("成功添加 popularity_score 字段")
		
		// 更新现有记录的 popularity_score 值
		updateQueries := []string{
			`UPDATE majors SET popularity_score = 95 WHERE name LIKE '%计算机%' OR name LIKE '%软件%' OR name LIKE '%人工智能%';`,
			`UPDATE majors SET popularity_score = 90 WHERE name LIKE '%电子%' OR name LIKE '%通信%' OR name LIKE '%自动化%';`,
			`UPDATE majors SET popularity_score = 85 WHERE name LIKE '%金融%' OR name LIKE '%经济%' OR name LIKE '%管理%';`,
			`UPDATE majors SET popularity_score = 80 WHERE name LIKE '%医学%' OR name LIKE '%临床%' OR name LIKE '%护理%';`,
			`UPDATE majors SET popularity_score = 75 WHERE name LIKE '%机械%' OR name LIKE '%土木%' OR name LIKE '%建筑%';`,
			`UPDATE majors SET popularity_score = 70 WHERE popularity_score = 0;`,
		}
		
		for i, query := range updateQueries {
			if err := db.PostgreSQL.Exec(query).Error; err != nil {
				db.Logger.Warnf("更新专业热度分数失败 (查询 %d): %v", i+1, err)
			} else {
				db.Logger.Infof("完成专业热度分数更新 (查询 %d)", i+1)
			}
		}
	}

	// 创建索引
	if err := db.createIndices(); err != nil {
		return fmt.Errorf("创建索引失败: %w", err)
	}

	db.Logger.Info("数据库迁移完成")
	return nil
}

// createIndices 创建数据库索引
func (db *DB) createIndices() error {
	db.Logger.Info("开始创建数据库索引...")
	
	// 创建复合索引 - 基于查询模式优化
	indices := []string{
		// 院校表核心查询索引
		"CREATE INDEX IF NOT EXISTS idx_universities_province_type ON universities(province, type)",
		"CREATE INDEX IF NOT EXISTS idx_universities_level_nature ON universities(level, nature)", 
		"CREATE INDEX IF NOT EXISTS idx_universities_province_level ON universities(province, level)",
		"CREATE INDEX IF NOT EXISTS idx_universities_type_level ON universities(type, level)",
		"CREATE INDEX IF NOT EXISTS idx_universities_active_recruiting ON universities(is_active, is_recruiting)",
		"CREATE INDEX IF NOT EXISTS idx_universities_rank_score ON universities(national_rank, popularity_score)",
		
		// 专业表查询索引
		"CREATE INDEX IF NOT EXISTS idx_majors_university_category ON majors(university_id, category)",
		"CREATE INDEX IF NOT EXISTS idx_majors_discipline_degree ON majors(discipline, degree_type)",
		"CREATE INDEX IF NOT EXISTS idx_majors_category_active ON majors(category, is_active)",
		"CREATE INDEX IF NOT EXISTS idx_majors_university_active ON majors(university_id, is_active)",
		
		// 录取数据表索引
		"CREATE INDEX IF NOT EXISTS idx_admission_data_year_province ON admission_data(year, province)",
		"CREATE INDEX IF NOT EXISTS idx_admission_data_university_year ON admission_data(university_id, year)",
		"CREATE INDEX IF NOT EXISTS idx_admission_data_score_rank ON admission_data(avg_score, min_rank)",
		"CREATE INDEX IF NOT EXISTS idx_admission_data_year_batch ON admission_data(year, batch_type)",
		
		// 搜索和分析表索引
		"CREATE INDEX IF NOT EXISTS idx_search_indices_type_province ON search_indices(type, province)",
		"CREATE INDEX IF NOT EXISTS idx_analysis_results_user_created ON analysis_results(user_id, created_at)",

		// pg_trgm 表达式索引：覆盖 services 层 LOWER(col) LIKE '%kw%' 模式。
		// 必须按 LOWER(col) 建表达式索引，planner 才会用——否则索引按 col 但 WHERE 按 LOWER(col)。
		// to_tsvector('chinese',...) 的旧方案在 postgres:15-alpine 上根本建不起来（无 chinese 词典）。
		"CREATE INDEX IF NOT EXISTS idx_universities_name_trgm ON universities USING gin (LOWER(name) gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_universities_code_trgm ON universities USING gin (LOWER(code) gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_universities_alias_trgm ON universities USING gin (LOWER(alias) gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_majors_name_trgm ON majors USING gin (LOWER(name) gin_trgm_ops)",
		// hot_searches.keyword 走 SearchService.AutoComplete 的高频路径（每次输入触发），
		// eef5eb7 的 follow-up：keyword 列还在 seq scan。补一条 trgm 表达式索引把它也带进索引扫描。
		"CREATE INDEX IF NOT EXISTS idx_hot_searches_keyword_trgm ON hot_searches USING gin (LOWER(keyword) gin_trgm_ops)",
	}

	createdCount := 0
	for _, indexSQL := range indices {
		if err := db.PostgreSQL.Exec(indexSQL).Error; err != nil {
			db.Logger.Warnf("创建索引失败: %s, 错误: %v", indexSQL, err)
		} else {
			createdCount++
			db.Logger.Debugf("成功创建索引: %s", indexSQL)
		}
	}
	
	db.Logger.Infof("索引创建完成，成功创建 %d/%d 个索引", createdCount, len(indices))
	return nil
}

// initElasticsearchIndices 初始化Elasticsearch索引
func (db *DB) initElasticsearchIndices() error {
	ctx := context.Background()

	// 大学索引映射
	universityMapping := `{
		"mappings": {
			"properties": {
				"id": {"type": "keyword"},
				"code": {"type": "keyword"},
				"name": {
					"type": "text",
					"analyzer": "ik_max_word",
					"search_analyzer": "ik_smart",
					"fields": {
						"keyword": {"type": "keyword"}
					}
				},
				"alias": {
					"type": "text",
					"analyzer": "ik_max_word"
				},
				"type": {"type": "keyword"},
				"level": {"type": "keyword"},
				"nature": {"type": "keyword"},
				"category": {"type": "keyword"},
				"province": {"type": "keyword"},
				"city": {"type": "keyword"},
				"description": {
					"type": "text",
					"analyzer": "ik_max_word"
				},
				"national_rank": {"type": "integer"},
				"overall_score": {"type": "float"},
				"is_active": {"type": "boolean"},
				"is_recruiting": {"type": "boolean"}
			}
		}
	}`

	// 专业索引映射
	majorMapping := `{
		"mappings": {
			"properties": {
				"id": {"type": "keyword"},
				"university_id": {"type": "keyword"},
				"code": {"type": "keyword"},
				"name": {
					"type": "text",
					"analyzer": "ik_max_word",
					"search_analyzer": "ik_smart",
					"fields": {
						"keyword": {"type": "keyword"}
					}
				},
				"category": {"type": "keyword"},
				"discipline": {"type": "keyword"},
				"sub_discipline": {"type": "keyword"},
				"degree_type": {"type": "keyword"},
				"description": {
					"type": "text",
					"analyzer": "ik_max_word"
				},
				"career_prospects": {
					"type": "text",
					"analyzer": "ik_max_word"
				},
				"employment_rate": {"type": "float"},
				"average_salary": {"type": "float"},
				"popularity_score": {"type": "float"},
				"is_recruiting": {"type": "boolean"},
				"is_active": {"type": "boolean"}
			}
		}
	}`

	// 创建索引
	indices := map[string]string{
		"universities": universityMapping,
		"majors":       majorMapping,
	}

	for indexName, mapping := range indices {
		exists, err := db.Elasticsearch.IndexExists(indexName).Do(ctx)
		if err != nil {
			return fmt.Errorf("检查索引 %s 是否存在失败: %w", indexName, err)
		}

		if !exists {
			_, err := db.Elasticsearch.CreateIndex(indexName).BodyString(mapping).Do(ctx)
			if err != nil {
				return fmt.Errorf("创建索引 %s 失败: %w", indexName, err)
			}
			db.Logger.Infof("创建Elasticsearch索引: %s", indexName)
		}
	}

	return nil
}

// Close 关闭所有数据库连接
func (db *DB) Close() error {
	var errs []error

	// 关闭PostgreSQL连接
	if db.PostgreSQL != nil {
		if sqlDB, err := db.PostgreSQL.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				errs = append(errs, fmt.Errorf("关闭PostgreSQL连接失败: %w", err))
			}
		}
	}

	// 关闭Redis连接
	if db.Redis != nil {
		if err := db.Redis.Close(); err != nil {
			errs = append(errs, fmt.Errorf("关闭Redis连接失败: %w", err))
		}
	}

	// Elasticsearch客户端不需要显式关闭

	if len(errs) > 0 {
		return fmt.Errorf("关闭数据库连接时发生错误: %v", errs)
	}

	db.Logger.Info("所有数据库连接已关闭")
	return nil
}

// Health 检查所有数据库连接健康状态
func (db *DB) Health(ctx context.Context) map[string]bool {
	status := make(map[string]bool)

	// 检查PostgreSQL
	if db.PostgreSQL != nil {
		if sqlDB, err := db.PostgreSQL.DB(); err == nil {
			status["postgresql"] = sqlDB.PingContext(ctx) == nil
		} else {
			status["postgresql"] = false
		}
	}

	// 检查Redis
	if db.Redis != nil {
		status["redis"] = db.Redis.Ping(ctx).Err() == nil
	}

	// 检查Elasticsearch
	if db.Elasticsearch != nil {
		_, _, err := db.Elasticsearch.Ping(db.Config.ElasticsearchURL).Do(ctx)
		status["elasticsearch"] = err == nil
	}

	return status
}

// GetConnectionPoolStats 获取数据库连接池统计信息
func (db *DB) GetConnectionPoolStats() map[string]interface{} {
	if db.PostgreSQL == nil {
		return map[string]interface{}{
			"error": "PostgreSQL connection not initialized",
		}
	}

	sqlDB, err := db.PostgreSQL.DB()
	if err != nil {
		return map[string]interface{}{
			"error": fmt.Sprintf("Failed to get SQL DB: %v", err),
		}
	}

	return map[string]interface{}{
		"max_open_conns":      sqlDB.Stats().MaxOpenConnections,
		"open_conns":          sqlDB.Stats().OpenConnections,
		"in_use":              sqlDB.Stats().InUse,
		"idle":                sqlDB.Stats().Idle,
		"wait_count":          sqlDB.Stats().WaitCount,
		"wait_duration":       sqlDB.Stats().WaitDuration.String(),
		"max_idle_closed":     sqlDB.Stats().MaxIdleClosed,
		"max_lifetime_closed": sqlDB.Stats().MaxLifetimeClosed,
		"max_idle_conns":      sqlDB.Stats().MaxIdleClosed,
	}
}
