# 部署拓扑

## 概述

tishi 采用单机 Docker Compose 部署，包含 3 个容器：Go 应用、PostgreSQL、Nginx。

## 拓扑图

```
┌─────────────────────────────────────────────────────────────┐
│                      宿主机 (Linux)                          │
│                                                             │
│  ┌───────────────── Docker Compose ───────────────────────┐ │
│  │                                                         │ │
│  │  ┌─────────────┐  ┌──────────────┐  ┌──────────────┐  │ │
│  │  │   Nginx     │  │  tishi       │  │ PostgreSQL   │  │ │
│  │  │             │  │  (Go binary) │  │              │  │ │
│  │  │  :80/:443   │  │              │  │  :5432       │  │ │
│  │  │             │  │  API Server  │  │              │  │ │
│  │  │  静态文件    │──▶│  :8080      │──▶│  tishi_db   │  │ │
│  │  │  /dist/*   │  │              │  │              │  │ │
│  │  │             │  │  Scheduler   │  │  data volume │  │ │
│  │  │  反向代理    │  │  Collector   │  │              │  │ │
│  │  │  /api/* ───▶│  │  Analyzer    │  │              │  │ │
│  │  │             │  │  Generator   │  │              │  │ │
│  │  └─────────────┘  └──────────────┘  └──────────────┘  │ │
│  │        │                                                │ │
│  │        │ 端口映射                                        │ │
│  └────────┼────────────────────────────────────────────────┘ │
│           │                                                   │
│      :80/:443 ◀──── 用户浏览器                                 │
└───────────────────────────────────────────────────────────────┘
```

## 容器定义

### tishi（Go 应用）

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o /tishi ./cmd/tishi

FROM gcr.io/distroless/static-debian12
COPY --from=builder /tishi /tishi
ENTRYPOINT ["/tishi", "server"]
```

- **端口**：8080（内部，不对外暴露）
- **功能**：API Server + Scheduler（内置 Collector/Analyzer/Generator）
- **健康检查**：`GET /healthz`
- **环境变量**：数据库连接、GitHub Token 等

### PostgreSQL

- **镜像**：`postgres:16-alpine`
- **端口**：5432（内部网络，不对外暴露）
- **数据持久化**：Docker Volume `pg_data`
- **初始化**：通过 `tishi migrate` 命令执行 DDL

### Nginx

- **镜像**：`nginx:1.25-alpine`
- **端口**：80/443（对外暴露）
- **职责**：
  - 托管 Astro 构建产物（`/dist/*`）
  - 反向代理 API 请求（`/api/*` → `tishi:8080`）
  - HTTPS 终止（Let's Encrypt）
  - Gzip 压缩
  - 静态资源缓存头

## Docker Compose 配置

```yaml
# docker-compose.yml
version: "3.9"

services:
  tishi:
    build: .
    restart: unless-stopped
    depends_on:
      postgres:
        condition: service_healthy
    environment:
      - DATABASE_URL=postgres://tishi:${DB_PASSWORD}@postgres:5432/tishi_db?sslmode=disable
      - GITHUB_TOKENS=${GITHUB_TOKENS}
      - TZ=UTC
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:8080/healthz"]
      interval: 30s
      timeout: 5s
      retries: 3
    networks:
      - tishi-net

  postgres:
    image: postgres:16-alpine
    restart: unless-stopped
    environment:
      - POSTGRES_DB=tishi_db
      - POSTGRES_USER=tishi
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - pg_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U tishi -d tishi_db"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - tishi-net

  nginx:
    image: nginx:1.25-alpine
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./web/dist:/usr/share/nginx/html:ro
      - ./deploy/nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./deploy/nginx/certs:/etc/nginx/certs:ro
    depends_on:
      - tishi
    networks:
      - tishi-net

networks:
  tishi-net:
    driver: bridge

volumes:
  pg_data:
```

## 网络架构

```
互联网
  │
  ▼
Nginx (:80/:443)
  │
  ├── /api/*  ───▶  tishi:8080  ───▶  postgres:5432
  │
  └── /*      ───▶  /usr/share/nginx/html (静态文件)
```

- 所有容器在同一 Docker bridge 网络 `tishi-net` 内
- 仅 Nginx 对外暴露端口
- PostgreSQL 和 tishi 不直接对外

## 前端构建 & 部署流程

```bash
# 1. 构建 Astro 静态站点（在 CI 或本地执行）
cd web && npm run build

# 2. 构建产物在 web/dist/，Nginx 挂载此目录

# 3. 如需更新前端，重新构建后 Nginx 自动生效（volume 挂载）
```

## 备份策略

```bash
# PostgreSQL 数据备份（建议每日 cron）
docker-compose exec postgres pg_dump -U tishi tishi_db | gzip > backup_$(date +%Y%m%d).sql.gz
```

## 监控 & 日志

- **应用日志**：`docker-compose logs -f tishi`
- **健康检查**：Docker 内置 healthcheck，异常自动重启
- **未来扩展**：可接入 Prometheus（Go 应用暴露 `/metrics`）+ Grafana

## 相关文档

- [系统架构](system-architecture.md)
- [技术选型](tech-stack.md)
- [部署指南](../guides/deployment.md)
