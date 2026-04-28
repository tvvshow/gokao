# 技术债务分析报告

## 1. 代码重复度

| 重复模式 | 位置 | 影响 |
|---------|------|------|
| JWT 认证逻辑 | `pkg/auth/middleware.go` vs `services/user-service/internal/middleware/permission.go` | 两套实现，后者未使用 |
| 配置加载 | 每个服务各自实现 `config/config.go` | 可提取到 `pkg/config` |
| 错误处理 | 部分服务仍有本地错误类型，未统一使用 `pkg/errors` | 不一致的错误响应格式 |

## 2. 死代码清单

| 文件 | 状态 | 建议 |
|------|------|------|
| `recommendation-service/pkg/cppbridge/hybrid_bridge.go` | legacycpp 分支 | ✅ 已删除 |
| `recommendation-service/pkg/cppbridge/memory_safe.go` | legacycpp 分支 | ✅ 已删除 |
| `user-service/internal/middleware/permission.go` 中的多个桩函数 | 硬编码返回，未使用 | 删除或实现 |

## 3. 性能瓶颈

| 优先级 | 问题 | 位置 | 影响 |
|--------|------|------|------|
| P1 | **Gateway 主文件过大** | `services/api-gateway/main.go` | 可维护性下降，评审成本高 |
| P1 | **循环内数据库查询 (N+1)** | `data-service/internal/services/algorithm_service.go:MatchVolunteers` | 对每个院校循环查库 |
| P1 | **限流策略全局硬编码** | `api-gateway/main.go:57-58` | 10 req/s 不适合所有路由 |
| P2 | **支付列表 API 参数设计不合理** | `payment-service/internal/handlers/payment_handler.go` | `user_id` 解析脆弱，前端难直接消费 |

## 4. 算法能效与改进建议

**当前状态**：
- 推荐算法：`cppengine` 路径已生效，支持 C++ 引擎优先 + Go 回退
- 录取预测：线性回归/分位数，计算密集但未缓存
- legacycpp 分支已清理，构建路径单一化

**改进建议**：
1. 维持 `cppengine` 单路径，避免引入新的并行旧实现
2. 预计算 + 位图索引：按分数段、省份、学科预建索引
3. `MatchVolunteers` 使用 goroutine 池并行处理
4. 热门查询结果预加载到 Redis

## 5. 上线差距评估

### Critical（阻塞上线）

| 差距 | 解决方案 |
|------|----------|
| 无 |

### High

| 差距 | 解决方案 |
|------|----------|
| 限流策略粗糙 | 按路由差异化配置 |
| 无结构化日志/链路追踪 | 集成 OpenTelemetry |
| 错误处理不统一 | 全服务迁移到 `pkg/errors` |

### Medium

| 差距 | 解决方案 |
|------|----------|
| 测试覆盖率不足 | 核心算法补充单元测试，目标 70% |
| API 文档不完整 | Swagger 同步更新新路由 |
| 压力测试未执行 | 执行压测验证 P99 < 500ms |

## 6. 后续推进方向

### 短期（1-2 天）
- 统一前后端响应格式（逐步消除双解析）
- 支付与会员数据从 localStorage 迁移到后端主存储
- 清理历史审计文档中的过时结论

### 中期（1 周）
- 为 `MatchVolunteers` 添加 goroutine 并行处理
- 配置数据库连接池
- 集成 OpenTelemetry

### 长期（2 周）
- 压力测试（目标：10k 并发，P99 < 500ms）
- 补充 API 文档
- K8s 部署清单编写

## 7. 总结

| 维度 | 状态 | 说明 |
|------|------|------|
| 主功能路径 | ✅ 基本完整 | 推荐生成、用户认证、数据查询可走通 |
| 认证授权 | ✅ 已生效 | 关键路由已挂载 JWT，推荐路由强制鉴权 |
| 推荐性能 | ✅ 已改善 | C++ 引擎路径已启用，保留回退机制 |
| 可观测性 | ⚠️ 基础 | 有 Prometheus metrics，无链路追踪 |
| 代码质量 | ⚠️ 中等 | 主文件偏大，前端仍有本地存储债务 |

**结论**：系统已跨过阻塞上线门槛，当前重点是“数据一致性与工程质量”治理。
