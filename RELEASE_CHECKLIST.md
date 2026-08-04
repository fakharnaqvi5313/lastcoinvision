# Release Checklist

A runbook for launching a real (non-`lcv-devnet`) network — `lcv-testnet-1` or, eventually,
`lastcoinvision-1`. This is a testnet-launch checklist today; nothing here should be read as
mainnet readiness, and no launch date is implied.

**Status (2026-08-04):** `lcv-testnet-1` exists and is producing blocks on the production
VPS, as a single, team-hosted bootstrap validator — this satisfies §3's "dry-run privately
among the launch coordinator's own validators first" step, nothing more. The genesis
mechanics below are for real (production timing, real supply math, self-bond carved from
the public allocation), but the six allocation addresses are still placeholders and no
external validator has been onboarded — see DECISIONS.md for the full record. Checked boxes
below reflect what's true of *this instance's mechanics*, not a fully launched network.

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

**None of this applies yet to `lcv-testnet-1`'s sole current validator** — it runs on a
shared VPS with a plaintext `test`-backend key, no sentry topology, and no HSM, which is
fine for a first bootstrap node but not what this section describes. Every box here stays
open for the first *real* external validator.

- [ ] Hardware meets the baseline in
      [`docs/docs/validator-handbook.md`](docs/docs/validator-handbook.md#hardware) (no
      chain-specific load-tested numbers exist yet — treat those as a floor, not a
      guarantee).
- [ ] Consensus key generated and held per
      [`docs/docs/validator-handbook.md`](docs/docs/validator-handbook.md#key-management-and-hsm-guidance) —
      HSM/remote-signer-backed, never duplicated across hosts.
- [ ] Operator key kept separate from the validator host.
- [ ] Sentry-node topology in place before the validator ever connects to a public network —
      see [`deploy/k8s/README.md`](deploy/k8s/README.md) for the reference 3-sentry topology,
      or the equivalent for a non-k8s deployment.
- [ ] `cosmovisor` installed and configured (`scripts/upgrade/cosmovisor-setup.sh`) *before*
      first start, so the validator is upgrade-ready from block 1.
- [ ] Monitoring wired up for missed blocks / signing rate — the 50%-of-100-blocks downtime
      threshold isn't a lot of margin (see the validator handbook).
- [ ] Firewall confirmed: only the sentry's P2P port is publicly reachable; the validator's
      own P2P/RPC are not.
- [ ] Gentx submitted and verified against the canonical genesis before the launch coordinator
      collects and merges all validators' gentxs.

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
7. **Confirm the security posture**: `SECURITY.md`'s reporting channel is live and monitored,
   the findings in `AUDIT_SCOPE.md` are still accurate for this network's actual genesis
   parameters (rates, periods, etc. — not just the devnet defaults this document was written
   against).
8. **Announce, with an explicit "testnet/devnet" label** — no legal, investment, or
   regulatory-approval language anywhere in the announcement, matching the constraint that
   already applies to every other public-facing surface in this repo (`web/`, `docs/`).

## 4. Post-launch

- [ ] Confirm the nightly e2e workflow (`.github/workflows/nightly-e2e.yml`) is pointed at
      the new network, not just `lcv-devnet`, if it's meant to monitor it continuously.
- [ ] Confirm cosmovisor's `upgrades/` staging process (see `scripts/upgrade/README.md`) is
      rehearsed against this specific network before the first real governance-triggered
      upgrade is proposed.
- [ ] Revisit `SECURITY.md`'s "no bug bounty program" line — if launching to real external
      validators/users, decide whether that's still accurate before publishing.
