package keeper

import (
	"context"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/runtime"

	"github.com/lastcoinvision/lcv/x/taxrail/types"
)

// -- Tax rules --

func (k Keeper) SetTaxRule(ctx context.Context, rule types.TaxRule) error {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz, err := k.cdc.Marshal(&rule)
	if err != nil {
		return err
	}
	store.Set(types.TaxRuleKey(rule.JurisdictionCode), bz)
	return nil
}

func (k Keeper) GetTaxRule(ctx context.Context, jurisdictionCode string) (types.TaxRule, bool) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.TaxRuleKey(jurisdictionCode))
	if bz == nil {
		return types.TaxRule{}, false
	}
	var rule types.TaxRule
	k.cdc.MustUnmarshal(bz, &rule)
	return rule, true
}

func (k Keeper) GetAllTaxRules(ctx context.Context) []*types.TaxRule {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iterator := store.Iterator(types.TaxRuleKeyPrefix, storeIteratorEnd(types.TaxRuleKeyPrefix))
	defer iterator.Close()

	var rules []*types.TaxRule
	for ; iterator.Valid(); iterator.Next() {
		var rule types.TaxRule
		k.cdc.MustUnmarshal(iterator.Value(), &rule)
		rules = append(rules, &rule)
	}
	return rules
}

// -- Jurisdiction ledgers --

func (k Keeper) SetJurisdictionLedger(ctx context.Context, ledger types.JurisdictionLedger) error {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz, err := k.cdc.Marshal(&ledger)
	if err != nil {
		return err
	}
	store.Set(types.JurisdictionLedgerKey(ledger.JurisdictionCode), bz)
	return nil
}

func (k Keeper) GetJurisdictionLedger(ctx context.Context, jurisdictionCode string) types.JurisdictionLedger {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.JurisdictionLedgerKey(jurisdictionCode))
	if bz == nil {
		return types.JurisdictionLedger{JurisdictionCode: jurisdictionCode, Denom: types.DefaultDenom, Balance: math.ZeroInt()}
	}
	var ledger types.JurisdictionLedger
	k.cdc.MustUnmarshal(bz, &ledger)
	return ledger
}

// AddToJurisdictionLedger credits amount to a jurisdiction's tracked
// collected-tax balance.
func (k Keeper) AddToJurisdictionLedger(ctx context.Context, jurisdictionCode string, amount math.Int) error {
	ledger := k.GetJurisdictionLedger(ctx, jurisdictionCode)
	ledger.Balance = ledger.Balance.Add(amount)
	return k.SetJurisdictionLedger(ctx, ledger)
}

// SubtractFromJurisdictionLedger debits amount from a jurisdiction's
// tracked collected-tax balance, failing if the balance is insufficient.
func (k Keeper) SubtractFromJurisdictionLedger(ctx context.Context, jurisdictionCode string, amount math.Int) error {
	ledger := k.GetJurisdictionLedger(ctx, jurisdictionCode)
	if ledger.Balance.LT(amount) {
		return types.ErrInsufficientLedger
	}
	ledger.Balance = ledger.Balance.Sub(amount)
	return k.SetJurisdictionLedger(ctx, ledger)
}

func (k Keeper) GetAllJurisdictionLedgers(ctx context.Context) []*types.JurisdictionLedger {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iterator := store.Iterator(types.JurisdictionLedgerKeyPrefix, storeIteratorEnd(types.JurisdictionLedgerKeyPrefix))
	defer iterator.Close()

	var ledgers []*types.JurisdictionLedger
	for ; iterator.Valid(); iterator.Next() {
		var ledger types.JurisdictionLedger
		k.cdc.MustUnmarshal(iterator.Value(), &ledger)
		ledgers = append(ledgers, &ledger)
	}
	return ledgers
}

// storeIteratorEnd returns the exclusive end key for iterating all keys
// sharing prefix: prefix with its last byte incremented, which works as
// long as the prefix doesn't end in 0xff (none of ours do).
func storeIteratorEnd(prefix []byte) []byte {
	end := make([]byte, len(prefix))
	copy(end, prefix)
	end[len(end)-1]++
	return end
}
