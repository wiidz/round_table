#!/bin/sh
# Fix Docker named-volume ownership (often root) before dropping to roundtable (uid 1000).
set -e

fixdir() {
	dir="$1"
	mkdir -p "$dir"
	chown -R roundtable:roundtable "$dir"
}

fixdir /app/data/workspaces
fixdir /app/data/profiles/participants
fixdir /app/data/profiles/principals
fixdir /app/data/profiles/moderator
fixdir /app/data/knowledge/participants
fixdir /app/data/knowledge/principals
fixdir /app/data/knowledge/shared
fixdir /app/data/transport

exec su-exec roundtable "$@"
