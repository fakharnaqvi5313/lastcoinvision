package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lastcoinvision/lcv/x/identity/types"
)

func (k msgServer) AddIssuer(goCtx context.Context, msg *types.MsgAddIssuer) (*types.MsgAddIssuerResponse, error) {
	if k.GetAuthority() != msg.Authority {
		return nil, types.ErrInvalidSigner
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	if k.IsAuthorizedIssuer(ctx, msg.IssuerAddress) {
		return nil, types.ErrIssuerAlreadyExists
	}

	issuer := types.Issuer{
		Address: msg.IssuerAddress,
		Name:    msg.Name,
		AddedAt: ctx.BlockTime().Unix(),
	}
	if err := k.SetIssuer(ctx, issuer); err != nil {
		return nil, err
	}

	if err := ctx.EventManager().EmitTypedEvent(&types.EventIssuerAdded{
		Address: issuer.Address,
		Name:    issuer.Name,
	}); err != nil {
		return nil, err
	}

	return &types.MsgAddIssuerResponse{}, nil
}

func (k msgServer) RemoveIssuer(goCtx context.Context, msg *types.MsgRemoveIssuer) (*types.MsgRemoveIssuerResponse, error) {
	if k.GetAuthority() != msg.Authority {
		return nil, types.ErrInvalidSigner
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.IsAuthorizedIssuer(ctx, msg.IssuerAddress) {
		return nil, types.ErrIssuerNotFound
	}
	k.DeleteIssuer(ctx, msg.IssuerAddress)

	if err := ctx.EventManager().EmitTypedEvent(&types.EventIssuerRemoved{
		Address: msg.IssuerAddress,
	}); err != nil {
		return nil, err
	}

	return &types.MsgRemoveIssuerResponse{}, nil
}
