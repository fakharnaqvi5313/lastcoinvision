package registry_test

import (
	"testing"

	keepertest "github.com/lastcoinvision/lcv/testutil/keeper"
	"github.com/lastcoinvision/lcv/testutil/nullify"
	registry "github.com/lastcoinvision/lcv/x/registry/module"
	"github.com/lastcoinvision/lcv/x/registry/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.RegistryKeeper(t)
	registry.InitGenesis(ctx, k, genesisState)
	got := registry.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
