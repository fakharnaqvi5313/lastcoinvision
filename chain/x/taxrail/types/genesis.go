package types

import (
	"fmt"

	"cosmossdk.io/math"
)

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

	seenRules := make(map[string]bool, len(gs.TaxRules))
	for _, r := range gs.TaxRules {
		if seenRules[r.JurisdictionCode] {
			return fmt.Errorf("duplicate tax rule for jurisdiction code: %s", r.JurisdictionCode)
		}
		seenRules[r.JurisdictionCode] = true
		if r.Rate.IsNegative() || r.Rate.GT(math.LegacyOneDec()) {
			return fmt.Errorf("tax rule for %s has rate outside [0, 1]: %s", r.JurisdictionCode, r.Rate)
		}
	}

	seenLedgers := make(map[string]bool, len(gs.JurisdictionLedgers))
	for _, l := range gs.JurisdictionLedgers {
		if seenLedgers[l.JurisdictionCode] {
			return fmt.Errorf("duplicate jurisdiction ledger for code: %s", l.JurisdictionCode)
		}
		seenLedgers[l.JurisdictionCode] = true
		if l.Balance.IsNegative() {
			return fmt.Errorf("jurisdiction ledger for %s has negative balance: %s", l.JurisdictionCode, l.Balance)
		}
	}

	return gs.Params.Validate()
}
