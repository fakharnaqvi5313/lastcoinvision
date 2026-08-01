package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lastcoinvision/lcv/x/taxrail/types"
)

// TaxedTransfer atomically splits msg.Amount between the jurisdiction's
// collector treasury (the taxrail module account, tracked per-jurisdiction
// via the ledger) and the recipient, per the registered rate. If no rule is
// registered for the jurisdiction, the transfer is rejected outright rather
// than silently defaulting to 0% tax.
func (k msgServer) TaxedTransfer(goCtx context.Context, msg *types.MsgTaxedTransfer) (*types.MsgTaxedTransferResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Amount.Denom != types.DefaultDenom {
		return nil, types.ErrUnsupportedDenom
	}
	rule, found := k.GetTaxRule(ctx, msg.JurisdictionCode)
	if !found {
		return nil, types.ErrNoTaxRule
	}

	taxAmount := rule.Rate.MulInt(msg.Amount.Amount).TruncateInt()
	remainderAmount := msg.Amount.Amount.Sub(taxAmount)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}
	recipient, err := sdk.AccAddressFromBech32(msg.Recipient)
	if err != nil {
		return nil, err
	}

	if taxAmount.IsPositive() {
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, sdk.NewCoins(sdk.NewCoin(msg.Amount.Denom, taxAmount))); err != nil {
			return nil, err
		}
		if err := k.AddToJurisdictionLedger(ctx, msg.JurisdictionCode, taxAmount); err != nil {
			return nil, err
		}
	}
	if remainderAmount.IsPositive() {
		if err := k.bankKeeper.SendCoins(ctx, sender, recipient, sdk.NewCoins(sdk.NewCoin(msg.Amount.Denom, remainderAmount))); err != nil {
			return nil, err
		}
	}

	taxCoin := sdk.NewCoin(msg.Amount.Denom, taxAmount)
	remainderCoin := sdk.NewCoin(msg.Amount.Denom, remainderAmount)

	if err := ctx.EventManager().EmitTypedEvent(&types.EventTaxedTransfer{
		Sender:           msg.Sender,
		Recipient:        msg.Recipient,
		JurisdictionCode: msg.JurisdictionCode,
		TaxAmount:        taxCoin,
		RemainderAmount:  remainderCoin,
	}); err != nil {
		return nil, err
	}

	return &types.MsgTaxedTransferResponse{TaxAmount: taxCoin, RemainderAmount: remainderCoin}, nil
}
