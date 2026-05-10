# 技术债务审计与推进计划

**初版**: 2026-05-10（全量审计）
**复审**: 2026-05-10（status update）
**审计范围**: 6 微服务 + 14 pkg 模块 + C++ 算法模块

---

## 0. 状态总览

| 类别 | 总数 | ✓ FIXED | ⚠ PARTIAL | ✗ PENDING |
|------|------|---------|-----------|-----------|
| 严重 (P-01~P-10) | 10 | **9** | 0 | 1 |
| 中等 (P-11~P-25) | 15 | **8** | 3 | 4 |
| 重复代码 (A~I) | 9 | 3 | 3 | 3 |
| 算法 (Phase 4) | 5 | 0 | 0 | 5 |
| **合计** | **39** | **20** | **6** | **13** |

复审依据：逐文件 grep + 关键文件 read 验证（非仅看 commit message）。

---

## 1. 已完成项（19）

| ID | 项目 | 修复 commit | 验证位置 |
|----|------|-------------|----------|
| P-02 | GetUserPermissions N+1 | c237473 | `services/user-service/internal/services/auth_service.go:236-237`（JOIN） |
| P-03 | RequireAuth 每请求查库 | c237473 | `services/user-service/internal/middleware/permission.go:108-131`（Redis 缓存 5min） |
| P-04 | RequirePermission 空实现 | c237473 | `services/user-service/internal/middleware/permission.go:208-264`（真权限校验） |
| P-06 | c.JSON 双写 | c237473 | `services/payment-service/internal/handlers/payment_handler.go:160-183` |
| P-07 | 订单号纳秒冲突 | c237473 | `services/payment-service/internal/service/payment_service.go:546`（`%d%04d` + rand） |
| P-09 | 事务路径 db=nil panic | c237473 | `services/payment-service/internal/repository/payment_repository.go:46-49` |
| P-10 | callback body 重复读取 | c237473 | `services/payment-service/internal/handlers/payment_handler.go:170-172` |
| P-11 | Prometheus 重复注册 | 编 main.go 改造 | `services/api-gateway/main.go:203-224`（per-router registry） |
| P-12 | 缓存中间件失效 | 674ad84 | `services/api-gateway/main.go:171-200`（bodyWriter 包装） |
| P-15* | 7 条 stats 查询 → 4 | 44c0d71 | `services/data-service/internal/services/university_service.go:415`（合 1 + 3 GROUP BY） |
| P-16 | 冒泡排序 | 732b9e4 | `services/data-service/internal/services/performance_service.go:380` |
| P-18 | ORDER BY SQL 注入 | 732b9e4 | `services/data-service/internal/services/university_service.go:327-337`（白名单） |
| P-20 | 无 HTTP Server 超时 | 732b9e4 | `services/user-service/main.go:210-217` |
| P-21 | membership 缓存形同虚设 | 7043249 | `services/payment-service/internal/services/membership_service.go:131-137,544` |
| B | Config 辅助函数重复 | 80094dc | 各服务 `getEnv*` 已委托 `pkg/config` |
| C | CORS 中间件 5 处重复 | 1f12cf3 | 已统一到 `pkg/middleware` |
| D | RequestID/TraceID 4 处重复 | 1f12cf3 | 已统一到 `pkg/middleware` |
| (G 部分) | JWTClaims 重复 | 历史 commit | 仅 `services/user-service/internal/models/models.go:62` 一处定义 |
| (CI) | Node 20 deprecation | 5358b99/8989b9d | annotations 13 → 1 |

\* 标 \* 表示部分完成，详见下文 PARTIAL 列表。

---

## 2. 未完成路线图（20 项 + 5 算法）

按风险/优先级分四个阶段。每项包含：现状 → 期望 → 步骤 → 验证 → 工作量。

---

### 阶段 A — 安全/正确性（必须先于功能）

#### A.1 [✗ PENDING] P-08 支付幂等性缺失

**现状**：`CreatePayment` / `RefundPayment` 接口无去重，重试触发多笔订单/退款。grep `Idempotency-Key|SetNX` 全仓 0 命中。

**期望**：客户端可通过 `Idempotency-Key` HTTP header 标识请求；服务端用 Redis `SET NX EX` 抢锁，重复键返回首次结果。

