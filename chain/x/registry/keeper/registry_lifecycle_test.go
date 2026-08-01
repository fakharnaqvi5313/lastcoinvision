package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/lastcoinvision/lcv/testutil/keeper"
	identitytypes "github.com/lastcoinvision/lcv/x/identity/types"
	"github.com/lastcoinvision/lcv/x/registry/keeper"
	"github.com/lastcoinvision/lcv/x/registry/types"
)

const (
	testJurisdictionCode = "us-ca"
	testAuthorityAddr    = "lcv1jurisdictionauthority000000000000000000"
	testOwnerAddr        = "lcv1initialowner00000000000000000000000000"
	testNewOwnerAddr     = "lcv1newowner000000000000000000000000000000"
)

// TestAcceptance_AssetLifecycle is the Phase C acceptance flow for
// x/registry: jurisdiction authority set via gov -> asset registered ->
// transfer between two attested parties succeeds and is recorded in
// history -> freeze blocks further transfer.
func TestAcceptance_AssetLifecycle(t *testing.T) {
	k, ctx, identityKeeper := keepertest.RegistryKeeperWithIdentity(t)
	ms := keeper.NewMsgServerImpl(k)

	identityKeeper.SetTier(testOwnerAddr, identitytypes.AccountTier_ACCOUNT_TIER_ATTESTED)
	identityKeeper.SetTier(testNewOwnerAddr, identitytypes.AccountTier_ACCOUNT_TIER_ATTESTED)

	// "jurisdiction authority set via gov"
	_, err := ms.SetJurisdictionAuthority(ctx, &types.MsgSetJurisdictionAuthority{
		Authority:        k.GetAuthority(),
		JurisdictionCode: testJurisdictionCode,
		Address:          testAuthorityAddr,
	})
	require.NoError(t, err)

	// A non-authority signer cannot set itself as the jurisdiction authority.
	_, err = ms.SetJurisdictionAuthority(ctx, &types.MsgSetJurisdictionAuthority{
		Authority:        testOwnerAddr,
		JurisdictionCode: testJurisdictionCode,
		Address:          testOwnerAddr,
	})
	require.ErrorIs(t, err, types.ErrInvalidSigner)

	// "asset registered" — only the jurisdiction authority may register.
	_, err = ms.RegisterAsset(ctx, &types.MsgRegisterAsset{
		Creator:          testOwnerAddr,
		AssetType:        "land_parcel",
		JurisdictionCode: testJurisdictionCode,
		MetadataHash:     "sha256:deedhash",
		Owner:            "did:lcv:" + testOwnerAddr,
	})
	require.ErrorIs(t, err, types.ErrNotJurisdictionAuthority)

	registerResp, err := ms.RegisterAsset(ctx, &types.MsgRegisterAsset{
		Creator:          testAuthorityAddr,
		AssetType:        "land_parcel",
		JurisdictionCode: testJurisdictionCode,
		MetadataHash:     "sha256:deedhash",
		Owner:            "did:lcv:" + testOwnerAddr,
	})
	require.NoError(t, err)
	require.NotEmpty(t, registerResp.Id)

	asset, found := k.GetAsset(ctx, registerResp.Id)
	require.True(t, found)
	require.Equal(t, types.AssetStatus_ASSET_STATUS_ACTIVE, asset.Status)

	// "transfer between two attested parties succeeds"
	_, err = ms.TransferAsset(ctx, &types.MsgTransferAsset{
		Creator:  testOwnerAddr,
		AssetId:  registerResp.Id,
		NewOwner: "did:lcv:" + testNewOwnerAddr,
	})
	require.NoError(t, err)

	asset, found = k.GetAsset(ctx, registerResp.Id)
	require.True(t, found)
	require.Equal(t, "did:lcv:"+testNewOwnerAddr, asset.Owner)

	// "...and is recorded in history"
	history := k.GetTransferHistory(ctx, registerResp.Id)
	require.Len(t, history, 1)
	require.Equal(t, "did:lcv:"+testOwnerAddr, history[0].From)
	require.Equal(t, "did:lcv:"+testNewOwnerAddr, history[0].To)

	// "freeze blocks further transfer"
	_, err = ms.FreezeAsset(ctx, &types.MsgFreezeAsset{
		Creator: testAuthorityAddr,
		AssetId: registerResp.Id,
		Reason:  "suspected fraud investigation",
	})
	require.NoError(t, err)

	asset, found = k.GetAsset(ctx, registerResp.Id)
	require.True(t, found)
	require.Equal(t, types.AssetStatus_ASSET_STATUS_FROZEN, asset.Status)
	require.Equal(t, "suspected fraud investigation", asset.FreezeReason)

	_, err = ms.TransferAsset(ctx, &types.MsgTransferAsset{
		Creator:  testNewOwnerAddr,
		AssetId:  registerResp.Id,
		NewOwner: "did:lcv:" + testOwnerAddr,
	})
	require.ErrorIs(t, err, types.ErrAssetFrozen)
}

