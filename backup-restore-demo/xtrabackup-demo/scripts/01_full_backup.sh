#!/bin/bash
# ============================================================
# 01_full_backup.sh — 写入初始数据 + 执行全量备份
# 运行环境: xb_tool 容器
# ============================================================
set -euo pipefail

MYSQL_HOST="xb_mysql"
MYSQL_USER="root"
MYSQL_PASS="Root1234!"
BACKUP_DIR="/backup/full"

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'
log()  { echo -e "${GREEN}[$(date '+%H:%M:%S')] $*${NC}"; }
warn() { echo -e "${YELLOW}[$(date '+%H:%M:%S')] $*${NC}"; }

# ── Step 1: 写入初始数据 ─────────────────────────────────────
log "📝 Step 1: 向 MySQL 写入初始演示数据..."
mysql -h"$MYSQL_HOST" -u"$MYSQL_USER" -p"$MYSQL_PASS" < /scripts/seed_data.sql
log "初始数据写入完成"

# ── Step 2: 执行全量备份 ─────────────────────────────────────
log "💾 Step 2: 开始全量备份 → $BACKUP_DIR"
rm -rf "$BACKUP_DIR"
mkdir -p "$BACKUP_DIR"

xtrabackup \
  --backup \
  --host="$MYSQL_HOST" \
  --port=3306 \
  --user="$MYSQL_USER" \
  --password="$MYSQL_PASS" \
  --datadir=/var/lib/mysql \
  --target-dir="$BACKUP_DIR" \
  --parallel=2 \
  --no-server-version-check \
  2>&1 | tail -5

log "✅ 全量备份完成"
ls -lh "$BACKUP_DIR" | head -10

# ── Step 3: Prepare 全量备份（--apply-log-only 保留未提交事务供增量合并）
log "🔧 Step 3: Prepare 全量备份（apply-log-only）..."
xtrabackup \
  --prepare \
  --apply-log-only \
  --target-dir="$BACKUP_DIR" \
  2>&1 | tail -3

log "✅ 全量 Prepare 完成，可进行增量备份"
echo ""
warn "👉 下一步: 运行 02_incremental_backup.sh 写入更多数据并做增量备份"
