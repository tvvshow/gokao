# Execution Plan — GaokaoHub (Go + C++ Hybrid)

## 0. 概览
- 节奏：18 周，9 人团队（架构1、Go后端3、C++2、前端2、QA1，DevOps由架构/后端分担）
- 目标：上市 SaaS + 学校 B2B 版本；P99 200ms；QPS 峰值 50k；99.9% 可用性；首年 5M RMB 收入
- 架构基线：Go 70%（API/业务/数据），C++ 30%（算法/AI/许可/付费核心）
- 验收方式：每阶段 DoD + KPI，端到端可演示 + 日志/指标可见

## 1. 里程碑（18 周节奏）
- W1-W2 基座与环境：代码仓库、Makefile、Conan/Go modules、Docker 镜像、K8s Dev 集群、CI 门禁
- W3-W4 用户与鉴权：注册/登录、JWT+Refresh、RBAC、基础审计日志、OpenAPI 文档
- W5-W6 数据域模型：PostgreSQL 表结构、迁移、种子数据；Redis 缓存策略；Elasticsearch 索引
- W7-W8 核心匹配（C++ 引擎 v1）：约束建模、排序策略、接口封装（gRPC/FFI），基准对拍
- W9 AI 推荐（推理 v1）：ONNX Runtime 接入、离线特征流水线、在线召回/重排策略雏形
- W10 支付与订单：微信/支付宝沙箱、订单生命周期、对账、退款、风险控制策略
- W11 许可与防破解：VMProtect 加壳、硬件指纹、反调试、远程许可证校验；Go garble 混淆
- W12 安全加固：WAF/限流、IP信誉、敏感数据加密、Secrets 管理、依赖与镜像扫描
- W13-W14 联调与压测：混合负载压测（API/匹配/支付）、容量预测、瓶颈优化、降级/熔断
- W15 监控与可观测：Prometheus/Grafana、Tracing、审计看板、告警策略（SLA/SLO）
- W16 试点灰度：小范围学校/地区用户，灰度发布/回滚预案，埋点验证留存/转化
- W17 上线准备：Chaos 演练、备份/恢复演练、演练版蓝绿切换、运维SOP；RTO/RPO 验证（目标：RTO < 1h，RPO < 15m），形成佐证材料
- W18 正式发布与复盘：发布、稳定性监控、首周复盘与路线图更新；RTO/RPO 达标确认并纳入SLA

