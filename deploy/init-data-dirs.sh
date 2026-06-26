#!/bin/sh
# Create host data directories for bind mounts (run once on the server).
set -e
cd "$(dirname "$0")/.."
mkdir -p \
  data/workspaces \
  data/profiles/participants \
  data/profiles/principals \
  data/profiles/moderator \
  data/knowledge/participants \
  data/knowledge/principals \
  data/knowledge/shared \
  data/transport
echo "data/ directories ready under $(pwd)/data"
