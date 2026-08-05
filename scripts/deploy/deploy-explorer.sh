#!/usr/bin/env bash
# One-command deploy for explorer/ (explorer.lastcoinvision.com) to the
# production VPS: builds ping.pub's static output (Docker, matching
# explorer/Dockerfile's build stage exactly) with CHAIN_CONFIG defaulted to
# lcv-testnet-1.json, then rsyncs it to the nginx docroot. Same pattern as
# deploy-docs.sh — nginx serves the static files directly, no app process to
# restart. Nginx vhost + SSL are assumed already set up on the server (see
# README.md).
#
# If `docker build` hits registry timeouts (seen once on macOS Docker
# Desktop — see DECISIONS.md), build natively instead as a one-off:
#   cd explorer/vendor/ping-pub
#   rm -f ./chains/testnet/*.json ./chains/mainnet/*.json
#   cp ../../config/lcv-testnet-1.json ./chains/testnet/ ./chains/mainnet/
#   yarn install --ignore-engines --frozen-lockfile && yarn build
#   git checkout -- . && git clean -fd chains/   # never leave the submodule modified
# then rsync vendor/ping-pub/dist/ to $REMOTE_EXPLORER_DIR by hand.
#
# Real host/path/key values live in a local, gitignored config file
# (scripts/deploy/config.local.sh — see config.example.sh) rather than
# hardcoded here, since they're this server's actual infrastructure
# details, not something that belongs in git history.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
EXPLORER_DIR="$ROOT_DIR/explorer"
CONFIG_FILE="$ROOT_DIR/scripts/deploy/config.local.sh"
CHAIN_CONFIG="${CHAIN_CONFIG:-lcv-testnet-1.json}"

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
: "${REMOTE_EXPLORER_DIR:?set in config.local.sh}"
: "${EXPLORER_HEALTH_URL:?set in config.local.sh}"

echo "==> Building explorer/ (CHAIN_CONFIG=$CHAIN_CONFIG)"
IMAGE_TAG="lcv-explorer-deploy:$(date +%s)"
docker build --build-arg "CHAIN_CONFIG=$CHAIN_CONFIG" -t "$IMAGE_TAG" "$EXPLORER_DIR"

echo "==> Extracting static output"
CONTAINER_ID="$(docker create "$IMAGE_TAG")"
TMP_DIST="$(mktemp -d)"
docker cp "$CONTAINER_ID:/usr/share/nginx/html/." "$TMP_DIST/"
docker rm "$CONTAINER_ID" >/dev/null
docker rmi "$IMAGE_TAG" >/dev/null

echo "==> Syncing to $VPS_USER@$VPS_HOST:$REMOTE_EXPLORER_DIR/"
rsync -az --delete \
  -e "ssh -i $SSH_KEY -o BatchMode=yes" \
  "$TMP_DIST/" \
  "$VPS_USER@$VPS_HOST:$REMOTE_EXPLORER_DIR/"
rm -rf "$TMP_DIST"

echo "==> Verifying $EXPLORER_HEALTH_URL"
code="$(curl -s -o /dev/null -w '%{http_code}' "$EXPLORER_HEALTH_URL")"
if [ "$code" != "200" ]; then
  echo "FAIL: $EXPLORER_HEALTH_URL returned HTTP $code"
  exit 1
fi

echo "==> Deployed successfully: $EXPLORER_HEALTH_URL (HTTP $code)"
