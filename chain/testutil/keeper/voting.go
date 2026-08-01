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

	"github.com/lastcoinvision/lcv/x/voting/keeper"
	"github.com/lastcoinvision/lcv/x/voting/types"
)

// FakeVotingIdentityKeeper is a minimal in-memory stand-in for the real
// x/identity keeper, just enough to exercise x/voting's eligibility gating
// in unit tests without pulling in the full identity module.
type FakeVotingIdentityKeeper struct {
	credentials map[string]map[string]bool // address -> credential_type -> active
}

func NewFakeVotingIdentityKeeper() *FakeVotingIdentityKeeper {
	return &FakeVotingIdentityKeeper{credentials: map[string]map[string]bool{}}
}

// GrantCredential marks address as holding an active credential of
// credentialType, as if x/identity had issued and authorized it.
func (f *FakeVotingIdentityKeeper) GrantCredential(address, credentialType string) {
	if f.credentials[address] == nil {
		f.credentials[address] = map[string]bool{}
	}
	f.credentials[address][credentialType] = true
}

func (f *FakeVotingIdentityKeeper) HasActiveCredentialOfType(_ context.Context, address, credentialType string) bool {
	return f.credentials[address][credentialType]
}

var _ types.IdentityKeeper = (*FakeVotingIdentityKeeper)(nil)

func VotingKeeper(t testing.TB) (keeper.Keeper, sdk.Context) {
	k, ctx, _ := VotingKeeperWithIdentity(t)
	return k, ctx
}

// VotingKeeperWithIdentity is like VotingKeeper but also returns the fake
// identity keeper wired in, so tests can grant eligibility via
// GrantCredential.
func VotingKeeperWithIdentity(t testing.TB) (keeper.Keeper, sdk.Context, *FakeVotingIdentityKeeper) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)
	identityKeeper := NewFakeVotingIdentityKeeper()

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
