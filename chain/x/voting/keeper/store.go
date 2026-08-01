package keeper

import (
	"context"
	"encoding/binary"
	"strconv"

	"github.com/cosmos/cosmos-sdk/runtime"

	"github.com/lastcoinvision/lcv/x/voting/types"
)

// -- Elections --

func (k Keeper) SetElection(ctx context.Context, election types.Election) error {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz, err := k.cdc.Marshal(&election)
	if err != nil {
		return err
	}
	store.Set(types.ElectionKey(election.Id), bz)
	return nil
}

func (k Keeper) GetElection(ctx context.Context, id string) (types.Election, bool) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.ElectionKey(id))
	if bz == nil {
		return types.Election{}, false
	}
	var election types.Election
	k.cdc.MustUnmarshal(bz, &election)
	return election, true
}

func (k Keeper) GetAllElections(ctx context.Context) []*types.Election {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iterator := store.Iterator(types.ElectionKeyPrefix, storeIteratorEnd(types.ElectionKeyPrefix))
	defer iterator.Close()

	var elections []*types.Election
	for ; iterator.Valid(); iterator.Next() {
		var election types.Election
		k.cdc.MustUnmarshal(iterator.Value(), &election)
		elections = append(elections, &election)
	}
	return elections
}

// NextElectionID allocates and persists the next sequential election id.
func (k Keeper) NextElectionID(ctx context.Context) string {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.NextElectionSeqKey)
	var seq uint64
	if bz != nil {
		seq = binary.BigEndian.Uint64(bz)
	}
	next := make([]byte, 8)
	binary.BigEndian.PutUint64(next, seq+1)
	store.Set(types.NextElectionSeqKey, next)
	return "election-" + strconv.FormatUint(seq+1, 10)
}

// -- Registered voters --

func (k Keeper) SetRegisteredVoter(ctx context.Context, rv types.RegisteredVoter) error {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz, err := k.cdc.Marshal(&rv)
	if err != nil {
		return err
	}
	store.Set(types.RegisteredVoterKey(rv.ElectionId, rv.Voter), bz)
	return nil
}

func (k Keeper) IsRegistered(ctx context.Context, electionID, voter string) bool {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return store.Has(types.RegisteredVoterKey(electionID, voter))
}

func (k Keeper) CountRegisteredVoters(ctx context.Context, electionID string) int64 {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefix := types.RegisteredVoterPrefix(electionID)
	iterator := store.Iterator(prefix, storeIteratorEnd(prefix))
	defer iterator.Close()

	var count int64
	for ; iterator.Valid(); iterator.Next() {
		count++
	}
	return count
}

func (k Keeper) GetAllRegisteredVoters(ctx context.Context) []*types.RegisteredVoter {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iterator := store.Iterator(types.RegisteredVoterKeyPrefix, storeIteratorEnd(types.RegisteredVoterKeyPrefix))
	defer iterator.Close()

	var voters []*types.RegisteredVoter
	for ; iterator.Valid(); iterator.Next() {
		var rv types.RegisteredVoter
		k.cdc.MustUnmarshal(iterator.Value(), &rv)
		voters = append(voters, &rv)
	}
	return voters
}

// -- Vote commitments --

func (k Keeper) SetVoteCommitment(ctx context.Context, vc types.VoteCommitment) error {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz, err := k.cdc.Marshal(&vc)
	if err != nil {
		return err
	}
	store.Set(types.VoteCommitmentKey(vc.ElectionId, vc.Nullifier), bz)
	return nil
}

func (k Keeper) GetVoteCommitment(ctx context.Context, electionID, nullifier string) (types.VoteCommitment, bool) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.VoteCommitmentKey(electionID, nullifier))
	if bz == nil {
		return types.VoteCommitment{}, false
	}
	var vc types.VoteCommitment
	k.cdc.MustUnmarshal(bz, &vc)
	return vc, true
}

func (k Keeper) GetVoteCommitmentsByElection(ctx context.Context, electionID string) []types.VoteCommitment {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefix := types.VoteCommitmentPrefix(electionID)
	iterator := store.Iterator(prefix, storeIteratorEnd(prefix))
	defer iterator.Close()

	var commitments []types.VoteCommitment
	for ; iterator.Valid(); iterator.Next() {
		var vc types.VoteCommitment
		k.cdc.MustUnmarshal(iterator.Value(), &vc)
		commitments = append(commitments, vc)
	}
	return commitments
}

func (k Keeper) GetAllVoteCommitments(ctx context.Context) []*types.VoteCommitment {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iterator := store.Iterator(types.VoteCommitmentKeyPrefix, storeIteratorEnd(types.VoteCommitmentKeyPrefix))
	defer iterator.Close()

	var commitments []*types.VoteCommitment
	for ; iterator.Valid(); iterator.Next() {
		var vc types.VoteCommitment
		k.cdc.MustUnmarshal(iterator.Value(), &vc)
		commitments = append(commitments, &vc)
	}
	return commitments
}

// -- Electoral authorities --

func (k Keeper) SetElectoralAuthority(ctx context.Context, auth types.ElectoralAuthority) error {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz, err := k.cdc.Marshal(&auth)
	if err != nil {
		return err
	}
	store.Set(types.ElectoralAuthorityKey(auth.Address), bz)
	return nil
}

func (k Keeper) IsElectoralAuthority(ctx context.Context, address string) bool {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return store.Has(types.ElectoralAuthorityKey(address))
}

func (k Keeper) DeleteElectoralAuthority(ctx context.Context, address string) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store.Delete(types.ElectoralAuthorityKey(address))
}

func (k Keeper) GetAllElectoralAuthorities(ctx context.Context) []*types.ElectoralAuthority {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iterator := store.Iterator(types.ElectoralAuthorityKeyPrefix, storeIteratorEnd(types.ElectoralAuthorityKeyPrefix))
	defer iterator.Close()

	var auths []*types.ElectoralAuthority
	for ; iterator.Valid(); iterator.Next() {
		var auth types.ElectoralAuthority
		k.cdc.MustUnmarshal(iterator.Value(), &auth)
		auths = append(auths, &auth)
	}
	return auths
}

// storeIteratorEnd returns the exclusive end key for iterating all keys
// sharing prefix: prefix with its last byte incremented, which works as
// long as the prefix doesn't end in 0xff (none of ours do).
func storeIteratorEnd(prefix []byte) []byte {
	end := make([]byte, len(prefix))
	copy(end, prefix)
	end[len(end)-1]++
	return end
}
