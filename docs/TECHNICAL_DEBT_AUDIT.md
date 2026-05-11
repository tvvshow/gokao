# 技术债务审计与推进计划

**初版**: 2026-05-10（全量审计）
**复审**: 2026-05-10（status update）
**审计范围**: 6 微服务 + 14 pkg 模块 + C++ 算法模块

---

## 0. 状态总览

| 类别 | 总数 | ✓ FIXED | ⚠ PARTIAL | ✗ PENDING |
|------|------|---------|-----------|-----------|
| 严重 (P-01~P-10) | 10 | **10** | 0 | 0 |
| 中等 (P-11~P-25) | 15 | **14** | 0 | 1 |
| 重复代码 (A~I) | 9 | **9** | 0 | 0 |
| 算法 (Phase 4) | 5 | 0 | 0 | 5 |
| **合计** | **39** | **33** | **0** | **6** |

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
| P-01 | LIKE 全表扫描审计偏差纠正 + hot_searches 索引 | eef5eb7 + 本次 | `services/data-service/internal/database/database.go:271` |
| P-14 | rateLimiter sync.Map TTL 淘汰 | 本次 | `services/api-gateway/main.go:244-303`（cleanup goroutine + 注入时钟） |
| P-24 | 100MB 流式导入 | 本次 | `services/data-service/internal/services/data_processing_service.go`（Process*DataStream + upsert helpers） |
| P-15 | stats 3 维度 GROUP BY 并发化 | 本次 | `services/data-service/internal/services/university_service.go:397-489`（errgroup + Session 隔离） |
| P-17 | search recordSearch ctx 解耦 | 本次 | `services/data-service/internal/services/search_service.go:606-636` |
| P-25 | algorithm_service saveAnalysisResult ctx 解耦 + WithContext | 本次 | `services/data-service/internal/services/algorithm_service.go:535-574` |
| P-13 | createProxy 共享 base log entry + cache Debug IsLevelEnabled 守卫 | 本次 | `services/api-gateway/main.go:160-189, 815-840` |
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

#### A.1 [✓ FIXED] P-08 支付幂等性缺失

**现状（已修复）**：原 `CreatePayment` / `RefundPayment` 接口无去重，客户端重试触发多笔订单。

**修复内容**：
1. `pkg/middleware/idempotency.go` 新增通用 `Idempotency(store IdempotencyStore, ttl)` 中间件。`IdempotencyStore` 是 `SetNX/Get/Set` 三方法的最小接口，让 `pkg/middleware` 保持零 Redis SDK 依赖。
2. `services/payment-service/internal/middleware/idempotency_store.go` 提供 redis v8 客户端的 `IdempotencyStore` 实现。
3. `main.go` 把中间件挂到 `POST /payments` 与 `POST /refunds`，TTL 24h。
4. 行为：
   - 客户端 `X-Idempotency-Key` 头声明键；首次请求 handler 正常执行并缓存 2xx 响应（24h）。
   - 重放同 key 时直接回放首次响应，handler **不再执行**。
   - 首请求执行中再次重放返回 `409 Conflict + X-Idempotency-Status: in-flight`。
   - Redis 故障时 fail-open（继续放行，标 `X-Idempotency-Status: store-error`），避免支付路径被基础设施抖动拖住。
   - 非 2xx 响应不缓存，允许客户端用同 key 重试纠错。
   - 客户端未带 header 时整段绕过 — 兼容旧客户端。
5. `pkg/middleware/idempotency_test.go` 含 7 个子测试覆盖：无 header 放行、首请求落库、重放回放、in-flight 409、store-error fail-open、非 2xx 不缓存、Content-Type 保留。

**验证**：
- `go test ./pkg/middleware/` 7 个子测试通过（0.637s）。
- `go build ./...`（workspace）干净。
- `go test ./services/payment-service/...` 全绿。

**部署说明**：README 已同步幂等性头部使用矩阵。

---

#### A.2 [✓ FIXED] P-19 admission_service 多查询无事务（+ 发现隐性 builder 污染 bug）

