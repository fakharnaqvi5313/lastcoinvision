package types

import "fmt"

// this line is used by starport scaffolding # genesis/types/import

// DefaultIndex is the default global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		// this line is used by starport scaffolding # genesis/types/default
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// this line is used by starport scaffolding # genesis/types/validate

	seenElections := make(map[string]bool, len(gs.Elections))
	for _, e := range gs.Elections {
		if seenElections[e.Id] {
			return fmt.Errorf("duplicate election id: %s", e.Id)
		}
		seenElections[e.Id] = true
	}

	seenAuthorities := make(map[string]bool, len(gs.ElectoralAuthorities))
	for _, a := range gs.ElectoralAuthorities {
		if seenAuthorities[a.Address] {
			return fmt.Errorf("duplicate electoral authority address: %s", a.Address)
		}
		seenAuthorities[a.Address] = true
	}

	for _, rv := range gs.RegisteredVoters {
		if !seenElections[rv.ElectionId] {
			return fmt.Errorf("registered voter references unknown election id: %s", rv.ElectionId)
		}
	}
	for _, vc := range gs.VoteCommitments {
		if !seenElections[vc.ElectionId] {
			return fmt.Errorf("vote commitment references unknown election id: %s", vc.ElectionId)
		}
	}

	return gs.Params.Validate()
}
