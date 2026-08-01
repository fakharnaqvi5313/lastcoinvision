package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/lastcoinvision/lcv/x/taxrail/types"
)

// RegisterInvariants registers all x/taxrail invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "ledger-matches-module-balance",
		LedgerMatchesModuleBalanceInvariant(k))
}

// LedgerMatchesModuleBalanceInvariant checks that the sum of every
// jurisdiction's tracked collected-tax balance equals the taxrail module
// account's actual bank balance. The two are updated together by every
// code path that touches either (TaxedTransfer, SpendCollectedTax), so any
// divergence would mean a bug let coins move without the ledger following,
// or vice versa.
func LedgerMatchesModuleBalanceInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		return ledgerMatchesModuleBalanceInvariant(ctx, k)
	}
}

func ledgerMatchesModuleBalanceInvariant(ctx context.Context, k Keeper) (string, bool) {
	total := math.ZeroInt()
	for _, ledger := range k.GetAllJurisdictionLedgers(ctx) {
		total = total.Add(ledger.Balance)
	}

	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
	actual := k.bankKeeper.GetBalance(ctx, moduleAddr, types.DefaultDenom).Amount

	if !total.Equal(actual) {
		return sdk.FormatInvariant(
			types.ModuleName,
			"ledger-matches-module-balance",
			fmt.Sprintf("sum of jurisdiction ledgers (%s) does not equal module account balance (%s)", total, actual),
		), true
	}
	return "", false
}
