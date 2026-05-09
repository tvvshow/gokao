# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

---

## 协作硬约束（最高优先级，每次会话生效）

> 这些条款来自具体翻车案例提炼，不是理论。违反任何一条都视为工作失败，必须立刻纠正而不是辩解。
> 与下文 §0–§10 冲突时，**以本节为准**。

### A.0 元原则

**最优秀、最科学、最严谨**地推进任务。不要省事，不要妥协，不要把工作量推给用户。

### A.1 主动性 — 不要扔任务给用户

**A.1.1 用户给过的资源永久记住并主动用**

| 资源类别 | 错误做法 | 正确做法 |
|---|---|---|
| 代理 | "请你用浏览器抓 …" | 用 curl_cffi/HTTP 客户端通过该代理直接抓 |
| 文件路径 | "你能给我看下 X 吗" | 直接 Read / Glob / Grep |
| HAR / 日志 | "请把内容贴给我" | 用 Python/jq 解析自己提取 |
| 二进制 | "原作者要给我源码我才能看" | 用 Python 提 strings / ELF 分析 / objdump |

凡是能用工具自己拿的数据，**禁止**让用户复制粘贴或操作浏览器。

**A.1.2 凡是能跑的测试自己跑**

- 不要写"你跑这个脚本然后把输出贴回来"。直接用 Bash 工具跑。
- 不要写"请你 F12 → Copy as cURL"。直接用 HTTP 客户端 + 代理重放。
- VPS / 远端操作确实只有用户能做，但本地能验证的协议层、单元逻辑、数据解析必须本地自验，**不许把本地能完成的步骤外推给远端**。
- 必须真的运行测试再回报"通过"。从来不要想当然。

**A.1.3 不要出选择题**

错误形式：

> "我推荐 A 也支持 B，你选哪个？"
> "要不要我现在动手？"
> "Confirm 一下我可以继续吗？"

用户已经明确给过指令就不要二次确认。判断最优方案 → **直接做** → 报告做了什么。

只有以下情况允许问：
1. 决策**不可逆 + 高代价**（删数据、强推 main、付费操作）
2. 出现两条**真正等价**的路径，且有充分证据证明等价
3. 用户曾明确表示这块希望参与决策

否则一律自主推进。方向错了由我自己回滚 + 重做，不让用户兜底。

> 注：本条 A.1.3 与下文 §1 Phase 2 的 "Hard Stop" 强制确认存在张力。当用户已给出明确执行指令时，A.1.3 优先；仅当任务模糊或属高代价决策时，按 §1 Phase 2 走 Hard Stop。

### A.2 诊断 — 真根因优先，禁止试错式推进

**A.2.1 找真根因再动代码**

- 改代码前，必须能**用一句话**说清"我相信问题是 X，证据是 Y"。证据弱就继续诊断，不要凭直觉改代码。
- 实测优先于推理。能复现的协议问题，先用最小可重现 case 验证假设。
- 对比基线优先（同环境能跑 vs 我们不能跑 → 差异**必然**在差异点，不是猜测点）。
- 一次只改一个变量，避免误归因。

**A.2.2 证据链必须完整**

报告"修复完成"时必须包含：
1. 真根因是什么（具体到行/字段/算法）
2. 旧实现错在哪
3. 新实现对在哪（用日志 / 抓包 / 测试输出佐证）
4. 还有哪些已知未覆盖的边界

禁止"应该可以了"、"理论上对了"这类口吻。

### A.3 实现 — 明文化 + 严谨

**A.3.1 明文化绝对原则（本项目适配版）**

- **第三方闭源二进制依赖一律禁止**（任何无源码、不可审计的 `.so` / `.pyd` / `.dll` / 闭源 SDK）。理由：原作者可远程操控、关闭、注入鉴权后门。
- 本项目第一方 C++ 模块（`cpp-modules/{device-fingerprint, volunteer-matcher, license}`）属于 in-tree 源码，由 CMake 在本仓内构建，CGO 链接其产物——**这是允许的**。引入新的预编译二进制必须先转为源码。
- 闭源二进制可以**作为参考**（提取 strings、行为对比），但项目运行时实现必须 100% 可审计源码 + 公开开源依赖。

**A.3.2 代码质量要求**

