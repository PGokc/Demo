#!/bin/bash
# ============================================================
# restore.sh — 在【宿主机】执行完整恢复流程
# 用法: bash scripts/restore.sh
# ============================================================
set -euo pipefail

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; RED='\033[0;31m'; NC='\033[0m'
log()  { echo -e "${GREEN}[$(date '+%H:%M:%S')] $*${NC}"; }
warn() { echo -e "${YELLOW}[$(date '+%H:%M:%S')] $*${NC}"; }
info() { echo -e "${CYAN}[$(date '+%H:%M:%S')] $*${NC}"; }
err()  { echo -e "${RED}[$(date '+%H:%M:%S')] $*${NC}"; }

echo ""
echo "════════════════════════════════════════════"
echo "        🔄  XtraBackup 正式恢复流程"
echo "════════════════════════════════════════════"
echo ""

# ── Step 1: 停止 MySQL ───────────────────────────────────────
log "Step 1: 停止 MySQL 容器..."
docker stop xb_mysql
log "✅ xb_mysql 已停止"
echo ""

# ── Step 2: 清空数据目录 ─────────────────────────────────────
log "Step 2: 清空 MySQL 数据目录..."
docker exec xb_tool bash -c "find /var/lib/mysql -mindepth 1 -delete && echo '数据目录已清空'"
echo ""

# ── Step 3: copy-back ────────────────────────────────────────
log "Step 3: 执行 xtrabackup --copy-back..."
docker exec xb_tool xtrabackup \
  --copy-back \
  --target-dir=/backup/full \
  --datadir=/var/lib/mysql \
  --no-server-version-check
log "✅ copy-back 完成"
echo ""

# ── Step 4: 启动 MySQL ───────────────────────────────────────
log "Step 4: 启动 MySQL 容器..."
docker start xb_mysql

warn "⏳ 等待 MySQL 就绪..."
for i in $(seq 1 30); do
  if docker exec xb_mysql mysqladmin ping -uroot -pRoot1234! --silent 2>/dev/null; then
    log "✅ MySQL 已就绪"
    break
  fi
  echo -n "."
  sleep 2
done
echo ""

# ── Step 5: 验证数据 ─────────────────────────────────────────
log "Step 5: 查询恢复后的数据..."
docker exec xb_mysql mysql -uroot -pRoot1234! 2>/dev/null << 'SQL'
SELECT '─── 数据库列表 ───' AS '';
SHOW DATABASES;

SELECT '─── orders 表全部数据 ───' AS '';
SELECT id, customer, product, amount, created_at FROM shopdb.orders ORDER BY id;

SELECT '─── 汇总 ───' AS '';
SELECT COUNT(*) AS total_rows, SUM(amount) AS total_amount FROM shopdb.orders;
SQL

echo ""
log "🎉 恢复完成！数据已验证"
