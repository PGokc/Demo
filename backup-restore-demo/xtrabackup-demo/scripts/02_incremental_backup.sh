#!/bin/bash
# ============================================================
# 02_incremental_backup.sh — 写入新数据 + 执行增量备份（两轮）
# 运行环境: xb_tool 容器
# ============================================================
set -euo pipefail

MYSQL_HOST="xb_mysql"
MYSQL_USER="root"
MYSQL_PASS="Root1234!"
FULL_DIR="/backup/full"
INC1_DIR="/backup/inc1"
INC2_DIR="/backup/inc2"

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; NC='\033[0m'
log()   { echo -e "${GREEN}[$(date '+%H:%M:%S')] $*${NC}"; }
warn()  { echo -e "${YELLOW}[$(date '+%H:%M:%S')] $*${NC}"; }
info()  { echo -e "${CYAN}[$(date '+%H:%M:%S')] $*${NC}"; }

mysql_exec() { mysql -h"$MYSQL_HOST" -u"$MYSQL_USER" -p"$MYSQL_PASS" shopdb -e "$1"; }

# ── Step 1: 写入第二批数据 ───────────────────────────────────
log "📝 Step 1: 写入第二批数据（增量1 之前）..."
mysql_exec "INSERT INTO orders (customer, product, amount) VALUES
  ('Frank', 'Mac Mini',    4599.00),
  ('Grace', 'HomePod',     2299.00),
  ('Hank',  'Magic Mouse',  529.00);"
mysql_exec "SELECT COUNT(*) AS total_rows FROM orders;"

# ── Step 2: 第一次增量备份（基于 full）────────────────────────
log "💾 Step 2: 第一次增量备份（inc1，基于 full）→ $INC1_DIR"
rm -rf "$INC1_DIR" && mkdir -p "$INC1_DIR"

xtrabackup \
  --backup \
  --host="$MYSQL_HOST" \
  --port=3306 \
  --user="$MYSQL_USER" \
  --password="$MYSQL_PASS" \
  --datadir=/var/lib/mysql \
  --target-dir="$INC1_DIR" \
  --incremental-basedir="$FULL_DIR" \
  --parallel=2 \
  --no-server-version-check \
  2>&1 | tail -5

log "✅ inc1 备份完成"
info "inc1 大小: $(du -sh $INC1_DIR | cut -f1)  （对比 full: $(du -sh $FULL_DIR | cut -f1)）"

# ── Step 3: 写入第三批数据 ───────────────────────────────────
log "📝 Step 3: 写入第三批数据（增量2 之前）..."
mysql_exec "INSERT INTO orders (customer, product, amount) VALUES
  ('Ivy',  'iPhone 15 Pro', 9999.00),
  ('Jack', 'MacBook Air',   8999.00);"
mysql_exec "SELECT COUNT(*) AS total_rows FROM orders;"

# ── Step 4: 第二次增量备份（基于 inc1）───────────────────────
log "💾 Step 4: 第二次增量备份（inc2，基于 inc1）→ $INC2_DIR"
rm -rf "$INC2_DIR" && mkdir -p "$INC2_DIR"

xtrabackup \
  --backup \
  --host="$MYSQL_HOST" \
  --port=3306 \
  --user="$MYSQL_USER" \
  --password="$MYSQL_PASS" \
  --datadir=/var/lib/mysql \
  --target-dir="$INC2_DIR" \
  --incremental-basedir="$INC1_DIR" \
  --parallel=2 \
  --no-server-version-check \
  2>&1 | tail -5

log "✅ inc2 备份完成"

# ── Step 5: Prepare 合并（关键步骤）──────────────────────────
log "🔧 Step 5: Prepare — 合并 inc1 → full（apply-log-only）..."
xtrabackup --prepare --apply-log-only \
  --target-dir="$FULL_DIR" \
  --incremental-dir="$INC1_DIR" \
  2>&1 | tail -3

log "🔧 Step 6: Prepare — 合并 inc2 → full（最终，不加 apply-log-only）..."
xtrabackup --prepare \
  --target-dir="$FULL_DIR" \
  --incremental-dir="$INC2_DIR" \
  2>&1 | tail -3

log "✅ 所有增量已合并到 /backup/full，可进行恢复"
echo ""
warn "👉 下一步: 运行 03_simulate_crash.sh 模拟数据库故障"
