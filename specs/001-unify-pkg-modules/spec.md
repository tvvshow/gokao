# Feature Specification: Shared Module Unification

**Feature Branch**: `001-unify-pkg-modules`  
**Created**: 2026-04-24  
**Status**: Draft  
**Input**: User description: "本次仅涉及：1. api-gateway/go.mod 删除 replace 指令 2. payment-service/go.mod 删除 replace 指令 3. 运行 go mod tidy 同步依赖 4. 验证构建和基本功能 不修改其他服务的 go.mod"

## Clarifications

### Session 2026-04-24

- Q: 本次变更范围是否仅限两个服务并禁止扩展到其他服务？ → A: 是，仅修改 `api-gateway/go.mod` 与 `payment-service/go.mod`，执行 `go mod tidy`，并验证两服务构建与基本功能；不修改其他服务的 `go.mod`。

## User Scenarios & Testing *(mandatory)*

### User Story 1 - 清理冲突 replace 规则 (Priority: P1)

作为后端维护者，我希望仅在两个目标服务中删除与 `pkg` 命名空间冲突的 `replace` 指令，以降低依赖分叉和维护成本。

**Why this priority**: 这是本次变更的核心目标，直接决定依赖解析一致性。  
**Independent Test**: 仅交付该故事时，两个服务的 `go.mod` 不再包含冲突 `replace` 规则。

**Acceptance Scenarios**:

1. **Given** `api-gateway/go.mod` 存在冲突 `replace`，**When** 完成修改，**Then** 冲突 `replace` 被删除。  
2. **Given** `payment-service/go.mod` 存在冲突 `replace`，**When** 完成修改，**Then** 冲突 `replace` 被删除。

---

### User Story 2 - 同步依赖并验证可运行性 (Priority: P2)

作为维护者，我希望在两个目标服务上完成依赖同步与基本验证，确保本次改动可用。

**Why this priority**: 防止“仅改文件不验证”导致后续不可构建或不可运行。  
**Independent Test**: 仅交付该故事时，两个服务完成依赖同步并有构建/运行验证结果。

**Acceptance Scenarios**:

1. **Given** 两个目标服务已完成 `go.mod` 修改，**When** 执行依赖同步，**Then** 依赖状态更新且未引入新增冲突。  
2. **Given** 依赖同步完成，**When** 执行构建与基本功能验证，**Then** 输出可复核的验证结果。

---

### Edge Cases

- 若本地缺失 `go` 或容器工具，必须保留“无法完成源码级构建验证”的明确记录。  
- 若删除冲突 `replace` 后触发新依赖冲突，必须在结果中标注并保留失败证据。  
- 若运行验证依赖外部组件（如数据库），必须区分“代码问题”与“环境依赖问题”。

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 系统 MUST 仅修改 `services/api-gateway/go.mod` 和 `services/payment-service/go.mod`。  
- **FR-002**: 系统 MUST 删除上述两个文件中与 `pkg` 命名空间冲突的 `replace` 指令。  
- **FR-003**: 系统 MUST 在两个目标服务中执行依赖同步（`go mod tidy`）。  
- **FR-004**: 系统 MUST 对两个目标服务执行构建验证并输出结果。  
- **FR-005**: 系统 MUST 对两个目标服务执行基本功能验证并输出结果。  
- **FR-006**: 系统 MUST NOT 修改其他服务的 `go.mod` 文件。  
- **FR-007**: 系统 MUST 在验证结果中区分“构建问题”与“环境依赖问题”。

### Constitution Alignment *(mandatory)*

- **CA-001 Excellence**: 变更必须带有可复核验证结果，而非仅文件编辑。  
- **CA-002 Reuse**: 沿用现有依赖体系与目录结构，不引入平行机制。  
- **CA-003 No Simplification**: 不以“未验证”代替交付，必须给出验证结论。  
- **CA-004 Stack Boundary**: 本次仅触及 Go 模块依赖管理边界，不扩展技术栈。  
- **CA-005 Quality Gates**: 两个目标服务均需完成依赖同步和验证记录。

### Key Entities *(include if feature involves data)*

- **Target Service Module File**: 目标服务的 `go.mod` 文件，属性包括文件路径、冲突 `replace` 状态。  
- **Dependency Sync Result**: 依赖同步结果，属性包括是否执行、是否成功、失败原因。  
- **Build/Run Verification Record**: 构建与基本功能验证记录，属性包括命令、结果、环境限制。

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `api-gateway/go.mod` 与 `payment-service/go.mod` 的冲突 `replace` 删除完成率达到 100%。  
- **SC-002**: 两个目标服务依赖同步执行覆盖率达到 100%。  
- **SC-003**: 两个目标服务构建验证执行覆盖率达到 100%，并有明确成功/失败结论。  
- **SC-004**: 两个目标服务基本功能验证执行覆盖率达到 100%，并有明确成功/失败结论。  
- **SC-005**: 其他服务 `go.mod` 被修改数量为 0。

## Assumptions

- “基本功能验证”至少包括服务可启动或核心健康检查可访问的证据。  
- 若受环境限制无法完成源码级构建，允许使用可执行运行验证并明确说明限制。  
- 本次不包含其他服务 `go.mod` 清理，也不包含全仓库路径统一改造。
