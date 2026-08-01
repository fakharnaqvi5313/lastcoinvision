package types_test

import (
	"testing"

	"github.com/lastcoinvision/lcv/x/voting/types"
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
				Elections: []*types.Election{
					{Id: "election-1"},
				},
				RegisteredVoters: []*types.RegisteredVoter{
					{ElectionId: "election-1", Voter: "lcv1voter"},
				},
				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: true,
		},
		{
			desc: "invalid: registered voter references unknown election",
			genState: &types.GenesisState{
				RegisteredVoters: []*types.RegisteredVoter{
					{ElectionId: "no-such-election", Voter: "lcv1voter"},
				},
			},
			valid: false,
		},
		{
			desc: "invalid: duplicate election id",
			genState: &types.GenesisState{
				Elections: []*types.Election{
					{Id: "election-1"},
					{Id: "election-1"},
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
