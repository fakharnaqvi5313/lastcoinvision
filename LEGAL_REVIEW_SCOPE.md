# Legal & Regulatory Review Scope

**This document is not legal advice, and nothing in this repository constitutes legal advice.**
It exists to brief whoever LastcoinVision engages as real counsel — the same role
`AUDIT_SCOPE.md` plays for a security auditor: what to review, what's already documented, and
where the project's own uncertainty is, so a lawyer's time goes toward analysis instead of
discovery. `MAINNET_LAUNCH_RUNBOOK.md`'s legal-review gate requires "real counsel" — this
document is preparation for that engagement, not a substitute for it.

## Why this matters for LastcoinVision specifically

Most "is my token a security" guidance assumes a generic fundraising token. LastcoinVision has
some things that generic guidance may not directly cover, which is exactly why this needs a
lawyer familiar with the specifics rather than a generic checklist:

- **Government-facing modules with real-world legal analogues.** `x/voting` runs elections;
  `x/registry` is a land-registry / asset-title system; `x/taxrail` routes tax payments;
  `x/identity` anchors credential hashes (W3C-style DIDs). Each of these maps onto an existing,
  regulated real-world domain (election law, property/title law, tax remittance, identity/KYC
  regulation) in a way a purely financial token doesn't — worth flagging to counsel explicitly
  rather than assuming "crypto token counsel" alone covers it.
- **Dual token/asset model.** `ulcv` (the native staking/gas token) and the tax/registry/voting
  modules are one chain, but functionally different concerns — counsel should look at whether
  `ulcv` itself needs securities analysis independent of what any particular government or
  enterprise deployment built on top of the chain does.
- **A defined token allocation table already exists** (see `DECISIONS.md`'s genesis-construction
  entries and `scripts/genesis/init-testnet-genesis.sh`): ecosystem grants 25%, foundation
  treasury 20%, team 15% (4-year vest, 1-year cliff), investors 15% (3-year vest, 1-year cliff),
  public 8%, community airdrop 5%, staking rewards 12% (emitted, not pre-allocated). Vesting
  schedules and allocation percentages are exactly the kind of fact pattern securities analysis
  turns on — this table is real and current as of this document, not illustrative.
- **No token sale has happened.** There is no whitepaper soliciting investment, no ICO, no
  public sale process anywhere in this repo or its associated properties (`lastcoinvision.com`,
  `faucet.lastcoinvision.com`) — worth confirming explicitly with counsel rather than assumed,
  since that fact is load-bearing for a lot of applicable-law analysis.
- **Testnet-only today, `2611` is a documented placeholder EVM chain id** (`modules/evm/SPEC.md`)
  — nothing currently in production has real economic value. This review's job is to clear the
  path *before* that changes, not to bless something already live.

## What's already documented (starting points, not conclusions)

- `web/token/page.tsx`'s existing disclaimer language (site copy, not legal analysis) and
  `web/lib/roadmap.ts`'s current framing of what's shipped vs. pending — useful for seeing how
  the project currently describes itself publicly, which counsel may want to revise regardless
  of the underlying legal analysis.
- The full allocation/vesting table referenced above.
- `docs/docs/validator-handbook.md` and `RELEASE_CHECKLIST.md` for how staking/validation
  actually works mechanically — relevant if counsel's analysis touches staking-as-investment-
  contract questions (a live, actively-discussed area of crypto regulatory analysis generally,
  not resolved here).

## Explicitly not addressed anywhere in this repo, and not for this document to guess at

- Which jurisdiction(s)' law actually apply — depends on where LastcoinVision is organized,
  where the team/company is based, and where token holders/validators/users are, none of which
  this document assumes or states.
- Whether `ulcv` or any allocation bucket constitutes a security under any specific framework
  (Howey, MiCA, or otherwise) — this is exactly the question counsel is being engaged to answer,
  not one this document attempts to pre-answer.
- AML/KYC obligations for the faucet, any future exchange listing, or the `x/identity` module's
  account-tier system (pseudonymous/attested/institutional) — the tiers exist on-chain today;
  whether operating them creates any regulatory obligation is a question for counsel.
- Government/enterprise customer contracts for anyone deploying `x/voting`/`x/registry`/
  `x/taxrail` in a real jurisdiction — those are separate agreements between LastcoinVision (or
  a deploying party) and that customer, out of scope for a review of the chain itself.

## What would close `MAINNET_LAUNCH_RUNBOOK.md`'s legal gate

Real engagement with counsel, a written opinion or memo addressing at minimum the token
classification question and any AML/KYC exposure identified above, and either remediation of
anything flagged or a documented, counsel-informed decision to accept a specific risk — mirroring
`AUDIT_FINDINGS_RESPONSE.md`'s "fixed or explicitly accepted in writing" pattern for the audit
gate, since the same discipline (no silent gaps) applies here. Not started; needs the user to
actually engage counsel — this document only prepares for that conversation.
