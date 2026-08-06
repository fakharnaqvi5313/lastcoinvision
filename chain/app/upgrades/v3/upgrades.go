// Package v3 is the first named software upgrade rehearsed for real on a
// persistent, live network (lcv-testnet-1) rather than a disposable devnet —
// see MAINNET_LAUNCH_RUNBOOK.md's "run under real conditions... at least one
// real governance-triggered upgrade rehearsed" gate. Like v2, it contains no
// state migrations of its own; the point of this one is exercising the real
// governance-vote-halt-swap sequence against production validator
// infrastructure, not a code change. See scripts/upgrade/README.md.
package v3

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/types/module"
)

// UpgradeName must match the "name" field of the MsgSoftwareUpgrade
// governance proposal and the cosmovisor upgrades/<name>/ directory.
const UpgradeName = "v3"

func CreateUpgradeHandler(mm *module.Manager, configurator module.Configurator) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
