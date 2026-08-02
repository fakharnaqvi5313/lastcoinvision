package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cast"

	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/gogoproto/proto"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	"cosmossdk.io/client/v2/autocli"
	clienthelpers "cosmossdk.io/client/v2/helpers"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/circuit"
	circuitkeeper "cosmossdk.io/x/circuit/keeper"
	circuittypes "cosmossdk.io/x/circuit/types"
	"cosmossdk.io/x/evidence"
	evidencekeeper "cosmossdk.io/x/evidence/keeper"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	feegrantmodule "cosmossdk.io/x/feegrant/module"
	"cosmossdk.io/x/nft"
	nftkeeper "cosmossdk.io/x/nft/keeper"
	nftmodule "cosmossdk.io/x/nft/module"
	"cosmossdk.io/x/upgrade"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/ethereum/go-ethereum/common"
	corevm "github.com/ethereum/go-ethereum/core/vm"
	_ "github.com/ethereum/go-ethereum/eth/tracers/js"     // force-load js tracer
	_ "github.com/ethereum/go-ethereum/eth/tracers/native" // force-load native tracer

	evmante "github.com/cosmos/evm/ante"
	antetypes "github.com/cosmos/evm/ante/types"
	evmconfig "github.com/cosmos/evm/config"
	evmencoding "github.com/cosmos/evm/encoding"
	evmaddress "github.com/cosmos/evm/encoding/address"
	evmmempool "github.com/cosmos/evm/mempool"
	precompiletypes "github.com/cosmos/evm/precompiles/types"
	srvflags "github.com/cosmos/evm/server/flags"
	evmutils "github.com/cosmos/evm/utils"
	"github.com/cosmos/evm/x/erc20"
	erc20keeper "github.com/cosmos/evm/x/erc20/keeper"
	erc20types "github.com/cosmos/evm/x/erc20/types"
	evmfeemarket "github.com/cosmos/evm/x/feemarket"
	evmfeemarketkeeper "github.com/cosmos/evm/x/feemarket/keeper"
	evmfeemarkettypes "github.com/cosmos/evm/x/feemarket/types"
	"github.com/cosmos/evm/x/precisebank"
	precisebankkeeper "github.com/cosmos/evm/x/precisebank/keeper"
	precisebanktypes "github.com/cosmos/evm/x/precisebank/types"
	"github.com/cosmos/evm/x/vm"
	evmkeeper "github.com/cosmos/evm/x/vm/keeper"
	evmtypes "github.com/cosmos/evm/x/vm/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	authcodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkmempool "github.com/cosmos/cosmos-sdk/types/mempool"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/group"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"
	groupmodule "github.com/cosmos/cosmos-sdk/x/group/module"
	"github.com/cosmos/cosmos-sdk/x/mint"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	ica "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/types"
	icahostkeeper "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/types"
	ibctransfer "github.com/cosmos/ibc-go/v10/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v10/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v10/modules/core"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v10/modules/core/keeper"
	ibctm "github.com/cosmos/ibc-go/v10/modules/light-clients/07-tendermint"

	emissionmodulekeeper "github.com/lastcoinvision/lcv/x/emission/keeper"
	emissionmodule "github.com/lastcoinvision/lcv/x/emission/module"
	emissionmoduletypes "github.com/lastcoinvision/lcv/x/emission/types"
	feemarketmodulekeeper "github.com/lastcoinvision/lcv/x/feemarket/keeper"
	feemarketmodule "github.com/lastcoinvision/lcv/x/feemarket/module"
	feemarketmoduletypes "github.com/lastcoinvision/lcv/x/feemarket/types"
	identitymodulekeeper "github.com/lastcoinvision/lcv/x/identity/keeper"
	identitymodule "github.com/lastcoinvision/lcv/x/identity/module"
	identitymoduletypes "github.com/lastcoinvision/lcv/x/identity/types"
	registrymodulekeeper "github.com/lastcoinvision/lcv/x/registry/keeper"
	registrymodule "github.com/lastcoinvision/lcv/x/registry/module"
	registrymoduletypes "github.com/lastcoinvision/lcv/x/registry/types"
	taxrailmodulekeeper "github.com/lastcoinvision/lcv/x/taxrail/keeper"
	taxrailmodule "github.com/lastcoinvision/lcv/x/taxrail/module"
	taxrailmoduletypes "github.com/lastcoinvision/lcv/x/taxrail/types"
	votingmodulekeeper "github.com/lastcoinvision/lcv/x/voting/keeper"
	votingmodule "github.com/lastcoinvision/lcv/x/voting/module"
	votingmoduletypes "github.com/lastcoinvision/lcv/x/voting/types"

	"github.com/lastcoinvision/lcv/docs"
)

const (
	// Name is the name of the application.
	Name = "lcv"
	// AccountAddressPrefix is the prefix for accounts addresses.
	AccountAddressPrefix = "lcv"
	// ChainCoinType is the coin type of the chain.
	ChainCoinType = 118

	// EVMChainID is the EIP-155 replay-protection chain id used by this
	// chain's EVM. Arbitrary placeholder for devnet/testnet — a real
	// mainnet launch must first reserve a permanent id via
	// https://github.com/ethereum-lists/chains (chainlist.org) rather than
	// reusing this value. See modules/evm/SPEC.md.
	EVMChainID = 2611

	// EVMBaseDenom is the denom EVM gas/state transitions are priced and
	// held in — the chain's existing bond denom, unchanged from Phase A.
	EVMBaseDenom = "ulcv"
	// EVMExtendedDenom is the atto-precision (18-decimal) denom
	// cosmos/evm's x/precisebank tracks internally so ulcv (6 decimals)
	// can still be represented at full EVM precision without minting any
	// new supply — see modules/evm/SPEC.md.
	EVMExtendedDenom = "alcv"
	// EVMDisplayDenom is the human-facing unit name; combined with the
	// bank denom metadata registered in bankAppModuleBasic, this is what
	// tells cosmos/evm that EVMBaseDenom has 6 decimals, not 18.
	EVMDisplayDenom = "lcv"
)

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string
)

