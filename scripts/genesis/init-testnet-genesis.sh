#!/usr/bin/env bash
# Builds genesis for a real (non-devnet) LastcoinVision network — lcv-testnet-1
# by default — as a single initial validator, matching RELEASE_CHECKLIST.md's
# "dry-run privately among the launch coordinator's own validators first"
# step. Unlike scripts/genesis/init-genesis.sh (which uses `lcvd multi-node`,
# a devnet-only tool that auto-funds each validator from outside the fixed
# 10B supply), this script uses the standard init/gentx/collect-gentxs flow
# and funds the validator's self-bond by carving it out of the public
# allocation via gengenesis's -validator-self-bond-* flags — auditable in the
# genesis.json itself, not conjured from nowhere.
#
# Production-vs-devnet differences from init-genesis.sh, deliberately:
#   - No --devnet-fast-gov: x/gov uses real (48h/60-day-class) voting/deposit
#     periods, not the 30s local-testing window.
#   - No timeout_commit override: keeps CometBFT's real default instead of
#     the devnet's 1s convenience value.
#   - No enabled-unsafe-cors: browser-based dApps should go through nginx's
#     own scoped CORS headers on the reverse proxy, not the node's own
#     wildcard-origin escape hatch.
#   - Vesting cliffs/periods were already real calendar durations in
#     gengenesis (oneYearSeconds, etc.) even for devnet — nothing to change
#     there.
#
# The six allocation-bucket addresses this script generates are PLACEHOLDERS.
# It prints a clear manifest at the end; every one of those addresses must be
# replaced with real, custody-controlled addresses (multisigs, vesting
# administrators, etc.) before any network built from this genesis is
# treated as carrying real value. Swapping them is exactly re-running this
# script with ECOSYSTEM_ADDR / TREASURY_ADDR / etc. set to the real
# addresses instead of letting it generate its own.
set -euo pipefail

CHAIN_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../chain" && pwd)"
CHAIN_ID="${CHAIN_ID:-lcv-testnet-1}"
DENOM="ulcv"
KEYRING_BACKEND="test"
MONIKER="${MONIKER:-lcv-testnet-validator-0}"
OUTPUT_DIR="${OUTPUT_DIR:-$HOME/.lcv-testnet-1}"
KEYS_HOME="${KEYS_HOME:-$OUTPUT_DIR/.allocation-keys}"
VALIDATOR_IP="${VALIDATOR_IP:?set VALIDATOR_IP to the public IP this validator advertises (embedded in its gentx)}"
# 1,000,000 LCV — a small, round slice of the 800,000,000 LCV (8%) public
# allocation, self-bonded to bootstrap consensus. Not a claim about what any
# future validator's stake should be; just enough for this node to be the
# sole initial validator.
SELF_BOND_AMOUNT="${SELF_BOND_AMOUNT:-1000000000000}"

BIN_DIR="$(mktemp -d)"
trap 'rm -rf "$BIN_DIR"' EXIT

echo "==> Building lcvd and gengenesis"
(cd "$CHAIN_DIR" && go build -o "$BIN_DIR/lcvd" ./cmd/lcvd && go build -o "$BIN_DIR/gengenesis" ./cmd/gengenesis)
LCVD="$BIN_DIR/lcvd"
GENGENESIS="$BIN_DIR/gengenesis"

echo "==> Wiping $OUTPUT_DIR"
rm -rf "$OUTPUT_DIR"
mkdir -p "$KEYS_HOME"

# Each of these six can be pre-set in the environment (a real,
# custody-controlled address) to skip placeholder generation for that
# bucket. The validator's own operator key is always freshly generated —
# it isn't a token-allocation bucket, it's this specific node's identity.
: "${ECOSYSTEM_ADDR:=}"
: "${TREASURY_ADDR:=}"
: "${TEAM_ADDR:=}"
: "${INVESTORS_ADDR:=}"
: "${PUBLIC_ADDR:=}"
: "${COMMUNITY_ADDR:=}"

addr_of() { "$LCVD" keys show "$1" -a --keyring-backend "$KEYRING_BACKEND" --home "$KEYS_HOME"; }
exitf() { printf "$@" >&2; exit 1; }

# Portable stand-in for an associative array (macOS ships bash 3.2, which
# predates `declare -A`): bucket names and their env-var names are kept as
# two parallel lists instead.
BUCKET_NAMES="ecosystem treasury team investors public community"
bucket_env_var() {
  case "$1" in
    ecosystem) echo ECOSYSTEM_ADDR ;;
    treasury) echo TREASURY_ADDR ;;
    team) echo TEAM_ADDR ;;
    investors) echo INVESTORS_ADDR ;;
    public) echo PUBLIC_ADDR ;;
    community) echo COMMUNITY_ADDR ;;
    *) exitf "unknown allocation bucket %s (bug in init-testnet-genesis.sh)\n" "$1" ;;
  esac
}

