# 部署指南

## 概述

v1.0 采用三阶段独立部署，每个阶段可部署在不同机器上，通过 Git 仓库交换数据。也支持单机简化部署。

## 前置条件

### Machine A（数据采集 + 分析）

| 要求 | 说明 |
|------|------|
| OS | Linux / macOS |
| 配置 | 最低 1 CPU / 512MB RAM |
| Go | ≥ 1.23（构建 CLI 用） |
| Git | ≥ 2.x |

### Machine B（SSG 构建）

| 要求 | 说明 |
|------|------|
| OS | Linux / macOS |
| 配置 | 最低 1 CPU / 1GB RAM |
| Node.js | ≥ 20 LTS |
| pnpm | ≥ 9.x |
| Git | ≥ 2.x |

### Machine C（Web 服务）

| 要求 | 说明 |
|------|------|
| OS | Linux |
| 配置 | 最低 1 CPU / 256MB RAM |
| Nginx | ≥ 1.24 或 Docker |

## 三机部署

### Machine A：数据采集 + LLM 分析

```bash
# 1. 克隆项目
git clone https://github.com/zbb88888/tishi.git
cd tishi

# 2. 配置
cp config.yaml.example config.yaml
vim config.yaml
# 配置 llm.api_key 和 github.tokens

# 3. 构建
make build

# 4. 手动运行 pipeline 一次，检查正常
./bin/tishi scrape
./bin/tishi analyze
./bin/tishi score
./bin/tishi generate --type weekly
./bin/tishi push

# 5. 配置 crontab
crontab -e
```

Crontab 配置：

```cron
# 每日 00:00 UTC 运行完整 pipeline
0 0 * * *  cd /opt/tishi && flock -n /tmp/tishi-pipeline.lock ./scripts/run-pipeline.sh >> /var/log/tishi/pipeline.log 2>&1

# 每周日 06:00 UTC 生成周报
0 6 * * 0  cd /opt/tishi && ./bin/tishi generate --type weekly && ./bin/tishi push >> /var/log/tishi/weekly.log 2>&1
```

Pipeline 脚本 `scripts/run-pipeline.sh`：

```bash
#!/bin/bash
set -euo pipefail

cd "$(dirname "$0")/.."
LOG_PREFIX="[$(date -u +%Y-%m-%dT%H:%M:%SZ)]"

echo "$LOG_PREFIX Starting pipeline..."

# 拉取最新数据（其他机器可能推送了审核结果）
git pull --rebase origin main

# 抓取 Trending + 过滤 AI 项目 + API enrichment
./bin/tishi scrape
echo "$LOG_PREFIX scrape done"

# LLM 分析（仅对新项目或过期项目）
./bin/tishi analyze
echo "$LOG_PREFIX analyze done"

# 评分排名
./bin/tishi score
echo "$LOG_PREFIX score done"

# 推送数据到 Git
./bin/tishi push
echo "$LOG_PREFIX push done"
```

### Machine B：Astro SSG 构建

```bash
# 1. 克隆项目
git clone https://github.com/zbb88888/tishi.git
cd tishi

# 2. 安装前端依赖
cd web && pnpm install && cd ..

# 3. 配置 crontab — 在 Machine A pipeline 完成后运行
crontab -e
```

```cron
# 每日 02:00 UTC（Machine A 00:00 跑完后）
0 2 * * *  cd /opt/tishi && git pull --rebase origin main && cd web && pnpm build >> /var/log/tishi/build.log 2>&1
```

构建产物在 `web/dist/`，可通过以下方式发布到 Machine C：

- `rsync -avz web/dist/ machineC:/var/www/tishi/`
- 或将 `dist/` 推送到另一个 Git 仓库

### Machine C：Nginx 静态托管

