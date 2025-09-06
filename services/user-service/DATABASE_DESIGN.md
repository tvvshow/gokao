# 高考志愿填报系统用户服务数据库设计文档

## 概述

本文档描述了高考志愿填报系统用户服务的完整数据库设计，包括用户管理、权限控制、会员系统、设备管理、会话管理等核心功能。

## 设计原则

- **可扩展性**: 支持横向扩展和数据分区
- **安全性**: 实现RBAC权限控制和设备指纹验证
- **性能优化**: 合理的索引设计和查询优化
- **数据完整性**: 完善的约束和触发器
- **GDPR合规**: 支持用户数据删除和隐私保护

## 数据库架构

### 技术栈
- **数据库**: PostgreSQL 14+
- **ORM**: GORM (Go)
- **缓存**: Redis
- **扩展**: uuid-ossp, pg_trgm, pg_cron (可选)

### 核心表结构

## 1. 用户管理模块

### 1.1 用户表 (users)

**功能**: 存储用户基础信息和会员状态

| 字段名 | 类型 | 约束 | 描述 |
|--------|------|------|------|
| id | UUID | PRIMARY KEY | 用户唯一标识 |
| username | VARCHAR(50) | UNIQUE, NOT NULL | 用户名 |
| email | VARCHAR(100) | UNIQUE, NOT NULL | 邮箱 |
| phone | VARCHAR(20) | UNIQUE | 手机号 |
| password | VARCHAR(255) | NOT NULL | 密码哈希 |
| nickname | VARCHAR(50) | | 昵称 |
| avatar | VARCHAR(255) | | 头像URL |
| gender | VARCHAR(10) | | 性别 (male/female/other) |
| birthday | TIMESTAMP | | 生日 |
| province | VARCHAR(50) | INDEX | 省份 |
| city | VARCHAR(50) | | 城市 |
| school | VARCHAR(100) | | 学校 |
| grade | VARCHAR(20) | | 年级 |
| status | VARCHAR(20) | INDEX, DEFAULT 'active' | 状态 (active/inactive/suspended) |
| is_verified | BOOLEAN | INDEX, DEFAULT false | 是否已验证 |
| membership_level | VARCHAR(20) | INDEX, DEFAULT 'free' | 会员等级 |
| membership_expiry | TIMESTAMP | INDEX | 会员到期时间 |
| max_devices | INTEGER | DEFAULT 1 | 最大设备数 |
| trial_used | BOOLEAN | DEFAULT false | 是否使用过试用 |
| trial_expiry | TIMESTAMP | | 试用到期时间 |
| last_login_at | TIMESTAMP | | 最后登录时间 |
| last_login_ip | VARCHAR(45) | | 最后登录IP |
| login_count | BIGINT | DEFAULT 0 | 登录次数 |
| created_at | TIMESTAMP | NOT NULL | 创建时间 |
| updated_at | TIMESTAMP | NOT NULL | 更新时间 |
| deleted_at | TIMESTAMP | INDEX | 软删除时间 |

**索引设计**:
- `idx_users_membership_status`: (membership_level, membership_expiry)
- `idx_users_location`: (province, city)
- `idx_users_activity`: (status, last_login_at)
- `idx_users_created_date`: (DATE(created_at))

**约束**:
- 会员到期时间逻辑检查
- 设备数量限制 (1-10)
- 生日合理性检查

## 2. 权限管理模块

### 2.1 角色表 (roles)

**功能**: 定义系统角色

| 字段名 | 类型 | 约束 | 描述 |
|--------|------|------|------|
| id | SERIAL | PRIMARY KEY | 角色ID |
| name | VARCHAR(50) | UNIQUE, NOT NULL | 角色名称 |
| description | VARCHAR(255) | | 角色描述 |
| is_system | BOOLEAN | DEFAULT false | 是否系统角色 |
| created_at | TIMESTAMP | NOT NULL | 创建时间 |
| updated_at | TIMESTAMP | NOT NULL | 更新时间 |

**默认角色**:
- `admin`: 系统管理员
- `user`: 普通用户
- `basic`: 基础会员
- `premium`: 高级会员
- `enterprise`: 企业会员
- `moderator`: 内容审核员

### 2.2 权限表 (permissions)

**功能**: 定义系统权限

| 字段名 | 类型 | 约束 | 描述 |
|--------|------|------|------|
| id | SERIAL | PRIMARY KEY | 权限ID |
| name | VARCHAR(100) | UNIQUE, NOT NULL | 权限名称 |
| description | VARCHAR(255) | | 权限描述 |
| resource | VARCHAR(50) | | 资源类型 |
| action | VARCHAR(50) | | 操作类型 |
| created_at | TIMESTAMP | NOT NULL | 创建时间 |
| updated_at | TIMESTAMP | NOT NULL | 更新时间 |

