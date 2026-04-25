# 高考志愿填报系统通用约束指导规范
# (天网协议 v1.0 — Gaokao Constraint Protocol)

**文档ID**: `MRED-T1-GaokaoConstraints`
**项目名称**: `Gaokao`
**版本**: `1.0`
**状态**: `ENFORCEABLE`

---

## Section 1: 核心约束原则

### 1.1 最优秀原则

- **MUST**: 所有实现必须采用行业最佳实践，追求技术卓越
- **MUST**: Go 服务接口响应时间 P99 不超过 200ms（推荐服务除外）
- **MUST**: 代码质量必须符合企业级标准，无技术债务残留
- **MUST NOT**: 接受"够用就行"的解决方案，必须追求最优解
- **MUST NOT**: 在没有 benchmark 数据支撑的情况下自称"性能优化"

### 1.2 有即复用原则

- **MUST**: 编码前必须扫描 `pkg/` 目录下的共享 Go 包，优先复用成熟功能
- **MUST**: 引入外部代码必须声明来源，格式：`// SOURCE: [URL or module path]`
- **MUST**: 前端组件编写前必须检查 `frontend/src/components/` 是否已有同类实现
- **MUST NOT**: 重复实现已存在的核心功能（如 JWT 验证、Redis 缓存封装、错误码定义）
- **MUST NOT**: 在多个微服务中各自实现相同的业务逻辑，应提取至 `pkg/`

### 1.3 不允许简化原则

- **MUST**: 必须完整实现所有功能需求，不能简化核心逻辑
- **MUST**: 必须处理所有边界条件、错误路径与网络超时场景
- **MUST**: API 必须实现完整的参数校验、错误返回与日志记录
- **MUST NOT**: 使用 `// TODO` 注释替代实现
- **MUST NOT**: 跳过复杂推荐算法，用随机排序或硬编码列表替代

### 1.4 不允许逃避原则

- **MUST**: 必须直面所有技术挑战，不能绕过复杂要求
- **MUST**: 微服务间通信失败必须有完整的降级与重试策略
- **MUST**: 支付相关逻辑必须实现幂等性，不得忽略重复请求风险
- **MUST NOT**: 用 `panic` 替代真实错误处理
- **MUST NOT**: 以"以后再优化"为借口跳过当前安全或性能要求

### 1.5 不允许占位、虚假、虚拟实现原则

- **MUST**: 所有 Go 函数必须真实可运行，不得返回硬编码 mock 数据作为生产逻辑
- **MUST**: 所有前端接口调用必须对接真实后端 API，不得长期使用 `mock.ts`
- **MUST**: 所有数据库操作必须真实执行，不得用内存 map 替代 PostgreSQL 查询
- **MUST NOT**: 使用 `return nil, nil` 替代真实错误处理
- **MUST NOT**: 前端组件使用静态假数据渲染，而不调用 API

---

## Section 2: 代码质量约束

### 2.1 Go 后端规范

- **MUST**: 使用 Go 1.25 标准，充分利用泛型、`errors.Join`、`slog` 等现代特性
- **MUST**: 所有错误必须使用 `fmt.Errorf("...: %w", err)` 进行包装，保留调用链
- **MUST**: 使用 `context.Context` 传递请求上下文，支持超时与取消
- **MUST**: 所有对外 HTTP Handler 必须通过 Gin 的 `ShouldBindJSON` 进行参数校验
- **MUST NOT**: 在 Handler 层直接编写业务逻辑，必须分层（Handler → Service → Repository）
- **MUST NOT**: 使用裸 `goroutine` 而不做错误捕获与生命周期管理

### 2.2 前端规范（Vue 3 + TypeScript）

- **MUST**: 所有组件使用 `<script setup lang="ts">` 语法，禁止 Options API
- **MUST**: 所有 API 请求类型必须通过 TypeScript interface 显式定义，禁止 `any`
- **MUST**: 使用 Pinia 管理全局状态，禁止组件间直接共享响应式变量
- **MUST**: 组件 Props 必须通过 `defineProps<T>()` 显式声明类型
- **MUST NOT**: 直接操作 DOM，必须通过 Vue 响应式系统
- **MUST NOT**: 在 `setup()` 中编写超过 100 行的逻辑，应提取为 Composable

### 2.3 C++ 模块规范（cpp-modules/）

- **MUST**: 使用 C++17 标准，利用 `std::optional`、`std::variant`、`std::string_view`
- **MUST**: 所有 CGO 接口函数必须有明确的内存所有权说明注释
- **MUST**: 使用智能指针（`std::unique_ptr`、`std::shared_ptr`）管理动态内存
- **MUST NOT**: 在 CGO 边界传递裸指针而不说明生命周期
- **MUST NOT**: 存在内存泄漏（通过 Valgrind 或 AddressSanitizer 验证）

