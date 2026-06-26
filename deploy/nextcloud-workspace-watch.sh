#!/bin/bash
# ============================================
# RoundTable Workspace → Nextcloud 实时同步脚本
# 检测 data/workspaces 变更后自动触发 occ files:scan
# 适合通过宝塔的计划任务管理（建议每 1 分钟）
# ============================================

set -euo pipefail

# ====== 配置 ======
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
WORKSPACE="${ROUND_TABLE_WORKSPACE:-$REPO_ROOT/data/workspaces}"
NEXTCLOUD_CONTAINER="${NEXTCLOUD_CONTAINER:-nas-nextcloud-1}"
NEXTCLOUD_USER="${NEXTCLOUD_USER:-hujiayilu}"
OCC_PATH="${NEXTCLOUD_USER}/files/round_table_workspace"
LOCKFILE="/tmp/roundtable-nextcloud-workspace-sync.lock"
MTIME_CACHE="/tmp/roundtable-nextcloud-workspace-mtime"

# ====== 防并发锁 ======
exec 200>"$LOCKFILE"
flock -n 200 || exit 0

# ====== 检查 workspace 目录 ======
if [ ! -d "$WORKSPACE" ]; then
	echo "roundtable nextcloud watch: workspace not found: $WORKSPACE" >&2
	exit 1
fi

# ====== 检查是否有文件变更 ======
LATEST=0
if [ -f "$MTIME_CACHE" ]; then
	LATEST=$(find "$WORKSPACE" -type f -newer "$MTIME_CACHE" 2>/dev/null | wc -l)
else
	LATEST=1
fi

if [ "$LATEST" -eq 0 ]; then
	exit 0
fi

# ====== 触发 nextcloud 文件扫描 ======
docker exec -u www-data "$NEXTCLOUD_CONTAINER" \
	php occ files:scan --path="$OCC_PATH" 2>/dev/null

# ====== 更新 mtime 标记 ======
touch "$MTIME_CACHE"
