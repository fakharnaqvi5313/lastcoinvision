package voting

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lastcoinvision/lcv/x/voting/keeper"
	"github.com/lastcoinvision/lcv/x/voting/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// this line is used by starport scaffolding # genesis/module/init
	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}
	for _, auth := range genState.ElectoralAuthorities {
		if err := k.SetElectoralAuthority(ctx, *auth); err != nil {
			panic(err)
		}
	}
	for _, election := range genState.Elections {
		if err := k.SetElection(ctx, *election); err != nil {
			panic(err)
		}
	}
	for _, voter := range genState.RegisteredVoters {
		if err := k.SetRegisteredVoter(ctx, *voter); err != nil {
			panic(err)
		}
	}
	for _, vc := range genState.VoteCommitments {
		if err := k.SetVoteCommitment(ctx, *vc); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns the module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)
	genesis.Elections = k.GetAllElections(ctx)
	genesis.RegisteredVoters = k.GetAllRegisteredVoters(ctx)
	genesis.VoteCommitments = k.GetAllVoteCommitments(ctx)
	genesis.ElectoralAuthorities = k.GetAllElectoralAuthorities(ctx)

	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