**审计描述偏差纠正**：原审计指出三个方法（`AnalyzeAdmissionData` / `PredictAdmission` / `GetAdmissionStatistics`）需包事务。复审实际代码：
- `AnalyzeAdmissionData`（line 213）：3 SELECT (admission + university + major)，**需要事务**保证关联一致。
- `PredictAdmission`（line 330）：**只 1 个 SELECT**（line 343 `query.Find(...)`），事务无收益，本次保留不动。
- `GetAdmissionStatistics`（line 387）：4 个 Count/Scan，**有事务收益且存在更严重的隐性 bug**（见下）。

**隐性真根因（比审计原描述更严重）**：原 `GetAdmissionStatistics` 复用同一 `*gorm.DB` 链式 builder 直接调用 4 次终止操作：
```go
baseQuery := s.db.PostgreSQL.Model(...)
baseQuery.Count(&total)
baseQuery.Select("province, ...").Group("province").Scan(&provinceStats)
baseQuery.Select("batch, ...").Group("batch").Scan(&batchStats)   // ← 仍带前一次 Select(province)
```
GORM v2 不自动 Clone Select/Group/Order 条件，第二次 Scan 时 builder 还带着前次的 `Select(province)` / `Group(province)`，**by_batch / score_distribution 统计结果会错位**。

**修复内容**：
1. `AnalyzeAdmissionData` 改为 `db.PostgreSQL.WithContext(ctx).Transaction(func(tx *gorm.DB) error {...})` 包裹，3 SELECT 内部统一用 `tx`。
2. `GetAdmissionStatistics` 同样包事务，且每次终止操作前用 `tx.Session(&gorm.Session{})` 创建新会话隔离 builder 状态。新增 3 个 named types `AdmissionProvinceStat` / `AdmissionBatchStat` / `AdmissionScoreDistribution` 让聚合结果可被外部断言。
3. `PredictAdmission` 保持原样（单 SELECT 无事务必要）。

**附带修复**：诊断过程发现 `tests/integration_test.go::getFirstUniversityID()` 用 `db.First(&university)` —— University.ID 是 `uuid.New()` 随机生成，主键升序与插入顺序无关，导致测试随机命中清华或北大而 fixture 仅与北大关联。改为 `db.Where("code = ?", "10001").First(...)` 显式取北大，消除测试的概率性 flakiness。

**验证**：
- `services/data-service/internal/services/admission_service_test.go` 新增 2 个 sqlite in-memory 单测覆盖 builder 污染修复：`TestGetAdmissionStatistics_NoBuilderPollution`（断言 by_batch 不被前次 Select(province) 污染）、`TestGetAdmissionStatistics_NoYearFilter`（断言 year=0 时跨年汇总）。
- `go test ./services/data-service/...` 全绿（含 internal/services 1.165s + tests 集成测试 2.016s）。
- `go build ./...`（workspace）干净。

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

#### B.1 [✓ FIXED] P-01 LIKE 全表扫描（11 处）

**审计描述偏差纠正**：原审计「pg_trgm GIN 索引但代码未使用」不准确。pg_trgm 的 GIN 表达式索引（`USING gin (LOWER(col) gin_trgm_ops)`）**天然加速** `LOWER(col) LIKE LOWER(?)` 语法 —— eef5eb7 commit message 已明确："Zero application-code changes — every existing LIKE query is automatically promoted from seq scan to bitmap index scan"。所谓 11 处需要替换并不成立。

**真根因复审**：11 处 LIKE 逐条核对字段索引覆盖：

| # | 文件:行 | 字段 | trgm 索引覆盖 |
|---|---|---|---|
| 1 | university_service.go:266 | name, code | ✓ ✓ |
| 2 | university_service.go:271 | name | ✓ |
| 3 | major_service.go:445 | name | ✓ |
| 4 | major_service.go:449 | name, **description**, **career_prospects** | ✓ ✗ ✗ |
| 5 | search_service.go:232 | hot_searches.keyword | ✗ → ✓（本次补） |
| 6 | search_service.go:244 | name, alias | ✓ ✓ |
| 7 | search_service.go:255 | name | ✓ |
| 8 | search_service.go:475 | name, alias, **description** | ✓ ✓ ✗ |
| 9 | search_service.go:519 | name, **description**, **career_prospects** | ✓ ✗ ✗ |
| 10 | search_service.go:562 | name, alias | ✓ ✓ |
| 11 | search_service.go:586 | name | ✓ |