**权限分类**:
- 用户管理: user:read, user:write, user:delete, user:verify
- 角色权限: role:read, role:write, role:delete, permission:manage
- 会员管理: membership:*, order:*
- 设备管理: device:*
- 会话管理: session:*
- 审计日志: audit:*
- 系统管理: system:*, admin:all

### 2.3 用户角色关联表 (user_roles)

**功能**: 用户和角色的多对多关联

| 字段名 | 类型 | 约束 | 描述 |
|--------|------|------|------|
| user_id | UUID | PRIMARY KEY | 用户ID |
| role_id | INTEGER | PRIMARY KEY | 角色ID |
| created_at | TIMESTAMP | NOT NULL | 分配时间 |

### 2.4 角色权限关联表 (role_permissions)

**功能**: 角色和权限的多对多关联

| 字段名 | 类型 | 约束 | 描述 |
|--------|------|------|------|
| role_id | INTEGER | PRIMARY KEY | 角色ID |
| permission_id | INTEGER | PRIMARY KEY | 权限ID |
| created_at | TIMESTAMP | NOT NULL | 分配时间 |

## 3. 设备管理模块

### 3.1 设备指纹表 (device_fingerprints)

**功能**: 管理用户设备绑定和安全验证

| 字段名 | 类型 | 约束 | 描述 |
|--------|------|------|------|
| id | UUID | PRIMARY KEY | 设备指纹ID |
| user_id | UUID | NOT NULL, INDEX | 用户ID |
| device_id | VARCHAR(255) | UNIQUE, NOT NULL | 设备唯一标识 |
| device_name | VARCHAR(100) | | 设备名称 |
| device_type | VARCHAR(20) | INDEX | 设备类型 (mobile/tablet/desktop) |
| platform | VARCHAR(50) | | 平台 |
| browser | VARCHAR(100) | | 浏览器 |
| browser_version | VARCHAR(50) | | 浏览器版本 |
| os | VARCHAR(50) | | 操作系统 |
| os_version | VARCHAR(50) | | 系统版本 |
| screen_resolution | VARCHAR(20) | | 屏幕分辨率 |
| timezone | VARCHAR(50) | | 时区 |
| language | VARCHAR(10) | | 语言 |
| user_agent | VARCHAR(500) | | User Agent |
| ip_address | VARCHAR(45) | INDEX | IP地址 |
| location | VARCHAR(100) | | 地理位置 |
| is_active | BOOLEAN | INDEX, DEFAULT true | 是否活跃 |
| is_trusted | BOOLEAN | DEFAULT false | 是否信任 |
| last_seen_at | TIMESTAMP | INDEX | 最后活跃时间 |
| created_at | TIMESTAMP | NOT NULL | 创建时间 |
| updated_at | TIMESTAMP | NOT NULL | 更新时间 |
| deleted_at | TIMESTAMP | INDEX | 软删除时间 |

**索引设计**:
- `idx_device_fingerprints_user_active`: (user_id, is_active, last_seen_at)
- `idx_device_fingerprints_type_platform`: (device_type, platform)
- `idx_device_fingerprints_ip`: (ip_address, created_at)

## 4. 会员订单模块

### 4.1 会员订单表 (membership_orders)

**功能**: 管理会员购买订单和支付信息

| 字段名 | 类型 | 约束 | 描述 |
|--------|------|------|------|
| id | UUID | PRIMARY KEY | 订单ID |
| user_id | UUID | NOT NULL, INDEX | 用户ID |
| order_no | VARCHAR(50) | UNIQUE, NOT NULL | 订单号 |
| product_name | VARCHAR(100) | NOT NULL | 产品名称 |
| membership_level | VARCHAR(20) | INDEX, NOT NULL | 会员等级 |
| duration | INTEGER | NOT NULL | 会员时长(天) |
| original_price | BIGINT | NOT NULL | 原价(分) |
| discount_price | BIGINT | DEFAULT 0 | 优惠金额(分) |
| final_price | BIGINT | NOT NULL | 实付金额(分) |
| currency | VARCHAR(10) | DEFAULT 'CNY' | 货币类型 |
| payment_method | VARCHAR(50) | INDEX | 支付方式 |
| payment_provider | VARCHAR(50) | | 支付服务商 |
| payment_id | VARCHAR(100) | INDEX | 支付ID |
| discount_code | VARCHAR(50) | | 优惠码 |
| status | VARCHAR(20) | INDEX, NOT NULL | 订单状态 |
| paid_at | TIMESTAMP | INDEX | 支付时间 |
| expired_at | TIMESTAMP | INDEX | 过期时间 |
| refunded_at | TIMESTAMP | | 退款时间 |
| refund_amount | BIGINT | DEFAULT 0 | 退款金额 |
| refund_reason | VARCHAR(255) | | 退款原因 |
| notes | TEXT | | 备注 |
| created_at | TIMESTAMP | NOT NULL | 创建时间 |
| updated_at | TIMESTAMP | NOT NULL | 更新时间 |
| deleted_at | TIMESTAMP | INDEX | 软删除时间 |

