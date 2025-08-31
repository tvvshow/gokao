# 🎯 高考志愿填报助手 - 详细使用说明

## 📋 目录

1. [系统概述](#系统概述)
2. [系统架构](#系统架构)
3. [快速开始](#快速开始)
4. [功能使用指南](#功能使用指南)
5. [API接口说明](#api接口说明)
6. [常见问题](#常见问题)
7. [故障排除](#故障排除)

---

## 📖 系统概述

高考志愿填报助手是一个基于 **Go + C++ 混合架构** 的智能推荐系统，为高考考生提供：

### 🎯 核心功能
- **智能志愿推荐** - AI算法分析历史录取数据，个性化推荐最适合的院校和专业
- **录取概率预测** - 基于考生成绩和院校历史数据，精确预测录取概率
- **专业分析** - 详细的专业信息、就业前景、薪资水平分析
- **风险评估** - 多维度风险分析，帮助考生合理分配志愿
- **实时数据查询** - 最新的院校信息、招生计划、录取分数线

### 🏗️ 技术特色
- **微服务架构** - 高可用、可扩展的分布式系统
- **AI驱动** - C++高性能算法引擎 + 机器学习推荐
- **实时缓存** - Redis缓存提供毫秒级响应
- **安全防护** - 设备指纹、用户认证、数据加密
- **付费服务** - 灵活的会员制度和付费功能

---

## 🏗️ 系统架构

### 🔧 服务组件

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   用户端应用     │────│   API 网关      │────│   数据服务      │
│  (Web/Mobile)   │    │   (8080端口)    │    │   (8082端口)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │                        │
                       ┌─────────────────┐    ┌─────────────────┐
                       │   用户服务      │    │   推荐服务      │
                       │   (8081端口)    │    │   (8083端口)    │
                       └─────────────────┘    └─────────────────┘
                                │                        │
                       ┌─────────────────┐    ┌─────────────────┐
                       │   支付服务      │    │   C++算法引擎   │
                       │   (8084端口)    │    │  (设备指纹/AI)  │
                       └─────────────────┘    └─────────────────┘
                                │
                    ┌─────────────────────────────────┐
                    │         数据层                   │
                    │  PostgreSQL + Redis + ES        │
                    └─────────────────────────────────┘
```

### 🛠️ 核心组件说明

1. **API网关 (8080)** - 统一入口，负载均衡，限流防护
2. **数据服务 (8082)** - 院校、专业、录取数据查询
3. **用户服务 (8081)** - 用户注册、登录、权限管理
4. **推荐服务 (8083)** - AI算法推荐、志愿匹配
5. **支付服务 (8084)** - 会员订阅、付费功能
6. **C++模块** - 高性能算法、设备指纹、许可证管理

---

## 🚀 快速开始

### 📋 环境要求

- **操作系统**: Windows 10/11, macOS, Linux
- **Docker**: 20.0+ 
- **Docker Compose**: 2.0+
- **内存**: 4GB+ 
- **存储**: 10GB+

### ⚡ 一键启动

```bash
# 1. 克隆项目
git clone https://github.com/oktetopython/gaokao.git
cd gaokao

# 2. 启动所有服务
docker-compose up -d

# 3. 验证服务状态
curl http://localhost:8080/healthz
curl http://localhost:8082/health
```

### 🔍 验证安装

```bash
# 检查所有容器状态
docker-compose ps

# 应该看到以下服务正在运行:
# ✅ gaokao-api-gateway-1   (端口 8080)
# ✅ gaokao-data-service-1  (端口 8082) 
# ✅ gaokao-postgres-1      (数据库)
# ✅ gaokao-redis-1         (缓存)
```

---

## 📚 功能使用指南

### 1️⃣ 院校查询功能

#### 🔍 搜索院校
```bash
# 获取院校列表
curl "http://localhost:8082/api/v1/universities?page=1&pageSize=10"

# 搜索特定院校
curl "http://localhost:8082/api/v1/universities/search?name=北京&page=1"

# 按省份筛选
curl "http://localhost:8082/api/v1/universities?province=北京&page=1"
```

#### 📊 院校详情
```bash
# 获取院校详细信息
curl "http://localhost:8082/api/v1/universities/{university_id}"

# 获取院校统计信息
curl "http://localhost:8082/api/v1/universities/statistics"
```

### 2️⃣ 专业查询功能

#### 🎓 专业搜索
```bash
# 获取专业列表
curl "http://localhost:8082/api/v1/majors?page=1&pageSize=10"

# 搜索专业
curl "http://localhost:8082/api/v1/majors/search?name=计算机&page=1"

# 按学科分类
curl "http://localhost:8082/api/v1/majors?category=工学&page=1"
```

#### 💼 就业分析
```bash
# 获取专业就业数据
curl "http://localhost:8082/api/v1/majors/{major_id}/employment"

# 薪资水平分析
curl "http://localhost:8082/api/v1/majors/{major_id}/salary"
```

### 3️⃣ 录取数据分析

#### 📈 历史录取数据
```bash
# 获取录取数据
curl "http://localhost:8082/api/v1/admission/data?year=2023&university_id={id}"

# 录取趋势分析
curl "http://localhost:8082/api/v1/admission/analyze?university_id={id}&years=5"
```

#### 🎯 录取概率预测
```bash
# 预测录取概率
curl -X POST "http://localhost:8082/api/v1/admission/predict" \
  -H "Content-Type: application/json" \
  -d '{
    "score": 650,
    "province": "北京",
    "science_type": "理科",
    "university_id": "uuid",
    "major_id": "uuid"
  }'
```

### 4️⃣ 智能推荐系统

#### 🤖 AI志愿推荐
```bash
# 获取志愿推荐
curl -X POST "http://localhost:8082/api/v1/algorithm/match" \
  -H "Content-Type: application/json" \
  -d '{
    "score": 650,
    "province": "北京", 
    "science_type": "理科",
    "preferences": {
      "location": ["北京", "上海"],
      "major_categories": ["工学", "理学"],
      "risk_tolerance": "moderate"
    }
  }'
```

#### ⚖️ 风险评估
```bash
# 获取风险评估
curl "http://localhost:8082/api/v1/algorithm/risk-tolerance"

# 推荐类型
curl "http://localhost:8082/api/v1/algorithm/recommend-types"
```

### 5️⃣ 用户账户管理

#### 👤 用户注册登录
```bash
# 用户注册
curl -X POST "http://localhost:8081/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "student001",
    "email": "student@example.com",
    "password": "password123",
    "phone": "13800138000"
  }'

# 用户登录
curl -X POST "http://localhost:8081/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "student001",
    "password": "password123"
  }'
```

#### 💳 会员服务
```bash
# 获取会员信息
curl -H "Authorization: Bearer {token}" \
  "http://localhost:8081/api/v1/user/membership"

# 升级会员
curl -X POST -H "Authorization: Bearer {token}" \
  "http://localhost:8084/api/v1/payment/subscribe" \
  -d '{"plan": "premium", "duration": 12}'
```

---

## 🔌 API接口说明

### 🌐 基础信息

| 服务 | 端口 | 基础URL | 文档 |
|------|------|---------|------|
| API网关 | 8080 | http://localhost:8080 | /swagger |
| 数据服务 | 8082 | http://localhost:8082 | /swagger |
| 用户服务 | 8081 | http://localhost:8081 | /swagger |
| 推荐服务 | 8083 | http://localhost:8083 | /swagger |
| 支付服务 | 8084 | http://localhost:8084 | /swagger |

### 📝 通用响应格式

```json
{
  "success": true,
  "message": "操作成功",
  "data": {
    // 具体数据
  },
  "timestamp": 1693123456
}
```

### 🚨 错误响应格式

```json
{
  "success": false,
  "message": "错误描述",
  "error": {
    "code": "ERROR_CODE",
    "message": "详细错误信息"
  },
  "timestamp": 1693123456
}
```

### 🔑 认证方式

```bash
# 在请求头中携带JWT令牌
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

---

## ❓ 常见问题

### Q1: 如何获取API访问权限？
**A**: 用户需要先注册账户，免费用户有基础查询权限，高级功能需要升级会员。

### Q2: API有访问频率限制吗？
**A**: 是的，有以下限制：
- 免费用户: 100次/小时
- 基础会员: 1000次/小时  
- 高级会员: 10000次/小时

### Q3: 数据多久更新一次？
**A**: 
- 院校基础信息: 每年更新
- 录取分数线: 每年高考后更新
- 招生计划: 实时更新
- 就业数据: 每半年更新

### Q4: 支持哪些支付方式？
**A**: 支持微信支付、支付宝、银联卡等主流支付方式。

### Q5: 如何获取技术支持？
**A**: 
- 📧 邮箱: support@gaokaohub.com
- 💬 在线客服: 工作日 9:00-18:00
- 📱 技术QQ群: 123456789

---

## 🔧 故障排除

### 🐛 常见问题解决

#### 1️⃣ 服务无法启动
```bash
# 检查端口占用
netstat -tulpn | grep :8080

# 重启服务
docker-compose restart api-gateway

# 查看日志
docker-compose logs api-gateway
```

#### 2️⃣ 数据库连接失败
```bash
# 检查数据库状态
docker-compose ps postgres

# 重启数据库
docker-compose restart postgres

# 测试连接
docker exec gaokao-postgres-1 pg_isready -U gaokao_user
```

#### 3️⃣ API响应缓慢
```bash
# 检查性能指标
curl "http://localhost:8082/api/v1/performance/metrics"

# 清理缓存
curl -X POST "http://localhost:8082/api/v1/performance/clear-cache"

# 预热缓存
curl -X POST "http://localhost:8082/api/v1/performance/warmup-cache"
```

#### 4️⃣ 内存使用过高
```bash
# 检查容器资源使用
docker stats

# 重启高内存容器
docker-compose restart data-service

# 优化配置
# 在 docker-compose.yml 中调整内存限制
```

### 🛠️ 高级调试

#### 开启调试模式
```bash
# 设置环境变量
export GIN_MODE=debug
export LOG_LEVEL=debug

# 重启服务
docker-compose up -d
```

#### 性能监控
```bash
# 系统性能总览
curl "http://localhost:8082/api/v1/performance/summary"

# 缓存命中率
curl "http://localhost:8082/api/v1/performance/cache-stats"

# 重置性能指标
curl -X POST "http://localhost:8082/api/v1/performance/reset"
```

---

## 🎯 最佳实践

### 💡 性能优化建议

1. **合理使用分页**: 单次查询记录数不超过100条
2. **启用缓存**: 频繁查询的数据会被自动缓存
3. **批量查询**: 使用批量接口减少网络请求
4. **异步处理**: 大量数据处理使用异步接口

### 🔒 安全使用建议

1. **保护API密钥**: 不要在前端代码中暴露敏感信息
2. **使用HTTPS**: 生产环境必须使用加密连接
3. **及时更新**: 定期更新系统和依赖包
4. **监控异常**: 设置API调用监控和报警

### 📊 数据使用建议

1. **合理缓存**: 客户端缓存不变的基础数据
2. **增量更新**: 只获取变化的数据
3. **数据校验**: 对重要数据进行客户端校验
4. **备份策略**: 重要数据本地备份

---

## 📞 技术支持

### 🆘 联系方式
- **技术文档**: https://docs.gaokaohub.com
- **API文档**: http://localhost:8080/swagger
- **GitHub**: https://github.com/oktetopython/gaokao
- **邮箱支持**: tech-support@gaokaohub.com

### 📋 报告问题
提交问题时请提供：
1. 错误详细描述
2. 复现步骤
3. 系统环境信息
4. 相关日志信息

---

## 📄 更新日志

### v1.0.0 (2023-08-30)
- ✅ 完整的微服务架构
- ✅ AI智能推荐系统
- ✅ 用户管理和支付系统
- ✅ C++高性能算法引擎
- ✅ 完整的CI/CD流水线

---

**🎓 祝愿所有考生都能进入心仪的大学！**