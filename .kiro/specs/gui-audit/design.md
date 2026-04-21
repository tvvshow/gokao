# GUI审计设计文档

## Overview

本设计文档详细分析高考志愿填报系统前端GUI的架构、代码质量、安全性、用户体验等方面，并提供具体的改进建议。

## 架构分析

### 当前架构

```
frontend/src/
├── api/                    # API客户端层
│   ├── api-client.ts       # 统一HTTP客户端
│   ├── recommendation.ts   # 推荐API
│   ├── university.ts       # 院校API
│   └── user.ts             # 用户API
├── components/             # 可复用组件
│   ├── analysis/           # 分析相关组件
│   ├── AppHeader.vue       # 全局头部
│   ├── AppFooter.vue       # 全局底部
│   └── ...                 # 其他组件
├── composables/            # 组合式函数
│   └── useResponsive.ts    # 响应式工具
├── router/                 # 路由配置
│   └── index.ts            # 路由定义
├── services/               # 服务层
│   └── api.ts              # API服务
├── stores/                 # 状态管理
│   ├── recommendation.ts   # 推荐状态
│   └── user.ts             # 用户状态
├── styles/                 # 样式文件
│   └── design-system.css   # 设计系统
├── types/                  # TypeScript类型
│   ├── api.ts              # API类型
│   ├── recommendation.ts   # 推荐类型
│   ├── university.ts       # 院校类型
│   └── user.ts             # 用户类型
├── views/                  # 页面视图
│   ├── HomePage*.vue       # 首页
│   ├── LoginPage.vue       # 登录页
│   ├── RecommendationPage.vue  # 推荐页
│   └── ...                 # 其他页面
├── App.vue                 # 根组件
└── main.ts                 # 入口文件
```

### 架构评估

| 方面 | 评分 | 说明 |
|------|------|------|
| 组件化 | ⭐⭐⭐⭐☆ | 组件划分合理，但部分组件过大 |
| 状态管理 | ⭐⭐⭐⭐☆ | Pinia使用规范，状态分离清晰 |
| 路由设计 | ⭐⭐⭐⭐⭐ | 懒加载、路由守卫完善 |
| 类型系统 | ⭐⭐⭐☆☆ | 类型定义存在，但不够完整 |
| API层 | ⭐⭐⭐☆☆ | 存在Git冲突，需要修复 |

## 组件分析

### Components

#### AppHeader.vue
- **优点**: 响应式设计、暗色模式支持、用户菜单完善
- **问题**: 组件较大(~200行)，可考虑拆分

#### RecommendationCard.vue
- **优点**: 功能完整、样式美观
- **问题**: 缺少错误边界处理

### Views

#### HomePageModern.vue
- **优点**: 现代化设计、动画效果好
- **问题**: 统计数据硬编码，应从API获取

#### LoginPage.vue
- **优点**: 表单验证完善、用户体验好
- **问题**: 密码强度验证不足

#### RecommendationPage.vue
- **优点**: 功能丰富、交互流畅
- **问题**: 
  - 组件过大(~500行)，应拆分
  - `recommendationApi`未导入但被使用
  - 缺少错误边界

## 数据模型

### User类型
```typescript
interface User {
  id: string
  username: string
  email: string
  phone?: string
  avatar?: string
  membershipLevel: 'free' | 'basic' | 'premium'
  membershipExpiry?: string
  createdAt: string
  updatedAt: string
}
```

### Recommendation类型
```typescript
interface Recommendation {
  id: string
  university: University
  type: 'aggressive' | 'moderate' | 'conservative'
  admissionProbability: number
  matchScore: number
  recommendReason: string
  riskLevel: 'low' | 'medium' | 'high'
  suggestedMajors: Array<{id: string; name: string; probability: number}>
  historicalData: Array<{minScore: number; avgScore: number; maxScore: number; year: number}>
}
```

## 发现的问题

### 🔴 严重问题

