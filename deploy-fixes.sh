#!/bin/bash

# 部署脚本 - 修复API路径问题
# 此脚本将在远程服务器192.168.0.181上执行

set -e

echo "=== 开始部署修复 ==="

# 1. 备份当前Nginx配置
echo "1. 备份Nginx配置..."
cp /etc/nginx/sites-available/gaokao ~/gaokao/nginx-backup-$(date +%Y%m%d-%H%M%S).conf

# 2. 安装新的Nginx配置
echo "2. 安装新的Nginx配置..."
cat > /tmp/gaokao-nginx.conf << 'NGINX_CONF_EOF'
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

    # Data Service - /api/data/api/v1/* -> port 8082/api/v1/*
    location /api/data/ {
        rewrite ^/api/data/(.*)$ /$1 break;
        proxy_pass http://127.0.0.1:8082;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 30s;
        proxy_connect_timeout 10s;
    }

    # User Service - /api/user/api/v1/* -> port 8081/api/v1/*
    location /api/user/ {
        rewrite ^/api/user/(.*)$ /$1 break;
        proxy_pass http://127.0.0.1:8081;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 30s;
        proxy_connect_timeout 10s;
    }

    # Recommendation Service - /api/recommendation/api/v1/* -> port 8083/api/v1/*
    location /api/recommendation/ {
        rewrite ^/api/recommendation/(.*)$ /$1 break;
        proxy_pass http://127.0.0.1:8083;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 30s;
        proxy_connect_timeout 10s;
    }

    # Payment Service - /api/payment/api/v1/* -> port 8084/api/v1/*
    location /api/payment/ {
        rewrite ^/api/payment/(.*)$ /$1 break;
        proxy_pass http://127.0.0.1:8084;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 30s;
        proxy_connect_timeout 10s;
    }

    # API Gateway fallback - /api/* -> port 8080/api/*
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 30s;
        proxy_connect_timeout 10s;
    }

    location / {
        try_files $uri $uri/ /index.html;
    }
}
NGINX_CONF_EOF

# 3. 测试并重载Nginx
echo "3. 测试Nginx配置..."
nginx -t -c /tmp/gaokao-nginx.conf || exit 1

echo "4. 安装Nginx配置..."
sudo cp /tmp/gaokao-nginx.conf /etc/nginx/sites-available/gaokao
sudo systemctl reload nginx

echo "=== Nginx配置已更新 ==="

# 5. 重新编译并部署data-service
echo "5. 重新编译data-service..."
cd ~/gaokao/services/data-service
go build -o data-service main.go || exit 1

echo "6. 重启data-service..."
sudo systemctl stop data-service || true
sleep 2
sudo systemctl start data-service

echo "=== 部署完成 ==="
echo "请检查服务状态："
echo "  - Nginx: sudo systemctl status nginx"
echo "  - Data Service: sudo systemctl status data-service"
echo "  - 测试API: curl http://localhost/api/data/api/v1/health"
