package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lastcoinvision/lcv/x/voting/types"
)

func (k msgServer) AddElectoralAuthority(goCtx context.Context, msg *types.MsgAddElectoralAuthority) (*types.MsgAddElectoralAuthorityResponse, error) {
	if k.GetAuthority() != msg.Authority {
		return nil, types.ErrInvalidSigner
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.SetElectoralAuthority(ctx, types.ElectoralAuthority{
		Address: msg.Address,
		Name:    msg.Name,
		AddedAt: ctx.BlockTime().Unix(),
	}); err != nil {
		return nil, err
	}

	if err := ctx.EventManager().EmitTypedEvent(&types.EventElectoralAuthorityAdded{
		Address: msg.Address,
		Name:    msg.Name,
	}); err != nil {
		return nil, err
	}

	return &types.MsgAddElectoralAuthorityResponse{}, nil
}

func (k msgServer) RemoveElectoralAuthority(goCtx context.Context, msg *types.MsgRemoveElectoralAuthority) (*types.MsgRemoveElectoralAuthorityResponse, error) {
	if k.GetAuthority() != msg.Authority {
		return nil, types.ErrInvalidSigner
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	k.DeleteElectoralAuthority(ctx, msg.Address)

	if err := ctx.EventManager().EmitTypedEvent(&types.EventElectoralAuthorityRemoved{
		Address: msg.Address,
	}); err != nil {
		return nil, err
	}

	return &types.MsgRemoveElectoralAuthorityResponse{}, nil
}
