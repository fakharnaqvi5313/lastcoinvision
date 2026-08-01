package types

const (
	// ModuleName defines the module name
	ModuleName = "identity"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_identity"
)

const (
	// DefaultAttestedCredentialType is the credential_type value that
	// grants the "attested" tier by default.
	DefaultAttestedCredentialType = "kyc"
)

var (
	ParamsKey = []byte("p_identity")

	// DIDKeyPrefix + <address> -> DIDDocument
	DIDKeyPrefix = []byte("did/")

	// CredentialKeyPrefix + <id> -> Credential
	CredentialKeyPrefix = []byte("credential/")

	// CredentialBySubjectKeyPrefix + <subject> + "/" + <id> -> presence marker,
	// a secondary index letting TierOf and CredentialsBySubject iterate a
	// subject's credentials without scanning the whole credential set.
	CredentialBySubjectKeyPrefix = []byte("credential_by_subject/")

	// IssuerKeyPrefix + <address> -> Issuer
	IssuerKeyPrefix = []byte("issuer/")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

// DIDKey returns the store key for the DID document controlled by address.
func DIDKey(address string) []byte {
	return append(DIDKeyPrefix, []byte(address)...)
}

// CredentialKey returns the store key for a credential by id.
func CredentialKey(id string) []byte {
	return append(CredentialKeyPrefix, []byte(id)...)
}

// CredentialBySubjectKey returns the secondary-index store key for a
// (subject, credential id) pair.
func CredentialBySubjectKey(subject, id string) []byte {
	key := append(CredentialBySubjectKeyPrefix, []byte(subject)...)
	key = append(key, '/')
	return append(key, []byte(id)...)
}

// CredentialBySubjectPrefix returns the iteration prefix for all of a
// subject's credential-id index entries.
func CredentialBySubjectPrefix(subject string) []byte {
	key := append(CredentialBySubjectKeyPrefix, []byte(subject)...)
	return append(key, '/')
}

// IssuerKey returns the store key for an issuer by address.
func IssuerKey(address string) []byte {
	return append(IssuerKeyPrefix, []byte(address)...)
}