### 2.4 数据层规范

- **MUST**: PostgreSQL 查询必须使用参数化语句，禁止字符串拼接 SQL
- **MUST**: Redis 缓存必须设置合理的 TTL，禁止永不过期的业务缓存
- **MUST**: Elasticsearch 索引 Mapping 必须在 `docs/` 中有对应文档
- **MUST NOT**: 在事务外执行需要原子性的多步数据库操作
- **MUST NOT**: 直接在业务代码中硬编码数据库连接字符串

---

## Section 3: 微服务架构约束

### 3.1 服务边界

- **MUST**: 严格遵守服务职责划分：
  - `api-gateway`：路由、认证鉴权、限流，不含业务逻辑
  - `data-service`：院校/专业数据查询，集成 Elasticsearch
  - `user-service`：用户注册、登录、档案管理
  - `payment-service`：支付流程，幂等性处理
  - `recommendation-service`：调用 C++ 模块，生成推荐结果
  - `monitoring-service`：指标采集，对接 Prometheus
- **MUST NOT**: 服务间直接操作对方的数据库
- **MUST NOT**: 绕过 `api-gateway` 让前端直接访问内部服务

### 3.2 服务间通信

- **MUST**: 服务间调用必须设置超时（推荐不超过 3s）
- **MUST**: 关键服务调用必须实现熔断器模式（Circuit Breaker）
- **MUST**: 所有内部 API 必须在 `docs/` 中有接口文档
- **MUST NOT**: 使用同步 RPC 调用可异步处理的非关键链路
- **MUST NOT**: 硬编码服务地址，必须通过环境变量或服务发现

### 3.3 API Gateway 规范

- **MUST**: 所有对外接口必须有 Swagger 注释，并通过 CI 校验
- **MUST**: 认证中间件必须统一在 Gateway 层处理，下游服务信任 Gateway 传递的用户信息
- **MUST**: 限流策略必须可配置，不得硬编码阈值
- **MUST NOT**: 在 Gateway 层实现业务逻辑

---

## Section 4: 性能约束

### 4.1 后端性能基准

- **MUST**: `data-service` 院校列表查询接口 P99 < 100ms（有缓存时）
- **MUST**: `recommendation-service` 推荐接口 P99 < 500ms
- **MUST**: `user-service` 登录/注册接口 P99 < 200ms
- **MUST**: 所有接口必须有对应的基准测试（Go `_test.go` Benchmark 函数）
- **MUST NOT**: 在热路径中执行 N+1 查询

### 4.2 前端性能基准

- **MUST**: 首屏 LCP（最大内容绘制）< 2.5s（生产构建）
- **MUST**: 使用 Vite 动态 `import()` 对路由级组件进行懒加载
- **MUST**: 大列表（院校列表、志愿列表）必须使用虚拟滚动
- **MUST NOT**: 在 `main.ts` 中同步引入超过 5 个大型第三方库
- **MUST NOT**: 将未压缩的静态资源直接部署（必须通过 Vite build）

### 4.3 缓存策略

- **MUST**: 院校基础数据（变化频率低）缓存 TTL ≥ 1 小时
- **MUST**: 用户推荐结果缓存 TTL ≤ 30 分钟，支持主动失效
- **MUST**: Redis 缓存 Key 必须遵循命名规范：`{service}:{entity}:{id}`
- **MUST NOT**: 对写多读少的数据滥用缓存

### 4.4 CGO/C++ 推荐模块性能

- **MUST**: CGO 调用必须最小化跨边界频率，批量传递数据
- **MUST**: C++ 推荐核心算法必须有独立的性能测试用例
- **MUST NOT**: 在请求热路径上频繁申请/释放大块内存

---

## Section 5: 安全约束

### 5.1 认证与授权

- **MUST**: JWT Token 必须设置合理过期时间（Access Token ≤ 2h，Refresh Token ≤ 7d）
- **MUST**: 敏感操作（支付、修改密码）必须二次校验身份
- **MUST**: 用户密码必须使用 bcrypt（cost ≥ 12）存储，禁止明文或 MD5
- **MUST NOT**: 在日志中记录密码、Token 或完整的支付卡号
- **MUST NOT**: 在前端 localStorage 中明文存储敏感用户信息

### 5.2 数据安全

- **MUST**: 所有外部输入必须经过校验和转义，防止 SQL 注入与 XSS
- **MUST**: 文件上传必须校验 MIME 类型与文件大小
- **MUST**: 涉及高考成绩、身份证号等敏感字段，数据库存储时必须加密
- **MUST NOT**: API 错误响应中暴露数据库错误详情或内部堆栈
- **MUST NOT**: 允许未认证用户访问任何用户私有数据接口

### 5.3 支付安全

