package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/lastcoinvision/lcv/x/voting/types"
)

func (k Keeper) Election(ctx context.Context, req *types.QueryElectionRequest) (*types.QueryElectionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	election, found := k.GetElection(ctx, req.Id)
	if !found {
		return nil, status.Error(codes.NotFound, "election not found")
	}
	return &types.QueryElectionResponse{Election: &election}, nil
}

// ElectionResults recomputes the tally live from every revealed vote
// commitment — there is no separately stored "official" result. Anyone can
// call this and recompute the same answer, per the top-level spec.
func (k Keeper) ElectionResults(ctx context.Context, req *types.QueryElectionResultsRequest) (*types.QueryElectionResultsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	counts := map[string]int64{}
	var totalRevealed int64
	for _, vc := range k.GetVoteCommitmentsByElection(ctx, req.ElectionId) {
		if !vc.Revealed {
			continue
		}
		counts[vc.RevealedChoice]++
		totalRevealed++
	}

	election, _ := k.GetElection(ctx, req.ElectionId)
	tallies := make([]*types.OptionTally, 0, len(election.Options))
	for _, opt := range election.Options {
		tallies = append(tallies, &types.OptionTally{Option: opt, Count: counts[opt]})
	}

	return &types.QueryElectionResultsResponse{
		Tallies:         tallies,
		TotalRevealed:   totalRevealed,
		TotalRegistered: k.CountRegisteredVoters(ctx, req.ElectionId),
	}, nil
}

func (k Keeper) ElectoralAuthorityList(ctx context.Context, req *types.QueryElectoralAuthorityListRequest) (*types.QueryElectoralAuthorityListResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	return &types.QueryElectoralAuthorityListResponse{Authorities: k.GetAllElectoralAuthorities(ctx)}, nil
}