func init() {
	var err error
	clienthelpers.EnvPrefix = Name
	DefaultNodeHome, err = clienthelpers.GetNodeHomeDirectory("." + Name)
	if err != nil {
		panic(err)
	}
}

// maccPerms is the map of module account permissions. See BlockedAddresses
// for accounts that additionally can't receive funds via a plain MsgSend.
var maccPerms = map[string][]string{
	authtypes.FeeCollectorName:                    nil,
	distrtypes.ModuleName:                         nil,
	minttypes.ModuleName:                          {authtypes.Minter},
	stakingtypes.BondedPoolName:                   {authtypes.Burner, authtypes.Staking},
	stakingtypes.NotBondedPoolName:                {authtypes.Burner, authtypes.Staking},
	govtypes.ModuleName:                           {authtypes.Burner},
	nft.ModuleName:                                nil,
	ibctransfertypes.ModuleName:                   {authtypes.Minter, authtypes.Burner},
	icatypes.ModuleName:                           nil,
	wasmtypes.ModuleName:                          {authtypes.Burner},
	emissionmoduletypes.StakingRewardsAccountName: nil,
	// Cosmos EVM modules
	evmtypes.ModuleName:          {authtypes.Minter, authtypes.Burner},
	evmfeemarkettypes.ModuleName: nil,
	erc20types.ModuleName:        {authtypes.Minter, authtypes.Burner},
	precisebanktypes.ModuleName:  {authtypes.Minter, authtypes.Burner},
	// Transient holding account for the fee-market ante decorator: fees
	// pass through here so the burned portion of the base fee can
	// actually be burned before the remainder is forwarded on.
	feemarketmoduletypes.ModuleName: {authtypes.Burner},
	// Pooled collector treasury for MsgTaxedTransfer's tax share, tracked
	// per-jurisdiction via JurisdictionLedger. No Minter/Burner permission:
	// x/taxrail only ever moves already-existing coins in (from senders)
	// and out (via governance-gated MsgSpendCollectedTax), never mints or
	// burns.
	taxrailmoduletypes.ModuleName: nil,
}

// GetMaccPerms returns a copy of the module account permissions.
//
// NOTE: This is solely to be used for testing purposes.
func GetMaccPerms() map[string][]string {
	dup := make(map[string][]string, len(maccPerms))
	for k, v := range maccPerms {
		dup[k] = v
	}
	return dup
}

// BlockedAddresses returns all the app's blocked account addresses: module
// accounts that must never receive funds via a plain MsgSend (only through
// their module's own designated code path), plus every EVM precompile
// address (Ethereum's own precompiles and cosmos/evm's static precompiles),
// which likewise must never be treated as an ordinary bank account.
func BlockedAddresses() map[string]bool {
	blocked := map[string]bool{
		authtypes.FeeCollectorName:     true,
		distrtypes.ModuleName:          true,
		minttypes.ModuleName:           true,
		stakingtypes.BondedPoolName:    true,
		stakingtypes.NotBondedPoolName: true,
		nft.ModuleName:                 true,
		// Collected tax should only ever arrive via MsgTaxedTransfer's split
		// logic (which credits the matching JurisdictionLedger in the same
		// handler call), never a generic MsgSend that would leave the
		// ledger under-counting the module account's real balance and trip
		// LedgerMatchesModuleBalanceInvariant.
		taxrailmoduletypes.ModuleName: true,
		// govtypes.ModuleName is deliberately NOT blocked: it may receive
		// funds directly (e.g. community-pool funding).
	}

	blockedPrecompilesHex := append([]string{}, evmtypes.AvailableStaticPrecompiles...)
	for _, addr := range corevm.PrecompiledAddressesPrague {
		blockedPrecompilesHex = append(blockedPrecompilesHex, addr.Hex())
	}
	for _, precompile := range blockedPrecompilesHex {
		blocked[evmutils.Bech32StringFromHexAddress(precompile)] = true
	}

	return blocked
}

// getGovProposalHandlers return the chain proposal handlers.
func getGovProposalHandlers() []govclient.ProposalHandler {
	var govProposalHandlers []govclient.ProposalHandler
	return govProposalHandlers
}

var (
	_ runtime.AppI            = (*App)(nil)
	_ servertypes.Application = (*App)(nil)
)

// App extends an ABCI application, but with most of its parameters exported.
type App struct {
	*baseapp.BaseApp

	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry codectypes.InterfaceRegistry

	// keys to access the substores
	keys  map[string]*storetypes.KVStoreKey
	tkeys map[string]*storetypes.TransientStoreKey

	pendingTxListeners []evmante.PendingTxListener
	clientCtx          client.Context

	// keepers
	AccountKeeper         authkeeper.AccountKeeper
	BankKeeper            bankkeeper.Keeper
	StakingKeeper         *stakingkeeper.Keeper
	DistrKeeper           distrkeeper.Keeper
	ConsensusParamsKeeper consensusparamkeeper.Keeper
	SlashingKeeper        slashingkeeper.Keeper
	MintKeeper            mintkeeper.Keeper
	GovKeeper             govkeeper.Keeper
	CrisisKeeper          *crisiskeeper.Keeper
	UpgradeKeeper         *upgradekeeper.Keeper
	AuthzKeeper           authzkeeper.Keeper
	EvidenceKeeper        evidencekeeper.Keeper
	FeeGrantKeeper        feegrantkeeper.Keeper
	GroupKeeper           groupkeeper.Keeper
	NFTKeeper             nftkeeper.Keeper
	CircuitBreakerKeeper  circuitkeeper.Keeper

	// IBC
	IBCKeeper           *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	ICAControllerKeeper icacontrollerkeeper.Keeper
	ICAHostKeeper       icahostkeeper.Keeper
	TransferKeeper      ibctransferkeeper.Keeper
	tmLightClientModule ibctm.LightClientModule

	// CosmWasm
	WasmKeeper wasmkeeper.Keeper

	// Cosmos EVM
	EVMFeeMarketKeeper evmfeemarketkeeper.Keeper
	EVMKeeper          *evmkeeper.Keeper
	Erc20Keeper        erc20keeper.Keeper
	PreciseBankKeeper  precisebankkeeper.Keeper
	EVMMempool         *evmmempool.ExperimentalEVMMempool

	// chain modules
	EmissionKeeper  emissionmodulekeeper.Keeper
	FeemarketKeeper feemarketmodulekeeper.Keeper
	IdentityKeeper  identitymodulekeeper.Keeper
	RegistryKeeper  registrymodulekeeper.Keeper
	VotingKeeper    votingmodulekeeper.Keeper
	TaxrailKeeper   taxrailmodulekeeper.Keeper

	// the module manager
	ModuleManager      *module.Manager
	BasicModuleManager module.BasicManager

	// simulation manager
	sm *module.SimulationManager

	// module configurator
	configurator module.Configurator
}

