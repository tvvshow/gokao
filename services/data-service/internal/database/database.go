package database

import (
	"context"
	"data-service/internal/config"
	"data-service/internal/models"
	"fmt"
	"time"

	"github.com/olivere/elastic/v7"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

	// 初始化PostgreSQL连接
	if err := db.initPostgreSQL(); err != nil {
		return nil, fmt.Errorf("初始化PostgreSQL失败: %w", err)
	}

	// 初始化Redis连接
	if err := db.initRedis(); err != nil {
		return nil, fmt.Errorf("初始化Redis失败: %w", err)
	}

	// 初始化Elasticsearch连接
	if err := db.initElasticsearch(); err != nil {
		log.Warnf("初始化Elasticsearch失败: %v", err)
		// Elasticsearch不是必需的，继续执行
	}

	return db, nil
}

// initPostgreSQL 初始化PostgreSQL连接
func (db *DB) initPostgreSQL() error {
	// 配置GORM日志
	var gormLogger logger.Interface
	if db.Config.Environment == "debug" {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Error)
	}

	// 连接PostgreSQL
	conn, err := gorm.Open(postgres.Open(db.Config.DatabaseURL), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	})
	if err != nil {
		return fmt.Errorf("连接PostgreSQL失败: %w", err)
	}

	// 配置连接池
	sqlDB, err := conn.DB()
	if err != nil {
		return fmt.Errorf("获取SQL DB实例失败: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)                // 最大空闲连接数
	sqlDB.SetMaxOpenConns(100)               // 最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Hour * 24) // 连接最大生存时间

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("PostgreSQL连接测试失败: %w", err)
	}

	db.PostgreSQL = conn
	db.Logger.Info("PostgreSQL连接成功")

	// 自动迁移数据库表
	if err := db.migrate(); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	return nil
}

// initRedis 初始化Redis连接
func (db *DB) initRedis() error {
	opts := &redis.Options{
		Addr:         db.Config.RedisURL,
		Password:     db.Config.RedisPassword,
		DB:           db.Config.RedisDB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	}

	client := redis.NewClient(opts)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := client.Ping(ctx).Err(); err != nil {
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

	// 测试简单模型
	if err := db.PostgreSQL.AutoMigrate(&models.SimpleTest{}); err != nil {
		return fmt.Errorf("迁移SimpleTest模型失败: %w", err)
	}

	// 定义所有需要迁移的模型 - 暂时注释掉复杂的模型
	/*models := []interface{}{
		&models.University{},
		&models.Major{},
		&models.AdmissionData{},
		&models.SearchIndex{},
		&models.AnalysisResult{},
		&models.HotSearch{},
		&models.DataStatistics{},
	}

	// 执行自动迁移
	for _, model := range models {
		if err := db.PostgreSQL.AutoMigrate(model); err != nil {
			return fmt.Errorf("迁移模型 %T 失败: %w", model, err)
		}
	}*/

	// 创建索引
	if err := db.createIndices(); err != nil {
		return fmt.Errorf("创建索引失败: %w", err)
	}

	db.Logger.Info("数据库迁移完成")
	return nil
}

// createIndices 创建数据库索引
func (db *DB) createIndices() error {
	// 暂时跳过复杂索引创建
	db.Logger.Info("跳过索引创建 - 使用简单测试模式")
	return nil

	/*
	// 创建复合索引
	indices := []string{
		"CREATE INDEX IF NOT EXISTS idx_universities_province_type ON universities(province, type)",
		"CREATE INDEX IF NOT EXISTS idx_universities_level_nature ON universities(level, nature)",
		"CREATE INDEX IF NOT EXISTS idx_majors_university_category ON majors(university_id, category)",
		"CREATE INDEX IF NOT EXISTS idx_majors_discipline_degree ON majors(discipline, degree_type)",
		"CREATE INDEX IF NOT EXISTS idx_admission_data_year_province ON admission_data(year, province)",
		"CREATE INDEX IF NOT EXISTS idx_admission_data_university_year ON admission_data(university_id, year)",
		"CREATE INDEX IF NOT EXISTS idx_admission_data_score_rank ON admission_data(avg_score, min_rank)",
		"CREATE INDEX IF NOT EXISTS idx_search_indices_type_province ON search_indices(type, province)",
		"CREATE INDEX IF NOT EXISTS idx_hot_searches_date_category ON hot_searches(date, category)",
		"CREATE INDEX IF NOT EXISTS idx_analysis_results_user_created ON analysis_results(user_id, created_at)",
	}

	for _, indexSQL := range indices {
		if err := db.PostgreSQL.Exec(indexSQL).Error; err != nil {
			db.Logger.Warnf("创建索引失败: %s, 错误: %v", indexSQL, err)
		}
	}

	return nil
	*/
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