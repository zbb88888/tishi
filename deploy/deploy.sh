#!/usr/bin/env bash
# deploy.sh — tishi Stage 2+3 deployment script
# Pulls latest data/ from Git, builds Astro SSG, deploys to Nginx.
#
# Usage:
#   ./deploy/deploy.sh                 # default: deploy to /var/www/tishi
#   DEPLOY_DIR=/opt/www ./deploy/deploy.sh   # custom deploy directory
#
# Flow:
#   git pull (data/) → npm install → npm run build → rsync dist/ → reload nginx

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

DEPLOY_DIR="${DEPLOY_DIR:-/var/www/tishi}"
WEB_DIR="${PROJECT_ROOT}/web"
DIST_DIR="${WEB_DIR}/dist"

log() { echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"; }

# ── Step 1: Pull latest data ──
log "Pulling latest data/ from Git..."
cd "${PROJECT_ROOT}"
git pull --ff-only origin main

# ── Step 2: Install frontend deps (if needed) ──
if [ ! -d "${WEB_DIR}/node_modules" ]; then
    log "Installing frontend dependencies..."
    cd "${WEB_DIR}" && npm install
fi

# ── Step 3: Build Astro SSG ──
log "Building Astro static site..."
cd "${WEB_DIR}"
npm run build

if [ ! -d "${DIST_DIR}" ]; then
    log "ERROR: Build failed — dist/ not found"
    exit 1
fi

# ── Step 4: Deploy to Nginx serve directory ──
log "Deploying to ${DEPLOY_DIR}..."
sudo mkdir -p "${DEPLOY_DIR}"
sudo rsync -a --delete "${DIST_DIR}/" "${DEPLOY_DIR}/"

# ── Step 5: Reload Nginx ──
if command -v nginx >/dev/null 2>&1; then
    log "Reloading Nginx..."
    sudo nginx -t && sudo nginx -s reload
    log "Nginx reloaded successfully."
else
    log "WARN: nginx not found — skip reload. Copy dist/ manually if needed."
fi

log "Deploy complete. Site: ${DEPLOY_DIR}"