// New returns a reference to an initialized App.
func New(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) (*App, error) {
	// evmtypes.SetChainConfig (called inside evmkeeper.NewKeeper below) is
	// a process-global, set-once value: a second call panics unless the
	// FIRST call used exactly evmtypes.DefaultEVMChainID as a sentinel.
	// cmd/lcvd/cmd/root.go builds a throwaway app.New(...) purely to read
	// back its codec/AutoCLI options (via simtestutil.EmptyAppOptions-style
	// options with no evm-chain-id key set) before the real app is built
	// for `start` et al. (from real config with evm.evm-chain-id set) — so
	// resolving the id from appOpts, defaulting to the sentinel rather than
	// this chain's real EVMChainID constant, is what lets the second,
	// real construction actually take effect instead of panicking. See
	// DECISIONS.md.
	evmChainID := cast.ToUint64(appOpts.Get(srvflags.EVMChainID))
	if evmChainID == 0 {
		evmChainID = evmtypes.DefaultEVMChainID
	}

	// evmencoding.MakeConfig (rather than a hand-built interface registry,
	// as used for the wasmd-only milestone) registers the eth_secp256k1
	// pubkey type, EIP-712 signing support, and custom GetSigners for
	// MsgEthereumTx/MsgConvertERC20 that cosmos/evm requires — see
	// DECISIONS.md and modules/evm/SPEC.md.
	encodingConfig := evmencoding.MakeConfig(evmChainID)
	appCodec := encodingConfig.Codec
	legacyAmino := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry
	txConfig := encodingConfig.TxConfig

	bApp := baseapp.NewBaseApp(Name, logger, db, txConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)
	bApp.SetTxEncoder(txConfig.TxEncoder())

	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, consensusparamtypes.StoreKey, crisistypes.StoreKey,
		upgradetypes.StoreKey, feegrant.StoreKey, evidencetypes.StoreKey,
		circuittypes.StoreKey, authzkeeper.StoreKey, nftkeeper.StoreKey, group.StoreKey,
		// ibc keys
		ibcexported.StoreKey, ibctransfertypes.StoreKey,
		icahosttypes.StoreKey, icacontrollertypes.StoreKey,
		// wasm key
		wasmtypes.StoreKey,
		// Cosmos EVM keys
		evmtypes.StoreKey, evmfeemarkettypes.StoreKey, erc20types.StoreKey, precisebanktypes.StoreKey,
		// chain module keys
		emissionmoduletypes.StoreKey, feemarketmoduletypes.StoreKey,
		identitymoduletypes.StoreKey, registrymoduletypes.StoreKey,
		votingmoduletypes.StoreKey, taxrailmoduletypes.StoreKey,
	)

	tkeys := storetypes.NewTransientStoreKeys(evmtypes.TransientKey, evmfeemarkettypes.TransientKey)

	if err := bApp.RegisterStreamingServices(appOpts, keys); err != nil {
		return nil, err
	}

	app := &App{
		BaseApp:           bApp,
		legacyAmino:       legacyAmino,
		appCodec:          appCodec,
		txConfig:          txConfig,
		interfaceRegistry: interfaceRegistry,
		keys:              keys,
		tkeys:             tkeys,
	}

	authAddr := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	app.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[consensusparamtypes.StoreKey]),
		authAddr,
		runtime.EventService{},
	)
	bApp.SetParamStore(app.ConsensusParamsKeeper.ParamsStore)

	// evmaddress.NewEvmCodec (rather than a plain bech32 codec) accepts
	// both hex ("0x...") and bech32 address strings and treats them as the
	// same 20-byte address space — required so a MetaMask-style
	// Keccak256-derived address and its bech32 encoding always refer to
	// the same account.
	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		evmaddress.NewEvmCodec(AccountAddressPrefix),
		AccountAddressPrefix,
		authAddr,
	)

	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		app.AccountKeeper,
		BlockedAddresses(),
		authAddr,
		logger,
	)

	app.StakingKeeper = stakingkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[stakingtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		authAddr,
		evmaddress.NewEvmCodec(AccountAddressPrefix+"valoper"),
		evmaddress.NewEvmCodec(AccountAddressPrefix+"valcons"),
	)

	app.MintKeeper = mintkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[minttypes.StoreKey]),
		app.StakingKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		authtypes.FeeCollectorName,
		authAddr,
	)

	app.DistrKeeper = distrkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[distrtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		authtypes.FeeCollectorName,
		authAddr,
	)

	app.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec,
		legacyAmino,
		runtime.NewKVStoreService(keys[slashingtypes.StoreKey]),
		app.StakingKeeper,
		authAddr,
	)

	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[feegrant.StoreKey]),
		app.AccountKeeper,
	)

	app.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(app.DistrKeeper.Hooks(), app.SlashingKeeper.Hooks()),
	)

	app.CircuitBreakerKeeper = circuitkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[circuittypes.StoreKey]),
		authAddr,
		app.AccountKeeper.AddressCodec(),
	)
	app.SetCircuitBreaker(&app.CircuitBreakerKeeper)

	app.AuthzKeeper = authzkeeper.NewKeeper(
		runtime.NewKVStoreService(keys[authzkeeper.StoreKey]),
		appCodec,
		app.MsgServiceRouter(),
		app.AccountKeeper,
	)

	groupConfig := group.DefaultConfig()
	app.GroupKeeper = groupkeeper.NewKeeper(
		keys[group.StoreKey],
		appCodec,
		app.MsgServiceRouter(),
		app.AccountKeeper,
		groupConfig,
	)

	app.NFTKeeper = nftkeeper.NewKeeper(
		runtime.NewKVStoreService(keys[nftkeeper.StoreKey]),
		appCodec,
		app.AccountKeeper,
		app.BankKeeper,
	)

	skipUpgradeHeights := map[int64]bool{}
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}
	homePath := cast.ToString(appOpts.Get("home"))
	app.UpgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		runtime.NewKVStoreService(keys[upgradetypes.StoreKey]),
		appCodec,
		homePath,
		app.BaseApp,
		authAddr,
	)

	app.IBCKeeper = ibckeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[ibcexported.StoreKey]),
		nil,
		app.UpgradeKeeper,
		authAddr,
	)

	app.TransferKeeper = ibctransferkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[ibctransfertypes.StoreKey]),
		nil,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.MsgServiceRouter(),
		app.AccountKeeper,
		app.BankKeeper,
		authAddr,
	)

	app.ICAHostKeeper = icahostkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[icahosttypes.StoreKey]),
		nil,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.AccountKeeper,
		app.MsgServiceRouter(),
		app.GRPCQueryRouter(),
		authAddr,
	)

	app.ICAControllerKeeper = icacontrollerkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[icacontrollertypes.StoreKey]),
		nil,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.MsgServiceRouter(),
		authAddr,
	)

	govRouter := govv1beta1.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govv1beta1.ProposalHandler)
	govConfig := govtypes.DefaultConfig()
	govKeeper := govkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[govtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		app.DistrKeeper,
		app.MsgServiceRouter(),
		govConfig,
		authAddr,
	)
	govKeeper.SetLegacyRouter(govRouter)
	app.GovKeeper = *govKeeper.SetHooks(govtypes.NewMultiGovHooks())

	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[evidencetypes.StoreKey]),
		app.StakingKeeper,
		app.SlashingKeeper,
		app.AccountKeeper.AddressCodec(),
		runtime.ProvideCometInfoService(),
	)
	app.EvidenceKeeper = *evidenceKeeper

	app.CrisisKeeper = crisiskeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[crisistypes.StoreKey]),
		invCheckPeriod(appOpts),
		app.BankKeeper,
		authtypes.FeeCollectorName,
		authAddr,
		app.AccountKeeper.AddressCodec(),
	)

	// Cosmos EVM keepers. EVMKeeper is constructed before Erc20Keeper
	// because it takes a pointer to app.Erc20Keeper (assigned afterward —
	// fine, since the pointer is only dereferenced at call time, by which
	// point Erc20Keeper is populated), matching cosmos/evm's own reference
	// app (evmd) ordering exactly.
	app.EVMFeeMarketKeeper = evmfeemarketkeeper.NewKeeper(
		appCodec, authtypes.NewModuleAddress(govtypes.ModuleName),
		keys[evmfeemarkettypes.StoreKey],
		tkeys[evmfeemarkettypes.TransientKey],
	)

	app.PreciseBankKeeper = precisebankkeeper.NewKeeper(
		appCodec,
		keys[precisebanktypes.StoreKey],
		app.BankKeeper,
		app.AccountKeeper,
	)

	app.EVMKeeper = evmkeeper.NewKeeper(
		appCodec, keys[evmtypes.StoreKey], tkeys[evmtypes.TransientKey], keys,
		authtypes.NewModuleAddress(govtypes.ModuleName),
		app.AccountKeeper,
		app.PreciseBankKeeper,
		app.StakingKeeper,
		app.EVMFeeMarketKeeper,
		&app.ConsensusParamsKeeper,
		&app.Erc20Keeper,
		evmChainID,
		cast.ToString(appOpts.Get(srvflags.EVMTracer)),
	).WithStaticPrecompiles(
		precompiletypes.DefaultStaticPrecompiles(
			*app.StakingKeeper,
			app.DistrKeeper,
			app.PreciseBankKeeper,
			&app.Erc20Keeper,
			&app.TransferKeeper,
			app.IBCKeeper.ChannelKeeper,
			app.GovKeeper,
			app.SlashingKeeper,
			appCodec,
		),
	)

	app.Erc20Keeper = erc20keeper.NewKeeper(
		keys[erc20types.StoreKey],
		appCodec,
		authtypes.NewModuleAddress(govtypes.ModuleName),
		app.AccountKeeper,
		app.PreciseBankKeeper,
		app.EVMKeeper,
		app.StakingKeeper,
		&app.TransferKeeper,
	)

	// chain modules
	app.EmissionKeeper = emissionmodulekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[emissionmoduletypes.StoreKey]),
		logger,
		authAddr,
		app.BankKeeper,
	)
	app.FeemarketKeeper = feemarketmodulekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[feemarketmoduletypes.StoreKey]),
		logger,
		authAddr,
	)
	app.IdentityKeeper = identitymodulekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[identitymoduletypes.StoreKey]),
		logger,
		authAddr,
		app.GroupKeeper,
	)
	app.RegistryKeeper = registrymodulekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[registrymoduletypes.StoreKey]),
		logger,
		authAddr,
		app.IdentityKeeper,
	)
	app.VotingKeeper = votingmodulekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[votingmoduletypes.StoreKey]),
		logger,
		authAddr,
		app.IdentityKeeper,
	)
	app.TaxrailKeeper = taxrailmodulekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[taxrailmoduletypes.StoreKey]),
		logger,
		authAddr,
		app.BankKeeper,
	)

	// CosmWasm keeper — registered last among stateful keepers since it can
	// depend on virtually everything above (bank, staking, gov, distribution,
	// IBC). See ibc.go for router/IBC wiring and the CodeUploadAccess /
	// InstantiateDefaultPermission params set at genesis.
	wasmDir := homePath + "/wasm"
	wasmConfig, err := wasm.ReadNodeConfig(appOpts)
	if err != nil {
		return nil, fmt.Errorf("error while reading wasm config: %w", err)
	}
	app.WasmKeeper = wasmkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[wasmtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		distrkeeper.NewQuerier(app.DistrKeeper),
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeperV2,
		app.TransferKeeper,
		app.MsgServiceRouter(),
		app.GRPCQueryRouter(),
		wasmDir,
		wasmConfig,
		wasmtypes.VMConfig{},
		wasmkeeper.BuiltInCapabilities(),
		authAddr,
	)

	if err := app.registerIBCModules(); err != nil {
		return nil, err
	}

	anteHandler, err := NewAnteHandler(HandlerOptions{
		AccountKeeper:          app.AccountKeeper,
		BankKeeper:             app.BankKeeper,
		FeemarketKeeper:        app.FeemarketKeeper,
		FeegrantKeeper:         app.FeeGrantKeeper,
		SignModeHandler:        txConfig.SignModeHandler(),
		ExtensionOptionChecker: antetypes.HasDynamicFeeExtensionOption,
		IBCKeeper:              app.IBCKeeper,
		EvmKeeper:              app.EVMKeeper,
		EvmFeeMarketKeeper:     app.EVMFeeMarketKeeper,
		MaxTxGasWanted:         cast.ToUint64(appOpts.Get(srvflags.EVMMaxTxGasWanted)),
		PendingTxListener:      app.onPendingTx,
	})
	if err != nil {
		return nil, err
	}
	app.SetAnteHandler(anteHandler)

	// Set the EVM priority-nonce mempool. If you wish to use the SDK's
	// default (FIFO) mempool instead, remove this call.
	if err := app.configureEVMMempool(appOpts, logger); err != nil {
		return nil, fmt.Errorf("failed to configure EVM mempool: %w", err)
	}

	/****  Module Options ****/

	app.ModuleManager = module.NewManager(
		genutil.NewAppModule(app.AccountKeeper, app.StakingKeeper, app, txConfig),
		auth.NewAppModule(appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts, nil),
		vesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper, nil),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		gov.NewAppModule(appCodec, &app.GovKeeper, app.AccountKeeper, app.BankKeeper, nil),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper, nil, nil),
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, nil, app.interfaceRegistry),
		distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, nil),
		staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, nil),
		upgrade.NewAppModule(app.UpgradeKeeper, app.AccountKeeper.AddressCodec()),
		evidence.NewAppModule(app.EvidenceKeeper),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		groupmodule.NewAppModule(appCodec, app.GroupKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		nftmodule.NewAppModule(appCodec, app.NFTKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		consensus.NewAppModule(appCodec, app.ConsensusParamsKeeper),
		crisis.NewAppModule(app.CrisisKeeper, false, nil),
		circuit.NewAppModule(appCodec, app.CircuitBreakerKeeper),
		// non-sdk modules
		wasm.NewAppModule(appCodec, &app.WasmKeeper, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.MsgServiceRouter(), nil),
		ibc.NewAppModule(app.IBCKeeper),
		ibctransfer.NewAppModule(app.TransferKeeper),
		ica.NewAppModule(&app.ICAControllerKeeper, &app.ICAHostKeeper),
		ibctm.NewAppModule(app.tmLightClientModule),
		// Cosmos EVM modules
		vm.NewAppModule(app.EVMKeeper, app.AccountKeeper, app.BankKeeper, app.AccountKeeper.AddressCodec()),
		evmfeemarket.NewAppModule(app.EVMFeeMarketKeeper),
		erc20.NewAppModule(app.Erc20Keeper, app.AccountKeeper),
		precisebank.NewAppModule(app.PreciseBankKeeper, app.BankKeeper, app.AccountKeeper),
		// chain modules
		emissionmodule.NewAppModule(appCodec, app.EmissionKeeper, app.AccountKeeper, app.BankKeeper),
		feemarketmodule.NewAppModule(appCodec, app.FeemarketKeeper, app.AccountKeeper, app.BankKeeper),
		identitymodule.NewAppModule(appCodec, app.IdentityKeeper, app.AccountKeeper, app.BankKeeper),
		registrymodule.NewAppModule(appCodec, app.RegistryKeeper, app.AccountKeeper, app.BankKeeper),
		votingmodule.NewAppModule(appCodec, app.VotingKeeper, app.AccountKeeper, app.BankKeeper),
		taxrailmodule.NewAppModule(appCodec, app.TaxrailKeeper, app.AccountKeeper, app.BankKeeper),
	)

	app.BasicModuleManager = module.NewBasicManagerFromManager(
		app.ModuleManager,
		map[string]module.AppModuleBasic{
			genutiltypes.ModuleName:      genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
			govtypes.ModuleName:          gov.NewAppModuleBasic(getGovProposalHandlers()),
			wasmtypes.ModuleName:         wasmAppModuleBasic{},
			banktypes.ModuleName:         bankAppModuleBasic{},
			evmtypes.ModuleName:          evmAppModuleBasic{},
			evmfeemarkettypes.ModuleName: evmFeemarketAppModuleBasic{},
		},
	)
	app.BasicModuleManager.RegisterLegacyAminoCodec(legacyAmino)
	app.BasicModuleManager.RegisterInterfaces(interfaceRegistry)

	app.ModuleManager.SetOrderPreBlockers(
		upgradetypes.ModuleName,
		authtypes.ModuleName,
		evmtypes.ModuleName,
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	app.ModuleManager.SetOrderBeginBlockers(
		minttypes.ModuleName,
		// chain modules: emission must run after mint (which is pinned to
		// zero inflation, see genesis) and before distribution, so the
		// staking-rewards transfer for this block lands in the fee
		// collector in time for distribution's BeginBlocker to allocate it
		// to the proposer/validators/community pool alongside tx fees.
		emissionmoduletypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		authz.ModuleName,
		genutiltypes.ModuleName,
		ibcexported.ModuleName,
		ibctransfertypes.ModuleName,
		icatypes.ModuleName,
		feemarketmoduletypes.ModuleName,
		identitymoduletypes.ModuleName,
		registrymoduletypes.ModuleName,
		votingmoduletypes.ModuleName,
		taxrailmoduletypes.ModuleName,
		wasmtypes.ModuleName,
		// Cosmos EVM modules — erc20 and evm-feemarket must run before evm
		// itself (evm's base-fee adjustment reads the current block's gas
		// usage tallied by feemarket's own BeginBlocker).
		erc20types.ModuleName,
		evmfeemarkettypes.ModuleName,
		evmtypes.ModuleName,
		precisebanktypes.ModuleName,
	)

	app.ModuleManager.SetOrderEndBlockers(
		crisistypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
		genutiltypes.ModuleName,
		ibcexported.ModuleName,
		ibctransfertypes.ModuleName,
		icatypes.ModuleName,
		emissionmoduletypes.ModuleName,
		feemarketmoduletypes.ModuleName,
		identitymoduletypes.ModuleName,
		registrymoduletypes.ModuleName,
		votingmoduletypes.ModuleName,
		taxrailmoduletypes.ModuleName,
		wasmtypes.ModuleName,
		// Cosmos EVM modules
		evmtypes.ModuleName,
		erc20types.ModuleName,
		evmfeemarkettypes.ModuleName,
		precisebanktypes.ModuleName,
	)

	genesisModuleOrder := []string{
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		crisistypes.ModuleName,
		ibcexported.ModuleName,
		// Cosmos EVM modules — evm must be initialized before genutil so
		// that gentx transactions built against this genesis see correct
		// feemarket params (mirrors cosmos/evm's own reference app).
		evmtypes.ModuleName,
		evmfeemarkettypes.ModuleName,
		erc20types.ModuleName,
		precisebanktypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		ibctransfertypes.ModuleName,
		icatypes.ModuleName,
		feegrant.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		nft.ModuleName,
		group.ModuleName,
		consensusparamtypes.ModuleName,
		circuittypes.ModuleName,
		wasmtypes.ModuleName,
		emissionmoduletypes.ModuleName,
		feemarketmoduletypes.ModuleName,
		identitymoduletypes.ModuleName,
		registrymoduletypes.ModuleName,
		votingmoduletypes.ModuleName,
		taxrailmoduletypes.ModuleName,
	}
	app.ModuleManager.SetOrderInitGenesis(genesisModuleOrder...)
	app.ModuleManager.SetOrderExportGenesis(genesisModuleOrder...)

	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	if err := app.ModuleManager.RegisterServices(app.configurator); err != nil {
		return nil, err
	}

	autocliv1.RegisterQueryServer(app.GRPCQueryRouter(), runtimeservices.NewAutoCLIQueryService(app.ModuleManager.Modules))

	reflectionSvc, err := runtimeservices.NewReflectionService()
	if err != nil {
		return nil, err
	}
	reflectionv1.RegisterReflectionServiceServer(app.GRPCQueryRouter(), reflectionSvc)

	overrideModules := map[string]module.AppModuleSimulation{
		authtypes.ModuleName: auth.NewAppModule(app.appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts, nil),
		// The EVM stack's own randomized simulation genesis states don't
		// know about this chain's denom (bank), EvmDenom (vm), or its
		// rescaled base fee (evm feemarket) — see the App/Basic-manager
		// overrides above and DECISIONS.md for why each is fixed rather
		// than randomized here.
		banktypes.ModuleName:         bankAppModuleSimulation{AppModule: bank.NewAppModule(app.appCodec, app.BankKeeper, app.AccountKeeper, nil)},
		evmtypes.ModuleName:          evmAppModuleSimulation{AppModule: vm.NewAppModule(app.EVMKeeper, app.AccountKeeper, app.BankKeeper, app.AccountKeeper.AddressCodec())},
		evmfeemarkettypes.ModuleName: evmFeemarketAppModuleSimulation{AppModule: evmfeemarket.NewAppModule(app.EVMFeeMarketKeeper)},
	}
	app.sm = module.NewSimulationManagerFromAppModules(app.ModuleManager.Modules, overrideModules)
	app.sm.RegisterStoreDecoders()

	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)

	app.SetInitChainer(app.InitChainer)
	app.SetPreBlocker(app.PreBlocker)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)

	protoFiles, err := proto.MergedRegistry()
	if err != nil {
		return nil, err
	}
	if err := msgservice.ValidateProtoAnnotations(protoFiles); err != nil {
		fmt.Fprintln(io.Discard, err.Error())
	}

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			return nil, err
		}
	}

	return app, nil
}

