#!/usr/bin/env bash
# Builds a fresh N-validator LastcoinVision devnet genesis: validator
# homes + keys + gentxs (via `lcvd multi-node`), then patches the resulting
# genesis.json with the full token allocation table (ecosystem, treasury,
# team/investor vesting, staking-rewards pool, public, community) and pins
# x/mint to zero inflation (via the `gengenesis` tool). Idempotent: re-running
# wipes and rebuilds $OUTPUT_DIR and $KEYS_HOME from scratch.
set -euo pipefail

CHAIN_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../chain" && pwd)"
NUM_VALIDATORS="${NUM_VALIDATORS:-4}"
CHAIN_ID="${CHAIN_ID:-lcv-devnet}"
DENOM="ulcv"
KEYRING_BACKEND="test"
OUTPUT_DIR="${OUTPUT_DIR:-$HOME/.lcv-devnet}"
KEYS_HOME="${KEYS_HOME:-$OUTPUT_DIR/.allocation-keys}"
VALIDATOR_STAKE="${VALIDATOR_STAKE:-400000000}" # 400 LCV self-bond; multi-node funds each
                                                 # validator account with a hardcoded 500 LCV
                                                 # (see chain/cmd/lcvd/cmd/testnet_multi_node.go),
                                                 # so this must stay comfortably under that with
                                                 # room left over for gentx gas.
BASE_PORT="${BASE_PORT:-26650}"                    # node0 uses 26650-26659, node1 26660-26669, etc.

BIN_DIR="$(mktemp -d)"
trap 'rm -rf "$BIN_DIR"' EXIT

echo "==> Building lcvd and gengenesis"
(cd "$CHAIN_DIR" && go build -o "$BIN_DIR/lcvd" ./cmd/lcvd && go build -o "$BIN_DIR/gengenesis" ./cmd/gengenesis)
LCVD="$BIN_DIR/lcvd"
GENGENESIS="$BIN_DIR/gengenesis"

echo "==> Wiping $OUTPUT_DIR"
rm -rf "$OUTPUT_DIR"
mkdir -p "$KEYS_HOME"

echo "==> Creating allocation-bucket keys in $KEYS_HOME"
ALLOCATION_NAMES=(ecosystem treasury team investors public community)
for name in "${ALLOCATION_NAMES[@]}"; do
  "$LCVD" keys add "$name" --keyring-backend "$KEYRING_BACKEND" --home "$KEYS_HOME" --output json \
    > "$KEYS_HOME/${name}.key.json" 2>&1
done

addr_of() { "$LCVD" keys show "$1" -a --keyring-backend "$KEYRING_BACKEND" --home "$KEYS_HOME"; }
ECOSYSTEM_ADDR="$(addr_of ecosystem)"
TREASURY_ADDR="$(addr_of treasury)"
TEAM_ADDR="$(addr_of team)"
INVESTORS_ADDR="$(addr_of investors)"
PUBLIC_ADDR="$(addr_of public)"
COMMUNITY_ADDR="$(addr_of community)"

# Build a comma-separated stake list so every validator self-bonds the same
# amount (multi-node requires one entry per validator).
STAKE_LIST="$VALIDATOR_STAKE"
for i in $(seq 2 "$NUM_VALIDATORS"); do STAKE_LIST="$STAKE_LIST,$VALIDATOR_STAKE"; done

# Ports: multi-node's --list-ports takes one P2P-ish base port per node and
# derives the rest; space nodes far enough apart to avoid collisions across
# all of P2P/RPC/gRPC/API/pprof.
PORT_LIST=""
for i in $(seq 0 $((NUM_VALIDATORS - 1))); do
  port=$((BASE_PORT + i * 100))
  [ -z "$PORT_LIST" ] && PORT_LIST="$port" || PORT_LIST="$PORT_LIST,$port"
done

echo "==> Generating $NUM_VALIDATORS validator homes via 'lcvd multi-node'"
"$LCVD" multi-node \
  --v "$NUM_VALIDATORS" \
  --chain-id "$CHAIN_ID" \
  --output-dir "$OUTPUT_DIR" \
  --keyring-backend "$KEYRING_BACKEND" \
  --node-dir-prefix validator \
  --validators-stake-amount "$STAKE_LIST" \
  --list-ports "$PORT_LIST" \
  --minimum-gas-prices "0${DENOM}" \
  > "$OUTPUT_DIR.multi-node.log" 2>&1

CANONICAL_GENESIS="$OUTPUT_DIR/validator0/config/genesis.json"

echo "==> Patching $CANONICAL_GENESIS with the token allocation table"
"$GENGENESIS" \
  --genesis "$CANONICAL_GENESIS" \
  --ecosystem-addr "$ECOSYSTEM_ADDR" \
  --treasury-addr "$TREASURY_ADDR" \
  --team-addr "$TEAM_ADDR" \
  --investors-addr "$INVESTORS_ADDR" \
  --public-addr "$PUBLIC_ADDR" \
  --community-addr "$COMMUNITY_ADDR"

echo "==> Distributing patched genesis to all $NUM_VALIDATORS validator homes"
for i in $(seq 0 $((NUM_VALIDATORS - 1))); do
  dst="$OUTPUT_DIR/validator$i/config/genesis.json"
  [ "$dst" = "$CANONICAL_GENESIS" ] && continue
  cp "$CANONICAL_GENESIS" "$dst"
done

echo "==> Validating genesis on every validator home"
for i in $(seq 0 $((NUM_VALIDATORS - 1))); do
  "$LCVD" genesis validate-genesis --home "$OUTPUT_DIR/validator$i"
done

cat <<EOF

==> Done. $NUM_VALIDATORS validator homes ready under $OUTPUT_DIR (validator0 .. validator$((NUM_VALIDATORS - 1))).

Allocation addresses (keys + mnemonics saved under $KEYS_HOME):
  ecosystem_grants (25%):    $ECOSYSTEM_ADDR
  foundation_treasury (20%): $TREASURY_ADDR
  team (15%, 4y/1y cliff):   $TEAM_ADDR
  investors (15%, 3y/1y cliff): $INVESTORS_ADDR
  public (8%):               $PUBLIC_ADDR
  community_airdrop (5%):    $COMMUNITY_ADDR
  staking_rewards (12%):     module account, no keys (see modules/emission/SPEC.md)

Run 'make devnet' from the repo root to start all $NUM_VALIDATORS nodes.
EOF
