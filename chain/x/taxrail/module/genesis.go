package taxrail

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lastcoinvision/lcv/x/taxrail/keeper"
	"github.com/lastcoinvision/lcv/x/taxrail/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// this line is used by starport scaffolding # genesis/module/init
	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}
	for _, rule := range genState.TaxRules {
		if err := k.SetTaxRule(ctx, *rule); err != nil {
			panic(err)
		}
	}
	for _, ledger := range genState.JurisdictionLedgers {
		if err := k.SetJurisdictionLedger(ctx, *ledger); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns the module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)
	genesis.TaxRules = k.GetAllTaxRules(ctx)
	genesis.JurisdictionLedgers = k.GetAllJurisdictionLedgers(ctx)

	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
