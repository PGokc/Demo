#!/bin/bash
# ============================================================
# 04_restore.sh — 预演恢复：把备份 copy 到 restore_tmp 验证文件完整性
# 运行环境: xb_tool 容器
# 注意: 这一步只是预检，不影响正在运行的 MySQL
# ============================================================
set -euo pipefail

BACKUP_DIR="/backup/full"
RESTORE_TMP="/backup/restore_tmp"

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; NC='\033[0m'
log()  { echo -e "${GREEN}[$(date '+%H:%M:%S')] $*${NC}"; }
warn() { echo -e "${YELLOW}[$(date '+%H:%M:%S')] $*${NC}"; }
info() { echo -e "${CYAN}[$(date '+%H:%M:%S')] $*${NC}"; }

# ── 检查备份是否 prepare 完成 ────────────────────────────────
log "🔍 检查备份状态..."
if ! grep -q "to_lsn" "$BACKUP_DIR/xtrabackup_checkpoints" 2>/dev/null; then
  echo "❌ 备份未 prepare，请先运行 02_incremental_backup.sh"
  exit 1
fi

BACKUP_TYPE=$(grep "backup_type" "$BACKUP_DIR/xtrabackup_checkpoints" | awk '{print $3}')
info "备份类型: $BACKUP_TYPE（log-applied 表示已 prepare 完成）"
echo ""

# ── 预演：copy-back 到临时目录 ───────────────────────────────
log "📂 预演恢复：copy-back → $RESTORE_TMP"
rm -rf "$RESTORE_TMP"
mkdir -p "$RESTORE_TMP"

xtrabackup \
  --copy-back \
  --target-dir="$BACKUP_DIR" \
  --datadir="$RESTORE_TMP" \
  --no-server-version-check \
  2>&1 | grep -E "completed OK|error|Error" | tail -5

log "✅ 预演完成，数据已复制到 $RESTORE_TMP"
echo ""

info "文件统计:"
echo "  备份目录 .ibd: $(find $BACKUP_DIR  -name '*.ibd' | wc -l) 个"
echo "  预演目录 .ibd: $(find $RESTORE_TMP -name '*.ibd' | wc -l) 个"
echo "  预演目录大小:  $(du -sh $RESTORE_TMP | cut -f1)"
echo ""

if [ -d "$RESTORE_TMP/shopdb" ]; then
  log "✅ shopdb 库文件存在"
  ls -lh "$RESTORE_TMP/shopdb/"
else
  echo "❌ shopdb 目录未找到，备份可能有问题"
  exit 1
fi

echo ""
warn "👉 预检通过，下一步运行 05_verify.sh 执行真正恢复并验证数据"
