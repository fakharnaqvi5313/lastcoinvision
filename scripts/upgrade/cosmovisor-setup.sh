#!/usr/bin/env bash
# Installs cosmovisor and lays out $DAEMON_HOME/cosmovisor/ so lcvd can be
# supervised and hot-swapped at a governance-approved upgrade height,
# instead of an operator manually stopping/swapping/restarting the binary
# (see docs/docs/validator-handbook.md's "Software upgrades" section,
# which this replaces the "manual" half of).
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DAEMON_HOME="${1:-$HOME/.lcv-devnet/validator0}"
LCVD_BIN="$ROOT_DIR/chain/build/lcvd"

if [ ! -x "$LCVD_BIN" ]; then
  echo "==> No build/lcvd found, building it"
  (cd "$ROOT_DIR" && make build)
fi

if ! command -v cosmovisor >/dev/null 2>&1; then
  echo "==> Installing cosmovisor"
  go install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@v1.7.1
fi

echo "==> Laying out cosmovisor directories under $DAEMON_HOME"
mkdir -p "$DAEMON_HOME/cosmovisor/genesis/bin"
cp "$LCVD_BIN" "$DAEMON_HOME/cosmovisor/genesis/bin/lcvd"

cat <<EOF

==> Done. Start the node under cosmovisor with:

  export DAEMON_NAME=lcvd
  export DAEMON_HOME=$DAEMON_HOME
  export DAEMON_ALLOW_DOWNLOAD_BINARIES=false
  export DAEMON_RESTART_AFTER_UPGRADE=true
  cosmovisor run start --home "\$DAEMON_HOME"

To pre-stage a new binary for an upgrade named "v2" ahead of time (cosmovisor
switches to it automatically the instant governance's plan takes effect —
see submit-upgrade-proposal.sh):

  mkdir -p $DAEMON_HOME/cosmovisor/upgrades/v2/bin
  cp <new-lcvd-binary> $DAEMON_HOME/cosmovisor/upgrades/v2/bin/lcvd
EOF
