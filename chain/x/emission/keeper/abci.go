package keeper

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/lastcoinvision/lcv/x/emission/types"
)

// GetLastEmissionTime returns the unix time (seconds) up to which staking
// rewards have already been transferred out of the pool. Before genesis
// InitGenesis runs this is unset; callers should treat that as "no time has
// elapsed yet" by falling back to params.StartTime.
func (k Keeper) GetLastEmissionTime(ctx sdk.Context) (int64, bool) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.LastEmissionTimeKey)
	if bz == nil {
		return 0, false
	}
	return int64(binary.BigEndian.Uint64(bz)), true
}

// SetLastEmissionTime records the unix time (seconds) up to which staking
// rewards have been transferred out of the pool.
func (k Keeper) SetLastEmissionTime(ctx sdk.Context, unixTime int64) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(unixTime))
	store.Set(types.LastEmissionTimeKey, bz)
}

// EmitBlockRewards transfers the amount due under the exponential decay
// curve for the time elapsed since the last block from the staking rewards
// pool module account into the fee collector, where x/distribution's own
// BeginBlocker (which must run after this one, see app_config.go) allocates
// it to the proposer/validators/community pool alongside collected fees.
//
// This never mints: it only ever moves coins that were pre-funded into the
// staking rewards account at genesis, so a depleted pool simply means no
// further rewards are emitted rather than any invariant violation.
func (k Keeper) EmitBlockRewards(ctx sdk.Context) error {
	params := k.GetParams(ctx)
	if params.HalfLifeSeconds == 0 {
		// Params not initialized yet (e.g. genesis InitGenesis hasn't run
		// for this module in a chain that predates it) — nothing to do.
		return nil
	}

	now := ctx.BlockTime().Unix()
	lastEmission, ok := k.GetLastEmissionTime(ctx)
	if !ok {
		lastEmission = params.StartTime
	}
	if now <= lastEmission {
		return nil
	}

	elapsedFrom := lastEmission - params.StartTime
	elapsedTo := now - params.StartTime
	if elapsedFrom < 0 {
		elapsedFrom = 0
	}

	due := emittedBetween(params.TotalRewardPool, params.HalfLifeSeconds, elapsedFrom, elapsedTo)
	if due.IsPositive() {
		available := k.bankKeeper.GetBalance(ctx, authtypes.NewModuleAddress(params.StakingRewardsAccount), params.MintDenom)
		if due.GT(available.Amount) {
			due = available.Amount
		}
	}

	k.SetLastEmissionTime(ctx, now)

	if !due.IsPositive() {
		return nil
	}

	coins := sdk.NewCoins(sdk.NewCoin(params.MintDenom, due))
	return k.bankKeeper.SendCoinsFromModuleToModule(ctx, params.StakingRewardsAccount, authtypes.FeeCollectorName, coins)
}
