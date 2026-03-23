#!/bin/bash
# ============================================================
# 03_simulate_crash.sh — 模拟故障：写入脏数据后删库
# 运行环境: xb_tool 容器 (通过 mysql client 操作)
# ============================================================
set -euo pipefail

MYSQL_HOST="xb_mysql"
MYSQL_USER="root"
MYSQL_PASS="Root1234!"

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'
log()  { echo -e "${GREEN}[$(date '+%H:%M:%S')] $*${NC}"; }
err()  { echo -e "${RED}[$(date '+%H:%M:%S')] $*${NC}"; }
warn() { echo -e "${YELLOW}[$(date '+%H:%M:%S')] $*${NC}"; }

mysql_exec() { mysql -h"$MYSQL_HOST" -u"$MYSQL_USER" -p"$MYSQL_PASS" -e "$1"; }

warn "⚠️  模拟灾难场景..."
echo ""

# 先看看现在有多少数据
log "📊 灾难前数据快照:"
mysql_exec "SELECT COUNT(*) AS total_orders FROM shopdb.orders;"
mysql_exec "SELECT * FROM shopdb.orders ORDER BY id;"

echo ""
warn "💥 Step 1: 误操作 — 写入一堆垃圾数据..."
mysql_exec "INSERT INTO shopdb.orders (customer, product, amount) VALUES
  ('HACKER', '@@@@DELETE ALL', -99999.99),
  ('ERROR',  'CORRUPTED',       0.00);"

warn "💥 Step 2: 删库跑路！DROP DATABASE shopdb..."
mysql_exec "DROP DATABASE shopdb;"

echo ""
err "🔥 数据库已删除！现在验证一下..."
mysql_exec "SHOW DATABASES;" 2>&1 || true

echo ""
warn "此时 shopdb 已不存在，需要从备份恢复"
warn "👉 下一步: 运行 04_restore.sh 进行恢复"
