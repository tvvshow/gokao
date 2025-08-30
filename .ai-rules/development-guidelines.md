# 高考志愿填报助手 - 开发规范和指导原则

## 🎯 项目愿景

基于"高考志愿填报助手"AI编码分阶段指导方案，实现高质量、高性能的混合推荐引擎系统。

## 📋 开发原则

### 核心价值观
1. **零代码重复**: 避免任何形式的代码重复
2. **零冗余**: 精简高效，每行代码都有存在价值  
3. **零占位符**: 不使用TODO、FIXME等占位符，所有代码必须完整实现
4. **高质量标准**: 生产就绪的代码质量
5. **性能优先**: 优化为王，追求极致性能

### 技术选型原则
- **Go**: 微服务、API服务、并发处理
- **C++**: 高性能算法、计算密集型任务
- **混合架构**: 发挥各语言优势，CGO桥接
- **云原生**: Docker容器化，Kubernetes编排

## 🔧 编码规范

### Go语言规范

#### 1. 代码结构
```go
// ✅ 正确的包结构
package handlers

import (
    "context"
    "net/http"
    
    "github.com/gin-gonic/gin"
    "github.com/oktetopython/gaokao/recommendation-service/internal/config"
)

// ✅ 结构体命名和字段
type AnalyticsHandler struct {
    logger *logrus.Logger
    config *config.Config
    bridge cppbridge.HybridRecommendationBridge
}
```

#### 2. 错误处理
```go
// ✅ 完整的错误处理
func (h *AnalyticsHandler) GetStats(c *gin.Context) {
    ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
    defer cancel()

    stats, err := h.bridge.GetHybridStats()
    if err != nil {
        h.logger.Errorf("获取混合引擎统计失败: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "获取统计信息失败",
            "details": err.Error(),
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "status": "success",
        "data":   stats,
    })
}
```

#### 3. 测试规范
```go
// ✅ 完整的单元测试
func TestAnalyticsHandler_GetStats(t *testing.T) {
    tests := []struct {
        name           string
        expectedStatus int
        expectError    bool
    }{
        {
            name:           "成功获取统计数据",
            expectedStatus: http.StatusOK,
            expectError:    false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 测试实现
            assert.Equal(t, tt.expectedStatus, w.Code)
        })
    }
}
```

### C++规范

#### 1. 类设计
```cpp
// ✅ 现代C++设计
class HybridRecommendationEngine {
private:
    std::unique_ptr<TraditionalMatcher> traditional_matcher_;
    std::unique_ptr<AIRecommendationEngine> ai_engine_;
    std::shared_ptr<Config> config_;
    mutable std::shared_mutex cache_mutex_;
    
public:
    explicit HybridRecommendationEngine(const std::string& config_path);
    ~HybridRecommendationEngine() = default;
    
    // 禁用拷贝和移动
    HybridRecommendationEngine(const HybridRecommendationEngine&) = delete;
    HybridRecommendationEngine& operator=(const HybridRecommendationEngine&) = delete;
    
    std::vector<VolunteerRecommendation> GenerateRecommendations(
        const Student& student, int max_volunteers) const;
};
```

#### 2. 内存管理
```cpp
// ✅ RAII和智能指针
class ResourceManager {
private:
    std::unique_ptr<ModelLoader> model_loader_;
    std::vector<std::unique_ptr<FeatureExtractor>> extractors_;
    
public:
    ResourceManager() : model_loader_(std::make_unique<ModelLoader>()) {
        // 资源初始化
    }
    
    // 自动资源清理，无需手动delete
};
```

#### 3. 异常安全
```cpp
// ✅ 异常安全的资源管理
void ProcessBatch(const std::vector<Student>& students) {
    std::lock_guard<std::mutex> lock(process_mutex_);
    
    try {
        auto results = algorithm_engine_->ProcessBatch(students);
        cache_manager_->StoreResults(results);
    } catch (const std::exception& e) {
        logger_->Error("批处理失败: {}", e.what());
        throw; // 重新抛出，让上层处理
    }
}
```

## 🏗️ 架构模式

### 1. 分层架构
```
┌─────────────────┐
│   HTTP API 层   │ ← Gin框架，路由处理
├─────────────────┤
│   业务逻辑层    │ ← Service层，业务规则
├─────────────────┤  
│   C++桥接层     │ ← CGO桥接，数据转换
├─────────────────┤
│   算法引擎层    │ ← C++实现，高性能计算
├─────────────────┤
│   数据存储层    │ ← PostgreSQL, Redis
└─────────────────┘
```

### 2. 依赖注入
```go
// ✅ 依赖注入模式
type ServiceContainer struct {
    Config     *config.Config
    Logger     *logrus.Logger
    Bridge     cppbridge.HybridRecommendationBridge
    Analytics  AnalyticsService
}

func NewServiceContainer(configPath string) (*ServiceContainer, error) {
    cfg, err := config.Load()
    if err != nil {
        return nil, err
    }
    
    bridge, err := cppbridge.NewHybridRecommendationBridge(cfg.CPP.ConfigPath)
    if err != nil {
        return nil, err
    }
    
    return &ServiceContainer{
        Config:    cfg,
        Logger:    logrus.New(),
        Bridge:    bridge,
        Analytics: services.NewAnalyticsService(bridge),
    }, nil
}
```

### 3. 接口设计
```go
// ✅ 接口驱动设计
type RecommendationEngine interface {
    GenerateRecommendations(ctx context.Context, student *Student, options *Options) (*RecommendationPlan, error)
    GetExplanation(ctx context.Context, recommendation *Recommendation) (string, error)
    UpdateWeights(ctx context.Context, weights *FusionWeights) error
}

type CacheManager interface {
    Get(ctx context.Context, key string, result interface{}) error
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
}
```

