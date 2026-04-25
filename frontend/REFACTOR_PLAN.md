# 前端代码整改计划

> 生成时间: 2026-01-16
> 目标: 消除过度工程化，精简代码库

---

## 一、问题概述

本轮清理（死代码/空壳组件/重复类型/过度抽象）已完成，当前剩余问题集中在：

- 认证与会话：refresh_token 链路未闭环
- API 契约对齐：分页参数、筛选枚举、排序字段
- 路由与入口一致性：详情页/会员/模拟入口
- 功能缺口：支付、会员、个人中心、数据分析

---

## 二、审计发现（本轮更新，按严重度排序）

### 高危/阻断
- **刷新令牌链路未闭环**：登录未存 `refresh_token`，401 刷新失败时不会触发登出；登出也未清理 `refresh_token`，导致过期状态滞留。
  - 位置: `src/api/api-client.ts`、`src/stores/user.ts`、`src/api/user.ts`
- **API 契约不一致导致筛选/分页失效**：院校/专业分页使用 `limit` 而后端为 `page_size`；筛选 `type/level` 使用中文值，后端枚举为 `undergraduate/graduate/vocational`、`double_first_class/ordinary`。
  - 位置: `src/views/UniversitiesPageModern.vue`、`src/views/MajorsPage.vue`、`src/api/university.ts`

### 中危/功能缺口
- **专业详情路由缺失**：列表点击跳转 `/majors/:id`，但无路由注册，直接 404。
  - 位置: `src/views/MajorsPage.vue`、`src/router/index.ts`
- **院校详情仍为占位**：`/universities/:id` 仅重定向，点击仍提示“开发中”，`detailId` query 未消费。
  - 位置: `src/router/index.ts`、`src/views/UniversitiesPageModern.vue`、`src/views/RecommendationPage.vue`
- **首页功能入口与路由不一致**：`/membership` 入口仍展示但路由被注释；`/simulation` 仅复用推荐页，功能描述与实际不符。
  - 位置: `src/config/constants.ts`、`src/views/HomePageModern.vue`、`src/router/index.ts`
- **支付/会员链路仍为 TODO**：store 与组件均为占位实现，接口契约未定。
  - 位置: `src/stores/payment.ts`、`src/components/PaymentForm.vue`、`src/components/MembershipStatus.vue`、`src/components/OrderHistory.vue`、`src/views/MembershipPage.vue`
- **数据分析/个人中心为占位实现**：使用空态或模拟数据，未接入真实 API。
  - 位置: `src/views/AnalysisPage.vue`、`src/views/ProfilePage.vue`

### 低危/质量与一致性
- **RegisterPage 仍为模拟逻辑且未挂路由**：形成死代码与重复维护成本。
  - 位置: `src/views/RegisterPage.vue`
- **首页统计字段映射不完整**：统计接口缺少 `majorCount/userCount/accuracyRate` 时长期回退默认值。
  - 位置: `src/views/HomePageModern.vue`
- **专业列表失败时使用 mock 数据**：上线后可能掩盖真实 API 故障。
  - 位置: `src/views/MajorsPage.vue`

### 已修复（本轮确认）
- 统一 `auth_token`，修复导出报告 `Blob`，修复院校分页二次 `slice`，新增 `/simulation` 与 `/universities/:id` 路由占位，补齐默认 logo 资源。

---

## 三、执行计划

### 阶段0: 关键功能修复（优先级: 最高）

1. **补齐 refresh_token 链路**：登录保存 refresh_token，登出清理；401 无 refresh_token 时强制登出。
2. **分页参数对齐后端**：院校/专业由 `limit` 改为 `page_size`。
3. **筛选枚举映射**：UI 中文值映射后端枚举（`type/level`），或后端兼容中文值。
4. **排序字段对齐**：`sortBy` 转换为 `sort_by` + `sort_order` 透传后端。
5. **路由入口一致性**：补齐 `/majors/:id`，明确 `/membership` 与 `/simulation` 的入口策略。

### 阶段1: 删除死代码（已完成）

#### 1.1 删除未使用的页面组件

```bash
# 以下文件无任何引用，直接删除
rm src/views/HomePage.vue           # 552行，路由使用HomePageModern
rm src/views/HomePageSimple.vue     # 385行，无引用
rm src/views/UniversitiesPage.vue   # 565行，路由使用Modern版本
```

**验证方式**: 删除后运行 `npm run build`，确认无报错

#### 1.2 删除未使用的composable

```bash
# useResponsive.ts 无任何文件导入
rm src/composables/useResponsive.ts  # 297行
```

**验证方式**: 全局搜索 `useResponsive` 确认无引用

---

### 阶段2: 移除空壳组件（已完成）

#### 2.1 删除占位analysis组件

```bash
# 这些组件仅显示"功能正在开发中"或随机模拟，无实际功能
rm src/components/analysis/ComparisonAnalysis.vue   # 45行
rm src/components/analysis/EmploymentAnalysis.vue   # 45行
rm src/components/analysis/TrendAnalysis.vue        # 44行
rm src/components/analysis/ProbabilityAnalysis.vue  # 84行（随机数模拟）
rm -rf src/components/analysis/                     # 删除整个目录
```

