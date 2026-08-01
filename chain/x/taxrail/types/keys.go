package types

const (
	// ModuleName defines the module name
	ModuleName = "taxrail"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_taxrail"
)

var (
	ParamsKey = []byte("p_taxrail")

	// TaxRuleKeyPrefix + <jurisdiction_code> -> TaxRule
	TaxRuleKeyPrefix = []byte("tax_rule/")

	// JurisdictionLedgerKeyPrefix + <jurisdiction_code> -> JurisdictionLedger
	JurisdictionLedgerKeyPrefix = []byte("jurisdiction_ledger/")
)

// DefaultDenom is the only denom x/taxrail operates in. MsgTaxedTransfer
// and MsgSpendCollectedTax both reject any other denom outright rather
// than tracking a per-denom ledger, since the top-level spec's
// split-payment primitive is scoped to ulcv — see modules/taxrail/SPEC.md.
const DefaultDenom = "ulcv"

func KeyPrefix(p string) []byte {
	return []byte(p)
}

// TaxRuleKey returns the store key for a jurisdiction's tax rule.
func TaxRuleKey(jurisdictionCode string) []byte {
	return append(TaxRuleKeyPrefix, []byte(jurisdictionCode)...)
}

// JurisdictionLedgerKey returns the store key for a jurisdiction's
// collected-tax ledger.
func JurisdictionLedgerKey(jurisdictionCode string) []byte {
	return append(JurisdictionLedgerKeyPrefix, []byte(jurisdictionCode)...)
}
