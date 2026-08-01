# `x/registry`

Code: [`chain/x/registry`](../../chain/x/registry) (see [DECISIONS.md](../../DECISIONS.md)
for why module code lives under `chain/x/` rather than here).

## Purpose

A protocol-level asset registry for government-recorded, non-fungible real-world records —
land parcels, vehicles, and similar assets identified by a jurisdiction code and a metadata
hash. Assets are **records to be tracked, not tokens to be traded**: there is no AMM, no
order book, and no fungibility between assets. The only state transitions are registration,
ownership transfer, and jurisdiction-authority-ordered freezing, each fully auditable via an
append-only transfer history.

## Jurisdiction authorities

- `MsgSetJurisdictionAuthority(authority, jurisdiction_code, address)` is gated on the module
  `authority` (defaults to the `x/gov` module account) — one governance-set address per
  jurisdiction code, upserted on each call. This mirrors `x/identity`'s issuer-registration
  pattern but is naturally 1:1 (a jurisdiction has exactly one current authority, not an open
  set), so it's a single keyed record rather than a list.
- Only the current jurisdiction authority for an asset's `jurisdiction_code` may register new
  assets in that jurisdiction or freeze/unfreeze assets already in it.

## Assets

- `MsgRegisterAsset(creator, asset_type, jurisdiction_code, metadata_hash, owner)` — callable
  only by the jurisdiction's authority. `owner` is a `did:lcv:<address>` DID string, not a raw
  address, so the registry's notion of "owner" is identity-linked from the start. The asset
  `id` is deterministic: `sha256(jurisdiction_code|asset_type|metadata_hash|registered_at)`.
  `metadata_hash` is a hash of off-chain asset documentation (deed scan, title record, etc.)
  — the document itself never touches the chain, matching `x/identity`'s credential-hash
  pattern.
- `MsgTransferAsset(creator, asset_id, new_owner)` — the signer must be the current owner (via
  DID-to-address mapping) or the jurisdiction authority. The asset must not be `FROZEN`
  (`ErrAssetFrozen`). **Both** the outgoing and incoming owner addresses must hold at least
  the `attested` tier per `x/identity` (`ErrPartyNotAttested`) — government-grade asset
  transfer requires both parties to have passed identity attestation, not just the sender.
- `MsgFreezeAsset(creator, asset_id, reason)` — callable only by the asset's jurisdiction
  authority. `reason` is a required, on-chain, human-readable string (e.g. "suspected fraud
  investigation") — freezing is not a silent action. There is no separate unfreeze message;
  the same jurisdiction-authority gate on `RegisterAsset`/transfers means freeze is currently
  one-directional (see Limitations).

## Transfer history

Every successful `MsgTransferAsset` appends an immutable `TransferRecord{asset_id, from, to,
transferred_at}` to a per-asset, chronologically ordered log (keyed by an auto-incrementing
sequence number, zero-padded for lexicographic = chronological store iteration). `TransferHistory`
query returns the full list for an asset — anyone can reconstruct an asset's complete chain of
custody from on-chain data alone.

## State

- `asset/<id>` → `Asset` (asset_type, jurisdiction_code, metadata_hash, owner DID, status,
  freeze_reason, registered_at, updated_at).
- `transfer_record/<asset_id>/<zero-padded seq>` → `TransferRecord`.
- `transfer_seq` (global counter) → next sequence number.
- `jurisdiction_authority/<jurisdiction_code>` → `JurisdictionAuthority`.

## Queries

`AssetByID(id)`, `TransferHistory(asset_id)`, `JurisdictionAuthorityOf(jurisdiction_code)`.

## Events

`EventJurisdictionAuthoritySet`, `EventAssetRegistered`, `EventAssetTransferred`,
`EventAssetFrozen`.

## Invariants

`TransferHistoryReferentialIntegrityInvariant` — every stored `TransferRecord.asset_id` must
reference an asset that actually exists in state. Registered in
`RegisterInvariants`/`InvariantRegistry` so a bug that leaves an orphaned transfer record
halts the chain rather than silently corrupting the audit trail.

## Cross-module dependency

`x/registry` depends on `x/identity` through a narrow expected-keeper interface
(`IdentityKeeper.ComputeTier`) to check the `attested`/`institutional` tier gate on transfers
— it does not re-implement or duplicate identity/credential logic.

## Limitations

- **No unfreeze message.** Freezing is currently one-directional; reversing a freeze requires
  a manual state migration or a future `MsgUnfreezeAsset` (straightforward to add — same
  authority gate — but not implemented since it wasn't in the acceptance criteria).
- **Jurisdiction authority is a single address, not a role hierarchy.** There's no delegation,
  multi-sig, or sub-jurisdiction structure; a real deployment would likely put a `x/group`
  policy account or another module account behind this address rather than an EOA.
- **`asset_type` is a free-form string**, not a governance-controlled enum — same tradeoff as
  `x/identity`'s `credential_type`, made for the same reason (flexibility now, governance
  restriction is a possible later addition).
