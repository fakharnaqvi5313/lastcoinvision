#!/usr/bin/env bash
# Builds a "pre-v3" lcvd binary — chain/app/upgrades.go with only the v3
# handler's registration commented out (v2 stays registered, since a real
# chain crossing v3 has already crossed v2) — for
# submit-v3-upgrade-dryrun.sh to exercise against on a disposable devnet
# before anyone touches the real lcv-testnet-1 validators.
#
# Same reasoning as build-pre-upgrade-binary.sh (see its own comment and
# the 2026-08-03 DECISIONS.md entry): a binary that already has a plan's
# handler registered can't schedule that same plan again.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
UPGRADES_GO="$ROOT_DIR/chain/app/upgrades.go"
OUT="${1:-$ROOT_DIR/chain/build/lcvd-pre-v3}"

BACKUP="$(mktemp)"
cp "$UPGRADES_GO" "$BACKUP"
restore() { cp "$BACKUP" "$UPGRADES_GO"; rm -f "$BACKUP"; }
trap restore EXIT

python3 - "$UPGRADES_GO" <<'PYEOF'
import sys
path = sys.argv[1]
with open(path) as f:
    content = f.read()

handler_marker = '\tapp.UpgradeKeeper.SetUpgradeHandler(\n\t\tv3.UpgradeName,\n\t\tv3.CreateUpgradeHandler(app.ModuleManager, app.Configurator()),\n\t)\n'
if handler_marker not in content:
    print("FAIL: v3 registration block not found verbatim -- has upgrades.go changed shape?", file=sys.stderr)
    sys.exit(1)
content = content.replace(handler_marker, "\t// v3 registration disabled for build-pre-v3-binary.sh\n")

import_marker = '\tv3 "github.com/lastcoinvision/lcv/app/upgrades/v3"\n'
if import_marker not in content:
    print("FAIL: v3 import not found verbatim -- has upgrades.go changed shape?", file=sys.stderr)
    sys.exit(1)
content = content.replace(import_marker, "")

with open(path, "w") as f:
    f.write(content)
PYEOF

echo "==> Building pre-v3 binary (v3 handler registration disabled) -> $OUT"
(cd "$ROOT_DIR/chain" && go build -o "$OUT" ./cmd/lcvd)

echo "==> Done. upgrades.go restored to its normal (v2+v3-registered) state."
