package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const didPrefix = "did:lcv:"

// nullifier derives the one-vote-per-DID-per-election nullifier described
// in the top-level spec: sha256(DID + electionID). It's a pure function of
// (voter, electionID), so looking a voter's own commitment back up never
// requires storing or searching by address — recomputing this same value is
// enough. See modules/voting/SPEC.md's Limitations for why this is not a
// ZK nullifier and therefore doesn't hide voter identity, only choice.
func nullifier(voterAddress, electionID string) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s%s|%s", didPrefix, voterAddress, electionID)))
	return hex.EncodeToString(sum[:])
}

// commitmentHash is the commit-reveal hash a voter computes off-chain at
// commit time (sha256(choice + salt)) and the chain recomputes at reveal
// time to check it matches.
func commitmentHash(choice, salt string) string {
	sum := sha256.Sum256([]byte(choice + "|" + salt))
	return hex.EncodeToString(sum[:])
}
