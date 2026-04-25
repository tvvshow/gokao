# API路径修复 - 部署指南

## 问题诊断

**核心问题**：后端服务API路由不一致
- data-service: `/v1` (无`/api`前缀)
- user-service: `/api/v1`
- recommendation-service: `/api/v1`

## 已完成的修复

### 1. 前端API路径统一
- ✅ 修改 `frontend/src/api/api-client.ts`
- ✅ 修改 `frontend/src/api/university.ts`
- ✅ 修改 `frontend/src/api/user.ts`
- ✅ 修改 `frontend/src/api/recommendation.ts`
- ✅ 修改 `frontend/src/views/MajorDetailPage.vue`

**统一路径格式**：`/api/{service}/api/v1/...`

### 2. Nginx配置
已创建 `nginx-final.conf`，路径重写规则：

```nginx
# Data Service
location /api/data/api/ {
    rewrite ^/api/data/api/(.*)$ /v1/$1 break;  # /api/data/api/v1/... -> /v1/...
    proxy_pass http://127.0.0.1:8082;
}

# User Service
location /api/user/api/ {
    rewrite ^/api/user/api/(.*)$ /api/v1/$1 break;
    proxy_pass http://127.0.0.1:8081;
}

# Recommendation Service
location /api/recommendation/api/ {
    rewrite ^/api/recommendation/api/(.*)$ /api/v1/$1 break;
    proxy_pass http://127.0.0.1:8083;
}
```

### 3. 前端构建
- ✅ 已完成生产构建：`frontend/dist/`

## 需要手动完成的步骤

由于远程服务器需要sudo权限，请在**192.168.0.181**服务器上执行以下操作：

### 步骤1：更新Nginx配置

```bash
# SSH登录到服务器
ssh pestxo@192.168.0.181

# 备份当前配置
sudo cp /etc/nginx/sites-available/gaokao ~/gaokao/nginx-backup-$(date +%Y%m%d).conf

# 复制新配置
sudo cp ~/gaokao/gaokao-nginx.conf /etc/nginx/sites-available/gaokao

# 测试配置
sudo nginx -t

# 重载Nginx
sudo systemctl reload nginx
```

### 步骤2：验证部署

```bash
# 测试data-service API
curl "http://localhost/api/data/api/v1/universities?page=1&page_size=3"

# 应该返回JSON数据，而不是404
```

### 步骤3：前端测试

访问以下地址验证：
- 本地：http://192.168.0.181
- 远程：https://gaokao.pkuedu.eu.org

检查浏览器控制台，API请求应该成功返回数据。

## 测试清单

### 数据完整性检查
```bash
# 检查数据库连接
docker exec -it gaokao-postgres psql -U gaokao -d gaokao_db -c "SELECT COUNT(*) FROM universities;"
docker exec -it gaokao-postgres psql -U gaokao -d gaokao_db -c "SELECT COUNT(*) FROM majors;"

# 预期结果
# universities: ~2700+
# majors: ~28000+
```

### API端点测试
```bash
# Data Service
curl "http://192.168.0.181/api/data/api/v1/universities?page=1&page_size=5"
curl "http://192.168.0.181/api/data/api/v1/majors?page=1&page_size=5"

# User Service
curl "http://192.168.0.181/api/user/api/v1/users/profile"
# (需要Bearer token)

# Recommendation Service
curl -X POST "http://192.168.0.181/api/recommendation/api/v1/recommendations/generate"
# (需要POST数据)
```

## 前端UI布局检查清单

### 已知问题
1. **响应式布局** - 移动端适配
2. **元素对齐** - 卡片和按钮对齐
3. **颜色对比度** - 深色模式下的可读性
4. **表单验证** - 输入验证和错误提示

### 测试页面
- [ ] 首页 (`/`)
- [ ] 院校查询 (`/universities`)
- [ ] 专业分析 (`/majors`)
- [ ] 智能推荐 (`/recommendation`)
- [ ] 数据分析 (`/analysis`)
- [ ] 个人中心 (`/profile`)

## 回滚方案

如果新配置有问题：

```bash
# 恢复Nginx配置
sudo cp ~/gaokao/nginx-backup-YYYYMMDD.conf /etc/nginx/sites-available/gaokao
sudo systemctl reload nginx

# 恢复前端（如果有备份）
cd ~/gaokao/backups
tar -xzf frontend-backup-YYYYMMDD-HHMMSS.tar.gz -C /var/www/gaokao/
```

## 后续优化建议

1. **统一后端路由**：修改data-service使用`/api/v1`前缀
2. **API Gateway**：考虑使用API Gateway统一路由
3. **健康检查**：添加更完善的健康检查端点
4. **监控**：添加API监控和日志聚合

## 文件位置

- Nginx配置：`/mnt/d/mybitcoin/gaokao/nginx-final.conf`
- 前端构建：`/mnt/d/mybitcoin/gaokao/frontend/dist/`
- 部署脚本：`/mnt/d/mybitcoin/gaokao/deploy-api-fix.sh`
- 本文档：`/mnt/d/mybitcoin/gaokao/API_FIX_GUIDE.md`

---

**创建时间**：2026-01-20
**状态**：等待手动部署Nginx配置
