package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	identitytypes "github.com/lastcoinvision/lcv/x/identity/types"
)

// IdentityKeeper defines the expected interface for the identity module,
// used to enforce that both parties to an asset transfer hold at least the
// "attested" tier, per the top-level spec.
type IdentityKeeper interface {
	ComputeTier(ctx context.Context, address string) (identitytypes.AccountTier, string)
}

// AccountKeeper defines the expected interface for the Account module.
type AccountKeeper interface {
	GetAccount(context.Context, sdk.AccAddress) sdk.AccountI // only used for simulation
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface for the Bank module.
type BankKeeper interface {
	SpendableCoins(context.Context, sdk.AccAddress) sdk.Coins
	// Methods imported from bank should be defined here
}

// ParamSubspace defines the expected Subspace interface for parameters.
type ParamSubspace interface {
	Get(context.Context, []byte, interface{})
	Set(context.Context, []byte, interface{})
}