func TestTransferRequiresBothPartiesAttested(t *testing.T) {
	k, ctx, identityKeeper := keepertest.RegistryKeeperWithIdentity(t)
	ms := keeper.NewMsgServerImpl(k)

	identityKeeper.SetTier(testOwnerAddr, identitytypes.AccountTier_ACCOUNT_TIER_ATTESTED)
	// testNewOwnerAddr deliberately left at the pseudonymous default.

	require.NoError(t, k.SetJurisdictionAuthority(ctx, types.JurisdictionAuthority{
		JurisdictionCode: testJurisdictionCode,
		Address:          testAuthorityAddr,
	}))
	registerResp, err := ms.RegisterAsset(ctx, &types.MsgRegisterAsset{
		Creator:          testAuthorityAddr,
		AssetType:        "vehicle",
		JurisdictionCode: testJurisdictionCode,
		MetadataHash:     "sha256:vinhash",
		Owner:            "did:lcv:" + testOwnerAddr,
	})
	require.NoError(t, err)

	_, err = ms.TransferAsset(ctx, &types.MsgTransferAsset{
		Creator:  testOwnerAddr,
		AssetId:  registerResp.Id,
		NewOwner: "did:lcv:" + testNewOwnerAddr,
	})
	require.ErrorIs(t, err, types.ErrPartyNotAttested)
}

func TestOnlyJurisdictionAuthorityMayFreeze(t *testing.T) {
	k, ctx, identityKeeper := keepertest.RegistryKeeperWithIdentity(t)
	ms := keeper.NewMsgServerImpl(k)
	identityKeeper.SetTier(testOwnerAddr, identitytypes.AccountTier_ACCOUNT_TIER_ATTESTED)

	require.NoError(t, k.SetJurisdictionAuthority(ctx, types.JurisdictionAuthority{
		JurisdictionCode: testJurisdictionCode,
		Address:          testAuthorityAddr,
	}))
	registerResp, err := ms.RegisterAsset(ctx, &types.MsgRegisterAsset{
		Creator:          testAuthorityAddr,
		AssetType:        "vehicle",
		JurisdictionCode: testJurisdictionCode,
		MetadataHash:     "sha256:vinhash",
		Owner:            "did:lcv:" + testOwnerAddr,
	})
	require.NoError(t, err)

	_, err = ms.FreezeAsset(ctx, &types.MsgFreezeAsset{
		Creator: testOwnerAddr,
		AssetId: registerResp.Id,
		Reason:  "self-freeze attempt",
	})
	require.ErrorIs(t, err, types.ErrNotJurisdictionAuthority)
}

func TestTransferHistoryReferentialIntegrityInvariant(t *testing.T) {
	k, ctx := keepertest.RegistryKeeper(t)

	msg, broken := keeper.TransferHistoryReferentialIntegrityInvariant(k)(ctx)
	require.False(t, broken)
	require.Empty(t, msg)

	require.NoError(t, k.AppendTransferRecord(ctx, types.TransferRecord{
		AssetId: "nonexistent-asset",
		From:    "did:lcv:lcv1a",
		To:      "did:lcv:lcv1b",
	}))

	msg, broken = keeper.TransferHistoryReferentialIntegrityInvariant(k)(ctx)
	require.True(t, broken)
	require.Contains(t, msg, "nonexistent-asset")
}