**步骤**：
1. 在 `pkg/middleware` 新增 `Idempotency(redis, ttl)` 中间件：从 header 读 key，`SETNX idem:<key> <hash> EX <ttl>`，未抢到则查 `idem:<key>:result` 直接返回。
2. handler 成功响应后将 `(status, body)` 写入 `idem:<key>:result`。
3. payment-service 路由中，`POST /payments`、`POST /refunds` 注册该中间件，TTL 24h。
4. 文档：`POST` 接口的 Swagger 注释加上 `X-Idempotency-Key` header 必填（`payment-service` Swagger init 后自动同步）。

**验证**：
- 单测：构造同 key 两次请求，断言第二次走缓存且响应字节级相等。
- 集成：用 `httptest` server 模拟客户端重试，订单表只新增 1 条。
- 失败注入：第一次 handler 返回前 panic，键应过期或留尾巴 → 验证 TTL 行为。

**工作量**：~6h（含中间件 + 单测 + 集成测试）。

---

#### A.2 [✗ PENDING] P-19 admission_service 多查询无事务

**现状**：`services/data-service/internal/services/admission_service.go` 中 `PredictAdmission`（line 330）等方法做多次 DB 读取却无事务包裹，`AnalyzeAdmissionData`（line 213）同。grep `tx |Begin|Transaction` 在该文件 0 命中。

**期望**：所有读多条数据后做计算的方法用 `db.Transaction(ctx, func(tx) {...})` 包裹，保证读到一致快照。

**步骤**：
1. 列出该文件中"多 SELECT 计算"的方法：`AnalyzeAdmissionData`、`PredictAdmission`、`GetAdmissionStatistics`。
2. 改造为 `s.db.PostgreSQL.WithContext(ctx).Transaction(...)`，事务内用 `tx.Where(...)` 替换 `s.db.PostgreSQL.Where(...)`。
3. 评估隔离级别：`gorm.WithCustomConfig(&gorm.Config{IsolationLevel: sql.LevelRepeatableRead})` 是否需要——如下游有 admission 写入并发，则需。
4. 单测：mock 两次返回不一致的数据，断言 PredictAdmission 行为符合预期（要么取一致快照，要么报错）。

**验证**：
- 单测覆盖每个改造方法。
- `go test -race` 多 goroutine 并发场景。

**工作量**：~4h。

---

#### A.3 [✓ FIXED] P-05 payment-service 双 PaymentService 共存

**现状（已修复）**：原 `services/payment_service.go`（raw SQL 旧版 dead code）与 `service/payment_service.go`（repository 新版）共存，且 `handler/`（单数 legacy）/ `handlers/`（复数生产）平行污染。

**修复内容（本次推进）**：
1. 迁移 `internal/services/membership_service.go` → `internal/service/membership_service.go`（包名 services → service）；这是 `services/` 目录中唯一的生产代码。
2. 删除 `internal/services/payment_service.go`（dead 旧 PaymentService，main.go 未使用）。
3. 删除 `internal/handler/wechat_handler.go`（legacy build tag 屏蔽的 dead handler）。
4. 删除 `internal/services/` 与 `internal/handler/` 空目录。
5. 同步修正调用方：`main.go`、`internal/middleware/membership_middleware.go`。
6. 清理 `internal/services/*_new_test.go` 两份 legacy 残骸（引用已不存在的 `NewOrderServiceNew`/`NewMembershipServiceNew`，加 -tags legacy 亦无法编译）。

**验证**：
- `go build ./...` workspace 全绿。
- `go test ./...` payment-service 全绿（`handlers` 包 0.695s 通过；含 `membership_handler_routes_test.go` 7 个子测试）。

**残留**：根目录 `integration_test.go`、`payment_functional_test.go` 与 `internal/handlers/{integration,payment_handler,membership_handler}_test.go` 仍带 `//go:build legacy` 标签且 import 已删除的 `internal/services` 包 —— baseline 即不可编译状态，未在本次清理范围。下次集中清理 legacy 时一并处理。

---

### 阶段 B — 性能瓶颈

#### B.1 [✗ PENDING] P-01 LIKE 全表扫描（11 处）

**现状**：data-service 中 11 处 `LOWER(name) LIKE LOWER('%keyword%')`，分布于：
- `university_service.go:266, 271`
- `search_service.go:232, 244, 255, 475, 519, 562, 586`
- `major_service.go:445, 449`

