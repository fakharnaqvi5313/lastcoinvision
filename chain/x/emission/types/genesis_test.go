package types_test

import (
	"testing"

	"github.com/lastcoinvision/lcv/x/emission/types"
	"github.com/stretchr/testify/require"
)

func TestGenesisState_Validate(t *testing.T) {
	tests := []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc: "valid genesis state",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: true,
		},
		{
			desc: "invalid: empty mint denom",
			genState: &types.GenesisState{
				Params: types.NewParams("", types.StakingRewardsAccountName, types.DefaultTotalRewardPool, types.DefaultHalfLifeSeconds, 0),
			},
			valid: false,
		},
		{
			desc: "invalid: zero half life",
			genState: &types.GenesisState{
				Params: types.NewParams(types.DefaultMintDenom, types.StakingRewardsAccountName, types.DefaultTotalRewardPool, 0, 0),
			},
			valid: false,
		},
		// this line is used by starport scaffolding # types/genesis/testcase
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
