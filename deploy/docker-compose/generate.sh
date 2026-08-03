#!/usr/bin/env bash
# Generates a 4-validator devnet genesis for the docker-compose stack, then
# rewrites each validator's config for container-to-container networking:
# scripts/genesis/init-genesis.sh (reused as-is, zero duplication) assumes
# one host running 4 processes on 127.0.0.1 with a different port per
# validator; docker-compose runs 4 separate containers instead, each with
# its own loopback, so every service binds the *same* default port and
# reaches its peers by service name (lcvd0..lcvd3) rather than by port
# offset. See README.md for the full rationale.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
COMPOSE_DIR="$ROOT_DIR/deploy/docker-compose"
OUTPUT_DIR="$COMPOSE_DIR/.devnet"
NUM_VALIDATORS=4
LCVD="$ROOT_DIR/chain/build/lcvd"

if [ ! -x "$LCVD" ]; then
  echo "==> No build/lcvd found, building it"
  (cd "$ROOT_DIR" && make build)
fi

sed_inplace() {
  if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' "$1" "$2"
  else
    sed -i "$1" "$2"
  fi
}

echo "==> Generating $NUM_VALIDATORS-validator genesis under $OUTPUT_DIR"
OUTPUT_DIR="$OUTPUT_DIR" NUM_VALIDATORS="$NUM_VALIDATORS" "$ROOT_DIR/scripts/genesis/init-genesis.sh"

echo "==> Collecting node IDs"
NODE_IDS=()
for i in $(seq 0 $((NUM_VALIDATORS - 1))); do
  NODE_IDS+=("$("$LCVD" comet show-node-id --home "$OUTPUT_DIR/validator$i")")
done

PEERS=""
for i in $(seq 0 $((NUM_VALIDATORS - 1))); do
  [ -n "$PEERS" ] && PEERS="$PEERS,"
  PEERS="${PEERS}${NODE_IDS[$i]}@lcvd$i:26656"
done

echo "==> Rewriting config.toml / app.toml for container-to-container networking"
for i in $(seq 0 $((NUM_VALIDATORS - 1))); do
  cfg="$OUTPUT_DIR/validator$i/config/config.toml"
  app="$OUTPUT_DIR/validator$i/config/app.toml"

  # Every peer except this validator's own node ID (a node isn't its own persistent peer).
  self_id="${NODE_IDS[$i]}"
  own_peers="$(echo "$PEERS" | tr ',' '\n' | grep -v "^${self_id}@" | paste -sd, -)"
  sed_inplace "s#^persistent_peers = .*#persistent_peers = \"$own_peers\"#" "$cfg"

  # P2P must be rewritten before RPC: RPC's line becomes "tcp://0.0.0.0:..."
  # too once patched, and would otherwise also match P2P's own pattern.
  sed_inplace 's#^laddr = "tcp://0.0.0.0:[0-9]*"#laddr = "tcp://0.0.0.0:26656"#' "$cfg"
  sed_inplace 's#^laddr = "tcp://127.0.0.1:[0-9]*"#laddr = "tcp://0.0.0.0:26657"#' "$cfg"

  sed_inplace 's#^address = "tcp://localhost:[0-9]*"#address = "tcp://0.0.0.0:1317"#' "$app"
  sed_inplace 's#^address = "localhost:[0-9]*"#address = "0.0.0.0:9090"#' "$app"
  sed_inplace 's#^address = "127\.0\.0\.1:[0-9]*"#address = "0.0.0.0:8545"#' "$app"
  sed_inplace 's#^ws-address = "127\.0\.0\.1:[0-9]*"#ws-address = "0.0.0.0:8546"#' "$app"
done

echo "==> Extracting a funded account's mnemonic for the faucet"
FAUCET_MNEMONIC="$(python3 -c "
import json
print(json.load(open('$OUTPUT_DIR/.allocation-keys/public.key.json'))['mnemonic'])
")"

cat > "$COMPOSE_DIR/.env" <<EOF
FAUCET_MNEMONIC=$FAUCET_MNEMONIC
EOF

echo "==> Wrote $COMPOSE_DIR/.env (faucet funded from the 'public' allocation bucket)"
echo "==> Ready — run 'docker compose up' from deploy/docker-compose/"
