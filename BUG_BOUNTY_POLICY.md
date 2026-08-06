# Bug Bounty Policy — DRAFT, NOT LIVE

This document is a **ready-to-launch draft**, not an active program. `SECURITY.md` still
correctly says there is no bug bounty program today — that stays true until this document is
approved, funded, and a launch date is set. It exists now, ahead of that, for the same reason
`AUDIT_FINDINGS_RESPONSE.md` and `VALIDATOR_HSM_RUNBOOK.md` were written before real findings or
real hardware existed: `MAINNET_LAUNCH_RUNBOOK.md`'s bug-bounty gate can be satisfied either by a
real live program *or* a documented decision to launch without one and why — deciding that
question, and having the terms ready either way, shouldn't happen under launch-week pressure.

## Why this, and why now

Once mainnet has real value, `SECURITY.md`'s current "devnet/testnet-stage, no real value at
stake" framing for having no bounty stops being true. A funded bounty gives good-faith
researchers a reason to report through a private channel instead of exploiting or disclosing
publicly — the entire point of coordinated disclosure, which only works if reporting is worth
someone's time.

## Scope

Same scope as `SECURITY.md`: `chain/`, `contracts/`, `sdk/`/`cli/`/`faucet/`, `deploy/`. Same
exclusions: `web/`/`docs/` content issues, and anything already listed as a known/accepted
limitation in `AUDIT_SCOPE.md` (a report of a documented limitation isn't a new finding and
isn't eligible for payout).

**New in-scope surface since `SECURITY.md` was written**: `packages/wallet-core`,
`packages/tx-builders`, `apps/desktop`, `apps/mobile` — the wallet apps hold real user keys the
moment they're pointed at a real network, which SECURITY.md's original scope list predates.

## Severity and payout tiers

Uses the same severity definitions as `AUDIT_FINDINGS_RESPONSE.md` (Critical/High/Medium/Low),
so a finding gets the same classification whether it comes from a paid audit or a bounty
reporter — one severity scale for the whole project, not two competing ones.

Payout ranges below are placeholders pending a funding decision — filled in with real numbers
once someone with budget authority sets them, not guessed here:

| Severity | Definition (from `AUDIT_FINDINGS_RESPONSE.md`) | Payout range |
|---|---|---|
| Critical | Direct path to fund loss, consensus failure/chain halt, double-spend, or forging a credential/identity/vote result | **TBD** — needs a funding decision |
| High | Same categories, but conditional on a specific precondition/privileged role, or a straightforward DoS | **TBD** |
| Medium | Real bug, limited impact | **TBD** |
| Low / Informational | Code-quality/defense-in-depth, no demonstrated exploit path | Typically out of scope for payout, credited only (matches most programs' norms — most Low/Informational findings aren't worth a bounty program's overhead to individually price) |

A first-cut reference point once ready to fill this in: look at what comparable-stage Cosmos
SDK chains actually pay (several publish ranges via Immunefi or their own docs) rather than
picking numbers with no anchor — not fabricated here since it depends on real treasury capacity,
not a general "best practice" number.

## Where reports go

Same private channel `SECURITY.md` already establishes (GitHub private security advisory, or
direct email if that's unavailable) — a bounty program doesn't need a new intake channel, just a
funded response on the other end of the existing one.

## Response commitment (once live)

Upgrades `SECURITY.md`'s current honest "no formal SLA, small project" framing to something a
paid program needs to actually deliver on:

- Acknowledgment within a stated window (e.g. 48h) — a real number, set when this launches, not
  "reasonable time."
- A severity classification and payout decision within a stated window, using the triage
  process already defined in `AUDIT_FINDINGS_RESPONSE.md`.
- Payment via a stated method, decided at launch (this repo doesn't manage payments today and
  won't presume how these get made).

## Platform: self-run vs. Immunefi/HackerOne/etc.

Two real paths once funded, not a false binary — pick based on how much triage overhead is
tolerable:

- **Self-run** (the private-channel approach above): zero platform fee, but the LastcoinVision
  team handles all triage, communication, and payment logistics directly.
- **A hosted platform** (Immunefi is the most common choice for Cosmos-ecosystem chains):
  platform takes a fee and handles researcher-facing UX/reputation/payment rails, in exchange
  for less direct control and a public program listing that itself signals "real value is now
  at stake here" — a call to make deliberately at launch time, not by default.

## Launching this for real

1. Set real payout numbers (needs budget authority — a treasury allocation decision, separate
   from and downstream of the treasury-multisig gate this document doesn't touch).
2. Pick self-run vs. hosted platform.
3. Set response-time commitments the team can actually meet.
4. Update `SECURITY.md` to point here instead of saying "no bug bounty program."
5. Update `MAINNET_LAUNCH_RUNBOOK.md`'s bug-bounty gate to reflect the real launched state (or,
   if the decision is to launch without one, record that decision and its reasoning there
   instead — the gate accepts either outcome, it just can't stay silently unaddressed).

Until all five happen, this stays a draft and `SECURITY.md`'s "no bug bounty" statement remains
accurate.
