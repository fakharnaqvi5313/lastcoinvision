package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/lastcoinvision/lcv/x/taxrail/types"
)

func (k Keeper) TaxRule(ctx context.Context, req *types.QueryTaxRuleRequest) (*types.QueryTaxRuleResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	rule, found := k.GetTaxRule(ctx, req.JurisdictionCode)
	if !found {
		return nil, status.Error(codes.NotFound, "no tax rule registered for this jurisdiction code")
	}
	return &types.QueryTaxRuleResponse{Rule: &rule}, nil
}

func (k Keeper) JurisdictionBalance(ctx context.Context, req *types.QueryJurisdictionBalanceRequest) (*types.QueryJurisdictionBalanceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ledger := k.GetJurisdictionLedger(ctx, req.JurisdictionCode)
	return &types.QueryJurisdictionBalanceResponse{Ledger: &ledger}, nil
}
