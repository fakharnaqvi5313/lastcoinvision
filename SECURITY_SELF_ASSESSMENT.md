# Security Self-Assessment

**Prepared 2026-08-05. This is an internal self-assessment, not a professional audit.**

This document exists to give a professional auditor (or any external reviewer) a running
start — current tool output, a manual review of the code added since the last internal pass,
and an honest account of what's been checked and what hasn't. It does not substitute for an
independent audit, and nothing in it should be represented as one. `AUDIT_SCOPE.md` remains
the authoritative scope document for briefing an auditor; this document is the *evidence*
behind that brief, refreshed as of today.

## What this is and isn't

- **Is**: automated tool output re-run fresh today, plus a manual code review of everything
  written since the last internal security pass (2026-08-03), by the same people building the
  project.
- **Isn't**: an independent audit. No external security firm has reviewed this code as of
  this writing. If you're an auditor using this as a starting point — verify everything here
  yourself; treat it as a map of where to look, not a substitute for looking.

## 1. Automated tooling — re-run fresh today (2026-08-05)

All five tools from the 2026-08-03 pass (`DECISIONS.md`, `AUDIT_SCOPE.md`) re-run against the
current `main` branch, not cached results.

| Tool | Scope | Result |
|---|---|---|
| `gosec -exclude-generated` | `chain/...` | 0 issues (unchanged from 2026-08-03) |
| `staticcheck` | `chain/...` | 213 raw hits, all in generated protobuf/pulsar code or transitional cosmos-sdk `x/crisis` deprecation warnings (SA1019) with no replacement shipped in SDK v0.53.6 yet — 0 actionable, unchanged from 2026-08-03. (An earlier run today silently produced an empty result file due to a redirect issue on my end, not a real "0 findings" — re-ran directly to confirm the real count before writing this down.) |
| `cargo audit` (both CosmWasm contracts) | `contracts/cosmwasm/*` | 2 unmaintained-but-not-vulnerable transitive deps (`derivative`, `paste`) — unchanged from 2026-08-03 |
| `npm audit` | `sdk`, `cli`, `faucet`, `web`, `docs` | **0 vulnerabilities of any severity, all 5 packages** — improved from 2026-08-03 (a handful of accepted low-severity/dev-only findings existed then; current dependency versions have none) |
| `slither` | `contracts/solidity/` | Same 9 findings as 2026-08-03: `block.timestamp` for day-scale deadline comparisons (accepted pattern) and low-level `.call{value:}()` for ETH transfers with checked return values (the correct modern pattern, not a finding) |

**Nothing new turned up by tooling.** The one real bug tooling caught in this project's
history — the fee-market `RequiredFee` integer overflow — was found and fixed 2026-08-03
(`DECISIONS.md`); it stays the single best example of the failure mode to specifically
re-check by hand (an unbounded tx-supplied `uint64` reaching arithmetic without a range check).

## 2. Manual review: code added since 2026-08-03

The custom modules (`x/identity`, `x/registry`, `x/voting`, `x/taxrail`, `x/emission`,
`x/feemarket`) haven't changed since the last pass — no need to re-review logic that hasn't
moved. What *has* been added since then is almost entirely infrastructure/tooling, not
consensus-critical code, with one exception reviewed in full below.

### `chain/cmd/gengenesis/main.go` — the `-itn-rewards-file` extension (added 2026-08-05)

This is the one piece of genuinely new logic that touches token allocation math, so it got a
real line-by-line read, not just a tool pass (which also came back clean — see §1).

**What it does**: carves Incentivized Testnet reward vesting accounts out of the
`community_airdrop` bucket, mirroring the pre-existing validator self-bond carve-out pattern.
Verified correct:
- Total-supply invariant holds exactly (the allocation math is integer-exact for all seven
  buckets at the real 10B-LCV, 6-decimal supply — no rounding dust; confirmed both
  algebraically and via a real end-to-end run against a local devnet before this was merged).
- Refuses to run if the rewards total would exceed or zero out the entire bucket.
- Vesting period math (`vestingPeriods`) correctly puts any integer-division remainder into
  the *last* period, so the sum of periods always equals the input amount exactly, never
  short by rounding.

**One real gap, not previously reported anywhere**: `gengenesis` never checks that the
operator-supplied addresses — the six allocation-bucket addresses, `-validator-self-bond-addr`,
and every address in an `-itn-rewards-file` — are pairwise distinct. If an operator
accidentally reused an address across two buckets (e.g. `-team-addr` equal to
`-investors-addr`, or an ITN reward address colliding with `-community-addr`), `gengenesis`
would silently write two separate account/balance entries for the same address into
`app_state`. This is **not silent corruption in practice** — `x/bank`'s own genesis
validation explicitly rejects duplicate balance entries for one address, so `lcvd genesis
validate-genesis` (already a mandatory step in every genesis-construction runbook in this
repo) would catch and fail loudly on this, not silently double-credit an address. But
`gengenesis` itself gives no early, specific error pointing at *which* address collided — an
operator would get a generic bank-module validation failure and have to work backward to find
the mistake. **Recommendation: add an explicit duplicate-address check to `gengenesis` with a
clear error message**, as defense in depth even though the current design already fails safe.

