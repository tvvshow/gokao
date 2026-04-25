# Design Document: Git合并冲突修复

## Overview

本设计文档描述Git合并冲突修复的技术方案，包括冲突识别、修复策略、验证流程等。

## Architecture

### 冲突分布架构

```
gaokao/
├── services/
│   ├── monitoring-service/          # 39个冲突
│   │   └── internal/alerts/
│   │       └── alert_manager.go     # 告警管理核心
│   ├── payment-service/             # 37个冲突
│   │   └── internal/
│   │       ├── repository/
│   │       │   └── payment_repository.go  # 22个冲突
│   │       ├── handlers/
│   │       │   └── payment_handler.go     # 14个冲突
│   │       └── models/
│   │           └── payment_models.go      # 1个冲突
│   └── api-gateway/                 # 11个冲突
│       └── main.go
├── pkg/                             # 10个冲突
│   ├── wechat_pay.go               # 5个冲突
│   ├── alipay.go                   # 3个冲突
│   └── metrics.go                  # 2个冲突
└── tests/                          # 3个冲突
```

### 修复策略设计

```
┌─────────────────────────────────────────────────────────────┐
│                    冲突修复决策流程                           │
├─────────────────────────────────────────────────────────────┤
│  1. 识别冲突类型                                             │
│     ├── 代码风格冲突 → 保留Go标准tab缩进                      │
│     ├── 模块路径冲突 → 统一为github.com/oktetopython/gaokao  │
│     ├── 功能实现冲突 → 保留功能更完整的版本                    │
│     └── 架构设计冲突 → 选择与项目整体架构一致的版本            │
│                                                             │
│  2. 执行修复                                                 │
│     ├── 删除冲突标记 (<<<<<<, =======, >>>>>>>)             │
│     ├── 保留选定的代码块                                     │
│     └── 修正import路径                                       │
│                                                             │
│  3. 验证修复                                                 │
│     ├── 语法检查 (go fmt)                                   │
│     ├── 编译验证 (go build)                                 │
│     └── 测试验证 (go test)                                  │
└─────────────────────────────────────────────────────────────┘
```

## Components

### 1. monitoring-service/internal/alerts/alert_manager.go

**冲突数量**: 39个

**修复策略**:
- 保留完整的告警管理实现
- 包含邮件、钉钉、微信通知功能
- 保留metrics依赖和相关方法

**关键决策**: 选择HEAD版本，因为包含更完整的通知渠道支持

### 2. payment-service/internal/repository/payment_repository.go

**冲突数量**: 22个

**修复策略**:
- 使用PaymentOrder模型（与项目架构一致）
- 修正module path为`github.com/oktetopython/gaokao`
- 保留完整的事务支持和行锁功能

**关键决策**: 选择PaymentOrder模型以保持与其他服务的一致性

### 3. payment-service/internal/handlers/payment_handler.go

**冲突数量**: 14个

**修复策略**:
- 保持与repository层的一致性
- 修正import路径
- 保留完整的错误处理逻辑

### 4. api-gateway/main.go

**冲突数量**: 11个

**修复策略**:
- 保留正确的Go代码风格（tab缩进）
- 保留完整的路由配置
- 保留中间件链

## Data Flow

```
冲突文件 → 识别冲突标记 → 分析冲突类型 → 选择修复策略 → 执行修复 → 验证 → 完成
```

## Error Handling

1. **编译错误**: 检查import路径和类型定义
2. **测试失败**: 检查模型字段和方法签名
3. **依赖错误**: 清理缓存并重新下载依赖

## Security Considerations

- 修复过程中不引入新的安全漏洞
- 保留原有的认证和授权逻辑
- 确保敏感信息处理方式不变

