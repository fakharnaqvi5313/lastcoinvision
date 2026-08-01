package feemarket

import (
	"math/rand"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/lastcoinvision/lcv/testutil/sample"
	feemarketsimulation "github.com/lastcoinvision/lcv/x/feemarket/simulation"
	"github.com/lastcoinvision/lcv/x/feemarket/types"
)

// avoid unused import issue
var (
	_ = feemarketsimulation.FindAccount
	_ = rand.Rand{}
	_ = sample.AccAddress
	_ = sdk.AccAddress{}
	_ = simulation.MsgEntryKind
)

const (
// this line is used by starport scaffolding # simapp/module/const
)

// GenerateGenesisState creates a randomized GenState of the module.
//
// The mandatory base-fee ante decorator (x/feemarket/ante) enforces
// base_fee*gasWanted in every delivered tx, not just at CheckTx/mempool
// time — that's the point of an EIP-1559-style base fee. But the SDK's
// generic simulation operations (used by x/staking, x/bank, x/gov, etc. —
// none of which know about our custom fee market) generate fees via
// simulation.RandomFees with a fixed 10,000,000 DefaultGenTxGas, which has
// no way to target our DefaultBaseFee. At the real DefaultBaseFee that
// combination requires 1000ulcv per tx that RandomFees has no reason to
// clear, so the full-app block simulator fails almost immediately with
// "insufficient fee" on ordinary staking/bank/gov operations, independent
// of anything this chain's own custom modules do.
//
// This is a simulation-genesis-only params override — it does not touch
// types.DefaultParams(), which is what a real chain's genesis actually
// uses. A near-zero (but positive) base fee was tried first, but many
// generic SDK simulation operations run against freshly-created accounts
// with zero spendable balance and rely on being able to submit a genuinely
// zero-fee tx (simulation.RandomFees returns `nil, nil` — not an error —
// for an account with no spendable coins at all); any positive floor still
// rejects those. So the simulation genesis sets both BaseFee and
// MinBaseFee to exactly zero (allowed by Validate(), which only requires
// MinBaseFee to be non-negative) purely so the ante decorator's
// `paid.LT(required.Amount)` check can never fail here. This matches how
// the SDK's own staking/slashing/gov modules already randomize away from
// their mainnet defaults for this same harness (see the "Selected randomly
// generated ... parameters" log lines any simulation run prints).
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	feemarketGenesis := types.GenesisState{
		Params: types.NewParams(
			types.DefaultFeeDenom,
			math.LegacyZeroDec(),
			math.LegacyZeroDec(),
			types.DefaultTargetGas,
			types.DefaultMaxBaseFeeChangeDenominator,
			types.DefaultBurnPercentage,
		),
		// this line is used by starport scaffolding # simapp/module/genesisState
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&feemarketGenesis)
}

// RegisterStoreDecoder registers a decoder.
func (am AppModule) RegisterStoreDecoder(_ simtypes.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)

	// this line is used by starport scaffolding # simapp/module/operation

	return operations
}

// ProposalMsgs returns msgs used for governance proposals for simulations.
func (am AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{
		// this line is used by starport scaffolding # simapp/module/OpMsg
	}
}