func invCheckPeriod(appOpts servertypes.AppOptions) uint {
	return cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod))
}

// BeginBlocker application updates every begin block
func (app *App) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	return app.ModuleManager.BeginBlock(ctx)
}

// EndBlocker application updates every end block
func (app *App) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.ModuleManager.EndBlock(ctx)
}

// InitChainer application update at chain initialization
func (app *App) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	var genesisState GenesisState
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		return nil, err
	}
	if err := app.UpgradeKeeper.SetModuleVersionMap(ctx, app.ModuleManager.GetVersionMap()); err != nil {
		return nil, err
	}
	return app.ModuleManager.InitGenesis(ctx, app.appCodec, genesisState)
}

func (app *App) PreBlocker(ctx sdk.Context, _ *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
	return app.ModuleManager.PreBlock(ctx)
}

// LoadHeight loads a particular height
func (app *App) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// LegacyAmino returns App's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// AppCodec returns App's app codec.
func (app *App) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns App's interfaceRegistry.
func (app *App) InterfaceRegistry() codectypes.InterfaceRegistry {
	return app.interfaceRegistry
}

// TxConfig returns App's tx config.
func (app *App) TxConfig() client.TxConfig {
	return app.txConfig
}

// GetKey returns the KVStoreKey for the provided store key.
func (app *App) GetKey(storeKey string) *storetypes.KVStoreKey {
	return app.keys[storeKey]
}

