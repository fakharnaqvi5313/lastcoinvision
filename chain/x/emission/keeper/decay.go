package keeper

import (
	"cosmossdk.io/math"
)

// maxTaylorTerms bounds the series expansion used by expNeg. Elapsed time in
// units of half-lives (x = ln(2)*t/halfLife) stays in the tens even after a
// century of chain uptime at a multi-year half-life, and the Taylor series
// for e^-x needs roughly x+40 terms to converge past math.LegacyDec's 18
// decimal digits of precision, so this cap is generous headroom rather than
// a tight bound.
const maxTaylorTerms = 300

// taylorEpsilon is the term magnitude below which further Taylor terms can't
// move the sum within Dec's precision, so the series can stop early.
var taylorEpsilon = math.LegacyNewDecWithPrec(1, 18)

// expNeg computes e^-x for x >= 0 using a fixed-point Taylor series,
// entirely in cosmossdk.io/math.LegacyDec arithmetic (arbitrary-precision
// big.Int under the hood) so every validator computes the exact same result
// deterministically — a plain float64 math.Exp would risk subtle
// cross-platform divergence in consensus-critical state transitions.
func expNeg(x math.LegacyDec) math.LegacyDec {
	if x.IsNegative() {
		// Should never happen (x = lambda * elapsedSeconds with
		// elapsedSeconds >= 0), but clamp defensively rather than return a
		// nonsensical result.
		x = math.LegacyZeroDec()
	}

	sum := math.LegacyOneDec()
	term := math.LegacyOneDec()
	for k := int64(1); k <= maxTaylorTerms; k++ {
		// term_k = term_(k-1) * (-x / k)
		term = term.Mul(x).QuoInt64(k).Neg()
		sum = sum.Add(term)
		if term.Abs().LT(taylorEpsilon) {
			break
		}
	}
	return sum
}

// emittedBetween returns the amount (in totalRewardPool's unit) emitted by
// the exponential decay curve between elapsedFrom and elapsedTo seconds
// since the curve's start time, where the curve is defined so that the
// asymptotic total emitted as elapsed time goes to infinity equals
// totalRewardPool, and the emission rate halves every halfLifeSeconds:
//
//	lambda = ln(2) / halfLifeSeconds
//	emitted(t1, t2) = totalRewardPool * (e^(-lambda*t1) - e^(-lambda*t2))
//
// elapsedTo must be >= elapsedFrom; both must be >= 0.
func emittedBetween(totalRewardPool math.Int, halfLifeSeconds, elapsedFrom, elapsedTo int64) math.Int {
	if elapsedTo <= elapsedFrom || halfLifeSeconds <= 0 || totalRewardPool.IsZero() {
		return math.ZeroInt()
	}

	// ln(2) to 18 decimal digits.
	ln2 := math.LegacyMustNewDecFromStr("0.693147180559945309")
	lambda := ln2.QuoInt64(halfLifeSeconds)

	x1 := lambda.MulInt64(elapsedFrom)
	x2 := lambda.MulInt64(elapsedTo)

	fraction := expNeg(x1).Sub(expNeg(x2))
	if !fraction.IsPositive() {
		return math.ZeroInt()
	}

	poolDec := math.LegacyNewDecFromInt(totalRewardPool)
	return poolDec.Mul(fraction).TruncateInt()
}
