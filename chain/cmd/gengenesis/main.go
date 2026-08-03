// Command gengenesis patches an existing genesis.json (as produced by
// `lcvd multi-node` or `lcvd init` + `lcvd genesis gentx` + `lcvd genesis
// collect-gentxs`) with LastcoinVision's fixed token allocation table:
// ecosystem/grants, foundation treasury, team (vesting), investors
// (vesting), the staking-rewards module account, public sale, and
// community/airdrop — plus pinning x/mint to zero inflation, since all
// staking rewards come from x/emission's pre-funded decay curve instead.
//
// It never invents tokenomics numbers of its own: the percentages below are
// exactly the ones from the top-level spec (25/20/15/15/12/8/5, summing to
// 100% of the fixed 10,000,000,000 LCV supply) and are intentionally not
// exposed as flags, so a typo on the command line can't silently change
// what governance and the public were told the allocation is.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	// Blank import for its init() side effect: sets the "lcv" bech32
	// prefixes globally, exactly as the lcvd binary itself does.
	_ "github.com/lastcoinvision/lcv/app"
	emissiontypes "github.com/lastcoinvision/lcv/x/emission/types"
)

const (
	denom = "ulcv"

	// oneLCV is 1 LCV in ulcv (6 decimals).
	oneLCV = 1_000_000

	// totalSupply is the fixed 10,000,000,000 LCV total supply, in ulcv.
	// Never minted beyond this: x/mint is pinned to zero inflation below,
	// and x/emission only ever transfers out of the pre-funded
	// staking-rewards pool, never mints.
	totalSupplyLCV = 10_000_000_000

	oneYearSeconds = 365 * 24 * 3600
)

// allocation is one row of the top-level spec's token allocation table.
type allocation struct {
	name       string
	percentage int64 // of totalSupplyLCV; all seven must sum to 100
}

var allocations = []allocation{
	{"ecosystem_grants", 25},
	{"foundation_treasury", 20},
	{"team", 15},
	{"investors", 15},
	{"staking_rewards", 12},
	{"public", 8},
	{"community_airdrop", 5},
}

func amountFor(pct int64) math.Int {
	return math.NewInt(totalSupplyLCV).MulRaw(oneLCV).MulRaw(pct).QuoRaw(100)
}

