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
	mintDenom, stakingRewardsAccount string,
	totalRewardPool math.Int,
	halfLifeSeconds, startTime int64,
) Params {
	return Params{
		MintDenom:             mintDenom,
		StakingRewardsAccount: stakingRewardsAccount,
		TotalRewardPool:       totalRewardPool,
		HalfLifeSeconds:       halfLifeSeconds,
		StartTime:             startTime,
	}
}

// DefaultParams returns a default set of parameters. StartTime is left at
// zero here and must be set to the genesis time by InitGenesis, since it
// isn't known until genesis is actually created.
func DefaultParams() Params {
	return NewParams(
		DefaultMintDenom,
		StakingRewardsAccountName,
		DefaultTotalRewardPool,
		DefaultHalfLifeSeconds,
		0,
	)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateDenom(p.MintDenom); err != nil {
		return err
	}
	if p.StakingRewardsAccount == "" {
		return fmt.Errorf("staking rewards account must not be empty")
	}
	if p.TotalRewardPool.IsNegative() {
		return fmt.Errorf("total reward pool must not be negative: %s", p.TotalRewardPool)
	}
	if p.HalfLifeSeconds <= 0 {
		return fmt.Errorf("half life seconds must be positive: %d", p.HalfLifeSeconds)
	}
	if p.StartTime < 0 {
		return fmt.Errorf("start time must not be negative: %d", p.StartTime)
	}
	return nil
}

func validateDenom(denom string) error {
	if denom == "" {
		return fmt.Errorf("mint denom must not be empty")
	}
	return nil
}
