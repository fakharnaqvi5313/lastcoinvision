package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/runtime"

	"github.com/lastcoinvision/lcv/x/identity/types"
)

// -- DID documents --

func (k Keeper) SetDID(ctx context.Context, doc types.DIDDocument) error {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz, err := k.cdc.Marshal(&doc)
	if err != nil {
		return err
	}
	store.Set(types.DIDKey(doc.Controller), bz)
	return nil
}

func (k Keeper) GetDID(ctx context.Context, address string) (types.DIDDocument, bool) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.DIDKey(address))
	if bz == nil {
		return types.DIDDocument{}, false
	}
	var doc types.DIDDocument
	k.cdc.MustUnmarshal(bz, &doc)
	return doc, true
}

func (k Keeper) GetAllDIDs(ctx context.Context) []*types.DIDDocument {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iterator := store.Iterator(types.DIDKeyPrefix, storeIteratorEnd(types.DIDKeyPrefix))
	defer iterator.Close()

	var docs []*types.DIDDocument
	for ; iterator.Valid(); iterator.Next() {
		var doc types.DIDDocument
		k.cdc.MustUnmarshal(iterator.Value(), &doc)
		docs = append(docs, &doc)
	}
	return docs
}

// -- Credentials --

func (k Keeper) SetCredential(ctx context.Context, cred types.Credential) error {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz, err := k.cdc.Marshal(&cred)
	if err != nil {
		return err
	}
	store.Set(types.CredentialKey(cred.Id), bz)
	store.Set(types.CredentialBySubjectKey(cred.Subject, cred.Id), []byte{1})
	return nil
}

func (k Keeper) GetCredential(ctx context.Context, id string) (types.Credential, bool) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.CredentialKey(id))
	if bz == nil {
		return types.Credential{}, false
	}
	var cred types.Credential
	k.cdc.MustUnmarshal(bz, &cred)
	return cred, true
}

// GetCredentialsBySubject returns every credential ever issued to subject,
// active or revoked.
func (k Keeper) GetCredentialsBySubject(ctx context.Context, subject string) []types.Credential {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefix := types.CredentialBySubjectPrefix(subject)
	iterator := store.Iterator(prefix, storeIteratorEnd(prefix))
	defer iterator.Close()

	var creds []types.Credential
	for ; iterator.Valid(); iterator.Next() {
		id := string(iterator.Key()[len(prefix):])
		if cred, found := k.GetCredential(ctx, id); found {
			creds = append(creds, cred)
		}
	}
	return creds
}

func (k Keeper) GetAllCredentials(ctx context.Context) []*types.Credential {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iterator := store.Iterator(types.CredentialKeyPrefix, storeIteratorEnd(types.CredentialKeyPrefix))
	defer iterator.Close()

	var creds []*types.Credential
	for ; iterator.Valid(); iterator.Next() {
		var cred types.Credential
		k.cdc.MustUnmarshal(iterator.Value(), &cred)
		creds = append(creds, &cred)
	}
	return creds
}

// -- Issuers --

func (k Keeper) SetIssuer(ctx context.Context, issuer types.Issuer) error {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz, err := k.cdc.Marshal(&issuer)
	if err != nil {
		return err
	}
	store.Set(types.IssuerKey(issuer.Address), bz)
	return nil
}

func (k Keeper) GetIssuer(ctx context.Context, address string) (types.Issuer, bool) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.IssuerKey(address))
	if bz == nil {
		return types.Issuer{}, false
	}
	var issuer types.Issuer
	k.cdc.MustUnmarshal(bz, &issuer)
	return issuer, true
}

func (k Keeper) IsAuthorizedIssuer(ctx context.Context, address string) bool {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return store.Has(types.IssuerKey(address))
}

// DeleteIssuer removes an issuer's store entry. Named distinctly from the
// MsgRemoveIssuer handler (keeper.msgServer.RemoveIssuer) to avoid the two
// methods shadowing each other through msgServer's embedding of Keeper,
// which would otherwise make the handler call itself instead of this.
func (k Keeper) DeleteIssuer(ctx context.Context, address string) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store.Delete(types.IssuerKey(address))
}

func (k Keeper) GetAllIssuers(ctx context.Context) []*types.Issuer {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iterator := store.Iterator(types.IssuerKeyPrefix, storeIteratorEnd(types.IssuerKeyPrefix))
	defer iterator.Close()

	var issuers []*types.Issuer
	for ; iterator.Valid(); iterator.Next() {
		var issuer types.Issuer
		k.cdc.MustUnmarshal(iterator.Value(), &issuer)
		issuers = append(issuers, &issuer)
	}
	return issuers
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
