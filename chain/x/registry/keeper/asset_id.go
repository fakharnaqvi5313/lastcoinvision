package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// assetID deterministically derives an asset's id from its registration
// details, so re-registering identical-looking details at a different time
// never collides with an existing asset.
func assetID(jurisdictionCode, assetType, metadataHash string, registeredAt int64) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%s|%d", jurisdictionCode, assetType, metadataHash, registeredAt)))
	return hex.EncodeToString(sum[:])
}
