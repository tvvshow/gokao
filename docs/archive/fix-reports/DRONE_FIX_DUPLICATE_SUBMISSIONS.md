# 🔧 修复Drone CI重复提交问题

## 问题描述

你的Drone CI出现重复提交问题，webhook显示同时收到了 `push` 和 `check_suite.completed` 两个事件，导致同一次代码推送触发两次CI/CD流程。

## 根本原因

1. **GitHub Webhook配置**：GitHub同时发送多个事件类型到Drone
2. **Drone Runner配置缺失**：drone-runner没有设置事件过滤，处理所有接收到的事件
3. **事件重复处理**：同一次推送被处理两次

## 解决方案

### 方法1：配置Drone Runner事件过滤（推荐）

在 `drone-runner` 服务中添加 `DRONE_LIMIT_EVENTS=push` 环境变量：

```yaml
drone-runner:
  environment:
    - DRONE_LIMIT_EVENTS=push  # 只处理push事件
```

### 方法2：修改GitHub Webhook配置

1. 进入GitHub仓库设置
2. 选择 Settings → Webhooks
3. 编辑现有webhook
4. 在事件选择中，只勾选 `push` 事件
5. 取消勾选 `check_suite` 相关事件

## 应用修复

### 步骤1：更新Drone配置

使用提供的 `drone-docker-compose.yml` 文件替换你当前的配置：

```bash
# 停止当前Drone服务
docker-compose down

# 使用新配置启动
docker-compose -f drone-docker-compose.yml up -d
```

### 步骤2：验证修复

1. 推送一个测试提交：
   ```bash
   git add .
   git commit -m "test: 验证Drone CI修复"
   git push origin master
   ```

2. 检查Drone UI，确认只有一个构建被触发

3. 检查webhook日志，确认只处理了push事件

## 配置文件对比

### 修复前（问题配置）
```yaml
drone-runner:
  environment:
    - DRONE_RPC_HOST=drone-server
    - DRONE_RPC_PROTO=http
    - DRONE_RPC_SECRET=6b6c389c06c7e33f3eab00bf1b483725
    - DRONE_RUNNER_NAME=my-runner
    - DRONE_RUNNER_CAPACITY=2
    # ❌ 缺少事件过滤，处理所有事件
```

### 修复后（正确配置）
```yaml
drone-runner:
  environment:
    - DRONE_RPC_HOST=drone-server
    - DRONE_RPC_PROTO=http
    - DRONE_RPC_SECRET=6b6c389c06c7e33f3eab00bf1b483725
    - DRONE_RUNNER_NAME=my-runner
    - DRONE_RUNNER_CAPACITY=2
    - DRONE_LIMIT_EVENTS=push  # ✅ 只处理push事件
```

## 验证修复效果

### 预期结果
- ✅ 每次push只触发一次CI/CD
- ✅ Webhook日志只显示push事件处理
- ✅ 构建历史中没有重复构建

### 如果问题仍然存在

1. **检查GitHub Webhook配置**：
   - 确保只选择了必要的事件
   - 检查是否有多个webhook端点

2. **检查Drone日志**：
   ```bash
   docker logs drone-runner
   docker logs drone-server
   ```

3. **重启Drone服务**：
   ```bash
   docker-compose -f drone-docker-compose.yml restart
   ```

## 其他优化建议

### .drone.yml 触发器配置

确保你的 `.drone.yml` 中的触发器配置正确：

```yaml
trigger:
  branch:
    - master
  event:
    - push  # 只监听push事件
```

### 监控和日志

- 定期检查Drone构建历史
- 监控webhook响应时间
- 设置构建失败通知

---

## 总结

通过添加 `DRONE_LIMIT_EVENTS=push` 环境变量，Drone Runner将只处理push事件，忽略其他GitHub事件，从而解决重复CI/CD提交问题。这是最简单有效的解决方案，不需要修改GitHub webhook配置。

**应用修复后，请测试推送代码验证问题是否解决。**