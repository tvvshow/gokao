#!/bin/bash

# 完整部署脚本 - 修复API路径问题
# 此脚本在本地WSL运行，将部署到远程服务器192.168.0.181

set -e

SERVER="pestxo@192.168.0.181"
PASSWORD="satanking"
REMOTE_DIR="~/gaokao"
FRONTEND_DIST="/mnt/d/mybitcoin/gaokao/frontend/dist"
NGINX_CONF="/mnt/d/mybitcoin/gaokao/nginx-final.conf"

echo "=== 开始部署到 192.168.0.181 ==="

# 1. 备份远程前端和Nginx配置
echo "1. 备份远程文件..."
sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no $SERVER "
cd $REMOTE_DIR
mkdir -p backups
sudo cp /etc/nginx/sites-available/gaokao backups/gaokao-backup-\$(date +%Y%m%d-%H%M%S).conf
tar -czf backups/frontend-backup-\$(date +%Y%m%d-%H%M%S).tar.gz /var/www/gaokao/ 2>/dev/null || true
"

# 2. 上传新的Nginx配置
echo "2. 上传Nginx配置..."
sshpass -p "$PASSWORD" scp -o StrictHostKeyChecking=no $NGINX_CONF $SERVER:$REMOTE_DIR/gaokao-nginx.conf

# 3. 上传前端构建产物
echo "3. 上传前端构建产物..."
sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no $SERVER "
rm -rf /var/www/gaokao/*
mkdir -p /var/www/gaokao
"

sshpass -p "$PASSWORD" scp -o StrictHostKeyChecking=no -r $FRONTEND_DIST/* $SERVER:/var/www/gaokao/

# 4. 安装Nginx配置并重载
echo "4. 安装并重载Nginx..."
sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no $SERVER "
cd $REMOTE_DIR
sudo tee /etc/nginx/sites-available/gaokao > /dev/null < gaokao-nginx.conf
sudo nginx -t && sudo systemctl reload nginx
echo '✓ Nginx配置已更新'
"

# 5. 验证部署
echo "5. 验证部署..."
sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no $SERVER "
echo '=== 前端文件 ==='
ls -lh /var/www/gaokao/ | head -10

echo ''
echo '=== Nginx状态 ==='
sudo systemctl status nginx | head -5

echo ''
echo '=== 测试API ==='
echo 'Data Service Health:'
curl -s http://localhost/api/data/api/v1/health | head -5

echo ''
echo '测试前端访问:'
curl -s http://localhost/ | grep -o '<title>.*</title>'
"

echo ""
echo "=== 部署完成 ==="
echo ""
echo "请访问以下地址测试："
echo "  - 本地: http://192.168.0.181"
echo "  - 远程: https://gaokao.pkuedu.eu.org"
echo ""
echo "API端点测试："
echo "  - 大学列表: curl http://192.168.0.181/api/data/api/v1/universities?page=1&page_size=5"
echo "  - 健康检查: curl http://192.168.0.181/api/data/api/v1/health"
