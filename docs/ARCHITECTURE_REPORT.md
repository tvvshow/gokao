高考志愿填报系统 — 完整架构分析报告
======================================
生成日期: 2026-04-24


1. 模块职责与架构拓扑
----------------------

  ┌──────────────────────────────────────────────────────┐
  │                    Frontend (Vue 3)                    │
  │              :80 (nginx) / :3000 (dev)                │
  └────────────────────────┬─────────────────────────────┘
                           │ HTTP /api/{svc}/v1/...
  ┌────────────────────────▼─────────────────────────────┐
  │              API Gateway (:8080)                       │
  │   JWT 透传 · 限流(token bucket) · Redis 缓存           │
  │   Prometheus 指标 · CORS · 反向代理                    │
  └──┬──────────┬──────────┬──────────┬──────────────────┘
     │          │          │          │
     ▼          ▼          ▼          ▼
  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────────────────────┐
  │ User │ │ Data │ │Payment│ │ Recommendation(:8084)│
  │:8083 │ │:8082 │ │:8084 │ │ CGO -> C++ volunteer-  │
  │JWT   │ │PG+ES │ │适配器 │ │ matcher.so            │
  │RBAC  │ │+Redis│ │模式   │ │ + ML (Gorgonia)       │
  └──────┘ └──────┘ └──────┘ └──────────────────────┘
                                ┌──────────────────┐
                                │Monitoring(:8080) │
                                │ 告警+指标(未完成) │
                                └──────────────────┘


2. 各模块职责详述
-----------------

2.1 api-gateway (API 网关)
   - 反向代理：将 /v1/{service} 路由到对应后端服务
   - JWT 透传：从请求头提取 user_id/username/role 并转发
   - 令牌桶限流：每 IP 10 req/s，burst 20
   - Redis 缓存中间件：GET 请求自动缓存
   - Prometheus 指标：请求计数 + 延迟直方图
   - CORS + 安全头 + 输入验证 + Request ID
   - 负载均衡器结构体存在但未与 ProxyManager 集成
   - 826 行单体 main.go，代理/限流/缓存/指标全部耦合

2.2 data-service (数据服务) — 最成熟
   架构: handler -> service -> DB
   存储: PostgreSQL (GORM) + Redis + Elasticsearch
   功能:
   - 大学信息 CRUD + 搜索 (中文分词 ik_max_word)
   - 专业信息 CRUD + 搜索
   - 录取数据查询 + 趋势分析 + 预测
   - 全局搜索 + 自动补全 + 热词
   - 志愿匹配算法 (服务端)
   - 性能监控 + 缓存管理
   - 数据库迁移管理
   中间件: Logger/Recovery/CORS/Security/RateLimit/分页验证/性能监控

2.3 user-service (用户服务)
   - 注册/登录/Token 刷新/登出
   - RBAC 权限模型 (user -> role -> permission)
   - JWT 认证中间件 + 权限中间件 (RequirePermission)
   - 设备指纹 (CGO 链接 OpenSSL + device-fingerprint C++ 模块)
   - 自有 auth 实现，与 pkg/auth 并存
   - CORS 被注释(由 Gateway 处理)

2.4 payment-service (支付服务)
   - 适配器模式：微信支付/支付宝/银联 (adapters/)
   - 支付单创建/查询/关闭/退款/回调
   - 会员系统 (整个被注释)
   - 适配器工厂 (被注释)
   - Repository 模式 (payment_repository.go)
   - Stub 适配器用于未实现的支付渠道

2.5 recommendation-service (推荐服务)
   - CGO -> C++ 混合推荐引擎 (hybrid_bridge.go)
   - ML 引擎 (Gorgonia + tensor)
   - 回退链: EnhancedRule -> SimpleRule -> Mock
   - 缓存: Redis 优先, 内存回退
   - 数据同步服务 (从 data-service 拉数据)
   - 权重配置服务
   - 特征工程模块 (advanced_feature_engineering.go)

2.6 monitoring-service (监控服务) — 骨架
   - Redis 地址硬编码 localhost:6379
   - 端口硬编码 :8080 (与 api-gateway 冲突)
   - 仅 /metrics (空壳) 和 /alerts 两个端点
   - 未被 docker-compose 编排
   - 使用 zap 日志库 (其他服务全用 logrus)

