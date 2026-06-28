#!/bin/sh
# Ensure bind-mounted data dirs are writable by roundtable (uid 1000).
set -e

fixdir() {
	dir="$1"
	mkdir -p "$dir"
	chown -R roundtable:roundtable "$dir"
	chmod -R u+rwX,g+rwX "$dir"
}

fixdir /app/data/workspaces
fixdir /app/data/profiles/participants
fixdir /app/data/profiles/principals
fixdir /app/data/profiles/moderator
fixdir /app/data/knowledge/participants
fixdir /app/data/knowledge/principals
fixdir /app/data/knowledge/shared
fixdir /app/data/transport
fixdir /app/data/logs

# Host network: bind directly on the host. Prefer deploy/.env ROUND_TABLE_ADDR; else derive from port vars.
if [ -z "$ROUND_TABLE_ADDR" ]; then
	port="${ROUND_TABLE_HTTP_PORT:-${ROUND_TABLE_WEB_PORT:-7777}}"
	port="${port#:}"
	export ROUND_TABLE_ADDR=":${port}"
fi

if ! su-exec roundtable sh -c 'touch /app/data/workspaces/.write-test && rm -f /app/data/workspaces/.write-test'; then
	echo "roundtable: FATAL: cannot write to /app/data/workspaces (check bind mount permissions)" >&2
	exit 1
fi

exec su-exec roundtable "$@"
