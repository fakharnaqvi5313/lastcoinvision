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

*(Empty — no audit has delivered a findings list yet. AftPak's initial report, verified in
`DECISIONS.md` 2026-08-06, contained automated-tooling output already covered by
`AUDIT_SCOPE.md`/`SECURITY_SELF_ASSESSMENT.md`, not a discrete findings list requiring this
process. This section gets a row per finding once a real one arrives.)*

| # | Date | Source | Severity | Description | Outcome | Evidence |
|---|------|--------|----------|-------------|---------|----------|
| — | — | — | — | — | — | — |

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