eef5eb7 已部署 pg_trgm GIN 索引但代码未使用。

**期望**：替换为 `name % keyword`（pg_trgm 相似度运算符）或 `to_tsvector('simple', name) @@ plainto_tsquery('simple', keyword)`（全文）。中文场景走 Elasticsearch。

**步骤**：
1. 识别每条查询的语义：模糊匹配（用 trgm）vs 精确分词（用 tsvector）。
2. 在 `pkg/database` 抽 `BuildFuzzyMatch(column, keyword) clause.Expr` helper。
3. 批量替换 11 处，统一通过 helper。
4. 跑 EXPLAIN ANALYZE 验证 GIN 索引被使用（无 Seq Scan）。

**验证**：
- 单测：构造 100K 行假数据，搜索 P95 < 50ms（相比 LIKE 数百 ms）。
- 中文搜索通过 Elasticsearch path 已经存在（`searchUniversitiesES`），确认前端流量真的走 ES。

**工作量**：~4h。

---

#### B.2 [⚠ PARTIAL] P-15 stats 查询合并未到位

**现状**：`university_service.go:399-454` 已把 4 个 COUNT 合到 1 条 SELECT（`Total/By985/By211/ByDoubleFirst`），但 `ByProvince`/`ByType`/`ByNature` 仍是 3 条独立 `GROUP BY` 查询。

**期望**：进一步合一——用 `WITH` CTE 或并行 goroutine 三发，把 4 次 round-trip 降到 1 次或并发 1 次。

**步骤**：
1. 评估 PG 单条 CTE 是否可读：
   ```sql
   WITH base AS (...), p AS (SELECT province, COUNT(*) FROM base GROUP BY province), ...
   SELECT * FROM p, t, n;
   ```
   单条但需 `Raw(...).Scan` 多结果集。可读性差。
2. **推荐方案**：3 条 GROUP BY 改成 `errgroup.Group` 并发执行，3 路 round-trip 并行 → 单次延迟为最慢一路。
3. 单测：mock DB，断言 3 次 SELECT 在并发窗口内发出。

**验证**：
- 基准：`go test -bench=BenchmarkGetStatistics`，对比串行/并发耗时。

**工作量**：~2h。

---

#### B.3 [✗ PENDING] P-14 rateLimiter sync.Map 无 TTL

**现状**：`services/api-gateway/main.go:251-269` 用 `sync.Map` 存 IP→bucket，无淘汰逻辑。长时间运行后 IP 集合膨胀，内存泄漏。

**期望**：bucket 超过 N 分钟未访问自动清除。

**步骤**：
1. 给 `rateBucket` 加 `last time.Time` 字段（已有，但未用作淘汰）。
2. `newRateLimiter` 启 cleanup goroutine：每 1min 扫一次，删除 `time.Since(last) > 10min` 的 entry。
3. shutdown 时关闭 cleanup（`context.WithCancel`）。

**验证**：
- 单测：插入 1000 个 key，等 1s 后调 cleanup（test 用注入时钟），断言 sync.Map 长度归 0。

**工作量**：~2h。

---

#### B.4 [⚠ PARTIAL] P-17 search goroutine ctx 解耦

**现状**：`search_service.go:606-630` 用 `context.WithTimeout(context.Background(), 5*time.Second)`。**没用传入的 `ctx`**，请求取消不传播。但 5s timeout 兜底，不是真泄漏。

**期望**：goroutine 内派生自请求 ctx：`recordCtx, cancel := context.WithTimeout(ctx, 5*time.Second)`，请求取消能立刻终止 goroutine。

**步骤**：
1. 改 `recordSearch` 函数体，把 line 608 的 `Background()` 换成入参 `ctx`。
2. 同样审 `algorithm_service.go:538`（saveAnalysisResult），同样改造。

**验证**：
- 单测：构造可取消的 ctx，goroutine 启动后立即取消；检查 DB write 未发生（mock DB）。

**工作量**：~1h（机械改）。

---

#### B.5 [✗ PENDING] P-25 algorithm_service goroutine 无 ctx 取消

合并到 B.4 一并改造。

---

#### B.6 [✗ PENDING] P-24 100MB 文件全量读入内存

