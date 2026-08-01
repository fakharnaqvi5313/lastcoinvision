package types

import (
	fmt "fmt"

	"cosmossdk.io/math"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	feeDenom string,
	baseFee, minBaseFee math.LegacyDec,
	targetGas, maxBaseFeeChangeDenominator int64,
	burnPercentage math.LegacyDec,
) Params {
	return Params{
		FeeDenom:                    feeDenom,
		BaseFee:                     baseFee,
		MinBaseFee:                  minBaseFee,
		TargetGas:                   targetGas,
		MaxBaseFeeChangeDenominator: maxBaseFeeChangeDenominator,
		BurnPercentage:              burnPercentage,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		DefaultFeeDenom,
		DefaultBaseFee,
		DefaultMinBaseFee,
		DefaultTargetGas,
		DefaultMaxBaseFeeChangeDenominator,
		DefaultBurnPercentage,
	)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if p.FeeDenom == "" {
		return fmt.Errorf("fee denom must not be empty")
	}
	if p.BaseFee.IsNegative() {
		return fmt.Errorf("base fee must not be negative: %s", p.BaseFee)
	}
	if p.MinBaseFee.IsNegative() {
		return fmt.Errorf("min base fee must not be negative: %s", p.MinBaseFee)
	}
	if p.BaseFee.LT(p.MinBaseFee) {
		return fmt.Errorf("base fee %s must not be below min base fee %s", p.BaseFee, p.MinBaseFee)
	}
	if p.TargetGas <= 0 {
		return fmt.Errorf("target gas must be positive: %d", p.TargetGas)
	}
	if p.MaxBaseFeeChangeDenominator <= 0 {
		return fmt.Errorf("max base fee change denominator must be positive: %d", p.MaxBaseFeeChangeDenominator)
	}
	if p.BurnPercentage.IsNegative() || p.BurnPercentage.GT(math.LegacyOneDec()) {
		return fmt.Errorf("burn percentage must be between 0 and 1: %s", p.BurnPercentage)
	}
	return nil
}
