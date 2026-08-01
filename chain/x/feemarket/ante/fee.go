package ante

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	feemarketkeeper "github.com/lastcoinvision/lcv/x/feemarket/keeper"
	feemarkettypes "github.com/lastcoinvision/lcv/x/feemarket/types"
)

// Decorator enforces and collects the EIP-1559-style dynamic base fee,
// replacing the standard Cosmos SDK auth.DeductFeeDecorator in the ante
// chain (see app/ante.go). It requires every fee to be paid entirely in
// params.FeeDenom and to be at least base_fee * gasWanted; any amount paid
// above that minimum is treated as a validator tip. Of the base-fee portion
// specifically, burn_percentage is burned and the remainder — plus the full
// tip — goes to the fee collector for x/distribution to allocate.
type Decorator struct {
	feemarketKeeper feemarketkeeper.Keeper
	bankKeeper      feemarkettypes.BankKeeper
}

func NewDecorator(feemarketKeeper feemarketkeeper.Keeper, bankKeeper feemarkettypes.BankKeeper) Decorator {
	return Decorator{feemarketKeeper: feemarketKeeper, bankKeeper: bankKeeper}
}

func (d Decorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "tx must implement sdk.FeeTx")
	}

	// InitChain-time and simulated txs skip fee enforcement/collection,
	// matching standard Cosmos SDK ante behavior.
	if ctx.BlockHeight() == 0 || simulate {
		return next(ctx, tx, simulate)
	}

	params := d.feemarketKeeper.GetParams(ctx)
	feeCoins := feeTx.GetFee()

	for _, c := range feeCoins {
		if c.Denom != params.FeeDenom {
			return ctx, errorsmod.Wrapf(sdkerrors.ErrInvalidCoins,
				"fee must be paid entirely in %s, got %s", params.FeeDenom, feeCoins)
		}
	}

	required := d.feemarketKeeper.RequiredFee(ctx, feeTx.GetGas())
	paid := feeCoins.AmountOf(params.FeeDenom)
	if paid.LT(required.Amount) {
		return ctx, errorsmod.Wrapf(sdkerrors.ErrInsufficientFee,
			"insufficient fee: got %s%s, required at least %s%s", paid, params.FeeDenom, required.Amount, params.FeeDenom)
	}

	payer := feeTx.FeePayer()
	paidCoins := sdk.NewCoins(sdk.NewCoin(params.FeeDenom, paid))
	if err := d.bankKeeper.SendCoinsFromAccountToModule(ctx, payer, feemarkettypes.ModuleName, paidCoins); err != nil {
		return ctx, errorsmod.Wrap(err, "fee market: collect fee")
	}

	burn, _ := d.feemarketKeeper.BurnSplit(ctx, required.Amount)
	if burn.IsPositive() {
		if err := d.bankKeeper.BurnCoins(ctx, feemarkettypes.ModuleName, sdk.NewCoins(sdk.NewCoin(params.FeeDenom, burn))); err != nil {
			return ctx, errorsmod.Wrap(err, "fee market: burn base fee")
		}
	}

	forward := paid.Sub(burn)
	if forward.IsPositive() {
		forwardCoins := sdk.NewCoins(sdk.NewCoin(params.FeeDenom, forward))
		if err := d.bankKeeper.SendCoinsFromModuleToModule(ctx, feemarkettypes.ModuleName, authtypes.FeeCollectorName, forwardCoins); err != nil {
			return ctx, errorsmod.Wrap(err, "fee market: forward fee to collector")
		}
	}

	return next(ctx, tx, simulate)
}
