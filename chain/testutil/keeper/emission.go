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

	"github.com/lastcoinvision/lcv/x/emission/keeper"
	"github.com/lastcoinvision/lcv/x/emission/types"
)

// fakeEmissionBankKeeper is a minimal in-memory stand-in for the real bank
// keeper, just enough to exercise x/emission's decay-transfer logic in unit
// tests without pulling in the full bank module.
type fakeEmissionBankKeeper struct {
	balances map[string]sdk.Coins
}

func newFakeEmissionBankKeeper() *fakeEmissionBankKeeper {
	return &fakeEmissionBankKeeper{balances: map[string]sdk.Coins{}}
}

func (f *fakeEmissionBankKeeper) SpendableCoins(_ context.Context, addr sdk.AccAddress) sdk.Coins {
	return f.balances[addr.String()]
}

func (f *fakeEmissionBankKeeper) GetBalance(_ context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return sdk.NewCoin(denom, f.balances[addr.String()].AmountOf(denom))
}

func (f *fakeEmissionBankKeeper) SendCoinsFromModuleToModule(_ context.Context, senderModule, recipientModule string, amt sdk.Coins) error {
	sender := authtypes.NewModuleAddress(senderModule).String()
	recipient := authtypes.NewModuleAddress(recipientModule).String()
	f.balances[sender] = f.balances[sender].Sub(amt...)
	f.balances[recipient] = f.balances[recipient].Add(amt...)
	return nil
}

func (f *fakeEmissionBankKeeper) FundModuleAccount(moduleName string, amt sdk.Coins) {
	addr := authtypes.NewModuleAddress(moduleName).String()
	f.balances[addr] = f.balances[addr].Add(amt...)
}

func EmissionKeeper(t testing.TB) (keeper.Keeper, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)

	bankKeeper := newFakeEmissionBankKeeper()
	bankKeeper.FundModuleAccount(types.StakingRewardsAccountName, sdk.NewCoins(sdk.NewCoin(types.DefaultMintDenom, types.DefaultTotalRewardPool)))

	k := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(storeKey),
		log.NewNopLogger(),
		authority.String(),
		bankKeeper,
	)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	if err := k.SetParams(ctx, types.DefaultParams()); err != nil {
		panic(err)
	}

	return k, ctx
}
