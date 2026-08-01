package types_test

import (
	"testing"

	"cosmossdk.io/math"

	"github.com/lastcoinvision/lcv/x/taxrail/types"
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
				TaxRules: []*types.TaxRule{
					{JurisdictionCode: "us-ca", Rate: math.LegacyNewDecWithPrec(15, 2)},
				},
				JurisdictionLedgers: []*types.JurisdictionLedger{
					{JurisdictionCode: "us-ca", Denom: types.DefaultDenom, Balance: math.NewInt(1000)},
				},
				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: true,
		},
		{
			desc: "invalid: rate above 1",
			genState: &types.GenesisState{
				TaxRules: []*types.TaxRule{
					{JurisdictionCode: "us-ca", Rate: math.LegacyNewDec(2)},
				},
			},
			valid: false,
		},
		{
			desc: "invalid: negative ledger balance",
			genState: &types.GenesisState{
				JurisdictionLedgers: []*types.JurisdictionLedger{
					{JurisdictionCode: "us-ca", Denom: types.DefaultDenom, Balance: math.NewInt(-1)},
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
