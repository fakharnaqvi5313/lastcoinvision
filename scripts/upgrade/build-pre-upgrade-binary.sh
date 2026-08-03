#!/usr/bin/env bash
# Builds a "pre-v2" lcvd binary — chain/app/app.go with the v2 upgrade
# handler's registration commented out — for the walkthrough in
# submit-upgrade-proposal.sh / swap-and-resume.sh to exercise against.
#
# Why this exists: every binary this repo normally builds already has the
# "v2" handler registered (chain/app/upgrades.go calls
# setupUpgradeHandlers() unconditionally — correct for a chain that has
# actually crossed that upgrade in its history). But that means running
# the walkthrough against a normal `make build` binary fails immediately:
# cosmos-sdk's x/upgrade module refuses to let a binary that ALREADY knows
# about a named upgrade schedule that same upgrade again — see the
# "BINARY UPDATED BEFORE TRIGGER" safety check in
# cosmossdk.io/x/upgrade/abci.go, and the 2026-08-03 DECISIONS.md entry
# this discovery is written up in. A real chain doesn't hit this: the
# upgrade only gets *scheduled* once, on the binary that predates it.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
APP_GO="$ROOT_DIR/chain/app/app.go"
OUT="${1:-$ROOT_DIR/chain/build/lcvd-pre-v2}"

BACKUP="$(mktemp)"
cp "$APP_GO" "$BACKUP"
restore() { cp "$BACKUP" "$APP_GO"; rm -f "$BACKUP"; }
trap restore EXIT

sed -i '' 's/^\tapp\.setupUpgradeHandlers()$/\t\/\/ app.setupUpgradeHandlers() -- disabled for build-pre-upgrade-binary.sh/' "$APP_GO"
if ! grep -q "disabled for build-pre-upgrade-binary.sh" "$APP_GO"; then
  echo "FAIL: sed didn't match — has app.go's setupUpgradeHandlers() call moved or been reworded?"
  exit 1
fi

echo "==> Building pre-v2 binary (v2 handler registration disabled) -> $OUT"
(cd "$ROOT_DIR/chain" && go build -o "$OUT" ./cmd/lcvd)

echo "==> Done. app.go restored to its normal (v2-registered) state."
