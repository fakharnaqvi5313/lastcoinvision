# Software upgrades: cosmovisor + governance walkthrough

Replaces the "manual: stop, swap the binary, restart" note in
[the validator handbook](../../docs/docs/validator-handbook.md#software-upgrades) with a
supervised, automated upgrade path, and demonstrates the governance flow that triggers it end
to end against a live devnet — including the actual halt-and-swap, not just a "no halt
observed" check.

## How a named upgrade works

1. A binary registers an **upgrade handler** for a plan name (e.g. `"v2"`) — see
   [`chain/app/upgrades/v2`](../../chain/app/upgrades/v2/upgrades.go), wired in
   `chain/app/upgrades.go`. Even a release with no state migrations needs one: without a
   registered handler for the plan's name, a node halts at the upgrade height instead of
   crossing it — the entire point being to force an operator (or cosmovisor) to swap in a
   binary that actually knows how to handle it.
2. Governance submits and passes a `MsgSoftwareUpgrade` proposal naming that plan and a target
   height.
3. **Critically, the binary running *before* the upgrade height must not already have that
   plan's handler registered.** cosmos-sdk's `x/upgrade` module checks this on every block
   once a plan is scheduled and panics immediately — `BINARY UPDATED BEFORE TRIGGER!` — if it
   does. This was discovered the hard way while building this walkthrough; see the
   2026-08-03 DECISIONS.md entry. It means every binary this repo builds going forward
   (`chain/build/lcvd`, with `v2` always registered) is only ever the *post*-upgrade binary —
   which is exactly correct for a chain that has already crossed `v2`, but means the
   walkthrough needs a genuinely different "before" binary to demonstrate against.
4. In production, every validator stages the new binary at
   `$DAEMON_HOME/cosmovisor/upgrades/<name>/bin/lcvd` ahead of the vote passing. At the target
   height, cosmovisor detects the halt, swaps `current` to the staged binary, and restarts —
   no operator intervention at the exact height.

## Files

- `build-pre-upgrade-binary.sh` — builds a temporary "pre-v2" `lcvd` (the one line in
  `app.go` that registers the `v2` handler commented out, restored immediately after via a
  trap) — purely so the walkthrough below has a genuine "before" binary to run against.
- `cosmovisor-setup.sh` — installs cosmovisor and lays out `$DAEMON_HOME/cosmovisor/genesis/bin/`
  from `chain/build/lcvd`, for supervising a node the same way a real validator would.
- `submit-upgrade-proposal.sh` — the full, self-contained walkthrough: builds the pre-v2
  binary, starts a fresh 4-validator devnet on it, submits a real `MsgSoftwareUpgrade` naming
  the `v2` plan, votes yes from all 4 validators, waits for it to pass, confirms the pre-v2
  binary actually halts at the target height with the expected `UPGRADE "v2" NEEDED` error,
  swaps in the real (v2-aware) `chain/build/lcvd`, and confirms the chain resumes and keeps
  producing blocks past that height.

## Try it

```bash
./scripts/upgrade/submit-upgrade-proposal.sh
```

Runs to completion on its own (no `make devnet` needed first — it manages its own devnet
lifecycle against `$HOME/.lcv-devnet`, stopping and wiping any existing one). Verified
end-to-end: submit → vote → pass → **halt** (confirmed via the exact panic message) → binary
swap → resume past the upgrade height.

`cosmovisor-setup.sh` is separate because it demonstrates the *supervision* half — run it and
start a node with `cosmovisor run start` instead of `lcvd start` to see cosmovisor's own
upgrade-info.json detection and restart cycle fire for real, rather than the manual kill/swap/
restart `submit-upgrade-proposal.sh` does itself.

## Adding a real upgrade later

When there's an actual v3 with a real migration:

1. `mkdir chain/app/upgrades/v3`, write its `CreateUpgradeHandler` (this is where an actual
   state migration — a new param, a store key change — goes; `v2`'s is empty on purpose).
2. Register it in `chain/app/upgrades.go` alongside `v2` (old handlers stay registered so
   nodes state-syncing through that height still cross it cleanly — only the binary that
   predates `v3` itself must not have `v3` registered yet).
3. Bump module `ConsensusVersion()`s if any module's state format changed, so
   `mm.RunMigrations` actually has something to do.
4. Build the new binary, stage it at every validator's `cosmovisor/upgrades/v3/bin/lcvd`
   **before** the vote passes — the currently-running binary at that point genuinely
   predates `v3`, so there's no premature-upgrade conflict this time.
5. Submit the `MsgSoftwareUpgrade` proposal (same shape as `submit-upgrade-proposal.sh`, name
   `"v3"`) and vote it through.
