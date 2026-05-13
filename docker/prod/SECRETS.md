# Production Secrets 准备指引

> 此文档由运维侧执行。`secrets/` 目录被 `.gitignore` 忽略，所有真实密钥不入仓。
> compose 启动前必须在 `docker/prod/secrets/` 内准备好以下文件，否则 secrets 挂载失败。

## 1. 创建 secrets 目录

```bash
cd docker/prod
mkdir -p secrets
chmod 700 secrets
```

## 2. 必备 secret 文件

| 文件 | 用途 | 生成方式 |
|---|---|---|
| `postgres_password.txt` | Postgres 主密码 | `openssl rand -base64 32 > secrets/postgres_password.txt` |
| `redis_password.txt` | Redis AUTH 密码 | `openssl rand -base64 32 > secrets/redis_password.txt` |
| `jwt_secret.txt` | JWT 签名密钥（≥ 32 字节） | `openssl rand -base64 64 > secrets/jwt_secret.txt` |
| `aes_key.txt` | AES-256 加密密钥（32 字节十六进制或 base64） | `openssl rand -base64 32 > secrets/aes_key.txt` |
| `grafana_admin_password.txt` | Grafana 管理员密码 | `openssl rand -base64 24 > secrets/grafana_admin_password.txt` |
| `tls.crt` | 站点证书（PEM）— 必须签名为 `${DOMAIN_NAME}` | Let's Encrypt / 商业 CA |
| `tls.key` | 站点私钥（PEM） | 与 tls.crt 配套 |

注：以上 7 个为系统启动**强依赖**。secrets 加载失败 → 容器 CrashLoop。

## 3. 可选 secret 文件（按业务能力开启）

| 文件 | 用途 | 缺省影响 |
|---|---|---|
| `llm_api_key.txt` | OpenAI 兼容 LLM API key | recommendation-service 仍能跑，仅 LLM 增强禁用 |
| `alipay_private_key.pem` | 支付宝商户私钥 | alipay 适配器禁用 |
| `alipay_public_key.pem` | 支付宝平台公钥 | alipay 适配器禁用 |
| `wechat_api_key.txt` | 微信支付 API v3 密钥 | wechat 适配器禁用 |

> 注：payment-service 的 adapter factory 会按 secret 是否存在动态启用对应支付通道。
> 缺失时仅日志告警，不阻塞服务启动。

## 4. 占位测试 secrets（**禁止用于生产**）

如需先跑通编排再换正式 secret，可生成占位（自签证书 + 弱口令）：

```bash
# 一键生成全套占位（仅供本机演练）
cd docker/prod
mkdir -p secrets && cd secrets

openssl rand -base64 32 > postgres_password.txt
openssl rand -base64 32 > redis_password.txt
openssl rand -base64 64 > jwt_secret.txt
openssl rand -base64 32 > aes_key.txt
openssl rand -base64 24 > grafana_admin_password.txt

# 自签 TLS（127.0.0.1 + localhost）— 仅本机演练
openssl req -x509 -newkey rsa:4096 -nodes \
  -keyout tls.key -out tls.crt -days 365 \
  -subj "/CN=localhost" \
  -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"

# 可选项一律置空（adapter 自动跳过）
: > llm_api_key.txt
: > alipay_private_key.pem
: > alipay_public_key.pem
: > wechat_api_key.txt

chmod 600 *.txt *.pem *.key *.crt
```

## 5. 验证

```bash
# 从仓库根运行
docker compose -f docker/prod/docker-compose.prod.yml \
  --env-file docker/prod/.env config > /dev/null && echo "compose ok"

# 启动核心数据面
docker compose -f docker/prod/docker-compose.prod.yml \
  --env-file docker/prod/.env up -d postgres redis

# 等待 30s 后检查 healthcheck
docker compose -f docker/prod/docker-compose.prod.yml ps
```

## 6. 轮换

- TLS：在证书过期前 ≥ 7 天替换 `tls.crt` + `tls.key`，`docker compose restart nginx`
- 数据库密码：需协调下游服务（rolling restart 顺序：先重建 secret → 重启微服务 → 最后重启 postgres）
- JWT：轮换会使所有现有 token 失效，应在低峰期窗口操作

## 7. Let's Encrypt 自动签发（可选）

> **何时启用**：拿到真域名 + 域名 A 记录指到本机 + 80/443 端口对外可达后。
> 启用前 `tls.crt`/`tls.key` 仍然必备（占位自签即可，nginx 首启需要它才能起来）。

### 7.1 准备

```bash
# 项目根 .env 加两行
echo "DOMAIN_NAME=gaokaohub.example.com" >> docker/prod/.env
echo "LETSENCRYPT_EMAIL=ops@example.com" >> docker/prod/.env
```

### 7.2 首签（一次性）

`nginx` 必须先起来才能响应 `/.well-known/acme-challenge/`，所以顺序是：先用占位自签 `tls.crt` 启动 nginx，再签真证。

```bash
cd docker/prod

# (a) 用占位自签先启 nginx
docker compose -f docker-compose.prod.yml up -d nginx

# (b) 跑一次性 certbot 容器（HTTP-01 challenge 经 nginx /.well-known/acme-challenge 透传）
docker compose -f docker-compose.prod.yml --profile letsencrypt run --rm certbot \
  certonly --webroot -w /var/www/certbot \
  -d "${DOMAIN_NAME}" \
  -m "${LETSENCRYPT_EMAIL}" \
  --agree-tos --no-eff-email --non-interactive

# 成功后 /etc/letsencrypt/live/${DOMAIN_NAME}/{fullchain,privkey}.pem 已生成
docker compose -f docker-compose.prod.yml --profile letsencrypt run --rm \
  --entrypoint sh certbot -c "ls /etc/letsencrypt/live/${DOMAIN_NAME}/"
```

### 7.3 切换 nginx 到 LE 证书

编辑 `docker/prod/nginx/nginx.conf` 的 443 server block：

```diff
-    ssl_certificate     /run/secrets/tls_cert;
-    ssl_certificate_key /run/secrets/tls_key;
+    ssl_certificate     /etc/letsencrypt/live/${DOMAIN_NAME}/fullchain.pem;
+    ssl_certificate_key /etc/letsencrypt/live/${DOMAIN_NAME}/privkey.pem;
```

注：nginx.conf 不支持环境变量替换，`${DOMAIN_NAME}` 必须手动改成真实域名。改完：

```bash
docker compose -f docker-compose.prod.yml restart nginx
```

### 7.4 启常驻续期

```bash
docker compose -f docker-compose.prod.yml --profile letsencrypt up -d certbot-renewer
```

`certbot-renewer` 每 12h 跑一次 `certbot renew`：
- 证书距过期 ≥ 30 天 → no-op
- 证书距过期 < 30 天 → 续签 + `--deploy-hook` 给 `gaokao-nginx-prod` 发 SIGHUP（nginx hot reload，不中断连接）

`--deploy-hook` 依赖 docker socket（`renewer` 容器内 mount 了 `/var/run/docker.sock`）。这是 opt-in profile 的明确风险点，运维评估后启用。

### 7.5 验证

```bash
# 看 renewer 第一次循环日志
docker logs gaokao-certbot-renewer

# 手动跑一次 dry-run（不真签）
docker compose -f docker-compose.prod.yml --profile letsencrypt exec certbot-renewer \
  certbot renew --dry-run
```

### 7.6 撤回

若需回 docker secrets 模式：恢复 nginx.conf 的 `ssl_certificate /run/secrets/tls_cert` 两行，`docker compose stop certbot-renewer`，重启 nginx。`letsencrypt_certs` volume 数据保留，再次启用零成本。
