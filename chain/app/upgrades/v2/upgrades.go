// Package v2 is a sample named software upgrade: it contains no state
// migrations of its own (there's nothing in v2 that needs one), so
// CreateUpgradeHandler just runs whatever module migrations are pending —
// the minimum any named upgrade handler has to do to let the chain cross
// the upgrade height without halting. See scripts/upgrade/README.md for
// the governance walkthrough this exists to demonstrate.
package v2

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/types/module"
)

// UpgradeName must match the "name" field of the MsgSoftwareUpgrade
// governance proposal and the cosmovisor upgrades/<name>/ directory.
const UpgradeName = "v2"

func CreateUpgradeHandler(mm *module.Manager, configurator module.Configurator) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
