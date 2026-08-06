# Audit Scope

This document is for external auditors: what to review, what's already known and accepted,
and where to focus. LastcoinVision (LCV) is devnet/testnet-stage software — this is not a
mainnet audit, and nothing here should be read as a claim that an audit has already happened.
No audit has been performed; this is the scope document that would guide one.

See [`SECURITY_SELF_ASSESSMENT.md`](SECURITY_SELF_ASSESSMENT.md) for current tool output and a
manual review of recently-added code — an internal self-assessment meant to give an auditor a
running start, explicitly not a substitute for independent review.

## Building and running

```bash
make devnet    # 4-validator local devnet
make test      # Go unit/integration tests
make lint      # gosec + staticcheck (see chain/ findings below)
```

See [`README.md`](README.md) and [`docs/docs/getting-started/run-a-node.md`](docs/docs/getting-started/run-a-node.md).

## In scope

| Component | Path | Reference |
|---|---|---|
| Chain application | `chain/app/` | ante handler, module wiring, upgrade handlers |
| `x/emission` | `chain/x/emission/` | [`modules/emission/SPEC.md`](modules/emission/SPEC.md) |
| `x/feemarket` | `chain/x/feemarket/` | [`modules/feemarket/SPEC.md`](modules/feemarket/SPEC.md) |
| `x/identity` | `chain/x/identity/` | [`modules/identity/SPEC.md`](modules/identity/SPEC.md) |
| `x/registry` | `chain/x/registry/` | [`modules/registry/SPEC.md`](modules/registry/SPEC.md) |
| `x/voting` | `chain/x/voting/` | [`modules/voting/SPEC.md`](modules/voting/SPEC.md) |
| `x/taxrail` | `chain/x/taxrail/` | [`modules/taxrail/SPEC.md`](modules/taxrail/SPEC.md) |
| CosmWasm integration | `chain/` (wasmd wiring) | [`modules/wasm/SPEC.md`](modules/wasm/SPEC.md) |
| EVM integration | `chain/` (cosmos/evm wiring) | [`modules/evm/SPEC.md`](modules/evm/SPEC.md) |
| Reference Solidity contracts | `contracts/solidity/src/` | `LCVEscrow.sol`, `InvoiceFactor.sol`, `CommerceEscrow.sol` |
| Reference CosmWasm contracts | `contracts/cosmwasm/` | `consent-registry`, `procurement-ledger` |
| TypeScript SDK | `sdk/` | wallet/tx-signing helpers, typed module clients |
| Wallet key management | `packages/wallet-core/` | mnemonic generation, dual Cosmos+EVM HD derivation, encrypted keystore — **highest-priority new surface**, holds real private key material once shipped |
| Wallet tx wrappers | `packages/tx-builders/` | fixed-fee wrappers over `sdk`'s module clients for the wallet apps |
| Faucet | `faucet/` | funds testnet accounts — check for drain/abuse vectors |
| Deployment manifests | `deploy/` | docker-compose, k8s — check for hardcoded secrets, exposed ports |

**Out of scope:** `web/` (marketing site) and `docs/` (documentation site) — content bugs
there are ordinary issues, not security findings.

## Known, accepted limitations (please don't re-report these)

### `x/voting`: commit-reveal does not provide ballot secrecy