2.7 C++ 模块 (cpp-modules/)
   - device-fingerprint: 设备识别和加密 (OpenSSL)
   - volunteer-matcher: 录取预测算法 (CMake, OpenSSL, JsonCpp)
   - license: 许可证验证
   均独立 CMake 构建，产出 .so/.a

2.8 Frontend (Vue 3 + TypeScript + Vite)
   路由: 11 条 (会员路由被注释，/simulation 和 /recommendation 共享组件)
   Store: user / recommendation / payment (payment 全桩代码)
   API 层: Axios 实例 + Token 自动刷新 + 并发 401 排队
   组件: 顶部导航 + 卡片 + 对比弹窗 + 虚拟列表 + 骨架屏
   样式: Tailwind CSS + Element Plus + 自定义 design-system.css
   测试: Vitest + 属性测试 (fast-check, 8 个场景)


3. 数据流向
-----------

  用户 -> Frontend(nginx:80) -> /api/data/v1/...     -> API Gateway(:8080)
                                  -> /v1/data/...        -> data-service(:8082)

  用户 -> Frontend(nginx:80) -> /api/user/v1/...     -> API Gateway(:8080)
                                  -> /v1/users/...       -> user-service(:8083)

  用户 -> Frontend(nginx:80) -> /api/recommendation/v1/... -> API Gateway(:8080)
                                  -> /v1/recommendations/... -> rec-svc(:8084)

  关键问题:
  - 前端 API 路径: /api/{service}/v1/{resource}
  - Gateway 代理剥离前缀后: /v1/{resource}
  - 各服务注册路径: /api/v1/{resource}
  存在两层前缀不一致，目前靠反向代理的 TrimPrefix 勉强工作。


4. 依赖关系
-----------

go.work (Go 1.25.5) 注册 17 个 module:
  . (root)
  pkg/auth, pkg/database, pkg/errors, pkg/logger,
  pkg/middleware, pkg/models, pkg/scripts, pkg/utils
  services/api-gateway, services/data-service,
  services/monitoring-service,
  services/monitoring-service/internal/alerts,
  services/monitoring-service/internal/metrics,
  services/payment-service, services/recommendation-service,
  services/user-service

  go.work 遗漏 (目录存在但未注册):
  pkg/cache, pkg/discovery, pkg/health, pkg/metrics,
  pkg/response, pkg/shared, pkg/testutil