**订单状态**:
- `pending`: 待支付
- `paid`: 已支付
- `cancelled`: 已取消
- `refunded`: 已退款
- `expired`: 已过期

**索引设计**:
- `idx_membership_orders_status_time`: (status, created_at, user_id)
- `idx_membership_orders_payment`: (payment_method, paid_at)
- `idx_membership_orders_revenue`: (membership_level, final_price, paid_at)

**约束**:
- 价格逻辑检查: final_price = original_price - discount_price
- 会员时长检查: 1-3650天
- 支付时间逻辑检查
- 退款逻辑检查

## 5. 会话管理模块

### 5.1 用户会话表 (user_sessions)

**功能**: 管理JWT会话和安全验证

| 字段名 | 类型 | 约束 | 描述 |
|--------|------|------|------|
| id | UUID | PRIMARY KEY | 会话ID |
| user_id | UUID | NOT NULL, INDEX | 用户ID |
| device_id | VARCHAR(255) | INDEX | 设备ID |
| session_token | VARCHAR(255) | UNIQUE, NOT NULL | 会话令牌 |
| refresh_token | VARCHAR(255) | UNIQUE | 刷新令牌 |
| ip_address | VARCHAR(45) | INDEX | IP地址 |
| user_agent | VARCHAR(500) | | User Agent |
| location | VARCHAR(100) | | 地理位置 |
| is_active | BOOLEAN | INDEX, DEFAULT true | 是否活跃 |
| expires_at | TIMESTAMP | INDEX, NOT NULL | 过期时间 |
| refresh_expires_at | TIMESTAMP | INDEX | 刷新令牌过期时间 |
| last_activity_at | TIMESTAMP | INDEX | 最后活动时间 |
| created_at | TIMESTAMP | NOT NULL | 创建时间 |
| updated_at | TIMESTAMP | NOT NULL | 更新时间 |
| deleted_at | TIMESTAMP | INDEX | 软删除时间 |

**索引设计**:
- `idx_user_sessions_activity`: (user_id, is_active, last_activity_at)
- `idx_user_sessions_expiry`: (expires_at, is_active)
- `idx_user_sessions_device`: (device_id, user_id)

**约束**:
- 会话时间逻辑检查: expires_at > created_at
- 活动时间检查: last_activity_at >= created_at
- 刷新令牌时间检查: refresh_expires_at > expires_at

## 6. 审计日志模块

### 6.1 审计日志表 (audit_logs)

**功能**: 记录用户操作和系统事件

| 字段名 | 类型 | 约束 | 描述 |
|--------|------|------|------|
| id | UUID | PRIMARY KEY | 日志ID |
| user_id | UUID | INDEX | 用户ID (可为空) |
| action | VARCHAR(100) | INDEX, NOT NULL | 操作类型 |
| resource | VARCHAR(100) | INDEX | 资源类型 |
| resource_id | VARCHAR(100) | INDEX | 资源ID |
| details | TEXT | | 详细信息 |
| ip | VARCHAR(45) | INDEX | IP地址 |
| user_agent | VARCHAR(500) | | User Agent |
| status | VARCHAR(20) | INDEX | 状态 |
| created_at | TIMESTAMP | NOT NULL | 创建时间 |

**索引设计**:
- `idx_audit_logs_user_action`: (user_id, action, created_at)
- `idx_audit_logs_resource`: (resource, resource_id, action)
- `idx_audit_logs_date_action`: (DATE(created_at), action)

### 6.2 登录尝试表 (login_attempts)

**功能**: 记录登录尝试和安全事件

| 字段名 | 类型 | 约束 | 描述 |
|--------|------|------|------|
| id | SERIAL | PRIMARY KEY | 记录ID |
| username | VARCHAR(50) | INDEX, NOT NULL | 用户名 |
| ip | VARCHAR(45) | INDEX, NOT NULL | IP地址 |
| user_agent | VARCHAR(500) | | User Agent |
| success | BOOLEAN | INDEX | 是否成功 |
| reason | VARCHAR(255) | | 失败原因 |
| created_at | TIMESTAMP | NOT NULL | 创建时间 |

