package keeper

import (
	"context"
	"fmt"
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
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/stretchr/testify/require"

	"github.com/lastcoinvision/lcv/x/identity/keeper"
	"github.com/lastcoinvision/lcv/x/identity/types"
)

// FakeGroupKeeper is a minimal in-memory stand-in for the real x/group
// keeper, just enough to exercise x/identity's institutional-tier detection
// in unit tests without pulling in the full group module.
type FakeGroupKeeper struct {
	policies map[string]*group.GroupPolicyInfo
}

func NewFakeGroupKeeper() *FakeGroupKeeper {
	return &FakeGroupKeeper{policies: map[string]*group.GroupPolicyInfo{}}
}

// SetGroupPolicy registers address as a group-policy account with the given
// metadata, as if it had been created via a real MsgCreateGroupPolicy.
func (f *FakeGroupKeeper) SetGroupPolicy(address, metadata string) {
	f.policies[address] = &group.GroupPolicyInfo{Address: address, Metadata: metadata}
}

func (f *FakeGroupKeeper) GroupPolicyInfo(_ context.Context, req *group.QueryGroupPolicyInfoRequest) (*group.QueryGroupPolicyInfoResponse, error) {
	info, ok := f.policies[req.Address]
	if !ok {
		return nil, fmt.Errorf("group policy not found for address %s", req.Address)
	}
	return &group.QueryGroupPolicyInfoResponse{Info: info}, nil
}

var _ types.GroupKeeper = (*FakeGroupKeeper)(nil)

func IdentityKeeper(t testing.TB) (keeper.Keeper, sdk.Context) {
	k, ctx, _ := IdentityKeeperWithGroup(t)
	return k, ctx
}

// IdentityKeeperWithGroup is like IdentityKeeper but also returns the fake
// group keeper wired in, so tests can register institutional accounts via
// SetGroupPolicy.
func IdentityKeeperWithGroup(t testing.TB) (keeper.Keeper, sdk.Context, *FakeGroupKeeper) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)
	groupKeeper := NewFakeGroupKeeper()

	k := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(storeKey),
		log.NewNopLogger(),
		authority.String(),
		groupKeeper,
	)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	if err := k.SetParams(ctx, types.DefaultParams()); err != nil {
		panic(err)
	}

	return k, ctx, groupKeeper
}
