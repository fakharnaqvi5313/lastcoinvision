package types

import "context"

// ZKVerifier is the extension point for zero-knowledge selective-disclosure
// credential proofs. MsgPresentCredential today only performs a signature
// check (the tx signer must equal the credential's subject, already
// enforced by standard Cosmos SDK tx authentication before the handler
// runs) — VerifyProof is not yet called from the message handler and this
// stub implementation always returns true.
//
// TODO(identity): once a concrete ZK scheme is chosen (e.g. a Groth16 or
// PLONK circuit proving a predicate over the off-chain credential document
// without revealing it), wire a real implementation of this interface into
// the keeper and have MsgPresentCredential call VerifyProof(ctx, cred,
// proof) instead of relying solely on tx-signer identity. See
// modules/identity/SPEC.md's Limitations section.
type ZKVerifier interface {
	// VerifyProof checks proof against cred, returning whether it's valid.
	VerifyProof(ctx context.Context, cred Credential, proof []byte) (bool, error)
}

// StubZKVerifier is the placeholder ZKVerifier used until a real scheme is
// implemented. It performs no cryptographic verification at all.
type StubZKVerifier struct{}

var _ ZKVerifier = StubZKVerifier{}

// VerifyProof always returns true: presenting a credential today is
// authenticated entirely by the transaction signature (presenter ==
// subject), not by this proof. See the TODO on ZKVerifier.
func (StubZKVerifier) VerifyProof(_ context.Context, _ Credential, _ []byte) (bool, error) {
	return true, nil
}
