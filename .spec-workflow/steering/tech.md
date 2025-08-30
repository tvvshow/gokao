# Technology Stack

## Project Type
**高考志愿填报专家系统** - 混合桌面与Web应用，采用Go+C++混合架构的企业级商业化高并发系统，包含AI推荐引擎、付费许可验证、安全加壳与防逆向工程。

## Core Technologies

### Primary Language(s)
- **Go 1.21+**: 主框架语言（70%）- API服务、业务逻辑、数据处理
- **C++ 20/23**: 核心算法语言（30%）- 匹配算法、AI推理、许可验证、付费功能
- **Runtime/Compiler**: GCC 11+/Clang 15+ for C++, Go compiler 1.21+
- **Language-specific tools**: Go modules, CMake 3.20+, garble (Go代码混淆), VMProtect (C++加壳)

### Key Dependencies/Libraries

**Go生态系统:**
- **Gin/Echo v4.9+**: 高性能Web框架与API路由
- **GORM v1.25+**: ORM数据库映射与迁移管理
- **go-redis v9.0+**: Redis缓存客户端
- **jwt-go v4.5+**: JWT身份认证与授权
- **validator v10.15+**: 输入参数验证
- **zap v1.25+**: 结构化日志记录
- **wire v0.5+**: 依赖注入框架

**C++生态系统:**
- **Eigen 3.4+**: 矩阵运算与数值计算（志愿匹配算法）
- **ONNX Runtime 1.16+**: AI模型推理引擎
- **nlohmann/json 3.11+**: JSON解析与序列化
- **SQLiteCpp 3.3+**: 轻量级数据库操作（本地缓存）
- **OpenSSL 3.0+**: 加密算法与证书验证
- **Boost 1.82+**: 通用C++库（字符串、文件系统、网络）
- **VMProtect SDK**: 商业代码保护与许可验证

### Application Architecture
**微服务架构 + 模块化设计**
- **Go微服务层**: API网关、用户服务、数据服务、支付服务、通知服务
- **C++核心模块**: 算法引擎、AI推理服务、许可验证服务、安全模块
- **跨语言通信**: gRPC (内部服务间), REST API (客户端交互), 共享内存 (高频调用)
- **模块间解耦**: 事件驱动架构，消息队列（Redis Streams）

### Data Storage
- **Primary storage**: PostgreSQL 15+ (主数据库) - 用户数据、学校信息、历史数据
- **Caching**: Redis 7.0+ (分布式缓存) - 会话、热点数据、实时排名
- **Search Engine**: Elasticsearch 8.9+ - 学校搜索、专业检索、智能推荐
- **Time Series (optional)**: InfluxDB 2.7+（仅用于高级分析报表，不参与运维监控；运维与系统指标统一由 Prometheus 提供）
- **Data formats**: Protocol Buffers (gRPC), JSON (REST API), MessagePack (高性能序列化)

### External Integrations
- **APIs**: 
  - 教育部官方API (学校数据同步)
  - 第三方支付接口 (微信支付、支付宝)
  - 短信通知服务 (阿里云SMS)
  - 邮件服务 (SendGrid/腾讯云)
- **Protocols**: HTTP/2, gRPC, WebSocket (实时通知), TCP/IP (内部通信)
- **Authentication**: OAuth 2.0, JWT, API Key, 双因子认证 (TOTP)

### Monitoring & Dashboard Technologies
- **Observability Stack**: OpenTelemetry 1.20+ (统一遥测标准), Prometheus 2.47+ (指标), Tempo/Jaeger (链路追踪), Loki (日志聚合)
- **Dashboard Framework**: Vue.js 3.3+ + Element Plus 2.3+ (管理后台), Grafana 10.1+ (运维看板)
- **Real-time Communication**: WebSocket + Server-Sent Events (实时通知)
- **Visualization Libraries**: ECharts 5.4+ (数据可视化), Chart.js 4.4+ (简单图表)
- **State Management**: Pinia 2.1+ (Vue状态管理)

## Development Environment

### Build & Development Tools
- **Build System**: 
  - Go: `go build` + Makefile + Docker multi-stage builds
  - C++: CMake 3.20+ + Conan 2.0+ (包管理)
- **Package Management**: Go modules, Conan (C++), npm (前端)
- **Development workflow**: 热重载 (Air for Go), 自动测试 (watch mode), Docker Compose开发环境

### Code Quality Tools
- **Static Analysis**: 
  - Go: `golangci-lint`, `go vet`, `staticcheck`
  - C++: `clang-tidy`, `cppcheck`, SonarQube
