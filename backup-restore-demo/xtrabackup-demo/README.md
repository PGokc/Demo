# XtraBackup 备份恢复 Demo

> MySQL 8.0 + Percona XtraBackup 8.0 完整演示

## 目录结构

```
xtrabackup-demo/
├── docker-compose.yml           # 容器编排
├── run_demo.sh                  # 一键运行脚本（在宿主机执行）
├── conf/
│   └── my.cnf                   # MySQL 配置（开启 binlog/GTID）
└── scripts/
    ├── seed_data.sql            # 初始演示数据
    ├── 01_full_backup.sh        # 写入数据 + 全量备份（xb_tool 容器内）
    ├── 02_incremental_backup.sh # 写入数据 + 增量备份两轮（xb_tool 容器内）
    ├── 03_simulate_crash.sh     # 模拟删库故障（xb_tool 容器内）
    ├── 04_restore.sh            # 预检：copy-back 到临时目录（xb_tool 容器内）
    └── 05_verify.sh             # 正式恢复 + 验证数据（宿主机执行）
```

## 快速开始

```bash
# 1. 进入目录
cd xtrabackup-demo

# 2. 给脚本加执行权限
chmod +x run_demo.sh scripts/*.sh

# 3. 一键运行完整演示（有交互暂停，按 Enter 分步查看）
./run_demo.sh
```

## 分步手动运行

```bash
# 启动容器（会自动初始化 volume 权限）
docker compose down -v
docker compose up -d
sleep 20

# Step 1: 写入初始数据 + 全量备份（在 xb_tool 容器内执行）
docker exec xb_tool bash /scripts/01_full_backup.sh

# Step 2: 写入新数据 + 增量备份两轮（在 xb_tool 容器内执行）
docker exec xb_tool bash /scripts/02_incremental_backup.sh

# Step 3: 模拟删库（在 xb_tool 容器内执行）
docker exec xb_tool bash /scripts/03_simulate_crash.sh

# Step 4: 预检恢复，copy-back 到临时目录验证文件（在 xb_tool 容器内执行）
docker exec xb_tool bash /scripts/04_restore.sh

# Step 5: 正式恢复 + 验证数据（在宿主机执行，会停库再启库）
bash scripts/05_verify.sh
```

## 各步骤说明

| 步骤 | 运行环境 | 说明 |
|------|---------|------|
| 01_full_backup.sh | xb_tool 容器 | 写入5条订单 → 全量备份 → prepare（apply-log-only）|
| 02_incremental_backup.sh | xb_tool 容器 | 再写5条 → 增量备份1、2 → 合并 prepare |
| 03_simulate_crash.sh | xb_tool 容器 | 写入脏数据 → DROP DATABASE shopdb |
| 04_restore.sh | xb_tool 容器 | copy-back 到 /backup/restore_tmp，验证文件完整性，不停库 |
| 05_verify.sh | **宿主机** | 停库 → 清空数据目录 → copy-back → 修复权限 → 启库 → 查询验证 |

## 备份流程图

```
MySQL 写入数据（5条）
       │
       ▼
[01] 全量备份 /backup/full
       │  --backup + --prepare --apply-log-only
       │
       ▼
MySQL 写入新数据（+3条）
       │
       ▼
[02] 增量备份1 /backup/inc1（基于 full）
       │
       ▼
MySQL 继续写入（+2条）
       │
       ▼
[02] 增量备份2 /backup/inc2（基于 inc1）
       │  prepare 合并: full ← inc1 ← inc2
       │
       ▼
[03] 故障模拟：DROP DATABASE shopdb
       │
       ▼
[04] 预检：copy-back → /backup/restore_tmp（验证文件）
       │
       ▼
[05] 正式恢复：copy-back → /var/lib/mysql → 启动 MySQL
       │
       ▼
    查询验证：10条数据全部恢复 ✅
```

## 注意事项

- `docker compose` 创建的 volume 名带项目前缀，如 `xtrabackup-demo_backup_data`
- `05_verify.sh` 必须在**宿主机**执行，容器内没有 docker 命令
- 每次重新演示用 `docker compose down -v` 清理，避免残留 volume 干扰

## 环境要求

- Docker Engine 20.10+
- docker compose v2
- 磁盘空间 2GB+

## 清理环境

```bash
docker compose down -v   # 删除容器和全部数据卷，彻底清理
```
