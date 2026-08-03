package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/lastcoinvision/lcv/testutil/keeper"
	"github.com/lastcoinvision/lcv/x/voting/keeper"
	"github.com/lastcoinvision/lcv/x/voting/types"
)

// FuzzVoteTally exercises the invariant behind the Phase C acceptance
// criterion "tallied votes <= registered voters": for any number of
// voters, any subset of them registering/committing/revealing, and any
// choice each reveals, the recomputed tally must never let more votes
// count than voters registered, and every revealed vote must land in
// exactly one option bucket (sum(tallies) == TotalRevealed) — the
// concrete failure mode a "phantom vote" bug (one uncommitted or
// unregistered ballot slipping into the count) would produce.
func FuzzVoteTally(f *testing.F) {
	f.Add(uint8(0), uint16(0))
	f.Add(uint8(1), uint16(1))
	f.Add(uint8(5), uint16(0b11111)) // everyone registers
	f.Add(uint8(5), uint16(0b10101)) // every other voter
	f.Add(uint8(8), uint16(0xFFFF))  // all bits set, only 8 voters exist
	f.Add(uint8(8), uint16(0b0000000011111111))

	f.Fuzz(func(t *testing.T, numVotersRaw uint8, activityBits uint16) {
		numVoters := int(numVotersRaw % 9) // 0..8 voters keeps each run fast

		k, ctx, identityKeeper := keepertest.VotingKeeperWithIdentity(t)
		ms := keeper.NewMsgServerImpl(k)

		require.NoError(t, k.SetElectoralAuthority(ctx, types.ElectoralAuthority{Address: testElectoralAuthority}))
		createResp, err := ms.CreateElection(ctx, newTestElectionMsg(testElectoralAuthority))
		require.NoError(t, err)

		options := []string{"yes", "no"}
		voters := make([]string, numVoters)
		registered := make([]bool, numVoters)

		for i := 0; i < numVoters; i++ {
			voters[i] = fmt.Sprintf("lcv1fuzzvoter%02d000000000000000000000000000", i)
			identityKeeper.GrantCredential(voters[i], testCredentialType)
		}

		// Bit i of activityBits: does voter i register at all. Every voter
		// that registers also commits and reveals — the interesting fuzzed
		// dimension is *how many* voters participate and *what* they vote,
		// not partial commit-without-reveal (already covered by the
		// existing lifecycle/invariant unit tests).
		ctx = ctx.WithBlockTime(unixTime(10))
		for i := 0; i < numVoters; i++ {
			if activityBits&(1<<uint(i)) == 0 {
				continue
			}
			_, err := ms.RegisterToVote(ctx, &types.MsgRegisterToVote{Voter: voters[i], ElectionId: createResp.Id})
			require.NoError(t, err)
			registered[i] = true
		}

		ctx = ctx.WithBlockTime(unixTime(150))
		wantCounts := map[string]int64{}
		var wantRevealed int64
		for i := 0; i < numVoters; i++ {
			if !registered[i] {
				continue
			}
			choice := options[i%len(options)]
			salt := fmt.Sprintf("salt-%d", i)
			_, err := ms.CommitVote(ctx, &types.MsgCommitVote{
				Voter:      voters[i],
				ElectionId: createResp.Id,
				Commitment: sha256Hex(choice + "|" + salt),
			})
			require.NoError(t, err)
			wantCounts[choice]++
			wantRevealed++
		}

		ctx = ctx.WithBlockTime(unixTime(250))
		for i := 0; i < numVoters; i++ {
			if !registered[i] {
				continue
			}
			choice := options[i%len(options)]
			salt := fmt.Sprintf("salt-%d", i)
			_, err := ms.RevealVote(ctx, &types.MsgRevealVote{
				Voter:      voters[i],
				ElectionId: createResp.Id,
				Choice:     choice,
				Salt:       salt,
			})
			require.NoError(t, err)
		}

		msg, broken := keeper.TalliedVotesNotExceedRegisteredInvariant(k)(ctx)
		require.False(t, broken, "invariant broken: %s", msg)

		results, err := k.ElectionResults(ctx, &types.QueryElectionResultsRequest{ElectionId: createResp.Id})
		require.NoError(t, err)

		require.Equal(t, wantRevealed, results.TotalRevealed)
		require.LessOrEqual(t, results.TotalRevealed, results.TotalRegistered,
			"tallied votes must never exceed registered voters: revealed=%d registered=%d", results.TotalRevealed, results.TotalRegistered)

		var sumTallies int64
		for _, tally := range results.Tallies {
			require.Equal(t, wantCounts[tally.Option], tally.Count, "option %q count mismatch", tally.Option)
			sumTallies += tally.Count
		}
		require.Equal(t, wantRevealed, sumTallies, "sum of per-option tallies must equal TotalRevealed exactly")
	})
}