func main() {
	var (
		genesisPath   = flag.String("genesis", "", "path to genesis.json to patch (required)")
		ecosystemAddr = flag.String("ecosystem-addr", "", "ecosystem/grants pool address (required)")
		treasuryAddr  = flag.String("treasury-addr", "", "foundation treasury address (required)")
		teamAddr      = flag.String("team-addr", "", "team vesting account address (required)")
		investorsAddr = flag.String("investors-addr", "", "investors vesting account address (required)")
		publicAddr    = flag.String("public-addr", "", "public sale address (required)")
		communityAddr = flag.String("community-addr", "", "community/airdrop address (required)")
		devnetFastGov = flag.Bool("devnet-fast-gov", false, "shorten x/gov voting/deposit periods to seconds instead of days, for fast local governance testing (never use for a real testnet/mainnet genesis)")
	)
	flag.Parse()

	if *genesisPath == "" || *ecosystemAddr == "" || *treasuryAddr == "" || *teamAddr == "" ||
		*investorsAddr == "" || *publicAddr == "" || *communityAddr == "" {
		flag.Usage()
		os.Exit(2)
	}

	var totalPct int64
	for _, a := range allocations {
		totalPct += a.percentage
	}
	if totalPct != 100 {
		exitf("allocation table sums to %d%%, not 100%% — refusing to guess which entry is wrong", totalPct)
	}

	registry := codectypes.NewInterfaceRegistry()
	authtypes.RegisterInterfaces(registry)
	vestingtypes.RegisterInterfaces(registry)
	govtypesv1.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	appGenesis, err := genutiltypes.AppGenesisFromFile(*genesisPath)
	if err != nil {
		exitf("reading genesis: %v", err)
	}

	var appState map[string]json.RawMessage
	if err := json.Unmarshal(appGenesis.AppState, &appState); err != nil {
		exitf("unmarshalling app_state: %v", err)
	}

	var authGenState authtypes.GenesisState
	cdc.MustUnmarshalJSON(appState[authtypes.ModuleName], &authGenState)

	var bankGenState banktypes.GenesisState
	cdc.MustUnmarshalJSON(appState[banktypes.ModuleName], &bankGenState)

	nextAccNum := nextAccountNumber(cdc, authGenState.Accounts)
	genesisTime := appGenesis.GenesisTime
	if genesisTime.IsZero() {
		genesisTime = time.Now()
	}

	addAccount := func(addr sdk.AccAddress, amount math.Int) {
		acc := authtypes.NewBaseAccount(addr, nil, nextAccNum, 0)
		nextAccNum++
		anyAcc, err := codectypes.NewAnyWithValue(acc)
		if err != nil {
			exitf("packing account %s: %v", addr, err)
		}
		authGenState.Accounts = append(authGenState.Accounts, anyAcc)
		bankGenState.Balances = append(bankGenState.Balances, banktypes.Balance{
			Address: addr.String(),
			Coins:   sdk.NewCoins(sdk.NewCoin(denom, amount)),
		})
	}

	addVestingAccount := func(addrStr string, amount math.Int, cliffSeconds, totalSeconds int64, numPeriodsAfterCliff int) {
		addr, err := sdk.AccAddressFromBech32(addrStr)
		if err != nil {
			exitf("invalid address %q: %v", addrStr, err)
		}
		base := authtypes.NewBaseAccount(addr, nil, nextAccNum, 0)
		nextAccNum++

		coins := sdk.NewCoins(sdk.NewCoin(denom, amount))
		startTime := genesisTime.Unix()
		endTime := startTime + totalSeconds

		periods := vestingPeriods(amount, cliffSeconds, totalSeconds, numPeriodsAfterCliff)

		bva, err := vestingtypes.NewBaseVestingAccount(base, coins, endTime)
		if err != nil {
			exitf("building base vesting account for %s: %v", addrStr, err)
		}
		pva := vestingtypes.NewPeriodicVestingAccountRaw(bva, startTime, periods)

		anyAcc, err := codectypes.NewAnyWithValue(pva)
		if err != nil {
			exitf("packing vesting account %s: %v", addrStr, err)
		}
		authGenState.Accounts = append(authGenState.Accounts, anyAcc)
		bankGenState.Balances = append(bankGenState.Balances, banktypes.Balance{
			Address: addrStr,
			Coins:   coins,
		})
	}

	addModuleAccount := func(name string, amount math.Int) {
		base := authtypes.NewBaseAccount(authtypes.NewModuleAddress(name), nil, nextAccNum, 0)
		nextAccNum++
		macc := authtypes.NewModuleAccount(base, name)
		anyAcc, err := codectypes.NewAnyWithValue(macc)
		if err != nil {
			exitf("packing module account %s: %v", name, err)
		}
		authGenState.Accounts = append(authGenState.Accounts, anyAcc)
		bankGenState.Balances = append(bankGenState.Balances, banktypes.Balance{
			Address: macc.GetAddress().String(),
			Coins:   sdk.NewCoins(sdk.NewCoin(denom, amount)),
		})
	}

	for _, a := range allocations {
		amount := amountFor(a.percentage)
		switch a.name {
		case "ecosystem_grants":
			addr := mustAddr(*ecosystemAddr)
			addAccount(addr, amount)
		case "foundation_treasury":
			addr := mustAddr(*treasuryAddr)
			addAccount(addr, amount)
		case "public":
			addr := mustAddr(*publicAddr)
			addAccount(addr, amount)
		case "community_airdrop":
			addr := mustAddr(*communityAddr)
			addAccount(addr, amount)
		case "team":
			// 4-year vest, 1-year cliff, then quarterly release over the
			// remaining 3 years (12 quarters).
			addVestingAccount(*teamAddr, amount, oneYearSeconds, 4*oneYearSeconds, 12)
		case "investors":
			// 3-year vest, 1-year cliff, then quarterly release over the
			// remaining 2 years (8 quarters).
			addVestingAccount(*investorsAddr, amount, oneYearSeconds, 3*oneYearSeconds, 8)
		case "staking_rewards":
			// Funds x/emission's decay curve; see modules/emission/SPEC.md.
			// No Minter/Burner permission (set in app/app_config.go) — this
			// account can only ever be drawn down, never topped up on-chain.
			addModuleAccount(emissiontypes.StakingRewardsAccountName, amount)
		default:
			exitf("unhandled allocation bucket %q (bug in gengenesis)", a.name)
		}
	}

	// Leave Supply empty so x/bank recomputes the true total from balances
	// at InitGenesis, rather than requiring us to hand-add these new
	// balances to whatever total `multi-node` already put here.
	bankGenState.Supply = nil

	appState[authtypes.ModuleName] = cdc.MustMarshalJSON(&authGenState)
	appState[banktypes.ModuleName] = cdc.MustMarshalJSON(&bankGenState)
	appState[minttypes.ModuleName] = zeroInflationMintGenesis(cdc)

	if *devnetFastGov {
		appState[govtypes.ModuleName] = fastGovGenesis(cdc, appState[govtypes.ModuleName])
	}

	newAppStateBz, err := json.Marshal(appState)
	if err != nil {
		exitf("marshalling app_state: %v", err)
	}
	appGenesis.AppState = newAppStateBz

	if err := appGenesis.ValidateAndComplete(); err != nil {
		exitf("validating patched genesis: %v", err)
	}
	if err := appGenesis.SaveAs(*genesisPath); err != nil {
		exitf("saving genesis: %v", err)
	}

	fmt.Printf("gengenesis: patched %s with %d allocation accounts (mint pinned to zero inflation)\n", *genesisPath, len(allocations))
}

