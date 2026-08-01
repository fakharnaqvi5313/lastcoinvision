package taxrail_test

import (
	"testing"

	keepertest "github.com/lastcoinvision/lcv/testutil/keeper"
	"github.com/lastcoinvision/lcv/testutil/nullify"
	taxrail "github.com/lastcoinvision/lcv/x/taxrail/module"
	"github.com/lastcoinvision/lcv/x/taxrail/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.TaxrailKeeper(t)
	taxrail.InitGenesis(ctx, k, genesisState)
	got := taxrail.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
