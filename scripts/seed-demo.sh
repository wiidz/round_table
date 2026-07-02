#!/usr/bin/env bash
# Copy demo templates into runtime data/ (gitignored).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DEMO="$ROOT/data/_templates/demo"
SCENARIO_3R="$ROOT/data/_templates/scenarios/3-round-debate"
BRIEF_SRC="$ROOT/data/_templates/briefs/decision-review/BRIEF.yaml"

cd "$ROOT"

mkdir -p \
  data/workspaces \
  data/profiles/participants \
  data/profiles/principals \
  data/briefs

echo "→ workspace mtg-demo-001"
rm -rf data/workspaces/mtg-demo-001
cp -R "$DEMO/workspaces/mtg-demo-001" data/workspaces/

echo "→ participants (skeptic, pragmatist)"
for id in skeptic pragmatist; do
  mkdir -p "data/profiles/participants/$id"
  cp "$SCENARIO_3R/profiles/$id/"*.md "data/profiles/participants/$id/"
done

echo "→ principal demo"
mkdir -p data/profiles/principals/demo
cp "$DEMO/profiles/principals/demo/USER.md" data/profiles/principals/demo/

echo "→ brief template decision-review"
mkdir -p data/briefs/decision-review
cp "$BRIEF_SRC" data/briefs/decision-review/

echo ""
echo "Demo data ready under data/"
echo "  make server-dev   # terminal 1"
echo "  make web-dev      # terminal 2"
echo "  Open Meetings → mtg-demo-001"
