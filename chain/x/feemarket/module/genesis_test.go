package feemarket_test

import (
	"testing"

	keepertest "github.com/lastcoinvision/lcv/testutil/keeper"
	"github.com/lastcoinvision/lcv/testutil/nullify"
	feemarket "github.com/lastcoinvision/lcv/x/feemarket/module"
	"github.com/lastcoinvision/lcv/x/feemarket/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.FeemarketKeeper(t)
	feemarket.InitGenesis(ctx, k, genesisState)
	got := feemarket.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