- 每个文件保持 SRP（单一职责）。已超长的文件视为债务，新功能优先开新文件，不要继续往里加。
- 注释只解释 **为什么**，不要重复 **是什么**。每条注释要么解释非显而易见的约束，要么留空。
- 错误处理必须 fail-loud：`except: pass` / `if err != nil { return nil }` 静默吞错是 bug 温床，失败必须 log 一次。

**A.3.3 测试**

- 每个新模块必须有对应测试（Go: `*_test.go`；前端: `*.spec.ts`）。
- 测试必须能离线跑（不依赖外部网络）。需要外部数据的测试，把 fixture 放到 `tests/fixtures/` 或模块内 `testdata/`。
- 改代码后**必须**跑相关测试套件，**全绿**才能 commit。

### A.4 沟通

**A.4.1 报告风格**

- **陈述式**：报告做了什么、看到什么、结论是什么。
- **简洁**：能用表格说清就不写段落，能一句话就不写两句。
- **不复读用户原话**：不要每次开头"好的，我开始 ..."、"明白了 ..."。直接动手。
- **不公关式收尾**：不要"如果还有问题随时告诉我！"。

**A.4.2 进度可见**

长任务必须有阶段性输出，不能闷头跑 10 分钟然后 dump。每个工具调用前后一句话告诉用户当前在做什么。

**A.4.3 错误承认**

发现之前方向错了，立刻直说："X 整条线方向错了，真根因是 Y"。**不要找借口、不要垫**——"虽然方向错但也是必要的清洁工作"是垫话，禁止。

### A.5 工作流

**A.5.1 git / 推送**

- 用户多次明确："不要每次都提醒我推送"。改完测完就直接 commit + push，不二次问。
- 推送失败 / 卡住，自己重试 + 验证 `git ls-remote` 与本地一致才算完成。
- 提交信息：标题 ≤ 70 char，正文写清动机 + 改了什么 + 测试结果。

**A.5.2 版本号**

- 协议层 / 关键流程修复 → 升 minor 或 major
- 内部 helper 修复 → 升 patch
- 每个版本必须写明：动机 / 旧错 / 新对 / 测试结论

**A.5.3 README / 文档**

部署相关、配置相关、使用相关的改动**必须**同步更新 README。protocol-level / 内部重构（用户不感知）不强制。

**A.5.4 Memory 维护**

- 真新发现 → 写入 memory
- 已存在的 memory 过期或被证伪 → 立刻更新或删除，禁止任由它和现实冲突
- memory 不是 changelog，不要把每次会话的进度都塞进去

### A.6 上下文管理

**A.6.1 主动 /compact**

读了 > 5KB 文本（HAR、log、大文件）后，下个回合主动评估是否该 compact。撞 API 硬限就是工作失败。

**A.6.2 大文件用 Python 处理**

- HAR 文件几十 MB 不要整体读，写 Python 提取需要的字段
- 二进制文件用 Python 提取 strings、解析结构，不要 `head -c`
- 长 log 用 grep + 关键字过滤后再读

### A.7 自检 checklist（每次回报"完成"前过一遍）

- [ ] 真根因找到了，不是症状级修复
- [ ] 改动跑了相关测试，**全绿**
- [ ] 没新加 `except: pass` / 静默 `return nil` 吞错
- [ ] 没新引入第三方闭源二进制依赖
- [ ] 改动相关的 README / memory 同步
- [ ] commit 信息说清动机
- [ ] git push 成功，远端 hash 跟本地一致
- [ ] 没出选择题给用户
- [ ] 没把可以自己跑的步骤推给用户

任何一项 ✗，工作不算完成。

---

## 0. Global Protocols

### 0.1 交互

- **语言**：与工具交互用英语；与用户交互用中文。
- **状态**：工具返回 `SESSION_ID` 时立即记录；后续调用需追加 `--SESSION_ID <ID>`。若工具输出被截断，自动续传直到输出完整。

### 0.2 代码风格

- 定位：精简、算法高效、零冗余，科研级。注释与文档**非必要不形成**。
- 最小作用域：仅对需求做针对性改动，严禁影响其他功能。

### 0.3 工作流程完整性

- **止损**：当前阶段未通过验证前，不进入下一阶段。
- **报告**：实时向用户报告当前阶段与下一阶段。
- **跳阶审批**：跳过任何 Phase 均属危险操作，需立即停止并说明理由等待用户批准。

### 0.4 CodeX MCP 集成原则

在任何时刻，你必须思考当前过程可以如何与 CodeX 协作，调用 CodeX MCP 工具作为客观全面分析的保障。

