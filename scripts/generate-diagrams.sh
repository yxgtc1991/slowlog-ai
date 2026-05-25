#!/usr/bin/env bash
# 从 docs/diagrams/*.mmd 生成 PNG（GoLand 可直接预览 .png）
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DIR="$ROOT/docs/diagrams"
TOOLS="$ROOT/scripts/diagram-tools"
MMDC="$TOOLS/node_modules/.bin/mmdc"

if [[ ! -x "$MMDC" ]]; then
  echo "installing @mermaid-js/mermaid-cli in scripts/diagram-tools ..."
  (cd "$TOOLS" && npm init -y >/dev/null 2>&1 && npm install @mermaid-js/mermaid-cli)
fi

if ! "$TOOLS/node_modules/puppeteer-core/package.json" >/dev/null 2>&1; then
  :
fi
if ! ls "$TOOLS/node_modules/puppeteer"/.local-chromium 2>/dev/null && \
   ! ls /var/folders/*/T/cursor-sandbox-cache/*/puppeteer/chrome-headless-shell 2>/dev/null; then
  echo "installing headless chrome for puppeteer ..."
  (cd "$TOOLS" && npx puppeteer browsers install chrome-headless-shell)
fi

cd "$DIR"
for mmd in *.mmd; do
  [[ -f "$mmd" ]] || continue
  out="${mmd%.mmd}.png"
  echo "→ $out"
  "$MMDC" -i "$mmd" -o "$out" -b white -w 1200
done
echo "OK: $DIR/*.png"