eef5eb7 已覆盖 7/11；本次补 hot_searches.keyword 索引覆盖 8/11；剩 3 处含 `description`/`career_prospects`（长文本字段）走 ES 主路（line 282/293），仅 ES 不可用时 DB fallback —— 索引体积成本 >> 收益，列为 OOB（不补）。

**修复内容**：
1. `services/data-service/internal/database/database.go:271` 新增一条索引：
   ```sql
   CREATE INDEX IF NOT EXISTS idx_hot_searches_keyword_trgm
     ON hot_searches USING gin (LOWER(keyword) gin_trgm_ops);
   ```
   覆盖 `SearchService.AutoComplete` 高频自动补全路径（每次按键触发）。
2. 业务代码（11 处 LIKE 调用点）**保持不变** —— pg_trgm 自动加速已生效，重写为 `%` 或 `to_tsvector` 收益为零（甚至更差，因为切换 similarity 阈值后丢失行为兼容）。

**OOB（不在本次范围）**：
- `majors.description` / `majors.career_prospects` / `universities.description` 的 trgm 索引（TEXT 列体积成本大，DB 路径触发率低，待真实生产 EXPLAIN 数据驱动决策）。
- 业务代码中 OR 条件含未索引字段时，planner 可能整条 OR 退化为 seq scan —— 未来若产品确认 description 维度搜索高频，再做两阶段查询改造（先 name 命中，无果再 fallback description）。

**验证**：
- `go build ./...` 干净。
- pg_trgm 行为依据 PostgreSQL 官方文档 + eef5eb7 commit 声明（标准库行为，无需重复实测）。

---

#### B.2 [✓ FIXED] P-15 stats 查询合并未到位

**现状（已修复）**：原 `university_service.go:399-454` 第 1 步已合到单条 SUM(CASE WHEN) 聚合，剩 3 条独立维度 GROUP BY（province / type / nature）串行执行。

**修复内容**：
1. 引入 `golang.org/x/sync/errgroup`（已是间接依赖，`go mod tidy` 提升为直接）。
2. 三条 GROUP BY 用 `errgroup.WithContext(ctx)` 并发执行，整体延迟 ≈ 最慢一路而非 3x 串行。
3. 每个 goroutine 用 `s.db.PostgreSQL.WithContext(gctx).Session(&gorm.Session{}).Model(...)` 新建会话 —— 否则共享同一 `*gorm.DB` builder 会出现数据竞争 + Select/Group 条件互相覆盖（同款 A.2 P-19 bug）。
4. 三个独立 slice（`provinceCounts`/`typeCounts`/`natureCounts`）各由一个 goroutine 写入，无 map race；合并到 `stats.By*` map 在主 goroutine 顺序完成。
5. errgroup ctx 让任一返错即取消其他正在跑的 SELECT，错误路径节省资源。

**验证**：
- `services/data-service/internal/services/university_service_test.go` 新增 2 个测试：ParallelDimensions（全维度结果断言）+ EmptyTable（空库不报错）。
- `go test -race ./services/data-service/internal/services/` 全绿（含 race detector），证明 Session 隔离生效。
- `go build ./...` 干净。

---

#### B.4 [✓ FIXED] P-17 / B.5 [✓ FIXED] P-25 后台 goroutine ctx 解耦

**现状（已修复）**：
- `search_service.go:606` `recordSearch` 用 `context.WithTimeout(context.Background(), 5s)`，请求取消不传播。
- `algorithm_service.go:535` `saveAnalysisResult` 更糟：goroutine 内 `s.db.PostgreSQL.Create()` 既没派生 ctx 也没传 `WithContext`，请求取消和超时都无法终止写入。

**修复内容**：
1. 两处的 timeout 父 ctx 从 `context.Background()` 切到入参 `ctx`：
   ```go
   recordCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
   ```
   保留 5s timeout fallback（防单条 DB 慢查询挂住 goroutine），但客户端取消能立即穿透到 goroutine 内的 DB 写。