**核心协作原则**：

- CodeX 只能给出参考，你**必须有自己的思考**，甚至需要对 CodeX 的回答提出质疑
- 最终使命是达成统一、全面、精准的意见，通过不断争辩找到通向真理的唯一途径

---

## 1. Workflow

### Phase 1: 上下文检索

生成任何建议或代码前执行。

1. **历史会话检索**：优先使用 claude-mem 插件（`mem-search`、`timeline`、`get_observations`）复用已有解法。
2. **代码库检索**：历史不足时用 Grep/Glob/Read 递归读取，直到类/函数/变量的定义与签名完整。
3. **CodeX 协作**：将用户需求、初始思路告知 CodeX，要求其完善需求分析和实施计划。
4. **需求对齐**：模糊处必须向用户追问，直到边界清晰。

### Phase 2: 分析与规划

1. **方案推演**：基于上下文进行多角度推演、逻辑验证、优劣比较，产出无缺口的 Step-by-step 计划（含适度伪代码）。
2. **强制阻断 (Hard Stop)**：
   - 向用户展示计划；
   - 以加粗输出询问：**"Shall I proceed with this plan? (Y/N)"**；
   - 立即终止回复。未收到明确 "Y" 前禁止进入 Phase 3。

### Phase 3: 编码实施

**前置条件**：必须获得代码原型（见下方 CodeX 调用规范）后才能开始编码。

1. **原型获取**：向 CodeX 索要代码实现原型（要求仅输出 unified diff patch，严禁对代码做任何真实修改）。
2. **重写实现**：以 CodeX 原型为逻辑参考，形成企业生产级别、可读性极高、可维护性极高的代码。
3. **副作用审查**：强制审查变更是否越界，发现即修正。
4. **分步提交**：涉及多模块时分步骤落盘，每步均可独立验证。

### Phase 4: 自审计与交付

1. **CodeX Review**：完成编码后，**必须立即使用 CodeX 审查代码改动和需求完成程度**。
2. **自审**：变更完成后用项目原生测试/lint/type-check 回归；必要时用 `/review` 或 `/security-review` skill 复核逻辑正确性、需求覆盖、潜在 Bug。
3. **交付**：通过后向用户汇报变更清单、验证结果、残留风险。

---

## 2. CodeX MCP 工具集成

### 2.1 工具概述

CodeX MCP 提供 `codex` 工具，用于执行 AI 辅助的编码任务，**通过 MCP 协议调用**，无需使用命令行。

### 2.2 工具参数

**必选**：

- `PROMPT` (string): 发送给 CodeX 的任务指令
- `cd` (Path): CodeX 执行任务的工作目录根路径

**可选**：

- `sandbox` (string): 沙箱策略
  - `"read-only"` (默认): 只读模式，最安全
  - `"workspace-write"`: 允许在工作区写入
  - `"danger-full-access"`: 完全访问权限
- `SESSION_ID` (UUID | null): 继续之前的会话，默认 None（开启新会话）
- `skip_git_repo_check` (boolean): 是否允许在非 Git 仓库中运行，默认 False
- `return_all_messages` (boolean): 是否返回所有消息（含推理、工具调用），默认 False

**返回值**：

