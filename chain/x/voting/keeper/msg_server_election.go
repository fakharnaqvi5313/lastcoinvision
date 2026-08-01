package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lastcoinvision/lcv/x/voting/types"
)

func (k msgServer) CreateElection(goCtx context.Context, msg *types.MsgCreateElection) (*types.MsgCreateElectionResponse, error) {
	if !k.IsElectoralAuthority(goCtx, msg.Creator) {
		return nil, types.ErrNotElectoralAuthority
	}
	if msg.RegistrationStart >= msg.RegistrationEnd ||
		msg.RegistrationEnd > msg.VotingStart ||
		msg.VotingStart >= msg.VotingEnd ||
		msg.VotingEnd > msg.RevealEnd {
		return nil, types.ErrInvalidElectionWindows
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	id := k.NextElectionID(ctx)
	election := types.Election{
		Id:                        id,
		Title:                     msg.Title,
		Description:               msg.Description,
		EligibilityCredentialType: msg.EligibilityCredentialType,
		Options:                   msg.Options,
		RegistrationStart:         msg.RegistrationStart,
		RegistrationEnd:           msg.RegistrationEnd,
		VotingStart:               msg.VotingStart,
		VotingEnd:                 msg.VotingEnd,
		RevealEnd:                 msg.RevealEnd,
		CreatedBy:                 msg.Creator,
		CreatedAt:                 ctx.BlockTime().Unix(),
	}
	if err := k.SetElection(ctx, election); err != nil {
		return nil, err
	}

	if err := ctx.EventManager().EmitTypedEvent(&types.EventElectionCreated{
		Id:        election.Id,
		Title:     election.Title,
		CreatedBy: election.CreatedBy,
	}); err != nil {
		return nil, err
	}

	return &types.MsgCreateElectionResponse{Id: id}, nil
}
