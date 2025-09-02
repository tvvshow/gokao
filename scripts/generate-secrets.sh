#!/bin/bash

# 🔐 高考志愿填报系统 - 安全密钥生成脚本
# 用于生成生产环境所需的各种密钥和证书

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONFIG_DIR="${PROJECT_ROOT}/config"
SECRETS_DIR="${CONFIG_DIR}/secrets"
CERTS_DIR="${CONFIG_DIR}/certs"

# 创建目录
mkdir -p "$SECRETS_DIR"
mkdir -p "$CERTS_DIR"

echo "🔐 开始生成安全密钥和证书..."

# ===========================================
# 生成JWT密钥
# ===========================================
echo "📝 生成JWT密钥..."
JWT_SECRET=$(openssl rand -base64 64 | tr -d '\n')
echo "JWT_SECRET=${JWT_SECRET}" > "${SECRETS_DIR}/jwt.env"
echo "✅ JWT密钥已生成: ${SECRETS_DIR}/jwt.env"

# ===========================================
# 生成数据库密码
# ===========================================
echo "🗄️ 生成数据库密码..."
DB_PASSWORD=$(openssl rand -base64 32 | tr -d '\n')
REDIS_PASSWORD=$(openssl rand -base64 32 | tr -d '\n')

cat > "${SECRETS_DIR}/database.env" << EOF
DB_PASSWORD=${DB_PASSWORD}
REDIS_PASSWORD=${REDIS_PASSWORD}
EOF
echo "✅ 数据库密码已生成: ${SECRETS_DIR}/database.env"

# ===========================================
# 生成AES加密密钥
# ===========================================
echo "🔒 生成AES加密密钥..."
AES_KEY=$(openssl rand -base64 32 | tr -d '\n')
echo "AES_ENCRYPTION_KEY=${AES_KEY}" > "${SECRETS_DIR}/encryption.env"
echo "✅ AES密钥已生成: ${SECRETS_DIR}/encryption.env"

# ===========================================
# 生成RSA密钥对
# ===========================================
echo "🔑 生成RSA密钥对..."
openssl genrsa -out "${CERTS_DIR}/gaokao-private.pem" 2048
openssl rsa -in "${CERTS_DIR}/gaokao-private.pem" -pubout -out "${CERTS_DIR}/gaokao-public.pem"
echo "✅ RSA密钥对已生成: ${CERTS_DIR}/"

# ===========================================
# 生成自签名SSL证书 (开发环境)
# ===========================================
echo "🌐 生成SSL证书 (开发环境)..."
openssl req -x509 -newkey rsa:2048 -keyout "${CERTS_DIR}/dev-private.key" \
    -out "${CERTS_DIR}/dev-certificate.crt" -days 365 -nodes \
    -subj "/C=CN/ST=Beijing/L=Beijing/O=GaokaoHub/CN=localhost"
echo "✅ SSL证书已生成: ${CERTS_DIR}/"

# ===========================================
# 生成API密钥
# ===========================================
echo "🔧 生成API密钥..."
API_SECRET=$(openssl rand -hex 32)
WEBHOOK_SECRET=$(openssl rand -hex 32)

cat > "${SECRETS_DIR}/api.env" << EOF
API_SECRET=${API_SECRET}
WEBHOOK_SECRET=${WEBHOOK_SECRET}
EOF
echo "✅ API密钥已生成: ${SECRETS_DIR}/api.env"

# ===========================================
# 生成会话密钥
# ===========================================
echo "👤 生成会话密钥..."
SESSION_SECRET=$(openssl rand -base64 64 | tr -d '\n')
CSRF_SECRET=$(openssl rand -base64 32 | tr -d '\n')

cat > "${SECRETS_DIR}/session.env" << EOF
SESSION_SECRET=${SESSION_SECRET}
CSRF_SECRET=${CSRF_SECRET}
EOF
echo "✅ 会话密钥已生成: ${SECRETS_DIR}/session.env"

# ===========================================
# 生成Docker Secrets文件
# ===========================================
echo "🐳 生成Docker Secrets..."
cat > "${SECRETS_DIR}/docker-secrets.yml" << EOF
version: '3.8'

secrets:
  jwt_secret:
    file: ./secrets/jwt.env
  db_password:
    file: ./secrets/database.env
  encryption_key:
    file: ./secrets/encryption.env
  api_secret:
    file: ./secrets/api.env
  session_secret:
    file: ./secrets/session.env
  ssl_private_key:
    file: ./certs/dev-private.key
  ssl_certificate:
    file: ./certs/dev-certificate.crt
  rsa_private_key:
    file: ./certs/gaokao-private.pem
  rsa_public_key:
    file: ./certs/gaokao-public.pem
EOF
echo "✅ Docker Secrets配置已生成: ${SECRETS_DIR}/docker-secrets.yml"

# ===========================================
# 生成环境变量文件
# ===========================================
echo "⚙️ 生成环境变量文件..."
cat > "${CONFIG_DIR}/.env.production" << EOF
# 🔒 生产环境配置 - 自动生成
# 生成时间: $(date)

# JWT配置
JWT_SECRET=${JWT_SECRET}
JWT_EXPIRE_HOURS=24
JWT_REFRESH_EXPIRE_HOURS=168

# 数据库配置
DB_HOST=postgres
DB_PORT=5432
DB_NAME=gaokao_db
DB_USER=gaokao_user
DB_PASSWORD=${DB_PASSWORD}
DB_SSL_MODE=require

# Redis配置
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=${REDIS_PASSWORD}
REDIS_DB=0