The single most important limitation in this codebase. Documented in full in
[`modules/voting/SPEC.md`](modules/voting/SPEC.md#limitations--ballot-secrecy-is-not-provided-by-this-design):
every `MsgRevealVote` is a signed, public transaction, so anyone can link a voter's address to
their revealed choice. This is a **design limitation, not a bug** — a real anonymous-voting
scheme needs ZK proofs or threshold encryption, neither of which is implemented. Internet
voting for real elections is explicitly not claimed anywhere in this repo.

### `x/identity`: no real zero-knowledge verification

`MsgPresentCredential` is authenticated by tx-signer identity alone. The `ZKVerifier`
interface (`chain/x/identity/types/zk.go`) and its `StubZKVerifier` implementation are a
marked extension point for a future selective-disclosure scheme — `VerifyProof` is defined
but not called from any handler yet. See
[`modules/identity/SPEC.md`](modules/identity/SPEC.md#limitations).

### `x/registry`: `credential_type`/asset metadata are free-form strings

Not governance-controlled enums. See each module's own SPEC.md Limitations section for the
full list (registry, taxrail, and emission each have one).

### `packages/wallet-core`: internal security review (2026-08-06), before any UI ships

Full detail in `DECISIONS.md`; summarized here for auditors. This package has no UI yet
(desktop/mobile apps haven't started) — this review covered the library in isolation.

- `npm audit`: 0 vulnerabilities in the package's own dependency tree.
- `eslint-plugin-security` (recommended rule set): 0 findings.
- Manual checklist, verified against the actual code (not assumed):
  - No `console.*` calls anywhere in `packages/wallet-core/src` or `packages/tx-builders/src`.
  - No network calls (`fetch`/`http`/etc.) anywhere in `wallet-core` — it's a pure library;
    networking lives in `tx-builders`/`sdk`, which only ever receive an already-constructed
    signer, never a mnemonic or private key.
  - No mnemonic or private key value is ever interpolated into an `Error` message.
  - `Math.random` is not used anywhere for key material — `mnemonic.ts` uses `@cosmjs/crypto`'s
    `Random` (backed by `crypto.getRandomValues`) exclusively; the one match for the string
    `Math.random` in the codebase is a comment explicitly stating it's *not* used.
  - Keystore encryption (`keystore.ts`) uses `@noble/ciphers`' `xchacha20poly1305`, an
    authenticated cipher — confirmed by test: tampering with the ciphertext or nonce, or using
    the wrong password, all fail closed (throw) rather than returning garbage plaintext.
  - No code path can reach CosmJS's "auto" fee-estimation sentinel — enforced by a grep-based
    `npm run lint` guard in both `wallet-core` and `tx-builders`, not just convention.
  - `LcvWallet` exposes `cosmosAddress()` and `evmAddress` as separately-named members; nothing
    in the codebase merges them into one field — the dual-derivation design (see `derive.ts`)
    is structurally impossible to present as "one account" from this API.

**Not yet done, by design:** an *external*, independent audit of this package specifically.
This was an internal review gate before starting UI work (Phase 2 of the wallet build), not a
substitute for one — see the mainnet-readiness gate in `DECISIONS.md`, which requires an
external audit of this exact package before it's ever pointed at real funds.

### Dependency-scanner findings reviewed and accepted this phase (2026-08-03)

Full detail in `DECISIONS.md`; summarized here for auditors who want the short version:

- **gosec**: 0 findings after `-exclude-generated` (protobuf-generated `.pb.go`/`.pulsar.go`
  files are excluded — they trip G115/G101 on varint-encoding math and gRPC method-path
  strings, neither a real issue). 8 remaining `#nosec` annotations in hand-written code, each
  with an inline justification comment: two unix-timestamp int64/uint64 conversions
  (`x/emission/keeper/abci.go`), one block-gas-limit conversion bounded by the consensus
  param (`x/feemarket/keeper/abci.go`), one cosmetic devnet chain-id-suffix RNG
  (`cmd/lcvd/cmd/testnet_multi_node.go`), a copy-gentx-files helper whose paths are
  self-enumerated, not user input (`cmd/lcvd/cmd/testnet_multi_node.go`), and a test-only
  reflection helper (`testutil/nullify/nullify.go`).
- **A real bug gosec did catch and that got fixed**: `x/feemarket/keeper/abci.go`'s
  `RequiredFee` cast a transaction's own attacker-controlled `gasWanted uint64` to `int64`
  before multiplying — a `gasWanted` at or above 2^63 would wrap negative, making the
  required fee negative and letting any paid amount (including zero) pass the ante handler's
  fee check. Fixed by using `math.NewIntFromUint64` instead of the cast; regression test at
  `TestRequiredFeeDoesNotOverflowForLargeGasWanted`. **Auditors should specifically look for
  the same pattern (an unbounded tx-supplied `uint64` field cast to `int64` before
  arithmetic) anywhere else in the ante/fee-handling path.**
- **staticcheck**: 0 actionable findings after removing one genuinely dead function
  (`mustFeeMarketParams` in `app/ante.go`) and suppressing two confirmed-false-positive
  SA4006 reports in `app/sim_test.go` (the flagged variables are read later in the same
  function — verified by hand). Remaining SA1019 deprecation warnings are all either
  cosmos-sdk's own x/crisis/invariant-registry transitional deprecations (no replacement
  ships yet in SDK v0.53.6 — removing our invariant registrations would be a regression, not
  a fix) or auto-generated `.pb.gw.go` grpc-gateway files using an older codegen template.
- **cargo audit** (both CosmWasm contracts): 2 "unmaintained" warnings (`derivative`,
  `paste`) — transitive deps with no known vulnerability, just no active maintainer. No fix
  available upstream from `cosmwasm-std`'s own dependency tree.
- **npm audit** (all 5 Node packages — `sdk`, `cli`, `faucet`, `web`, `docs`): the critical
  `protobufjs@6.11.6` RCE (transitive via `@confio/ics23`, pulled in by `@cosmjs/stargate
  @0.32.4`) was fixed at the root by bumping the whole `@cosmjs/*` family to `0.39.0`, which
  dropped the `ics23` dependency entirely (`sdk`, `faucet`). `web`'s Next.js was bumped
  14→16 and React 18→19 to clear a long list of real Next.js CVEs (DoS/SSRF/cache-poisoning,
  fixed only in >=15.5.21); `next`'s own internal `postcss`/`sharp` pins were forced to
  patched versions via npm `overrides`. Remaining accepted findings: one low-severity
  `esbuild` dev-server-only issue (`sdk`/`cli`/`faucet`/`web`, Windows-only, not present in
  any shipped output), and Docusaurus 3.x's own vendored webpack build-tooling
  (`webpack-dev-server`, `sockjs`, `serialize-javascript`, `uuid` — `docs`) which has no
  newer upstream release yet and is dev-server-only, never present in the deployed static
  site.
- **Slither** (Solidity contracts): fixed one style-only "uninitialized local" (Solidity
  zero-initializes value types by language guarantee; added an explicit `= 0` anyway).
  Remaining findings: `block.timestamp` used for day/hour-scale deadline comparisons (the
  standard, accepted pattern — a few seconds of miner/validator drift doesn't matter at that
  granularity) and low-level `.call{value:}()` for ETH transfers (the *correct* modern
  pattern post-Istanbul, not `.transfer()`/`.send()`'s fixed 2300-gas stipend — every call
  site's return value is checked via `require(ok, ...)`, verified by hand). All three
  contracts' fund-releasing functions update state (`m.released = true`, etc.) **before** the
  external call — checks-effects-interactions is followed throughout; no reentrancy finding.

## Fuzz coverage

Three property-based Go native fuzz targets (`go test -fuzz=... -fuzztime=5m`), each run for
just over 5 minutes with zero failures (2026-08-03):

- `FuzzMsgTaxedTransfer` (`chain/x/taxrail/keeper/msg_server_transfer_fuzz_test.go`,
  34,806 executions / 323s): for any rate in `RegisterTaxRule`'s validated `[0,1]` range and
  any amount, `tax + remainder == amount` exactly, neither half negative, and the
  jurisdiction-ledger/module-balance invariant holds.
- `FuzzMsgTransferAsset` (`chain/x/registry/keeper/msg_server_asset_transfer_fuzz_test.go`,
  33,345 executions / 315s): models the four independent preconditions (frozen status,
  caller authorization, both parties' attestation) and checks a transfer succeeds if and
  only if all four hold, for every combination the fuzzer generates.
- `FuzzVoteTally` (`chain/x/voting/keeper/tally_fuzz_test.go`, 33,783 executions / 305s): for
  any number of voters and any subset registering/committing/revealing, tallied votes never
  exceed registered voters, and the sum of per-option tallies equals `TotalRevealed` exactly.

## What auditors should prioritize

1. The ante handler / fee-market path (`chain/app/ante.go`, `chain/x/feemarket/ante/`) —
   given the `RequiredFee` overflow found this phase, this is the highest-value area to
   re-check for the same class of bug (unbounded tx-supplied integers reaching arithmetic
   without an explicit range check).
2. `x/voting`'s nullifier/commit-reveal implementation against the limitations already
   documented — confirm nothing beyond what's claimed.
3. The EVM↔Cosmos bank bridge (`modules/evm/SPEC.md`) — bidirectional value bridges between
   two different accounting representations are a classic source of real exploits elsewhere
   in the ecosystem.
4. `contracts/solidity/` and `contracts/cosmwasm/` reference contracts, since these are the
   ones most likely to be copied directly into a real dApp.
