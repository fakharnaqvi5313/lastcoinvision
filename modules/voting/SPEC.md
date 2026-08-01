# `x/voting`

Code: [`chain/x/voting`](../../chain/x/voting) (see [DECISIONS.md](../../DECISIONS.md)
for why module code lives under `chain/x/` rather than here).

## Purpose

On-chain elections with credential-gated eligibility, one-vote-per-identity enforcement, and
a commit-reveal ballot scheme whose tally anyone can independently recompute from public
chain state. This is **not** a zero-knowledge or fully anonymous voting system — see
Limitations for exactly what secrecy guarantee this design does and does not provide.

## Electoral authorities

`MsgAddElectoralAuthority` / `MsgRemoveElectoralAuthority` are authority-gated (module
`authority`, defaults to `x/gov`) and maintain a flat authorized set, mirroring
`x/identity`'s issuer-registry pattern — electoral authorities aren't jurisdiction-scoped like
`x/registry`'s, so a set rather than a 1:1 map is the right shape. Only an address in this set
may create elections.

## Election lifecycle

- `MsgCreateElection(created_by, title, description, eligibility_credential_type, options,
  registration_start, registration_end, voting_start, voting_end, reveal_end)` — validates
  strict window ordering: `registration_start < registration_end <= voting_start < voting_end
  <= reveal_end`. `eligibility_credential_type` names an `x/identity` credential type (e.g.
  `"voter-registration"`) rather than hardcoding a tier, so different elections can require
  different credentials.
- `MsgRegisterToVote(voter, election_id)` — only during the registration window
  (`ErrNotRegistrationWindow`), only once per voter per election (`ErrAlreadyRegistered`),
  and only if the voter holds an active credential of the election's
  `eligibility_credential_type` per `x/identity` (`ErrNotEligible`).
- `MsgCommitVote(voter, election_id, commitment)` — only during the voting window
  (`ErrNotVotingWindow`), only for registered voters (`ErrNotRegistered`), and only once per
  voter per election — enforced by a **nullifier**: `sha256("did:lcv:"+voter+"|"+election_id)`.
  The nullifier, not the voter address, is what's stored against the commitment, so a second
  commit attempt from the same voter for the same election collides on the nullifier
  (`ErrAlreadyCommitted`) without the commitment record itself needing to name the voter.
  `commitment = sha256(choice+"|"+salt)`, computed off-chain by the voter and submitted as an
  opaque hash.
- `MsgRevealVote(voter, election_id, choice, salt)` — only during the reveal window
  (`ErrNotRevealWindow`), only for a nullifier with an existing, not-yet-revealed commitment
  (`ErrNoCommitment` / `ErrAlreadyRevealed`), and only if `sha256(choice+"|"+salt)` matches the
  stored commitment exactly (`ErrCommitmentMismatch`) and `choice` is one of the election's
  declared `options` (`ErrInvalidChoice`).

## Tally

`ElectionResults` is a **live query**, not a stored value: it iterates every `VoteCommitment`
for the election and counts `revealed = true` entries by `revealed_choice`, every time it's
called. There is no separate "tally" record that could drift from the underlying commitments
— per the top-level spec's "anyone can recompute the tally from on-chain data," recomputing
*is* the only way results are ever produced.

## State

- `election/<id>` → `Election`; `election_seq` (global counter) generates ids via
  `strconv.FormatUint(seq+1, 10)`.
- `registered_voter/<election_id>/<voter_address>` → presence marker.
- `vote_commitment/<election_id>/<nullifier>` → `VoteCommitment` (commitment, revealed,
  revealed_choice, committed_at, revealed_at).
- `electoral_authority/<address>` → presence marker.

## Queries

`Election(id)`, `ElectionResults(id)` (live tally, see above), `ElectoralAuthorityList`.

## Events

Five typed events (authority added/removed, election created, vote committed, vote revealed).
Deliberately, `EventVoteCommitted` and `EventVoteRevealed` carry the nullifier and (for
reveal) the choice, but **not** the voter's address — see Limitations for why this doesn't
actually achieve anonymity, but it avoids adding one more redundant on-chain linkage beyond
what tx signing already provides.

## Invariants

`TalliedVotesNotExceedRegisteredInvariant` — for every election, the number of `revealed=true`
commitments can never exceed the number of registered voters. Since a commitment can only be
created for a registered voter's nullifier and reveal only flips an existing commitment's
flag, this should hold by construction; it's registered as a chain-halting invariant to catch
any future regression.

## Cross-module dependency

`x/voting` depends on `x/identity` through a narrow expected-keeper interface
(`IdentityKeeper.HasActiveCredentialOfType`) for registration eligibility.

## Limitations — ballot secrecy is NOT provided by this design

This is the most important limitation in this module and is stated explicitly, not buried:

- **The commit-reveal scheme prevents double-voting and lets anyone verify the tally, but it
  does not hide who voted for what from the chain's own observers.** Every `MsgRevealVote` is
  a signed transaction from the voter's own account, broadcast and included in a public block.
  Anyone watching the mempool or reading block history can trivially link the revealing
  address to the revealed `choice` — the "commitment" only hides the vote between the commit
  and reveal phases, not after reveal, and not from the network layer (IP address, timing)
  even before that.
- **This is fundamentally different from a real anonymous-voting scheme** (e.g. one built on
  zero-knowledge proofs, blind signatures, or a mixnet), where the tally can be verified
  *without* any observer being able to link a specific voter to a specific choice. No such
  mechanism is implemented here. If ballot secrecy against on-chain observers is a hard
  requirement, this module does not meet it as designed — a ZK reveal step (prove "I hold a
  valid commitment for choice X" without revealing which committed nullifier it came from)
  would be required, and is a natural extension point (see `x/identity`'s `ZKVerifier` stub
  for the same class of future work) but is out of scope for Phase C.
- **The nullifier is address-derived, not committed to independently.** `sha256("did:lcv:"+
  voter+"|"+election_id)` is fully computable by anyone who knows the voter's address and the
  election id — it prevents double-voting but is not itself a secrecy mechanism; recovering
  which nullifier belongs to which voter is trivial for anyone with the voter list (which is
  itself public, since `MsgRegisterToVote` is a signed, broadcast transaction).
- **Coercion/vote-buying resistance is not addressed.** Because reveal is a voter-signed
  transaction naming the choice in the clear, a voter can always prove how they voted to a
  third party after the fact (by showing the transaction), which is the classic
  coercion/vote-selling vector that real secret-ballot systems are specifically designed to
  prevent. This design does not prevent it.
