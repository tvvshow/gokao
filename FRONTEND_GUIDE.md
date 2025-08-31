# 🎓 高考志愿填报助手 - 前端界面使用指南

## 🎉 重要更新：用户友好的Web界面已完成！

现在您可以通过现代化的Web界面使用高考志愿填报助手，不再需要只使用API文档界面。

## 🚀 快速启动

### 方式一：一键启动全栈系统
```bash
# 启动完整系统（推荐）
./scripts/start-full-stack.sh
```

### 方式二：分别启动
```bash
# 1. 启动后端服务
docker-compose up -d postgres redis data-service api-gateway

# 2. 启动前端界面
./scripts/start-frontend.sh
```

### 方式三：Docker完整部署
```bash
# 使用Docker启动所有服务
docker-compose up -d
```

## 🌐 访问地址

启动成功后，您可以访问：

- **🏠 前端界面**: http://localhost:3000 ← **用户使用这个**
- **🔧 API文档**: http://localhost:8080/swagger/index.html ← 开发者使用
- **🗄️ 数据服务**: http://localhost:8082

## 📱 界面功能

### 🏠 首页 (http://localhost:3000)
- 系统介绍和功能预览
- 快速成绩查询入口
- 热门院校展示
- 统计数据可视化

### 🔍 院校查询 (http://localhost:3000/universities)
- 🔎 多条件搜索筛选（名称、省份、类型、层次）
- 📊 高级筛选（分数线、排名、学校规模）
- 🎯 院校详情查看
- ❤️ 收藏功能
- 📈 对比分析

### 🤖 智能推荐 (http://localhost:3000/recommendation)
- 📝 考生信息录入（分数、省份、文理科）
- 🎯 个性化偏好设置（地区、专业、院校类型）
- 🎲 风险承受度选择（保守型/稳健型/激进型）
- 🏆 AI推荐结果展示
- 📊 录取概率预测
- 📄 推荐报告导出

### 📊 数据分析 (http://localhost:3000/analysis)
- 📈 历史录取分数线趋势
- 🎯 录取概率预测
- ⚠️ 风险评估报告
- 📉 专业就业前景分析

### 👤 用户中心 (http://localhost:3000/profile)
- 👤 用户注册登录
- 🏷️ 个人信息管理
- 💎 会员服务
- 📋 历史记录

## 🎯 使用流程

### 对于普通用户：
1. 访问 http://localhost:3000
2. 点击"开始智能推荐"
3. 填写考生基本信息（分数、省份、文理科）
4. 设置个人偏好（意向地区、专业类别、风险承受度）
5. 点击"生成智能推荐"
6. 查看AI推荐的院校列表
7. 可以收藏、对比感兴趣的院校
8. 导出推荐报告

### 对于开发者：
- API文档：http://localhost:8080/swagger/index.html
- 直接调用后端API接口进行开发测试

## 🛠️ 技术架构

```
🌐 用户界面 (Vue.js 3 + Element Plus)
     ↓ HTTP API
🚪 API网关 (:8080)
     ↓ 服务调用
🔧 数据服务 (:8082) → 🗄️ PostgreSQL + 📦 Redis
🤖 推荐服务 (:8083) → 🧠 C++算法引擎
💳 支付服务 (:8084)
👤 用户服务 (:8081)
```

## 🎨 界面特色

- ✨ 现代化设计，响应式布局
- 📱 支持桌面端、平板端、移动端
- 🎯 直观的用户体验
- 📊 丰富的数据可视化
- ⚡ 快速响应，支持实时搜索
- 🔒 安全的用户认证
- 💾 本地数据缓存

## 🔧 开发模式

如果您想修改前端界面：

```bash
cd frontend
npm install
npm run dev
```

这将启动开发服务器，支持热更新。

## 📞 帮助与支持

如果遇到问题：

1. 检查所有服务是否正常启动：`docker-compose ps`
2. 查看服务日志：`docker-compose logs [service-name]`
3. 重启服务：`docker-compose restart`
4. 完全重新部署：`docker-compose down && docker-compose up -d`

---

**🎊 现在您已经拥有了完整的用户界面，不再需要使用开发者API文档进行操作！**