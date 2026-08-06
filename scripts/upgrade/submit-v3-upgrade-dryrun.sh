#!/usr/bin/env bash
# Local, disposable-devnet dry run of the v3 upgrade -- the exact same
# governance-vote-halt-swap sequence intended for real on lcv-testnet-1
# (see MAINNET_LAUNCH_RUNBOOK.md's "rehearse a real governance-triggered
# upgrade" gate), run here first against throwaway infrastructure so any
# bug in the v3 handler or this sequence itself surfaces before it ever
# touches production validators. Adapted from submit-upgrade-proposal.sh
# (the v2 walkthrough) -- see that script's comments for why two binaries
# are needed instead of one.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
OUTPUT_DIR="${OUTPUT_DIR:-$HOME/.lcv-devnet}"
CHAIN_ID="${CHAIN_ID:-lcv-devnet}"
NUM_VALIDATORS="${NUM_VALIDATORS:-4}"
RPC="http://localhost:26657"
UPGRADE_HEIGHT_OFFSET="${UPGRADE_HEIGHT_OFFSET:-40}"

PRE_V3_BIN="$ROOT_DIR/chain/build/lcvd-pre-v3"
V3_BIN="$ROOT_DIR/chain/build/lcvd"
PID_DIR="$OUTPUT_DIR/pids"

echo "==> Building pre-v3 binary (v3 handler not yet registered, v2 still is)"
"$ROOT_DIR/scripts/upgrade/build-pre-v3-binary.sh" "$PRE_V3_BIN"

echo "==> Building the real (v3-aware) lcvd, if needed"
if [ ! -x "$V3_BIN" ]; then
  (cd "$ROOT_DIR" && make build)
fi

echo "==> Generating a fresh $NUM_VALIDATORS-validator devnet"
OUTPUT_DIR="$OUTPUT_DIR" NUM_VALIDATORS="$NUM_VALIDATORS" "$ROOT_DIR/scripts/genesis/init-genesis.sh" > /dev/null

echo "==> Starting validators on the pre-v3 binary"
mkdir -p "$PID_DIR"
for i in $(seq 0 $((NUM_VALIDATORS - 1))); do
  "$PRE_V3_BIN" start --home "$OUTPUT_DIR/validator$i" > "$OUTPUT_DIR/validator$i/node.log" 2>&1 &
  echo $! > "$PID_DIR/validator$i.pid"
done
sleep 5

current_height() {
  curl -sf "$RPC/status" 2>/dev/null | grep -o '"latest_block_height":"[0-9]*"' | grep -o '[0-9]*' || echo ""
}

for i in $(seq 1 30); do
  h="$(current_height)"
  [ -n "$h" ] && [ "$h" -gt 0 ] && break
  sleep 1
done

height_now="$(current_height)"
upgrade_height=$((height_now + UPGRADE_HEIGHT_OFFSET))
echo "==> Current height: $height_now — targeting upgrade at height $upgrade_height"

LCVD="$PRE_V3_BIN"
GOV_ADDR="$("$LCVD" query auth module-account gov --node "$RPC" | grep address | awk '{print $2}')"
echo "==> Gov module account: $GOV_ADDR"

PROPOSAL_FILE="$(mktemp)"
cat > "$PROPOSAL_FILE" <<EOF
{
  "messages": [
    {
      "@type": "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
      "authority": "$GOV_ADDR",
      "plan": {
        "name": "v3",
        "height": "$upgrade_height",
        "info": "Dry run of the v3 upgrade rehearsal (see chain/app/upgrades/v3), ahead of a real submission to lcv-testnet-1."
      }
    }
  ],
  "metadata": "ipfs://sample",
  "deposit": "10000000ulcv",
  "title": "v3 upgrade dry run",
  "summary": "Local disposable-devnet rehearsal of the v3 governance-triggered upgrade before it touches lcv-testnet-1."
}
EOF

