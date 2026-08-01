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

	"github.com/lastcoinvision/lcv/x/taxrail/keeper"
	"github.com/lastcoinvision/lcv/x/taxrail/types"
)

// FakeTaxrailBankKeeper is a minimal in-memory stand-in for the real bank
// keeper, just enough to exercise x/taxrail's split-payment logic in unit
// tests without pulling in the full bank module.
type FakeTaxrailBankKeeper struct {
	balances map[string]sdk.Coins
}

func NewFakeTaxrailBankKeeper() *FakeTaxrailBankKeeper {
	return &FakeTaxrailBankKeeper{balances: map[string]sdk.Coins{}}
}

func (f *FakeTaxrailBankKeeper) FundAccount(addr sdk.AccAddress, amt sdk.Coins) {
	f.balances[addr.String()] = f.balances[addr.String()].Add(amt...)
}

func (f *FakeTaxrailBankKeeper) SpendableCoins(_ context.Context, addr sdk.AccAddress) sdk.Coins {
	return f.balances[addr.String()]
}

func (f *FakeTaxrailBankKeeper) GetBalance(_ context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return sdk.NewCoin(denom, f.balances[addr.String()].AmountOf(denom))
}

func (f *FakeTaxrailBankKeeper) SendCoins(_ context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error {
	f.balances[fromAddr.String()] = f.balances[fromAddr.String()].Sub(amt...)
	f.balances[toAddr.String()] = f.balances[toAddr.String()].Add(amt...)
	return nil
}

func (f *FakeTaxrailBankKeeper) SendCoinsFromAccountToModule(_ context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	recipient := authtypes.NewModuleAddress(recipientModule).String()
	f.balances[senderAddr.String()] = f.balances[senderAddr.String()].Sub(amt...)
	f.balances[recipient] = f.balances[recipient].Add(amt...)
	return nil
}

func (f *FakeTaxrailBankKeeper) SendCoinsFromModuleToAccount(_ context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	sender := authtypes.NewModuleAddress(senderModule).String()
	f.balances[sender] = f.balances[sender].Sub(amt...)
	f.balances[recipientAddr.String()] = f.balances[recipientAddr.String()].Add(amt...)
	return nil
}

func TaxrailKeeper(t testing.TB) (keeper.Keeper, sdk.Context) {
	k, ctx, _ := TaxrailKeeperWithBank(t)
	return k, ctx
}

func TaxrailKeeperWithBank(t testing.TB) (keeper.Keeper, sdk.Context, *FakeTaxrailBankKeeper) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)

	bankKeeper := NewFakeTaxrailBankKeeper()

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

	return k, ctx, bankKeeper
}