**One minor procedural note, not a vulnerability**: `gengenesis` falls back to the operator's
local wall-clock time as the vesting `genesisTime` if the genesis file being patched doesn't
already have one set. In every actual production runbook in this repo (`scripts/genesis/init-
testnet-genesis.sh`), `lcvd init` runs first and always sets a genesis time before `gengenesis`
touches the file, so this fallback path shouldn't trigger in the real workflow — but it's worth
an auditor's attention if the tool is ever used in a different order.

### Everything else added since 2026-08-03

Deployment scripts (`scripts/deploy/*.sh`), the `deploy/k8s/` manifests (already
end-to-end-verified against a real cluster, see `DECISIONS.md` 2026-08-05), `explorer/`
config, and website content changes — none of this is in `AUDIT_SCOPE.md`'s in-scope table
(deployment manifests are explicitly in scope for "hardcoded secrets, exposed ports," which was
checked as part of that work: no secrets committed, `config.local.sh` gitignored, VPS firewall
rules reviewed and confirmed correct after every change — see `DECISIONS.md`'s same-day
entries for the explorer and second-validator work).

## 3. Dependency / supply-chain status

As of 2026-08-05: **7 open Dependabot alerts** (down from 121 initially, then 26, then 9, then
7 across several triage rounds — full history in `DECISIONS.md`). All 7 are either:
- Confirmed **not compiled into the shipped `lcvd` binary at all** (`docker/docker`,
  `docker/cli`, `go-chi/chi` — pulled in only via a build-tag-gated tool import, verified via
  `go list -deps`), or
- **Reachable but with no upstream fix available yet** (`pion/dtls`, `pion/stun` — pulled in
  via unused `go-ethereum` P2P code; `lcvd` uses CometBFT for P2P, not devp2p).

Every alert that was reachable *and* had a fix available has already been patched (full
before/after triage table in `DECISIONS.md`'s 2026-08-05 entries). Cross-checked with
`govulncheck` (call-graph-level reachability, stricter than Dependabot's package-level
matching) in addition to `go list -deps` — see `DECISIONS.md` for the one thing `govulncheck`
additionally flagged that Dependabot didn't (`golang.org/x/crypto/openpgp`, reachable via
`cmd.Execute`, no fix available — worth an auditor's independent look at what in that call
chain actually needs the deprecated `armor` package).

## 4. Known design limitations (unchanged, restated from `AUDIT_SCOPE.md`)

- **`x/voting` does not provide ballot secrecy** — every revealed vote is a public, signed
  transaction. This is the single most important thing to understand about this codebase's
  actual guarantees; see `modules/voting/SPEC.md`'s own limitations section.
- **`x/identity`'s `ZKVerifier` is a stub** — no real zero-knowledge verification is wired up
  yet.
- Full list, with reasoning, in `AUDIT_SCOPE.md`.

## 5. Infrastructure notes (as of 2026-08-05)

- `lcv-testnet-1` runs two validators on two independently-operated VPSes. Neither meets the
  hardware/HSM/sentry-topology bar in `docs/docs/validator-handbook.md` or
  `RELEASE_CHECKLIST.md` §2 — both are bootstrap-stage, consensus keys are plaintext files
  (`test`-backend keyrings), acceptable for the current no-real-value testnet stage, not for
  mainnet. This is a deliberate, documented tradeoff (`DECISIONS.md`), not an oversight.
- Real allocation addresses exist for all 6 non-module buckets; `treasury`/`ecosystem` are
  cosmos-sdk native multisigs in *form* but not yet in substance — all signer keys currently
  live on one machine, so a single compromise still defeats the threshold. Full history in
  `DECISIONS.md`, 2026-08-05.
- VPS firewalls reviewed and tightened where found lacking (one VPS had no firewall at all
  before the second-validator onboarding; fixed same day).

## 6. What a professional audit should prioritize

Same four items `AUDIT_SCOPE.md` already lists, still the highest-value places to spend a
limited engagement's time — nothing in this pass changes that prioritization:

1. The ante handler / fee-market path — re-check for the same overflow class already found
   and fixed once.
2. `x/voting`'s commit-reveal implementation against its documented (non-)guarantees.
3. The EVM↔Cosmos bank bridge.
4. The reference Solidity/CosmWasm contracts, since they're the ones most likely to be copied
   directly into a real dApp.

Plus, newly surfaced by this pass: the `gengenesis` duplicate-address gap in §2 above, and
independent verification of the multisig/custody claims in §5 — both are process/tooling
gaps around token allocation, not runtime vulnerabilities, but exactly the kind of thing worth
an external reviewer's eyes before any real value is at stake.
