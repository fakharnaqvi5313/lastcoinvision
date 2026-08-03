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

	// Unlike RequiredFee's gasWanted (a single tx's own uint64 field, fully
	// attacker-controlled with no upstream bound — see the fix above and
	// DECISIONS.md), GasConsumedToLimit is capped by the block gas limit
	// consensus param across the whole block; reaching anywhere near
	// int64's range would require a limit no real chain configures.
	gasUsed := int64(ctx.BlockGasMeter().GasConsumedToLimit()) // #nosec G115
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
	// gasWanted comes straight from the tx (feeTx.GetGas() in
	// x/feemarket/ante), so it must never go through an int64 cast: a
	// gasWanted at or above 2^63 would wrap negative, making the required
	// fee negative and letting any paid amount (including zero) satisfy
	// the ante's paid.LT(required) check — a full fee-market bypass. Found
	// by gosec (G115); see DECISIONS.md and AUDIT_SCOPE.md.
	amount := params.BaseFee.MulInt(math.NewIntFromUint64(gasWanted)).Ceil().TruncateInt()
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