// GetStoreKeys returns all the registered store keys.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetStoreKeys() []storetypes.StoreKey {
	keys := make([]storetypes.StoreKey, 0, len(app.keys))
	for _, k := range app.keys {
		keys = append(keys, k)
	}
	return keys
}

// DefaultGenesis returns a default genesis from the registered
// AppModuleBasic's, marshalled to json.RawMessage per module.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) DefaultGenesis() map[string]json.RawMessage {
	return app.BasicModuleManager.DefaultGenesis(app.appCodec)
}

// Configurator returns the app's module configurator.
func (app *App) Configurator() module.Configurator {
	return app.configurator
}

// SimulationManager implements the SimulationApp interface.
func (app *App) SimulationManager() *module.SimulationManager {
	return app.sm
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *App) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	cmtservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	app.BasicModuleManager.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	if err := server.RegisterSwaggerAPI(apiSvr.ClientCtx, apiSvr.Router, apiConfig.Swagger); err != nil {
		panic(err)
	}
	docs.RegisterOpenAPIService(Name, apiSvr.Router)
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *App) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.GRPCQueryRouter(), clientCtx, app.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *App) RegisterTendermintService(clientCtx client.Context) {
	cmtservice.RegisterTendermintService(
		clientCtx,
		app.GRPCQueryRouter(),
		app.interfaceRegistry,
		app.Query,
	)
}

