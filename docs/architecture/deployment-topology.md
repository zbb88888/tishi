# 部署拓扑

## 概述

tishi v1.0 采用三机解耦部署，每个 Stage 独立运行在不同机器上，通过 Git 仓库交换数据。

## 拓扑图

```
┌───────────────────────────────────────────────────────────────────────────────────────┐
│                              tishi v1.0 部署拓扑                                       │
│                                                                                       │
│  Machine A (采集+分析)          Machine B (SSG 构建)          Machine C (发布)           │
│  ┌───────────────────┐        ┌───────────────────┐        ┌───────────────────┐      │
│  │ cron (每日触发)     │        │ cron (每日触发)     │        │                   │      │
│  │       ↓            │        │       ↓            │        │    Nginx          │      │
│  │ tishi scrape       │        │  git pull data/    │        │    :80/:443       │      │
│  │       ↓            │        │       ↓            │        │       ↓           │      │
│  │ tishi analyze      │        │  npm run build     │        │   serve dist/     │      │
│  │       ↓            │  Git   │       ↓            │ rsync  │                   │      │
│  │ tishi score        │ ─────→ │  dist/ 输出        │ ─────→ │   用户浏览器访问    │      │
│  │       ↓            │ push   │                   │        │                   │      │
│  │ tishi generate     │        └───────────────────┘        └───────────────────┘      │
│  │       ↓            │                                                                │
│  │ tishi push         │                                                                │
│  │  (git push data/)  │                                                                │
│  └───────────────────┘                                                                │
│                                                                                       │
│  ❶ 仅需 Go runtime     ❷ 仅需 Node.js + Git         ❸ 仅需 Nginx                      │
│     + GitHub Token        + npm                         + 静态文件                      │
│     + LLM API Key                                                                     │
└───────────────────────────────────────────────────────────────────────────────────────┘
```

## Machine A: 采集+分析

**需求**：Go 1.23+, Git, 网络（访问 GitHub + LLM API）

**运行内容**：

- `tishi` Go 二进制 (单文件，~15MB)
- cron job 每日触发完整 pipeline

**环境变量**：

```bash
TISHI_GITHUB_TOKENS=ghp_xxx,ghp_yyy    # GitHub Token (支持多个轮换)
TISHI_LLM_PROVIDER=deepseek             # deepseek 或 qwen
TISHI_LLM_API_KEY=sk-xxx               # LLM API Key
TISHI_DATA_DIR=./data                   # 数据目录 (Git 仓库)
```

**Cron 配置**：

```cron
# 每日 UTC 00:00 运行完整 pipeline
0 0 * * * cd /opt/tishi && ./tishi scrape && ./tishi analyze && ./tishi score && ./tishi generate && ./tishi push
```

## Machine B: SSG 构建

**需求**：Node.js 20+, npm, Git

**运行内容**：

- Astro SSG 项目 (web/ 目录)
- cron job 每日 git pull + build

**Cron 配置**：

```cron
# 每日 UTC 01:00 (给 Stage 1 一小时完成)
0 1 * * * cd /opt/tishi && git pull && cd web && npm run build && rsync -avz dist/ machineC:/var/www/tishi/
```

## Machine C: 发布

**需求**：Nginx

**Nginx 配置**：

```nginx
server {
    listen 80;
    server_name tishi.example.com;
    root /var/www/tishi;
    index index.html;

    location / {
        try_files $uri $uri/ /404.html;
    }

    # 静态资源长缓存
    location ~* \.(js|css|png|jpg|svg|woff2)$ {
        expires 30d;
        add_header Cache-Control "public, immutable";
    }

    gzip on;
    gzip_types text/html text/css application/javascript application/json;
}
```

## 单机部署简化方案

如果三台机器的资源不可用，可以在单机上运行全部 Stage：

```bash
# 单机一键 pipeline
cd /opt/tishi
./tishi scrape && ./tishi analyze && ./tishi score && ./tishi generate
cd web && npm run build
# Nginx 直接 serve web/dist/
```

Docker Compose 仅包含 Nginx：

```yaml
version: "3.9"
services:
  nginx:
    image: nginx:1.25-alpine
    ports:
      - "80:80"
    volumes:
      - ./web/dist:/usr/share/nginx/html:ro
      - ./deploy/nginx/nginx.conf:/etc/nginx/nginx.conf:ro
```

不再需要 PostgreSQL 和 Go 常驻服务容器。

## 与 v0.x 架构对比

| 维度 | v0.x | v1.0 |
|------|------|------|
| 容器数 | 3 (Go + PG + Nginx) | 1 (仅 Nginx) |
| 数据库 | PostgreSQL 常驻 | 无 (JSON 文件) |
| Go 进程 | 常驻 (API Server + Scheduler) | 按需运行 (CLI) |
| 数据同步 | PG 共享访问 | Git push/pull |
| 耦合度 | 所有模块依赖同一 PG | 三阶段完全独立 |

## 相关文档

- [系统架构](system-architecture.md)
- [技术选型](tech-stack.md)
- [部署指南](../guides/deployment.md)
