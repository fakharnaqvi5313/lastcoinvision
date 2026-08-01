package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/lastcoinvision/lcv/testutil/keeper"
	"github.com/lastcoinvision/lcv/x/identity/keeper"
	"github.com/lastcoinvision/lcv/x/identity/types"
)

const (
	testIssuerAddr  = "lcv1issuer00000000000000000000000000000000"
	testSubjectAddr = "lcv1subject0000000000000000000000000000000"
)

// TestAcceptance_IssuerCredentialTierLifecycle is the Phase B acceptance
// test from the top-level spec: issuer added via gov -> issues credential
// to account -> account tier becomes attested -> revocation flips status ->
// tier reverts.
func TestAcceptance_IssuerCredentialTierLifecycle(t *testing.T) {
	k, ctx := keepertest.IdentityKeeper(t)
	ms := keeper.NewMsgServerImpl(k)

	// Before anything, the account is pseudonymous.
	tier, _ := k.ComputeTier(ctx, testSubjectAddr)
	require.Equal(t, types.AccountTier_ACCOUNT_TIER_PSEUDONYMOUS, tier)

	// "issuer added via gov": AddIssuer is authority-gated, i.e. only
	// executable via a passed governance proposal in production.
	_, err := ms.AddIssuer(ctx, &types.MsgAddIssuer{
		Authority:     k.GetAuthority(),
		IssuerAddress: testIssuerAddr,
		Name:          "Test KYC Provider",
	})
	require.NoError(t, err)
	require.True(t, k.IsAuthorizedIssuer(ctx, testIssuerAddr))

	// A non-authority signer must not be able to add itself as an issuer.
	_, err = ms.AddIssuer(ctx, &types.MsgAddIssuer{
		Authority:     testSubjectAddr,
		IssuerAddress: testSubjectAddr,
		Name:          "Self-appointed",
	})
	require.ErrorIs(t, err, types.ErrInvalidSigner)

	// "issues credential to account"
	issueResp, err := ms.IssueCredential(ctx, &types.MsgIssueCredential{
		Issuer:         testIssuerAddr,
		Subject:        testSubjectAddr,
		CredentialType: types.DefaultAttestedCredentialType,
		CredentialHash: "deadbeef",
	})
	require.NoError(t, err)
	require.NotEmpty(t, issueResp.Id)

	cred, found := k.GetCredential(ctx, issueResp.Id)
	require.True(t, found)
	require.Equal(t, types.CredentialStatus_CREDENTIAL_STATUS_ACTIVE, cred.Status)

	// "account tier becomes attested"
	tier, roleMetadata := k.ComputeTier(ctx, testSubjectAddr)
	require.Equal(t, types.AccountTier_ACCOUNT_TIER_ATTESTED, tier)
	require.Empty(t, roleMetadata)

	// An unauthorized address may not revoke.
	_, err = ms.RevokeCredential(ctx, &types.MsgRevokeCredential{
		Issuer:       testSubjectAddr,
		CredentialId: issueResp.Id,
	})
	require.ErrorIs(t, err, types.ErrNotCredentialIssuer)

	// "revocation flips status"
	_, err = ms.RevokeCredential(ctx, &types.MsgRevokeCredential{
		Issuer:       testIssuerAddr,
		CredentialId: issueResp.Id,
	})
	require.NoError(t, err)

	cred, found = k.GetCredential(ctx, issueResp.Id)
	require.True(t, found)
	require.Equal(t, types.CredentialStatus_CREDENTIAL_STATUS_REVOKED, cred.Status)

	// Revoking twice is rejected.
	_, err = ms.RevokeCredential(ctx, &types.MsgRevokeCredential{
		Issuer:       testIssuerAddr,
		CredentialId: issueResp.Id,
	})
	require.ErrorIs(t, err, types.ErrCredentialRevoked)

	// "tier reverts"
	tier, _ = k.ComputeTier(ctx, testSubjectAddr)
	require.Equal(t, types.AccountTier_ACCOUNT_TIER_PSEUDONYMOUS, tier)
}

// TestAcceptance_RemovedIssuerCredentialsNoLongerGrantAttested checks the
// documented invariant that removing an issuer doesn't retroactively
// invalidate a credential record, but DOES stop it from granting the
// attested tier, since TierOf re-checks issuer authorization live.
func TestIssuerRemovalRevokesAttestedTier(t *testing.T) {
	k, ctx := keepertest.IdentityKeeper(t)
	ms := keeper.NewMsgServerImpl(k)

	_, err := ms.AddIssuer(ctx, &types.MsgAddIssuer{
		Authority:     k.GetAuthority(),
		IssuerAddress: testIssuerAddr,
		Name:          "Test KYC Provider",
	})
	require.NoError(t, err)

	issueResp, err := ms.IssueCredential(ctx, &types.MsgIssueCredential{
		Issuer:         testIssuerAddr,
		Subject:        testSubjectAddr,
		CredentialType: types.DefaultAttestedCredentialType,
		CredentialHash: "deadbeef",
	})
	require.NoError(t, err)

	tier, _ := k.ComputeTier(ctx, testSubjectAddr)
	require.Equal(t, types.AccountTier_ACCOUNT_TIER_ATTESTED, tier)

	_, err = ms.RemoveIssuer(ctx, &types.MsgRemoveIssuer{
		Authority:     k.GetAuthority(),
		IssuerAddress: testIssuerAddr,
	})
	require.NoError(t, err)

	// The credential record itself is untouched...
	cred, found := k.GetCredential(ctx, issueResp.Id)
	require.True(t, found)
	require.Equal(t, types.CredentialStatus_CREDENTIAL_STATUS_ACTIVE, cred.Status)

	// ...but it no longer grants the attested tier.
	tier, _ = k.ComputeTier(ctx, testSubjectAddr)
	require.Equal(t, types.AccountTier_ACCOUNT_TIER_PSEUDONYMOUS, tier)
}

