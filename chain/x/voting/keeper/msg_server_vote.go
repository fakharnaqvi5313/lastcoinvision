package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lastcoinvision/lcv/x/voting/types"
)

func (k msgServer) RegisterToVote(goCtx context.Context, msg *types.MsgRegisterToVote) (*types.MsgRegisterToVoteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	election, found := k.GetElection(ctx, msg.ElectionId)
	if !found {
		return nil, types.ErrElectionNotFound
	}
	now := ctx.BlockTime().Unix()
	if now < election.RegistrationStart || now >= election.RegistrationEnd {
		return nil, types.ErrNotRegistrationWindow
	}
	if k.IsRegistered(ctx, msg.ElectionId, msg.Voter) {
		return nil, types.ErrAlreadyRegistered
	}
	if !k.identityKeeper.HasActiveCredentialOfType(ctx, msg.Voter, election.EligibilityCredentialType) {
		return nil, types.ErrNotEligible
	}

	if err := k.SetRegisteredVoter(ctx, types.RegisteredVoter{
		ElectionId:   msg.ElectionId,
		Voter:        msg.Voter,
		RegisteredAt: now,
	}); err != nil {
		return nil, err
	}

	if err := ctx.EventManager().EmitTypedEvent(&types.EventVoterRegistered{
		ElectionId: msg.ElectionId,
	}); err != nil {
		return nil, err
	}

	return &types.MsgRegisterToVoteResponse{}, nil
}

func (k msgServer) CommitVote(goCtx context.Context, msg *types.MsgCommitVote) (*types.MsgCommitVoteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	election, found := k.GetElection(ctx, msg.ElectionId)
	if !found {
		return nil, types.ErrElectionNotFound
	}
	now := ctx.BlockTime().Unix()
	if now < election.VotingStart || now >= election.VotingEnd {
		return nil, types.ErrNotVotingWindow
	}
	if !k.IsRegistered(ctx, msg.ElectionId, msg.Voter) {
		return nil, types.ErrNotRegistered
	}

	null := nullifier(msg.Voter, msg.ElectionId)
	if _, found := k.GetVoteCommitment(ctx, msg.ElectionId, null); found {
		return nil, types.ErrAlreadyCommitted
	}

	if err := k.SetVoteCommitment(ctx, types.VoteCommitment{
		ElectionId:  msg.ElectionId,
		Nullifier:   null,
		Commitment:  msg.Commitment,
		CommittedAt: now,
	}); err != nil {
		return nil, err
	}

	if err := ctx.EventManager().EmitTypedEvent(&types.EventVoteCommitted{
		ElectionId: msg.ElectionId,
		Nullifier:  null,
	}); err != nil {
		return nil, err
	}

	return &types.MsgCommitVoteResponse{}, nil
}

func (k msgServer) RevealVote(goCtx context.Context, msg *types.MsgRevealVote) (*types.MsgRevealVoteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	election, found := k.GetElection(ctx, msg.ElectionId)
	if !found {
		return nil, types.ErrElectionNotFound
	}
	now := ctx.BlockTime().Unix()
	if now < election.VotingEnd || now >= election.RevealEnd {
		return nil, types.ErrNotRevealWindow
	}

	validChoice := false
	for _, opt := range election.Options {
		if opt == msg.Choice {
			validChoice = true
			break
		}
	}
	if !validChoice {
		return nil, types.ErrInvalidChoice
	}

	null := nullifier(msg.Voter, msg.ElectionId)
	commitment, found := k.GetVoteCommitment(ctx, msg.ElectionId, null)
	if !found {
		return nil, types.ErrNoCommitment
	}
	if commitment.Revealed {
		return nil, types.ErrAlreadyRevealed
	}
	if commitmentHash(msg.Choice, msg.Salt) != commitment.Commitment {
		return nil, types.ErrCommitmentMismatch
	}

	commitment.Revealed = true
	commitment.RevealedChoice = msg.Choice
	commitment.RevealedAt = now
	if err := k.SetVoteCommitment(ctx, commitment); err != nil {
		return nil, err
	}

	if err := ctx.EventManager().EmitTypedEvent(&types.EventVoteRevealed{
		ElectionId: msg.ElectionId,
		Nullifier:  null,
	}); err != nil {
		return nil, err
	}

	return &types.MsgRevealVoteResponse{}, nil
}