func (app *App) RegisterNodeService(clientCtx client.Context, cfg config.Config) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg)
}

// AutoCliOpts returns the autocli options for the app, built from the
// concrete module manager since this app is no longer depinject-wired.
func (app *App) AutoCliOpts() autocli.AppOptions {
	modules := make(map[string]appmodule.AppModule, len(app.ModuleManager.Modules))
	for name, m := range app.ModuleManager.Modules {
		if appModule, ok := m.(appmodule.AppModule); ok {
			modules[name] = appModule
		}
	}

	return autocli.AppOptions{
		Modules:               modules,
		ModuleOptions:         runtimeservices.ExtractAutoCLIOptions(app.ModuleManager.Modules),
		AddressCodec:          authcodec.NewBech32Codec(AccountAddressPrefix),
		ValidatorAddressCodec: authcodec.NewBech32Codec(AccountAddressPrefix + "valoper"),
		ConsensusAddressCodec: authcodec.NewBech32Codec(AccountAddressPrefix + "valcons"),
	}
}

// wasmAppModuleBasic overrides wasm's stock DefaultGenesis (which defaults
// to fully permissionless code upload) with the Phase D spec's requirement:
// gov-gated code upload, permissionless instantiate of already-approved
// code IDs. See modules/wasm/SPEC.md.
type wasmAppModuleBasic struct {
	wasm.AppModuleBasic
}