## 🧪 测试策略

### 1. 测试金字塔
```
    /\
   /E2E\      ← 端到端测试 (少量)
  /______\
 /集成测试 \    ← 服务间集成 (适量)  
/__________\
/  单元测试  \  ← 函数级测试 (大量)
______________
```

### 2. 测试覆盖率目标
- **单元测试**: ≥ 90%
- **集成测试**: ≥ 80%  
- **E2E测试**: 关键路径100%

### 3. 性能测试
```go
// ✅ 基准测试
func BenchmarkRecommendationGeneration(b *testing.B) {
    engine := setupTestEngine()
    student := createTestStudent()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := engine.GenerateRecommendations(context.Background(), student, nil)
        if err != nil {
            b.Fatalf("推荐生成失败: %v", err)
        }
    }
}

// ✅ 并发测试
func TestConcurrentRecommendations(t *testing.T) {
    engine := setupTestEngine()
    concurrency := 100
    requests := 1000
    
    // 并发测试实现
}
```

## 📊 性能标准

### 1. 响应时间要求
- **推荐生成**: < 100ms (P95)
- **分析查询**: < 50ms (P95)
- **配置更新**: < 10ms (P95)

### 2. 吞吐量要求  
- **并发用户**: 1000+
- **QPS**: 100+ (推荐请求)
- **批处理**: 1000+ students/batch

### 3. 资源使用
- **内存**: < 512MB (单实例)
- **CPU**: < 70% (正常负载)
- **缓存命中率**: > 85%

## 🔒 安全规范

### 1. 输入验证
```go
// ✅ 严格的输入验证
func (h *RecommendationHandler) ValidateStudent(student *Student) error {
    if student.StudentID == "" {
        return errors.New("学生ID不能为空")
    }
    
    if student.TotalScore < 0 || student.TotalScore > 750 {
        return errors.New("总分数必须在0-750之间")
    }
    
    if len(student.PreferredCities) > 10 {
        return errors.New("偏好城市数量不能超过10个")
    }
    
    return nil
}
```

### 2. 错误信息处理
```go
// ✅ 安全的错误响应
func handleError(c *gin.Context, err error, userMsg string) {
    // 记录详细错误到日志
    logger.Errorf("内部错误: %v", err)
    
    // 返回安全的用户友好消息
    c.JSON(http.StatusInternalServerError, gin.H{
        "error":   "internal_error",
        "message": userMsg,
        "code":    generateErrorCode(),
    })
}
```

### 3. 敏感数据保护
```go
// ✅ 敏感信息掩码
type Student struct {
    StudentID    string `json:"student_id"`
    Name         string `json:"name"`
    TotalScore   int    `json:"total_score"`
    IDCard       string `json:"-"`          // 不序列化
    Phone        string `json:"-"`          // 不序列化
}

func (s Student) SafeString() string {
    return fmt.Sprintf("Student{ID: %s, Name: %s***}", 
        s.StudentID, s.Name[:1])
}
```

## 📈 监控和日志

### 1. 结构化日志
```go
// ✅ 结构化日志记录
logger.WithFields(logrus.Fields{
    "user_id":      userID,
    "request_id":   requestID,
    "function":     "GenerateRecommendations",
    "duration_ms":  duration.Milliseconds(),
    "success":      true,
}).Info("推荐生成完成")
```

### 2. 指标收集
```go
// ✅ Prometheus指标
var (
    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "recommendation_request_duration_seconds",
            Help: "推荐请求处理时间",
        },
        []string{"method", "status"},
    )
    
    cacheHitRate = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "cache_hits_total",
            Help: "缓存命中次数",
        },
        []string{"cache_type"},
    )
)
```

## 🚀 部署指南

### 1. Docker配置
```dockerfile
# ✅ 多阶段构建
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -o recommendation-service main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/recommendation-service .
COPY --from=builder /app/config ./config
CMD ["./recommendation-service"]
```

### 2. 环境配置
```bash
# ✅ 生产环境变量
export SERVER_PORT=":8083"
export SERVER_MODE="release"
export LOG_LEVEL="info"
export REDIS_ENABLED="true"
export REDIS_HOST="redis-cluster"
export CPP_MAX_WORKERS="8"
export CPP_CACHE_SIZE="5000"
```

## 📝 代码审查清单

### 提交前检查
- [ ] 代码格式化 (`go fmt`, `clang-format`)
- [ ] 静态分析通过 (`go vet`, `cppcheck`)
- [ ] 单元测试通过 (`go test ./... -v`)
- [ ] 性能测试验证 (`go test -bench=.`)
- [ ] 内存泄漏检查 (`valgrind`, `go test -race`)
- [ ] 文档更新完整
- [ ] 提交信息规范

### 代码质量标准
- [ ] 函数长度 < 50行
- [ ] 圈复杂度 < 10
- [ ] 测试覆盖率 > 80%
- [ ] 无明显性能瓶颈
- [ ] 错误处理完整
- [ ] 日志记录合理

## 🔄 持续改进

### 技术债务管理
1. **定期重构**: 每sprint进行代码重构
2. **依赖更新**: 及时更新第三方依赖
3. **性能优化**: 持续性能监控和优化
4. **安全审计**: 定期安全漏洞扫描

### 知识分享
1. **技术文档**: 保持文档更新
2. **代码注释**: 关键逻辑添加注释
3. **架构决策**: 记录重要技术决策
4. **最佳实践**: 分享经验和教训

---

这份开发规范确保了团队在高质量标准下协作开发，任何新加入的开发者都能快速上手并遵循统一的开发标准。