# Implementation Plan: GUI审计修复任务

## Overview

本任务列表基于GUI审计发现的问题，按优先级排序，提供具体的修复步骤。使用Vitest进行单元测试和属性测试。

## Tasks

- [x] 1. 修复严重问题
  - [x] 1.1 修复api-client.ts中的Git合并冲突
    - 解决`<<<<<<< HEAD`和`>>>>>>> 0dd6b27`之间的冲突
    - 选择axios实现（功能更完善，包含请求拦截器）
    - 确保认证Token正确传递
    - _Requirements: 2.3, 3.1_

  - [ ]* 1.2 编写属性测试：源代码完整性检测
    - **Property 1: 源代码完整性**
    - 检测所有源代码文件不包含Git冲突标记
    - **Validates: Requirements 2.3**

  - [x] 1.3 修复RecommendationPage.vue中的未导入引用
    - 添加`import { recommendationApi } from '@/api/recommendation'`
    - 验证所有API调用正常工作
    - _Requirements: 2.1, 2.2_

- [x] 2. Checkpoint - 验证严重问题修复
  - 运行`npm run build`确保编译通过
  - 运行`npm run lint`确保无语法错误
  - 确保所有页面可正常访问

- [x] 3. 修复安全问题
  - [x] 3.1 完善API客户端认证机制
    - 确保所有请求携带Bearer Token
    - 实现Token过期自动刷新
    - 添加401响应的统一处理
    - _Requirements: 3.1, 3.4_

  - [ ]* 3.2 编写属性测试：API认证Token传递
    - **Property 4: API认证Token传递**
    - 验证所有需要认证的请求包含Authorization头
    - **Validates: Requirements 3.1**

  - [x] 3.3 加强表单验证
    - 添加密码强度验证（至少8位，包含大小写字母和数字）
    - 添加XSS防护，对用户输入进行转义
    - 验证所有表单的必填字段
    - _Requirements: 3.2, 3.3_

  - [ ]* 3.4 编写属性测试：表单验证完整性
    - **Property 5: 表单验证完整性**
    - 验证所有必填字段有对应的验证规则
    - **Validates: Requirements 3.2, 4.6**

- [x] 4. Checkpoint - 验证安全修复
  - 测试登录/注册流程
  - 测试Token过期处理
  - 测试表单验证
  - 运行属性测试

- [ ] 5. 优化代码质量
  - [ ] 5.1 拆分RecommendationPage.vue
    - 提取StudentInfoForm组件
    - 提取RecommendationResults组件
    - 提取RecommendationStats组件
    - _Requirements: 2.4_

  - [ ]* 5.2 编写属性测试：组件大小限制
    - **Property 3: 组件大小限制**
    - 验证所有Vue组件不超过500行
    - **Validates: Requirements 1.1, 2.4**

  - [ ] 5.3 完善TypeScript类型
    - 消除any类型使用
    - 添加缺失的类型定义
    - _Requirements: 2.1_

  - [ ]* 5.4 编写属性测试：TypeScript类型安全
    - **Property 2: TypeScript类型安全**
    - 验证any类型使用最小化
    - **Validates: Requirements 2.1**

  - [ ] 5.5 移除硬编码数据
    - 首页统计数据从API获取
    - 省份列表从配置文件获取
    - _Requirements: 2.1_

- [ ] 6. Checkpoint - 验证代码质量
  - 运行`npm run lint`
  - 运行`npm run type-check`
  - 运行属性测试
  - 代码审查

- [ ] 7. 优化用户体验
  - [ ] 7.1 添加加载状态
    - 为所有API调用添加loading状态
    - 添加骨架屏组件
    - _Requirements: 4.1_

  - [ ]* 7.2 编写属性测试：异步操作加载状态
    - **Property 6: 异步操作加载状态**
    - 验证所有异步API调用有对应的loading状态
    - **Validates: Requirements 4.1**

  - [ ] 7.3 统一错误处理
    - 创建全局错误边界组件
    - 统一错误提示样式
    - _Requirements: 4.2_

  - [ ] 7.4 优化响应式布局
    - 测试移动端适配
    - 修复布局问题
    - _Requirements: 4.3_

- [ ] 8. 优化性能和可访问性
  - [ ] 8.1 验证路由懒加载
    - 检查所有路由使用动态import
    - _Requirements: 1.3, 5.1_

  - [ ]* 8.2 编写属性测试：路由懒加载
    - **Property 8: 路由懒加载**
    - 验证所有路由组件使用动态import语法
    - **Validates: Requirements 1.3, 5.1**

  - [ ] 8.3 添加可访问性标签
    - 为交互元素添加aria-label
    - _Requirements: 6.1_

  - [ ]* 8.4 编写属性测试：可访问性标签
    - **Property 7: 可访问性标签**
    - 验证交互元素有适当的aria标签
    - **Validates: Requirements 6.1**

  - [ ] 8.5 实现虚拟滚动
    - 为院校列表添加虚拟滚动
    - 为推荐结果添加虚拟滚动
    - _Requirements: 5.4_

- [ ] 9. 添加集成测试
  - [ ]* 9.1 添加单元测试
    - 测试Store逻辑
    - 测试工具函数
    - _Requirements: 8.1, 8.2_

  - [ ]* 9.2 添加E2E测试
    - 测试登录流程
    - 测试推荐流程
    - _Requirements: 8.3_

- [ ] 10. Final Checkpoint
  - 运行所有属性测试
  - 运行所有单元测试
  - 性能测试
  - 用户验收测试

## Notes

- 任务按优先级排序，严重问题优先修复
- 每个Checkpoint确保前面的修复正确完成
- 标记`*`的任务为可选测试任务
- 属性测试使用Vitest + fast-check库
- 所有修复需要在开发环境验证后再部署
- 属性测试配置：最少100次迭代
