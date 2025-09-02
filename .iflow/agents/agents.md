---
name: Gaokao Assistant Agents
description: Multi-role AI agents workflow to design, develop, deploy and maintain a 高考填报助手 system
---

# 高考填报助手 Agents 工作流

## 角色 1: 产品设计师
目标：定义核心功能和用户体验。

**任务**：
- 明确用户目标：高考分数查询、志愿推荐、录取概率分析。
- 绘制原型图：主页、分数查询页、志愿填报模拟页、结果分析页。
- 定义核心功能：账号登录、分数输入、批次筛选、推荐列表、收藏与对比。
- 产出：PRD 文档、界面原型图、用户流程图。

---

## 角色 2: 系统架构师
目标：制定整体架构。

**任务**：
- 采用微服务架构：爬虫服务、API 服务、前端 Web、数据库。
- 数据存储：PostgreSQL（录取分数线、院校信息）、Redis 缓存。
- 技术栈：后端 Python(FastAPI)，前端 React，容器化部署。  
- 安全性设计：数据校验、API 限流、用户认证(JWT/OAuth2)。

---

## 角色 3: 数据采集工程师
目标：获取权威的高校录取数据。

**任务**：
- 收集 31 个省考试院官网列表，建立采集字典。
- 优先解析 PDF/Excel 公告，保证数据权威。
- 定时任务检测更新（RSS/公告爬取）。
- 数据解析入库，统一格式化。

---

## 角色 4: Go 语言专家
目标：实现高性能爬虫。

**任务**：
- 开发爬虫框架，支持并发抓取。
- 针对不同省份编写解析器（HTML/PDF/Excel）。
- 错误重试与日志记录。
- 输出统一 JSON 数据。

---

## 角色 5: C++ 软件工程师
目标：实现高效的录取概率计算引擎。

**任务**：
- 输入：考生分数、省份、科类、批次。
- 输出：推荐志愿、录取概率（基于历年数据拟合）。
- 算法：分数段匹配 + 正态分布估算。
- 提供 C++ 动态库，暴露 gRPC/REST 接口。

---

## 角色 6: DevOps 工程师
目标：实现容器化与 CI/CD。

**任务**：
- 编写 `Dockerfile`：基于 `python:3.11-slim`，包含爬虫和 API。  
- 编写 `docker-compose.yml`：包含 `app`、`db(Postgres)`、`pgadmin`。  
- 配置 Drone CI，要求：  
  - 在 `drone.yml` 中定义流水线：构建 → 测试 → 构建镜像 → 部署。  
  - 支持缓存依赖，加快构建速度。  
  - 镜像推送到 DockerHub 或私有镜像仓库。  
- 提供部署说明文档。  

**示例 drone.yml：**

```yaml
kind: pipeline
type: docker
name: default

steps:
  - name: install-deps
    image: python:3.11-slim
    commands:
      - pip install -r requirements.txt

  - name: run-tests
    image: python:3.11-slim
    commands:
      - pytest --maxfail=1 --disable-warnings -q

  - name: build-docker
    image: plugins/docker
    settings:
      repo: your-dockerhub-username/gaokao-scores
      tags: latest
      username:
        from_secret: docker_username
      password:
        from_secret: docker_password

  - name: deploy
    image: appleboy/ssh-action
    settings:
      host: your-server-ip
      username: root
      password:
        from_secret: ssh_password
      script:
        - cd /opt/gaokao-scores
        - git pull origin main
        - docker-compose down
        - docker-compose up -d --build

trigger:
  branch:
    - main
```

---

## 角色 7: 前端工程师
目标：开发可交互的 Web 界面。

**任务**：
- React + TailwindCSS 开发 UI。
- 页面：登录注册、分数录入、志愿推荐结果、对比分析。
- 与后端 API 对接。
- 支持 PWA，方便在手机使用。

---

## 角色 8: 测试工程师
目标：保障质量。

**任务**：
- 单元测试：爬虫解析、API、算法。
- 集成测试：端到端流程。
- 压力测试：并发查询性能。
- 自动化测试接入 CI/CD。

---

## 角色 9: 运维工程师
目标：上线与监控。

**任务**：
- 部署至云服务器（阿里云/华为云）。
- 配置 Nginx + HTTPS。
- 监控：Prometheus + Grafana。
- 日志收集与告警：ELK/ Loki。

---

## 角色 10: 项目经理
目标：全局把控进度与交付。

**任务**：
- 建立 Jira/飞书任务看板。
- 定义里程碑（MVP → Beta → 正式版）。
- 定期 Review 和 Retrospective。
- 风险评估与资源协调。
