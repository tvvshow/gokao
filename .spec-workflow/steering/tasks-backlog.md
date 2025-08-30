# Tasks Backlog — GaokaoHub

说明：以下任务从 execution-plan.md、tech.md、structure.md 萃取，面向 18 周周期，便于导入任务看板（可转 CSV）。
字段：模块 | 任务 | 明确产出 | 负责人 | 依赖 | 估时 | 验收标准(DoD)

## W1-W2 环境与基座
- DevOps | 初始化Git仓、分支保护、Commitlint/Conventional Commits | 仓库+保护规则 | DevOps | 无 | 4h | 规则生效且示例提交通过
- DevOps | CI模版（Go/C++/FE）+ SBOM 输出 | .github/ 或 GitLab CI | DevOps | 仓库 | 1d | 三语言能跑+产物上传
- Infra | K8s本地/云集群、镜像仓库、制品库 | 集群可用 | SRE | CI | 2d | Helm安装通过、私有仓拉取OK
- Security | 机密管理（KMS/Secrets）引导 | 文档+示例 | Security | CI | 0.5d | 本地/CI均可注入机密
- Observability | Prometheus/Loki/Tempo/Grafana | 监控栈可用 | SRE | Infra | 1.5d | 仪表盘有核心指标

## W3-W5 用户与数据
- User | 认证/注册/权限（RBAC） | user-service MVP | BE1 | Infra | 3d | OpenAPI+单测通过
- Data | 数据模型与导入流水线 | schema+迁移脚本 | BE2 | DB | 3d | 10w+样本导入成功
- API | Gateway 路由/限流/审计 | api-gateway MVP | BE3 | User | 2d | P99<120ms，错误<0.1%

## W4-W6 C++核心与编排
- C++ | core-algo 初版（场景约束求解） | lib+GTest | Cpp1 | Data | 5d | 准确率基线>95%
- C++ | license/anti-tamper | 静态库+示例 | Cpp2 | - | 3d | 许可绑定生效
- Go | match-service 调用C++ | gRPC编排 | BE1 | C++ | 2d | P99<200ms

## W6-W8 功能完善与前端
- FE | 高考志愿填报流程UI | Vue页面 | FE1 | API | 4d | 关键路径可用
- BE | 报表/导出 | report-service | BE2 | Data | 2d | 生成PDF/CSV

## W8-W10 支付/会员
- BE | payment-service 接入微信/支付宝 | 支付回调+签名 | BE3 | Infra | 4d | 沙盒支付通过
- BE | 会员/权益 | user-service 扩展 | BE1 | Payment | 2d | 权限控制生效

## W10-W12 质量与安全
- QA | 覆盖率提升至门禁目标 | 覆盖率报告 | QA | 各服务 | 4d | Go≥70/80, C++≥60/75, FE≥60
- Sec | SCA/镜像/IaC扫描零高危 | 报告+修复 | Security | CI | 2d | 高危=0
- Perf | Flamegraph瓶颈优化 | 报告+修复PR | BE/C++ | 监控 | 3d | 目标SLO达标

## W12-W14 灰度与SLO
- SRE | 灰度发布策略与回滚 | 文档+脚本 | SRE | CI | 2d | 一键灰度/回滚成功
- QA | 可用性与容量压测 | 报告 | QA | SRE | 3d | 错误预算在阈内

## W15 观测与审计
- Obs | 仪表盘完善（SLA/SLO） | Grafana面板 | SRE | 监控 | 1d | 看板覆盖三大服务
- Audit | 审计流水与脱敏 | 日志策略 | Security | User | 1d | PII最小化合规

## W16-W18 上线与复盘
- DR | 备份/恢复演练（RTO<1h, RPO<15m） | 演练记录 | SRE | Data | 1d | 佐证材料通过
- Rel | 正式发布与监控 | 发布记录 | All | 灰度 | 1d | 稳定运行1周
- Retro | 复盘与路线图 | 报告 | PM | 全体 | 0.5d | 行动项明确

## 持续任务（每周）
- 安全依赖升级、许可证合规扫描、关键模块变异测试（月度）

备注：桌面端（Win/macOS）归入 Phase 2，不在本周期交付范围。