**现状**：`services/data-service/internal/services/data_import_service.go:30` 用 `io.ReadAll(file)` 一次性读入。

**期望**：`json.NewDecoder(file).Decode(&v)` 流式解析；如果是 NDJSON/数组逐项处理，用 `decoder.Token()` 流式。

**步骤**：
1. 看 import 的输入格式（JSON 数组 / NDJSON / object）。
2. 重写为 `for decoder.More() { decoder.Decode(&item); process(item) }`。
3. 增加 `--max-memory` 监控点（pprof）确认内存峰值下降。

**验证**：
- 用 100MB 测试文件运行，内存峰值应 << 100MB（理想 <10MB）。

**工作量**：~3h。

---

#### B.7 [⚠ PARTIAL] P-13 Logrus 每请求 2x WithFields

**现状**：`api-gateway/main.go:753, 767` — 每请求构造 `logrus.Fields{}` 两次（pre + post）。在高 QPS 下分配开销可观。

**期望**：合并为 1 次构造（请求结束统一打），或用 `logrus.WithFields().Info()` 一次性。

**步骤**：
1. Read 753-769 上下文确认两次的目的（往往是 access log + error log 重复）。
2. 引入请求级 `*logrus.Entry`，附在 ctx 上，请求结束 1 次 flush。
3. 或采样：仅 5xx + 慢请求 (>500ms) 打 detailed log。

**验证**：
- 基准：`go test -bench=BenchmarkRequestPath`，分配数应减半。

**工作量**：~2h。

---

#### B.8 [✗ PENDING] P-22 / P-23 跨服务 debug log 求值 / JSON 序列化

**现状**：跨服务多处 `logger.Debug(fmt.Sprintf(...))` 即使日志级别 < Debug 仍执行 sprintf；缓存路径每次 `json.Marshal/Unmarshal` 不复用 buffer。

**期望**：
- Debug 路径用 lazy logging：`logger.Debugf(...)` 或 `if logger.IsLevelEnabled(DebugLevel) { ... }` 守卫。
- 缓存路径用 `sync.Pool[bytes.Buffer]` + `jsoniter`（替换标准 json）。

**步骤**：
1. grep 全仓 `logger.Debug.*Sprintf|logrus.Debug.*Sprintf` 列出所有点。
2. 引入 `pkg/logger.LazyDebug(fn func() string)` 辅助函数。
3. 缓存热路径替换 `encoding/json` → `github.com/json-iterator/go`。

**验证**：
- pprof CPU profile 对比，json/log allocs 降至原 30%。

**工作量**：~6h（跨服务，需谨慎）。

---

### 阶段 C — 重复代码治理

#### C.1 [✗ PENDING] F. BeforeCreate UUID hooks（16 处）

**现状**：16 处 `func (X *Y) BeforeCreate(tx *gorm.DB) error { if X.ID == uuid.Nil { X.ID = uuid.New() }; return nil }` 完全相同。分布于 user-service(7) + data-service(6) + payment-service(3)。

**期望**：在 `pkg/models` 定义 `UUIDBaseModel` 嵌入式基类，仅写一次 BeforeCreate。

**步骤**：
1. 在 `pkg/models` 新增：
   ```go
   type UUIDBaseModel struct {
       ID uuid.UUID `gorm:"type:uuid;primaryKey"`
       CreatedAt time.Time
       UpdatedAt time.Time
   }
   func (m *UUIDBaseModel) BeforeCreate(tx *gorm.DB) error {
       if m.ID == uuid.Nil { m.ID = uuid.New() }
       return nil
   }
   ```
2. 各模型把 `ID uuid.UUID` 字段替换为 `models.UUIDBaseModel` 嵌入。
3. 删除 16 处 BeforeCreate。
4. 跑测试确认 GORM 嵌入字段 hook 被正确识别（GORM 自动用最外层的 BeforeCreate，嵌入应也能 work，需测）。

**风险**：GORM 嵌入 hook 的可见性需验证；如果嵌入失效，回退方案是 `pkg/models.NewUUID()` helper + 各处显式调用。

**验证**：
- 集成测试：创建 X 记录，断言 `ID != uuid.Nil`。

**工作量**：~3h。

---

#### C.2 [⚠ PARTIAL] G. APIResponse 重复（仍 3 处）