func TestInstitutionalTierTakesPrecedenceOverAttested(t *testing.T) {
	k, ctx, groupKeeper := keepertest.IdentityKeeperWithGroup(t)
	ms := keeper.NewMsgServerImpl(k)

	_, err := ms.AddIssuer(ctx, &types.MsgAddIssuer{
		Authority:     k.GetAuthority(),
		IssuerAddress: testIssuerAddr,
		Name:          "Test KYC Provider",
	})
	require.NoError(t, err)
	_, err = ms.IssueCredential(ctx, &types.MsgIssueCredential{
		Issuer:         testIssuerAddr,
		Subject:        testSubjectAddr,
		CredentialType: types.DefaultAttestedCredentialType,
		CredentialHash: "deadbeef",
	})
	require.NoError(t, err)

	tier, _ := k.ComputeTier(ctx, testSubjectAddr)
	require.Equal(t, types.AccountTier_ACCOUNT_TIER_ATTESTED, tier, "sanity check before adding group policy")

	groupKeeper.SetGroupPolicy(testSubjectAddr, `{"role":"ministry_of_finance"}`)

	tier, roleMetadata := k.ComputeTier(ctx, testSubjectAddr)
	require.Equal(t, types.AccountTier_ACCOUNT_TIER_INSTITUTIONAL, tier)
	require.Equal(t, `{"role":"ministry_of_finance"}`, roleMetadata)
}

func TestDIDLifecycle(t *testing.T) {
	k, ctx := keepertest.IdentityKeeper(t)
	ms := keeper.NewMsgServerImpl(k)

	createResp, err := ms.CreateDID(ctx, &types.MsgCreateDID{Creator: testSubjectAddr})
	require.NoError(t, err)
	require.Equal(t, "did:lcv:"+testSubjectAddr, createResp.Id)

	doc, found := k.GetDID(ctx, testSubjectAddr)
	require.True(t, found)
	require.False(t, doc.Deactivated)

	// Can't create twice.
	_, err = ms.CreateDID(ctx, &types.MsgCreateDID{Creator: testSubjectAddr})
	require.ErrorIs(t, err, types.ErrDIDAlreadyExists)

	_, err = ms.UpdateDID(ctx, &types.MsgUpdateDID{
		Creator: testSubjectAddr,
		Service: []*types.ServiceEndpoint{{Id: "svc-1", Type: "MessagingService", ServiceEndpoint: "https://example.invalid"}},
	})
	require.NoError(t, err)

	doc, found = k.GetDID(ctx, testSubjectAddr)
	require.True(t, found)
	require.Len(t, doc.Service, 1)

	_, err = ms.DeactivateDID(ctx, &types.MsgDeactivateDID{Creator: testSubjectAddr})
	require.NoError(t, err)

	doc, found = k.GetDID(ctx, testSubjectAddr)
	require.True(t, found)
	require.True(t, doc.Deactivated)

	// A deactivated DID cannot be updated or deactivated again.
	_, err = ms.UpdateDID(ctx, &types.MsgUpdateDID{Creator: testSubjectAddr})
	require.ErrorIs(t, err, types.ErrDIDDeactivated)
	_, err = ms.DeactivateDID(ctx, &types.MsgDeactivateDID{Creator: testSubjectAddr})
	require.ErrorIs(t, err, types.ErrDIDDeactivated)
}

func TestPresentCredential(t *testing.T) {
	k, ctx := keepertest.IdentityKeeper(t)
	ms := keeper.NewMsgServerImpl(k)

	_, err := ms.AddIssuer(ctx, &types.MsgAddIssuer{
		Authority:     k.GetAuthority(),
		IssuerAddress: testIssuerAddr,
		Name:          "Test KYC Provider",
	})
	require.NoError(t, err)
	issueResp, err := ms.IssueCredential(ctx, &types.MsgIssueCredential{
		Issuer:         testIssuerAddr,
		Subject:        testSubjectAddr,
		CredentialType: types.DefaultAttestedCredentialType,
		CredentialHash: "deadbeef",
	})
	require.NoError(t, err)

	// Wrong presenter is rejected outright.
	_, err = ms.PresentCredential(ctx, &types.MsgPresentCredential{
		Presenter:    testIssuerAddr,
		CredentialId: issueResp.Id,
	})
	require.ErrorIs(t, err, types.ErrNotCredentialHolder)

	presentResp, err := ms.PresentCredential(ctx, &types.MsgPresentCredential{
		Presenter:    testSubjectAddr,
		CredentialId: issueResp.Id,
	})
	require.NoError(t, err)
	require.True(t, presentResp.Valid)

	_, err = ms.RevokeCredential(ctx, &types.MsgRevokeCredential{
		Issuer:       testIssuerAddr,
		CredentialId: issueResp.Id,
	})
	require.NoError(t, err)

	presentResp, err = ms.PresentCredential(ctx, &types.MsgPresentCredential{
		Presenter:    testSubjectAddr,
		CredentialId: issueResp.Id,
	})
	require.NoError(t, err, "presenting a revoked credential is not an error, it's a valid=false result")
	require.False(t, presentResp.Valid)
}
