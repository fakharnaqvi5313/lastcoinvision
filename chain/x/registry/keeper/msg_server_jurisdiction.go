package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lastcoinvision/lcv/x/registry/types"
)

func (k msgServer) SetJurisdictionAuthority(goCtx context.Context, msg *types.MsgSetJurisdictionAuthority) (*types.MsgSetJurisdictionAuthorityResponse, error) {
	if k.GetAuthority() != msg.Authority {
		return nil, types.ErrInvalidSigner
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	auth := types.JurisdictionAuthority{
		JurisdictionCode: msg.JurisdictionCode,
		Address:          msg.Address,
		SetAt:            ctx.BlockTime().Unix(),
	}
	if err := k.Keeper.SetJurisdictionAuthority(ctx, auth); err != nil {
		return nil, err
	}

	if err := ctx.EventManager().EmitTypedEvent(&types.EventJurisdictionAuthoritySet{
		JurisdictionCode: msg.JurisdictionCode,
		Address:          msg.Address,
	}); err != nil {
		return nil, err
	}

	return &types.MsgSetJurisdictionAuthorityResponse{}, nil
}
