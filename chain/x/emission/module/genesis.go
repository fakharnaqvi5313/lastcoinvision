package emission

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lastcoinvision/lcv/x/emission/keeper"
	"github.com/lastcoinvision/lcv/x/emission/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// this line is used by starport scaffolding # genesis/module/init
	params := genState.Params
	if params.StartTime == 0 {
		// Anchor the decay curve to actual genesis time rather than the
		// zero value, unless a genesis file explicitly overrides it (e.g.
		// to align a testnet's curve with an earlier mainnet's).
		params.StartTime = ctx.BlockTime().Unix()
	}
	if err := k.SetParams(ctx, params); err != nil {
		panic(err)
	}

	lastEmissionTime := genState.LastEmissionTime
	if lastEmissionTime == 0 {
		// Fresh genesis: nothing has emitted yet, so anchor to the same
		// start time as the params curve itself.
		lastEmissionTime = params.StartTime
	}
	k.SetLastEmissionTime(ctx, lastEmissionTime)
}

// ExportGenesis returns the module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)
	if lastEmissionTime, found := k.GetLastEmissionTime(ctx); found {
		genesis.LastEmissionTime = lastEmissionTime
	}

	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