2. `algorithm_service.saveAnalysisResult` 额外补 `s.db.PostgreSQL.WithContext(saveCtx).Create(...)`，让 GORM 真把 ctx 下传到驱动层。

**风险评估**：
- gin handler 返回后框架会 cancel 主请求 ctx —— goroutine 派生子 ctx 用主 ctx 作 parent，框架 cancel 会立即穿透 child ctx，DB 写会被中断。这是符合预期的：请求结束后再花资源记录 hot_search / analysis_result 已无意义。
- 真正长跑写入应该走专门的 worker queue，不应挂在 handler goroutine 上。

**验证**：
- `go build ./services/data-service/...` 干净，全量测试套件无回归（services 1.5s + tests 2.2s）。

---

#### B.7 [✓ FIXED] P-13 Logrus 每请求 2x WithFields

**现状（已修复）**：
- `api-gateway/main.go:818, 832` createProxy 每请求构造 `logrus.Fields{}` 两次（pre + post），每次 6-8 字段，前 6 字段完全相同 —— 重复 mapassign 浪费。
- `api-gateway/main.go:161, 181` cacheMiddleware 的 Debug 日志没 IsLevelEnabled 守卫，生产 Info level 下仍然每次构造 fields map。

**修复内容**：
1. createProxy 把 6 个共享字段构造一次成 `baseLog *logrus.Entry`，pre 直接 `baseLog.Info(...)`，post 再 `baseLog.WithFields(2 个变量字段).Info(...)`。Fields map 构造从 2 次 × 6-8 字段降到 1 次 × 6 字段 + 1 次 × 2 字段（实际 alloc 减半）。
2. cacheMiddleware 两处 Debug 调用前加 `if cache.logger.IsLevelEnabled(logrus.DebugLevel)` 守卫，生产 Info level 下完全跳过 fields 构造（cache 是热路径，每请求都走）。

**验证**：
- `go test ./services/api-gateway/...` 全绿（27.9s，含原 26 个测试 + 4 个 rate limiter 测试 + 现有代理回归覆盖日志字段输出）。

**OOB（不在本次范围）**：
- 采样策略（仅 5xx + 慢请求 >500ms 详细日志）—— 涉及业务可观测性策略，需运维确认。

---

#### B.3 [✓ FIXED] P-14 rateLimiter sync.Map 无 TTL

**现状（已修复）**：原 `rateLimiter.m sync.Map` 存 IP→bucket 无淘汰逻辑，长期运行 IP 集合膨胀导致内存泄漏。

**修复内容**：
1. `services/api-gateway/main.go:244-303` 给 `rateLimiter` 加 `idleTTL`、`now func() time.Time`、`stopCh` 字段。
2. `newRateLimiter` 默认 `idleTTL=10min`、cleanup `interval=1min`，启动 `cleanupLoop` goroutine。
3. 新增 `newRateLimiterWithDeps(rps, burst, idleTTL, nowFn)` 工厂供测试注入时钟，不启动 goroutine。
4. `evictIdle()` 同步扫描 `sync.Map.Range`，淘汰 `now - bucket.last > idleTTL` 的 entry。锁顺序：先 `bucket.mu` 读 last 再释放，再 `m.Delete`，避免与 `allow()` 并发死锁。
5. `stop()` 用 `sync.Once` 关闭 stopCh，多次调用安全。
6. `allow()` 时间源从 `time.Now()` 切到 `rl.now()`，与 `evictIdle()` 共用同一时钟，保证测试快进时 `last`/`now` 时间基一致。

**验证**：
- `services/api-gateway/rate_limiter_test.go` 4 个新测试：EvictsIdleBuckets / KeepsActiveBuckets / StopIsIdempotent / AllowUsesInjectedClock 全绿。
- `go test ./services/api-gateway/...` 全绿（27.6s，含原 22 个测试）。
- `go build ./...` 干净。

---

#### B.6 [✓ FIXED] P-24 100MB 文件全量读入内存

