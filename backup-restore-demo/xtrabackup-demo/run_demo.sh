#!/bin/bash
# ============================================================
# run_demo.sh — 一键运行完整演示流程
# ============================================================
set -euo pipefail

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; BOLD='\033[1m'; NC='\033[0m'
log()   { echo -e "${GREEN}${BOLD}[DEMO] $*${NC}"; }
info()  { echo -e "${CYAN}$*${NC}"; }
warn()  { echo -e "${YELLOW}$*${NC}"; }
sep()   { echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"; }

sep
echo -e "${BOLD}  🐬 XtraBackup 备份恢复完整演示${NC}"
sep
echo ""

if ! command -v docker &>/dev/null; then
  echo "❌ 未找到 docker，请先安装 Docker Engine"
  exit 1
fi

XB_EXEC="docker exec xb_tool bash"

# ── Step 0: 清理旧环境，重新启动 ────────────────────────────
log "Step 0: 清理旧环境，启动容器..."
docker compose down -v 2>/dev/null || true
docker compose up -d

echo ""
warn "⏳ 等待 init-perms 完成权限初始化..."
for i in $(seq 1 30); do
  STATUS=$(docker inspect --format='{{.State.Status}}' xb_init 2>/dev/null || echo "missing")
  if [ "$STATUS" = "exited" ]; then
    EXIT_CODE=$(docker inspect --format='{{.State.ExitCode}}' xb_init)
    if [ "$EXIT_CODE" = "0" ]; then
      echo "  ✅ 权限初始化成功"
      break
    else
      echo "  ❌ init-perms 失败，exit code: $EXIT_CODE"
      docker logs xb_init
      exit 1
    fi
  fi
  echo -n "."
  sleep 1
done

warn "⏳ 等待 MySQL 健康检查通过..."
for i in $(seq 1 40); do
  HEALTH=$(docker inspect --format='{{.State.Health.Status}}' xb_mysql 2>/dev/null || echo "none")
  if [ "$HEALTH" = "healthy" ]; then
    echo "  ✅ MySQL 已就绪"
    break
  fi
  echo -n "."
  sleep 2
done
echo ""
docker compose ps
echo ""

# ── Step 1: 全量备份 ─────────────────────────────────────────
sep
log "Step 1: 写入初始数据 + 全量备份"
sep
$XB_EXEC /scripts/01_full_backup.sh
echo ""
read -p "按 Enter 继续下一步..."

# ── Step 2: 增量备份 ─────────────────────────────────────────
sep
log "Step 2: 写入新数据 + 增量备份（两轮）"
sep
$XB_EXEC /scripts/02_incremental_backup.sh
echo ""
read -p "按 Enter 继续下一步..."

# ── Step 3: 模拟故障 ─────────────────────────────────────────
sep
log "Step 3: 模拟灾难 — 删库！"
sep
$XB_EXEC /scripts/03_simulate_crash.sh
echo ""
read -p "按 Enter 继续恢复..."

# ── Step 4: 预检恢复（容器内，不停库）────────────────────────
sep
log "Step 4: 预检 — copy-back 到临时目录验证文件完整性"
sep
$XB_EXEC /scripts/04_restore.sh
echo ""
read -p "按 Enter 执行真正恢复..."

# ── Step 5: 真正恢复（宿主机执行，需要停库）──────────────────
sep
log "Step 5: 正式恢复 — 停库 → copy-back → 启库 → 验证数据"
sep
bash "$(dirname "$0")/scripts/05_verify.sh"
echo ""

sep
log "🎉 演示完成！"
sep
info ""
info "常用命令:"
info "  进入 MySQL:     docker exec -it xb_mysql mysql -uroot -pRoot1234!"
info "  进入备份容器:   docker exec -it xb_tool bash"
info "  查看备份文件:   docker exec xb_tool ls -lh /backup/"
info "  清理所有数据:   docker compose down -v"
info ""
