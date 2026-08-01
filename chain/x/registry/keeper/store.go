package keeper

import (
	"context"
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/runtime"

	"github.com/lastcoinvision/lcv/x/registry/types"
)

// -- Assets --

func (k Keeper) SetAsset(ctx context.Context, asset types.Asset) error {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz, err := k.cdc.Marshal(&asset)
	if err != nil {
		return err
	}
	store.Set(types.AssetKey(asset.Id), bz)
	return nil
}

func (k Keeper) GetAsset(ctx context.Context, id string) (types.Asset, bool) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.AssetKey(id))
	if bz == nil {
		return types.Asset{}, false
	}
	var asset types.Asset
	k.cdc.MustUnmarshal(bz, &asset)
	return asset, true
}

func (k Keeper) GetAllAssets(ctx context.Context) []*types.Asset {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iterator := store.Iterator(types.AssetKeyPrefix, storeIteratorEnd(types.AssetKeyPrefix))
	defer iterator.Close()

	var assets []*types.Asset
	for ; iterator.Valid(); iterator.Next() {
		var asset types.Asset
		k.cdc.MustUnmarshal(iterator.Value(), &asset)
		assets = append(assets, &asset)
	}
	return assets
}

// -- Transfer history --

// AppendTransferRecord adds one entry to an asset's transfer history,
// preserving chronological order for TransferHistory queries.
func (k Keeper) AppendTransferRecord(ctx context.Context, record types.TransferRecord) error {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	seq := k.nextTransferSeq(ctx, record.AssetId)

	bz, err := k.cdc.Marshal(&record)
	if err != nil {
		return err
	}
	store.Set(types.TransferRecordKey(record.AssetId, seq), bz)
	return nil
}

func (k Keeper) nextTransferSeq(ctx context.Context, assetID string) uint64 {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	key := types.TransferSeqKey(assetID)
	bz := store.Get(key)
	var seq uint64
	if bz != nil {
		seq = binary.BigEndian.Uint64(bz)
	}
	next := make([]byte, 8)
	binary.BigEndian.PutUint64(next, seq+1)
	store.Set(key, next)
	return seq
}

// GetTransferHistory returns every transfer record for an asset in
// chronological order.
func (k Keeper) GetTransferHistory(ctx context.Context, assetID string) []*types.TransferRecord {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefix := types.TransferRecordPrefix(assetID)
	iterator := store.Iterator(prefix, storeIteratorEnd(prefix))
	defer iterator.Close()

	var records []*types.TransferRecord
	for ; iterator.Valid(); iterator.Next() {
		var record types.TransferRecord
		k.cdc.MustUnmarshal(iterator.Value(), &record)
		records = append(records, &record)
	}
	return records
}

func (k Keeper) GetAllTransferRecords(ctx context.Context) []*types.TransferRecord {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iterator := store.Iterator(types.TransferRecordKeyPrefix, storeIteratorEnd(types.TransferRecordKeyPrefix))
	defer iterator.Close()

	var records []*types.TransferRecord
	for ; iterator.Valid(); iterator.Next() {
		var record types.TransferRecord
		k.cdc.MustUnmarshal(iterator.Value(), &record)
		records = append(records, &record)
	}
	return records
}

// -- Jurisdiction authorities --

func (k Keeper) SetJurisdictionAuthority(ctx context.Context, auth types.JurisdictionAuthority) error {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz, err := k.cdc.Marshal(&auth)
	if err != nil {
		return err
	}
	store.Set(types.JurisdictionAuthorityKey(auth.JurisdictionCode), bz)
	return nil
}

func (k Keeper) GetJurisdictionAuthority(ctx context.Context, jurisdictionCode string) (types.JurisdictionAuthority, bool) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.JurisdictionAuthorityKey(jurisdictionCode))
	if bz == nil {
		return types.JurisdictionAuthority{}, false
	}
	var auth types.JurisdictionAuthority
	k.cdc.MustUnmarshal(bz, &auth)
	return auth, true
}

func (k Keeper) IsJurisdictionAuthority(ctx context.Context, jurisdictionCode, address string) bool {
	auth, found := k.GetJurisdictionAuthority(ctx, jurisdictionCode)
	return found && auth.Address == address
}

func (k Keeper) GetAllJurisdictionAuthorities(ctx context.Context) []*types.JurisdictionAuthority {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iterator := store.Iterator(types.JurisdictionAuthorityKeyPrefix, storeIteratorEnd(types.JurisdictionAuthorityKeyPrefix))
	defer iterator.Close()

	var auths []*types.JurisdictionAuthority
	for ; iterator.Valid(); iterator.Next() {
		var auth types.JurisdictionAuthority
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
