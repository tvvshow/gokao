# Feature Specification: Shared Module Unification

**Feature Branch**: `001-unify-pkg-modules`
**Created**: 2026-04-24
**Status**: In Progress
**Input**: User description: "本次仅涉及：1. api-gateway/go.mod 删除 replace 指令 2. payment-service/go.mod 删除 replace 指令 3. 运行 go mod tidy 同步依赖 4. 验证构建和基本功能 不修改其他服务的 go.mod"

## Clarifications

### Session 2026-04-24

- Q: 本次变更范围是否仅限两个服务并禁止扩展到其他服务？ → A: 是，仅修改 `api-gateway/go.mod` 与 `payment-service/go.mod`，执行 `go mod tidy`，并验证两服务构建与基本功能；不修改其他服务的 `go.mod`。

### Session 2026-04-25

- Q: 仓库模块前缀是否需要从 oktetopython 迁移到 gaokaohub？ → A: 否。`github.com/oktetopython/gaokao` 是仓库的正确地址，不需要迁移。架构报告中声称需要"统一到 gaokaohub"的判断是基于错误前提。
- Q: 删除 replace 后是否需要审计目标服务中的重复实现？ → A: 是。作为 US1 的审计子步骤，识别并标记与 pkg 共享库功能重叠的本地实现（CORS wrapper、auth wrapper、error handler），收敛异常路径归因。
- Q: 异常路径分类是否需要新增"设计债务"类别？ → A: 是。删除 replace 后可能暴露 pkg 模块接口不兼容，这既非代码 bug 也非环境问题，需归类为 Design Gap 留待后续迭代。

## User Scenarios & Testing *(mandatory)*

### User Story 1 - 清理冲突 replace 规则与复用审计 (Priority: P1) 🎯 MVP

作为后端维护者，我希望在两个目标服务中删除与当前依赖治理策略不一致的本地 `replace` 指令，并审计可复用的共享实现，以降低依赖分叉和代码重复。

**Why this priority**: 这是本次变更的核心目标，直接决定依赖解析一致性和代码收敛方向。
**Independent Test**: 仅交付该故事时，两个服务的 `go.mod` 不再包含冲突 `replace` 规则，且重复实现有明确审计结论。

**Acceptance Scenarios**:

1. **Given** `api-gateway/go.mod` 存在冲突 `replace`，**When** 完成修改，**Then** 冲突 `replace` 被删除。
2. **Given** `payment-service/go.mod` 存在冲突 `replace`，**When** 完成修改，**Then** 冲突 `replace` 被删除。
3. **Given** 目标服务存在与 `pkg/*` 功能重叠的本地实现，**When** 审计完成，**Then** 每项重叠有 keep/deprecate/remove 结论。
4. **Given** 审计结论为 remove 的本地实现，**When** 执行移除，**Then** `go build` 仍可通过或失败原因被归因为 Design Gap。

**判定规则 — 何为"冲突 replace"**:

删除条件（全部满足）:
1. `replace` 目标指向 `../../pkg/*` 形式的本地路径
2. 该 pkg 子模块已在 `go.work` 中注册
3. 删除后不影响编译（由 build 验证确认）

保留条件（任一满足）:
1. `replace` 目标指向服务内部子模块
2. 删除后 `go build` 失败且无 `go.work` 替代解析路径
3. `replace` 用于覆盖第三方依赖版本

**复用改造判定规则**:

移除/标记条件:
1. 服务本地实现与 `pkg/*` 共享库功能完全重叠
2. 服务已通过 `go.work` 可达对应 pkg 模块
3. 移除后功能等价

保留条件:
1. 本地实现有 pkg 共享库不具备的定制逻辑
2. 移除会导致编译失败且无等价替代
3. 功能仅有部分重叠但接口不兼容

---

### User Story 2 - 同步依赖并验证可运行性 (Priority: P2)

作为维护者，我希望在两个目标服务上完成依赖同步与基本验证，确保本次改动可用。

**Why this priority**: 防止"仅改文件不验证"导致后续不可构建或不可运行。
**Independent Test**: 仅交付该故事时，两个服务完成依赖同步并有构建/运行验证结果。

**Acceptance Scenarios**:

1. **Given** 两个目标服务已完成 `go.mod` 修改，**When** 执行依赖同步，**Then** 依赖状态更新且未引入新增冲突。
2. **Given** 依赖同步完成，**When** 执行构建与基本功能验证，**Then** 输出可复核的验证结果。

**异常路径分类规则**:

- **Code Issue**: 删除 replace 后 `go mod tidy` 或 `go build` 报错，且错误指向模块路径或依赖版本不匹配
- **Environment Limitation**: `go` 未安装、DB/Redis 不可达等外部阻塞
- **Design Gap**: 删除 replace 后暴露 pkg 模块接口不兼容，需回退 replace 或修改 pkg 接口

