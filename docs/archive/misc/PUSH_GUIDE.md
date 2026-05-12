# Git推送指南

## 当前状态

✅ 已完成2个本地提交：
1. `44ef09c` - fix: 修复API路径和Nginx配置，支持关键词模糊搜索
2. `b6dfd67` - docs: 添加API搜索功能修复报告

## 推送方式

### 方式1：配置SSH密钥（推荐）

```bash
# 1. 生成SSH密钥（如果没有）
ssh-keygen -t ed25519 -C "your_email@example.com"

# 2. 查看公钥
cat ~/.ssh/id_ed25519.pub

# 3. 将公钥添加到GitHub
# 访问：https://github.com/settings/keys
# 点击 "New SSH key"，粘贴公钥内容

# 4. 测试SSH连接
ssh -T git@github.com

# 5. 推送代码
git push origin master
```

### 方式2：使用GitHub Personal Access Token

```bash
# 1. 创建GitHub Token
# 访问：https://github.com/settings/tokens
# 点击 "Generate new token (classic)"
# 勾选 "repo" 权限
# 复制生成的token

# 2. 使用Token推送
git push https://<TOKEN>@github.com/oktetopython/gaokao.git master

# 示例（替换YOUR_TOKEN，不要把真实 token 写进命令历史或文档）：
# git push https://<YOUR_TOKEN>@github.com/oktetopython/gaokao.git master
```

### 方式3：使用Git凭据存储（一次性配置）

```bash
# 配置HTTPS凭据助手
git config --global credential.helper store

# 推送时会提示输入用户名和密码
# 用户名：GitHub用户名
# 密码：Personal Access Token（不是GitHub密码）
git push origin master
```

## 快速命令

### 查看当前分支状态
```bash
git status
git log --oneline -3
```

### 查看远程仓库配置
```bash
git remote -v
```

### 切换远程URL（如果需要）
```bash
# 使用SSH
git remote set-url origin git@github.com:oktetopython/gaokao.git

# 使用HTTPS
git remote set-url origin https://github.com/oktetopython/gaokao.git
```

## 推送成功后的验证

```bash
# 检查远程分支
git branch -r

# 验证提交历史
git log origin/master --oneline -3
```

## 本次提交内容摘要

**提交1：API路径和Nginx配置修复**
- 修复前端API路径：统一使用 `/api/{service}/api/v1/...` 格式
- 优化Nginx配置：正确的路径重写和Host header设置
- 添加部署脚本：自动化Nginx配置更新流程

**提交2：修复报告文档**
- 详细的修复说明
- 测试验证结果
- 部署指南
- 功能特性说明

## 文件变更统计

```
frontend/src/api/university.ts    | 10 +--
frontend/src/api/user.ts          |  8 +--
frontend/src/api/recommendation.ts |  6 +-
nginx-fixed.conf                  | 120 +++++++++++++++++++++
nginx.conf                        | 120 +++++++++++++++++++++
deploy-fixes.sh                   |  90 +++++++++++++++
API_SEARCH_FIX_REPORT.md          | 194 +++++++++++++++++++++++++++++
```

## 下一步操作

1. 选择上述推送方式之一
2. 执行推送命令
3. 验证推送成功
4. 在GitHub查看提交历史

---

**注意：** 请勿将包含敏感信息的Token提交到代码仓库！
