#!/bin/bash

# 前端部署脚本 - 部署修复后的构建到192.168.0.181服务器

echo "=== 开始部署前端到192.168.0.181 ==="

SERVER="192.168.0.181"
REMOTE_DIR="/var/www/gaokao"
LOCAL_DIST="/mnt/d/mybitcoin/gaokao/frontend/dist"

# 检查本地构建目录
if [ ! -d "$LOCAL_DIST" ]; then
    echo "错误：本地构建目录不存在: $LOCAL_DIST"
    exit 1
fi

echo "1. 备份服务器上的旧构建..."
ssh root@$SERVER "cd $REMOTE_DIR && tar -czf backup-$(date +%Y%m%d-%H%M%S).tar.gz * 2>/dev/null || true"

echo "2. 清理服务器上的旧文件..."
ssh root@$SERVER "rm -rf $REMOTE_DIR/*"

echo "3. 上传新构建..."
scp -r $LOCAL_DIST/* root@$SERVER:$REMOTE_DIR/

echo "4. 设置正确的文件权限..."
ssh root@$SERVER "chown -R www-data:www-data $REMOTE_DIR && chmod -R 755 $REMOTE_DIR"

echo "5. 重新加载Nginx配置..."
ssh root@$SERVER "nginx -t && systemctl reload nginx"

echo "=== 部署完成 ==="
echo "请访问 https://your-domain.com 测试前端功能"
