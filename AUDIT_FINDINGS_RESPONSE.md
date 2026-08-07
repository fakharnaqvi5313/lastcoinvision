# Audit Findings Response Process

This document governs how findings from a **professional audit engagement** (currently: the
AftPak.com engagement — see `DECISIONS.md`, 2026-08-06/2026-08-07) get triaged, resolved, and
recorded, so that `MAINNET_LAUNCH_RUNBOOK.md`'s audit gate — *"professional security audit
complete, and every finding either fixed or explicitly accepted in writing"* — can eventually be
closed on real, checkable evidence rather than a claim.

This is distinct from `SECURITY.md`, which covers ad-hoc vulnerability reports from anyone,
anytime. This document covers a structured audit's findings specifically — the kind that arrive
as a numbered list in a report, all at once, and need to be worked through systematically before
a launch decision gets made on top of them.

**Written now, before any findings exist to process**, on purpose: deciding the response bar
under no time pressure, rather than improvising it the day a report with a scary-looking Critical
finding actually lands.

## Severity definitions

Adapted to this project's actual threat model — a chain that will eventually hold real value and
has tax/voting/land-registry modules where "denial of service" and "silent incorrect state" both
matter, not just "funds stolen":

- **Critical** — direct path to fund loss, consensus failure/chain halt, double-spend, or
  forging a credential/identity/vote result. Blocks mainnet unconditionally; no written
  acceptance is available for Critical findings — they get fixed, or the launch doesn't happen.
- **High** — a real but more conditional path to the same categories above (requires a specific
  precondition, privileged role, or unusual sequence), or a straightforward way to grief/DoS a
  module (e.g. an unbounded loop over user-controlled input in a keeper). Must be fixed before
  mainnet; a written risk-acceptance is possible only with a specific, credible mitigating
  argument for why it's not exploitable in practice — not just "we didn't have time."
- **Medium** — real bug, but limited impact (wrong error message leaking internal state, a
  parameter that should be gated by governance but isn't yet, an edge case that degrades rather
  than breaks something). Fix before mainnet by default; written risk-acceptance is the normal
  path if the fix is disproportionate to the actual impact.
- **Low / Informational** — code-quality, defense-in-depth, or best-practice findings with no
  demonstrated exploit path (much of what automated scanners like `gosec`'s G115 findings
  already produce — see `AUDIT_SCOPE.md`). Tracked, fixed opportunistically, never blocks
  mainnet on its own.

If AftPak's (or any future auditor's) own severity labels don't map cleanly onto these, use this
scale for the actual go/no-go decision and note the mismatch rather than silently picking
whichever label sounds better.

## Triage process

For each finding, in order:

1. **Reproduce it independently before doing anything else** — this project's own pattern
   throughout `DECISIONS.md` (re-deriving every checkable AftPak claim against the actual repo
   rather than trusting the document) applies here too, doubly so: a finding that doesn't
   actually reproduce against current `main` doesn't get "fixed," it gets documented as stale or
   already-resolved, with the evidence (commit hash, test run, tool output) showing why.
2. **Classify severity** per the scale above, independent of whatever label the report used.
3. **Assign a response**: fix it, or write a risk-acceptance (see below) — every finding gets
   one of these two outcomes, nothing stays in limbo.
4. **Verify the response**: a "fixed" finding needs the same tool/reproduction from step 1 run
   again, confirming it no longer reproduces — not just a diff that looks plausible. A test
   regression is added wherever the finding's shape allows one, matching the project's existing
   convention (e.g. the fuzz tests already added for `MsgTaxedTransfer`/`MsgTransferAsset`/vote
   tally per `AUDIT_SCOPE.md`).
5. **Log it** in the Findings Log below, dated, with the outcome and links to the fix commit or
   the risk-acceptance writeup.

## Writing a risk-acceptance

A risk-acceptance is not "we decided not to worry about it." It must state, in writing, in the
Findings Log entry itself:

- The specific reason the finding isn't being fixed (disproportionate cost, already mitigated by
  a different control, requires a precondition assessed as unrealistic at this project's current
  scale, etc.) — not a placeholder, an actual argument someone could disagree with and evaluate.
- Who accepted it — this is a small project without a formal security team (see `SECURITY.md`),
  so in practice that's whoever has the authority to make the mainnet go/no-go call; name them
  when it happens rather than leaving this generic.
- What would change the decision later (e.g. "acceptable at testnet-scale transaction volume;
  revisit if throughput crosses X").

Critical findings cannot get a risk-acceptance under this process, by definition above — only
High/Medium can.

## Findings Log

**2026-08-07: AftPak's initial report's two tool categories (gosec, govulncheck), triaged.**
This is *not* the discrete, manually-reviewed findings list this document was written to
eventually process — it's the same automated-tooling output already covered by `AUDIT_SCOPE.md`/
`SECURITY_SELF_ASSESSMENT.md`. Logged anyway, at the user's request to resolve everything AftPak's
initial report identified, following this document's own reproduce-first triage process rather
than skipping it just because the source wasn't a formal audit findings list.

| # | Date | Source | Severity | Description | Outcome | Evidence |
|---|------|--------|----------|-------------|---------|----------|
| 1 | 2026-08-07 | AftPak initial report / `gosec` | Informational | 1,040 gosec findings (1,036 `G115` integer-overflow-conversion, 4 `G101` hardcoded-credentials) | **No fix needed — confirmed false-positive class.** Reproduced fresh (`gosec ./...`, no `-exclude-generated`): exactly 1,040, matching the report. Checked every issue's file path: **100% in generated `.pb.go`/`.pulsar.go`/`_grpc.pb.go` files, 0% in hand-written code.** The 4 `G101`s are gRPC method-path constants like `"/lcv.identity.Msg/RevokeCredential"` — the word "Credential" trips the heuristic, not an actual secret. Matches `AUDIT_SCOPE.md`'s existing documentation exactly; `gosec -exclude-generated` already gives 0. | `gosec -fmt json` output, filtered by file suffix — 0 non-generated hits |
| 2 | 2026-08-07 | AftPak initial report / `govulncheck` | — (12 separate CVEs) | 12 Go standard-library CVEs (`crypto/tls`, `net/textproto`, `crypto/x509`×3, `html/template`×3, `net`, `net/http`, `archive/tar`), all with a published fixed Go patch release | **Fixed.** Bumped `chain/go.mod`'s `go` directive from `1.25.8` to `1.25.12` (the highest version any of the 12 needed) — the toolchain was already cached locally. Rebuilt (`go version -m` on the binary confirms `go1.25.12`), full test suite still green (19/19 packages), re-ran `govulncheck`: all 12 gone. | `/tmp/govulncheck-after.txt` (19→7 vulnerabilities); `go test ./...` clean |
| 3 | 2026-08-07 | AftPak initial report / `govulncheck` | Low | `GO-2026-5932`: `golang.org/x/crypto/openpgp` (armor encode/decode) is unmaintained and unsafe by design | **Risk-accepted.** No fixed version exists (package deprecated, not patched). Reachable only through `cosmos-sdk`'s keyring's PGP-armor key-export/import CLI path — an operator-initiated local convenience feature, not something a network attacker can trigger, and not part of consensus/tx-processing. This is an inherited `cosmos-sdk` dependency shared by every chain using its keyring, not something LCV added. Accepted by: project maintainer (this session, per user's request to resolve AftPak's findings — revisit if `cosmos-sdk` ever removes the armor-export path or offers an alternative). | `govulncheck` call trace: `keyring.init` → `armor.*` only |
| 4 | 2026-08-07 | AftPak initial report / `govulncheck` | Low-Medium | `GO-2026-4740`: DoS in `github.com/shamaton/msgpack/v2`, pulled in by `wasmd` | **Risk-accepted, flagged for re-check.** Already at the latest available version (`v2.4.1`) — no fix exists upstream. Call trace shows only `wasm.init`-time codec registration, not a demonstrated runtime decode of attacker-controlled input — but this wasn't independently verified deeper into `wasmd`'s actual usage, so classified Low-Medium rather than dismissed as certainly unreachable. Revisit if a `wasmd` upgrade or CosmWasm contract execution path is found to actually decode msgpack from untrusted input. | `go list -m -versions` (v2.4.1 = latest); trace in `/tmp/govulncheck-full.txt` |
| 5 | 2026-08-07 | AftPak initial report / `govulncheck` | Low | `GO-2026-4479`: `github.com/pion/dtls/v2` random-nonce AES-GCM key-leak risk | **Risk-accepted.** Checked `v2.2.8`–`v2.2.12` (newer than the pinned `v2.2.7`) against `govulncheck`'s database — still no fixed version identified. Call trace is exclusively `init()`-time ciphersuite/package registration; `lcvd` doesn't run a DTLS/WebRTC handshake as part of normal operation (this dependency is transitive, likely via a networking library `cosmos/evm` or a P2P dependency pulls in for a feature LCV doesn't use). | `go list -m -versions github.com/pion/dtls/v2`; trace shows only `*.init` calls |
| 6 | 2026-08-07 | AftPak initial report / `govulncheck` | **Critical — fixed via a local vendor patch** | `GO-2025-3684` / `GHSA-mjfq-3qr2-6g84`: `github.com/cosmos/evm` "Partial Precompile State Writes." Per the GHSA advisory (severity **Critical**, fetched directly, not just the govulncheck summary): setting a low EVM call-gas limit lets a caller partially execute a precompile and error out mid-way without the partial state writes being reverted — the advisory's own example is claiming staking rewards without the claimable balance resetting to 0. | **Fixed.** A version bump alone could not resolve this: checked `v0.6.1` (pinned), `v0.7.1`, and `v1.0.0-rc2` (newest available) — `precompiles/common/precompile.go`'s `HandleGasError` is byte-for-byte identical in all three, none call any revert function. Traced the actual mechanism in `runNativeAction` (the shared entry point every stateful precompile's `Run()` funnels through, confirmed via `precompiles/distribution/distribution.go:98`): it takes a multi-store snapshot before the precompile's action runs (`stateDB.MultiStoreSnapshot()`), but never called `stateDB.RevertMultiStore(snapshot)` on any error/panic exit — the existing `RevertMultiStore` machinery was only wired into the journal-based revert path used when the *outer* EVM call reverts, not per-call. Confirmed LCV's `DefaultStaticPrecompiles()` includes `WithDistributionPrecompile(...)`, matching the advisory's exact example, and it's active at genesis (`evmGenesisState()`, `app.go`). **Vendored `cosmos/evm@v0.6.1` into `chain/third_party/cosmos-evm-patched`** (via a `go.mod` local `replace`, same pattern as the existing go-ethereum fork) with a single 4-line fix: a `defer` in `runNativeAction` that calls `stateDB.RevertMultiStore(snapshot)` whenever the function returns a non-nil error, covering every exit path (including `HandleGasError`'s panic recovery) rather than just the one the advisory's own file-specific diff targeted. Verified: LCV's full test suite (19/19 packages) and the vendored package's own existing test suite (`go test -tags=test ./precompiles/common/...`) both pass; a live call through the actual patched code path (`cast call` to the Distribution precompile's `communityPool()` on the running devnet, post-rebuild) returns correctly decoded data. **Not achieved**: a live reproduction of the actual exploit (a precisely gas-limited call that triggers a mid-write OutOfGas panic) — this needs exact calibration against the SDK's internal KV-store gas accounting, which wasn't reliably reproducible within scope. The fix is verified by direct code tracing and regression testing, not by demonstrating the pre-patch bug live. | GHSA-mjfq-3qr2-6g84 (fetched via WebFetch + `gh api repos/cosmos/evm/security-advisories/...`); `precompiles/common/precompile.go:60-96` (all 3 checked versions); `precompiles/distribution/distribution.go:97-98`; `app/app.go:678-679,1245`; `chain/third_party/cosmos-evm-patched/precompiles/common/precompile.go` (the patch, with inline comment) |
| 7 | 2026-08-07 | AftPak initial report / `govulncheck` | Medium | `GO-2024-2584` (slashing evasion), `GO-2023-1881` (`x/crisis` doesn't charge `ConstantFee`), `GO-2023-1821` (`x/crisis` doesn't cause chain halt) — all `cosmos-sdk`'s own advisories | **Risk-accepted.** Bumped `cosmos-sdk` to the latest available `v0.53.8` (from `v0.53.6`) specifically to check — confirmed via fresh `govulncheck` re-run that none of these three are fixed even at the latest patch release. `x/crisis` is confirmed actively wired into `app.go` (not dormant), so these are real, not moot — but they're long-standing, narrow-scope design-limitation advisories in a legacy SDK module that every current `cosmos-sdk` v0.53.x chain shares, with no available patched version. Accepted by: project maintainer (this session) — revisit if `cosmos-sdk` ships a fix, or before mainnet if `x/crisis`'s invariant-checking behavior specifically matters more once real value is at stake. | `go list -m -versions github.com/cosmos/cosmos-sdk` (v0.53.8 = latest); `app/app.go:822` |
| 8 | 2026-08-07 | GitHub Dependabot (7 new alerts after pushing the cosmos/evm vendor patch) | Low (docker/docker) / Low (pion/stun) | 4 alerts on `github.com/docker/docker` (`GHSA-rg2x-37c3-w2rh`, `GHSA-vp62-88p7-qqf5`, `GHSA-x86f-5xw2-fm2r`, `GHSA-x744-4wpc-v9h2`, `GHSA-pxq6-2prw-chj9` — docker-daemon/CLI container-management bugs) and 1 on `github.com/pion/stun/v2` (`GHSA-34rh-wp3j-6cxc`, DoS via malformed STUN attribute parsing), all on `chain/go.mod` | **Risk-accepted, strong evidence for both.** `go mod why github.com/docker/docker` returns `"main module does not need package github.com/docker/docker"` — it's a Minimal-Version-Selection bookkeeping entry in `go.sum`, not imported by any source file, so it never compiles into `lcvd` at all; this is the strongest possible "not reachable" signal, stronger than the usual govulncheck-reachability argument used elsewhere in this table. `pion/stun/v2` does compile in — real chain: `lastcoinvision/lcv/app` → `cosmos/evm/x/vm/keeper` → `go-ethereum/eth/tracers` → `go-ethereum/internal/ethapi` → `go-ethereum/p2p` → `go-ethereum/p2p/nat` → `pion/stun/v2` — but a fresh `govulncheck` run post-patch shows only 6 vulnerabilities total (down from 7 pre-patch, `GO-2025-3684` now gone per Finding #6), and `pion/stun` isn't among them: `lcvd` uses CometBFT's own P2P layer for consensus networking, never go-ethereum's standalone devp2p/NAT-traversal stack that pulls this in, so the code path is compiled but never executed. Both are transitive artifacts of vendoring `cosmos-evm-patched`'s own dependency graph (its test/tooling code, not `lcvd`'s), not something LCV's own code introduced a real path to. | `go mod why github.com/docker/docker` / `github.com/pion/stun/v2`; `/tmp/govulncheck-post-push.txt` (6 vulnerabilities, no docker/pion-stun entries) |

## Closing the audit gate

`MAINNET_LAUNCH_RUNBOOK.md`'s audit gate can be marked closed only when all of the following are
true simultaneously, checked by re-reading this document's actual state rather than assumed:

1. A completed audit report (not an "initial" or interim one) has been received.
2. Every finding in it has a row in the Findings Log above with a real Outcome.
3. Every Critical and High finding's Outcome is "fixed and verified" — none are open, none are
   risk-accepted (Critical) or risk-accepted without a credible written argument (High).
4. Medium/Low findings are either fixed or have a written risk-acceptance — not silently
   dropped.

Until then, per `DECISIONS.md`'s 2026-08-07 entry, the gate stays open and says so honestly.
