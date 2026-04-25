#!/bin/bash

# 彻底修复高考志愿填报系统网络连接问题
# 远端服务器: 192.168.0.181

echo "=========================================="
echo "开始彻底修复网络连接问题"
echo "=========================================="

SSH_HOST="pestxo@192.168.0.181"
SSH_PASS="satanking"

echo "1. 修复前端环境变量..."
sshpass -p "$SSH_PASS" ssh -o StrictHostKeyChecking=no "$SSH_HOST" << 'EOF'
cd /home/pestxo/gaokao
# 设置前端环境变量
cat > .env.production << 'ENV_EOF'
VITE_API_BASE_URL=/api
VITE_APP_TITLE=高考志愿填报助手
VITE_APP_DESCRIPTION=智能推荐系统，助您选择理想大学
VITE_HOST=gaokao.pkuedu.eu.org
NODE_ENV=production
ENV_EOF

echo "前端环境变量设置完成"
EOF

echo "2. 重新构建前端..."
sshpass -p "$SSH_PASS" ssh -o StrictHostKeyChecking=no "$SSH_HOST" "cd /home/pestxo/gaokao/frontend && npm run build:prod"

echo "3. 彻底修复Nginx配置..."
sshpass -p "$SSH_PASS" ssh -o StrictHostKeyChecking=no "$SSH_HOST" << 'EOF'
sudo cp /etc/nginx/sites-available/gaokao /etc/nginx/sites-available/gaokao.backup.$(date +%s)

sudo cat > /etc/nginx/sites-available/gaokao << 'NGINX_EOF'
server {
    listen 80;
    server_name gaokao.pkuedu.eu.org localhost;

    root /var/www/gaokao;
    index index.html;

    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/javascript application/javascript application/json application/xml application/xml+rss;

    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # 静态资源缓存
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
        add_header Vary Accept-Encoding;
    }

    # API Gateway - /api/* -> port 8080/*
    location /api/ {
        proxy_pass http://127.0.0.1:8080/;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_set_header X-Forwarded-Host \$host;
        proxy_set_header X-Forwarded-Port \$server_port;

        # 超时设置
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;

        # CORS配置
        add_header 'Access-Control-Allow-Origin' '*' always;
        add_header 'Access-Control-Allow-Methods' 'GET, POST, PUT, DELETE, OPTIONS, PATCH' always;
        add_header 'Access-Control-Allow-Headers' 'DNT,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization,X-Auth-Token' always;

        # OPTIONS请求处理
        if (\$request_method = 'OPTIONS') {
            add_header 'Access-Control-Allow-Origin' '*';
            add_header 'Access-Control-Allow-Methods' 'GET, POST, PUT, DELETE, OPTIONS, PATCH';
            add_header 'Access-Control-Allow-Headers' 'DNT,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization,X-Auth-Token';
            add_header 'Access-Control-Max-Age' 1728000;
            add_header 'Content-Type' 'text/plain; charset=utf-8';
            add_header 'Content-Length' 0;
            return 204;
        }
    }

    # 特定服务代理（兼容旧配置）
    location /api/v1/data/ {
        rewrite ^/api/v1/data/(.*)$ /v1/data/\$1 break;
        proxy_pass http://127.0.0.1:8082/;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    location /api/v1/users/ {
        rewrite ^/api/v1/users/(.*)$ /v1/users/\$1 break;
        proxy_pass http://127.0.0.1:8081/;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    location /api/v1/recommendations/ {
        rewrite ^/api/v1/recommendations/(.*)$ /v1/recommendations/\$1 break;
        proxy_pass http://127.0.0.1:8083/;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    # 前端路由
    location / {
        try_files \$uri \$uri/ /index.html;
    }

    # 健康检查
    location /health {
        access_log off;
        return 200 "healthy\n";
        add_header Content-Type text/plain;
    }

    # 错误页面
    error_page 404 /index.html;
    error_page 500 502 503 504 /50x.html;
    location = /50x.html {
        root /usr/share/nginx/html;
    }
}
NGINX_EOF

echo "Nginx配置更新完成"
EOF

echo "4. 更新前端文件..."
sshpass -p "$SSH_PASS" ssh -o StrictHostKeyChecking=no "$SSH_HOST" "sudo rm -rf /var/www/gaokao/* && sudo cp -r /home/pestxo/gaokao/frontend/dist/* /var/www/gaokao/ && sudo chown -R www-data:www-data /var/www/gaokao/"

echo "5. 重启Nginx并测试..."
sshpass -p "$SSH_PASS" ssh -o StrictHostKeyChecking=no "$SSH_HOST" "sudo nginx -t && sudo systemctl restart nginx"

echo "6. 验证修复结果..."
# 测试前端页面
HTTP_CODE=$(sshpass -p "$SSH_PASS" ssh -o StrictHostKeyChecking=no "$SSH_HOST" "curl -s -o /dev/null -w '%{http_code}' http://localhost")
echo "前端页面状态码: $HTTP_CODE"

# 测试API代理
API_TEST=$(sshpass -p "$SSH_PASS" ssh -o StrictHostKeyChecking=no "$SSH_HOST" "curl -s -H 'Accept: application/json' -o /dev/null -w '%{http_code}' http://localhost/api/data/v1/universities?page=1&page_size=1")
echo "API代理状态码: $API_TEST"

# 测试API Gateway
GATEWAY_TEST=$(sshpass -p "$SSH_PASS" ssh -o StrictHostKeyChecking=no "$SSH_HOST" "curl -s -o /dev/null -w '%{http_code}' http://localhost:8080/api/v1/data/universities?page=1&page_size=1")
echo "API Gateway状态码: $GATEWAY_TEST"

echo "=========================================="
echo "修复完成！"
echo ""
echo "请访问以下地址测试："
echo "前端: https://gaokao.pkuedu.eu.org"
echo ""
echo "如果仍有问题，请执行以下命令检查服务状态："
echo "ssh pestxo@192.168.0.181"
echo "cd ~/gaokao"
echo "./check-services.sh"
echo ""
echo "API测试命令："
echo "curl -H 'Accept: application/json' https://gaokao.pkuedu.eu.org/api/data/v1/universities?page=1&page_size=1"
echo "=========================================="