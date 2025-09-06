# 高考志愿填报助手 - 前端界面

这是一个基于 Vue.js 3 + TypeScript + Element Plus 的现代化前端应用，为高考志愿填报系统提供用户友好的Web界面。

## 🚀 快速开始

### 安装依赖
```bash
npm install
# 或者
yarn install
```

### 开发环境
```bash
npm run dev
# 或者
yarn dev
```

访问 http://localhost:3000

### 生产构建
```bash
npm run build
# 或者
yarn build
```

### 预览构建结果
```bash
npm run preview
# 或者
yarn preview
```

## 📁 项目结构

```
frontend/
├── src/
│   ├── api/                # API接口
│   │   ├── index.ts        # axios配置
│   │   ├── user.ts         # 用户相关API
│   │   └── university.ts   # 院校相关API
│   ├── components/         # 公共组件
│   │   ├── AppHeader.vue   # 头部导航
│   │   ├── AppFooter.vue   # 页脚
│   │   └── UniversityCard.vue # 院校卡片
│   ├── router/             # 路由配置
│   │   └── index.ts
│   ├── stores/             # 状态管理
│   │   └── user.ts         # 用户状态
│   ├── types/              # TypeScript类型定义
│   │   ├── user.ts
│   │   └── university.ts
│   ├── views/              # 页面组件
│   │   ├── HomePage.vue    # 首页
│   │   ├── UniversitiesPage.vue # 院校查询
│   │   └── RecommendationPage.vue # 智能推荐
│   ├── App.vue             # 根组件
│   ├── main.ts             # 入口文件
│   └── style.css           # 全局样式
├── index.html              # HTML模板
├── package.json
├── vite.config.ts          # Vite配置
└── tsconfig.json           # TypeScript配置
```

## 🎯 主要功能

### 🏠 首页
- 系统介绍和功能预览
- 快速成绩查询入口
- 热门院校展示
- 统计数据展示

### 🔍 院校查询
- 多条件搜索和筛选
- 院校列表展示
- 院校详情查看
- 收藏和对比功能

### 🤖 智能推荐
- 考生信息录入
- 个性化偏好设置
- AI推荐结果展示
- 风险评估和概率预测

### 📊 数据分析
- 录取概率预测
- 历史分数线趋势
- 专业就业前景分析

### 👤 用户系统
- 用户注册登录
- 个人信息管理
- 会员服务

## 🛠️ 技术栈

- **框架**: Vue.js 3 + TypeScript
- **UI组件库**: Element Plus
- **构建工具**: Vite
- **状态管理**: Pinia
- **路由**: Vue Router 4
- **HTTP客户端**: Axios
- **图表库**: ECharts

## 🔧 开发指南

### API代理配置
开发环境下，前端会自动代理API请求到后端服务：
```typescript
// vite.config.ts
server: {
  proxy: {
    '/api': {
      target: 'http://localhost:8080',
      changeOrigin: true
    }
  }
}
```

### 环境变量
可以创建 `.env.local` 文件配置本地环境变量：
```
VITE_API_BASE_URL=http://localhost:8080
```

### 代码规范
- 使用 TypeScript 进行类型检查
- 遵循 Vue 3 Composition API 风格
- 使用 Element Plus 组件库
- 保持组件单一职责原则

## 📱 响应式设计

应用采用响应式设计，支持：
- 桌面端 (1200px+)
- 平板端 (768px-1200px)
- 移动端 (<768px)

## 🔗 与后端集成

前端通过 HTTP API 与后端服务通信：
- 用户服务: http://localhost:8081
- 数据服务: http://localhost:8082  
- 推荐服务: http://localhost:8083
- 支付服务: http://localhost:8084

## 🚀 部署

### Docker部署
```bash
# 构建镜像
docker build -t gaokao-frontend .

# 运行容器
docker run -p 3000:3000 gaokao-frontend
```

### 静态部署
构建后的 `dist` 目录可以部署到任何静态文件服务器（如 Nginx、Apache、CDN 等）。

## 📄 许可证

MIT License