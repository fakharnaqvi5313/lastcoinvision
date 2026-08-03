#!/usr/bin/env bash
# Full, self-contained software-upgrade governance walkthrough: starts a
# fresh 4-validator devnet on a "pre-v2" binary (the v2 upgrade handler's
# registration disabled — see build-pre-upgrade-binary.sh), submits a real
# MsgSoftwareUpgrade proposal naming the "v2" plan registered in
# chain/app/upgrades/v2, votes yes from every validator, waits for it to
# pass, confirms the chain halts EXACTLY at the upgrade height with the
# expected "UPGRADE ... NEEDED" error (the correct, designed behavior for
# a binary that doesn't yet know about the upgrade), then swaps in the
# real (v2-aware) `chain/build/lcvd` binary and confirms the chain resumes
# and keeps producing blocks past that height.
#
# Why it needs two binaries, not one: cosmos-sdk's x/upgrade module
# refuses to let a binary that *already* has a plan's handler registered
# schedule that same plan again ("BINARY UPDATED BEFORE TRIGGER" — this
# was discovered the hard way running an earlier version of this script
# against a normal `make devnet`; see the 2026-08-03 DECISIONS.md entry).
# A real chain never hits this: the upgrade is only ever scheduled on the
# binary version that predates it, exactly like this script now does.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
OUTPUT_DIR="${OUTPUT_DIR:-$HOME/.lcv-devnet}"
CHAIN_ID="${CHAIN_ID:-lcv-devnet}"
NUM_VALIDATORS="${NUM_VALIDATORS:-4}"
RPC="http://localhost:26657"
UPGRADE_HEIGHT_OFFSET="${UPGRADE_HEIGHT_OFFSET:-40}" # blocks from now — generous for a ~1s block time

PRE_V2_BIN="$ROOT_DIR/chain/build/lcvd-pre-v2"
V2_BIN="$ROOT_DIR/chain/build/lcvd"
PID_DIR="$OUTPUT_DIR/pids"

echo "==> Building pre-v2 binary (v2 handler not yet registered)"
"$ROOT_DIR/scripts/upgrade/build-pre-upgrade-binary.sh" "$PRE_V2_BIN"

echo "==> Building the real (v2-aware) lcvd, if needed"
if [ ! -x "$V2_BIN" ]; then
  (cd "$ROOT_DIR" && make build)
fi

echo "==> Generating a fresh $NUM_VALIDATORS-validator devnet"
OUTPUT_DIR="$OUTPUT_DIR" NUM_VALIDATORS="$NUM_VALIDATORS" "$ROOT_DIR/scripts/genesis/init-genesis.sh" > /dev/null

echo "==> Starting validators on the pre-v2 binary"
mkdir -p "$PID_DIR"
for i in $(seq 0 $((NUM_VALIDATORS - 1))); do
  "$PRE_V2_BIN" start --home "$OUTPUT_DIR/validator$i" > "$OUTPUT_DIR/validator$i/node.log" 2>&1 &
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

LCVD="$PRE_V2_BIN"
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
        "name": "v2",
        "height": "$upgrade_height",
        "info": "Sample software upgrade for the Phase H governance walkthrough — see chain/app/upgrades/v2."
      }
    }
  ],
  "metadata": "ipfs://sample",
  "deposit": "10000000ulcv",
  "title": "Sample v2 software upgrade",
  "summary": "Demonstrates the full MsgSoftwareUpgrade governance flow against a live devnet."
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

echo "==> Waiting for the pre-v2 binary to halt at height $upgrade_height (expected!)"
halted=""
for i in $(seq 1 90); do
  if grep -q "UPGRADE \"v2\" NEEDED" "$OUTPUT_DIR/validator0/node.log" 2>/dev/null; then
    halted=1
    break
  fi
  sleep 1
done
if [ -z "$halted" ]; then
  echo "FAIL: pre-v2 binary never hit the expected 'UPGRADE \"v2\" NEEDED' halt"
  exit 1
fi
echo "    confirmed: pre-v2 binary halted with the expected upgrade-needed error"

echo "==> Swapping in the real (v2-aware) lcvd and restarting"
for i in $(seq 0 $((NUM_VALIDATORS - 1))); do
  kill "$(cat "$PID_DIR/validator$i.pid")" 2>/dev/null || true
done
sleep 2
for i in $(seq 0 $((NUM_VALIDATORS - 1))); do
  "$V2_BIN" start --home "$OUTPUT_DIR/validator$i" >> "$OUTPUT_DIR/validator$i/node.log" 2>&1 &
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
