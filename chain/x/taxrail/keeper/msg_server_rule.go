package keeper

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lastcoinvision/lcv/x/taxrail/types"
)

func (k msgServer) RegisterTaxRule(goCtx context.Context, msg *types.MsgRegisterTaxRule) (*types.MsgRegisterTaxRuleResponse, error) {
	if k.GetAuthority() != msg.Authority {
		return nil, types.ErrInvalidSigner
	}
	if msg.Rate.IsNegative() || msg.Rate.GT(math.LegacyOneDec()) {
		return nil, types.ErrInvalidRate
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.SetTaxRule(ctx, types.TaxRule{
		JurisdictionCode: msg.JurisdictionCode,
		Rate:             msg.Rate,
		SetAt:            ctx.BlockTime().Unix(),
	}); err != nil {
		return nil, err
	}

	if err := ctx.EventManager().EmitTypedEvent(&types.EventTaxRuleRegistered{
		JurisdictionCode: msg.JurisdictionCode,
		Rate:             msg.Rate,
	}); err != nil {
		return nil, err
	}

	return &types.MsgRegisterTaxRuleResponse{}, nil
}
