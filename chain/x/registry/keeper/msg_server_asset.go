package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	identitytypes "github.com/lastcoinvision/lcv/x/identity/types"
	"github.com/lastcoinvision/lcv/x/registry/types"
)

func (k msgServer) RegisterAsset(goCtx context.Context, msg *types.MsgRegisterAsset) (*types.MsgRegisterAssetResponse, error) {
	if !k.IsJurisdictionAuthority(goCtx, msg.JurisdictionCode, msg.Creator) {
		return nil, types.ErrNotJurisdictionAuthority
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	now := ctx.BlockTime().Unix()
	id := assetID(msg.JurisdictionCode, msg.AssetType, msg.MetadataHash, now)

	asset := types.Asset{
		Id:               id,
		AssetType:        msg.AssetType,
		JurisdictionCode: msg.JurisdictionCode,
		MetadataHash:     msg.MetadataHash,
		Owner:            msg.Owner,
		Status:           types.AssetStatus_ASSET_STATUS_ACTIVE,
		RegisteredAt:     now,
		UpdatedAt:        now,
	}
	if err := k.SetAsset(ctx, asset); err != nil {
		return nil, err
	}

	if err := ctx.EventManager().EmitTypedEvent(&types.EventAssetRegistered{
		Id:               asset.Id,
		JurisdictionCode: asset.JurisdictionCode,
		Owner:            asset.Owner,
	}); err != nil {
		return nil, err
	}

	return &types.MsgRegisterAssetResponse{Id: id}, nil
}

func (k msgServer) TransferAsset(goCtx context.Context, msg *types.MsgTransferAsset) (*types.MsgTransferAssetResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	asset, found := k.GetAsset(ctx, msg.AssetId)
	if !found {
		return nil, types.ErrAssetNotFound
	}
	if asset.Status == types.AssetStatus_ASSET_STATUS_FROZEN {
		return nil, types.ErrAssetFrozen
	}
	isOwner := msg.Creator == didToAddress(asset.Owner)
	isAuthority := k.IsJurisdictionAuthority(ctx, asset.JurisdictionCode, msg.Creator)
	if !isOwner && !isAuthority {
		return nil, types.ErrNotAssetOwnerOrAuthority
	}

	oldOwnerAddr := didToAddress(asset.Owner)
	newOwnerAddr := didToAddress(msg.NewOwner)
	if !k.isAttestedOrAbove(ctx, oldOwnerAddr) || !k.isAttestedOrAbove(ctx, newOwnerAddr) {
		return nil, types.ErrPartyNotAttested
	}

	if err := k.AppendTransferRecord(ctx, types.TransferRecord{
		AssetId:   asset.Id,
		From:      asset.Owner,
		To:        msg.NewOwner,
		Timestamp: ctx.BlockTime().Unix(),
	}); err != nil {
		return nil, err
	}

	asset.Owner = msg.NewOwner
	asset.UpdatedAt = ctx.BlockTime().Unix()
	if err := k.SetAsset(ctx, asset); err != nil {
		return nil, err
	}

	if err := ctx.EventManager().EmitTypedEvent(&types.EventAssetTransferred{
		Id:   asset.Id,
		From: oldOwnerAddr,
		To:   newOwnerAddr,
	}); err != nil {
		return nil, err
	}

	return &types.MsgTransferAssetResponse{}, nil
}

func (k msgServer) FreezeAsset(goCtx context.Context, msg *types.MsgFreezeAsset) (*types.MsgFreezeAssetResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	asset, found := k.GetAsset(ctx, msg.AssetId)
	if !found {
		return nil, types.ErrAssetNotFound
	}
	if !k.IsJurisdictionAuthority(ctx, asset.JurisdictionCode, msg.Creator) {
		return nil, types.ErrNotJurisdictionAuthority
	}

	asset.Status = types.AssetStatus_ASSET_STATUS_FROZEN
	asset.FreezeReason = msg.Reason
	asset.UpdatedAt = ctx.BlockTime().Unix()
	if err := k.SetAsset(ctx, asset); err != nil {
		return nil, err
	}

	if err := ctx.EventManager().EmitTypedEvent(&types.EventAssetFrozen{
		Id:     asset.Id,
		Reason: msg.Reason,
	}); err != nil {
		return nil, err
	}

	return &types.MsgFreezeAssetResponse{}, nil
}

// isAttestedOrAbove reports whether address holds at least the "attested"
// identity tier (attested or institutional).
func (k Keeper) isAttestedOrAbove(ctx context.Context, address string) bool {
	tier, _ := k.identityKeeper.ComputeTier(ctx, address)
	return tier == identitytypes.AccountTier_ACCOUNT_TIER_ATTESTED || tier == identitytypes.AccountTier_ACCOUNT_TIER_INSTITUTIONAL
}
