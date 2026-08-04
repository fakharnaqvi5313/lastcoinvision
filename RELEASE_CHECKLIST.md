# Release Checklist

A runbook for launching a real (non-`lcv-devnet`) network — `lcv-testnet-1` or, eventually,
`lastcoinvision-1`. This is a testnet-launch checklist today; nothing here should be read as
mainnet readiness, and no launch date is implied.

**Status (2026-08-04):** `lcv-testnet-1` exists and is producing blocks on the production
VPS, as a single, team-hosted bootstrap validator — this satisfies §3's "dry-run privately
among the launch coordinator's own validators first" step, nothing more. The genesis
mechanics below are for real (production timing, real supply math, self-bond carved from
the public allocation), but the six allocation addresses are still placeholders and no
external validator has been onboarded — see DECISIONS.md for the full record. Public RPC/API
access exists (`rpc.lastcoinvision.com`, `api.lastcoinvision.com`) alongside
`faucet.lastcoinvision.com`, all pointed at this same single validator. Checked boxes below
reflect what's true of *this instance's mechanics*, not a fully launched network.

**Update, same day:** cosmovisor now supervises the validator, and a missed-block/signing-rate
monitor runs every 5 minutes (§2). Identity's acceptance flow is verified end-to-end on
`lcv-testnet-1`; registry/taxrail/voting's bootstrap is in-flight via a real governance
proposal (deposited, voted, passing 2026-08-06T17:40:52Z — real 48h timing, not compressed)
(§3). See §2–§4 below for the itemized state — several boxes are now checked, several are
explicitly left open with the reason why (external validators, HSM, a genuinely new upgrade
handler, or a policy call only the user can make).

## 1. Genesis verification

Genesis construction for a real network must **not** reuse `scripts/genesis/init-genesis.sh`
unmodified — that script's `lcvd multi-node` step auto-funds each validator's self-bond from
outside the 10B allocation table (documented in `DECISIONS.md`, 2026-08-01), which is
acceptable for a disposable devnet but not for a network whose supply figures are meant to be
final.

- [ ] **Allocation table matches the published tokenomics exactly**: ecosystem/grants 25%,
      foundation treasury 20%, team 15% (4y vest, 1y cliff), investors 15% (3y vest, 1y
      cliff), staking rewards 12%, public 8%, community/airdrop 5%. Cross-check every genesis
      account's amount and vesting schedule against the allocation table, not just the total.
      **Structurally true on `lcv-testnet-1`** (verified percentages/vesting math), **but the
      six bucket addresses are still placeholders** (`scripts/genesis/init-testnet-genesis.sh`
      generated them) — this box stays unchecked until real, custody-controlled addresses
      replace them.
- [x] **Total supply is exactly 10,000,000,000 LCV** (`10_000_000_000_000_000 ulcv`) — sum
      every genesis balance and confirm it, don't sample. Verified on `lcv-testnet-1`'s
      genesis.json (2026-08-04).
- [x] **Validator self-bonds come from the public/ecosystem allocation** via a real,
      auditable distribution process (e.g. a documented pre-launch delegation round), not
      from `multi-node`'s devnet auto-funding. `gengenesis` now has
      `-validator-self-bond-addr`/`-amount` flags that carve the self-bond out of the public
      allocation instead of minting it — see DECISIONS.md.
- [x] **`x/mint` inflation is pinned to zero** (`gengenesis`'s `--devnet-fast-gov`-independent
      behavior) — staking rewards come only from the pre-funded `staking_rewards` module
      account's decay curve (`x/emission`), confirmed via `lcvd query emission params` and by
      inspecting the exported genesis JSON's `app_state.mint` section directly.
- [x] **Vesting cliffs and end times use real calendar dates**, not the small offsets used in
      local devnet testing. (This was already true of `gengenesis`'s vesting math even for
      devnet — nothing changed here.)
- [x] **`x/gov` voting/deposit periods are production values**, not the `--devnet-fast-gov`
      30-second window used for local testing (see `DECISIONS.md`'s governance-timing notes).
      Verified on `lcv-testnet-1`: 172800s (48h) voting and deposit periods.
- [x] **CometBFT `timeout_commit` is a production-appropriate value** (the devnet's reduced
      `1s` — see `DECISIONS.md`, 2026-08-03 — is a local-testing convenience, not meant for a
      real distributed validator set with real network latency). `lcv-testnet-1` uses
      CometBFT's real default (~5s), confirmed by observed block times.
- [x] **`lcvd validate-genesis` passes on the final, distributed genesis.json** — and every
      validator's own copy hashes identically (`sha256sum genesis.json` compared across all
      validators before anyone starts). Passes; the hash-match check is trivial with a single
      validator and needs re-verifying once more validators actually join.