**现状**：3 处独立定义：
- `services/user-service/internal/models/models.go:54`
- `services/payment-service/internal/models/payment_models.go:298`
- `services/recommendation-service/internal/handlers/simple_recommendation_handler.go:191`

`pkg/response` 已存在，data-service 已通过类型别名接入。

**期望**：3 处全部替换为 `pkg/response.Response`。

**步骤**：
1. 各服务 go.mod 加 `pkg/response` 依赖。
2. 用 type alias `type APIResponse = response.Response` 平滑迁移（避免 100+ 处 call site 改动）。
3. 后续逐步把 call site 改为直接引用 `response.Response`。

**验证**：
- 编译 + 测试。
- 跨服务接口契约不变（字段名/JSON 结构相同）。

**工作量**：~2h（type alias 阶段）+ 后续逐步。

---

#### C.3 [⚠ PARTIAL] A. handler 错误响应 100+ 处 gin.H{}

**现状**：payment-service handler 仍 30+ 处 `c.JSON(http.StatusXxx, gin.H{"error": ..., "message": ...})`。其他服务未审。

**期望**：统一通过 `pkg/response` 工厂函数：`response.BadRequest(c, "msg")`、`response.OK(c, data)` 等。

**步骤**：
1. 先在 `pkg/response` 补齐工厂函数（若缺）。
2. 写一个 codemod / sed 脚本批量替换最常见的 4 种 pattern（Bad/Internal/NotFound/OK）。
3. 人工审剩余的特殊 case（带 details 字段的）。

**验证**：
- grep `gin.H{` 在 services/ 中应大幅下降（从 100+ 到 < 20）。

**工作量**：~6h（机械化部分 1h，人工审 5h）。

---

#### C.4 [✗ PENDING] H. payment-service JSONB 双类型

**现状**：`services/payment-service/internal/models/`：
- `membership_models.go:11` → `type JSONB map[string]interface{}`
- `payment_models.go:15` → `type PaymentJSONB map[string]interface{}`

**期望**：合并为 `pkg/models.JSONB`（或 payment-service 内部 `JSONB`），删除 `PaymentJSONB`。

**步骤**：
1. 决定是否泛化到 `pkg/models`（其他服务也可能需要）。当前其他服务没用，先内部统一即可。
2. 把所有 `PaymentJSONB` 改名为 `JSONB`，删除 `payment_models.go:14-34` 的 `PaymentJSONB` 定义。
3. 测试。

**验证**：
- `go build ./...` + 测试。

**工作量**：~1h。

---

#### C.5 [✗ PENDING] E. monitoring/api-gateway Redis 未走 shareddb

**现状**：未审具体。需先 grep 确认 monitoring-service 与 api-gateway 各自如何初始化 Redis。

**步骤**：
1. 先做 spike：grep `redis.NewClient` in services/。如果有 `pkg/database.OpenRedis()` 或类似 helper，迁移。
2. 如 helper 不存在，先在 `pkg/database` 抽出 `OpenRedis(cfg) (*redis.Client, error)`，然后迁移。

**工作量**：先做调研 1h，看具体差异决定。

---

#### C.6 [⚠ PARTIAL] I. SecurityHeaders 仍 2 处自定义

**现状**：
- `services/api-gateway/main.go:461` 已用 `securityMiddleware.SecurityHeaders()` ✓
- `services/data-service/internal/middleware/middleware.go:224` 直接 `c.Header("X-Content-Type-Options", "nosniff")` 自定义 ✗
- `services/payment-service/internal/middleware/middleware.go:213-216` 自定义 ✗

**期望**：data-service / payment-service 也接 `pkg/middleware`。

**步骤**：与 C.1 类似，统一接入 pkg/middleware.SecurityHeaders。

**工作量**：~1h。

---

### 阶段 D — 算法升级（Phase 4）

⚠ **重要**：此阶段是用户最初路线图的 Phase 4，影响产品核心价值（推荐质量），与前面性能/重复治理是不同维度的工作。建议在阶段 A/B 完成、产品线稳定后再启动。

#### D.1 [✗ PENDING] 3.4 录取概率改正态分布 CDF

**现状**：`services/recommendation-service/pkg/cppbridge/simple_rule_bridge.go:229-260` 用硬编码 7 阶梯 switch，29→30 跳 10%，省份差异忽略。

