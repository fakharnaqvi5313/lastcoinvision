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

	seenDIDs := make(map[string]bool, len(gs.Dids))
	for _, d := range gs.Dids {
		if seenDIDs[d.Id] {
			return fmt.Errorf("duplicate DID id: %s", d.Id)
		}
		seenDIDs[d.Id] = true
	}

	seenCreds := make(map[string]bool, len(gs.Credentials))
	for _, c := range gs.Credentials {
		if seenCreds[c.Id] {
			return fmt.Errorf("duplicate credential id: %s", c.Id)
		}
		seenCreds[c.Id] = true
	}

	seenIssuers := make(map[string]bool, len(gs.Issuers))
	for _, i := range gs.Issuers {
		if seenIssuers[i.Address] {
			return fmt.Errorf("duplicate issuer address: %s", i.Address)
		}
		seenIssuers[i.Address] = true
	}

	return gs.Params.Validate()
}
