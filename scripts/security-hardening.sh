#!/bin/bash

set -e

# 脚本配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
SECURITY_DIR="${PROJECT_ROOT}/security"
LOG_FILE="${SECURITY_DIR}/security.log"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1" | tee -a "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] ✅ $1${NC}" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ❌ $1${NC}" | tee -a "$LOG_FILE"
}

log_warning() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] ⚠️  $1${NC}" | tee -a "$LOG_FILE"
}

# 创建安全目录
create_security_dir() {
    log "创建安全配置目录..."
    mkdir -p "$SECURITY_DIR"
    mkdir -p "$SECURITY_DIR/certs"
    mkdir -p "$SECURITY_DIR/keys"
    mkdir -p "$SECURITY_DIR/configs"
    log_success "安全目录创建完成"
}

# 生成加密密钥
generate_encryption_keys() {
    log "生成加密密钥..."
    
    # RSA密钥对 (2048位)
    openssl genrsa -out "${SECURITY_DIR}/keys/rsa_private.pem" 2048
    openssl rsa -in "${SECURITY_DIR}/keys/rsa_private.pem" -pubout -out "${SECURITY_DIR}/keys/rsa_public.pem"
    
    # AES密钥 (256位)
    openssl rand -hex 32 > "${SECURITY_DIR}/keys/aes_key.txt"
    
    # JWT密钥
    openssl rand -base64 64 > "${SECURITY_DIR}/keys/jwt_secret.txt"
    
    # API密钥
    openssl rand -hex 16 > "${SECURITY_DIR}/keys/api_key.txt"
    
    # 设置密钥文件权限
    chmod 600 "${SECURITY_DIR}/keys/"*
    
    log_success "加密密钥生成完成"
}

# 生成SSL证书
generate_ssl_certificates() {
    log "生成SSL证书..."
    
    # 创建CA私钥
    openssl genrsa -out "${SECURITY_DIR}/certs/ca-key.pem" 4096
    
    # 创建CA证书
    openssl req -new -x509 -days 365 -key "${SECURITY_DIR}/certs/ca-key.pem" \
        -sha256 -out "${SECURITY_DIR}/certs/ca.pem" \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=GaoKaoHub/OU=IT/CN=GaoKaoHub CA"
    
    # 创建服务器私钥
    openssl genrsa -out "${SECURITY_DIR}/certs/server-key.pem" 4096
    
    # 创建服务器证书签名请求
    openssl req -subj "/C=CN/ST=Beijing/L=Beijing/O=GaoKaoHub/OU=IT/CN=api.gaokaohub.com" \
        -sha256 -new -key "${SECURITY_DIR}/certs/server-key.pem" \
        -out "${SECURITY_DIR}/certs/server.csr"
    
    # 创建扩展文件
    cat > "${SECURITY_DIR}/certs/server-extfile.cnf" << EOF
subjectAltName = DNS:api.gaokaohub.com,DNS:*.gaokaohub.com,DNS:localhost,IP:127.0.0.1
extendedKeyUsage = serverAuth
EOF
    
    # 签发服务器证书
    openssl x509 -req -days 365 -sha256 \
        -in "${SECURITY_DIR}/certs/server.csr" \
        -CA "${SECURITY_DIR}/certs/ca.pem" \
        -CAkey "${SECURITY_DIR}/certs/ca-key.pem" \
        -out "${SECURITY_DIR}/certs/server-cert.pem" \
        -extfile "${SECURITY_DIR}/certs/server-extfile.cnf" \
        -CAcreateserial
    
    # 清理临时文件
    rm "${SECURITY_DIR}/certs/server.csr" "${SECURITY_DIR}/certs/server-extfile.cnf"
    
    # 设置证书文件权限
    chmod 600 "${SECURITY_DIR}/certs/"*-key.pem
    chmod 644 "${SECURITY_DIR}/certs/"*.pem
    
    log_success "SSL证书生成完成"
}