**期望**：`P = Φ((score - μ) / σ)`，μ/σ 来自该校该专业近 3 年录取数据。

**步骤**：
1. 在 admission 数据 schema 上确认 μ（avg_score）/σ（std_score）字段存在。如缺，补 SQL 计算。
2. 实现 `normalCDF(x float64) float64 { return 0.5 * (1 + math.Erf(x / math.Sqrt2)) }`。
3. `calculateAdmissionProbability` 重写：取最近 3 年记录，加权平均求 μ/σ，返回 `clamp(normalCDF(scoreDiff/sigma), 0.01, 0.99)`。
4. 同步改 `enhanced_rule_bridge.go:325` 同款方法。

**风险**：
- 旧测试断言概率值（"分数差 30 → 0.95"）会全部 fail。需重写期望值。
- σ 估计噪声大（3 年样本太少）→ 加最小 σ floor（如 σ ≥ 5）。

**验证**：
- 单测：σ=10、scoreDiff=20 → 应 ≈ Φ(2.0) = 0.9772。
- 输出对比：5 个典型 case 前后概率（确认无突变）。
- 业务验证：让产品/算法同学审一组 case 是否符合直觉。

**工作量**：~6h（含测试改写）。

---

#### D.2 [✗ PENDING] 3.3 "冲稳保"三分法

**现状**：`services/recommendation-service/internal/handlers/simple_recommendation_handler.go:1393-1399`：
```go
recType := "稳妥"
if rec.Probability >= 0.8 { recType = "稳妥" }
else if rec.Probability >= 0.6 { recType = "适中" }
else { recType = "冲刺" }
```

问题：缺"保底"类，命名非标准，仅基于 Probability 单维度，返回平铺无分组。

**期望**：标准三分法 — "冲" / "稳" / "保"，按类别分组返回，至少考虑 (probability, score_gap) 两维。

**步骤**：
1. 定义阈值（与产品/算法同学对齐）：
   - 保：probability ≥ 0.85 且 scoreDiff ≥ +20
   - 稳：probability ≥ 0.55 且 scoreDiff ≥ -5
   - 冲：probability ≥ 0.20 且 scoreDiff ≥ -25
2. 改返回结构：从 `[]Recommendation` 改为 `{ "rush": [...], "stable": [...], "safety": [...] }`。
3. 同步 `analyzer.go:165-174` 的话术（line 174 的格式化字符串"稳妥%d个，适中%d个，冲刺%d个"也得改）。
4. 改前端 ts 类型（`frontend/src/types/recommendation.ts` 之类）+ 视图组件（`RecommendationList.vue` 之类）。

**风险**：分组结构变更属 breaking API change，需协调前端同步上线。考虑双写 1-2 周（同时返回 `rec_type` + `category`）平滑过渡。

**验证**：
- 后端单测覆盖 3 类边界。
- 前端 e2e 验证三组都正确渲染。

**工作量**：~12h（后端 4h + 前端 6h + 测试 2h）。

---

#### D.3 [✗ PENDING] 3.1 ML 模块 stub

**现状**：`services/recommendation-service/pkg/ml/ml_enhanced_engine.go`：
- `DeepLearningModel.Predict`: `base + (gap/100)*0.2` 线性占位
- `CollaborativeFilter.GetAdjustment`: 硬返回 `0.02`
- `ContentBasedFilter.GetAdjustment`: 硬返回 `0.01`
- `ReinforcementLearner.Optimize`: `base + cf + cb` 简单加法
- `FeatureEngineering`: O(n²) 双循环 dead code

**期望**：要么真做（接 ML pipeline），要么全部删除——目前是误导性命名（声称"深度学习"实为常数返回）。

**推荐路径**：**先删除**（最小化代码维护成本，避免误导）。真上 ML 是另一个量级的工作（数据集 + 训练 + 部署），不在本路线图。

**步骤（删除路径）**：
1. grep 调用方：`MLEnhancedRecommendationEngine` 是否在生产路径上被使用？如果只是 dead code，直接删除整个 `pkg/ml/`。
2. 如有调用方，把调用方改为直接走 `simple_rule_bridge`/`enhanced_rule_bridge`。
3. 删除 `pkg/ml/ml_enhanced_engine.go`、相关 init、测试。

**验证**：
- `go build ./...` 通过。
- 推荐请求 happy path 不变。

