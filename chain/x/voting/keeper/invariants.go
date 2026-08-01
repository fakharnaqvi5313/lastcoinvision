package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lastcoinvision/lcv/x/voting/types"
)

// RegisterInvariants registers all x/voting invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "tallied-votes-not-exceed-registered",
		TalliedVotesNotExceedRegisteredInvariant(k))
}

// TalliedVotesNotExceedRegisteredInvariant checks, for every election, that
// the number of revealed votes never exceeds the number of registered
// voters — the acceptance-test invariant named explicitly in the top-level
// spec. CommitVote already enforces this per-voter via the nullifier
// uniqueness check and the "must be registered" precondition, so this
// invariant is a cross-check that those guarantees actually hold in
// stored state, not a substitute for them.
func TalliedVotesNotExceedRegisteredInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		return talliedVotesInvariant(ctx, k)
	}
}

func talliedVotesInvariant(ctx context.Context, k Keeper) (string, bool) {
	revealedByElection := map[string]int64{}
	for _, vc := range k.GetAllVoteCommitments(ctx) {
		if vc.Revealed {
			revealedByElection[vc.ElectionId]++
		}
	}

	var violations []string
	for electionID, revealed := range revealedByElection {
		registered := k.CountRegisteredVoters(ctx, electionID)
		if revealed > registered {
			violations = append(violations, fmt.Sprintf("election %s: %d revealed > %d registered", electionID, revealed, registered))
		}
	}

	if len(violations) > 0 {
		return sdk.FormatInvariant(
			types.ModuleName,
			"tallied-votes-not-exceed-registered",
			fmt.Sprintf("found %d election(s) with more revealed votes than registered voters: %v", len(violations), violations),
		), true
	}
	return "", false
}
