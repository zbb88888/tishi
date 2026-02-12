# 部署指南

## 概述

tishi 使用 Docker Compose 单机部署，包含 Go 应用、PostgreSQL、Nginx 三个容器。

## 前置条件

| 要求 | 说明 |
|------|------|
| 服务器 | Linux（Ubuntu 22.04+ / Debian 12+） |
| 配置 | 最低 1 CPU / 1GB RAM / 10GB 磁盘 |
| Docker | ≥ 24.x |
| Docker Compose | v2 |
| 域名（可选） | 用于 HTTPS |

## 部署步骤

### 1. 克隆项目

```bash
git clone https://github.com/zbb88888/tishi.git
cd tishi
```

### 2. 配置环境变量

```bash
cp .env.example .env
vim .env
```

必须配置的变量：

```bash
# 数据库密码（生产环境请使用强密码）
DB_PASSWORD=your_secure_password_here

# GitHub Token（至少一个，逗号分隔多个）
GITHUB_TOKENS=ghp_your_token_here

# 站点域名（可选，用于 Nginx 配置）
SITE_DOMAIN=tishi.example.com
```

### 3. 构建前端

```bash
cd web
pnpm install
pnpm build
cd ..
```

### 4. 启动服务

```bash
# 构建并启动所有容器
docker-compose up -d --build

# 查看状态
docker-compose ps

# 执行数据库迁移
docker-compose exec tishi /tishi migrate

# 查看日志
docker-compose logs -f
```

### 5. 验证

```bash
# 健康检查
curl http://localhost/healthz

# 测试 API
curl http://localhost/api/v1/categories

# 手动触发首次数据采集
docker-compose exec tishi /tishi collect
docker-compose exec tishi /tishi analyze
```

## HTTPS 配置（Let's Encrypt）

### 使用 Certbot

```bash
# 安装 certbot
apt install certbot

# 获取证书（先停止 Nginx 或使用 webroot 方式）
certbot certonly --standalone -d tishi.example.com

# 证书路径
# /etc/letsencrypt/live/tishi.example.com/fullchain.pem
# /etc/letsencrypt/live/tishi.example.com/privkey.pem

# 复制到项目目录
cp /etc/letsencrypt/live/tishi.example.com/fullchain.pem deploy/nginx/certs/
cp /etc/letsencrypt/live/tishi.example.com/privkey.pem deploy/nginx/certs/

# 重启 Nginx
docker-compose restart nginx
```

### 自动续期

```bash
# 添加 cron 任务
echo "0 0 1 * * certbot renew --quiet && docker-compose restart nginx" | crontab -
```

## 备份

### 自动备份脚本

```bash
#!/bin/bash
# scripts/backup.sh
set -euo pipefail

BACKUP_DIR="/backups/tishi"
DATE=$(date +%Y%m%d_%H%M%S)
KEEP_DAYS=30

mkdir -p "$BACKUP_DIR"

# 备份数据库
docker-compose exec -T postgres pg_dump -U tishi tishi_db | gzip > "$BACKUP_DIR/db_${DATE}.sql.gz"

# 清理旧备份
find "$BACKUP_DIR" -name "*.sql.gz" -mtime +$KEEP_DAYS -delete

echo "Backup completed: $BACKUP_DIR/db_${DATE}.sql.gz"
```

### 恢复

```bash
# 解压并恢复
gunzip < backup_20260212.sql.gz | docker-compose exec -T postgres psql -U tishi tishi_db
```

## 更新部署

```bash
# 拉取最新代码
git pull

# 重新构建前端
cd web && pnpm build && cd ..

# 重新构建并重启
docker-compose up -d --build

# 执行新的数据库迁移（如有）
docker-compose exec tishi /tishi migrate
```

## 监控

### 日志查看

```bash
# 所有服务日志
docker-compose logs -f

# 仅 tishi 应用日志
docker-compose logs -f tishi

# 仅最近 100 行
docker-compose logs --tail 100 tishi
```

### 健康检查

Docker 内置 healthcheck 会自动监控，容器异常时自动重启（`restart: unless-stopped`）。

```bash
# 查看容器健康状态
docker-compose ps
docker inspect tishi --format='{{.State.Health.Status}}'
```

### 磁盘监控

```bash
# 查看数据库数据量
docker-compose exec postgres psql -U tishi -d tishi_db -c "
SELECT schemaname, tablename,
       pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
"
```

## 故障排查

| 问题 | 排查 |
|------|------|
| 容器启动失败 | `docker-compose logs <service>` |
| 数据库连接失败 | 检查 `DB_PASSWORD` 环境变量，`pg_isready` 测试连接 |
| 采集无数据 | 检查 `GITHUB_TOKENS` 是否有效，查看采集日志 |
| 前端 404 | 确认 `web/dist/` 存在且已挂载到 Nginx |
| API 502 | tishi 容器可能未就绪，检查 healthcheck 状态 |

## 相关文档

- [部署拓扑](../architecture/deployment-topology.md) — 架构细节
- [配置说明](configuration.md) — 配置项参考
