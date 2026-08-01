package types_test

import (
	"testing"

	"github.com/lastcoinvision/lcv/x/identity/types"
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
				Issuers: []*types.Issuer{
					{Address: "lcv1issuer", Name: "Test Issuer", AddedAt: 1},
				},
				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: true,
		},
		{
			desc: "invalid: duplicate issuer address",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				Issuers: []*types.Issuer{
					{Address: "lcv1issuer", Name: "A", AddedAt: 1},
					{Address: "lcv1issuer", Name: "B", AddedAt: 2},
				},
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
