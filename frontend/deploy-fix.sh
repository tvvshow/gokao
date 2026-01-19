#!/bin/bash
# Frontend API Path Fix Deployment Script
# 修复前端API路径双重前缀问题

set -e  # 遇到错误立即退出

echo "=========================================="
echo "开始修复并部署前端"
echo "=========================================="

# 进入项目目录
cd ~/gaokao/frontend

# 备份现有dist目录
echo "1. 备份现有dist目录..."
if [ -d "dist" ]; then
    mv dist dist.backup.$(date +%Y%m%d_%H%M%S)
fi

# 拉取最新代码
echo "2. 拉取最新代码..."
git pull origin master

# 确认 .env.production 配置正确
echo "3. 检查环境变量配置..."
cat > .env.production << 'EOF'
# Production Environment Variables
# 端点已包含完整路径 /api/v1/...，baseURL应为空
VITE_API_BASE_URL=
EOF
echo "环境变量配置已更新："
cat .env.production

# 重新构建前端
echo "4. 重新构建前端..."
npm run build

# 检查构建结果
if [ ! -d "dist" ]; then
    echo "错误：构建失败，dist目录不存在"
    exit 1
fi

echo "5. 构建成功，检查关键文件..."
ls -lh dist/
ls -lh dist/assets/ | head -n 5

# 重新加载Nginx（如需要）
echo "6. 重新加载Nginx配置..."
sudo nginx -t && sudo systemctl reload nginx

echo "=========================================="
echo "部署完成！"
echo "=========================================="
echo ""
echo "测试步骤："
echo "1. 打开浏览器访问: http://gaokao.pkuedu.eu.org"
echo "2. 打开开发者工具Network面板"
echo "3. 检查API请求路径应为: /api/v1/... （不是 /api/api/v1/...）"
echo "4. 首页应正常加载，没有'资源不存在'错误"
echo ""
echo "如果仍有问题，请检查："
echo "- Nginx日志: sudo tail -f /var/log/nginx/access.log"
echo "- 浏览器控制台错误信息"
