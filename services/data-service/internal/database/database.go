package database

import (
	"context"
	"embed"
	"fmt"
	"time"

	"github.com/olivere/elastic/v7"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/tvvshow/gokao/services/data-service/internal/config"
	shareddb "github.com/tvvshow/gokao/pkg/database"
)

// embedMigrations 把 migrations/*.sql 嵌入二进制，避免运行时依赖文件系统路径。
//
//go:embed migrations/*.sql
var embedMigrations embed.FS

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

// initPostgreSQL 初始化 PostgreSQL 连接 + 应用版本化迁移。
func (db *DB) initPostgreSQL() error {
	conn, err := shareddb.OpenGorm(db.Config.DatabaseConfig, shareddb.GormOpenOptions{
		Production: db.Config.Environment != "debug",
	})
	if err != nil {
		return err
	}

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

// migrate 应用所有未运行的 goose 迁移到目标 DB。
//
// 与旧版 AutoMigrate + 散乱 ALTER / CREATE INDEX 的差异：
//   - schema 变更走版本化 SQL 文件（migrations/*.sql），可向下回滚；
//   - goose 在 goose_db_version 表中追踪版本号，二进制启动时自动 Up；
//   - 弃用 GORM 在迁移路径的 AutoMigrate 行为——模型仅承担运行时 ORM，结构演进权归迁移文件。
func (db *DB) migrate() error {
	db.Logger.Info("开始数据库迁移（goose）...")

	sqlDB, err := db.PostgreSQL.DB()
	if err != nil {
		return fmt.Errorf("获取 sql.DB: %w", err)
	}

	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	// SetLogger 把 goose 输出接到 logrus，统一日志格式。
	goose.SetLogger(gooseLogger{db.Logger})

	if err := goose.Up(sqlDB, "migrations"); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	db.Logger.Info("数据库迁移完成")
	return nil
}

// gooseLogger 把 goose 内部日志桥接到项目 logrus 实例。
type gooseLogger struct{ l *logrus.Logger }

func (g gooseLogger) Fatalf(format string, v ...interface{}) { g.l.Fatalf(format, v...) }
func (g gooseLogger) Printf(format string, v ...interface{}) { g.l.Infof(format, v...) }

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
