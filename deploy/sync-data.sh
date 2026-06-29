#!/usr/bin/env bash
# Sync runtime data/ between local dev and a remote RoundTable host (rsync).
#
# Does NOT sync: deploy/.env, data/_templates/ (tracked in git).
#
# Usage:
#   cp deploy/sync-data.env.example deploy/sync-data.env   # once
#   ./deploy/sync-data.sh push              # local  → remote
#   ./deploy/sync-data.sh pull              # remote → local
#   ./deploy/sync-data.sh push --dry-run    # preview only
#   ./deploy/sync-data.sh push --delete     # mirror (remove extras on target)
#
# Tip: stop server / docker on both sides before syncing roundtable.db (SQLite WAL).

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
ENV_FILE="$SCRIPT_DIR/sync-data.env"

SSH_CONTROL_DIR=""
RSYNC_SSH=""

usage() {
	cat <<'EOF'
Usage: deploy/sync-data.sh <push|pull|status> [--dry-run] [--delete]

  push    local data/  → remote
  pull    remote data/ → local
  status  dry-run summary (no changes)

Config: deploy/sync-data.env (see deploy/sync-data.env.example)
EOF
}

cleanup_ssh() {
	if [ -n "$RSYNC_SSH" ] && [ -n "${SYNC_HOST:-}" ]; then
		"$RSYNC_SSH" -O exit "$SYNC_HOST" 2>/dev/null || true
	fi
	if [ -n "$SSH_CONTROL_DIR" ] && [ -d "$SSH_CONTROL_DIR" ]; then
		rm -rf "$SSH_CONTROL_DIR"
	fi
}

init_ssh() {
	if [ -n "${SYNC_SSH_PASSWORD:-}" ] && ! command -v sshpass >/dev/null 2>&1; then
		echo "SYNC_SSH_PASSWORD is set but sshpass not found" >&2
		echo "  macOS: brew install hudochenkov/sshpass/sshpass" >&2
		echo "  更推荐: ssh-copy-id $SYNC_HOST  配置免密，无需存密码" >&2
		exit 1
	fi

	SSH_CONTROL_DIR="/tmp/rts$$"
	mkdir -p "$SSH_CONTROL_DIR"
	chmod 700 "$SSH_CONTROL_DIR"

	local wrap="$SSH_CONTROL_DIR/rsync-ssh"
	{
		echo '#!/usr/bin/env bash'
		if [ -n "${SYNC_SSH_PASSWORD:-}" ]; then
			printf 'export SSHPASS=%q\n' "$SYNC_SSH_PASSWORD"
			printf 'exec sshpass -e ssh '
		else
			printf 'exec ssh '
		fi
		# shellcheck disable=SC2086
		printf '%s ' $SYNC_SSH_OPTS
		# %C = hash of host/port/user — keeps ControlPath under macOS 104-byte limit
		printf '%q ' \
			-o ControlMaster=auto \
			"-o ControlPath=${SSH_CONTROL_DIR}/%C" \
			-o ControlPersist=120
		echo '"$@"'
	} >"$wrap"
	chmod 700 "$wrap"
	RSYNC_SSH="$wrap"
	trap cleanup_ssh EXIT
}

run_ssh() {
	"$RSYNC_SSH" "$@"
}

if [ $# -lt 1 ]; then
	usage >&2
	exit 1
fi

ACTION="$1"
shift

DRY_RUN=0
DELETE=0
while [ $# -gt 0 ]; do
	case "$1" in
	--dry-run) DRY_RUN=1 ;;
	--delete) DELETE=1 ;;
	-h | --help)
		usage
		exit 0
		;;
	*)
		echo "unknown option: $1" >&2
		usage >&2
		exit 1
		;;
	esac
	shift
done

if [ ! -f "$ENV_FILE" ]; then
	echo "missing $ENV_FILE — copy deploy/sync-data.env.example and edit SYNC_HOST / SYNC_REMOTE_ROOT" >&2
	exit 1
fi

# shellcheck disable=SC1090
source "$ENV_FILE"

SYNC_HOST="${SYNC_HOST:-}"
SYNC_REMOTE_ROOT="${SYNC_REMOTE_ROOT:-}"
SYNC_SSH_OPTS="${SYNC_SSH_OPTS:-}"
SYNC_SSH_PASSWORD="${SYNC_SSH_PASSWORD:-}"
SYNC_DELETE="${SYNC_DELETE:-0}"

if [ -z "$SYNC_HOST" ] || [ -z "$SYNC_REMOTE_ROOT" ]; then
	echo "SYNC_HOST and SYNC_REMOTE_ROOT must be set in $ENV_FILE" >&2
	exit 1
fi

if [ "$DELETE" -eq 1 ] || [ "$SYNC_DELETE" = "1" ]; then
	DELETE=1
fi

init_ssh

REMOTE="${SYNC_HOST}:${SYNC_REMOTE_ROOT}"

