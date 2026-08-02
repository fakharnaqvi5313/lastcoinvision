package emission_test

import (
	"testing"

	keepertest "github.com/lastcoinvision/lcv/testutil/keeper"
	"github.com/lastcoinvision/lcv/testutil/nullify"
	emission "github.com/lastcoinvision/lcv/x/emission/module"
	"github.com/lastcoinvision/lcv/x/emission/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.EmissionKeeper(t)
	emission.InitGenesis(ctx, k, genesisState)
	got := emission.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}

// TestGenesis_RoundTripsLastEmissionTime guards against a real bug found
// while verifying Phase D: ExportGenesis previously dropped
// LastEmissionTime entirely, and InitGenesis always reset it back to
// params.StartTime — so a chain export/import (upgrade, state-sync) would
// silently rewind the decay curve's bookkeeping to genesis time, discarding
// whatever time had actually elapsed. See DECISIONS.md.
func TestGenesis_RoundTripsLastEmissionTime(t *testing.T) {
	k, ctx := keepertest.EmissionKeeper(t)

	params := types.DefaultParams()
	params.StartTime = 1_700_000_000
	genesisState := types.GenesisState{
		Params:           params,
		LastEmissionTime: 1_700_500_000, // well after StartTime: emission has been running a while
	}

	emission.InitGenesis(ctx, k, genesisState)

	lastEmissionTime, found := k.GetLastEmissionTime(ctx)
	require.True(t, found)
	require.Equal(t, genesisState.LastEmissionTime, lastEmissionTime)

	exported := emission.ExportGenesis(ctx, k)
	require.Equal(t, genesisState.LastEmissionTime, exported.LastEmissionTime)
}

// TestGenesis_FreshGenesisAnchorsLastEmissionTimeToStartTime covers the
// zero-value case: a genuinely fresh genesis (LastEmissionTime unset)
// should still anchor to params.StartTime, exactly as before this fix.
func TestGenesis_FreshGenesisAnchorsLastEmissionTimeToStartTime(t *testing.T) {
	k, ctx := keepertest.EmissionKeeper(t)

	params := types.DefaultParams()
	params.StartTime = 1_700_000_000
	genesisState := types.GenesisState{Params: params}

	emission.InitGenesis(ctx, k, genesisState)

	lastEmissionTime, found := k.GetLastEmissionTime(ctx)
	require.True(t, found)
	require.Equal(t, params.StartTime, lastEmissionTime)
}
