package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AdjustBaseFee updates base_fee for the next block using the standard
// EIP-1559 formula, based on how much gas the block just finished executing
// (ctx.BlockGasMeter()) used relative to params.TargetGas:
//
//	delta = baseFee * (gasUsed - targetGas) / targetGas / maxBaseFeeChangeDenominator
//	newBaseFee = clamp(baseFee + delta, minBaseFee, +inf)
//
// A block exactly at the target leaves base_fee unchanged; a full block
// raises it by up to 1/maxBaseFeeChangeDenominator; an empty block lowers it
// by the same fraction, floored at min_base_fee so it can never reach zero
// (a zero base fee would make spam transactions free).
func (k Keeper) AdjustBaseFee(ctx sdk.Context) error {
	params := k.GetParams(ctx)
	if params.TargetGas == 0 {
		// Params not initialized yet — nothing to adjust.
		return nil
	}

	gasUsed := int64(ctx.BlockGasMeter().GasConsumedToLimit())
	gasDelta := gasUsed - params.TargetGas

	change := params.BaseFee.
		MulInt64(gasDelta).
		QuoInt64(params.TargetGas).
		QuoInt64(params.MaxBaseFeeChangeDenominator)

	newBaseFee := params.BaseFee.Add(change)
	if newBaseFee.LT(params.MinBaseFee) {
		newBaseFee = params.MinBaseFee
	}

	if newBaseFee.Equal(params.BaseFee) {
		return nil
	}
	params.BaseFee = newBaseFee
	return k.SetParams(ctx, params)
}

// RequiredFee returns the minimum total fee (base_fee * gasWanted) a
// transaction must pay in the current block, rounded up so that
// under-paying by a fractional unit is never possible.
func (k Keeper) RequiredFee(ctx sdk.Context, gasWanted uint64) sdk.Coin {
	params := k.GetParams(ctx)
	amount := params.BaseFee.MulInt64(int64(gasWanted)).Ceil().TruncateInt()
	return sdk.NewCoin(params.FeeDenom, amount)
}

// BurnSplit divides a base-fee payment into the portion to burn and the
// portion to forward to the fee collector, per params.BurnPercentage.
func (k Keeper) BurnSplit(ctx sdk.Context, baseFeeAmount math.Int) (burn, remainder math.Int) {
	params := k.GetParams(ctx)
	burn = math.LegacyNewDecFromInt(baseFeeAmount).Mul(params.BurnPercentage).TruncateInt()
	remainder = baseFeeAmount.Sub(burn)
	return burn, remainder
}