---

### Edge Cases

- 若本地缺失 `go` 或容器工具，必须保留"无法完成源码级构建验证"的明确记录。
- 若删除冲突 `replace` 后触发新依赖冲突，必须在结果中标注并保留失败证据。
- 若运行验证依赖外部组件（如数据库），必须区分"代码问题"与"环境依赖问题"。
- 若审计发现本地实现与 pkg 接口不兼容，标记为 Design Gap 而非强行移除。

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 系统 MUST 仅修改 `services/api-gateway/go.mod` 和 `services/payment-service/go.mod`。
- **FR-002**: 系统 MUST 删除上述两个文件中与当前依赖治理策略不一致的本地 `replace` 指令。
- **FR-003**: 系统 MUST 审计两个目标服务中与 `pkg/*` 功能重叠的本地实现，并产出 keep/deprecate/remove 结论。
- **FR-004**: 系统 MUST 在两个目标服务中执行依赖同步（`go mod tidy`）。
- **FR-005**: 系统 MUST 对两个目标服务执行构建验证并输出结果。
- **FR-006**: 系统 MUST 对两个目标服务执行基本功能验证并输出结果。
- **FR-007**: 系统 MUST NOT 修改其他服务的 `go.mod` 文件。
- **FR-008**: 系统 MUST 在验证结果中区分 Code Issue / Environment Limitation / Design Gap。
- **FR-009**: 系统 MUST 产出质量门槛结论（`Gate Passed` / `Gate Waived` / `Gate Blocked`）；若为 `Gate Waived`，MUST 记录阻塞证据、责任人（owner）、截止时间（deadline）与回补计划。

### Constitution Alignment *(mandatory)*

- **CA-001 Excellence**: 变更必须带有可复核验证结果，而非仅文件编辑。
- **CA-002 Reuse**: 沿用现有依赖体系与目录结构，不引入平行机制。
- **CA-003 No Simplification**: 不以"未验证"代替交付，必须给出验证结论。
- **CA-004 Stack Boundary**: 本次仅触及 Go 模块依赖管理边界，不扩展技术栈。
- **CA-005 Quality Gates**: 两个目标服务均需完成依赖同步和验证记录。

### Key Entities *(include if feature involves data)*

- **Target Service Module File**: 目标服务的 `go.mod` 文件，属性包括文件路径、冲突 `replace` 状态。
- **Reuse Audit Item**: 目标服务中的本地实现，属性包括文件路径、pkg 等价物、审计结论（keep/deprecate/remove）、Design Gap 标记。
- **Dependency Sync Result**: 依赖同步结果，属性包括是否执行、是否成功、失败原因。
- **Build/Run Verification Record**: 构建与基本功能验证记录，属性包括命令、结果、分类（Code Issue / Environment Limitation / Design Gap）。

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `api-gateway/go.mod` 与 `payment-service/go.mod` 的冲突 `replace` 删除完成率达到 100%。
- **SC-002**: 两个目标服务的复用审计完成率达到 100%（每项重叠有明确结论）。
- **SC-003**: 两个目标服务依赖同步执行覆盖率达到 100%。
- **SC-004**: 两个目标服务构建验证执行覆盖率达到 100%，并有明确成功/失败结论。
- **SC-005**: 两个目标服务基本功能验证执行覆盖率达到 100%，并有明确成功/失败结论。
- **SC-006**: 其他服务 `go.mod` 被修改数量为 0。
- **SC-007**: 所有失败项有明确的 Code Issue / Environment Limitation / Design Gap 分类。
- **SC-008**: 质量门槛结论覆盖率达到 100%（本次迭代必须明确给出 `Gate Passed` / `Gate Waived` / `Gate Blocked` 之一；若 `Gate Waived`，豁免元数据完整率达到 100%）。

## Assumptions

- "基本功能验证"至少包括服务可启动或核心健康检查可访问的证据。
- 基本功能验证判定标准固定为：
  - `api-gateway`: 进程启动后 15 秒内 `GET /healthz` 返回 HTTP 200（或同等健康端点返回 200）。
  - `payment-service`: 启动命令执行后进程在 15 秒内未出现 fatal/panic 并保持存活（若有健康端点则优先使用 HTTP 200 判定）。
- 若受环境限制无法完成源码级构建，允许使用可执行运行验证并明确说明限制。
- 本次不包含其他服务 `go.mod` 清理，也不包含全仓库路径统一改造。
- 仓库模块前缀以 `github.com/oktetopython/gaokao/*` 为准，本次不包含前缀迁移。
- 复用审计以审计为主、移除/标记为辅，不以激进重构为目标。