- [x] **Chain ID is final and correct** (`lcv-testnet-1`, or the real mainnet ID) — this
      cannot be changed after the first block without a hard fork. Running as `lcv-testnet-1`;
      a real mainnet launch still needs its own chain ID decided separately.

## 2. Validator onboarding

**Most of this still doesn't apply to `lcv-testnet-1`'s sole current validator** — it runs on
a shared VPS with a plaintext `test`-backend key and no sentry topology, which is fine for a
first bootstrap node but not what this section describes overall. The items below marked
done are genuinely done for this validator; the rest stay open for the first *real* external
validator (recruiting one is a coordination problem, not something to fake).

- [ ] Hardware meets the baseline in
      [`docs/docs/validator-handbook.md`](docs/docs/validator-handbook.md#hardware) (no
      chain-specific load-tested numbers exist yet — treat those as a floor, not a
      guarantee). Unassessed — this is a shared VPS hosting unrelated projects, not
      dedicated validator hardware.
- [ ] Consensus key generated and held per
      [`docs/docs/validator-handbook.md`](docs/docs/validator-handbook.md#key-management-and-hsm-guidance) —
      HSM/remote-signer-backed, never duplicated across hosts. `priv_validator_key.json` is a
      plain file on the VPS, not HSM-backed — real infrastructure investment needed before
      this applies.
- [x] Operator key kept separate from the validator host. The `validator0` operator key
      (used to sign the gentx) was generated and has only ever lived on the machine that ran
      `init-testnet-genesis.sh` — it was never copied to the VPS. Only the resulting signed
      gentx and the consensus key (`priv_validator_key.json`, which *should* live on the
      validator host) were deployed.
- [ ] Sentry-node topology in place before the validator ever connects to a public network —
      see [`deploy/k8s/README.md`](deploy/k8s/README.md) for the reference 3-sentry topology,
      or the equivalent for a non-k8s deployment. Not in place — this validator's P2P is
      directly reachable, a known/accepted tradeoff for a single bootstrap node, not a fix.
- [x] `cosmovisor` installed and configured (`scripts/upgrade/cosmovisor-setup.sh`) *before*
      first start, so the validator is upgrade-ready from block 1. Installed retroactively
      (2026-08-04, after first start, since this validator predates this checklist item being
      acted on) — `lcvd.service` now runs `cosmovisor run start`, confirmed via the process
      tree and continued block production after the switch. "Before first start" wasn't
      literally true, but it's true for every block from here on.
- [x] Monitoring wired up for missed blocks / signing rate — the 50%-of-100-blocks downtime
      threshold isn't a lot of margin (see the validator handbook). A systemd timer
      (`lcvd-monitor.timer`, every 5 minutes) checks `slashing signing-info` and validator
      status, logging a `WARNING` line above 20/100 missed blocks. Not a real alerting stack
      (no paging channel exists for this project) — a manual `grep WARNING` check, not a push
      notification.
- [ ] Firewall confirmed: only the sentry's P2P port is publicly reachable; the validator's
      own P2P/RPC are not. Not true here by design: this validator's own P2P (26656) *and*
      its RPC/REST (via `rpc.`/`api.lastcoinvision.com`) are both directly, intentionally
      public, since there's no sentry to front them and public testnet RPC access was
      explicitly requested. Accepted tradeoff for a bootstrap node, not the target end state.
- [x] Gentx submitted and verified against the canonical genesis before the launch coordinator
      collects and merges all validators' gentxs. Done for this validator's own bootstrap
      gentx (trivial with one validator — `collect-gentxs` and `validate-genesis` both
      passed). The "collects and merges *all* validators'" framing is about a multi-validator
      launch, which hasn't happened yet.

## 3. Testnet launch runbook

1. **Freeze genesis.** No further edits to `genesis.json` once distributed — treat it as
   immutable input from this point.
2. **Distribute genesis + persistent_peers/seeds list** to every onboarded validator. Confirm
   each validator's genesis.json hash matches before anyone proceeds.
3. **Dry-run privately** among the launch coordinator's own validators first (a private
   devnet using the *real* genesis, not a throwaway one) to catch genesis-level bugs before
   any external validator starts — see step 1's checklist for what to re-verify here.
