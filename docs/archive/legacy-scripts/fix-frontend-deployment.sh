#!/bin/bash

# 高考志愿填报系统前端部署修复脚本
# 远端服务器: 192.168.0.181

set -e

echo "=========================================="
echo "开始修复前端部署问题"
echo "=========================================="

# SSH连接信息
SSH_HOST="pestxo@192.168.0.181"
SSH_PASS="satanking"

echo "1. 修复前端环境变量配置..."
sshpass -p "$SSH_PASS" ssh -o StrictHostKeyChecking=no "$SSH_HOST" << 'EOF'
# 设置前端环境变量
cat > /home/pestxo/gaokao/.env.production << 'ENV_EOF'
# 生产环境配置 - 前端
VITE_API_BASE_URL=https://gaokao.pkuedu.eu.org
VITE_APP_TITLE=高考志愿填报助手
VITE_APP_DESCRIPTION=智能推荐系统，助您选择理想大学
VITE_HOST=gaokao.pkuedu.eu.org
NODE_ENV=production
ENV_EOF

echo "前端环境变量配置完成"
EOF

echo "2. 重新构建前端..."
sshpass -p "$SSH_PASS" ssh -o StrictHostKeyChecking=no "$SSH_HOST" "cd /home/pestxo/gaokao/frontend && npm run build:prod"

echo "3. 复制前端文件到网站目录..."
sshpass -p "$SSH_PASS" ssh -o StrictHostKeyChecking=no "$SSH_HOST" "sudo rm -rf /var/www/gaokao/* && sudo cp -r /home/pestxo/gaokao/frontend/dist/* /var/www/gaokao/ && sudo chown -R www-data:www-data /var/www/gaokao/"

echo "4. 修复Nginx配置..."
sshpass -p "$SSH_PASS" ssh -o StrictHostKeyChecking=no "$SSH_HOST" << 'EOF'
# 创建新的Nginx配置
sudo cat > /etc/nginx/sites-available/gaokao << 'NGINX_EOF'
server {
    listen 80;
    server_name gaokao.pkuedu.eu.org localhost;

    root /var/www/gaokao;
    index index.html;

    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/javascript application/javascript application/json application/xml;

    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;

    location /assets {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    location / {
        try_files $uri $uri/ /index.html;
    }

    # API Gateway - /api/* -> port 8080/*
    location /api/ {
        proxy_pass http://127.0.0.1:8080/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # CORS配置
        add_header 'Access-Control-Allow-Origin' '*' always;
        add_header 'Access-Control-Allow-Methods' 'GET, POST, PUT, DELETE, OPTIONS' always;
        add_header 'Access-Control-Allow-Headers' 'DNT,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization' always;

        if ($request_method = 'OPTIONS') {
            add_header 'Access-Control-Allow-Origin' '*';
            add_header 'Access-Control-Allow-Methods' 'GET, POST, PUT, DELETE, OPTIONS';
            add_header 'Access-Control-Allow-Headers' 'DNT,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization';
            add_header 'Access-Control-Max-Age' 1728000;
            add_header 'Content-Type' 'text/plain; charset=utf-8';
            add_header 'Content-Length' 0;
            return 204;
        }
    }

    # 健康检查
    location /health {
        access_log off;
        return 200 "healthy\n";
        add_header Content-Type text/plain;
    }
}
NGINX_EOF

echo "Nginx配置更新完成"
EOF

echo "5. 重启Nginx服务..."
sshpass -p "$SSH_PASS" ssh -o StrictHostKeyChecking=no "$SSH_HOST" "sudo nginx -t && sudo systemctl restart nginx"

echo "6. 验证修复结果..."
echo "测试前端访问..."
sshpass -p "$SSH_PASS" ssh -o StrictHostKeyChecking=no "$SSH_HOST" "curl -s -o /dev/null -w '%{http_code}' http://localhost"

echo "测试API代理..."
API_TEST=$(sshpass -p "$SSH_PASS" ssh -o StrictHostKeyChecking=no "$SSH_HOST" "curl -s -o /dev/null -w '%{http_code}' http://localhost/api/data/v1/universities?page=1&page_size=1")
echo "API状态码: $API_TEST"

echo "=========================================="
echo "修复完成！"
echo ""
echo "请访问以下地址测试："
echo "前端: https://gaokao.pkuedu.eu.org"
echo "API: https://gaokao.pkuedu.eu.org/api/data/v1/universities"
echo ""
echo "如果仍有问题，请检查："
echo "1. 服务状态: ~/gaokao/check-services.sh"
echo "2. 错误日志: tail -f ~/gaokao/logs/*.log"
echo "3. 网络连接: curl -v http://localhost/api/data/v1/universities"
echo "=========================================="