echo "==> Creating placeholder keys in $KEYS_HOME for any allocation bucket not pre-set in the environment"
for name in $BUCKET_NAMES; do
  env_var="$(bucket_env_var "$name")"
  if [ -n "${!env_var}" ]; then
    echo "    $name: using pre-set $env_var (${!env_var})"
    continue
  fi
  "$LCVD" keys add "$name" --keyring-backend "$KEYRING_BACKEND" --home "$KEYS_HOME" --output json \
    > "$KEYS_HOME/${name}.key.json" 2>&1
  printf -v "$env_var" '%s' "$(addr_of "$name")"
  echo "    $name: generated placeholder ${!env_var}"
done
"$LCVD" keys add validator0 --keyring-backend "$KEYRING_BACKEND" --home "$KEYS_HOME" --output json \
  > "$KEYS_HOME/validator0.key.json" 2>&1
VALIDATOR_ADDR="$(addr_of validator0)"

echo "==> lcvd init ($CHAIN_ID, moniker $MONIKER)"
"$LCVD" init "$MONIKER" --chain-id "$CHAIN_ID" --default-denom "$DENOM" --home "$OUTPUT_DIR" --overwrite \
  > "$OUTPUT_DIR.init.log" 2>&1

# Advertise this node's real reachable address so future peers can find it —
# left blank by `lcvd init`, and RPC/API/gRPC/EVM all default to loopback
# already (nothing else here needs to change for a public-facing node).
sed_inplace() { sed -i.bak "$1" "$2" && rm -f "$2.bak"; }
sed_inplace "s#^external_address = \"\"#external_address = \"${VALIDATOR_IP}:26656\"#" "$OUTPUT_DIR/config/config.toml"

CANONICAL_GENESIS="$OUTPUT_DIR/config/genesis.json"

echo "==> Patching $CANONICAL_GENESIS with the token allocation table + validator self-bond"
"$GENGENESIS" \
  --genesis "$CANONICAL_GENESIS" \
  --ecosystem-addr "$ECOSYSTEM_ADDR" \
  --treasury-addr "$TREASURY_ADDR" \
  --team-addr "$TEAM_ADDR" \
  --investors-addr "$INVESTORS_ADDR" \
  --public-addr "$PUBLIC_ADDR" \
  --community-addr "$COMMUNITY_ADDR" \
  --validator-self-bond-addr "$VALIDATOR_ADDR" \
  --validator-self-bond-amount "$SELF_BOND_AMOUNT"
  # No --devnet-fast-gov: real x/gov timing.

echo "==> Generating gentx for validator0 (self-bond: $SELF_BOND_AMOUNT$DENOM)"
"$LCVD" genesis gentx validator0 "${SELF_BOND_AMOUNT}${DENOM}" \
  --chain-id "$CHAIN_ID" \
  --moniker "$MONIKER" \
  --ip "$VALIDATOR_IP" \
  --keyring-backend "$KEYRING_BACKEND" \
  --keyring-dir "$KEYS_HOME" \
  --home "$OUTPUT_DIR"

echo "==> Collecting gentxs"
"$LCVD" genesis collect-gentxs --home "$OUTPUT_DIR"

echo "==> Validating genesis"
"$LCVD" genesis validate-genesis --home "$OUTPUT_DIR"

cat <<EOF

==> Done. Genesis + node home ready at $OUTPUT_DIR.

ANY ADDRESS BELOW MARKED "generated placeholder" ABOVE IS NOT REAL. Before
this network is treated as carrying real value, every one of those must be
replaced with a real, custody-controlled address — re-run this script with
ECOSYSTEM_ADDR=..., TREASURY_ADDR=..., etc. set in the environment to use
real addresses instead of generating new placeholders (see the per-bucket
log lines above for which ones still need replacing).

Allocation addresses (keys + mnemonics saved under $KEYS_HOME):
  ecosystem_grants (25%):       $ECOSYSTEM_ADDR
  foundation_treasury (20%):    $TREASURY_ADDR
  team (15%, 4y/1y cliff):      $TEAM_ADDR
  investors (15%, 3y/1y cliff): $INVESTORS_ADDR
  public (8%, minus self-bond): $PUBLIC_ADDR
  community_airdrop (5%):       $COMMUNITY_ADDR
  staking_rewards (12%):        module account, no keys (see modules/emission/SPEC.md)

Initial validator (self-bonded $SELF_BOND_AMOUNT$DENOM out of the public allocation):
  operator address: $VALIDATOR_ADDR
  moniker:           $MONIKER
  advertised P2P IP: $VALIDATOR_IP

Next: copy $OUTPUT_DIR to the validator machine (or start lcvd with
--home $OUTPUT_DIR directly here), then start lcvd. This genesis has NOT
had timeout_commit, CORS, or gov timing shortened for local convenience —
it's built with production values throughout.
EOF
