package types_test

import (
	"testing"

	"github.com/lastcoinvision/lcv/x/feemarket/types"
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
			desc: "invalid: base fee below min",
			genState: &types.GenesisState{
				Params: types.NewParams(
					types.DefaultFeeDenom,
					types.DefaultMinBaseFee.QuoInt64(2),
					types.DefaultMinBaseFee,
					types.DefaultTargetGas,
					types.DefaultMaxBaseFeeChangeDenominator,
					types.DefaultBurnPercentage,
				),
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