#### 1. Git合并冲突未解决
**文件**: `frontend/src/api/api-client.ts`
**问题**: 存在未解决的Git合并冲突标记
```typescript
```
**影响**: 代码无法正常编译运行
**修复**: 立即解决合并冲突，选择正确的实现

#### 2. 未导入的API引用
**文件**: `frontend/src/views/RecommendationPage.vue`
**问题**: 使用了`recommendationApi`但未导入
```typescript
const response = await recommendationApi.exportReport(recommendations.value)
```
**影响**: 运行时错误
**修复**: 添加正确的导入语句

### 🟡 中等问题

#### 3. 认证Token未在请求中携带
**文件**: `frontend/src/api/api-client.ts`
**问题**: fetch版本的API客户端未携带认证Token
**影响**: 需要认证的API请求会失败
**修复**: 在请求头中添加Authorization

#### 4. 组件过大
**文件**: `RecommendationPage.vue` (~500行)
**问题**: 单个组件代码量过大，难以维护
**修复**: 拆分为多个子组件

#### 5. 硬编码数据
**文件**: `HomePageModern.vue`
**问题**: 统计数据硬编码
```typescript
const stats = ref([
  { icon: BuildingIcon, value: '2700+', label: '合作高校' },
  // ...
])
```
**修复**: 从API获取真实数据

### 🟢 轻微问题

#### 6. 缺少加载状态
**多个文件**
**问题**: 部分API调用缺少加载状态显示
**修复**: 添加loading状态和骨架屏

#### 7. 错误处理不一致
**多个文件**
**问题**: 错误处理方式不统一
**修复**: 建立统一的错误处理机制

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

基于prework分析，以下是经过整合和去重后的核心正确性属性：

### Property 1: 源代码完整性
*For any* 源代码文件（.ts, .vue, .js），不应包含Git合并冲突标记（`<<<<<<<`, `=======`, `>>>>>>>`）
**Validates: Requirements 2.3**

### Property 2: TypeScript类型安全
*For any* TypeScript文件，`any`类型的使用次数应该最小化（每个文件不超过3处，且必须有注释说明原因）
**Validates: Requirements 2.1**

### Property 3: 组件大小限制
*For any* Vue组件文件，总行数不应超过500行，以保持单一职责原则
**Validates: Requirements 1.1, 2.4**

### Property 4: API认证Token传递
*For any* 需要认证的API请求，请求头中必须包含格式为`Bearer {token}`的Authorization头
**Validates: Requirements 3.1**

### Property 5: 表单验证完整性
*For any* 表单组件中的必填字段，必须定义对应的验证规则，且验证必须在表单提交前执行
**Validates: Requirements 3.2, 4.6**

### Property 6: 异步操作加载状态
*For any* 异步API调用，必须有对应的loading状态变量，在请求开始时设为true，请求结束时设为false
**Validates: Requirements 4.1**

### Property 7: 可访问性标签
*For any* 可交互元素（按钮、链接、表单控件），应该有适当的aria-label或可访问的文本内容
**Validates: Requirements 6.1**

### Property 8: 路由懒加载
*For any* 路由配置中的组件引用，应该使用动态import语法（`() => import(...)`）实现懒加载
**Validates: Requirements 1.3, 5.1**

## Error Handling

### 当前错误处理机制

1. **API层**: `api-client.ts`提供统一的HTTP错误处理
2. **组件层**: 使用try-catch和ElMessage显示错误
3. **路由层**: 路由守卫处理认证错误

### 改进建议

1. 建立全局错误边界组件
2. 统一错误码和错误消息
3. 添加错误日志上报机制
4. 实现优雅降级策略

## Testing Strategy

### 单元测试
- 使用Vitest测试组件逻辑
- 测试Store的状态管理
- 测试工具函数和composables

### 集成测试
- 测试组件间交互
- 测试API调用流程
- 测试路由导航

### E2E测试
- 使用Cypress或Playwright
- 覆盖关键用户流程
- 测试响应式布局

### 属性测试
- 使用fast-check进行属性测试
- 验证表单验证逻辑
- 验证数据转换函数
