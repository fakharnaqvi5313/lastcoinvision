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

	feemarketante "github.com/lastcoinvision/lcv/x/feemarket/ante"
	feemarketkeeper "github.com/lastcoinvision/lcv/x/feemarket/keeper"
	feemarkettypes "github.com/lastcoinvision/lcv/x/feemarket/types"
)

// BankKeeper is the union of what the standard ante decorators and the
// feemarket fee decorator each need — the real bank keeper implements all
// of it, but the standard cosmos-sdk ante.HandlerOptions' narrower
// authtypes.BankKeeper doesn't include BurnCoins.
type BankKeeper interface {
	authtypes.BankKeeper
	feemarkettypes.BankKeeper
}

// HandlerOptions mirrors cosmos-sdk's ante.HandlerOptions, plus the
// feemarket keeper needed to build feemarketante.Decorator.
type HandlerOptions struct {
	AccountKeeper          ante.AccountKeeper
	BankKeeper             BankKeeper
	FeemarketKeeper        feemarketkeeper.Keeper
	ExtensionOptionChecker ante.ExtensionOptionChecker
	FeegrantKeeper         ante.FeegrantKeeper
	SignModeHandler        *txsigning.HandlerMap
	SigGasConsumer         func(meter storetypes.GasMeter, sig signing.SignatureV2, params authtypes.Params) error
}

// NewAnteHandler mirrors cosmos-sdk's ante.NewAnteHandler decorator chain
// exactly, except NewDeductFeeDecorator is replaced with
// feemarketante.NewDecorator so the EIP-1559-style dynamic base fee (with
// partial burn) is what actually gets enforced and collected, instead of
// the standard flat min-gas-price fee deduction.
func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "account keeper is required for ante builder")
	}
	if options.BankKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "bank keeper is required for ante builder")
	}
	if options.SignModeHandler == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}

	sigGasConsumer := options.SigGasConsumer
	if sigGasConsumer == nil {
		sigGasConsumer = ante.DefaultSigVerificationGasConsumer
	}

	anteDecorators := []sdk.AnteDecorator{
		ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		ante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		feemarketante.NewDecorator(options.FeemarketKeeper, options.BankKeeper),
		ante.NewSetPubKeyDecorator(options.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, sigGasConsumer),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}
