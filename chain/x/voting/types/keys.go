package types

const (
	// ModuleName defines the module name
	ModuleName = "voting"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_voting"
)

var (
	ParamsKey = []byte("p_voting")

	// ElectionKeyPrefix + <id> -> Election
	ElectionKeyPrefix = []byte("election/")

	// NextElectionSeqKey -> next election id sequence number (uint64 big-endian)
	NextElectionSeqKey = []byte("next_election_seq")

	// RegisteredVoterKeyPrefix + <election_id> + "/" + <voter> -> RegisteredVoter
	RegisteredVoterKeyPrefix = []byte("registered_voter/")

	// VoteCommitmentKeyPrefix + <election_id> + "/" + <nullifier> -> VoteCommitment
	VoteCommitmentKeyPrefix = []byte("vote_commitment/")

	// ElectoralAuthorityKeyPrefix + <address> -> ElectoralAuthority
	ElectoralAuthorityKeyPrefix = []byte("electoral_authority/")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

// ElectionKey returns the store key for an election by id.
func ElectionKey(id string) []byte {
	return append(ElectionKeyPrefix, []byte(id)...)
}

// RegisteredVoterKey returns the store key for one voter's registration in
// one election.
func RegisteredVoterKey(electionID, voter string) []byte {
	key := append(RegisteredVoterKeyPrefix, []byte(electionID)...)
	key = append(key, '/')
	return append(key, []byte(voter)...)
}

// RegisteredVoterPrefix returns the iteration prefix for all of an
// election's registered voters.
func RegisteredVoterPrefix(electionID string) []byte {
	key := append(RegisteredVoterKeyPrefix, []byte(electionID)...)
	return append(key, '/')
}

// VoteCommitmentKey returns the store key for one voter's commitment
// (looked up by their deterministic nullifier) in one election.
func VoteCommitmentKey(electionID, nullifier string) []byte {
	key := append(VoteCommitmentKeyPrefix, []byte(electionID)...)
	key = append(key, '/')
	return append(key, []byte(nullifier)...)
}

// VoteCommitmentPrefix returns the iteration prefix for all of an
// election's vote commitments.
func VoteCommitmentPrefix(electionID string) []byte {
	key := append(VoteCommitmentKeyPrefix, []byte(electionID)...)
	return append(key, '/')
}

// ElectoralAuthorityKey returns the store key for an electoral authority by
// address.
func ElectoralAuthorityKey(address string) []byte {
	return append(ElectoralAuthorityKeyPrefix, []byte(address)...)
}