- **MUST**: 支付接口必须实现幂等性（基于业务订单号去重）
- **MUST**: 支付回调必须校验签名，拒绝伪造请求
- **MUST**: 订单金额必须在服务端计算，不信任客户端传入的价格
- **MUST NOT**: 在非 HTTPS 环境下传输支付相关数据

---

## Section 6: 测试约束

### 6.1 测试覆盖率

- **MUST**: Go 单元测试覆盖率 ≥ 80%（`make test-go` 输出）
- **MUST**: 推荐算法核心逻辑覆盖率 ≥ 95%
- **MUST**: 前端关键业务组件（志愿填报流程）必须有 Vitest 单元测试
- **MUST NOT**: 存在未测试的支付、认证等关键路径

### 6.2 测试类型要求

- **MUST**: 每个 Go Service 层函数必须有对应的单元测试（使用 `testify/mock`）
- **MUST**: API 层必须有集成测试，覆盖正常路径与错误路径
- **MUST**: 数据库操作使用 `testcontainers-go` 启动真实 PostgreSQL 进行测试
- **MUST**: C++ 模块必须有 GoogleTest 单元测试
- **MUST NOT**: 使用硬编码期望值替代真实计算结果的测试

### 6.3 端到端测试

- **MUST**: 核心用户旅程（注册 → 查询院校 → 生成推荐 → 填报志愿）必须有 E2E 测试
- **MUST**: E2E 测试必须在 CI 中可重复执行，不依赖外部真实服务
- **MUST NOT**: E2E 测试依赖共享的测试数据库状态

---

## Section 7: 可观测性约束

### 7.1 日志规范

- **MUST**: 使用 Go `slog` 结构化日志，字段统一（`service`、`trace_id`、`user_id`、`duration_ms`）
- **MUST**: 每个请求必须生成唯一 `trace_id` 并在服务间传递（通过 Header）
- **MUST**: 错误日志必须包含足够上下文，但不得包含敏感数据
- **MUST NOT**: 使用 `fmt.Println` 输出业务日志
- **MUST NOT**: 在循环内输出 DEBUG 级别日志而不加采样控制

### 7.2 Prometheus 指标

- **MUST**: 每个微服务必须暴露 `/metrics` 端点
- **MUST**: 至少上报以下指标：请求总数、请求耗时分布（histogram）、错误率
- **MUST**: `recommendation-service` 必须上报推荐算法执行耗时
- **MUST NOT**: 使用高基数标签（如用户 ID）作为 Prometheus Label

### 7.3 健康检查

- **MUST**: 每个服务必须实现 `/health/live` 和 `/health/ready` 端点
- **MUST**: `ready` 端点必须检查数据库、Redis 等依赖的连通性
- **MUST NOT**: 健康检查端点执行耗时超过 500ms

---

## Section 8: 工程规范约束

### 8.1 代码提交规范

- **MUST**: 提交信息必须遵循 Conventional Commits 格式
  - 格式：`<type>(<scope>): <description>`
  - 示例：`feat(recommendation): add score-range filtering`
  - 示例：`fix(api-gateway): correct JWT expiry check`
- **MUST**: 每次提交前执行 `make test` 与前端 lint/type-check
- **MUST NOT**: 提交包含敏感信息（密钥、密码）的代码
- **MUST NOT**: 在主分支直接提交，必须通过 Pull Request

### 8.2 目录结构规范

- **MUST**: 新 Go 微服务必须放在 `services/` 目录下，遵循现有目录结构
- **MUST**: 共享 Go 逻辑必须提取至 `pkg/` 并有完整单元测试
- **MUST**: 新 C++ 模块必须放在 `cpp-modules/` 下，CGO 绑定代码单独文件
- **MUST**: Docker 相关配置必须放在 `docker/` 目录，不得散落在根目录
- **MUST NOT**: 在 `frontend/` 目录外编写前端代码

### 8.3 Docker 与部署规范

- **MUST**: 每个服务必须有独立的多阶段 Dockerfile（builder + runtime）
- **MUST**: 生产镜像必须基于 distroless 或 Alpine，禁止使用 `ubuntu:latest`
- **MUST**: `docker-compose.yml` 中的端口映射必须与 README 保持同步
- **MUST**: 所有环境变量必须在 `docker-compose.yml` 中有注释说明
- **MUST NOT**: 将 `.env` 文件提交至代码仓库

### 8.4 Makefile 规范

- **MUST**: 所有常用操作必须有对应的 Make target（`build`、`test`、`lint`、`docker-up`）
- **MUST**: Make target 必须有简短注释说明用途
- **MUST NOT**: 在 CI 中直接执行未在 Makefile 中定义的复杂命令链

---

## Section 9: 文档约束

### 9.1 代码文档

