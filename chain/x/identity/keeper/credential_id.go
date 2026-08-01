package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// credentialID deterministically derives a credential's id from its
// issuance details, so the same (issuer, subject, type, hash, time) can
// never collide with a different credential and the id itself reveals
// nothing beyond what's already public in these fields.
func credentialID(issuer, subject, credentialType, credentialHash string, issuedAt int64) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%s|%s|%d", issuer, subject, credentialType, credentialHash, issuedAt)))
	return hex.EncodeToString(sum[:])
}