4.0 模块命名空间碎片化 (关键问题)
  pkg 目录使用 3 种不同的 module path 前缀:
  - github.com/oktetopython/gaokao/pkg/*      (auth, database, errors, logger, middleware, utils)
  - github.com/oktetopython/gaokao/pkg/*  (api-gateway 和 payment-service 的 replace 指令)
  - github.com/gaokao/shared        (shared 包专用)
  - github.com/oktetopython/gaokao/pkg/scripts  (scripts 包专用)
  这意味着每个服务的 go.mod 必须使用 replace 指令来映射 import 路径。
  go.work 在开发时缓解了此问题，但命名不统一是严重技术债务。

4.1 共享包实际使用情况
  17 个 pkg 目录中仅 5 个被服务实际 import:
  已使用: auth, errors, middleware, database, logger
  未使用: cache, discovery, health, metrics, models, api, response,
         scripts, shared, testutil, utils
  (其中 api/response/models 为空或仅有占位文件)

4.2 包编译状态
  - pkg/utils/password_validator.go: import "github.com/gaokao/pkg/errors" (模块路径不存在，应为 gaokaohub)，且缺少 "fmt" import
  - pkg/metrics/business.go:88 包含中文字符 "极"，语法错误
  - pkg/database/pool.go: createPostgresConnection 返回 stub 实现 (未完成)
  - pkg/models/: 仅有 go.mod，无任何 .go 源文件
  - pkg/api/: 空目录
  - pkg/response/: 仅有 test.txt，无代码

4.3 重复实现
  - 错误系统: pkg/errors 中 ErrorResponse 和 APIError 两套并存
  - 缓存系统: pkg/cache 中 CacheManager 和 MultiLevelCache 两套并存
  - 配置: pkg/database 中 connection.go 和 pool.go 有重复的 PoolConfig
  - JWT 认证: pkg/auth 和 pkg/middleware 各有实现

4.4 包可达性
  cache, discovery, metrics, health, testutil 无 go.mod，
  不是独立 Go module，无法被其他 module require。

  外部依赖特征:
  - 所有 Go 服务: Gin + logrus
  - data-service 额外: Elasticsearch v7 (olivere/elastic)
  - recommendation-service 额外: Gorgonia + tensor (ML)
  - user-service 额外: CGO -> OpenSSL + device-fingerprint
  - monitoring-service: zap (与其他服务不同的日志库)


5. 设计模式使用
---------------

  模式              位置                                    评价
  ─────────────────────────────────────────────────────────────────
  API Gateway        api-gateway/main.go:467-678            核心模式，但耦合在单文件
  适配器             payment-service/internal/adapters/     结构正确，未接入真实 SDK
  工厂               payment-service/internal/adapters/     被注释未使用
                   factory.go
  策略 (缓存)        recommendation-service/internal/       Redis/Memory 双实现
                   cache/
  回退链             recommendation-service/main.go:70-87   EnhancedRule->SimpleRule->Mock
  令牌桶限流         api-gateway/main.go:223-273            自实现，无持久化
  Repository         payment-service/internal/repository/   仅 payment-service 用
  Bridge (CGO)       recommendation-service/pkg/            Go<->C++ 桥接，含 mock
                   cppbridge/
  单例 (Config)      各服务 config.Load()                    sync.Once，但每个服务独立实现
  建造者             errors/api_error.go                    WithRequestID/WithRetryAfter 链式


6. 架构问题 (按严重程度)
-------------------------

  严重
  ----
  1. 模块命名空间碎片化
     pkg 目录使用 4 种不同的 module path 前缀 (gaokaohub/oktetopython/gaokao/gaokaohub-gaokao)
     每个服务需要 replace 指令来解析 import，脆弱且混乱

  2. 端口冲突
     docker-compose: payment-service 和 recommendation-service 同为 8084
     monitoring-service (:8080) 与 api-gateway (:8080) 冲突
     代码默认端口: api-gateway=8080, data=10082, user=10081,
                   rec=10083, payment=10084, monitoring=8080

  3. go.work 不完整
     7 个 pkg 子目录未被 workspace 注册，import 这些包在多 module 构建时会失败
     缺失: cache, discovery, health, metrics, response, shared, testutil
     5 个 pkg 无 go.mod (cache/discovery/metrics/health/testutil)，无法被其他 module require

  4. 编译错误
     pkg/utils/password_validator.go: import 不存在的模块路径 + 缺少 fmt import
     pkg/metrics/business.go:88: 中文字符导致的语法错误
     这两个包在 go build ./... 时会编译失败

  5. 前端<->后端路径不一致
     前端: /api/{service}/v1/{resource}
     Gateway 代理: /v1/{resource}
     后端注册: /api/v1/{resource}
     靠 TrimPrefix 勉强工作，脆弱且难调试

  中等
  ----
  4. 共享包利用率低
     pkg/auth/ 和 pkg/errors/ 设计完善，但只有 api-gateway 使用
     user-service 有自有的 auth 实现 (与 pkg/auth 功能重叠)
     data-service 未使用 pkg/errors

  5. api-gateway 单体化
     826 行 main.go，代理/限流/缓存/指标/中间件全部糅合

  6. 服务间通信无弹性
     无熔断器、无重试、无服务发现
     pkg/discovery/consul.go 存在但未使用

  7. 配置分散
     每个服务自有一套 config 结构体和加载逻辑，无统一配置中心

  8. payment-service/Dockerfile EXPOSE 8084 (与 recommendation 冲突)
     monitoring-service Redis 地址硬编码 localhost:6379
     monitoring-service 端口硬编码 :8080

  轻微
  ----
  9. 数据库迁移嵌在 database.NewDB() 启动路径中
     包含业务数据 UPDATE (专业热度分数硬编码)，不属于迁移层

  10. CORS 实现重复 3 处
      api-gateway + data-service + recommendation-service
      应统一由 Gateway 处理

  11. monitoring-service 占位状态
      未被 docker-compose 编排，无实际功能

  12. payment-service 会员模块整个被注释
      支付适配器未接入真实 SDK

  13. recommendation-service/config.go:47-48
      有乱码行 "type极速赛车开奖直播记录" 和重复字段 MaxSize

  14. data-service 热度分数 UPDATE 混在 migrate() 中
      (database.go:253-258)


7. 现存问题清单
---------------

  [严重] 模块命名空间碎片化: 4 种不同 module path 前缀
  [严重] pkg/utils/password_validator.go 编译错误 (import 不存在)
  [严重] pkg/metrics/business.go:88 中文字符语法错误
  [严重] payment/Dockerfile EXPOSE 8084 (与 recommendation 冲突)
  [严重] docker-compose.yml payment + recommendation 同端口 8084
  [严重] monitoring 端口 8080 硬编码 (与 api-gateway 冲突)
  [严重] go.work 缺少 7 个 pkg module
  [严重] 5 个 pkg 无 go.mod，无法被其他 module require
  [中等] 前端路径 vs Gateway 路径 vs 后端路径三层不一致
  [中等] 多个 CORS 实现 (api-gateway/data-service/rec-service)
  [中等] 多个 auth 实现 (pkg/auth + user-service 内建)
  [中等] pkg/errors 中 ErrorResponse 和 APIError 两套错误系统并存
  [中等] pkg/cache 中 CacheManager 和 MultiLevelCache 重复实现
  [中等] pkg/database 中 Postgres 连接为 stub (pool.go)
  [中等] config.go:47 乱码 "type极速赛车开奖直播记录"
  [中等] config.go:47-48 MaxSize 字段重复定义
  [轻微] pkg/api/, pkg/response/, pkg/models/ 为空或仅有占位文件
  [轻微] database.go:253-258 热度 UPDATE 混在迁移中
  [轻微] monitoring Redis 地址硬编码 localhost:6379
  [轻微] monitoring 未被 docker-compose 编排
  [轻微] payment 会员/适配器被注释，未实现
  [轻微] frontend services/ 和 composables/ 目录为空占位
  [轻微] frontend /simulation 和 /recommendation 共享同一组件


8. 优化方向
-----------

  短期 (R2 修复)
  - 统一模块命名空间: 将 pkg 下所有 module path 收敛到单一前缀 (如 github.com/oktetopython/gaokao/pkg/...)
  - 修复编译错误: pkg/utils 的 import + pkg/metrics 的中文字符
  - 删除空包或填充: api, response, models 要么删除要么给出实际代码
  - 为未注册 pkg 补齐 go.mod 或从 go.work 移除
  - 统一端口分配，修正 docker-compose 和代码默认值
  - 将 CORS 集中到 api-gateway，下游服务移除 CORS 中间件
  - 修复 config.go 乱码和重复字段
  - 统一前后端路径约定
  - 消除 pkg/errors 中双错误系统，保留 ErrorResponse/ErrorCode

  中期 (架构增强)
  - 拆分 api-gateway main.go: proxy/router/limiter/cache 各自独立包
  - 引入统一错误处理: 所有服务使用 pkg/errors
  - 统一配置管理: 提取公共 Config 基类型
  - 消除重复: 合并 CacheManager/MultiLevelCache，移除冗余 PoolConfig
  - 激活 pkg/discovery 或移除，不做半成品
  - 服务间调用加超时 + 重试 + 熔断
  - 完成 pkg/database Postgres 连接实现 (当前为 stub)

  长期 (平台化)
  - 引入消息队列 (异步推荐生成、大数据导入)
  - 分布式追踪 (OpenTelemetry)
  - 数据库迁移从启动路径剥离为独立 migrate 子命令
  - monitoring-service: 补全或移除
  - payment-service: 接真实支付 SDK 或移除会员模块


9. 前端架构要点 (补充)
----------------------

  亮点
  - Token 刷新机制: 401 自动 refresh + 并发请求排队
  - 中英文双重类型系统 (中文标签 <-> 英文枚举)
  - 三层错误处理: Axios 拦截器 + ErrorBoundary + Vue errorHandler
  - 属性测试覆盖 8 个场景 (fast-check)
  - 虚拟列表 (VirtualList.vue) + 骨架屏 (SkeletonCard/SkeletonList)

  问题
  - services/ 和 composables/ 目录为空占位
  - payment store 全为 TODO 桩代码
  - /recommendation 和 /simulation 共享组件 (SEO/语义不理想)
  - /login 和 /register 共享组件
  - Vite 代理指向 localhost:10080，与 docker-compose 8080 不一致


10. 推进建议
------------

  1. 立即: 解决端口冲突 + go.work 完整性问题 (影响构建正确性)
  2. 本周: 统一 CORS/错误处理/配置模式，消除跨服务不一致
  3. 本月: api-gateway 拆分解耦，monitoring-service 决定去留
  4. 下月: 消息队列 + 分布式追踪 + C++ 模块 CI 集成 (Option A 已落地)