// vestingPeriods builds a cliff-then-linear PeriodicVestingAccount period
// list. Cosmos SDK's PeriodicVestingAccount rejects any period with a
// zero/empty Amount (see (PeriodicVestingAccount).Validate:
// "!p.Amount.IsAllPositive()"), so a literal zero-coin period can't
// represent "nothing releases during the cliff" — instead, the cliff
// duration is folded into the *length* of the first releasing period, so
// nothing vests until cliffSeconds + one slice have elapsed, then
// numPeriodsAfterCliff equal slices release in total (the last absorbing
// any remainder from integer division), ending at exactly totalSeconds.
// This is the conventional "N-year cliff, then vests quarterly/monthly
// thereafter" schedule: the first unlock lands one period after the cliff,
// not a separate zero-amount instant at the cliff itself.
func vestingPeriods(amount math.Int, cliffSeconds, totalSeconds int64, numPeriodsAfterCliff int) vestingtypes.Periods {
	remainingSeconds := totalSeconds - cliffSeconds
	periodLen := remainingSeconds / int64(numPeriodsAfterCliff)
	periodAmt := amount.QuoRaw(int64(numPeriodsAfterCliff))

	periods := make(vestingtypes.Periods, 0, numPeriodsAfterCliff)
	allocatedSeconds := int64(0) // tracks post-cliff seconds only
	allocatedAmt := math.ZeroInt()
	for i := 0; i < numPeriodsAfterCliff; i++ {
		sliceLen := periodLen
		amt := periodAmt
		if i == numPeriodsAfterCliff-1 {
			sliceLen = remainingSeconds - allocatedSeconds
			amt = amount.Sub(allocatedAmt)
		}

		length := sliceLen
		if i == 0 {
			length += cliffSeconds
		}
		periods = append(periods, vestingtypes.Period{
			Length: length,
			Amount: sdk.NewCoins(sdk.NewCoin(denom, amt)),
		})

		allocatedSeconds += sliceLen
		allocatedAmt = allocatedAmt.Add(amt)
	}
	return periods
}

// fastGovGenesis shortens x/gov's voting and deposit periods to seconds so
// devnet operators can actually exercise a full proposal lifecycle in a
// terminal session instead of waiting out the mainnet-realistic 48h/60-day
// defaults. Never intended for a real testnet/mainnet genesis — see
// DECISIONS.md.
func fastGovGenesis(cdc codec.Codec, raw json.RawMessage) json.RawMessage {
	var genState govtypesv1.GenesisState
	cdc.MustUnmarshalJSON(raw, &genState)

	votingPeriod := 30 * time.Second
	expeditedVotingPeriod := 15 * time.Second
	depositPeriod := 60 * time.Second

	genState.Params.VotingPeriod = &votingPeriod
	genState.Params.ExpeditedVotingPeriod = &expeditedVotingPeriod
	genState.Params.MaxDepositPeriod = &depositPeriod

	return cdc.MustMarshalJSON(&genState)
}

func zeroInflationMintGenesis(cdc codec.Codec) json.RawMessage {
	genState := minttypes.DefaultGenesisState()
	genState.Params.MintDenom = denom
	genState.Params.InflationMin = math.LegacyZeroDec()
	genState.Params.InflationMax = math.LegacyZeroDec()
	genState.Params.InflationRateChange = math.LegacyZeroDec()
	genState.Minter.Inflation = math.LegacyZeroDec()
	genState.Minter.AnnualProvisions = math.LegacyZeroDec()
	return cdc.MustMarshalJSON(genState)
}

func nextAccountNumber(cdc codec.Codec, accounts []*codectypes.Any) uint64 {
	var maxNum uint64
	for _, any := range accounts {
		var acc sdk.AccountI
		if err := cdc.UnpackAny(any, &acc); err != nil {
			exitf("unpacking existing genesis account: %v", err)
		}
		if n := acc.GetAccountNumber(); n >= maxNum {
			maxNum = n + 1
		}
	}
	return maxNum
}

func mustAddr(s string) sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(s)
	if err != nil {
		exitf("invalid address %q: %v", s, err)
	}
	return addr
}

func exitf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "gengenesis: "+format+"\n", args...)
	os.Exit(1)
}
