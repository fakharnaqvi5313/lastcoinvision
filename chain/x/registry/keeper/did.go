package keeper

import "strings"

const didPrefix = "did:lcv:"

// didToAddress strips the "did:lcv:" prefix from a DID string, returning
// the bech32 address it identifies. Asset.owner and MsgTransferAsset's
// new_owner are DIDs (per the top-level spec's "owner DID" wording), but
// message signers and identity-tier lookups both operate on raw addresses,
// so this conversion happens at every boundary between the two.
func didToAddress(did string) string {
	return strings.TrimPrefix(did, didPrefix)
}
