# `x/identity`

Code: [`chain/x/identity`](../../chain/x/identity) (see [DECISIONS.md](../../DECISIONS.md)
for why module code lives under `chain/x/` rather than here).

## Purpose

Protocol-level W3C DID + Verifiable Credentials support: every address can register a
`did:lcv:<address>` DID document, governance-authorized issuers anchor credential **hashes**
(never the underlying personal data) against a subject address, and an address's trust tier
— pseudonymous, attested, or institutional — is derived on the fly from that state rather
than stored directly.

## DIDs

- `did:lcv:<bech32 address>`. An address may control at most one DID.
- `MsgCreateDID` / `MsgUpdateDID` / `MsgDeactivateDID`, all signed by the DID's own
  controller (`creator`). A DID document holds only W3C DID Core `VerificationMethod` and
  `ServiceEndpoint` entries — no personal data belongs in either.
- Deactivation is permanent: a deactivated DID cannot be updated or reactivated
  (`ErrDIDDeactivated`).

## Credential registry

- `MsgAddIssuer` / `MsgRemoveIssuer` are authority-gated (the module authority defaults to
  the `x/gov` module account), matching the top-level spec's "authorized issuers (added via
  governance)". Removing an issuer does **not** retroactively revoke credentials it already
  issued — the `Credential` record is untouched — but it does stop those credentials from
  granting the attested tier, since tier computation re-checks issuer authorization live
  (see `TestIssuerRemovalRevokesAttestedTier`).
- `MsgIssueCredential(issuer, subject, credential_type, credential_hash)` — only callable by
  a currently authorized issuer. `credential_hash` is a hash of the off-chain credential
  document; the document itself never touches the chain. The credential `id` is a
  deterministic `sha256(issuer|subject|credential_type|hash|issued_at)`.
- `MsgRevokeCredential` — callable by the original issuer or the module authority. Flips
  `status` from `ACTIVE` to `REVOKED`; there is no "expired" status, since expiry is a
  property of the off-chain document, not tracked here.
- `MsgPresentCredential(presenter, credential_id, proof)` — today, "verification" is
  entirely the standard Cosmos SDK tx-signature check: the handler requires
  `presenter == credential.subject`. The `proof` field and the `ZKVerifier` interface
  (`x/identity/types/zk.go`, `StubZKVerifier`) exist as the documented extension point for a
  future zero-knowledge selective-disclosure scheme, but `VerifyProof` is not yet called
  from the handler — see Limitations.

## Account tiers

`TierOf` (query) / `ComputeTier` (keeper) compute a tier on every call rather than storing
one, with this precedence:

1. **institutional** — the address is a `x/group` group-policy account
   (`GroupKeeper.GroupPolicyInfo`) with non-empty `Metadata`. This deliberately reuses
   `x/group`'s own metadata field as the "role" descriptor instead of inventing a parallel
   identity-module registry, per the top-level spec's literal wording: "institutional
   (group-module account with role metadata)".
2. **attested** — the address holds at least one `ACTIVE` credential whose
   `credential_type` equals `params.attested_credential_type` (default `"kyc"`), issued by
   an issuer that is *currently* authorized.
3. **pseudonymous** — the default for everyone else.

Institutional takes precedence over attested: an institutional account holding a KYC
credential is still reported as institutional (see
`TestInstitutionalTierTakesPrecedenceOverAttested`).

## State

- `Params` (key `p_identity`): `attested_credential_type`.
- `did/<address>` → `DIDDocument`.
- `credential/<id>` → `Credential`.
- `credential_by_subject/<subject>/<id>` → presence marker; a secondary index so
  `ComputeTier` and `CredentialStatus` lookups don't scan every credential ever issued.
- `issuer/<address>` → `Issuer`.

## Queries

`Params`, `DIDByAddress`, `CredentialStatus`, `IssuerList`, `TierOf` — all four custom
queries named in the top-level spec.

## Events

Typed events for every state change: `EventDIDCreated`, `EventDIDUpdated`,
`EventDIDDeactivated`, `EventIssuerAdded`, `EventIssuerRemoved`, `EventCredentialIssued`,
`EventCredentialRevoked`, `EventCredentialPresented`.

## Limitations

- **No real ZK verification yet.** `MsgPresentCredential` is authenticated solely by
  tx-signer identity. `ZKVerifier`/`StubZKVerifier` are a clearly marked stub — the
  interface exists so a real zero-knowledge selective-disclosure scheme can be wired in
  later without changing the message's wire format, but no such scheme is implemented, and
  `VerifyProof` isn't called yet.
- **`credential_type` is a free-form string**, not a governance-controlled enum. Any issuer
  can issue any `credential_type` value; only `params.attested_credential_type` has special
  meaning (granting the attested tier). Governance restricting which `credential_type`
  values are legal at all is a possible future addition, not implemented here.
- **Institutional role metadata is unstructured.** `x/group`'s `Metadata` field is a plain
  string (commonly JSON by convention, per that module's own docs) — `x/identity` does not
  parse or validate its structure, it's surfaced as-is via `TierOf`'s `role_metadata`.
- **No selective field disclosure.** Presenting a credential today is all-or-nothing (it's
  just a status check) — there's no mechanism yet for proving a predicate about the
  off-chain document (e.g. "holder is over 18") without the ZK extension point above.
