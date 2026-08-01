#!/usr/bin/env bash
# Starts every validator home under $OUTPUT_DIR as a background lcvd
# process, logging each to its own file and recording its PID so
# stop.sh can find it later. Runs scripts/genesis/init-genesis.sh first if
# $OUTPUT_DIR doesn't exist yet.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
CHAIN_DIR="$ROOT_DIR/chain"
NUM_VALIDATORS="${NUM_VALIDATORS:-4}"
OUTPUT_DIR="${OUTPUT_DIR:-$HOME/.lcv-devnet}"
PID_DIR="$OUTPUT_DIR/pids"

if [ ! -d "$OUTPUT_DIR/validator0" ]; then
  echo "==> No devnet found at $OUTPUT_DIR, generating one first"
  NUM_VALIDATORS="$NUM_VALIDATORS" OUTPUT_DIR="$OUTPUT_DIR" "$ROOT_DIR/scripts/genesis/init-genesis.sh"
fi

LCVD_BIN="$CHAIN_DIR/build/lcvd"
if [ ! -x "$LCVD_BIN" ]; then
  echo "==> No build/lcvd found, building it"
  mkdir -p "$CHAIN_DIR/build"
  (cd "$CHAIN_DIR" && go build -o build/lcvd ./cmd/lcvd)
fi

mkdir -p "$PID_DIR"

for i in $(seq 0 $((NUM_VALIDATORS - 1))); do
  dir="$OUTPUT_DIR/validator$i"
  pidfile="$PID_DIR/validator$i.pid"
  if [ -f "$pidfile" ] && kill -0 "$(cat "$pidfile")" 2>/dev/null; then
    echo "validator$i already running (pid $(cat "$pidfile"))"
    continue
  fi
  "$LCVD_BIN" start --home "$dir" > "$dir/node.log" 2>&1 &
  echo $! > "$pidfile"
  echo "validator$i started (pid $!), log: $dir/node.log"
done

cat <<EOF

==> Devnet running. Endpoints:
EOF
for i in $(seq 0 $((NUM_VALIDATORS - 1))); do
  rpc_port=$((26657 + i * 10))
  api_port=$((1317 + i * 10))
  grpc_port=$((9090 + i * 10))
  echo "  validator$i: RPC http://localhost:${rpc_port}  API http://localhost:${api_port}  gRPC localhost:${grpc_port}"
done
echo
echo "Example: $LCVD_BIN status --home $OUTPUT_DIR/validator0 | head -c 400"
echo "Stop with: make localnet-stop  (or scripts/devnet/stop.sh)"
