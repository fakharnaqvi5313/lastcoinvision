#!/usr/bin/env bash
# Lightweight liveness/signing-rate check for the lcv-testnet-1 validator —
# RELEASE_CHECKLIST.md §2's "monitoring wired up for missed blocks / signing
# rate" item. Not a full alerting stack (no external paging channel exists
# yet for this project); logs a clear WARNING line to stdout (captured by
# systemd/journald via the lcvd-monitor.service unit) whenever the
# validator's missed-block count approaches x/slashing's jail threshold, so
# `journalctl -u lcvd-monitor.service | grep WARNING` is a real, if manual,
# early-warning check until a real alerting channel exists.
set -euo pipefail

LCVD="/home/deploy/lcv-testnet-1/lcvd"
HOME_DIR="/home/deploy/lcv-testnet-1/data"
NODE="tcp://127.0.0.1:26657"

# 50 of the 100-block signed_blocks_window (see `lcvd query slashing
# params`) is when x/slashing actually jails the validator — warn well
# before that, not at the jailing threshold itself.
MISSED_BLOCKS_WARN_THRESHOLD="${MISSED_BLOCKS_WARN_THRESHOLD:-20}"

CONS_ADDR="$("$LCVD" comet show-address --home "$HOME_DIR")"
CONS_PUBKEY="$("$LCVD" comet show-validator --home "$HOME_DIR" | python3 -c "import json,sys; print(json.load(sys.stdin)['key'])")"

# query staking validators lists every validator on the network, not just
# this one -- with a single bootstrap validator, taking the first entry
# happened to be correct, but that broke silently the moment a second
# validator was onboarded (2026-08-05). Matched explicitly by this node's
# own consensus pubkey instead, so this script stays correct regardless of
# how many other validators exist.
STATUS_JSON="$("$LCVD" query staking validators --node "$NODE" --output json)"
VALIDATOR_STATUS="$(echo "$STATUS_JSON" | CONS_PUBKEY="$CONS_PUBKEY" python3 -c "
import json,os,sys
d = json.load(sys.stdin)
target = os.environ['CONS_PUBKEY']
for v in d.get('validators', []):
    # 'query staking validators --output json' uses the legacy
    # type/value field names for the pubkey Any, not @type/key like
    # 'comet show-validator' -- these are the same key, just serialized
    # differently by two different code paths in the same binary.
    if v.get('consensus_pubkey', {}).get('value') == target:
        print(v.get('status'))
        break
else:
    print('NOT_FOUND')
")"

SIGNING_JSON="$("$LCVD" query slashing signing-info "$CONS_ADDR" --node "$NODE" --output json)"
MISSED_BLOCKS="$(echo "$SIGNING_JSON" | python3 -c "
import json,sys
d = json.load(sys.stdin)
print(d.get('val_signing_info', {}).get('missed_blocks_counter', '0'))
")"
JAILED_UNTIL="$(echo "$SIGNING_JSON" | python3 -c "
import json,sys
d = json.load(sys.stdin)
print(d.get('val_signing_info', {}).get('jailed_until', ''))
")"

TIMESTAMP="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

echo "[$TIMESTAMP] status=$VALIDATOR_STATUS missed_blocks=$MISSED_BLOCKS jailed_until=$JAILED_UNTIL"

if [ "$VALIDATOR_STATUS" != "BOND_STATUS_BONDED" ]; then
  echo "[$TIMESTAMP] WARNING: validator status is $VALIDATOR_STATUS, not BOND_STATUS_BONDED"
fi

if [ "$MISSED_BLOCKS" -ge "$MISSED_BLOCKS_WARN_THRESHOLD" ]; then
  echo "[$TIMESTAMP] WARNING: missed_blocks=$MISSED_BLOCKS >= threshold $MISSED_BLOCKS_WARN_THRESHOLD (jailing happens at 50/100)"
fi
