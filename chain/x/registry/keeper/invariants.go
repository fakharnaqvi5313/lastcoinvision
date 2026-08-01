package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lastcoinvision/lcv/x/registry/types"
)

// RegisterInvariants registers all x/registry invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "transfer-history-referential-integrity",
		TransferHistoryReferentialIntegrityInvariant(k))
}

// TransferHistoryReferentialIntegrityInvariant checks that every recorded
// transfer references an asset that still exists — the transfer history is
// meant to be an immutable, fully verifiable audit trail of registered
// assets, so an orphaned record would indicate state corruption.
func TransferHistoryReferentialIntegrityInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		return transferHistoryInvariant(ctx, k)
	}
}

func transferHistoryInvariant(ctx context.Context, k Keeper) (string, bool) {
	var orphaned []string
	for _, record := range k.GetAllTransferRecords(ctx) {
		if _, found := k.GetAsset(ctx, record.AssetId); !found {
			orphaned = append(orphaned, record.AssetId)
		}
	}
	if len(orphaned) > 0 {
		return sdk.FormatInvariant(
			types.ModuleName,
			"transfer-history-referential-integrity",
			fmt.Sprintf("found %d transfer record(s) referencing nonexistent assets: %v", len(orphaned), orphaned),
		), true
	}
	return "", false
}
