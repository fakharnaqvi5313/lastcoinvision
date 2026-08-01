package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/taxrail module sentinel errors
var (
	ErrInvalidSigner      = sdkerrors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrInvalidRate        = sdkerrors.Register(ModuleName, 1101, "tax rate must be between 0 and 1 inclusive")
	ErrNoTaxRule          = sdkerrors.Register(ModuleName, 1102, "no tax rule registered for this jurisdiction code")
	ErrInsufficientLedger = sdkerrors.Register(ModuleName, 1103, "jurisdiction has not collected enough tax to spend this amount")
	ErrUnsupportedDenom   = sdkerrors.Register(ModuleName, 1104, "taxrail only operates in ulcv")
)
