# Copy to allocation-addresses.local.sh (gitignored) and source it before
# running init-testnet-genesis.sh, e.g.:
#
#   source scripts/genesis/allocation-addresses.local.sh
#   ./scripts/genesis/init-testnet-genesis.sh
#
# These are the real, custody-controlled addresses for each tokenomics
# bucket (see RELEASE_CHECKLIST.md's allocation-table item and DECISIONS.md
# for how/when each was generated) — not something init-testnet-genesis.sh
# should keep auto-generating placeholders for. Addresses themselves aren't
# secret; this file is gitignored to match the repo's existing
# config.local.sh/config.example.sh pattern, not because these values need
# hiding.

export ECOSYSTEM_ADDR="lcv1..."
export TREASURY_ADDR="lcv1..."
export TEAM_ADDR="lcv1..."
export INVESTORS_ADDR="lcv1..."
export PUBLIC_ADDR="lcv1..."
export COMMUNITY_ADDR="lcv1..."
