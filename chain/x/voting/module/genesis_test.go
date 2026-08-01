package voting_test

import (
	"testing"

	keepertest "github.com/lastcoinvision/lcv/testutil/keeper"
	"github.com/lastcoinvision/lcv/testutil/nullify"
	voting "github.com/lastcoinvision/lcv/x/voting/module"
	"github.com/lastcoinvision/lcv/x/voting/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.VotingKeeper(t)
	voting.InitGenesis(ctx, k, genesisState)
	got := voting.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
