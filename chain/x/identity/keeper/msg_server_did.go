package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lastcoinvision/lcv/x/identity/types"
)

func didID(address string) string {
	return fmt.Sprintf("did:lcv:%s", address)
}

func (k msgServer) CreateDID(goCtx context.Context, msg *types.MsgCreateDID) (*types.MsgCreateDIDResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, found := k.GetDID(ctx, msg.Creator); found {
		return nil, types.ErrDIDAlreadyExists
	}

	now := ctx.BlockTime().Unix()
	doc := types.DIDDocument{
		Id:                 didID(msg.Creator),
		Controller:         msg.Creator,
		VerificationMethod: msg.VerificationMethod,
		Service:            msg.Service,
		Deactivated:        false,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := k.SetDID(ctx, doc); err != nil {
		return nil, err
	}

	if err := ctx.EventManager().EmitTypedEvent(&types.EventDIDCreated{
		Id:         doc.Id,
		Controller: doc.Controller,
	}); err != nil {
		return nil, err
	}

	return &types.MsgCreateDIDResponse{Id: doc.Id}, nil
}

func (k msgServer) UpdateDID(goCtx context.Context, msg *types.MsgUpdateDID) (*types.MsgUpdateDIDResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	doc, found := k.GetDID(ctx, msg.Creator)
	if !found {
		return nil, types.ErrDIDNotFound
	}
	if doc.Deactivated {
		return nil, types.ErrDIDDeactivated
	}

	doc.VerificationMethod = msg.VerificationMethod
	doc.Service = msg.Service
	doc.UpdatedAt = ctx.BlockTime().Unix()

	if err := k.SetDID(ctx, doc); err != nil {
		return nil, err
	}

	if err := ctx.EventManager().EmitTypedEvent(&types.EventDIDUpdated{
		Id: doc.Id,
	}); err != nil {
		return nil, err
	}

	return &types.MsgUpdateDIDResponse{}, nil
}

func (k msgServer) DeactivateDID(goCtx context.Context, msg *types.MsgDeactivateDID) (*types.MsgDeactivateDIDResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	doc, found := k.GetDID(ctx, msg.Creator)
	if !found {
		return nil, types.ErrDIDNotFound
	}
	if doc.Deactivated {
		return nil, types.ErrDIDDeactivated
	}

	doc.Deactivated = true
	doc.UpdatedAt = ctx.BlockTime().Unix()

	if err := k.SetDID(ctx, doc); err != nil {
		return nil, err
	}

	if err := ctx.EventManager().EmitTypedEvent(&types.EventDIDDeactivated{
		Id: doc.Id,
	}); err != nil {
		return nil, err
	}

	return &types.MsgDeactivateDIDResponse{}, nil
}
