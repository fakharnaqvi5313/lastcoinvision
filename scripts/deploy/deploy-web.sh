#!/usr/bin/env bash
# One-command deploy for web/ (lastcoinvision.com) to the production VPS:
# builds the Next.js standalone output locally, rsyncs it to the existing
# systemd-managed deployment, and restarts the service. Nginx/SSL/systemd
# unit are assumed already set up on the server (see README.md) — this
# script only ships new application code, it never touches server config.
#
# Real host/path/key values live in a local, gitignored config file
# (scripts/deploy/config.local.sh — see config.example.sh) rather than
# hardcoded here, since they're this server's actual infrastructure
# details, not something that belongs in git history.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
WEB_DIR="$ROOT_DIR/web"
CONFIG_FILE="$ROOT_DIR/scripts/deploy/config.local.sh"

if [ ! -f "$CONFIG_FILE" ]; then
  echo "FAIL: $CONFIG_FILE not found."
  echo "Copy scripts/deploy/config.example.sh to config.local.sh and fill in your VPS details."
  exit 1
fi
# shellcheck source=/dev/null
source "$CONFIG_FILE"

: "${VPS_HOST:?set in config.local.sh}"
: "${VPS_USER:?set in config.local.sh}"          # owns the app files rsync writes to
: "${VPS_ADMIN_USER:?set in config.local.sh}"    # used for systemctl (needs root or passwordless sudo)
: "${SSH_KEY:?set in config.local.sh}"
: "${REMOTE_APP_DIR:?set in config.local.sh}"
: "${SERVICE_NAME:?set in config.local.sh}"
: "${APP_PORT:?set in config.local.sh}"
: "${HEALTH_URL:?set in config.local.sh}"

ssh_cmd() { ssh -i "$SSH_KEY" -o BatchMode=yes "$VPS_ADMIN_USER@$VPS_HOST" "$@"; }

echo "==> Building web/ (production standalone output)"
# NEXT_PUBLIC_LCV_RPC_URL is inlined into the client bundle at build time
# (Next.js convention for NEXT_PUBLIC_* vars) -- without it the homepage's
# live-stats strip silently falls back to localhost:26657 in every real
# visitor's browser and always shows "unavailable" (see DECISIONS.md,
# 2026-08-05). Pointed at the real public testnet RPC by default here;
# override by exporting NEXT_PUBLIC_LCV_RPC_URL before running this script.
(cd "$WEB_DIR" && rm -rf .next && NEXT_PUBLIC_LCV_RPC_URL="${NEXT_PUBLIC_LCV_RPC_URL:-https://rpc.lastcoinvision.com}" npm run build)

echo "==> Assembling standalone output (public/ + .next/static/)"
cp -r "$WEB_DIR/public" "$WEB_DIR/.next/standalone/public"
mkdir -p "$WEB_DIR/.next/standalone/.next"
cp -r "$WEB_DIR/.next/static" "$WEB_DIR/.next/standalone/.next/static"

echo "==> Syncing to $VPS_USER@$VPS_HOST:$REMOTE_APP_DIR/.next/standalone/"
rsync -az --delete \
  -e "ssh -i $SSH_KEY -o BatchMode=yes" \
  "$WEB_DIR/.next/standalone/" \
  "$VPS_USER@$VPS_HOST:$REMOTE_APP_DIR/.next/standalone/"

echo "==> Restarting $SERVICE_NAME.service"
ssh_cmd "systemctl restart $SERVICE_NAME.service"

echo "==> Waiting for the app to come back up"
for i in $(seq 1 15); do
  if ssh_cmd "curl -sf -o /dev/null http://127.0.0.1:$APP_PORT/"; then
    break
  fi
  if [ "$i" -eq 15 ]; then
    echo "FAIL: service did not come back up on the VPS after restart"
    ssh_cmd "journalctl -u $SERVICE_NAME.service -n 40 --no-pager"
    exit 1
  fi
  sleep 1
done

echo "==> Verifying $HEALTH_URL"
code="$(curl -s -o /dev/null -w '%{http_code}' "$HEALTH_URL")"
if [ "$code" != "200" ]; then
  echo "FAIL: $HEALTH_URL returned HTTP $code"
  exit 1
fi

echo "==> Deployed successfully: $HEALTH_URL (HTTP $code)"
