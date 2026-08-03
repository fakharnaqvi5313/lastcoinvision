package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/lastcoinvision/lcv/testutil/keeper"
	identitytypes "github.com/lastcoinvision/lcv/x/identity/types"
	"github.com/lastcoinvision/lcv/x/registry/keeper"
	"github.com/lastcoinvision/lcv/x/registry/types"
)

// callerRole selects which of the three real-world caller categories a
// fuzz case exercises, cycling through the input rather than trying to
// fuzz a raw address string (an arbitrary string is never the owner or
// the authority, which would only ever exercise the "stranger" branch).
type callerRole uint8

const (
	callerIsOwner callerRole = iota
	callerIsAuthority
	callerIsStranger
	numCallerRoles
)

// FuzzMsgTransferAsset fuzzes the four independent preconditions
// TransferAsset must enforce together — frozen status, caller
// authorization, and both parties' attestation — checking that a
// transfer succeeds if and only if all four hold, for every combination.
// This is the invariant a bypass of any single check (e.g. transferring a
// frozen asset, or between an attested and a merely-pseudonymous party)
// would violate; modeled as a fuzz target (rather than a fixed table of
// cases) so the property is checked against however Go's fuzzer chooses
// to combine and mutate the flags, not just the combinations a human
// enumerates by hand.
func FuzzMsgTransferAsset(f *testing.F) {
	for frozen := 0; frozen < 2; frozen++ {
		for role := 0; role < int(numCallerRoles); role++ {
			for ownerAttested := 0; ownerAttested < 2; ownerAttested++ {
				for newOwnerAttested := 0; newOwnerAttested < 2; newOwnerAttested++ {
					f.Add(frozen != 0, uint8(role), ownerAttested != 0, newOwnerAttested != 0)
				}
			}
		}
	}

	f.Fuzz(func(t *testing.T, frozen bool, roleRaw uint8, ownerAttested bool, newOwnerAttested bool) {
		role := callerRole(roleRaw % uint8(numCallerRoles))

		k, ctx, identityKeeper := keepertest.RegistryKeeperWithIdentity(t)
		ms := keeper.NewMsgServerImpl(k)

		if ownerAttested {
			identityKeeper.SetTier(testOwnerAddr, identitytypes.AccountTier_ACCOUNT_TIER_ATTESTED)
		}
		if newOwnerAttested {
			identityKeeper.SetTier(testNewOwnerAddr, identitytypes.AccountTier_ACCOUNT_TIER_ATTESTED)
		}

		_, err := ms.SetJurisdictionAuthority(ctx, &types.MsgSetJurisdictionAuthority{
			Authority:        k.GetAuthority(),
			JurisdictionCode: testJurisdictionCode,
			Address:          testAuthorityAddr,
		})
		require.NoError(t, err)

		registerResp, err := ms.RegisterAsset(ctx, &types.MsgRegisterAsset{
			Creator:          testAuthorityAddr,
			AssetType:        "land_parcel",
			JurisdictionCode: testJurisdictionCode,
			MetadataHash:     "sha256:deedhash",
			Owner:            "did:lcv:" + testOwnerAddr,
		})
		require.NoError(t, err)

		if frozen {
			_, err := ms.FreezeAsset(ctx, &types.MsgFreezeAsset{
				Creator: testAuthorityAddr,
				AssetId: registerResp.Id,
				Reason:  "fuzz-induced freeze",
			})
			require.NoError(t, err)
		}

		var caller string
		switch role {
		case callerIsOwner:
			caller = testOwnerAddr
		case callerIsAuthority:
			caller = testAuthorityAddr
		default:
			caller = "lcv1completestranger00000000000000000000000"
		}

		_, err = ms.TransferAsset(ctx, &types.MsgTransferAsset{
			Creator:  caller,
			AssetId:  registerResp.Id,
			NewOwner: "did:lcv:" + testNewOwnerAddr,
		})

		authorized := role == callerIsOwner || role == callerIsAuthority
		shouldSucceed := !frozen && authorized && ownerAttested && newOwnerAttested

		if shouldSucceed {
			require.NoError(t, err, "expected transfer to succeed: frozen=%v role=%v ownerAttested=%v newOwnerAttested=%v", frozen, role, ownerAttested, newOwnerAttested)
			asset, found := k.GetAsset(ctx, registerResp.Id)
			require.True(t, found)
			require.Equal(t, "did:lcv:"+testNewOwnerAddr, asset.Owner)
		} else {
			require.Error(t, err, "expected transfer to be rejected: frozen=%v role=%v ownerAttested=%v newOwnerAttested=%v", frozen, role, ownerAttested, newOwnerAttested)
			asset, found := k.GetAsset(ctx, registerResp.Id)
			require.True(t, found)
			require.Equal(t, "did:lcv:"+testOwnerAddr, asset.Owner, "a rejected transfer must never change ownership")
		}
	})
}