func (wasmAppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(&wasmtypes.GenesisState{
		Params: wasmtypes.Params{
			CodeUploadAccess:             wasmtypes.AccessTypeAnyOfAddresses.With(authtypes.NewModuleAddress(govtypes.ModuleName)),
			InstantiateDefaultPermission: wasmtypes.AccessTypeEverybody,
		},
	})
}

// bankAppModuleBasic overrides bank's stock DefaultGenesis to additionally
// register denom metadata for ulcv. cosmos/evm's x/vm module derives the
// EVM coin's decimals from this metadata at genesis (looking up the
// DenomUnit matching the Display field) — without it, InitGenesis panics
// with "denom metadata could not be found". See modules/evm/SPEC.md.
type bankAppModuleBasic struct {
	bank.AppModuleBasic
}

func (bankAppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(bankGenesisWithDenomMetadata())
}

// bankGenesisWithDenomMetadata is shared by bankAppModuleBasic (real
// genesis, via `lcvd init`) and bankAppModuleSimulation (the randomized
// block simulator's genesis) — both need the ulcv/lcv denom metadata,
// since x/vm's InitGenesis panics without it regardless of which path
// produced the surrounding genesis.
func bankGenesisWithDenomMetadata() *banktypes.GenesisState {
	genState := banktypes.DefaultGenesisState()
	genState.DenomMetadata = append(genState.DenomMetadata, banktypes.Metadata{
		Description: "The native staking and gas token of LastcoinVision",
		Base:        EVMBaseDenom,
		Display:     EVMDisplayDenom,
		Name:        "LastcoinVision",
		Symbol:      "LCV",
		DenomUnits: []*banktypes.DenomUnit{
			{Denom: EVMBaseDenom, Exponent: 0, Aliases: []string{"microlcv"}},
			{Denom: EVMDisplayDenom, Exponent: 6},
		},
	})
	return genState
}

// bankAppModuleSimulation overrides bank's own randomized simulation
// genesis (which knows nothing about this chain's denom) with the fixed,
// metadata-bearing genesis above — registered in the simulation manager's
// overrideModules map in New(). Embeds the real bank.AppModule so
// WeightedOperations/RegisterStoreDecoder (bank's own randomized
// send/multisend operations) are unaffected; only genesis generation
// changes.
type bankAppModuleSimulation struct {
	bank.AppModule
}

func (bankAppModuleSimulation) GenerateGenesisState(simState *module.SimulationState) {
	simState.GenState[banktypes.ModuleName] = simState.Cdc.MustMarshalJSON(bankGenesisWithDenomMetadata())
}

// evmAppModuleBasic overrides the vm module's stock DefaultGenesis to set
// this chain's EVM denom (the existing ulcv bond denom, unchanged) and its
// 18-decimal extended denom (alcv, tracked internally by x/precisebank —
// see modules/evm/SPEC.md), and to enable every default static precompile
// and preinstall, matching cosmos/evm's own reference app.
type evmAppModuleBasic struct {
	vm.AppModuleBasic
}

func (evmAppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(evmGenesisState())
}

