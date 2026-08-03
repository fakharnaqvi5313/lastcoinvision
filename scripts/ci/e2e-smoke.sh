#!/usr/bin/env bash
# Minimal end-to-end smoke test against a running `make devnet`: waits for
# the chain to produce blocks, sends a real bank tx between two validator
# accounts, confirms it actually landed (not just passed CheckTx), and
# queries one custom module to confirm it's wired into the running binary.
# Used by the nightly e2e workflow; exits non-zero on any failure.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
CHAIN_DIR="$ROOT_DIR/chain"
OUTPUT_DIR="${OUTPUT_DIR:-$HOME/.lcv-devnet}"
CHAIN_ID="${CHAIN_ID:-lcv-devnet}"
LCVD="$CHAIN_DIR/build/lcvd"
RPC="http://localhost:26657"
API="http://localhost:1317"
HOME0="$OUTPUT_DIR/validator0"
HOME1="$OUTPUT_DIR/validator1"

echo "==> Waiting for $RPC to accept connections"
for i in $(seq 1 60); do
  if curl -sf "$RPC/status" > /dev/null 2>&1; then
    break
  fi
  if [ "$i" -eq 60 ]; then
    echo "FAIL: $RPC never came up after 60s"
    exit 1
  fi
  sleep 1
done

echo "==> Waiting for block height to advance past 3"
for i in $(seq 1 60); do
  height="$(curl -sf "$RPC/status" | grep -o '"latest_block_height":"[0-9]*"' | grep -o '[0-9]*' || echo 0)"
  if [ "${height:-0}" -ge 3 ]; then
    echo "    at height $height"
    break
  fi
  if [ "$i" -eq 60 ]; then
    echo "FAIL: chain stuck below height 3 after 60s (last seen: $height)"
    exit 1
  fi
  sleep 1
done

echo "==> Querying validator0/validator1 addresses"
ADDR0="$("$LCVD" keys show validator0 -a --keyring-backend test --home "$HOME0")"
ADDR1="$("$LCVD" keys show validator1 -a --keyring-backend test --home "$HOME1")"
echo "    validator0: $ADDR0"
echo "    validator1: $ADDR1"

echo "==> Querying validator0's bank balance"
curl -sf "$API/cosmos/bank/v1beta1/balances/$ADDR0" | tee /tmp/e2e-balance.json
grep -q "ulcv" /tmp/e2e-balance.json || { echo "FAIL: validator0 has no ulcv balance"; exit 1; }

echo "==> Sending 1000ulcv from validator0 to validator1"
TX_OUT="$("$LCVD" tx bank send "$ADDR0" "$ADDR1" 1000ulcv \
  --home "$HOME0" --keyring-backend test --chain-id "$CHAIN_ID" \
  --node "$RPC" --fees 500ulcv -y --output json)"
TXHASH="$(echo "$TX_OUT" | grep -o '"txhash":"[^"]*"' | cut -d'"' -f4)"
echo "    txhash: $TXHASH"
if [ -z "$TXHASH" ]; then
  echo "FAIL: no txhash returned from bank send"
  echo "$TX_OUT"
  exit 1
fi

echo "==> Waiting for tx to be included and confirming it actually succeeded"
CODE=""
for i in $(seq 1 20); do
  sleep 1
  QUERY_OUT="$("$LCVD" query tx "$TXHASH" --node "$RPC" --output json 2>/dev/null || true)"
  [ -z "$QUERY_OUT" ] && continue
  CODE="$(echo "$QUERY_OUT" | grep -o '"code":[0-9]*' | head -1 | grep -o '[0-9]*')"
  break
done
if [ "${CODE:-}" != "0" ]; then
  echo "FAIL: tx $TXHASH did not land with code 0 (got: ${CODE:-<not found>})"
  echo "$QUERY_OUT"
  exit 1
fi
echo "    confirmed: code 0"

echo "==> Querying x/emission params (custom module smoke check)"
curl -sf "$API/lastcoinvision/lcv/emission/params" | tee /tmp/e2e-emission-params.json
grep -q "total_reward_pool" /tmp/e2e-emission-params.json || { echo "FAIL: x/emission params missing expected field"; exit 1; }

echo "==> e2e smoke test passed"
