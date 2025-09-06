# 统一数据处理管道设计方案

## 1. 背景与目标

根据项目审计报告和对现有代码的分析，当前数据处理存在以下问题：
1. 采用"脚本驱动开发"模式，数据处理逻辑分散在多个独立脚本中
2. 数据流不统一，从爬虫到数据库的过程缺乏标准化
3. 缺乏统一的数据处理服务，维护困难且容易出错

本方案旨在设计一个统一的数据处理管道，解决上述问题，实现：
1. 废除"脚本驱动开发"模式
2. 将所有数据处理逻辑整合到Go编写的data-service中
3. 统一数据流从爬虫->JSON->数据库的过程
4. 选择合适的数据库迁移工具并集成到data-service中

## 2. 设计原则

1. **服务化**: 所有数据处理逻辑集中到data-service中，通过API提供服务
2. **标准化**: 统一数据处理流程，定义清晰的数据流和接口规范
3. **可扩展性**: 设计模块化架构，便于未来扩展新的数据源和处理逻辑
4. **可观测性**: 提供完整的日志记录和监控指标
5. **可靠性**: 实现错误处理、重试机制和数据一致性保证

## 3. 架构设计

### 3.1 整体架构

```
+----------------+    +------------------+    +-----------------+    +------------------+
|   Data Source  | -> |  Data Collector  | -> |  Data Processor | -> |  Data Storage    |
+----------------+    +------------------+    +-----------------+    +------------------+
                            |                        |                     |
                            v                        v                     v
                     +------------------+    +-----------------+    +------------------+
                     |  Crawler Scripts |    |  Data Service   |    |  PostgreSQL/ES   |
                     +------------------+    +-----------------+    +------------------+
```

### 3.2 核心组件

#### 3.2.1 Data Collector (数据收集器)
- 负责从各种数据源收集数据
- 支持多种数据源：网页爬虫、API接口、文件导入等
- 输出标准化的JSON格式数据

#### 3.2.2 Data Processor (数据处理器)
- 核心组件，集成在data-service中
- 负责数据清洗、转换、验证和入库
- 提供RESTful API接口供外部调用
- 支持批量处理和实时处理两种模式

#### 3.2.3 Data Storage (数据存储)
- 主存储：PostgreSQL数据库
- 索引存储：Elasticsearch（用于搜索功能）
- 缓存：Redis（用于热点数据缓存）

## 4. 数据流设计

### 4.1 标准化数据格式

定义统一的JSON数据格式：

```json
{
  "metadata": {
    "source": "string",
    "timestamp": "ISO8601",
    "version": "string",
    "record_count": "integer"
  },
  "universities": [
    {
      "id": "string",
      "name": "string",
      "code": "string",
      "province": "string",
      "city": "string",
      "level": "string",
      "type": "string",
      "founded_year": "integer",
      "website": "string"
    }
  ],
  "majors": [
    {
      "code": "string",
      "name": "string",
      "category": "string",
      "sub_category": "string",
      "education_level": "string",
      "description": "string"
    }
  ],
  "admissions": [
    {
      "year": "integer",
      "province": "string",
      "university_id": "string",
      "major_code": "string",
      "batch": "string",
      "min_score": "integer",
      "max_score": "integer",
      "avg_score": "integer",
      "admit_count": "integer",
      "subject_type": "string"
    }
  ]
}
```

### 4.2 处理流程

1. **数据收集**: 爬虫脚本或API调用收集原始数据，转换为标准化JSON格式
2. **数据传输**: 通过HTTP API或文件传输将JSON数据发送到data-service
3. **数据验证**: Data Processor验证数据格式和完整性
4. **数据转换**: 将数据转换为数据库模型格式
5. **数据入库**: 使用事务确保数据一致性，写入PostgreSQL数据库
6. **索引更新**: 更新Elasticsearch索引以支持搜索功能
7. **缓存更新**: 更新Redis缓存以提高查询性能

## 5. API设计

### 5.1 数据处理API

```
POST /api/v1/data/process
Content-Type: application/json

{
  "data_type": "universities|majors|admissions",
  "data": [...],
  "options": {
    "validate_only": false,
    "update_existing": true,
    "batch_size": 1000
  }
}
```

响应:
```json
{
  "status": "success|error",
  "message": "string",
  "processed_count": "integer",
  "error_count": "integer",
  "errors": []
}
```

### 5.2 数据导入API

```
POST /api/v1/data/import
Content-Type: multipart/form-data

file: JSON文件
type: universities|majors|admissions
```

## 6. 数据库迁移方案

### 6.1 选择的工具

选择 [GORM Migrator](https://gorm.io/docs/migration.html) 作为数据库迁移工具，原因：
1. 与项目中已使用的GORM ORM无缝集成
2. 支持自动迁移和手动迁移
3. 提供版本控制和回滚功能
4. Go语言原生支持，易于集成

### 6.2 迁移策略

1. **自动迁移**: 对于简单的表结构变更，使用GORM的AutoMigrate功能
2. **手动迁移**: 对于复杂的表结构变更，编写专门的迁移脚本
3. **版本控制**: 使用数据库表记录迁移版本，确保迁移顺序正确
4. **回滚支持**: 为每个迁移提供回滚操作

### 6.3 迁移接口设计

在data-service中提供迁移API：

```
POST /api/v1/migrate/up
POST /api/v1/migrate/down
GET /api/v1/migrate/status
```

## 7. 错误处理与监控

### 7.1 错误处理

1. **数据验证错误**: 记录错误详情，支持部分成功处理
2. **数据库错误**: 实现重试机制，区分临时性和永久性错误
3. **网络错误**: 实现超时和重试机制

### 7.2 监控指标

1. **处理速率**: 每秒处理的记录数
2. **错误率**: 处理失败的记录占比
3. **延迟**: 从数据收集到入库的总时间
4. **资源使用**: CPU、内存、数据库连接数等

## 8. 安全考虑

1. **API认证**: 数据处理API需要认证授权
2. **数据加密**: 敏感数据在传输和存储时加密
3. **访问控制**: 限制对数据处理接口的访问权限
4. **审计日志**: 记录所有数据处理操作

## 9. 实施计划

1. **第一阶段**: 设计并实现数据处理核心逻辑
2. **第二阶段**: 集成数据库迁移工具
3. **第三阶段**: 重构现有脚本，迁移到新的数据处理管道
4. **第四阶段**: 完善监控和错误处理机制
5. **第五阶段**: 编写文档和测试用例

## 10. 预期收益

1. **维护性提升**: 数据处理逻辑集中管理，降低维护成本
2. **可靠性增强**: 标准化流程和完善的错误处理提高数据质量
3. **扩展性改善**: 模块化设计便于添加新的数据源和处理逻辑
4. **性能优化**: 批量处理和缓存机制提高处理效率