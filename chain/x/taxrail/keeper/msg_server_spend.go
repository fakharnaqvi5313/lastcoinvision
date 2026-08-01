package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lastcoinvision/lcv/x/taxrail/types"
)

// SpendCollectedTax moves funds out of a jurisdiction's collected-tax
// ledger to a recipient. Authority-gated (governance), per the top-level
// spec's requirement that collector outflows require a governance
// proposal.
func (k msgServer) SpendCollectedTax(goCtx context.Context, msg *types.MsgSpendCollectedTax) (*types.MsgSpendCollectedTaxResponse, error) {
	if k.GetAuthority() != msg.Authority {
		return nil, types.ErrInvalidSigner
	}
	if msg.Amount.Denom != types.DefaultDenom {
		return nil, types.ErrUnsupportedDenom
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.SubtractFromJurisdictionLedger(ctx, msg.JurisdictionCode, msg.Amount.Amount); err != nil {
		return nil, err
	}

	recipient, err := sdk.AccAddressFromBech32(msg.Recipient)
	if err != nil {
		return nil, err
	}
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, sdk.NewCoins(msg.Amount)); err != nil {
		return nil, err
	}

	if err := ctx.EventManager().EmitTypedEvent(&types.EventCollectedTaxSpent{
		JurisdictionCode: msg.JurisdictionCode,
		Recipient:        msg.Recipient,
		Amount:           msg.Amount,
		Reason:           msg.Reason,
	}); err != nil {
		return nil, err
	}

	return &types.MsgSpendCollectedTaxResponse{}, nil
}
