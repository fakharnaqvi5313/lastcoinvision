package keeper_test

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	"github.com/stretchr/testify/require"

	testkeeper "github.com/lastcoinvision/lcv/testutil/keeper"
	"github.com/lastcoinvision/lcv/x/feemarket/types"
)

func TestAdjustBaseFeeRisesWhenBlockOverTarget(t *testing.T) {
	k, ctx := testkeeper.FeemarketKeeper(t)
	params := k.GetParams(ctx)

	gasMeter := storetypes.NewGasMeter(uint64(params.TargetGas) * 2)
	gasMeter.ConsumeGas(uint64(params.TargetGas)*2, "simulate full block")
	ctx = ctx.WithBlockGasMeter(gasMeter)

	require.NoError(t, k.AdjustBaseFee(ctx))

	newParams := k.GetParams(ctx)
	require.True(t, newParams.BaseFee.GT(params.BaseFee), "base fee should rise above target: got %s, was %s", newParams.BaseFee, params.BaseFee)
}

func TestAdjustBaseFeeFallsWhenBlockUnderTarget(t *testing.T) {
	k, ctx := testkeeper.FeemarketKeeper(t)
	// Start above the floor so a fall is observable.
	params := k.GetParams(ctx)
	params.BaseFee = params.MinBaseFee.MulInt64(10)
	require.NoError(t, k.SetParams(ctx, params))

	gasMeter := storetypes.NewGasMeter(uint64(params.TargetGas))
	// Consume nothing: empty block.
	ctx = ctx.WithBlockGasMeter(gasMeter)

	require.NoError(t, k.AdjustBaseFee(ctx))

	newParams := k.GetParams(ctx)
	require.True(t, newParams.BaseFee.LT(params.BaseFee), "base fee should fall below start: got %s, was %s", newParams.BaseFee, params.BaseFee)
}

func TestAdjustBaseFeeNeverBelowFloor(t *testing.T) {
	k, ctx := testkeeper.FeemarketKeeper(t)
	params := k.GetParams(ctx)

	gasMeter := storetypes.NewGasMeter(uint64(params.TargetGas))
	ctx = ctx.WithBlockGasMeter(gasMeter)

	// Repeatedly apply empty blocks; base fee must never drop below the floor.
	for i := 0; i < 50; i++ {
		require.NoError(t, k.AdjustBaseFee(ctx))
	}
	got := k.GetParams(ctx).BaseFee
	require.True(t, got.GTE(params.MinBaseFee), "base fee %s fell below floor %s", got, params.MinBaseFee)
}

func TestAdjustBaseFeeUnchangedAtTarget(t *testing.T) {
	k, ctx := testkeeper.FeemarketKeeper(t)
	params := k.GetParams(ctx)

	gasMeter := storetypes.NewGasMeter(uint64(params.TargetGas))
	gasMeter.ConsumeGas(uint64(params.TargetGas), "simulate exactly-on-target block")
	ctx = ctx.WithBlockGasMeter(gasMeter)

	require.NoError(t, k.AdjustBaseFee(ctx))

	newParams := k.GetParams(ctx)
	require.True(t, newParams.BaseFee.Equal(params.BaseFee), "base fee should stay unchanged at target: got %s, was %s", newParams.BaseFee, params.BaseFee)
}

func TestRequiredFeeAndBurnSplit(t *testing.T) {
	k, ctx := testkeeper.FeemarketKeeper(t)
	params := k.GetParams(ctx)

	gasWanted := uint64(200_000)
	required := k.RequiredFee(ctx, gasWanted)
	require.Equal(t, types.DefaultFeeDenom, required.Denom)
	require.True(t, required.Amount.IsPositive())

	burn, remainder := k.BurnSplit(ctx, required.Amount)
	require.True(t, burn.Add(remainder).Equal(required.Amount), "burn + remainder must equal the input amount exactly")

	wantBurn := required.Amount.ToLegacyDec().Mul(params.BurnPercentage).TruncateInt()
	require.Equal(t, wantBurn, burn)
}