#### 2.2 简化AnalysisPage.vue

修改 `src/views/AnalysisPage.vue`，移除对空壳组件的引用，改为简单提示页面。

---

### 阶段3: 合并重复类型定义（已完成）

#### 3.1 类型文件现状

| 文件 | 重复定义 |
|------|---------|
| `types/api.ts` | University, ApiResponse, UniversitySearchParams |
| `types/api-params.ts` | ApiResponse, UniversitySearchParams (重复) |
| `types/university.ts` | University (重复), Major |

#### 3.2 整合方案

**目标结构:**
```
types/
├── index.ts          # 统一导出
├── university.ts     # 院校、专业相关类型（保留）
├── user.ts           # 用户相关类型（保留）
├── recommendation.ts # 推荐相关类型（保留）
├── payment.ts        # 支付相关类型（保留）
└── api.ts            # API通用类型（合并api-params.ts）
```

**操作步骤:**
1. 将 `api-params.ts` 中独有的类型合并到 `api.ts`
2. 删除 `api-params.ts`
3. 更新所有 `import from 'api-params'` 为 `import from 'api'`

---

### 阶段4: 简化过度抽象（已完成）

#### 4.1 VirtualList评估

**现状**: 使用虚拟滚动处理列表
**问题**: 实际列表数据量通常 < 100条，虚拟滚动带来的复杂性远超收益

**方案A（保守）**: 保留但提高触发阈值
```typescript
// 当前: v-if="items.length > 5"
// 修改: v-if="items.length > 100"
```

**方案B（激进）**: 完全移除VirtualList，使用普通v-for
```bash
rm src/components/common/VirtualList.vue
# 更新引用处改为普通列表渲染
```

**建议**: 采用方案A，保留组件但提高阈值

#### 4.2 common组件评估

| 组件 | 使用次数 | 建议 |
|------|---------|-----|
| ErrorBoundary.vue | 1次 | 保留 |
| LoadingOverlay.vue | 2次 | 保留 |
| SkeletonCard.vue | 3次 | 保留 |
| SkeletonList.vue | 2次 | 保留 |
| VirtualList.vue | 2次 | 保留但简化 |

---

### 阶段5: 功能一致性修复（优先级: 中）

1. **注册入口收敛**：保留 `LoginPage` 的注册或 `RegisterPage`，删除/改造另一个。
2. **支付链路决策**：要么实现 API，要么移除 UI 入口；避免“可点但无效”。
3. **首页统计字段对齐**：补齐 `majorCount/userCount/accuracyRate` 或调整展示。
4. **模拟填报入口澄清**：若复用推荐页，需调整文案或隐藏入口。

---

## 四、执行检查清单

### 阶段0检查项
- [x] 统一 `auth_token`
- [ ] 补齐 refresh_token 存储与登出清理（401 无 refresh_token 强制登出）
- [x] 修复 `universityApi` 查询参数传递
- [x] 移除院校分页二次 `slice`（后端分页）
- [ ] 分页参数对齐 `page_size`（院校/专业）
- [ ] 筛选枚举映射 `type/level`
- [ ] 排序字段对齐 `sort_by/sort_order`
- [ ] 补齐 `/majors/:id` 或移除入口
- [ ] `/membership` 入口与路由一致
- [ ] `/simulation` 入口文案/功能一致

### 阶段1检查项
- [x] 删除 HomePage.vue
- [x] 删除 HomePageSimple.vue
- [x] 删除 UniversitiesPage.vue
- [x] 删除 useResponsive.ts
- [ ] 运行 `npm run type-check` 通过
- [ ] 运行 `npm run build` 通过

### 阶段2检查项
- [x] 删除 analysis 目录
- [x] 更新 AnalysisPage.vue
- [ ] 运行 `npm run build` 通过

### 阶段3检查项
- [x] 合并类型定义到 api.ts
- [x] 删除 api-params.ts
- [x] 更新所有导入语句
- [ ] 运行 `npm run type-check` 通过

### 阶段4检查项
- [x] 提高VirtualList触发阈值
- [ ] 运行 `npm run build` 通过

### 阶段5检查项
- [ ] 注册入口仅保留一种实现
- [ ] 支付/会员入口与后端契约一致（或移除入口）
- [ ] 首页统计字段对齐
- [ ] 模拟填报入口文案/功能一致

---

## 五、预期成果

| 指标 | 整改前 | 整改后 | 变化 |
|------|-------|-------|------|
| 源文件数 | 66 | 57 | -9 |
| 代码总行数 | ~12000 | ~9700 | -19% |
| 构建体积 | 1044KB | ~950KB | -9% |
| 类型文件 | 6个 | 5个 | -1 |

---

## 六、回滚方案

整改前创建Git分支备份:
```bash
git checkout -b backup/before-refactor
git add -A
git commit -m "backup: 整改前代码备份"
git checkout master
```

如需回滚:
```bash
git checkout backup/before-refactor -- src/
```

---

## 七、备注

- 阶段1、2可立即执行，风险极低
- 阶段3需仔细检查导入依赖
- 阶段4视实际需求决定是否执行
- 阶段0、5属于功能性修复，需与后端契约对齐
- 建议每个阶段完成后单独提交，便于追踪
