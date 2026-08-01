package keeper

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestExpNegZero(t *testing.T) {
	require.True(t, expNeg(math.LegacyZeroDec()).Equal(math.LegacyOneDec()))
}

func TestExpNegHalfLife(t *testing.T) {
	// e^-ln(2) == 0.5 exactly, by definition of ln.
	ln2 := math.LegacyMustNewDecFromStr("0.693147180559945309")
	got := expNeg(ln2)
	want := math.LegacyNewDecWithPrec(5, 1) // 0.5
	diff := got.Sub(want).Abs()
	require.Truef(t, diff.LT(math.LegacyNewDecWithPrec(1, 9)), "expNeg(ln2) = %s, want ~0.5", got)
}

func TestEmittedBetweenOneHalfLife(t *testing.T) {
	pool := math.NewInt(1_000_000_000)
	halfLife := int64(1000)

	// Over exactly one half-life, half the pool should be emitted.
	got := emittedBetween(pool, halfLife, 0, halfLife)
	want := pool.QuoRaw(2)
	diff := got.Sub(want).Abs()
	// Allow a tiny rounding tolerance (truncation to an integer amount).
	require.Truef(t, diff.LTE(math.NewInt(1)), "emittedBetween(one half-life) = %s, want ~%s", got, want)
}

func TestEmittedBetweenTenHalfLives(t *testing.T) {
	pool := math.NewInt(1_000_000_000)
	halfLife := int64(1000)

	// After 10 half-lives, 1 - 0.5^10 = 99.90234375% should be emitted.
	got := emittedBetween(pool, halfLife, 0, 10*halfLife)
	wantFraction := math.LegacyMustNewDecFromStr("0.9990234375")
	want := math.LegacyNewDecFromInt(pool).Mul(wantFraction).TruncateInt()
	diff := got.Sub(want).Abs()
	require.Truef(t, diff.LTE(math.NewInt(1)), "emittedBetween(10 half-lives) = %s, want ~%s", got, want)
}

func TestEmittedBetweenIsAdditive(t *testing.T) {
	// Splitting an interval into two adjacent pieces and summing must equal
	// emitting over the whole interval directly (no double counting, no gaps).
	pool := math.NewInt(1_200_000_000_000_000) // 1.2B LCV in ulcv
	halfLife := int64(2 * 365 * 24 * 3600)

	whole := emittedBetween(pool, halfLife, 0, 5*halfLife)
	part1 := emittedBetween(pool, halfLife, 0, 2*halfLife)
	part2 := emittedBetween(pool, halfLife, 2*halfLife, 5*halfLife)
	sum := part1.Add(part2)

	diff := whole.Sub(sum).Abs()
	require.Truef(t, diff.LTE(math.NewInt(1)), "whole = %s, part1+part2 = %s", whole, sum)
}

func TestEmittedBetweenNeverExceedsPool(t *testing.T) {
	pool := math.NewInt(1_200_000_000_000_000)
	halfLife := int64(2 * 365 * 24 * 3600)

	// Even over an enormous span (100 half-lives), emission must stay
	// strictly below the total pool (asymptotic limit, never reached).
	got := emittedBetween(pool, halfLife, 0, 100*halfLife)
	require.True(t, got.LT(pool))
}
