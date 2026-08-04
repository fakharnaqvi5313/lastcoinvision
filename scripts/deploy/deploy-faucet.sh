#!/usr/bin/env bash
# One-command deploy for faucet/ (faucet.lastcoinvision.com) to the
# production VPS: builds locally, syncs the build output + package
# manifests, installs production deps on the server (Node native deps
# shouldn't be rsynced cross-platform from macOS), and restarts the
# service. Never touches the server's .env — that holds the faucet's
# funded wallet mnemonic and is created once, by hand, on the server.
#
# Real host/path/key values live in a local, gitignored config file
# (scripts/deploy/config.local.sh — see config.example.sh) rather than
# hardcoded here.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
FAUCET_DIR="$ROOT_DIR/faucet"
CONFIG_FILE="$ROOT_DIR/scripts/deploy/config.local.sh"

if [ ! -f "$CONFIG_FILE" ]; then
  echo "FAIL: $CONFIG_FILE not found."
  echo "Copy scripts/deploy/config.example.sh to config.local.sh and fill in your VPS details."
  exit 1
fi
# shellcheck source=/dev/null
source "$CONFIG_FILE"

: "${VPS_HOST:?set in config.local.sh}"
: "${VPS_USER:?set in config.local.sh}"
: "${VPS_ADMIN_USER:?set in config.local.sh}"
: "${SSH_KEY:?set in config.local.sh}"
: "${REMOTE_FAUCET_DIR:?set in config.local.sh}"
: "${FAUCET_HEALTH_URL:?set in config.local.sh}"

ssh_cmd() { ssh -i "$SSH_KEY" -o BatchMode=yes "$VPS_ADMIN_USER@$VPS_HOST" "$@"; }
ssh_deploy_cmd() { ssh -i "$SSH_KEY" -o BatchMode=yes "$VPS_USER@$VPS_HOST" "$@"; }

echo "==> Building faucet/"
(cd "$FAUCET_DIR" && rm -rf dist && npm run build)

echo "==> Syncing dist/ + package manifests to $VPS_USER@$VPS_HOST:$REMOTE_FAUCET_DIR/ (leaving .env alone)"
rsync -az --delete --exclude .env \
  -e "ssh -i $SSH_KEY -o BatchMode=yes" \
  "$FAUCET_DIR/dist" "$FAUCET_DIR/package.json" "$FAUCET_DIR/package-lock.json" \
  "$VPS_USER@$VPS_HOST:$REMOTE_FAUCET_DIR/"

echo "==> Installing production dependencies on the server"
ssh_deploy_cmd "cd $REMOTE_FAUCET_DIR && npm ci --omit=dev"

echo "==> Restarting faucet.service"
ssh_cmd "systemctl restart faucet.service"

echo "==> Waiting for the app to come back up"
for i in $(seq 1 15); do
  if ssh_cmd "curl -sf -o /dev/null http://127.0.0.1:8088/health"; then
    break
  fi
  if [ "$i" -eq 15 ]; then
    echo "FAIL: service did not come back up on the VPS after restart"
    ssh_cmd "journalctl -u faucet.service -n 40 --no-pager"
    exit 1
  fi
  sleep 1
done

echo "==> Verifying $FAUCET_HEALTH_URL"
code="$(curl -s -o /dev/null -w '%{http_code}' "$FAUCET_HEALTH_URL")"
if [ "$code" != "200" ]; then
  echo "FAIL: $FAUCET_HEALTH_URL returned HTTP $code"
  exit 1
fi

echo "==> Deployed successfully: $FAUCET_HEALTH_URL (HTTP $code)"
