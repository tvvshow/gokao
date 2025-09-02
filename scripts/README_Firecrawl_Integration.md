# Firecrawl MCP 高校数据爬取系统

## 项目概述

本项目实现了基于 Firecrawl MCP (Model Context Protocol) 的高校数据爬取系统，能够从全国31个省份的教育考试院网站自动爬取3000+所高校的详细信息。

## 系统架构

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Go 主控制器    │───▶│  Python MCP脚本   │───▶│  Firecrawl API  │
│ (并发管理/报告)  │    │ (实际数据爬取)    │    │ (网页内容提取)  │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   进度跟踪       │    │   JSON数据交换    │    │   结构化数据     │
│   错误处理       │    │   结果文件        │    │   智能提取       │
│   质量验证       │    │   配置管理        │    │   重试机制       │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## 核心特性

### 🚀 技术特性
- **混合架构**: Go主控制器 + Python MCP脚本
- **智能提取**: 使用Firecrawl的LLM能力进行结构化数据提取
- **并发控制**: 支持省份级别的并发爬取（默认5个并发）
- **错误处理**: 内置重试机制和错误恢复
- **增量爬取**: 支持断点续传和进度保存
- **数据验证**: 多层次的数据质量检查

### 📊 数据覆盖
- **31个省份**: 覆盖全国所有省、自治区、直辖市
- **3000+高校**: 包括985、211、本科、专科等各类院校
- **结构化数据**: 高校名称、类型、985/211状态、官网等

## 文件结构

```
scripts/
├── firecrawl_university_crawler.go    # Go主控制器
├── firecrawl_crawler.py               # Python MCP脚本
├── province_config.json               # 31个省份配置
├── test_firecrawl_integration.go      # 集成测试脚本
├── README_Firecrawl_Integration.md    # 本文档
└── 输出文件/
    ├── crawl_results.json             # 爬取结果
    ├── crawl_report.json              # 详细报告
    └── firecrawl_crawler.log          # 运行日志
```

## 环境要求

### 必需软件
- **Go 1.19+**: 主控制器运行环境
- **Python 3.8+**: MCP脚本运行环境
- **Firecrawl API Key**: 用于访问Firecrawl服务

### Python依赖
```bash
pip install requests
# 如果使用本地MCP服务器，还需要:
# pip install mcp-server-firecrawl
```

## 快速开始

### 1. 环境配置

```bash
# 设置Firecrawl API密钥
export FIRECRAWL_API_KEY="your_api_key_here"

# Windows用户使用:
set FIRECRAWL_API_KEY=your_api_key_here

# 可选：配置重试参数
export FIRECRAWL_RETRY_MAX_ATTEMPTS=3
export FIRECRAWL_RETRY_INITIAL_DELAY=1000
```

### 2. 系统测试

```bash
# 运行集成测试
go run test_firecrawl_integration.go
```

### 3. 单省份测试

```bash
# 测试单个省份（推荐先测试）
python firecrawl_crawler.py 北京
```

### 4. 全量爬取

```bash
# 运行完整爬取系统
go run firecrawl_university_crawler.go
```

## 配置说明

### 省份配置 (province_config.json)

```json
{
  "provinces": [
    {
      "name": "北京",
      "code": "BJ",
      "base_url": "https://www.bjeea.cn",
      "search_keywords": ["高校", "大学", "学院", "高等院校"]
    }
  ]
}
```

### 环境变量配置

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `FIRECRAWL_API_KEY` | Firecrawl API密钥 | 必需 |
| `FIRECRAWL_RETRY_MAX_ATTEMPTS` | 最大重试次数 | 3 |
| `FIRECRAWL_RETRY_INITIAL_DELAY` | 初始重试延迟(ms) | 1000 |
| `FIRECRAWL_CREDIT_WARNING_THRESHOLD` | 信用额度警告阈值 | 100 |

## 数据结构

### 高校信息结构