echo "==> Submitting proposal"
SUBMIT_OUT="$("$LCVD" tx gov submit-proposal "$PROPOSAL_FILE" \
  --from validator0 --home "$OUTPUT_DIR/validator0" --keyring-backend test \
  --chain-id "$CHAIN_ID" --node "$RPC" --gas 300000 --fees 10000ulcv -y --output json)"
TXHASH="$(echo "$SUBMIT_OUT" | grep -o '"txhash":"[^"]*"' | cut -d'"' -f4)"
echo "    submit txhash: $TXHASH"
sleep 3

PROPOSAL_ID="$("$LCVD" query tx "$TXHASH" --node "$RPC" --output json | python3 -c "
import json, sys
tx = json.load(sys.stdin)
for event in tx.get('events', []):
    for attr in event.get('attributes', []):
        if attr.get('key') == 'proposal_id':
            print(attr['value'])
            sys.exit(0)
")"
if [ -z "$PROPOSAL_ID" ]; then
  echo "FAIL: could not determine proposal ID from tx $TXHASH"
  "$LCVD" query tx "$TXHASH" --node "$RPC"
  exit 1
fi
echo "    proposal ID: $PROPOSAL_ID"

echo "==> Voting yes from all $NUM_VALIDATORS validators"
for i in $(seq 0 $((NUM_VALIDATORS - 1))); do
  "$LCVD" tx gov vote "$PROPOSAL_ID" yes \
    --from "validator$i" --home "$OUTPUT_DIR/validator$i" --keyring-backend test \
    --chain-id "$CHAIN_ID" --node "$RPC" --gas 300000 --fees 10000ulcv -y --output json > /dev/null
  echo "    validator$i voted yes"
  sleep 1
done

proposal_status() {
  "$LCVD" query gov proposal "$PROPOSAL_ID" --node "$RPC" --output json | python3 -c "
import json, sys
d = json.load(sys.stdin)
print(d.get('proposal', d).get('status', ''))
"
}

echo "==> Waiting for proposal to pass"
for i in $(seq 1 30); do
  sleep 1
  status="$(proposal_status)"
  [ "$status" = "PROPOSAL_STATUS_PASSED" ] && break
  if [ "$i" = 30 ]; then
    echo "FAIL: proposal $PROPOSAL_ID did not pass (status: $status)"
    exit 1
  fi
done
echo "    proposal $PROPOSAL_ID: $status"

echo "==> Waiting for the pre-v3 binary to halt at height $upgrade_height (expected!)"
halted=""
for i in $(seq 1 90); do
  if grep -q "UPGRADE \"v3\" NEEDED" "$OUTPUT_DIR/validator0/node.log" 2>/dev/null; then
    halted=1
    break
  fi
  sleep 1
done
if [ -z "$halted" ]; then
  echo "FAIL: pre-v3 binary never hit the expected 'UPGRADE \"v3\" NEEDED' halt"
  exit 1
fi
echo "    confirmed: pre-v3 binary halted with the expected upgrade-needed error"

echo "==> Swapping in the real (v3-aware) lcvd and restarting"
for i in $(seq 0 $((NUM_VALIDATORS - 1))); do
  kill "$(cat "$PID_DIR/validator$i.pid")" 2>/dev/null || true
done
sleep 2
for i in $(seq 0 $((NUM_VALIDATORS - 1))); do
  "$V3_BIN" start --home "$OUTPUT_DIR/validator$i" >> "$OUTPUT_DIR/validator$i/node.log" 2>&1 &
  echo $! > "$PID_DIR/validator$i.pid"
done
sleep 5

echo "==> Confirming the chain resumed and kept producing blocks past height $upgrade_height"
h_after="$(current_height)"
if [ -z "$h_after" ] || [ "$h_after" -le "$upgrade_height" ]; then
  echo "FAIL: chain did not resume past the upgrade height (height: ${h_after:-<none>})"
  exit 1
fi
echo "    height now: $h_after — upgrade applied cleanly after the binary swap"

echo "==> Stopping devnet validators"
for i in $(seq 0 $((NUM_VALIDATORS - 1))); do
  kill "$(cat "$PID_DIR/validator$i.pid")" 2>/dev/null || true
done
