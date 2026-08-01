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

	seenAssets := make(map[string]bool, len(gs.Assets))
	for _, a := range gs.Assets {
		if seenAssets[a.Id] {
			return fmt.Errorf("duplicate asset id: %s", a.Id)
		}
		seenAssets[a.Id] = true
	}

	seenJurisdictions := make(map[string]bool, len(gs.JurisdictionAuthorities))
	for _, j := range gs.JurisdictionAuthorities {
		if seenJurisdictions[j.JurisdictionCode] {
			return fmt.Errorf("duplicate jurisdiction authority code: %s", j.JurisdictionCode)
		}
		seenJurisdictions[j.JurisdictionCode] = true
	}

	for _, r := range gs.TransferRecords {
		if !seenAssets[r.AssetId] {
			return fmt.Errorf("transfer record references unknown asset id: %s", r.AssetId)
		}
	}

	return gs.Params.Validate()
}
