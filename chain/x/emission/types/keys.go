package types

const (
	// ModuleName defines the module name
	ModuleName = "emission"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_emission"
)

var (
	ParamsKey = []byte("p_emission")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
