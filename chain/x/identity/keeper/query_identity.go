package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/lastcoinvision/lcv/x/identity/types"
)

func (k Keeper) DIDByAddress(ctx context.Context, req *types.QueryDIDByAddressRequest) (*types.QueryDIDByAddressResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	doc, found := k.GetDID(ctx, req.Address)
	if !found {
		return nil, status.Error(codes.NotFound, "no DID for this address")
	}
	return &types.QueryDIDByAddressResponse{Did: &doc}, nil
}

func (k Keeper) CredentialStatus(ctx context.Context, req *types.QueryCredentialStatusRequest) (*types.QueryCredentialStatusResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	cred, found := k.GetCredential(ctx, req.Id)
	if !found {
		return nil, status.Error(codes.NotFound, "credential not found")
	}
	return &types.QueryCredentialStatusResponse{Credential: &cred}, nil
}

func (k Keeper) IssuerList(ctx context.Context, req *types.QueryIssuerListRequest) (*types.QueryIssuerListResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	return &types.QueryIssuerListResponse{Issuers: k.GetAllIssuers(ctx)}, nil
}

func (k Keeper) TierOf(ctx context.Context, req *types.QueryTierOfRequest) (*types.QueryTierOfResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	tier, roleMetadata := k.ComputeTier(ctx, req.Address)
	return &types.QueryTierOfResponse{Tier: tier, RoleMetadata: roleMetadata}, nil
}