**工作量**：~2h（删除路径）；~40h+（真实现路径，超出本路线图）。

---

#### D.4 [✗ PENDING] 3.5 Score / Confidence 语义冲突

**现状**：`generateEnhancedRecommendations`（推 reco handler 第 849 行附近）用 confidence 覆盖 Score 字段，导致 Score 原本表示"多维匹配分"的语义丢失。

**期望**：保留 Score（多维匹配）+ 新增 Confidence（概率置信度）两个独立字段。

**步骤**：
1. Read 实际 line 849 周围确认覆盖发生位置。
2. 分离字段：response struct 加 `Confidence float64` 字段。
3. 计算路径：Score 仍由 bridge 层算多维匹配，Confidence 由 P-D.1 的 normalCDF 输出。
4. 前端：列表卡片同时展示两个数值（设计同学决定 UX）。

**工作量**：~4h（与 D.2 一并做）。

---

#### D.5 [✗ PENDING] 3.7 analytics 硬编码数据

**现状**：
- `services/recommendation-service/internal/services/analytics_service.go:336-398` `GetRecommendationTrends` 全 mock
- 同文件 `:741-763` `getTopRecommendations` 硬编码"清华大学/北京大学/上海交通大学"

**期望**：从 ClickHouse / data-service / 推荐日志聚合真数据。

**步骤**：
1. 评估数据源：是否已有推荐日志表？没有则需新建（建议 ClickHouse / Postgres `recommendation_log` 表，user_id + university_id + score + timestamp）。
2. 改造为 `SELECT university_id, COUNT(*) FROM recommendation_log WHERE created_at > NOW() - INTERVAL '7 days' GROUP BY university_id ORDER BY count DESC LIMIT 10`。
3. 加 Redis 缓存（5min TTL）避免每个分析请求都打 DB。

**风险**：如无推荐日志表，需要先建数据管道（应用层埋点 → 写入），是基础设施级工作。

**工作量**：~1d 起（含管道建设）。

---

## 3. 优先级建议

| 优先级 | 项目 | 理由 |
|--------|------|------|
| **P0**（立即） | A.1 P-08 幂等性 / A.2 P-19 事务 / A.3 P-05 双 Service 整合 | 安全/正确性，触线生产即事故 |
| **P1**（本周） | B.1 P-01 LIKE / B.3 P-14 rateLimiter / B.6 P-24 流式读 | 性能与稳定性瓶颈 |
| **P2**（下周） | B.2 P-15 stats / B.4 P-17/P-25 ctx / B.7 P-13 logrus | 优化项 |
| **P3**（第 3-4 周） | C.1 BeforeCreate / C.2 APIResponse / C.4 JSONB / C.6 SecurityHeaders | 重复代码治理 |
| **P4**（持续） | C.3 gin.H / C.5 Redis init | 长尾治理 |
| **P5**（产品决策后） | D.1 CDF / D.2 三分法 / D.4 Score 分离 | 算法升级，需产品同步 |
| **P6**（评估后） | D.3 ML stub 删除 / D.5 真数据接入 | D.3 易；D.5 需基础设施投入 |

---

## 4. 验证基线

每个改动 PR 必须满足：
- `go test ./...` 全绿
- 相关 unit test 覆盖率 ≥ 60%（与 CI 阈值对齐）
- `go vet ./...` 0 warning
- `gosec ./...` 无新增高危
- `golangci-lint run` 通过（沿用现 v1 配置；如启动 .golangci.yml v2 迁移则同时升级 action v6→v8）

跨服务改动（如 C.1、C.2）需要前端 e2e 配合：
- `cd frontend && npm run test:unit && npm run test:e2e`

---

## 附录：本次复审用到的 grep 凭证

为下次复审可重放，记录关键 grep 命令：

```bash
# A.1 幂等性
rg "Idempotency-Key|SetNX|SETNX" services/

# B.1 LIKE
rg "LIKE|ILIKE" services/data-service/internal/services/

# C.1 BeforeCreate 计数
rg -c "func\s+\(\w+\s+\*\w+\)\s+BeforeCreate" services/

# G. APIResponse 重复
rg "type\s+APIResponse\s+struct" services/

# H. JSONB 重复
rg "type\s+\w*JSON\w*\s+|JSONB" services/payment-service/internal/models/
```
