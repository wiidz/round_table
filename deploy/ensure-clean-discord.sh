#!/bin/sh
# Remove orphan Discord transport after compose merge (single server + Supervisor).
# Run on the Linux host before docker compose up -d --build.
set -e
cd "$(dirname "$0")/.."

echo "Stopping compose stack…"
docker compose down --remove-orphans 2>/dev/null || true

if docker ps -a --format '{{.Names}}' | grep -qx roundtable-discord; then
	echo "Removing orphan container roundtable-discord…"
	docker rm -f roundtable-discord
fi

remaining=$(pgrep -fc roundtable-discord 2>/dev/null || true)
if [ "${remaining:-0}" -gt 0 ]; then
	echo "WARNING: roundtable-discord still running on host (count=$remaining):"
	pgrep -af roundtable-discord || true
	echo "Kill manually if needed: pkill -f roundtable-discord"
	exit 1
fi

echo "Discord transport clean — only roundtable-server should run transport via Supervisor."
