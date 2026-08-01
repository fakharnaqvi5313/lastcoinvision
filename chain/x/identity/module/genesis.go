package identity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lastcoinvision/lcv/x/identity/keeper"
	"github.com/lastcoinvision/lcv/x/identity/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// this line is used by starport scaffolding # genesis/module/init
	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}
	for _, doc := range genState.Dids {
		if err := k.SetDID(ctx, *doc); err != nil {
			panic(err)
		}
	}
	for _, cred := range genState.Credentials {
		if err := k.SetCredential(ctx, *cred); err != nil {
			panic(err)
		}
	}
	for _, issuer := range genState.Issuers {
		if err := k.SetIssuer(ctx, *issuer); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns the module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)
	genesis.Dids = k.GetAllDIDs(ctx)
	genesis.Credentials = k.GetAllCredentials(ctx)
	genesis.Issuers = k.GetAllIssuers(ctx)

	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
