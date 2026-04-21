# Drone CI/CD 配置指南

## 🚀 Drone 自托管配置完成

您的高考志愿填报助手项目现在已经配置了完整的Drone CI/CD流水线。

## 📋 必需的 Secrets 配置

在Drone UI中需要配置以下秘钥：

### Docker Registry 配置
```bash
# Docker Hub或私有Registry凭据
docker_username: your_docker_username
docker_password: your_docker_password
```

### 部署环境配置
```bash
# 测试环境 (develop分支自动部署)
staging_host: staging.gaokaohub.com
staging_user: deploy
staging_key: |
  -----BEGIN OPENSSH PRIVATE KEY-----
  your_staging_private_key_here
  -----END OPENSSH PRIVATE KEY-----

# 生产环境 (tag部署)
production_host: gaokaohub.com
production_user: deploy
production_key: |
  -----BEGIN OPENSSH PRIVATE KEY-----
  your_production_private_key_here
  -----END OPENSSH PRIVATE KEY-----
```

### 通知配置
```bash
# Slack通知 (可选)
slack_webhook: https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK

# 钉钉通知 (可选)
dingtalk_webhook: https://oapi.dingtalk.com/robot/send?access_token=YOUR_TOKEN

# FOSSA许可证检查 (可选)
fossa_api_key: your_fossa_api_key
```

## 🔧 Drone服务器配置

### 1. 启用仓库
在Drone UI中：
1. 登录到您的Drone实例
2. 找到 `gaokaohub/gaokao` 仓库
3. 点击"ACTIVATE"激活仓库
4. 进入仓库设置页面

### 2. 配置Secrets
在仓库设置中添加上述所有必需的secrets：

```bash
# 在Drone UI中逐个添加，或使用CLI
drone secret add gaokaohub/gaokao docker_username your_username
drone secret add gaokaohub/gaokao docker_password your_password
# ... 添加其他secrets
```

### 3. 配置Webhooks (如果需要)
确保GitHub webhook指向您的Drone服务器：
- Payload URL: `https://your-drone-server.com/hook`
- Content type: `application/json`
- Events: Push, Pull request, Tag

## 🛠️ CI/CD 流水线说明

### 主流水线 (gaokaohub-ci)
1. **restore-cache** - 恢复Go模块和依赖缓存
2. **go-deps** - 下载Go依赖
3. **code-quality** - 代码质量检查 (golangci-lint, staticcheck)
4. **go-test** - Go单元测试和覆盖率
5. **cpp-build** - C++模块构建和测试
6. **security-scan** - 安全扫描 (gosec)
7. **build-images** - Docker镜像构建和推送
8. **image-security-scan** - 镜像安全扫描 (Trivy)
9. **deploy-staging** - 部署到测试环境 (develop分支)
10. **deploy-production** - 部署到生产环境 (tags)
11. **cleanup-and-notify** - 清理和通知
12. **rebuild-cache** - 重建缓存

### 通知流水线 (notify)
- 发送构建成功/失败通知到Slack

### 安全审计流水线 (security-audit)
- 每日定时运行
- 依赖漏洞检查
- 许可证合规检查

## 🎯 触发条件

### 自动触发
- **Push到master** → 完整CI + Docker镜像构建推送
- **Push到develop** → 完整CI + 部署到测试环境
- **Pull Request** → 代码检查和测试
- **创建Tag (v*)** → 完整CI + 部署到生产环境
- **定时任务** → 每日安全扫描

### 分支策略
- `master` - 生产分支，镜像推送到registry
- `develop` - 开发分支，自动部署到测试环境  
- `feature/*` - 功能分支，只运行测试

## 🚀 激活Drone CI/CD

### 1. 立即激活
```bash
# 如果有drone CLI
drone repo enable gaokaohub/gaokao

# 或在Drone UI中手动激活
```

### 2. 测试流水线
```bash
# 推送一个小的更改来测试
git add .
git commit -m "test: 激活Drone CI/CD"
git push origin master
```

### 3. 查看构建状态
访问您的Drone UI查看构建进度：
- `https://your-drone-server.com/gaokaohub/gaokao`

## 📊 构建状态徽章

添加构建状态徽章到README：
```markdown
[![Build Status](https://your-drone-server.com/api/badges/gaokaohub/gaokao/status.svg)](https://your-drone-server.com/gaokaohub/gaokao)
```

## 🔧 高级配置

### 自定义构建环境
可以修改 `.drone.yml` 中的镜像版本：
```yaml
environment:
  GO_VERSION: "1.21"  # Go版本
  DOCKER_BUILDKIT: "1"  # 启用Docker BuildKit
```

### 并行构建
当前配置支持：
- Go服务并行测试
- C++模块并行构建
- Docker镜像并行构建

### 缓存优化
- Go模块缓存: `./go/pkg/mod`
- Node模块缓存: `./node_modules`
- 构建缓存: `./.cache`

## 📱 监控和通知

### 构建通知
- ✅ 成功通知 → Slack + 钉钉
- ❌ 失败通知 → Slack
- 📊 构建统计 → Drone仪表板

### 安全监控
- 🔍 代码安全扫描 (gosec)
- 🐳 镜像漏洞扫描 (Trivy)
- 📦 依赖漏洞检查 (Nancy)
- ⚖️ 许可证合规 (FOSSA)

## 🆘 故障排除

### 常见问题
1. **构建失败** - 检查日志，通常是依赖或测试问题
2. **部署失败** - 检查SSH密钥和目标服务器连接
3. **镜像推送失败** - 检查Docker凭据配置
4. **通知不工作** - 检查Webhook URL配置

### 调试命令
```bash
# 检查Drone配置
drone lint .drone.yml

# 查看构建日志
drone build logs gaokaohub/gaokao <build-number>

# 重新触发构建
drone build restart gaokaohub/gaokao <build-number>
```

---

🎉 **您的Drone CI/CD现在已完全配置并准备就绪！**

只需在Drone UI中激活仓库并配置secrets，然后推送代码即可开始自动化构建和部署。