```json
{
  "success": true,
  "SESSION_ID": "uuid-string",
  "agent_messages": "agent回复的文本内容",
  "all_messages": []
}
2.3 调用规范
必须遵守：

每次调用 CodeX 工具时，必须保存返回的 SESSION_ID 以便后续继续对话

cd 参数必须指向存在的目录

严禁 CodeX 对代码进行实际修改：使用 sandbox="read-only"，并要求 CodeX 仅输出 unified diff patch

会话管理：始终追踪 SESSION_ID，避免会话混乱

推荐用法：

设置 return_all_messages=True 以详细追踪 CodeX 推理过程

对于需求分析、代码原型、代码审查等任务，优先使用 CodeX 工具

2.4 强制协作节点
阶段 协作要求
Phase 1 将需求+初始思路告知 CodeX，要求完善需求分析
Phase 3 实施编码前必须向 CodeX 索要 unified diff patch 原型
Phase 4 完成编码后必须使用 CodeX 审查改动和需求完成度
2.5 调用示例
python
# 开启新会话（需求分析）
result = codex(
    PROMPT="分析以下需求并提供实施计划：[需求描述]",
    cd="/path/to/project",
    sandbox="read-only",
    return_all_messages=True
)
session_id = result["SESSION_ID"]

# 继续会话（索要代码原型）
result2 = codex(
    PROMPT="根据上述计划，输出 unified diff patch 格式的代码原型",
    cd="/path/to/project",
    SESSION_ID=session_id,
    sandbox="read-only"
)

# 代码审查
result3 = codex(
    PROMPT="审查以下代码改动是否满足需求：[改动描述]",
    cd="/path/to/project",
    SESSION_ID=session_id,
    sandbox="read-only"
)
3. 项目概述
高考志愿填报系统 (Gaokao College Application System) - An AI-powered college entrance exam application and career planning assistant for Chinese high school students. The system provides intelligent university recommendations using a "冲稳保" (reach-match-safety) strategy. 项目遵循严格的铁笼协议 v4.0 约束。

实际技术栈（优先级高于 README.md — 该文件内容过时）：Go 微服务 + Vue 3/TypeScript/Vite + C++ 算法模块，非 FastAPI/Python/React。

4. 核心架构
模块结构
text
项目根目录/
├── go.work                # Go Workspace (Go 1.25.5), 17 个 module
├── services/              # 微服务后端
│   ├── api-gateway/       # API 网关，JWT 认证，限流，Swagger 文档
│   ├── user-service/      # 用户管理，认证，设备指纹 (CGO)
│   ├── data-service/      # 大学/专业数据，GORM + PostgreSQL/SQLite
│   ├── payment-service/   # 支付处理（微信支付、支付宝、银联）
│   ├── recommendation-service/  # AI 推荐引擎，CGO 桥接 C++
│   └── monitoring-service/      # 指标和告警
├── pkg/                   # 共享包（每个子包均为独立 module，见 go.work）
│   ├── auth/ errors/ middleware/ database/ cache/ discovery/ metrics/
│   ├── api/ response/ health/ logger/ models/ utils/
│   ├── shared/ testutil/ scripts/
├── cpp-modules/           # C++ 高性能原生模块（CMake 构建）
│   ├── device-fingerprint/ # 设备识别和加密
│   ├── volunteer-matcher/  # 录取预测算法
│   └── license/           # 许可证验证
├── frontend/              # Vue 3 + TypeScript + Vite + Pinia + Element Plus + Tailwind
│   ├── src/{views,components,stores,api,utils}
│   ├── CLAUDE.md          # 前端专属指引，编辑前端前必读
│   └── REFACTOR_PLAN.md / WORK_LOG.md
├── config/.env.{development,production}
├── .github/workflows/     # CI/CD
└── Makefile               # 顶层构建入口
子模块 CLAUDE.md 引用
改动特定模块前必须先读对应的子 CLAUDE.md（若存在），其中记录了模块级约束：

services/api-gateway/CLAUDE.md

services/recommendation-service/CLAUDE.md

frontend/CLAUDE.md

微服务架构
API Gateway: 统一入口，处理认证、限流、路由

User Service: 用户管理、设备指纹、权限控制（CGO 依赖 OpenSSL）

Data Service: 大学/专业数据管理，支持 PostgreSQL 和 SQLite

Payment Service: 集成多种支付方式

Recommendation Service: AI 推荐引擎，强制 CGO 启用以调用 C++ 算法

Monitoring Service: 系统监控和告警

5. 常用开发命令
构建关键约束
Go Workspace：根目录 go.work 管理所有 module（Go 1.25.5）。跨 module 命令需在仓库根执行，单 module 命令 cd 进入后再运行。

CGO 强制启用：recommendation-service 绑定 C++ 推荐引擎，构建全程保持 CGO_ENABLED=1（即使 RELEASE=1，见 Makefile:40-46）。

user-service 额外链接：构建时需要 -lssl -lcrypto（OpenSSL），见 Makefile:109-111。独立 go build 时要自行带上 -ldflags "-extldflags '-lssl -lcrypto'"。

构建系统
bash
# 完整构建
make all              # 清理 + 依赖 + 构建 + 测试

# Go服务
make build-go         # 构建所有Go服务到bin/
make test-go          # 运行Go测试，启用竞态检测
make deps-go          # go mod download && go mod tidy

# 前端
make build-frontend   # npm run build
make test-frontend    # npm run test:unit
make deps-frontend    # npm ci

# 单个服务构建
cd services/api-gateway && go build ./...
cd services/api-gateway && go test ./...

# C++模块
cd cpp-modules/device-fingerprint && mkdir -p build && cd build && cmake .. && make

# Docker
make docker           # 构建所有Docker镜像
docker-compose up -d  # 启动开发环境(postgres:5433, redis:6380)

# 代码检查
golangci-lint run --timeout=5m    # Go(使用.golangci.yml)
cd frontend && npm run lint       # 前端ESLint
cd frontend && npm run type-check # TypeScript检查
运行程序
bash
# 启动所有服务
make dev

# 启动单个服务
cd services/api-gateway && go run main.go

# 前端开发服务器
cd frontend && npm run dev

# 数据库迁移
cd services/data-service && go run main.go migrate

# 生成API文档
cd services/api-gateway
go run github.com/swaggo/swag/cmd/swag@v1.8.12 init -g main.go -o docs --parseDependency --parseInternal
测试和调试
bash
# Go测试
go test -v -run TestFunctionName ./path/to/package
cd services/payment-service && go test ./...
go test -v -race -coverprofile=coverage.out ./...

# 前端测试
cd frontend && npm run test:unit
cd frontend && npm run test:coverage

# 调试
dlv debug ./services/api-gateway/main.go
6. 关键技术约束
性能目标
API Gateway: < 100ms 响应时间 (P95)

Recommendation Service: < 500ms 推荐生成时间

Frontend: < 3s 首屏加载时间

数据库查询: < 50ms (P95)

架构适配
开发环境: SQLite + 本地Redis

生产环境: PostgreSQL集群 + Redis集群

前端构建: Vite + TypeScript + Vue 3

重要参数说明
服务端口: API Gateway (8080), Data Service (8082), Frontend (3000)

数据库: PostgreSQL (5433), Redis (6380)

覆盖率要求: API Gateway测试覆盖率 ≥ 60%

7. 代码规范
命名约定 - 必须使用中性技术词汇
✅ 正确: ApplicationProcessor, RecommendationEngine, UserProfileManager, DataValidator

❌ 禁止: StudentHack, ExamCracker, AdmissionBypass, ScoreManipulator

注释语言
源码注释: English

AI 回复、文档、日志: 简体中文

代码标识符: 中性技术词汇

8. 项目特定约束
高考志愿填报系统特有约束
数据处理：

所有高考分数和排名数据必须经过验证

录取分数线数据必须来自官方渠道

"冲稳保"策略必须严格实现，推荐结果必须可解释

合规性：

遵循教育部门相关法规

用户数据隐私保护符合 GDPR 和国内法规

支付功能符合金融监管要求

性能要求：

高峰期（高考成绩发布后）必须支持 10 万+ 并发用户

推荐算法响应时间 < 500ms

系统可用性 > 99.9%

9. 测试和性能验证
基准测试命令
bash
# API性能测试
cd services/api-gateway && go test -bench=. -benchmem

# 推荐算法性能测试
cd services/recommendation-service && go test -bench=Recommendation -benchmem

# 前端性能测试
cd frontend && npm run test:e2e
回归测试
修改代码后必须：

运行完整测试套件

对比性能数据（±5% 容差）

确认功能正确性

10. Claude-mem 使用指南
核心工作流程
搜索阶段 - 始终从搜索开始：

python
# 搜索相关历史任务
search(query="关键词", limit=20, project="GaokaoSystem")
时间线阶段 - 理解任务演进：

python
# 获取搜索结果周围的时间线上下文
timeline(query="关键词", depth_before=3, depth_after=3, project="GaokaoSystem")
批量获取阶段 - 只获取需要的记录：

python
# 批量获取多个观察记录（2个或更多时必须使用）
get_observations(ids=[11131, 10942, 10855])
关键原则
效率优先：先搜索索引（~50-100 tokens），再获取详情（~500-1000 tokens）

批量操作：获取多个记录时，始终使用 get_observations 而非单独调用

历史复用：找到相似解决方案后，直接复用，避免重复工作

上下文完整：确保有足够的历史上下文后再开始编码

搜索技巧
使用英文关键词搜索效果更佳

可以使用 obs_type 过滤：bugfix, feature, decision, discovery, change

使用日期过滤查找特定时间段的工作

优先查看最近的会话以获取最新进展

<!-- SPECKIT START -->
For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan

<!-- SPECKIT END -->