package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lastcoinvision/lcv/x/identity/types"
)

func (k msgServer) IssueCredential(goCtx context.Context, msg *types.MsgIssueCredential) (*types.MsgIssueCredentialResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.IsAuthorizedIssuer(ctx, msg.Issuer) {
		return nil, types.ErrIssuerNotFound
	}

	issuedAt := ctx.BlockTime().Unix()
	id := credentialID(msg.Issuer, msg.Subject, msg.CredentialType, msg.CredentialHash, issuedAt)

	cred := types.Credential{
		Id:             id,
		Subject:        msg.Subject,
		Issuer:         msg.Issuer,
		CredentialType: msg.CredentialType,
		CredentialHash: msg.CredentialHash,
		Status:         types.CredentialStatus_CREDENTIAL_STATUS_ACTIVE,
		IssuedAt:       issuedAt,
	}
	if err := k.SetCredential(ctx, cred); err != nil {
		return nil, err
	}

	if err := ctx.EventManager().EmitTypedEvent(&types.EventCredentialIssued{
		Id:             cred.Id,
		Subject:        cred.Subject,
		Issuer:         cred.Issuer,
		CredentialType: cred.CredentialType,
	}); err != nil {
		return nil, err
	}

	return &types.MsgIssueCredentialResponse{Id: id}, nil
}

func (k msgServer) RevokeCredential(goCtx context.Context, msg *types.MsgRevokeCredential) (*types.MsgRevokeCredentialResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	cred, found := k.GetCredential(ctx, msg.CredentialId)
	if !found {
		return nil, types.ErrCredentialNotFound
	}
	if cred.Issuer != msg.Issuer && k.GetAuthority() != msg.Issuer {
		return nil, types.ErrNotCredentialIssuer
	}
	if cred.Status == types.CredentialStatus_CREDENTIAL_STATUS_REVOKED {
		return nil, types.ErrCredentialRevoked
	}

	cred.Status = types.CredentialStatus_CREDENTIAL_STATUS_REVOKED
	cred.RevokedAt = ctx.BlockTime().Unix()
	if err := k.SetCredential(ctx, cred); err != nil {
		return nil, err
	}

	if err := ctx.EventManager().EmitTypedEvent(&types.EventCredentialRevoked{
		Id:     cred.Id,
		Issuer: msg.Issuer,
	}); err != nil {
		return nil, err
	}

	return &types.MsgRevokeCredentialResponse{}, nil
}

// PresentCredential is authenticated entirely by tx-signer identity today
// (the presenter must be the credential's subject) — see the ZKVerifier
// TODO in x/identity/types/zk.go for the planned extension point.
func (k msgServer) PresentCredential(goCtx context.Context, msg *types.MsgPresentCredential) (*types.MsgPresentCredentialResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	cred, found := k.GetCredential(ctx, msg.CredentialId)
	if !found {
		return nil, types.ErrCredentialNotFound
	}
	if cred.Subject != msg.Presenter {
		return nil, types.ErrNotCredentialHolder
	}

	valid := cred.Status == types.CredentialStatus_CREDENTIAL_STATUS_ACTIVE

	if err := ctx.EventManager().EmitTypedEvent(&types.EventCredentialPresented{
		Id:        cred.Id,
		Presenter: msg.Presenter,
		Valid:     valid,
	}); err != nil {
		return nil, err
	}

	return &types.MsgPresentCredentialResponse{Valid: valid}, nil
}