### 6.3 刷新令牌表 (refresh_tokens)

**功能**: 管理JWT刷新令牌

| 字段名 | 类型 | 约束 | 描述 |
|--------|------|------|------|
| id | UUID | PRIMARY KEY | 令牌ID |
| user_id | UUID | NOT NULL, INDEX | 用户ID |
| token | VARCHAR(255) | UNIQUE, NOT NULL | 令牌值 |
| expires_at | TIMESTAMP | INDEX, NOT NULL | 过期时间 |
| is_revoked | BOOLEAN | INDEX, DEFAULT false | 是否撤销 |
| created_at | TIMESTAMP | NOT NULL | 创建时间 |
| updated_at | TIMESTAMP | NOT NULL | 更新时间 |

## 7. 数据视图

### 7.1 用户统计视图 (user_statistics)

提供用户数量、会员分布、注册趋势等统计信息。

### 7.2 会员订单统计视图 (membership_order_statistics)

提供订单统计、收入分析、会员转化等数据。

### 7.3 设备指纹统计视图 (device_fingerprint_statistics)

提供设备类型分布、平台统计、使用情况等信息。

### 7.4 用户会话统计视图 (user_session_statistics)

提供会话时长、活跃度、并发量等统计。

### 7.5 性能监控视图 (performance_metrics)

提供表大小、索引大小、行数等性能指标。

## 8. 数据分区策略

### 8.1 用户表分区 (按省份)

为每个省份创建分区表，提高查询性能：
```sql
-- 示例：为北京创建分区
SELECT create_province_partition('北京');
```

### 8.2 审计日志分区 (按年份)

按年份对审计日志进行分区：
```sql
-- 示例：为2024年创建分区
SELECT create_year_partition('audit_logs', 2024);
```

## 9. 数据清理策略

### 9.1 自动清理函数

- `cleanup_audit_logs(90)`: 清理90天前的审计日志
- `cleanup_login_attempts(24)`: 清理24小时前的登录尝试
- `cleanup_refresh_tokens()`: 清理过期的刷新令牌
- `cleanup_user_sessions(30)`: 清理30天前的非活跃会话
- `cleanup_device_fingerprints(180)`: 清理180天前的非活跃设备
- `cleanup_expired_orders(365)`: 清理365天前的已取消订单
- `update_user_membership_status()`: 更新过期会员状态

### 9.2 定时任务 (pg_cron)

```sql
-- 每日凌晨2点清理审计日志
SELECT cron.schedule('cleanup-audit-logs', '0 2 * * *', 'SELECT cleanup_audit_logs(90);');

-- 每日凌晨更新会员状态
SELECT cron.schedule('update-membership-status', '0 0 * * *', 'SELECT update_user_membership_status();');
```

## 10. 安全措施

### 10.1 数据加密

- 密码使用bcrypt哈希
- 敏感字段考虑使用pgcrypto扩展
- JWT令牌使用强随机生成

### 10.2 访问控制

- 基于RBAC的细粒度权限控制
- 设备指纹验证
- IP地址白名单/黑名单
- 登录频率限制

### 10.3 审计合规

- 完整的操作日志记录
- 数据变更追踪
- GDPR数据删除支持
- 定期安全扫描

## 11. 性能优化

### 11.1 索引策略

- 复合索引优化查询
- 部分索引减少存储
- 表达式索引支持特殊查询

### 11.2 查询优化

- 视图预计算常用统计
- 分区表提高大表查询
- 连接池优化连接管理

### 11.3 缓存策略

- Redis缓存热点数据
- 会话状态缓存
- 权限信息缓存

## 12. 扩展性设计

### 12.1 水平扩展

- 按省份分片用户数据
- 读写分离
- 数据库集群

### 12.2 垂直扩展

- 模块化表设计
- 微服务架构支持
- 异步处理

## 13. 监控告警

### 13.1 性能监控

- 查询性能分析
- 连接池监控
- 死锁检测

### 13.2 业务监控

- 注册转化率
- 会员续费率
- 异常登录检测

## 14. 数据备份恢复

### 14.1 备份策略

- 每日全量备份
- 实时WAL备份
- 跨地域备份

### 14.2 恢复测试

- 定期恢复演练
- RTO/RPO指标
- 灾难恢复预案

## 总结

本数据库设计充分考虑了高考志愿填报系统的业务需求，实现了完整的用户管理、权限控制、会员系统、设备管理等功能。通过合理的表结构设计、索引优化、分区策略和安全措施，确保系统的高性能、高可用性和高安全性。

设计遵循PostgreSQL最佳实践，支持水平扩展和垂直扩展，具备良好的可维护性和扩展性，能够满足系统长期发展的需求。