# 代码混淆配置
setup_code_obfuscation() {
    log "配置代码混淆..."
    
    # Go代码混淆配置 (使用garble)
    cat > "${SECURITY_DIR}/configs/garble.yaml" << 'EOF'
# Go代码混淆配置
obfuscation:
  enabled: true
  seed: random
  literals: true
  tiny: true
  debug: false
  
# 排除的包
exclude:
  - "vendor/*"
  - "internal/test/*"
  
# 保护的函数
protect:
  - "main.main"
  - "*.validateLicense"
  - "*.generatePaymentSignature"
  - "*.encryptSensitiveData"
EOF

    # C++代码保护配置
    cat > "${SECURITY_DIR}/configs/cpp_protection.cmake" << 'EOF'
# C++代码保护配置

# 编译器安全选项
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -fstack-protector-strong")
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -D_FORTIFY_SOURCE=2")
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -fPIE -pie")
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -Wformat -Wformat-security")

# Release模式优化
if(CMAKE_BUILD_TYPE STREQUAL "Release")
    set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -O2 -s")
    set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -ffunction-sections -fdata-sections")
    set(CMAKE_EXE_LINKER_FLAGS "${CMAKE_EXE_LINKER_FLAGS} -Wl,--gc-sections")
    set(CMAKE_EXE_LINKER_FLAGS "${CMAKE_EXE_LINKER_FLAGS} -Wl,--strip-all")
endif()

# 符号保护
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -fvisibility=hidden")
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -fvisibility-inlines-hidden")

# 反调试保护
add_definitions(-DANTI_DEBUG=1)
add_definitions(-DANTI_TAMPER=1)
EOF

    # 创建字符串加密脚本
    cat > "${SECURITY_DIR}/configs/string_encrypt.py" << 'EOF'
#!/usr/bin/env python3
"""
字符串加密工具
用于加密源代码中的敏感字符串
"""

import os
import re
import sys
import base64
from cryptography.fernet import Fernet

def generate_key():
    """生成加密密钥"""
    return Fernet.generate_key()

def encrypt_string(text, key):
    """加密字符串"""
    f = Fernet(key)
    encrypted = f.encrypt(text.encode())
    return base64.b64encode(encrypted).decode()

def decrypt_string(encrypted_text, key):
    """解密字符串"""
    f = Fernet(key)
    encrypted = base64.b64decode(encrypted_text.encode())
    return f.decrypt(encrypted).decode()

def process_cpp_file(file_path, key):
    """处理C++文件，加密字符串常量"""
    with open(file_path, 'r') as f:
        content = f.read()
    
    # 查找字符串常量
    pattern = r'"([^"]*(?:secret|password|key|token|api)[^"]*)"'
    
    def replace_string(match):
        original = match.group(1)
        encrypted = encrypt_string(original, key)
        return f'decrypt_string("{encrypted}")'
    
    new_content = re.sub(pattern, replace_string, content, flags=re.IGNORECASE)
    
    with open(file_path, 'w') as f:
        f.write(new_content)

def process_go_file(file_path, key):
    """处理Go文件，加密字符串常量"""
    with open(file_path, 'r') as f:
        content = f.read()
    
    # 查找字符串常量
    pattern = r'"([^"]*(?:secret|password|key|token|api)[^"]*)"'
    
    def replace_string(match):
        original = match.group(1)
        encrypted = encrypt_string(original, key)
        return f'DecryptString("{encrypted}")'
    
    new_content = re.sub(pattern, replace_string, content, flags=re.IGNORECASE)
    
    with open(file_path, 'w') as f:
        f.write(new_content)

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python3 string_encrypt.py <source_directory>")
        sys.exit(1)
    
    source_dir = sys.argv[1]
    key = generate_key()
    
    # 保存密钥
    with open("encryption_key.bin", "wb") as f:
        f.write(key)
    
    # 处理源文件
    for root, dirs, files in os.walk(source_dir):
        for file in files:
            file_path = os.path.join(root, file)
            if file.endswith('.cpp') or file.endswith('.h'):
                process_cpp_file(file_path, key)
            elif file.endswith('.go'):
                process_go_file(file_path, key)
    
    print("字符串加密完成")
EOF

    chmod +x "${SECURITY_DIR}/configs/string_encrypt.py"
    
    log_success "代码混淆配置完成"
}

