package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/voting module sentinel errors
var (
	ErrInvalidSigner          = sdkerrors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrNotElectoralAuthority  = sdkerrors.Register(ModuleName, 1101, "signer is not an authorized electoral authority")
	ErrInvalidElectionWindows = sdkerrors.Register(ModuleName, 1102, "election windows must be non-overlapping and strictly increasing: registration, voting, reveal")
	ErrElectionNotFound       = sdkerrors.Register(ModuleName, 1103, "election not found")
	ErrNotRegistrationWindow  = sdkerrors.Register(ModuleName, 1104, "not within the election's registration window")
	ErrNotEligible            = sdkerrors.Register(ModuleName, 1105, "voter does not hold an active eligibility credential from an authorized issuer")
	ErrAlreadyRegistered      = sdkerrors.Register(ModuleName, 1106, "voter is already registered for this election")
	ErrNotRegistered          = sdkerrors.Register(ModuleName, 1107, "voter is not registered for this election")
	ErrNotVotingWindow        = sdkerrors.Register(ModuleName, 1108, "not within the election's voting window")
	ErrAlreadyCommitted       = sdkerrors.Register(ModuleName, 1109, "a vote has already been committed for this voter and election")
	ErrNotRevealWindow        = sdkerrors.Register(ModuleName, 1110, "not within the election's reveal window")
	ErrNoCommitment           = sdkerrors.Register(ModuleName, 1111, "no vote commitment found for this voter and election")
	ErrAlreadyRevealed        = sdkerrors.Register(ModuleName, 1112, "this vote has already been revealed")
	ErrCommitmentMismatch     = sdkerrors.Register(ModuleName, 1113, "revealed choice and salt do not hash to the stored commitment")
	ErrInvalidChoice          = sdkerrors.Register(ModuleName, 1114, "choice is not one of the election's options")
)
