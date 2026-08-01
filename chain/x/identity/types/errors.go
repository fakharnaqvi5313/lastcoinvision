package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/identity module sentinel errors
var (
	ErrInvalidSigner       = sdkerrors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrDIDAlreadyExists    = sdkerrors.Register(ModuleName, 1101, "a DID already exists for this address")
	ErrDIDNotFound         = sdkerrors.Register(ModuleName, 1102, "no DID exists for this address")
	ErrDIDDeactivated      = sdkerrors.Register(ModuleName, 1103, "DID is deactivated")
	ErrIssuerAlreadyExists = sdkerrors.Register(ModuleName, 1104, "issuer is already authorized")
	ErrIssuerNotFound      = sdkerrors.Register(ModuleName, 1105, "issuer is not authorized")
	ErrCredentialNotFound  = sdkerrors.Register(ModuleName, 1106, "credential not found")
	ErrNotCredentialIssuer = sdkerrors.Register(ModuleName, 1107, "only the issuing address (or module authority) may revoke this credential")
	ErrCredentialRevoked   = sdkerrors.Register(ModuleName, 1108, "credential is already revoked")
	ErrNotCredentialHolder = sdkerrors.Register(ModuleName, 1109, "presenter does not match credential subject")
)
