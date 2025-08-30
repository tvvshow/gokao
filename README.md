# 高考志愿填报系统 (College Entrance Exam Application)

一个基于AI的高考志愿填报和职业规划助手系统，为中国高中学生、家长和老师提供智能化的大学申请解决方案。

## 系统架构

- **后端**: FastAPI + Python (微服务架构)
- **数据库**: PostgreSQL + Redis
- **AI/ML**: XGBoost/LightGBM + 特征存储
- **前端**: React/Next.js + Ant Design
- **移动端**: Flutter
- **高性能计算**: C++ gRPC服务
- **基础设施**: Docker + Kubernetes

## 核心功能

- 🎯 AI智能推荐 (冲稳保方案生成)
- 🏫 院校专业查询 (2700+高校, 1400+专业)
- 📊 就业前景分析
- 🎲 志愿模拟填报
- 🔍 智能推荐 (就近、热门、潜力专业)
- 📱 新高考政策支持 (3+1+2, 3+3模式)
- 💬 社区互动功能

## 快速开始

### 开发环境部署

```bash
# 克隆项目
git clone <repository-url>
cd college-entrance-exam-app

# 启动开发环境
docker-compose up -d

# 运行数据库迁移
python manage.py migrate

# 启动开发服务器
python manage.py runserver
```

### 项目结构

```
college-entrance-exam-app/
├── backend/                 # 后端服务
│   ├── app/                # FastAPI应用
│   ├── models/             # 数据模型
│   ├── services/           # 业务逻辑服务
│   ├── api/                # API路由
│   └── ml/                 # 机器学习模块
├── frontend/               # Web前端 (React/Next.js)
├── mobile/                 # 移动端 (Flutter)
├── cpp-scoring/            # C++高性能计算服务
├── infrastructure/         # 基础设施配置
│   ├── docker/            # Docker配置
│   ├── k8s/               # Kubernetes配置
│   └── monitoring/        # 监控配置
└── docs/                   # 文档
```

## 开发指南

详见 [开发文档](docs/development.md)

## 部署指南

详见 [部署文档](docs/deployment.md)

## 贡献指南

详见 [贡献指南](docs/contributing.md)

## 许可证

MIT License