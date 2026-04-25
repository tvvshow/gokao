# SSL证书配置

## 证书文件说明

此目录用于存放SSL证书文件，用于HTTPS加密通信：

```
ssl/
├── gaokao.example.com.crt     # 证书文件
├── gaokao.example.com.key     # 私钥文件
└── ca-bundle.crt             # CA证书包
```

## 证书获取方式

### 方式1: Let's Encrypt免费证书
```bash
# 安装certbot
sudo apt-get install certbot

# 获取证书
sudo certbot certonly --standalone -d gaokao.example.com -d www.gaokao.example.com

# 复制证书到项目目录
sudo cp /etc/letsencrypt/live/gaokao.example.com/fullchain.pem ./ssl/gaokao.example.com.crt
sudo cp /etc/letsencrypt/live/gaokao.example.com/privkey.pem ./ssl/gaokao.example.com.key
sudo cp /etc/letsencrypt/live/gaokao.example.com/chain.pem ./ssl/ca-bundle.crt
```

### 方式2: 自签名证书（测试环境）
```bash
# 生成私钥
openssl genrsa -out gaokao.example.com.key 2048

# 生成证书签名请求
openssl req -new -key gaokao.example.com.key -out gaokao.example.com.csr

# 生成自签名证书
openssl x509 -req -days 365 -in gaokao.example.com.csr -signkey gaokao.example.com.key -out gaokao.example.com.crt

# 生成CA证书包
cat gaokao.example.com.crt > ca-bundle.crt
```

### 方式3: 购买商业证书
从您的证书颁发机构(CA)购买证书，然后将证书文件复制到此目录。

## 权限设置
```bash
# 设置正确的文件权限
chmod 600 gaokao.example.com.key
chmod 644 gaokao.example.com.crt ca-bundle.crt
```

## 自动续期（Let's Encrypt）
添加cron任务：
```bash
# 编辑crontab
crontab -e

# 添加以下行（每月自动续期）
0 0 1 * * certbot renew --quiet --post-hook "docker-compose restart frontend"
```

## 注意事项
1. 确保证书文件路径正确
2. 设置适当的文件权限（私钥必须是600）
3. 定期检查证书有效期
4. 生产环境建议使用权威CA签发的证书