- **Formatting**: `gofmt`/`goimports`, `clang-format`
- **Testing Framework**: 
  - Go: `testing` + `testify`, `gomock` (模拟)
  - C++: Google Test + Google Mock
- **Documentation**: `godoc`, Doxygen (C++), Swagger/OpenAPI 3.0

### Version Control & Collaboration
- **VCS**: Git + GitLab Enterprise
- **Branching Strategy**: Git Flow (feature/develop/release/hotfix)
- **Code Review Process**: GitLab Merge Request + 至少2人审核 + CI/CD门禁

### Dashboard Development
- **Live Reload**: Vite 4.4+ (Vue热更新), Air (Go热重载)
- **Port Management**: 动态端口分配，环境变量配置
- **Multi-Instance Support**: Docker Compose + Nginx负载均衡

## Deployment & Distribution

- **Target Platform(s)**: 
  - 云端部署: 阿里云/腾讯云 Kubernetes集群
  - 桌面客户端: Windows 10+ (主要), macOS 12+ (次要)
- **Distribution Method**: 
  - SaaS Web应用 (主要收入来源)
  - 桌面客户端下载 (离线使用场景)
  - 学校定制版 (B2B销售)
- **Installation Requirements**: 
  - Web: 现代浏览器 (Chrome 90+, Firefox 88+, Safari 14+)
  - 桌面: 4GB RAM, 2GB存储空间, .NET Runtime 6.0+
- **Update Mechanism**: 
  - Web应用: 蓝绿部署，零停机更新
  - 桌面客户端: 自动更新服务，增量更新包

## Technical Requirements & Constraints

### Performance Requirements
- **响应时间**: API < 200ms (P99), 页面加载 < 2s
- **并发处理**: 10,000+ 同时在线用户，峰值 50,000+ QPS
- **内存使用**: Go服务 < 512MB per instance, C++模块 < 1GB
- **启动时间**: 服务启动 < 30s, 桌面应用 < 5s

### Compatibility Requirements
- **Platform Support**: Linux (Ubuntu 20.04+, CentOS 8+), Windows Server 2019+
- **Dependency Versions**: Go 1.21+, C++20标准, PostgreSQL 15+
- **Standards Compliance**: OAuth 2.0, OpenAPI 3.0, GDPR数据保护

### Security & Compliance
- **Security Requirements**: 
  - TLS 1.3加密传输
  - AES-256数据加密存储（KMS托管密钥，年度轮换；应用层 envelope encryption）
  - JWT + Refresh Token认证
  - Rate limiting防护 + WAF
  - C++代码VMProtect加壳保护
  - Go代码garble混淆
- **Compliance Standards**: 个人信息保护法合规, ISO 27001信息安全；支付场景不处理任何持卡人数据（WeChat/Alipay跳转+签名回调）
- **Data Protection**: PII最小化采集与分级；数据脱敏/掩码；分环境匿名化；数据保留与删除策略（默认1年，可配置）；敏感审计日志最小化
- **Key & Secrets**: 使用云KMS/HSM；密钥轮换年度、Secrets轮换季度；严格RBAC与访问审计
- **Threat Model**: 防逆向工程、防数据泄露、防DDoS攻击、防SQL注入、防供应链风险

### Scalability & Reliability
- **Expected Load**: 第一年10万注册用户，峰值期50万PV/日
- **Availability Requirements**: 99.9%可用性目标，RTO < 1小时，RPO < 15分钟
- **Growth Projections**: 3年内支持100万用户，水平扩展架构

## Technical Decisions & Rationale

### Decision Log
1. **Go+C++混合架构**: Go负责70%业务逻辑(开发效率高)，C++负责30%核心算法(性能优势+防逆向)，平衡商业化需求与开发成本
2. **微服务架构**: 支持团队并行开发，水平扩展，故障隔离，便于商业化定制
3. **VMProtect+garble双重保护**: C++核心算法强加壳保护知识产权，Go业务代码混淆防破解，确保商业价值
4. **PostgreSQL+Redis+Elasticsearch组合**: 关系数据+缓存+搜索的完整数据栈，支持复杂业务需求
5. **Kubernetes容器化部署**: 云原生架构，自动扩缩容，多云部署，降低运维复杂度

## Known Limitations

- **跨语言调用开销**: Go↔C++通信存在序列化开销，通过共享内存优化高频调用场景
- **C++构建复杂度**: 依赖管理和交叉编译相对复杂，通过Docker标准化构建环境解决
- **许可验证强依赖**: VMProtect加壳后调试困难，需要完善的日志和监控体系
- **初期开发成本**: 混合架构学习曲线陡峭，前期需要架构师指导和团队培训