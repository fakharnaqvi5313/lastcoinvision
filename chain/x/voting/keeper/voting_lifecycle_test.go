package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/lastcoinvision/lcv/testutil/keeper"
	"github.com/lastcoinvision/lcv/x/voting/keeper"
	"github.com/lastcoinvision/lcv/x/voting/types"
)

const (
	testElectoralAuthority = "lcv1electoralauthority00000000000000000000"
	testVoterAddr          = "lcv1voter00000000000000000000000000000000"
	testVoter2Addr         = "lcv1voter20000000000000000000000000000000"
	testCredentialType     = "national_id"
)

func newTestElectionMsg(creator string) *types.MsgCreateElection {
	return &types.MsgCreateElection{
		Creator:                   creator,
		Title:                     "Test Referendum",
		Description:               "A test referendum",
		EligibilityCredentialType: testCredentialType,
		Options:                   []string{"yes", "no"},
		RegistrationStart:         0,
		RegistrationEnd:           100,
		VotingStart:               100,
		VotingEnd:                 200,
		RevealEnd:                 300,
	}
}

// TestAcceptance_ElectionLifecycle is the Phase C acceptance flow for
// x/voting: electoral authority created -> election created -> eligible
// voter registers -> commits a vote -> reveals it -> the tally is
// recomputable live from on-chain data.
func TestAcceptance_ElectionLifecycle(t *testing.T) {
	k, ctx, identityKeeper := keepertest.VotingKeeperWithIdentity(t)
	ms := keeper.NewMsgServerImpl(k)

	require.NoError(t, k.SetElectoralAuthority(ctx, types.ElectoralAuthority{Address: testElectoralAuthority}))

	createResp, err := ms.CreateElection(ctx, newTestElectionMsg(testElectoralAuthority))
	require.NoError(t, err)
	require.NotEmpty(t, createResp.Id)

	// A non-authority cannot create elections.
	_, err = ms.CreateElection(ctx, newTestElectionMsg(testVoterAddr))
	require.ErrorIs(t, err, types.ErrNotElectoralAuthority)

	identityKeeper.GrantCredential(testVoterAddr, testCredentialType)

	ctx = ctx.WithBlockTime(unixTime(10))
	_, err = ms.RegisterToVote(ctx, &types.MsgRegisterToVote{Voter: testVoterAddr, ElectionId: createResp.Id})
	require.NoError(t, err)

	// An ineligible voter (no credential) cannot register.
	_, err = ms.RegisterToVote(ctx, &types.MsgRegisterToVote{Voter: testVoter2Addr, ElectionId: createResp.Id})
	require.ErrorIs(t, err, types.ErrNotEligible)

	// Commit during the voting window.
	commitment := "" // computed below via the same hash the handler expects
	choice, salt := "yes", "super-secret-salt"
	commitment = sha256Hex(choice + "|" + salt)

	ctx = ctx.WithBlockTime(unixTime(150))
	_, err = ms.CommitVote(ctx, &types.MsgCommitVote{
		Voter:      testVoterAddr,
		ElectionId: createResp.Id,
		Commitment: commitment,
	})
	require.NoError(t, err)

	// Double-commit is rejected.
	_, err = ms.CommitVote(ctx, &types.MsgCommitVote{
		Voter:      testVoterAddr,
		ElectionId: createResp.Id,
		Commitment: commitment,
	})
	require.ErrorIs(t, err, types.ErrAlreadyCommitted)

	// Reveal after the voting window closes.
	ctx = ctx.WithBlockTime(unixTime(250))
	_, err = ms.RevealVote(ctx, &types.MsgRevealVote{
		Voter:      testVoterAddr,
		ElectionId: createResp.Id,
		Choice:     choice,
		Salt:       salt,
	})
	require.NoError(t, err)

	// The tally is recomputable live from on-chain data.
	results, err := k.ElectionResults(ctx, &types.QueryElectionResultsRequest{ElectionId: createResp.Id})
	require.NoError(t, err)
	require.EqualValues(t, 1, results.TotalRevealed)
	require.EqualValues(t, 1, results.TotalRegistered)
	var yesCount int64
	for _, tally := range results.Tallies {
		if tally.Option == "yes" {
			yesCount = tally.Count
		}
	}
	require.EqualValues(t, 1, yesCount)
}

func TestRevealRejectsWrongSaltOrChoice(t *testing.T) {
	k, ctx, identityKeeper := keepertest.VotingKeeperWithIdentity(t)
	ms := keeper.NewMsgServerImpl(k)
	require.NoError(t, k.SetElectoralAuthority(ctx, types.ElectoralAuthority{Address: testElectoralAuthority}))
	createResp, err := ms.CreateElection(ctx, newTestElectionMsg(testElectoralAuthority))
	require.NoError(t, err)
	identityKeeper.GrantCredential(testVoterAddr, testCredentialType)

	ctx = ctx.WithBlockTime(unixTime(10))
	_, err = ms.RegisterToVote(ctx, &types.MsgRegisterToVote{Voter: testVoterAddr, ElectionId: createResp.Id})
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(unixTime(150))
	_, err = ms.CommitVote(ctx, &types.MsgCommitVote{
		Voter:      testVoterAddr,
		ElectionId: createResp.Id,
		Commitment: sha256Hex("yes|realsalt"),
	})
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(unixTime(250))
	_, err = ms.RevealVote(ctx, &types.MsgRevealVote{
		Voter:      testVoterAddr,
		ElectionId: createResp.Id,
		Choice:     "yes",
		Salt:       "wrongsalt",
	})
	require.ErrorIs(t, err, types.ErrCommitmentMismatch)
}

func TestTalliedVotesInvariant(t *testing.T) {
	k, ctx, identityKeeper := keepertest.VotingKeeperWithIdentity(t)
	ms := keeper.NewMsgServerImpl(k)
	require.NoError(t, k.SetElectoralAuthority(ctx, types.ElectoralAuthority{Address: testElectoralAuthority}))
	createResp, err := ms.CreateElection(ctx, newTestElectionMsg(testElectoralAuthority))
	require.NoError(t, err)
	identityKeeper.GrantCredential(testVoterAddr, testCredentialType)

	msg, broken := keeper.TalliedVotesNotExceedRegisteredInvariant(k)(ctx)
	require.False(t, broken)
	require.Empty(t, msg)

	ctx = ctx.WithBlockTime(unixTime(10))
	_, err = ms.RegisterToVote(ctx, &types.MsgRegisterToVote{Voter: testVoterAddr, ElectionId: createResp.Id})
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(unixTime(150))
	_, err = ms.CommitVote(ctx, &types.MsgCommitVote{
		Voter:      testVoterAddr,
		ElectionId: createResp.Id,
		Commitment: sha256Hex("yes|salt"),
	})
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(unixTime(250))
	_, err = ms.RevealVote(ctx, &types.MsgRevealVote{
		Voter:      testVoterAddr,
		ElectionId: createResp.Id,
		Choice:     "yes",
		Salt:       "salt",
	})
	require.NoError(t, err)

	// One registered, one revealed: still holds (1 <= 1).
	msg, broken = keeper.TalliedVotesNotExceedRegisteredInvariant(k)(ctx)
	require.False(t, broken)
	require.Empty(t, msg)
}