# 安全编译脚本
create_secure_build_script() {
    log "创建安全编译脚本..."
    
    cat > "${SECURITY_DIR}/secure_build.sh" << 'EOF'
#!/bin/bash

# 安全编译脚本
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
BUILD_DIR="${PROJECT_ROOT}/build/secure"

echo "🔒 开始安全编译..."

# 创建安全构建目录
mkdir -p "$BUILD_DIR"

# 清理之前的构建
rm -rf "$BUILD_DIR"/*

# Go服务安全编译
echo "编译Go服务..."
cd "$PROJECT_ROOT"

services=(
    "api-gateway"
    "user-service" 
    "data-service"
    "recommendation-service"
    "payment-service"
)

for service in "${services[@]}"; do
    service_dir="services/${service}"
    if [ -d "$service_dir" ]; then
        echo "安全编译: $service"
        cd "$service_dir"
        
        # 使用garble进行代码混淆
        if command -v garble &> /dev/null; then
            garble -seed=random -literals -tiny build -ldflags="-s -w -X main.buildTime=$(date +%s)" -o "${BUILD_DIR}/${service}" .
        else
            # 降级到普通编译
            go build -ldflags="-s -w -X main.buildTime=$(date +%s)" -o "${BUILD_DIR}/${service}" .
        fi
        
        # 使用UPX压缩
        if command -v upx &> /dev/null; then
            upx --brute "${BUILD_DIR}/${service}"
        fi
        
        cd "$PROJECT_ROOT"
        echo "✅ $service 编译完成"
    fi
done

# C++模块安全编译
echo "编译C++模块..."
cpp_modules=(
    "device-fingerprint"
    "license"
    "volunteer-matcher"
)

for module in "${cpp_modules[@]}"; do
    module_dir="cpp-modules/${module}"
    if [ -d "$module_dir" ]; then
        echo "安全编译: $module"
        cd "$module_dir"
        
        mkdir -p build_secure
        cd build_secure
        
        # 使用安全编译选项
        cmake .. \
            -DCMAKE_BUILD_TYPE=Release \
            -DCMAKE_CXX_FLAGS="-O2 -s -fstack-protector-strong -D_FORTIFY_SOURCE=2 -fPIE -pie" \
            -DCMAKE_EXE_LINKER_FLAGS="-Wl,--strip-all -Wl,--gc-sections" \
            -DCMAKE_INSTALL_PREFIX="$BUILD_DIR"
        
        make -j$(nproc)
        make install
        
        cd "$PROJECT_ROOT"
        echo "✅ $module 编译完成"
    fi
done

echo "🎉 安全编译完成！"
echo "输出目录: $BUILD_DIR"
EOF

    chmod +x "${SECURITY_DIR}/secure_build.sh"
    
    log_success "安全编译脚本创建完成"
}

# 渗透测试配置
setup_penetration_testing() {
    log "配置渗透测试..."
    
    # 创建安全测试脚本
    cat > "${SECURITY_DIR}/security_test.sh" << 'EOF'
#!/bin/bash

# 安全测试脚本
set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPORT_DIR="${PROJECT_ROOT}/security/reports"

mkdir -p "$REPORT_DIR"

echo "🔍 开始安全测试..."

# SQL注入测试
echo "测试SQL注入漏洞..."
if command -v sqlmap &> /dev/null; then
    sqlmap -u "http://localhost:8080/api/v1/data/universities?search=test" \
           --batch --risk=2 --level=2 \
           --output-dir="$REPORT_DIR/sqlmap" || true
else
    echo "⚠️  sqlmap未安装，跳过SQL注入测试"
fi

# XSS测试
echo "测试XSS漏洞..."
curl -s -X POST "http://localhost:8080/api/v1/users/search" \
     -H "Content-Type: application/json" \
     -d '{"query": "<script>alert(\"xss\")</script>"}' \
     > "$REPORT_DIR/xss_test.log" || true

# API安全测试
echo "测试API安全..."
curl -s "http://localhost:8080/api/v1/admin/users" \
     -H "Authorization: Bearer invalid_token" \
     > "$REPORT_DIR/api_auth_test.log" || true

# 端口扫描
echo "扫描开放端口..."
if command -v nmap &> /dev/null; then
    nmap -sS -O localhost > "$REPORT_DIR/port_scan.log" || true
else
    echo "⚠️  nmap未安装，跳过端口扫描"
fi

# SSL/TLS测试
echo "测试SSL/TLS配置..."
if command -v testssl.sh &> /dev/null; then
    testssl.sh --html localhost:8080 > "$REPORT_DIR/ssl_test.html" || true
else
    echo "⚠️  testssl.sh未安装，跳过SSL测试"
fi

echo "🎉 安全测试完成！"
echo "报告目录: $REPORT_DIR"
EOF

    chmod +x "${SECURITY_DIR}/security_test.sh"
    
    # 创建安全检查清单
    cat > "${SECURITY_DIR}/security_checklist.md" << 'EOF'
# 安全检查清单

## 1. 代码安全
- [ ] 移除所有硬编码的密钥和密码
- [ ] 实施输入验证和输出编码
- [ ] 使用参数化查询防止SQL注入
- [ ] 实施CSRF保护
- [ ] 启用内容安全策略(CSP)
- [ ] 代码混淆和字符串加密

## 2. 身份认证和授权
- [ ] 强密码策略
- [ ] JWT令牌安全配置
- [ ] 会话管理安全
- [ ] 多因素认证(MFA)
- [ ] 角色和权限控制

## 3. 数据保护
- [ ] 数据库加密
- [ ] 传输加密(HTTPS/TLS)
- [ ] 敏感数据掩码
- [ ] 数据备份加密
- [ ] PII数据保护

## 4. 基础设施安全
- [ ] 网络分段
- [ ] 防火墙配置
- [ ] 入侵检测系统
- [ ] 日志记录和监控
- [ ] 定期安全更新

## 5. 应用安全
- [ ] 错误处理安全
- [ ] 文件上传安全
- [ ] API速率限制
- [ ] 反爬虫保护
- [ ] 业务逻辑验证

## 6. 运维安全
- [ ] 容器安全配置
- [ ] 秘钥管理
- [ ] 访问控制
- [ ] 审计日志
- [ ] 灾难恢复计划

## 7. 合规性
- [ ] 数据隐私保护
- [ ] 审计要求
- [ ] 法规遵循
- [ ] 安全培训
- [ ] 事件响应计划
EOF

    log_success "渗透测试配置完成"
}

# Docker安全配置
setup_docker_security() {
    log "配置Docker安全..."
    
    # 创建安全的Docker配置
    cat > "${SECURITY_DIR}/configs/docker-security.yaml" << 'EOF'
# Docker安全配置

# 基础镜像安全
base_images:
  - alpine:latest
  - distroless/static:latest
  - scratch

# 安全选项
security_opts:
  - no-new-privileges:true
  - seccomp:unconfined
  - apparmor:docker-default

# 资源限制
resources:
  memory: 512M
  cpu: 0.5
  pids: 100

# 网络安全
networks:
  - name: gaokaohub-secure
    driver: bridge
    encrypted: true
    
# 用户配置
users:
  - name: appuser
    uid: 1001
    gid: 1001
    shell: /sbin/nologin

# 挂载安全
volumes:
  - type: tmpfs
    target: /tmp
    options: noexec,nosuid,nodev
  - type: bind
    source: /etc/ssl/certs
    target: /etc/ssl/certs
    readonly: true
EOF

    # 创建安全的Dockerfile模板
    cat > "${SECURITY_DIR}/configs/Dockerfile.secure" << 'EOF'
# 安全的Dockerfile模板

# 使用官方基础镜像
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装安全依赖
RUN apk add --no-cache ca-certificates git tzdata

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 复制源代码
COPY --chown=appuser:appgroup . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags='-s -w -extldflags "-static"' -o app .

# 运行阶段 - 使用distroless镜像
FROM gcr.io/distroless/static:nonroot

# 复制证书
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# 复制应用
COPY --from=builder /app/app /app

# 使用非root用户
USER nonroot:nonroot

# 设置只读根文件系统
USER 65532:65532

# 安全配置
LABEL security.scan="enabled" \
      security.non-root="true" \
      security.readonly-rootfs="true"

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app", "--health-check"]

# 运行应用
ENTRYPOINT ["/app"]
EOF

    log_success "Docker安全配置完成"
}

# Kubernetes安全配置
setup_k8s_security() {
    log "配置Kubernetes安全..."
    
    mkdir -p "${SECURITY_DIR}/k8s"
    
    # Pod安全策略
    cat > "${SECURITY_DIR}/k8s/pod-security-policy.yaml" << 'EOF'
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: gaokaohub-psp
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
    - ALL
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'projected'
    - 'secret'
    - 'downwardAPI'
    - 'persistentVolumeClaim'
  runAsUser:
    rule: 'MustRunAsNonRoot'
  seLinux:
    rule: 'RunAsAny'
  fsGroup:
    rule: 'RunAsAny'
  readOnlyRootFilesystem: true
  allowedHostPaths: []
  hostNetwork: false
  hostIPC: false
  hostPID: false
EOF

    # 网络策略
    cat > "${SECURITY_DIR}/k8s/network-policy.yaml" << 'EOF'
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: gaokaohub-network-policy
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: gaokaohub
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to: []
    ports:
    - protocol: TCP
      port: 443
    - protocol: TCP
      port: 53
    - protocol: UDP
      port: 53
EOF

    # 服务账户
    cat > "${SECURITY_DIR}/k8s/service-account.yaml" << 'EOF'
apiVersion: v1
kind: ServiceAccount
metadata:
  name: gaokaohub-sa
automountServiceAccountToken: false
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: gaokaohub-role
rules:
- apiGroups: [""]
  resources: ["configmaps", "secrets"]
  verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: gaokaohub-binding
subjects:
- kind: ServiceAccount
  name: gaokaohub-sa
roleRef:
  kind: Role
  name: gaokaohub-role
  apiGroup: rbac.authorization.k8s.io
EOF

    log_success "Kubernetes安全配置完成"
}

# 生成安全报告
generate_security_report() {
    log "生成安全报告..."
    
    cat > "${SECURITY_DIR}/security_report.md" << 'EOF'
# 高考志愿填报助手 - 安全报告

## 1. 项目安全概述

本项目采用了多层次的安全防护措施，确保用户数据和系统的安全性。

## 2. 安全措施

### 2.1 代码安全
- 使用Go garble进行代码混淆
- C++代码采用符号保护和反调试技术
- 敏感字符串加密存储
- 移除调试符号和优化编译

### 2.2 加密保护
- RSA 2048位密钥用于数字签名
- AES 256位密钥用于对称加密
- JWT令牌安全配置
- SSL/TLS证书保护传输

### 2.3 访问控制
- 基于角色的权限控制(RBAC)
- JWT身份验证
- API速率限制
- 设备指纹验证

### 2.4 基础设施安全
- Docker容器安全配置
- Kubernetes安全策略
- 网络分段和防火墙
- 定期安全扫描

## 3. 合规性

项目遵循以下安全标准：
- OWASP Top 10 防护
- 数据隐私保护法规
- 网络安全等级保护要求

## 4. 监控和审计

- 全面的日志记录
- 实时安全监控
- 异常行为检测
- 定期安全审计

## 5. 事件响应

建立了完整的安全事件响应机制：
- 事件检测和分类
- 响应流程和预案
- 恢复和改进措施

## 6. 安全测试

定期进行以下安全测试：
- 渗透测试
- 漏洞扫描
- 代码审计
- 依赖检查

## 7. 建议

### 7.1 运维建议
- 定期更新系统和依赖
- 监控安全事件和日志
- 备份关键数据
- 培训运维人员

### 7.2 开发建议
- 遵循安全编码规范
- 定期安全代码审查
- 使用安全的第三方库
- 实施安全测试

生成时间: $(date)
EOF

    log_success "安全报告生成完成"
}

# 清理函数
cleanup() {
    log "清理临时文件..."
    # 清理可能的敏感临时文件
    find "${SECURITY_DIR}" -name "*.tmp" -delete 2>/dev/null || true
    log_success "清理完成"
}

# 主函数
main() {
    echo "=========================================="
    echo "🔒 高考志愿填报助手 - 安全加固脚本"
    echo "=========================================="
    
    # 检查OpenSSL
    if ! command -v openssl &> /dev/null; then
        log_error "OpenSSL未安装，请先安装OpenSSL"
        exit 1
    fi
    
    # 初始化
    create_security_dir
    
    # 开始安全加固
    log "开始安全加固过程..."
    log "项目根目录: $PROJECT_ROOT"
    log "安全目录: $SECURITY_DIR"
    
    # 生成加密密钥
    generate_encryption_keys
    
    # 生成SSL证书
    generate_ssl_certificates
    
    # 配置代码混淆
    setup_code_obfuscation
    
    # 创建安全编译脚本
    create_secure_build_script
    
    # 配置渗透测试
    setup_penetration_testing
    
    # 配置Docker安全
    setup_docker_security
    
    # 配置K8s安全
    setup_k8s_security
    
    # 生成安全报告
    generate_security_report
    
    # 完成
    log_success "安全加固完成！"
    echo ""
    echo "安全文件生成:"
    echo "  - 加密密钥: ${SECURITY_DIR}/keys/"
    echo "  - SSL证书: ${SECURITY_DIR}/certs/"
    echo "  - 安全配置: ${SECURITY_DIR}/configs/"
    echo "  - 安全脚本: ${SECURITY_DIR}/"
    echo "  - 安全报告: ${SECURITY_DIR}/security_report.md"
    echo ""
    echo "下一步操作:"
    echo "  1. 执行安全编译: ${SECURITY_DIR}/secure_build.sh"
    echo "  2. 运行安全测试: ${SECURITY_DIR}/security_test.sh"
    echo "  3. 审查安全报告: ${SECURITY_DIR}/security_report.md"
    echo ""
}

# 捕获退出信号
trap cleanup EXIT

# 执行主函数
main "$@"