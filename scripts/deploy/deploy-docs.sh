#!/usr/bin/env bash
# One-command deploy for docs/ (docs.lastcoinvision.com) to the production VPS:
# builds the static Docusaurus site locally and rsyncs it to the nginx docroot.
# Unlike deploy-web.sh, there's no systemd service or app process to restart —
# nginx serves the static files directly, so a content-only deploy needs no
# server-side action beyond the rsync itself. Nginx/SSL vhost are assumed
# already set up on the server (see README.md).
#
# Real host/path/key values live in a local, gitignored config file
# (scripts/deploy/config.local.sh — see config.example.sh) rather than
# hardcoded here, since they're this server's actual infrastructure
# details, not something that belongs in git history.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DOCS_DIR="$ROOT_DIR/docs"
CONFIG_FILE="$ROOT_DIR/scripts/deploy/config.local.sh"

if [ ! -f "$CONFIG_FILE" ]; then
  echo "FAIL: $CONFIG_FILE not found."
  echo "Copy scripts/deploy/config.example.sh to config.local.sh and fill in your VPS details."
  exit 1
fi
# shellcheck source=/dev/null
source "$CONFIG_FILE"

: "${VPS_HOST:?set in config.local.sh}"
: "${VPS_USER:?set in config.local.sh}"        # owns the docroot rsync writes to
: "${SSH_KEY:?set in config.local.sh}"
: "${REMOTE_DOCS_DIR:?set in config.local.sh}"
: "${DOCS_HEALTH_URL:?set in config.local.sh}"

echo "==> Building docs/ (static Docusaurus output)"
(cd "$DOCS_DIR" && rm -rf build .docusaurus && npm run build)

echo "==> Syncing to $VPS_USER@$VPS_HOST:$REMOTE_DOCS_DIR/"
rsync -az --delete \
  -e "ssh -i $SSH_KEY -o BatchMode=yes" \
  "$DOCS_DIR/build/" \
  "$VPS_USER@$VPS_HOST:$REMOTE_DOCS_DIR/"

echo "==> Verifying $DOCS_HEALTH_URL"
code="$(curl -s -o /dev/null -w '%{http_code}' "$DOCS_HEALTH_URL")"
if [ "$code" != "200" ]; then
  echo "FAIL: $DOCS_HEALTH_URL returned HTTP $code"
  exit 1
fi

echo "==> Deployed successfully: $DOCS_HEALTH_URL (HTTP $code)"