RUNTIME_PATHS=(
	data/workspaces
	data/profiles/participants
	data/profiles/principals
	data/profiles/moderator
	data/knowledge/participants
	data/knowledge/principals
	data/knowledge/shared
	data/transport
	data/logs
)

RSYNC_FLAGS=(-a -z --human-readable)
if [ "$DRY_RUN" -eq 1 ]; then
	RSYNC_FLAGS+=(--dry-run --itemize-changes)
fi
if [ "$DELETE" -eq 1 ]; then
	RSYNC_FLAGS+=(--delete)
fi
RSYNC_FLAGS+=(-e "$RSYNC_SSH")

ensure_local_dirs() {
	for rel in "${RUNTIME_PATHS[@]}"; do
		mkdir -p "$REPO_ROOT/$rel"
	done
}

ensure_remote_dirs() {
	local mk=""
	for rel in "${RUNTIME_PATHS[@]}"; do
		mk="${mk}mkdir -p '${SYNC_REMOTE_ROOT}/${rel}'; "
	done
	run_ssh "$SYNC_HOST" "${mk}true"
}

sync_path() {
	local rel="$1"
	local src dst

	case "$ACTION" in
	push)
		src="$REPO_ROOT/$rel/"
		dst="$REMOTE/$rel/"
		;;
	pull)
		src="$REMOTE/$rel/"
		dst="$REPO_ROOT/$rel/"
		;;
	status)
		src="$REPO_ROOT/$rel/"
		dst="$REMOTE/$rel/"
		echo "== $rel (local → remote preview) =="
		rsync "${RSYNC_FLAGS[@]}" "$src" "$dst" || true
		echo "== $rel (remote → local preview) =="
		rsync "${RSYNC_FLAGS[@]}" "$dst" "$src" || true
		return 0
		;;
	*)
		echo "unknown action: $ACTION" >&2
		exit 1
		;;
	esac

	if [ ! -d "$REPO_ROOT/$rel" ] && [ "$ACTION" = "push" ]; then
		mkdir -p "$REPO_ROOT/$rel"
	fi

	echo "→ rsync $rel ($ACTION)"
	rsync "${RSYNC_FLAGS[@]}" "$src" "$dst"
}

sync_sqlite() {
	local db="data/roundtable.db"
	local src dst

	case "$ACTION" in
	push)
		[ -f "$REPO_ROOT/$db" ] || return 0
		src="$REPO_ROOT/$db"
		dst="$REMOTE/$db"
		;;
	pull)
		src="$REMOTE/$db"
		dst="$REPO_ROOT/$db"
		;;
	status)
		if [ -f "$REPO_ROOT/$db" ]; then
			echo "== $db (local → remote preview) =="
			rsync "${RSYNC_FLAGS[@]}" "$REPO_ROOT/$db" "$REMOTE/$db" || true
		fi
		echo "== $db (remote → local preview) =="
		rsync "${RSYNC_FLAGS[@]}" "$REMOTE/$db" "$REPO_ROOT/$db" 2>/dev/null || true
		return 0
		;;
	esac

	echo "→ rsync $db ($ACTION)"
	rsync "${RSYNC_FLAGS[@]}" "$src" "$dst"
}

warn_if_servers_running() {
	if pgrep -f 'roundtable-server|roundtable-discord|apps/server/cmd/roundtable|apps/server/cmd/discord' >/dev/null 2>&1; then
		echo "warning: RoundTable process(es) running locally — stop before syncing SQLite to avoid WAL corruption" >&2
	fi
	if [ "$ACTION" != "status" ] && [ "$DRY_RUN" -eq 0 ]; then
		if run_ssh "$SYNC_HOST" "pgrep -f 'roundtable-server|roundtable-discord' >/dev/null 2>&1"; then
			echo "warning: RoundTable process(es) running on $SYNC_HOST — consider: docker compose stop server" >&2
		fi
	fi
}

case "$ACTION" in
push | pull | status) ;;
*)
	echo "unknown action: $ACTION (want push, pull, or status)" >&2
	usage >&2
	exit 1
	;;
esac

echo "RoundTable data sync · action=$ACTION · remote=$REMOTE"
if [ -n "$SYNC_SSH_PASSWORD" ]; then
	echo "auth: password (sshpass) + connection reuse"
fi
if [ "$DELETE" -eq 1 ]; then
	echo "delete mode: ON (target will mirror source; extra files removed)"
fi
if [ "$DRY_RUN" -eq 1 ]; then
	echo "dry-run: no files will be changed"
fi

warn_if_servers_running

if [ "$ACTION" = "push" ]; then
	ensure_local_dirs
	ensure_remote_dirs
elif [ "$ACTION" = "pull" ]; then
	ensure_local_dirs
	ensure_remote_dirs
else
	ensure_local_dirs
fi

for rel in "${RUNTIME_PATHS[@]}"; do
	sync_path "$rel"
done

sync_sqlite

if [ "$ACTION" = "status" ]; then
	echo "done (status preview only)"
else
	echo "done ($ACTION)"
fi