4. **Publish the launch time** (a genesis time far enough in the future that every validator
   can start before it) and confirm each validator has started and is waiting at that height.
5. **At genesis time**, confirm the chain produces its first block and 2/3+ voting power is
   online (`lcvd query staking validators`, cross-checked against `lcvd status` on multiple
   validators).
6. **Verify the acceptance criteria from every phase still hold** on the new network, not
   just on `lcv-devnet`: a governance proposal can change fee-market params; the identity,
   registry, voting, and taxrail acceptance flows in each module's SPEC.md still pass;
   `docker compose up` from `deploy/docker-compose/` (or the k8s manifests) stands up
   explorer/faucet/website pointed at the new network correctly.
   **Partially verified (2026-08-04):** identity's `create-did`/`did-by-address` flow
   confirmed fully end-to-end on `lcv-testnet-1` (real tx, real query, matching result).
   Registry, taxrail, and voting all gate their bootstrap step (jurisdiction authority, tax
   rule, electoral authority) behind `x/gov`, so a real submission was made — proposal #1,
   bundling all three plus a `lcvfeemarket` `MsgUpdateParams` confirming the fee-market
   gov-path — deposited, entered `PROPOSAL_STATUS_VOTING_PERIOD` with the real 48h window,
   and was voted yes with the validator's full self-bond weight. **It will pass and execute
   automatically at 2026-08-06T17:40:52Z** — this can't be verified further without either
   waiting out that real window or compressing the network's own governance timing (which
   would defeat the point of using production timing in the first place). Re-check
   `lcvd query gov proposal 1` after that time and re-run the registry/taxrail/voting
   acceptance flows against `lcv-testnet-1` with the community address (now the bootstrapped
   authority for all three) once it has. `docker compose up` against this specific network
   was not attempted this pass.
7. **Confirm the security posture**: `SECURITY.md`'s reporting channel is live and monitored,
   the findings in `AUDIT_SCOPE.md` are still accurate for this network's actual genesis
   parameters (rates, periods, etc. — not just the devnet defaults this document was written
   against). **Verified (2026-08-04):** `AUDIT_SCOPE.md` documents code-level findings
   (gosec/staticcheck/fuzz results, design limitations), not devnet-specific numeric
   parameters — its claims hold identically on `lcv-testnet-1` since it runs the same audited
   code (only the cosmetic swagger-docs template changed since).
8. **Announce, with an explicit "testnet/devnet" label** — no legal, investment, or
   regulatory-approval language anywhere in the announcement, matching the constraint that
   already applies to every other public-facing surface in this repo (`web/`, `docs/`). Not
   done — a real public announcement is the user's call to make, not something to do
   unilaterally.

## 4. Post-launch

- [ ] Confirm the nightly e2e workflow (`.github/workflows/nightly-e2e.yml`) is pointed at
      the new network, not just `lcv-devnet`, if it's meant to monitor it continuously.
      **Assessed, not changed (2026-08-04):** this workflow spins up its own disposable
      4-validator devnet on the CI runner to test the *current commit's code*, and tears it
      down at the end — it isn't testing a persistent network's health, so "pointing" it at
      `lcv-testnet-1` would conflate code-correctness CI with live-network monitoring (and
      risk spamming a real network with test transactions on every run). That's a genuinely
      different job than this one — `lcvd-monitor.timer` (see §2) is the actual live-network
      monitoring piece.
- [ ] Confirm cosmovisor's `upgrades/` staging process (see `scripts/upgrade/README.md`) is
      rehearsed against this specific network before the first real governance-triggered
      upgrade is proposed. Cosmovisor itself is now installed and supervising `lcvd.service`
      (see §2), but a full rehearsal needs a genuinely new named upgrade handler that isn't
      already registered in the running binary (every binary this repo builds already has
      `v2` registered, so `v2` itself can't be used to demonstrate a real halt-and-swap on
      this network — see `scripts/upgrade/README.md`) plus a real governance proposal and
      another ~48h wait. Not attempted this pass.
- [ ] Revisit `SECURITY.md`'s "no bug bounty program" line — if launching to real external
      validators/users, decide whether that's still accurate before publishing. Not decided —
      this is a funding/commitment decision for the user, not something to resolve
      unilaterally.
