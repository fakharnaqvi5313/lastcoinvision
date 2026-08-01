package types

const (
	// ModuleName defines the module name
	ModuleName = "registry"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_registry"
)

var (
	ParamsKey = []byte("p_registry")

	// AssetKeyPrefix + <id> -> Asset
	AssetKeyPrefix = []byte("asset/")

	// TransferRecordKeyPrefix + <asset_id> + "/" + <seq> -> TransferRecord,
	// zero-padded seq so lexicographic iteration order equals chronological
	// order for the transfer-history query.
	TransferRecordKeyPrefix = []byte("transfer/")

	// TransferSeqKeyPrefix + <asset_id> -> next transfer sequence number
	// (uint64 big-endian), so each asset's history can be appended to
	// without re-scanning it first.
	TransferSeqKeyPrefix = []byte("transfer_seq/")

	// JurisdictionAuthorityKeyPrefix + <jurisdiction_code> -> JurisdictionAuthority
	JurisdictionAuthorityKeyPrefix = []byte("jurisdiction_authority/")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

// AssetKey returns the store key for an asset by id.
func AssetKey(id string) []byte {
	return append(AssetKeyPrefix, []byte(id)...)
}

// TransferRecordKey returns the store key for one transfer-history entry.
func TransferRecordKey(assetID string, seq uint64) []byte {
	key := append(TransferRecordKeyPrefix, []byte(assetID)...)
	key = append(key, '/')
	seqBz := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		seqBz[i] = byte(seq)
		seq >>= 8
	}
	return append(key, seqBz...)
}

// TransferRecordPrefix returns the iteration prefix for all of an asset's
// transfer-history entries.
func TransferRecordPrefix(assetID string) []byte {
	key := append(TransferRecordKeyPrefix, []byte(assetID)...)
	return append(key, '/')
}

// TransferSeqKey returns the store key for an asset's next transfer
// sequence number.
func TransferSeqKey(assetID string) []byte {
	return append(TransferSeqKeyPrefix, []byte(assetID)...)
}

// JurisdictionAuthorityKey returns the store key for a jurisdiction's
// authority record.
func JurisdictionAuthorityKey(jurisdictionCode string) []byte {
	return append(JurisdictionAuthorityKeyPrefix, []byte(jurisdictionCode)...)
}