- **MUST**: 所有导出的 Go 函数/类型必须有 godoc 注释（中英文均可）
- **MUST**: 所有 Gin Handler 必须有 Swagger 注释（`@Summary`、`@Param`、`@Success`、`@Failure`）
- **MUST**: C++ 模块所有公共函数必须有 Doxygen 风格注释
- **MUST**: Vue 组件 Props 复杂类型必须有 JSDoc 注释
- **MUST NOT**: 提交无注释的导出函数

### 9.2 架构文档

- **MUST**: `docs/` 目录必须包含最新的系统架构图（服务拓扑、数据流向）
- **MUST**: 每次新增微服务或修改服务间依赖，必须同步更新架构文档
- **MUST**: API 变更必须同步更新 Swagger 文档，CI 会校验
- **MUST NOT**: 文档与实际实现不符超过 1 个迭代周期

---

## Section 10: CI/CD 约束

### 10.1 CI 流水线必检项

每次 Pull Request 合并前，CI 必须通过以下全部检查：

- **MUST**: `make test`（Go 单元测试 + 前端单元测试）全部通过
- **MUST**: `cd frontend && npm run lint && npm run type-check` 无错误
- **MUST**: Go 代码通过 `golangci-lint`（配置见 `.golangci.yml`）
- **MUST**: Swagger 文档与代码注释一致性检查通过
- **MUST**: Docker 镜像构建成功（`make build`）
- **MUST NOT**: 跳过任何 CI 检查项合并代码

### 10.2 部署约束

- **MUST**: 生产部署前必须通过 Staging 环境验证
- **MUST**: 数据库 Migration 必须向前兼容，支持回滚
- **MUST**: 服务发布采用滚动更新，保证零停机
- **MUST NOT**: 直接修改生产数据库 Schema 而不经过 Migration 工具

---

## Section 11: 违规处理机制

### 11.1 违规分级

- **Critical**（立即停止，24 小时内修复）
  - 安全漏洞（SQL 注入、未鉴权接口暴露用户数据）
  - 支付逻辑缺陷（非幂等、金额客户端可篡改）
  - 内存泄漏导致服务崩溃
  - 敏感数据明文存储或日志泄露

- **High**（停止当前任务，48 小时内修复）
  - 单元测试覆盖率低于 80%
  - CI 流水线失败
  - API 文档与实现不一致
  - 服务间无超时保护

- **Medium**（记录，72 小时内修复）
  - 缺少结构化日志或 trace_id
  - 缓存 TTL 未设置
  - 文档缺失或过时

- **Low**（记录，下次提交前修复）
  - 命名不规范
  - 提交信息格式不符合 Conventional Commits
  - 注释缺失

### 11.2 修复验证

- **MUST**: 所有修复必须通过相同的 CI 验证
- **MUST**: Critical 和 High 级别修复必须经过 Code Review
- **MUST NOT**: 提交未经验证的修复

---

## Section 12: AI 辅助开发约束

> 本节针对使用 AI 编码助手（如 Claude、Cursor、GitHub Copilot 等）的场景

### 12.1 AI 生成代码审查

- **MUST**: AI 生成的所有代码必须经过人工审查，不得直接 commit
- **MUST**: AI 生成的测试用例必须验证其确实在测试正确的逻辑，而非恒真断言
- **MUST**: AI 生成的 SQL 必须经过 EXPLAIN ANALYZE 验证执行计划
- **MUST NOT**: 直接采用 AI 生成的安全相关代码（JWT、加密、签名验证）而不逐行审查

### 12.2 AI 输出质量要求

- **MUST**: 要求 AI 输出完整实现，不接受含 `// TODO` 或 `// implement me` 的代码
- **MUST**: 要求 AI 同时输出对应的单元测试
- **MUST**: 对 AI 生成的架构方案，必须与 `docs/` 中的现有架构文档进行一致性验证
- **MUST NOT**: 将 AI 的推断性描述（"这应该可以工作"）视为验证通过

### 12.3 上下文提供规范

- **MUST**: 向 AI 提问时必须提供当前服务的技术栈版本（Go 1.25、Vue 3、PostgreSQL 版本等）
- **MUST**: 提供现有相关代码片段，避免 AI 重新发明已有实现
- **MUST**: 明确说明约束条件（如"不使用 ORM"、"必须支持并发安全"）
- **MUST NOT**: 接受 AI 提供的与本协议冲突的"最佳实践"建议而不进行评估

---

## 总结

本规范（天网协议 v1.0）基于 Gaokao 高考志愿填报系统的技术栈与业务特性制定，覆盖 Go 微服务、Vue 3 前端、C++ 推荐模块、数据层、安全、可观测性、工程规范、AI 辅助开发等全方位约束。

所有约束通过 CI/CD 自动化强制执行，结合分级违规处理机制，确保项目在高考业务的高可靠性、数据安全性要求下达到企业级质量标准。

**核心目标**：将 Gaokao 系统打造为安全可靠、性能稳定、可维护性强的高考志愿填报服务标杆项目。