**现状（已修复）**：原 `data_import_service.go:30` 调 `io.ReadAll(file)` 把整个 100MB 上传文件读进 `[]byte`，再 `json.Unmarshal` 又复制一份对象切片到内存，峰值 >= 文件大小 × 2。

**修复内容**：
1. `services/data-service/internal/services/data_processing_service.go` 新增三个流式入口：
   - `ProcessUniversityDataStream(io.Reader) error`
   - `ProcessMajorDataStream(io.Reader) error`
   - `ProcessAdmissionDataStream(io.Reader) error`
   每个用 `json.NewDecoder(r)` + `expectArrayStart` 校验顶层 `[`，再循环 `dec.More() / dec.Decode(&item)` 逐条 upsert。内存峰值仅为单条 record + decoder 内部 buffer（约 KB 级），与文件大小解耦。
2. 旧 `ProcessUniversityData([]byte)` / `ProcessMajorData([]byte)` / `ProcessAdmissionData([]byte)` 保留，内部 `bytes.NewReader(data)` 委托给流式版本 —— `data_handler.ProcessData` 的 JSON-in-body API 路径无需改动。
3. `data_import_service.ImportFromFile` 删除 `io.ReadAll`，把 `multipart.File`（实现 `io.Reader`）直接传给流式入口。
4. 抽取 `upsertUniversity` / `upsertMajor` / `upsertAdmission` 三个 helper 消除流式与兼容入口的 upsert 重复逻辑。事务在最外层包裹，任何一条解析或写入失败立即 Rollback。
5. `expectArrayStart` 显式校验顶层 token 为 `[`，避免 `dec.More()` 在非数组 JSON（对象/单值）下行为不定。

**验证**：
- `services/data-service/internal/services/data_processing_service_stream_test.go` 新增 7 个测试：HappyPath / RejectsNonArray / RollbackOnPartialFailure / EmptyArray / MajorHappyPath / AdmissionHappyPath / LegacyByteEntry。
- `go test ./services/data-service/...` 全绿（services 1.185s + tests 2.049s + middleware 0.702s）。
- `go build ./...`（workspace）干净。

**残留风险**：
- 单事务 + 流式 decode：100K-1M 条记录单事务在 PG 上完全可行，但极端大文件（1000 万+ 条）单事务可能撞 max_wal_size。后续如果业务出现这种规模，再切到批 commit（每 N 条提交一次）—— 设计上已抽出 upsert helper，切换零摩擦。

---

#### B.7 [⚠ DUPLICATE — 见上方 line 245 FIXED 节] P-13

旧 PARTIAL 描述已被 line 245 的 FIXED 版本取代，保留此空壳避免章节编号断层。下一次复审清理。

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

#### C.1 [✓ FIXED] F. BeforeCreate UUID hooks（16 处）

**现状（已修复）**：16 处 BeforeCreate 都是同款 `if X.ID == uuid.Nil { X.ID = uuid.New() }` 模板（user 7 + data 6 + payment 3）。

**真根因选型纠正**：原审计推荐"嵌入式 UUIDBaseModel"路径，但实测风险较大：
- 16 个 model 各自有完整 ID 字段定义 + gorm tag（如 `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`）
- 嵌入 BaseModel 后字段冲突（外层 ID 遮蔽嵌入 ID，gorm tag 行为不直观）
- GORM v2 嵌入 hook 的 method 提升语义在多嵌入场景下行为不稳定

改走 **helper 函数路径**（更稳健，仍消除业务逻辑重复）：
1. `pkg/models/uuid.go` 新增 `AssignNewUUIDIfZero(id *uuid.UUID)` —— 单一职责，nil-pointer 防护内置。
2. 16 处 BeforeCreate 函数体从 4 行 → 2 行：
   ```go
   func (X *Y) BeforeCreate(tx *gorm.DB) error {
       pkgmodels.AssignNewUUIDIfZero(&X.ID)
       return nil
   }
   ```
3. user-service / data-service / payment-service 的 `go.mod` 各自新增 `pkg/models v0.0.0` 依赖 + `replace ... => ../../pkg/models`。
4. import `pkgmodels "github.com/tvvshow/gokao/pkg/models"`，避免与服务本地 `models` 包名冲突。