func evmGenesisState() *evmtypes.GenesisState {
	genState := evmtypes.DefaultGenesisState()
	genState.Params.EvmDenom = EVMBaseDenom
	genState.Params.ExtendedDenomOptions = &evmtypes.ExtendedDenomOptions{ExtendedDenom: EVMExtendedDenom}
	genState.Params.ActiveStaticPrecompiles = evmtypes.AvailableStaticPrecompiles
	genState.Preinstalls = evmtypes.DefaultPreinstalls
	return genState
}

// evmAppModuleSimulation overrides the vm module's own randomized
// simulation genesis (which would otherwise pick a random, non-ulcv
// EvmDenom, among other things) with the same fixed genesis DefaultGenesis
// uses. See bankAppModuleSimulation for why this override exists.
type evmAppModuleSimulation struct {
	vm.AppModule
}

func (evmAppModuleSimulation) GenerateGenesisState(simState *module.SimulationState) {
	simState.GenState[evmtypes.ModuleName] = simState.Cdc.MustMarshalJSON(evmGenesisState())
}

// evmFeemarketAppModuleBasic overrides the EVM feemarket module's stock
// DefaultGenesis, whose DefaultBaseFee (1_000_000_000) is calibrated for an
// 18-decimal-native EVM chain — i.e. "1 gwei" real value. On this chain the
// EVM's native denom (ulcv) has only 6 decimals, so that same raw number
// means 1_000_000_000 ulcv per gas unit (1,000 LCV/gas — found live: every
// tx failed with "insufficient funds for gas * price + value" because the
// EVM base fee had been priced a trillion times too high). Scaling the
// default down by 10^(18-6) gives the same real-world gas price the
// upstream default intends. See modules/evm/SPEC.md.
type evmFeemarketAppModuleBasic struct {
	evmfeemarket.AppModuleBasic
}

func (evmFeemarketAppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(evmFeemarketGenesisState())
}

func evmFeemarketGenesisState() *evmfeemarkettypes.GenesisState {
	genState := evmfeemarkettypes.DefaultGenesisState()
	decimalGap := math.LegacyNewDec(10).Power(18 - 6) // 10^12
	genState.Params.BaseFee = evmfeemarkettypes.DefaultBaseFee.Quo(decimalGap)
	return genState
}

// evmFeemarketAppModuleSimulation overrides the EVM feemarket module's own
// randomized simulation genesis with the same rescaled-base-fee genesis
// DefaultGenesis uses — otherwise the simulator hits the exact
// insufficient-funds bug found live (see DECISIONS.md).
type evmFeemarketAppModuleSimulation struct {
	evmfeemarket.AppModule
}

func (evmFeemarketAppModuleSimulation) GenerateGenesisState(simState *module.SimulationState) {
	simState.GenState[evmfeemarkettypes.ModuleName] = simState.Cdc.MustMarshalJSON(evmFeemarketGenesisState())
}

// onPendingTx notifies every listener registered via
// RegisterPendingTxListener (the JSON-RPC server's eth_subscribe
// "newPendingTransactions" feed) that a new Ethereum tx hash has entered
// the mempool.
func (app *App) onPendingTx(hash common.Hash) {
	for _, listener := range app.pendingTxListeners {
		listener(hash)
	}
}

// RegisterPendingTxListener is used by the JSON-RPC server to subscribe to
// pending-transaction notifications.
func (app *App) RegisterPendingTxListener(listener func(common.Hash)) {
	app.pendingTxListeners = append(app.pendingTxListeners, listener)
}

// GetMempool implements cosmosevmserver.Application.
func (app *App) GetMempool() sdkmempool.ExtMempool {
	return app.EVMMempool
}

// SetClientCtx implements cosmosevmserver.Application.
func (app *App) SetClientCtx(clientCtx client.Context) {
	app.clientCtx = clientCtx
}

// Close unsubscribes from the CometBFT event bus (if set) and closes the
// mempool and underlying BaseApp.
func (app *App) Close() error {
	var err error
	if app.EVMMempool != nil {
		err = app.EVMMempool.Close()
	}
	return errors.Join(err, app.BaseApp.Close())
}

// configureEVMMempool wires cosmos/evm's experimental priority-nonce
// mempool, required for correct EVM tx ordering/replacement semantics
// (nonce gaps, gas-price-based eviction) that the SDK's default mempool
// doesn't provide. Mirrors cosmos/evm's own evmd/mempool.go exactly.
func (app *App) configureEVMMempool(appOpts servertypes.AppOptions, logger log.Logger) error {
	if evmtypes.GetChainConfig() == nil {
		logger.Debug("evm chain config is not set, skipping mempool configuration")
		return nil
	}

	cosmosPoolMaxTx := evmconfig.GetCosmosPoolMaxTx(appOpts, logger)
	if cosmosPoolMaxTx < 0 {
		logger.Debug("app-side mempool is disabled, skipping evm mempool configuration")
		return nil
	}

	mempoolConfig := &evmmempool.EVMMempoolConfig{
		AnteHandler:      app.AnteHandler(),
		LegacyPoolConfig: evmconfig.GetLegacyPoolConfig(appOpts, logger),
		BlockGasLimit:    evmconfig.GetBlockGasLimit(appOpts, logger),
		MinTip:           evmconfig.GetMinTip(appOpts, logger),
	}

	evmMempool := evmmempool.NewExperimentalEVMMempool(
		app.CreateQueryContext,
		logger,
		app.EVMKeeper,
		app.EVMFeeMarketKeeper,
		app.txConfig,
		app.clientCtx,
		mempoolConfig,
		cosmosPoolMaxTx,
	)
	app.EVMMempool = evmMempool
	app.SetMempool(evmMempool)
	app.SetCheckTxHandler(evmmempool.NewCheckTxHandler(evmMempool))

	abciProposalHandler := baseapp.NewDefaultProposalHandler(evmMempool, app.BaseApp)
	abciProposalHandler.SetSignerExtractionAdapter(
		evmmempool.NewEthSignerExtractionAdapter(
			sdkmempool.NewDefaultSignerExtractionAdapter(),
		),
	)
	app.SetPrepareProposal(abciProposalHandler.PrepareProposalHandler())

	return nil
}
