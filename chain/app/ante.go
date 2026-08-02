package app

import (
	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	txsigning "cosmossdk.io/x/tx/signing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	ibcante "github.com/cosmos/ibc-go/v10/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v10/modules/core/keeper"

	cosmosevmante "github.com/cosmos/evm/ante"
	cosmosante "github.com/cosmos/evm/ante/cosmos"
	evmmonoante "github.com/cosmos/evm/ante/evm"
	anteinterfaces "github.com/cosmos/evm/ante/interfaces"
	evmtypes "github.com/cosmos/evm/x/vm/types"

	feemarketante "github.com/lastcoinvision/lcv/x/feemarket/ante"
	feemarketkeeper "github.com/lastcoinvision/lcv/x/feemarket/keeper"
	feemarkettypes "github.com/lastcoinvision/lcv/x/feemarket/types"
)

// BankKeeper is the union of what the standard ante decorators, our own
// feemarket decorator, and cosmos/evm's decorators each need.
type BankKeeper interface {
	authtypes.BankKeeper
	feemarkettypes.BankKeeper
	anteinterfaces.BankKeeper
}

// AccountKeeper is the union of what the standard ante decorators and
// cosmos/evm's decorators each need (the latter's is wider — it also needs
// GetSequence and AddressCodec).
type AccountKeeper interface {
	ante.AccountKeeper
	anteinterfaces.AccountKeeper
}

// HandlerOptions mirrors cosmos/evm's ante.HandlerOptions, plus our own
// feemarket keeper. It builds an ante handler that dispatches on tx type
// (see NewAnteHandler): a MsgEthereumTx goes through cosmos/evm's own mono
// decorator (which enforces its own, EVM-native feemarket, mandatory for
// EIP-1559-style gas pricing on the EVM side and for JSON-RPC
// compatibility); every other transaction — including our own registry,
// voting, taxrail, wasm messages, and standard SDK messages — goes through
// a decorator chain that keeps this chain's original Phase A x/feemarket
// decorator as the fee mechanism, unchanged from before EVM was added. See
// DECISIONS.md for why the two fee markets are kept deliberately separate
// rather than collapsing onto one.
type HandlerOptions struct {
	AccountKeeper          AccountKeeper
	BankKeeper             BankKeeper
	FeemarketKeeper        feemarketkeeper.Keeper
	ExtensionOptionChecker ante.ExtensionOptionChecker
	FeegrantKeeper         ante.FeegrantKeeper
	SignModeHandler        *txsigning.HandlerMap
	SigGasConsumer         func(meter storetypes.GasMeter, sig signing.SignatureV2, params authtypes.Params) error

	IBCKeeper          *ibckeeper.Keeper
	EvmKeeper          anteinterfaces.EVMKeeper
	EvmFeeMarketKeeper anteinterfaces.FeeMarketKeeper
	MaxTxGasWanted     uint64
	PendingTxListener  cosmosevmante.PendingTxListener
}

// Validate checks if the keepers are defined.
func (options HandlerOptions) Validate() error {
	if options.AccountKeeper == nil {
		return errorsmod.Wrap(sdkerrors.ErrLogic, "account keeper is required for ante builder")
	}
	if options.BankKeeper == nil {
		return errorsmod.Wrap(sdkerrors.ErrLogic, "bank keeper is required for ante builder")
	}
	if options.SignModeHandler == nil {
		return errorsmod.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}
	if options.IBCKeeper == nil {
		return errorsmod.Wrap(sdkerrors.ErrLogic, "ibc keeper is required for ante builder")
	}
	if options.EvmKeeper == nil {
		return errorsmod.Wrap(sdkerrors.ErrLogic, "evm keeper is required for ante builder")
	}
	if options.EvmFeeMarketKeeper == nil {
		return errorsmod.Wrap(sdkerrors.ErrLogic, "evm feemarket keeper is required for ante builder")
	}
	return nil
}