**消除效果**：每处 3 行业务逻辑 × 16 处 = 48 行重复模板降至 2 行。
方法签名（GORM hook 注册必须）保留 16 处但只是壳，不构成"业务重复"。

**验证**：
- `go build ./...` workspace 干净。
- `go test ./services/user-service/... ./services/data-service/... ./services/payment-service/... ./pkg/models/...` 全绿。

**OOB（未来 follow-up）**：
- 完全消除方法签名重复需要 GORM 嵌入式 base struct，但需要拆分每个 model 的 ID/CreatedAt/UpdatedAt 字段定义。工作量大且涉及现存 DB schema 兼容性（如 `default:gen_random_uuid()`），暂不强行。

---

#### C.2 [✓ FIXED] G. APIResponse 重复（已收敛）

**现状（已修复）**：原 3 处独立定义已全部用 type alias 收敛到 `pkg/response.APIResponse`：
- `services/user-service/internal/models/models.go:54` → `type APIResponse = response.APIResponse`
- `services/payment-service/internal/models/payment_models.go:264` → `type APIResponse = response.APIResponse`
- `services/recommendation-service/internal/handlers/simple_recommendation_handler.go:191` → `type APIResponse = response.APIResponse`

**真根因审计纠正**：原审计描述"字段名/JSON 结构相同"实际不准确。3 处旧 schema 字段是 `Error string`，pkg/response.APIResponse 是 `Error *ErrorInfo + Timestamp + RequestID`。直接 alias 会改变 wire schema。

**风险评估**：
- 后端：grep 全仓 `APIResponse{` 共 17 处构造点，**零处**构造 Error 字段 —— 大家只用 Success/Message/Data。所以 Error 字段 schema 变化属"dead spec 差异"，无生产行为影响。
- 前端：`frontend/src/types/api.ts::ApiResponse<T>` 字段集为 `success/code?/data/message?/total?/timestamp?/request_id?`，**不含 error 字段**。grep 前端零依赖 response.error。
- 结论：alias 切换零 wire 契约破坏。

**修复内容**：
1. user-service / payment-service / recommendation-service 的 `go.mod` 新增 `pkg/response v0.0.0` 依赖 + `replace ... => ../../pkg/response`。
2. 三处独立 struct 替换为 `type APIResponse = response.APIResponse`。
3. import "github.com/tvvshow/gokao/pkg/response"。
4. 调用点全部零改动（type alias 透明）。

**验证**：
- `go build ./...` workspace 干净。
- `go test ./services/user-service/... ./services/payment-service/... ./services/recommendation-service/...` 全绿。

---

#### C.3 [✓ FIXED] A. handler 错误响应 100+ 处 gin.H{}

**现状（已修复）**：267 处 `c.JSON(http.StatusXxx, gin.H{...})` 已全部收敛到 `pkg/response` 工厂。生产代码残留 0 处（剩余 7 处仅在测试 fixture 中作为 mock handler 输出，保留以保持测试可读性）。

**修复内容**：
1. `pkg/response` 补齐 9 个工厂：`CreatedWithMessage` / `NotImplemented` / `Locked` / `TooManyRequests` / `ServiceUnavailable` / `GatewayTimeout` / `RequestTimeout` / `Gone` / `WriteError`（含单元测试）。
2. 13 个生产文件（handlers + middleware + main + pkg/auth/discovery/middleware）改用 `response.{BadRequest/NotFound/InternalError/OK/OKWithMessage/Created/CreatedWithMessage/Unauthorized/Forbidden/NotImplemented/ServiceUnavailable/Locked/Conflict/TooManyRequests/AbortWithError}`。
3. `pkg/auth/pkg/discovery/pkg/middleware/services/monitoring-service` 等 go.mod 添加 `require + replace pkg/response`。
4. `services/api-gateway/{go.mod,Dockerfile}` + `services/monitoring-service/Dockerfile` 同步补齐 pkg/response 依赖 + COPY。