```go
type UniversityInfo struct {
    Name           string `json:"name"`            // 高校全称
    Province       string `json:"province"`        // 所在省份
    City           string `json:"city"`            // 所在城市
    UniversityType string `json:"university_type"` // 本科/专科/独立学院/民办
    Is985          bool   `json:"is_985"`          // 是否985高校
    Is211          bool   `json:"is_211"`          // 是否211高校
    Website        string `json:"website"`         // 官方网站
    Description    string `json:"description"`     // 高校简介
    EstablishedYear int   `json:"established_year"` // 建校年份
}
```

### 爬取结果结构

```go
type CrawlResult struct {
    Province        string           `json:"province"`         // 省份名称
    Success         bool             `json:"success"`          // 是否成功
    Universities    []UniversityInfo `json:"universities"`     // 高校列表
    ErrorMessage    string           `json:"error_message"`    // 错误信息
    UrlsDiscovered  int              `json:"urls_discovered"`  // 发现的URL数量
    ProcessingTime  float64          `json:"processing_time"`  // 处理时间(秒)
}
```

## 工作流程

### 三阶段爬取流程

1. **URL发现阶段**
   - 使用 `firecrawl_map` 工具
   - 发现各省教育考试院的高校信息页面
   - 过滤相关URL

2. **数据提取阶段**
   - 使用 `firecrawl_extract` 工具
   - 配合自定义JSON Schema
   - 智能提取结构化数据

3. **数据验证阶段**
   - 必填字段检查
   - 数据格式验证
   - 重复数据去除
   - 质量评分

### 并发控制策略

```
省份级并发 (最大5个)
├── 北京 ──┐
├── 上海 ──┤
├── 天津 ──┼── 并发处理
├── 重庆 ──┤
└── 河北 ──┘
     ⋮
```

## 监控和日志

### 日志文件
- `firecrawl_crawler.log`: Python脚本运行日志
- Go程序日志输出到控制台

### 监控指标
- 省份爬取成功率
- 高校数据完整性
- 平均处理时间
- 错误类型统计

## 故障排除

### 常见问题

1. **Python脚本调用失败**
   ```
   错误: 执行Python脚本失败
   解决: 检查Python环境和PATH配置
   ```

2. **Firecrawl API调用失败**
   ```
   错误: API密钥无效或额度不足
   解决: 检查FIRECRAWL_API_KEY设置
   ```

3. **数据提取质量差**
   ```
   原因: 网站结构变化或反爬虫机制
   解决: 调整JSON Schema或增加延迟
   ```

### 调试模式

```bash
# 启用详细日志
export FIRECRAWL_DEBUG=true

# 单省份调试
python firecrawl_crawler.py 北京
```

## 性能优化

### 建议配置
- **并发数**: 5个省份（避免触发反爬虫）
- **请求间隔**: 2-3秒（保护目标网站）
- **重试次数**: 3次（平衡成功率和效率）
- **超时设置**: 30秒（适应网络环境）

### 成本控制
- 使用 `maxAge` 参数启用缓存（500%性能提升）
- 设置信用额度警告阈值
- 优先爬取重点省份

## 扩展功能

### 未来计划
- [ ] 支持更多数据源（教育部官网等）
- [ ] 增加专业信息爬取
- [ ] 实现实时数据更新
- [ ] 添加数据可视化界面
- [ ] 支持自定义爬取规则

### 自定义开发

```python
# 扩展新的数据源
class CustomUniversityCrawler(FirecrawlUniversityCrawler):
    def discover_urls(self, province):
        # 自定义URL发现逻辑
        pass
    
    def extract_university_data(self, urls, province):
        # 自定义数据提取逻辑
        pass
```

## 许可证

本项目采用 MIT 许可证。详见 LICENSE 文件。

## 贡献指南

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 联系方式

如有问题或建议，请通过以下方式联系：
- 创建 GitHub Issue
- 发送邮件至项目维护者

---

**注意**: 请遵守目标网站的robots.txt规则和使用条款，合理使用爬虫功能。