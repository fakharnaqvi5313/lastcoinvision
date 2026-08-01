package keeper_test

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

func unixTime(sec int64) time.Time {
	return time.Unix(sec, 0).UTC()
}

// sha256Hex mirrors the unexported keeper.commitmentHash algorithm
// (sha256(choice + "|" + salt)) so tests can construct valid commitments
// from outside the keeper package.
func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}
