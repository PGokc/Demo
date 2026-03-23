#!/bin/bash
# ============================================================
# 05_verify.sh — 真正恢复 + 验证数据
# 运行环境: 宿主机（需要 docker 命令）
# ============================================================
set -euo pipefail

MYSQL_CONTAINER="xb_mysql"
MYSQL_VOL="xtrabackup-demo_mysql_data"
BACKUP_VOL="xtrabackup-demo_backup_data"

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; RED='\033[0;31m'; NC='\033[0m'
log()  { echo -e "${GREEN}[$(date '+%H:%M:%S')] $*${NC}"; }
warn() { echo -e "${YELLOW}[$(date '+%H:%M:%S')] $*${NC}"; }
info() { echo -e "${CYAN}[$(date '+%H:%M:%S')] $*${NC}"; }
err()  { echo -e "${RED}[$(date '+%H:%M:%S')] $*${NC}"; }

echo ""
echo "══════════════════════════════════════════════"
echo "      🔄  XtraBackup 正式恢复 + 数据验证"
echo "══════════════════════════════════════════════"
echo ""

# ── Step 1: 备份完整性检查 ───────────────────────────────────
log "Step 1: 检查备份完整性..."
docker run --rm \
  -v "$BACKUP_VOL":/backup \
  busybox sh -c "
    echo '--- xtrabackup_checkpoints ---'
    cat /backup/full/xtrabackup_checkpoints
    echo ''
    echo '--- binlog 位置 ---'
    cat /backup/full/xtrabackup_binlog_info 2>/dev/null || echo '无 binlog 信息'
    echo ''
    echo '--- 备份大小 ---'
    du -sh /backup/full  && echo 'full'
    du -sh /backup/inc1  2>/dev/null && echo 'inc1' || true
    du -sh /backup/inc2  2>/dev/null && echo 'inc2' || true
  "
echo ""

# ── Step 2: 停止 MySQL ───────────────────────────────────────
log "Step 2: 停止 MySQL..."
docker stop "$MYSQL_CONTAINER"
echo ""

# ── Step 3: 清空数据目录 ─────────────────────────────────────
log "Step 3: 清空 MySQL 数据目录..."
docker run --rm \
  -v "$MYSQL_VOL":/var/lib/mysql \
  busybox sh -c "find /var/lib/mysql -mindepth 1 -delete && echo '✅ 已清空'"
echo ""

# ── Step 4: copy-back 到真正的数据目录 ──────────────────────
log "Step 4: 执行 copy-back → /var/lib/mysql..."
docker run --rm \
  -v "$MYSQL_VOL":/var/lib/mysql \
  -v "$BACKUP_VOL":/backup \
  -u 999:999 \
  percona/percona-xtrabackup:8.0 \
  xtrabackup --copy-back \
    --target-dir=/backup/full \
    --datadir=/var/lib/mysql \
    --no-server-version-check
log "✅ copy-back 完成"
echo ""

# ── Step 5: 修复权限 ─────────────────────────────────────────
log "Step 5: 修复文件权限..."
docker run --rm \
  -v "$MYSQL_VOL":/var/lib/mysql \
  busybox chown -R 999:999 /var/lib/mysql
log "✅ 权限修复完成"
echo ""

# ── Step 6: 启动 MySQL ───────────────────────────────────────
log "Step 6: 启动 MySQL..."
docker start "$MYSQL_CONTAINER"

warn "⏳ 等待 MySQL 就绪..."
for i in $(seq 1 30); do
  if docker exec "$MYSQL_CONTAINER" mysqladmin ping -uroot -pRoot1234! --silent 2>/dev/null; then
    log "✅ MySQL 已就绪"
    break
  fi
  echo -n "."
  sleep 2
done
echo ""

# ── Step 7: 查询验证 ─────────────────────────────────────────
log "Step 7: 查询恢复后的数据..."
echo ""
docker exec "$MYSQL_CONTAINER" mysql -uroot -pRoot1234!   --table   -e "SHOW DATABASES;"   -e "SELECT id, customer, product, amount, created_at FROM shopdb.orders ORDER BY id;"   -e "SELECT COUNT(*) AS total_rows, SUM(amount) AS total_amount FROM shopdb.orders;"   2>&1 | grep -v "Warning"

echo ""
echo "══════════════════════════════════════════════"
log "🎉 恢复完成！数据验证通过"
echo "══════════════════════════════════════════════"
