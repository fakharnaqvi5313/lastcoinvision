package keeper

import (
	"context"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"

	identitytypes "github.com/lastcoinvision/lcv/x/identity/types"
	"github.com/lastcoinvision/lcv/x/registry/keeper"
	"github.com/lastcoinvision/lcv/x/registry/types"
)

// FakeIdentityKeeper is a minimal in-memory stand-in for the real x/identity
// keeper, just enough to exercise x/registry's attested-tier gating in unit
// tests without pulling in the full identity module. Every address defaults
// to pseudonymous unless set otherwise.
type FakeIdentityKeeper struct {
	tiers map[string]identitytypes.AccountTier
}

func NewFakeIdentityKeeper() *FakeIdentityKeeper {
	return &FakeIdentityKeeper{tiers: map[string]identitytypes.AccountTier{}}
}

func (f *FakeIdentityKeeper) SetTier(address string, tier identitytypes.AccountTier) {
	f.tiers[address] = tier
}

func (f *FakeIdentityKeeper) ComputeTier(_ context.Context, address string) (identitytypes.AccountTier, string) {
	if tier, ok := f.tiers[address]; ok {
		return tier, ""
	}
	return identitytypes.AccountTier_ACCOUNT_TIER_PSEUDONYMOUS, ""
}

var _ types.IdentityKeeper = (*FakeIdentityKeeper)(nil)

func RegistryKeeper(t testing.TB) (keeper.Keeper, sdk.Context) {
	k, ctx, _ := RegistryKeeperWithIdentity(t)
	return k, ctx
}

// RegistryKeeperWithIdentity is like RegistryKeeper but also returns the
// fake identity keeper wired in, so tests can mark addresses attested or
// institutional via SetTier.
func RegistryKeeperWithIdentity(t testing.TB) (keeper.Keeper, sdk.Context, *FakeIdentityKeeper) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)
	identityKeeper := NewFakeIdentityKeeper()

	k := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(storeKey),
		log.NewNopLogger(),
		authority.String(),
		identityKeeper,
	)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	if err := k.SetParams(ctx, types.DefaultParams()); err != nil {
		panic(err)
	}

	return k, ctx, identityKeeper
}
