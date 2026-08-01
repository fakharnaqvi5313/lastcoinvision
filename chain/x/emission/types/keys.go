package types

import "cosmossdk.io/math"

const (
	// ModuleName defines the module name
	ModuleName = "emission"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_emission"

	// DefaultMintDenom is the denom emitted as staking rewards.
	DefaultMintDenom = "ulcv"

	// StakingRewardsAccountName is the module account name holding the
	// genesis-funded staking rewards pool (12% of total supply). It is a
	// plain holding account: no Minter/Burner permission, since emissions
	// are transfers out of a pre-funded balance, never new minting.
	StakingRewardsAccountName = "staking_rewards"

	// DefaultHalfLifeSeconds is 2 years, chosen so that by year 10 (5
	// half-lives) ~96.9% of the reward pool has been emitted, matching the
	// spec's "10-year decaying emission" description while leaving a long
	// tail for chain longevity beyond that horizon.
	DefaultHalfLifeSeconds = int64(2 * 365 * 24 * 3600)
)

var (
	ParamsKey = []byte("p_emission")

	// LastEmissionTimeKey stores the unix time (seconds) up to which
	// emissions have already been transferred out of the staking rewards
	// pool, so BeginBlock can compute exactly the amount due since then.
	LastEmissionTimeKey = []byte("last_emission_time")

	// DefaultTotalRewardPool is 12% of the fixed 10,000,000,000 LCV total
	// supply, expressed in ulcv (1 LCV = 1,000,000 ulcv).
	DefaultTotalRewardPool = math.NewInt(1_200_000_000).MulRaw(1_000_000)
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