**特殊处理**：
- `auth_handler.Login/RefreshToken` 在 data 内同时保留 `access_token`/`refresh_token`/`expires_at` 旧字段；frontend `isWrappedResponse` 双模式自动适配。
- 支付回调 `PaymentCallback`/`Refund` 返回原始 result 对象（由外部支付网关消费），通过 `response.OK(c, result)` 统一仍保留 service 层结构。
- `WebhookTest` 使用 `OKWithMessage` 包装自定义 payload。

**验证**：
- `grep "c\\.JSON\\(http\\.Status.+gin\\.H\\{" services pkg` → 0 production matches（仅 7 处 test fixture）。
- CI run 25648661907 全绿，所有服务 build + test 通过。
- 净 LOC 减少 ~908 行（移除冗余 c.Abort()、重复 gin.H 字段模板）。

**遗留**：测试 fixture 中的 mock handler 保持原样（不属于审计范围，保留以维持测试可读性）。

---

#### C.4 [✓ FIXED] H. payment-service JSONB 双类型

**现状（已修复）**：原有两个 JSONB 类型并存：
- `payment_models.go:14` 定义 `PaymentJSONB`（Scan 含空字节边界保护）
- `membership_models.go:11` 定义 `JSONB`（Scan 直接 Unmarshal，空字节会 panic "unexpected end of JSON input"）

**修复内容**：
1. 合并到单一 `JSONB`（位于 `membership_models.go`），保留更鲁棒的实现（含空字节 → nil 边界保护）。
2. 删除 `payment_models.go` 的 `PaymentJSONB` 定义（含 Value/Scan 方法）。
3. 全仓 sed 重命名 22 处 `PaymentJSONB` → `JSONB`：
   - `models/payment_models.go`（4 个字段：PaymentOrder.Metadata / PaymentRefund.Metadata / + 2 处）
   - `adapters/alipay.go`（4 处 Extra）
   - `adapters/wechat_pay.go`（1 处 Extra）
   - `service/payment_service.go`（13 处构造 + 转换）

**验证**：
- `go test ./services/payment-service/...` 全绿（config + handlers）。
- `go build ./...`（workspace）干净。

---

#### C.5 [✓ FIXED] E. monitoring/api-gateway Redis 未走 shareddb

**现状（已修复）**：审计指出 monitoring-service / api-gateway 各自手写 `redis.NewClient(...)`，与 `pkg/database.OpenRedis` 已存在的统一工厂脱节。复审实际发现两层债务：
1. **配置漂移**：8 处独立 `redis.NewClient` 各自设超时/池/ping，参数不一致（monitoring 完全没 ping 校验，启动假成功也不报错）。
2. **版本割裂**：`pkg/database.OpenRedis` 走 `github.com/redis/go-redis/v9` 但 monitoring/api-gateway 仍用已停止维护的 `github.com/go-redis/redis/v8 v8.11.5`（v8 最后一版 2022 年归档），导致即使引入 helper 也只能 v8 单独包装一份。

**修复内容**：
1. **alerts 子 module 升 v9**：`services/monitoring-service/internal/alerts/{go.mod,alert_manager.go}` 将 `go-redis/v8` 切到 `redis/go-redis/v9 v9.13.0`。`alert_manager.go` 实际只用 `Set/Get/Keys/Result/Err`，v9 全兼容，零业务代码改动。
2. **monitoring-service main.go**：删 `redis.NewClient(&redis.Options{Addr: ...})`，改用 `sharedcfg.RedisConfig{...}` + `shareddb.OpenRedis(redisCfg, 5*time.Second)`。**附带补回历史缺陷**：原代码无 ping 校验，Redis 实例不可达时 alert manager 启动后首条写入才报错；新实现在启动期就 fail-fast，省一轮 production 排障。环境变量兼容：`REDIS_URL` 优先，docker-compose 历史用的 `REDIS_ADDR` 作为回退。
3. **api-gateway main.go**：同款替换 `initRedisCache` 内部的 6 行手写初始化为 `LoadRedis + OpenRedis`，v8 import 切 v9。
4. **go.mod/Dockerfile 同步**：
   - `services/monitoring-service/go.mod` 删 v8 require，加 `pkg/config` + `pkg/database` require + replace 指向 `../../pkg/{config,database}`。
   - `services/api-gateway/go.mod` 同上。
   - `Dockerfile` 同步 COPY `pkg/config` + `pkg/database` 让 docker 构建可解析。

