package app

import (
	v2 "github.com/lastcoinvision/lcv/app/upgrades/v2"
)

// setupUpgradeHandlers registers every named software upgrade this binary
// knows how to apply. A binary only needs the handler for upgrades at or
// after its own version — old handlers can be deleted once a chain has
// passed that height and nobody needs to state-sync through it anymore.
func (app *App) setupUpgradeHandlers() {
	app.UpgradeKeeper.SetUpgradeHandler(
		v2.UpgradeName,
		v2.CreateUpgradeHandler(app.ModuleManager, app.Configurator()),
	)
}
