# Release Checklist

A runbook for launching a real (non-`lcv-devnet`) network — `lcv-testnet-1` or, eventually,
`lastcoinvision-1`. This is a testnet-launch checklist today; nothing here should be read as
mainnet readiness, and no launch date is implied.

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
- [ ] **Total supply is exactly 10,000,000,000 LCV** (`10_000_000_000_000_000 ulcv`) — sum
      every genesis balance and confirm it, don't sample.
- [ ] **Validator self-bonds come from the public/ecosystem allocation** via a real,
      auditable distribution process (e.g. a documented pre-launch delegation round), not
      from `multi-node`'s devnet auto-funding.
- [ ] **`x/mint` inflation is pinned to zero** (`gengenesis`'s `--devnet-fast-gov`-independent
      behavior) — staking rewards come only from the pre-funded `staking_rewards` module
      account's decay curve (`x/emission`), confirmed via `lcvd query emission params` and by
      inspecting the exported genesis JSON's `app_state.mint` section directly.
- [ ] **Vesting cliffs and end times use real calendar dates**, not the small offsets used in
      local devnet testing.
- [ ] **`x/gov` voting/deposit periods are production values**, not the `--devnet-fast-gov`
      30-second window used for local testing (see `DECISIONS.md`'s governance-timing notes).
- [ ] **CometBFT `timeout_commit` is a production-appropriate value** (the devnet's reduced
      `1s` — see `DECISIONS.md`, 2026-08-03 — is a local-testing convenience, not meant for a
      real distributed validator set with real network latency).
- [ ] **`lcvd validate-genesis` passes on the final, distributed genesis.json** — and every
      validator's own copy hashes identically (`sha256sum genesis.json` compared across all
      validators before anyone starts).
- [ ] **Chain ID is final and correct** (`lcv-testnet-1`, or the real mainnet ID) — this
      cannot be changed after the first block without a hard fork.

## 2. Validator onboarding

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