```bash
# 1. 安装 Nginx
apt install nginx

# 2. 配置站点
cat > /etc/nginx/sites-available/tishi << 'EOF'
server {
    listen 80;
    server_name tishi.example.com;
    root /var/www/tishi;
    index index.html;

    gzip on;
    gzip_types text/html text/css application/javascript application/json image/svg+xml;
    gzip_min_length 1024;

    location / {
        try_files $uri $uri/ /404.html;
    }

    # 缓存静态资源
    location ~* \.(css|js|png|jpg|svg|woff2)$ {
        expires 30d;
        add_header Cache-Control "public, immutable";
    }

    # JSON 数据不缓存
    location ~* \.json$ {
        expires -1;
        add_header Cache-Control "no-cache";
    }
}
EOF

ln -s /etc/nginx/sites-available/tishi /etc/nginx/sites-enabled/
nginx -t && systemctl reload nginx
```

## 单机简化部署

三个阶段在同一台机器上运行：

```bash
# 1. 环境准备
# 需要 Go 1.23+, Node.js 20+, pnpm 9+, Nginx

# 2. 克隆 + 配置
git clone https://github.com/zbb88888/tishi.git
cd tishi
cp config.yaml.example config.yaml
vim config.yaml

# 3. 构建 CLI + 前端
make build
cd web && pnpm install && pnpm build && cd ..

# 4. 配置 Nginx → web/dist/
# 5. 配置 crontab
```

单机 Crontab：

```cron
# 每日 00:00 完整 pipeline
0 0 * * *  cd /opt/tishi && flock -n /tmp/tishi.lock ./scripts/run-pipeline.sh >> /var/log/tishi/pipeline.log 2>&1

# 每日 02:00 重新构建前端
0 2 * * *  cd /opt/tishi/web && pnpm build >> /var/log/tishi/build.log 2>&1

# 每周日 06:00 生成周报 + 重建前端
0 6 * * 0  cd /opt/tishi && ./bin/tishi generate --type weekly && cd web && pnpm build >> /var/log/tishi/weekly.log 2>&1
```

## HTTPS（Let's Encrypt）

```bash
apt install certbot python3-certbot-nginx
certbot --nginx -d tishi.example.com

# 自动续期（certbot 自带 timer）
systemctl enable certbot.timer
```

## 备份

v1.0 数据全部在 Git 仓库中，备份 = Git push 到远程仓库。

```bash
# 数据已通过 tishi push 推送到 Git 远程
# 额外备份：推送到第二个远程
git remote add backup git@backup-server:tishi-data.git
git push backup main
```

## 更新部署

```bash
# Machine A
cd /opt/tishi
git pull
make build
# crontab 自动生效

# Machine B
cd /opt/tishi
git pull
cd web && pnpm install && pnpm build

# Machine C（如果使用 rsync）
# Machine B 构建后自动 rsync
```

## 监控

### 日志

```bash
# Pipeline 日志
tail -f /var/log/tishi/pipeline.log

# 构建日志
tail -f /var/log/tishi/build.log

# Nginx 访问日志
tail -f /var/log/nginx/access.log
```

### 健康检查

```bash
# 检查数据是否每日更新
ls -lt data/rankings/ | head -5

# 检查前端构建时间
stat web/dist/index.html

# 检查 Nginx
curl -I http://tishi.example.com
```

## 故障排查

| 问题 | 排查 |
|------|------|
| Pipeline 无数据 | 检查 `config.yaml` 中的 Token/API Key |
| LLM 分析失败 | 检查 `llm.api_key` 和 provider 配置 |
| 前端构建失败 | `cd web && pnpm install` 重装依赖 |
| Nginx 404 | 确认 `root` 指向 `web/dist/` |
| Git push 冲突 | `git pull --rebase` 后重试 |
| 数据不更新 | 检查 crontab (`crontab -l`) 和 flock 锁 |

## 相关文档

- [部署拓扑](../architecture/deployment-topology.md) — 架构细节
- [配置说明](configuration.md) — 配置项参考