// NewAnteHandler returns an ante handler that routes a transaction to
// either the EVM-native handler (for a MsgEthereumTx carrying the
// ExtensionOptionsEthereumTx extension option) or this chain's own Cosmos
// tx handler, mirroring cosmos/evm's own ante.NewAnteHandler dispatch
// (github.com/cosmos/evm/ante) — forked here (rather than imported)
// because that package's two branch-builder functions are unexported, and
// this chain needs the Cosmos branch to use its own feemarket decorator
// instead of cosmos/evm's.
func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if err := options.Validate(); err != nil {
		return nil, err
	}

	return func(ctx sdk.Context, tx sdk.Tx, sim bool) (newCtx sdk.Context, err error) {
		var anteHandler sdk.AnteHandler

		txWithExtensions, ok := tx.(ante.HasExtensionOptionsTx)
		if ok {
			opts := txWithExtensions.GetExtensionOptions()
			if len(opts) > 0 {
				switch typeURL := opts[0].GetTypeUrl(); typeURL {
				case "/cosmos.evm.vm.v1.ExtensionOptionsEthereumTx":
					anteHandler = newEvmAnteHandler(ctx, options)
				case "/cosmos.evm.ante.v1.ExtensionOptionDynamicFeeTx":
					anteHandler = newCosmosAnteHandler(ctx, options)
				default:
					return ctx, errorsmod.Wrapf(
						sdkerrors.ErrUnknownExtensionOptions,
						"rejecting tx with unsupported extension option: %s", typeURL,
					)
				}
				return anteHandler(ctx, tx, sim)
			}
		}

		anteHandler = newCosmosAnteHandler(ctx, options)
		return anteHandler(ctx, tx, sim)
	}, nil
}

// newCosmosAnteHandler builds the decorator chain for every non-EVM
// transaction. Identical to cosmos/evm's own newCosmosAnteHandler except
// cosmosante.NewMinGasPriceDecorator + ante.NewDeductFeeDecorator (which
// together enforce cosmos/evm's own feemarket) are replaced with this
// chain's original feemarketante.NewDecorator (see app/ante.go's history
// pre-Phase-D and x/feemarket/ante).
func newCosmosAnteHandler(ctx sdk.Context, options HandlerOptions) sdk.AnteHandler {
	// cosmos/evm's own SigVerificationGasConsumer (not the SDK's
	// DefaultSigVerificationGasConsumer) is required here because accounts
	// on this chain may hold an eth_secp256k1 pubkey (see
	// hd.EthSecp256k1Option() in cmd/lcvd/cmd/root.go) — the standard
	// consumer's pubkey-type switch doesn't have a case for it and every
	// tx signed by such an account would fail with "unrecognized public
	// key type", including gentx delivery at genesis.
	sigGasConsumer := options.SigGasConsumer
	if sigGasConsumer == nil {
		sigGasConsumer = cosmosevmante.SigVerificationGasConsumer
	}

	evmFeemarketParams := options.EvmFeeMarketKeeper.GetParams(ctx)

	return sdk.ChainAnteDecorators(
		cosmosante.NewRejectMessagesDecorator(), // reject MsgEthereumTxs outside the EVM branch
		cosmosante.NewAuthzLimiterDecorator( // disable Msg types that must not appear inside an authz.MsgExec
			sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{}),
			sdk.MsgTypeURL(&sdkvesting.MsgCreateVestingAccount{}),
		),
		ante.NewSetUpContextDecorator(),
		ante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		feemarketante.NewDecorator(options.FeemarketKeeper, options.BankKeeper),
		ante.NewSetPubKeyDecorator(options.AccountKeeper), // must run before all signature verification decorators
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, sigGasConsumer),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
		ibcante.NewRedundantRelayDecorator(options.IBCKeeper),
		// Tracks gas used against the block gas target even for non-EVM
		// txs, since cosmos/evm's own feemarket's EIP-1559 base-fee
		// adjustment (for EVM txs) needs total per-block gas usage
		// regardless of which fee mechanism actually charged for it.
		evmmonoante.NewGasWantedDecorator(options.EvmKeeper, options.EvmFeeMarketKeeper, &evmFeemarketParams),
	)
}

// newEvmAnteHandler builds the decorator chain for a MsgEthereumTx,
// unmodified from cosmos/evm's own newMonoEVMAnteHandler (both pieces it
// uses are exported, so no fork was needed here — only the dispatcher and
// the Cosmos branch above needed forking).
func newEvmAnteHandler(ctx sdk.Context, options HandlerOptions) sdk.AnteHandler {
	evmParams := options.EvmKeeper.GetParams(ctx)
	feemarketParams := options.EvmFeeMarketKeeper.GetParams(ctx)

	return sdk.ChainAnteDecorators(
		evmmonoante.NewEVMMonoDecorator(
			options.AccountKeeper,
			options.EvmFeeMarketKeeper,
			options.EvmKeeper,
			options.MaxTxGasWanted,
			&evmParams,
			&feemarketParams,
		),
		cosmosevmante.NewTxListenerDecorator(options.PendingTxListener),
	)
}

func mustFeeMarketParams(options HandlerOptions) *interface {
} {
	return nil
}
