# Project Structure

## Directory Organization

```
project-root/
├── docs/                               # 文档（架构设计、API、部署、合规）
│   ├── architecture/                   # 架构与决策记录（ADR）
│   ├── api/                            # OpenAPI 规范与示例
│   ├── operations/                     # 运维手册、SOP与应急预案
│   └── security/                       # 威胁模型、渗透测试报告、加固项
├── deployments/                        # 部署与交付
│   ├── docker/                         # Dockerfile（Go/C++/前端）
│   ├── k8s/                            # Kubernetes YAML（Helm Chart 可选）
│   └── scripts/                        # CI/CD脚本、迁移脚本
├── frontend/                           # 前端（Vue 3 + Vite + Element Plus）
│   ├── src/
│   │   ├── assets/
│   │   ├── components/
│   │   ├── pages/
│   │   ├── store/                      # Pinia 状态
│   │   ├── router/
│   │   └── services/                   # API 调用封装
│   ├── public/
│   └── vite.config.ts
├── services/                           # 后端微服务（Go）
│   ├── api-gateway/                    # 统一入口（Gin/Echo）
│   ├── user-service/                   # 认证、账号、权限、会员
│   ├── data-service/                   # 院校/专业/批次/历史数据管理
│   ├── match-service/                  # 调用 C++ 算法服务的编排层
│   ├── payment-service/                # 订单、支付、退款、对账
│   ├── notify-service/                 # 短信/邮件/站内通知
│   ├── report-service/                 # 分析报表与导出
│   └── admin-service/                  # 后台管理接口
├── cpp/                                # C++ 核心模块
│   ├── core-algo/                      # 志愿匹配、约束求解、排序
│   ├── ai-inference/                   # ONNX Runtime 推理
│   ├── license/                        # 许可验证、硬件绑定、反调试
│   ├── security/                       # 代码完整性校验、反篡改
│   ├── include/                        # 公共头文件
│   ├── third_party/                    # 第三方依赖（Conan/源码）
│   └── CMakeLists.txt                  # 根构建配置
├── shared/                             # 跨语言共享资产
│   ├── proto/                          # gRPC/Protobuf 定义
│   ├── schemas/                        # JSON Schema / SQL / 迁移
│   ├── models/                         # 共享实体定义与DTO
│   └── configs/                        # 环境配置模板与约定
├── tools/                              # 工具与开发者生产力
│   ├── linters/
│   ├── hooks/                          # Git hooks（预提交、提交消息）
│   └── benchmarks/                     # 基准测试、压测脚本
├── tests/                              # 测试
│   ├── go/                             # Go 单元/集成测试
│   ├── cpp/                            # C++ 单元/性能测试（GTest）
│   ├── e2e/                            # 端到端测试（Playwright/自研）
│   └── testdata/
├── Makefile                            # 统一构建入口（go/cpp/frontend/docker）
├── README.md
├── LICENSE
└── .editorconfig
```

## Naming Conventions

### Files
- Components/Modules: PascalCase（前端组件），snake_case（C++），kebab-case（配置/脚本）
- Services/Handlers: UserService, DataService（Go以业务名+Service）
- Utilities/Helpers: util_*（Go/C++工具库，新增标注 // NEW: 原因）
- Tests: [filename]_test.go / [filename]_test.cpp / *.spec.ts

### Code
- Classes/Types: PascalCase（C++类、Go结构体）
- Functions/Methods: camelCase（Go/C++），前端组合式API使用驼峰
- Constants: UPPER_SNAKE_CASE（跨语言一致）
- Variables: camelCase（Go/TS），snake_case（C++局部/私有）

## Import Patterns

### Import Order
1. External dependencies
2. Internal modules
3. Relative imports
4. Style imports（前端）

### Module/Package Organization
- Go 使用 module 内绝对导入（services/*），内部相对路径最小化
- C++ 采用 include/ 与命名空间划分，公共头放入 cpp/include
- 前端使用 @ 别名指向 src，模块内相对导入
- 依赖管理：Go Modules / Conan / npm + lockfile 固定版本

## Code Structure Patterns

### Module/Class Organization
1. Imports/includes/dependencies
2. Constants and configuration
3. Type/interface definitions
4. Main implementation
5. Helper/utility functions
6. Exports/public API

### Function/Method Organization
- 输入校验优先
- 核心逻辑居中
- 错误处理贯穿
- 明确的返回路径

### File Organization Principles
- 单一职责（one module per file 尽量）
- 相关功能分组
- Public API 明确
- 细节隐藏（internal/、匿名命名空间）

## Code Organization Principles
1. Single Responsibility
2. Modularity
3. Testability
4. Consistency

## Module Boundaries
- Core（C++算法、AI、许可）与 Services（Go微服务）通过 gRPC 隔离
- Public API（shared/proto, shared/models）对外，内部细节隐藏在 cpp/* 与 services/* 内部
- Platform-specific（Windows/macOS 客户端）与 Cross-platform（服务端）分离
- Stable（已商业化功能）与 Experimental（实验特性）通过 feature flags 管控
- 依赖方向：前端 → API → 服务编排 → C++ 核心；禁止反向依赖

## Code Size Guidelines
- File size: < 500 行（例外需说明）
- Function size: < 50 行（复杂逻辑拆函数）
- Complexity: 圈复杂度 < 10（CI 检查）
- Nesting depth: ≤ 3 层

## Dashboard/Monitoring Structure
```
src/
└── dashboard/
    ├── server/           # Go微服务导出metrics + Prometheus抓取
    ├── client/           # Vue看板（ECharts/Chart.js）
    ├── shared/
    └── public/
```
- 与核心业务解耦，可独立关闭；独立路由与端口；尽量只依赖公共API

## Documentation Standards
- 所有公共API需在 docs/api 中提供 OpenAPI/示例
- 复杂算法、许可、安全模块强制在 docs/security 与 docs/architecture 留存设计说明
- README：每个 service 与 cpp 子模块均应有
- 遵循 GoDoc / Doxygen / Typedoc 约定
- 运维/灾备文档：
  - docs/operations/ops-guide.md — 运维SOP与告警处理流程
  - docs/operations/dr-runbook.md — 灾备/恢复（含RTO/RPO演练步骤与演练记录）
  - docs/security/key-rotation.md — 密钥与Secrets轮换策略（KMS/HSM，年度/季度）