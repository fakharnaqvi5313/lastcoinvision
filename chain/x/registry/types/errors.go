package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/registry module sentinel errors
var (
	ErrInvalidSigner              = sdkerrors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrNotJurisdictionAuthority   = sdkerrors.Register(ModuleName, 1101, "signer is not the jurisdiction authority for this jurisdiction code")
	ErrJurisdictionAuthorityUnset = sdkerrors.Register(ModuleName, 1102, "no jurisdiction authority is set for this jurisdiction code")
	ErrAssetNotFound              = sdkerrors.Register(ModuleName, 1103, "asset not found")
	ErrAssetFrozen                = sdkerrors.Register(ModuleName, 1104, "asset is frozen")
	ErrNotAssetOwnerOrAuthority   = sdkerrors.Register(ModuleName, 1105, "signer is neither the asset's current owner nor its jurisdiction authority")
	ErrPartyNotAttested           = sdkerrors.Register(ModuleName, 1106, "both current and new owner must hold at least the attested identity tier")
)
