package types

import "cosmossdk.io/math"

const (
	// ModuleName defines the module name. Named "lcvfeemarket", not just
	// "feemarket", because cosmos/evm ships its own module also named
	// "feemarket" (its EVM-side EIP-1559 base fee market) — module names
	// must be unique per app (they key the module manager, store, and
	// genesis maps), and that name is upstream/fixed. See DECISIONS.md.
	ModuleName = "lcvfeemarket"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_feemarket"

	// DefaultFeeDenom is the denom the dynamic base fee is charged in.
	DefaultFeeDenom = "ulcv"

	// DefaultTargetGas is the per-block gas usage the fee market targets.
	DefaultTargetGas = int64(30_000_000)

	// DefaultMaxBaseFeeChangeDenominator caps base_fee movement at 1/8
	// (12.5%) per block, matching Ethereum's EIP-1559 default.
	DefaultMaxBaseFeeChangeDenominator = int64(8)
)

var (
	ParamsKey = []byte("p_feemarket")

	// DefaultMinBaseFee is the floor base_fee can never adjust below:
	// 0.0001 ulcv per unit of gas.
	DefaultMinBaseFee = math.LegacyNewDecWithPrec(1, 4)

	// DefaultBaseFee starts at the floor; it adjusts from there based on
	// actual network usage from genesis onward.
	DefaultBaseFee = DefaultMinBaseFee

	// DefaultBurnPercentage: 30% of the base fee portion of every
	// transaction fee is burned, 70% goes to the fee collector for
	// validator/community-pool distribution, per the top-level spec.
	DefaultBurnPercentage = math.LegacyNewDecWithPrec(30, 2)
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