## 2. 服务与模块拆解（按仓内目录映射）
- services/api-gateway：API 路由、统一鉴权、中间件、SSE/WebSocket、限流
- services/user-service：注册/登录/JWT/Refresh、RBAC、会员等级、审计
- services/data-service：院校/专业/批次/历史数据、同步任务、ES 索引
- services/match-service：编排与策略，调用 C++ 引擎，缓存与灰度开关
- services/payment-service：订单、支付、退款、对账、发票接口
- services/notify-service：短信/邮件/站内信、模板/频控
- services/report-service：可下载报表、分析图表数据源
- services/admin-service：后台管理接口
- cpp/core-algo：约束求解/打分/排序、并行向量化
- cpp/ai-inference：ONNX Runtime 封装、模型加载/热更新
- cpp/license：许可校验、硬件指纹、反调试与环境探测
- cpp/security：完整性校验、反篡改检测
- frontend/*：登录/首页/志愿模拟/专业搜索/下单支付/个人中心/后台

## 3. 阶段 DoD 与 KPI
- 基座（W1-W2）
  - DoD：本地/Dev 环境一键启动（Docker Compose），CI 完成 lint/单测/镜像构建
  - KPI：主服务冷启动 < 30s；单测通过率 ≥ 95%
- 鉴权（W3-W4）
  - DoD：JWT+Refresh、RBAC、审计日志；OpenAPI 文档自动发布
  - KPI：登录 P99 < 150ms；暴力破解防护触发验证
- 数据（W5-W6）
  - DoD：PostgreSQL/Redis/ES 全量/增量流程；联调样本可查询
  - KPI：热门搜索 < 100ms；缓存命中 ≥ 85%
- 匹配引擎 v1（W7-W8）
  - DoD：gRPC 接口稳定，Go 服务稳定调用；对拍 50+ 样例通过
  - KPI：单次匹配 < 50ms（平均），P99 < 200ms
- AI 推荐 v1（W9）
  - DoD：推理接口/特征抽取可用；AB 开关
  - KPI：平均推理 < 30ms；召回率/命中率基线建立
- 支付（W10）
  - DoD：沙箱全流程、对账任务、风控拦截策略
  - KPI：成功率 ≥ 99.5%；对账差异 < 0.5%
- 许可（W11）
  - DoD：VMProtect 保护产物、远程校验、硬件绑定、异常告警
  - KPI：许可证校验 < 20ms；逃逸事件 0 起
- 安全（W12）
  - DoD：WAF/限流上线、密钥安全、依赖漏洞 0 高危
  - KPI：误报率 < 1%；阻断成功率 > 95%
- 压测（W13-W14）
  - DoD：混合压测脚本、瓶颈清单与修复、降级熔断演练
  - KPI：QPS 50k 峰值稳定，P99 < 200ms
- 可观测（W15）
  - DoD：关键指标/日志/Trace 可视；告警闭环（分级）
  - KPI：告警到响应 < 5 分钟；误报率 < 5%
- 灰度（W16）
  - DoD：分批发布/回滚、埋点闭环
  - KPI：Crash 率 < 0.1%；转化/留存指标达标
- 上线（W17-W18）
  - DoD：蓝绿切换/演练通过、SOP 完整；复盘材料沉淀
  - KPI：7 天 99.9% 可用，0 高危事故

## 4. 任务清单（样例与模板）
任务字段：模块 | 描述 | 输入/输出 | 依赖 | 负责人 | 预估(人日) | 验收 | 风险
- api-gateway | 接入 JWT + 限流 | OpenAPI→路由运行 | user-service | BE1 | 2d | 通过集成测试 | 高并发限流策略
- match-service | 编排 gRPC 调用 | gRPC proto→结果集 | core-algo | BE2 | 3d | 50 样例对拍 | 序列化开销
- core-algo | 约束求解器 v1 | 题库/约束→建议 | schemas | C++1 | 8d | 基准达标 | 数据边界
- ai-inference | ONNX 推理封装 | 模型→得分 | 模型仓库 | C++2 | 5d | 延迟 <30ms | 模型大小
- license | 硬件指纹+远程校验 | 许可→通过/拒绝 | 密钥服务 | C++1 | 4d | 逃逸=0 | 兼容性
- payment | 沙箱全链路 | 订单→支付态 | 网关 | BE3 | 5d | 回调对账OK | 异常对齐
- frontend | 志愿模拟页 | API→UI | api-gateway | FE1 | 5d | e2e 通过 | 交互复杂
- QA | 压测场景脚本 | N/A | 全链路 | QA | 4d | 指标达标 | 测试数据

附：提供 CSV/表格模版，便于导入看板工具。

## 5. CI/CD 与质量门禁
- Lint/格式化：golangci-lint, gofmt/goimports, clang-format, clang-tidy
- 测试矩阵：Go 单测+集成、C++ GTest、前端 Vitest/Playwright
- 覆盖率门禁：Go 单元覆盖率 ≥ 70%，关键包 ≥ 80%；C++ 单元覆盖率 ≥ 60%，核心算法/许可 ≥ 75%；前端单测 ≥ 60%
- 关键模块质量：对 cpp/license 与 cpp/core-algo 执行变异测试/差错注入（季度）
- 安全扫描：SCA（依赖）、容器镜像扫描、机密扫描、IaC 扫描
- 版本与分支：Git Flow；语义化版本；Release Notes 自动化
- 构建产物：多阶段 Docker；私有镜像仓库；SBOM 输出

## 6. 性能与容量计划
- SLO预算（P99延迟目标 / 错误预算）：
  - API Gateway：P99 120ms，错误率 < 0.1%
  - Match Service（含C++调用）：P99 200ms，错误率 < 0.2%
  - Payment Service：P99 150ms，错误率 < 0.1%
- 基准：核心路径 Flamegraph；热函数优化（并行/向量化/缓存）
- 容量：HPA 自动扩缩容；缓存分层；读写分离；降级熔断与限流
- 数据：ES 倒排索引调优；Redis 热点分片；PostgreSQL 索引策略

附：桌面端（Windows/macOS）不纳入首发发布范围，纳入 Phase 2（如需插入W17+包装/签名/自动更新，可在W16灰度后另起里程碑）。

## 7. 风险与回滚
- 风险：跨语言调用开销、C++ 构建复杂度、支付合规、模型偏差
- 预案：共享内存优化/批处理；Docker 标准化构建；对账/风控双轨；AB/灰度与回滚

## 8. 资源与预算（估算）
- 云资源：K8s 3 节点（8C16G 起）、RDS/Redis/ES 标准规格、对象存储
- 第三方：短信/邮件、支付费率、VMProtect 许可、监控与日志托管
- 成本控制：低环境弹性关闭、预留实例、按需扩容

---
本执行计划与以下文档一致：
- <product> .spec-workflow/steering/product.md
- <tech> .spec-workflow/steering/tech.md
- <structure> .spec-workflow/steering/structure.md