# 加密配置
AES_ENCRYPTION_KEY=${AES_KEY}
RSA_PRIVATE_KEY_PATH=/run/secrets/rsa_private_key
RSA_PUBLIC_KEY_PATH=/run/secrets/rsa_public_key

# API配置
API_SECRET=${API_SECRET}
WEBHOOK_SECRET=${WEBHOOK_SECRET}

# 会话配置
SESSION_SECRET=${SESSION_SECRET}
CSRF_SECRET=${CSRF_SECRET}

# 安全配置
ENVIRONMENT=production
DEBUG_MODE=false
LOG_LEVEL=info

# 服务端口
API_GATEWAY_PORT=8080
USER_SERVICE_PORT=8081
DATA_SERVICE_PORT=8082
PAYMENT_SERVICE_PORT=8083
RECOMMENDATION_SERVICE_PORT=8084

# CORS配置
CORS_ALLOWED_ORIGINS=https://yourdomain.com
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS

# 速率限制
RATE_LIMIT_REQUESTS_PER_MINUTE=60
RATE_LIMIT_BURST=10
EOF
echo "✅ 生产环境配置已生成: ${CONFIG_DIR}/.env.production"

# ===========================================
# 生成开发环境配置
# ===========================================
echo "🛠️ 生成开发环境配置..."
cat > "${CONFIG_DIR}/.env.development" << EOF
# 🛠️ 开发环境配置 - 自动生成
# 生成时间: $(date)

# JWT配置
JWT_SECRET=${JWT_SECRET}
JWT_EXPIRE_HOURS=24
JWT_REFRESH_EXPIRE_HOURS=168

# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_NAME=gaokao_dev
DB_USER=gaokao_user
DB_PASSWORD=${DB_PASSWORD}
DB_SSL_MODE=disable

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=${REDIS_PASSWORD}
REDIS_DB=1

# 加密配置
AES_ENCRYPTION_KEY=${AES_KEY}
RSA_PRIVATE_KEY_PATH=./config/certs/gaokao-private.pem
RSA_PUBLIC_KEY_PATH=./config/certs/gaokao-public.pem

# API配置
API_SECRET=${API_SECRET}
WEBHOOK_SECRET=${WEBHOOK_SECRET}

# 会话配置
SESSION_SECRET=${SESSION_SECRET}
CSRF_SECRET=${CSRF_SECRET}

# 开发配置
ENVIRONMENT=development
DEBUG_MODE=true
LOG_LEVEL=debug

# 服务端口
API_GATEWAY_PORT=8080
USER_SERVICE_PORT=8081
DATA_SERVICE_PORT=8082
PAYMENT_SERVICE_PORT=8083
RECOMMENDATION_SERVICE_PORT=8084

# CORS配置
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS

# 速率限制
RATE_LIMIT_REQUESTS_PER_MINUTE=120
RATE_LIMIT_BURST=20

# 开发特性
DEV_MODE=true
DEV_MOCK_PAYMENT=true
DEV_ENABLE_SWAGGER=true
EOF
echo "✅ 开发环境配置已生成: ${CONFIG_DIR}/.env.development"

# ===========================================
# 设置文件权限
# ===========================================
echo "🔐 设置文件权限..."
chmod 600 "${SECRETS_DIR}"/*.env
chmod 600 "${CERTS_DIR}"/*.pem
chmod 600 "${CERTS_DIR}"/*.key
chmod 644 "${CERTS_DIR}"/*.crt
chmod 600 "${CONFIG_DIR}"/.env.*

echo "✅ 文件权限已设置"

# ===========================================
# 生成密钥摘要
# ===========================================
echo "📋 生成密钥摘要..."
cat > "${SECRETS_DIR}/keys-summary.txt" << EOF
🔐 高考志愿填报系统 - 密钥摘要
生成时间: $(date)

📁 文件位置:
- JWT密钥: ${SECRETS_DIR}/jwt.env
- 数据库密码: ${SECRETS_DIR}/database.env
- 加密密钥: ${SECRETS_DIR}/encryption.env
- API密钥: ${SECRETS_DIR}/api.env
- 会话密钥: ${SECRETS_DIR}/session.env
- RSA密钥对: ${CERTS_DIR}/gaokao-*.pem
- SSL证书: ${CERTS_DIR}/dev-*

🔑 密钥指纹:
- JWT密钥 SHA256: $(echo -n "$JWT_SECRET" | sha256sum | cut -d' ' -f1)
- 数据库密码 SHA256: $(echo -n "$DB_PASSWORD" | sha256sum | cut -d' ' -f1)
- AES密钥 SHA256: $(echo -n "$AES_KEY" | sha256sum | cut -d' ' -f1)

⚠️  安全提醒:
1. 请妥善保管所有密钥文件
2. 生产环境部署前请更换默认密钥
3. 定期轮换密钥以提高安全性
4. 不要将密钥文件提交到版本控制系统
EOF

echo ""
echo "🎉 密钥生成完成！"
echo ""
echo "📋 生成的文件:"
echo "   - 密钥文件: ${SECRETS_DIR}/"
echo "   - 证书文件: ${CERTS_DIR}/"
echo "   - 环境配置: ${CONFIG_DIR}/.env.*"
echo ""
echo "⚠️  重要提醒:"
echo "   1. 请将 config/secrets/ 目录添加到 .gitignore"
echo "   2. 生产环境部署前请验证所有密钥"
echo "   3. 定期备份密钥文件"
echo ""
echo "🚀 下一步: 运行 'docker-compose up -d' 启动服务"
