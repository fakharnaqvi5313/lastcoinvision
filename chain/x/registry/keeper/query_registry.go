package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/lastcoinvision/lcv/x/registry/types"
)

func (k Keeper) AssetByID(ctx context.Context, req *types.QueryAssetByIDRequest) (*types.QueryAssetByIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	asset, found := k.GetAsset(ctx, req.Id)
	if !found {
		return nil, status.Error(codes.NotFound, "asset not found")
	}
	return &types.QueryAssetByIDResponse{Asset: &asset}, nil
}

func (k Keeper) TransferHistory(ctx context.Context, req *types.QueryTransferHistoryRequest) (*types.QueryTransferHistoryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	return &types.QueryTransferHistoryResponse{Records: k.GetTransferHistory(ctx, req.AssetId)}, nil
}

func (k Keeper) JurisdictionAuthorityOf(ctx context.Context, req *types.QueryJurisdictionAuthorityOfRequest) (*types.QueryJurisdictionAuthorityOfResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	auth, found := k.GetJurisdictionAuthority(ctx, req.JurisdictionCode)
	if !found {
		return nil, status.Error(codes.NotFound, "no jurisdiction authority set")
	}
	return &types.QueryJurisdictionAuthorityOfResponse{Authority: &auth}, nil
}