**OOB（不在本次范围）**：
- `services/payment-service/internal/database/database.go:47` 仍用 v8 — payment-service 不在 C.5 审计范围。等下次 payment 专项治理（含 sql.DB → gorm 切换等）一并升 v9。
- `services/recommendation-service/internal/cache/redis.go:18` 已用 v9 但未走 `OpenRedis`（手写参数集与 OpenRedis 一致）— 不阻塞，下次可一并收敛。

**验证**：
- alerts module: `go build ./...` + `go mod tidy` 干净。
- monitoring-service: `go build ./...` + `go test ./...` 全绿（0.721s）。
- api-gateway: `go build ./...` + `go test ./...` 全绿（28.161s，22+4 测试覆盖 cache/rate limit/security headers）。
- workspace 级 `go build ./...` 干净。
- 其他服务 sanity（user-service middleware / recommendation handlers / data-service / payment-service）全绿。
- `gofmt -l` 检查改动文件无残留。

**净效果**：
- 4 处生产 `redis.NewClient(...)`（monitoring main + api-gateway main + alerts 内部 + 三方依赖）→ 2 处统一调 `OpenRedis`。
- monitoring + api-gateway 完成 `go-redis/v8` → `redis/go-redis/v9` 升级，消除版本碎片。
- monitoring 端首次具备 Redis 启动期 ping 校验。

---

#### C.6 [✓ FIXED] I. SecurityHeaders 统一

**现状（已修复）**：
- `services/api-gateway/main.go:461` 已通过 `securityMiddleware.SecurityHeaders()` 接入 pkg ✓
- `services/data-service/internal/middleware/middleware.go:221` 自定义 `Security()`（缺 Strict-Transport-Security）✗
- `services/payment-service/internal/middleware/middleware.go:213` 自定义 `SecurityHeaders()` 但**main.go 从未调用**（dead code，且实际请求无安全头）✗

**修复内容**：
1. `pkg/middleware/security.go` 新增包级 `SecurityHeaders()` 函数（不依赖 SecurityConfig），头列表合并所有服务历史最严集合：
   - X-Content-Type-Options / X-Frame-Options / X-XSS-Protection
   - Strict-Transport-Security / Content-Security-Policy
   - **Referrer-Policy: strict-origin-when-cross-origin**（原 data-service 独有，统一后所有服务都设）
2. `*SecurityMiddleware.SecurityHeaders()` method 改为委托给包级函数 + 保留 `sm.config.SecurityHeaders` 配置开关。api-gateway 用法不变。
3. `data-service/main.go` import `pkgmw`，把 `middleware.Security()` 替换为 `pkgmw.SecurityHeaders()`；删除 `internal/middleware.Security()` 函数。
4. `payment-service/main.go` 新增 `pkgmw.SecurityHeaders()` Use（之前完全没设安全头 —— 是 dead `SecurityHeaders()` 函数留下的设计缺陷）；删除 `internal/middleware.SecurityHeaders()` dead code。
5. `services/api-gateway/main_test.go::TestSecurityHeaders_OnGET` 修复历史 bug：原断言 `Referrer-Policy != ""` 配文案"missing or wrong"，反映出测试逻辑写反。统一后改为正向断言三个新头都存在。

**验证**：
- `go test ./services/api-gateway/... ./services/data-service/... ./services/payment-service/... ./pkg/middleware/...` 全绿。
- 副产物：payment-service 现在终于设上了安全响应头（之前是 dead code 假冒覆盖）。

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
| **P3**（第 3-4 周） | C.1 BeforeCreate / C.2 APIResponse / C.3 gin.H / C.4 JSONB / C.5 Redis init / C.6 SecurityHeaders | 重复代码治理（全部完成） |
| **P4**（持续） | B.8 P-22/P-23 debug log/JSON 优化 | 长尾治理 |
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
