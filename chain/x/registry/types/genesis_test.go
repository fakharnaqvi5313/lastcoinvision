package types_test

import (
	"testing"

	"github.com/lastcoinvision/lcv/x/registry/types"
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
				Assets: []*types.Asset{
					{Id: "asset-1", JurisdictionCode: "us-ca"},
				},
				TransferRecords: []*types.TransferRecord{
					{AssetId: "asset-1", From: "did:lcv:lcv1a", To: "did:lcv:lcv1b"},
				},
				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: true,
		},
		{
			desc: "invalid: transfer record references unknown asset",
			genState: &types.GenesisState{
				TransferRecords: []*types.TransferRecord{
					{AssetId: "no-such-asset", From: "did:lcv:lcv1a", To: "did:lcv:lcv1b"},
				},
			},
			valid: false,
		},
		{
			desc: "invalid: duplicate asset id",
			genState: &types.GenesisState{
				Assets: []*types.Asset{
					{Id: "asset-1"},
					{Id: "asset-1"},
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
