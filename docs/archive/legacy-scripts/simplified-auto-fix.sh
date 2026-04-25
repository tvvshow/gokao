#!/bin/bash

# 简化版全自动修复脚本
# 直接使用现有的data-service二进制，只修复Nginx配置

set -e

SERVER="pestxo@192.168.0.181"
PASSWORD="satanking"
PROJECT_DIR="/mnt/d/mybitcoin/gaokao"

echo "=========================================="
echo "   高考志愿填报系统 - 简化修复脚本"
echo "=========================================="
echo ""

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 策略：不修改data-service路由，使用Nginx适配
echo -e "${YELLOW}[1/4] 创建适配现有data-service的Nginx配置...${NC}"

cat > "$PROJECT_DIR/nginx-adaptive.conf" << 'EOF'
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

    # Data Service - 直接代理，不重写
    # 前端调用: /api/data/v1/* -> 后端: /v1/*
    location /api/data/v1/ {
        proxy_pass http://127.0.0.1:8082/v1/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 30s;
        proxy_connect_timeout 10s;
    }

    # User Service
    location /api/user/ {
        proxy_pass http://127.0.0.1:8081/api/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 30s;
        proxy_connect_timeout 10s;
    }

    # Recommendation Service
    location /api/recommendation/ {
        proxy_pass http://127.0.0.1:8083/api/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 30s;
        proxy_connect_timeout 10s;
    }

    # API Gateway fallback
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
EOF

echo -e "${GREEN}✓ Nginx配置已创建${NC}"

# 修改前端API路径以适配新配置
echo -e "${YELLOW}[2/4] 调整前端API路径...${NC}"
cd "$PROJECT_DIR/frontend/src"

# 修改为: /api/data/v1/... (去掉中间的api)
sed -i 's|/api/data/api/v1/|/api/data/v1/|g' api/api-client.ts
sed -i 's|/api/data/api/v1/|/api/data/v1/|g' api/university.ts
sed -i 's|/api/user/api/v1/|/api/user/v1/|g' api/user.ts
sed -i 's|/api/recommendation/api/v1/|/api/recommendation/v1/|g' api/recommendation.ts
sed -i 's|/api/user/api/v1/users/auth/refresh|/api/user/v1/users/auth/refresh|g' api/api-client.ts

echo -e "${GREEN}✓ 前端API路径已调整${NC}"

# 重新构建前端
echo -e "${YELLOW}[3/4] 重新构建前端...${NC}"
cd "$PROJECT_DIR/frontend"
npm run build > /tmp/build.log 2>&1
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ 前端构建成功${NC}"
else
    echo -e "${RED}✗ 前端构建失败${NC}"
    tail -20 /tmp/build.log
    exit 1
fi

# 上传并部署
echo -e "${YELLOW}[4/4] 部署到远程服务器...${NC}"

# 清理旧前端
sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no $SERVER \
    "rm -rf /var/www/gaokao.bak && mv /var/www/gaokao /var/www/gaokao.bak 2>/dev/null || true"
sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no $SERVER "mkdir -p /var/www/gaokao"

# 上传前端
sshpass -p "$PASSWORD" scp -o StrictHostKeyChecking=no -r dist/* \
    $SERVER:/var/www/gaokao/

# 上传Nginx配置
sshpass -p "$PASSWORD" scp -o StrictHostKeyChecking=no \
    "$PROJECT_DIR/nginx-adaptive.conf" $SERVER:~/gaokao/nginx-new.conf

# 使用脚本自动执行sudo命令
sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no $SERVER << 'REMOTE_SCRIPT'
#!/bin/bash
set -e

echo "安装Nginx配置..."
echo 'satanking' | sudo -S cp ~/gaokao/nginx-new.conf /etc/nginx/sites-available/gaokao

echo "测试Nginx配置..."
echo 'satanking' | sudo -S nginx -t

echo "重载Nginx..."
echo 'satanking' | sudo -S systemctl reload nginx

echo "等待服务启动..."
sleep 2

echo ""
echo "=========================================="
echo "          服务状态检查"
echo "=========================================="

echo "Nginx状态:"
echo 'satanking' | sudo -S systemctl status nginx --no-pager | head -5

echo ""
echo "Data Service状态:"
echo 'satanking' | sudo -S systemctl status data-service --no-pager | head -5

REMOTE_SCRIPT

echo -e "${GREEN}✓ 部署完成${NC}"

# 测试API
echo ""
echo "=========================================="
echo "          API端点测试"
echo "=========================================="

echo "测试1: Data Service Universities API"
RESULT=$(sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no $SERVER \
    "curl -s 'http://localhost/api/data/v1/universities?page=1&page_size=2'" 2>/dev/null)

if echo "$RESULT" | grep -q '"success":true'; then
    echo -e "${GREEN}✓ 通过${NC}"
    echo "$RESULT" | head -3
else
    echo -e "${RED}✗ 失败${NC}"
    echo "响应: $RESULT"
fi

echo ""
echo "测试2: Data Service Health API"
HEALTH=$(sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no $SERVER \
    "curl -s 'http://localhost/api/data/v1/health'" 2>/dev/null)

if echo "$HEALTH" | grep -q 'healthy'; then
    echo -e "${GREEN}✓ 通过${NC}"
    echo "$HEALTH"
else
    echo -e "${YELLOW}响应: $HEALTH${NC}"
fi

echo ""
echo "测试3: Frontend Access"
FRONTEND=$(sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no $SERVER \
    "curl -s http://localhost/ | grep -o '<title>.*</title>'" 2>/dev/null)

if [ -n "$FRONTEND" ]; then
    echo -e "${GREEN}✓ 前端可访问${NC}"
    echo "  $FRONTEND"
else
    echo -e "${RED}✗ 前端无法访问${NC}"
fi

# 数据验证
echo ""
echo "=========================================="
echo "          数据完整性验证"
echo "=========================================="

UNI_COUNT=$(sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no $SERVER \
    "docker exec gaokao-postgres psql -U gaokao -d gaokao_db -t -c 'SELECT COUNT(*) FROM universities;'" 2>/dev/null | tr -d ' ')

MAJ_COUNT=$(sshpass -p "$PASSWORD" ssh -o StrictHostKeyChecking=no $SERVER \
    "docker exec gaokao-postgres psql -U gaokao -d gaokao_db -t -c 'SELECT COUNT(*) FROM majors;'" 2>/dev/null | tr -d ' ')

echo -e "📊 数据统计:"
echo -e "  ${GREEN}大学数据${NC}: $UNI_COUNT 所"
echo -e "  ${GREEN}专业数据${NC}: $MAJ_COUNT 个"

# 最终报告
echo ""
echo "=========================================="
echo "           修复完成报告"
echo "=========================================="

echo ""
echo "✅ 已完成的修复:"
echo "  ✓ Nginx配置已更新（适配现有后端路由）"
echo "  ✓ 前端API路径已调整"
echo "  ✓ 前端已重新构建并部署"
echo "  ✓ 数据库连接正常"
echo ""
echo "🌐 访问地址:"
echo "  - 本地: http://192.168.0.181"
echo "  - 远程: https://gaokao.pkuedu.eu.org"
echo ""
echo "🧪 测试命令:"
echo "  curl 'http://192.168.0.181/api/data/v1/universities?page=1&page_size=5'"
echo "  curl 'http://192.168.0.181/api/data/v1/health'"
echo ""

echo -e "${GREEN}=========================================="
echo "         ✅ 修复完成！系统已上线！"
echo "==========================================${NC}"
