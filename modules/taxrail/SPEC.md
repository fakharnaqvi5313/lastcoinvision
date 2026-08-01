# `x/taxrail`

Code: [`chain/x/taxrail`](../../chain/x/taxrail) (see [DECISIONS.md](../../DECISIONS.md)
for why module code lives under `chain/x/` rather than here).

## Purpose

A split-payment primitive: a single transfer message that atomically routes a
governance-registered percentage to a jurisdiction's collector treasury and the remainder to
the recipient, in one transaction. All flows — collection and later spend — are on-chain and
queryable.

## Tax rules

`MsgRegisterTaxRule(authority, jurisdiction_code, rate)` is gated on the module `authority`
(defaults to `x/gov`). `rate` is a `cosmossdk.io/math.LegacyDec` in `[0, 1]`
(`ErrInvalidRate` otherwise) — one rule per jurisdiction code, upserted on each call, same
1:1 shape as `x/registry`'s jurisdiction authority.

## Taxed transfers

`MsgTaxedTransfer(sender, recipient, jurisdiction_code, amount)`:

1. Rejects any denom other than `ulcv` (`ErrUnsupportedDenom`) — see Design decision below.
2. Looks up the jurisdiction's tax rule; **no rule means the transfer is rejected outright**
   (`ErrNoTaxRule`), not silently treated as 0% — an unconfigured jurisdiction should not let
   money move tax-free by omission.
3. Computes `tax_amount = floor(rate * amount)` and `remainder_amount = amount - tax_amount`
   (so `tax_amount + remainder_amount = amount` always, by construction and integer
   arithmetic — there's no floating point and no rounding remainder left unaccounted for).
4. Sends `tax_amount` from the sender to the taxrail module account (acting as the
   jurisdiction's collector treasury) and credits the jurisdiction's ledger; sends
   `remainder_amount` directly from sender to recipient.

The response returns both `tax_amount` and `remainder_amount` as `Coin`s so callers can verify
the split without a follow-up query.

## Collector treasury and spend

The taxrail module account holds all collected tax for every jurisdiction combined; a
per-jurisdiction `JurisdictionLedger` tracks each jurisdiction's share of that pooled balance.
`MsgSpendCollectedTax(authority, jurisdiction_code, recipient, amount, reason)` is
authority-gated — **collector outflows require a governance proposal**, per the top-level
spec, there is no path for a jurisdiction to spend its own collected tax unilaterally. It
debits the jurisdiction's ledger (failing with `ErrInsufficientLedger` if the jurisdiction
hasn't collected enough) and pays out from the module account to `recipient`, with a required
on-chain `reason` string.

## State

- `tax_rule/<jurisdiction_code>` → `TaxRule` (rate, set_at).
- `jurisdiction_ledger/<jurisdiction_code>` → `JurisdictionLedger` (denom, balance) — a
  jurisdiction with no ledger record yet reads back as a zero-value ledger rather than
  "not found," since every jurisdiction implicitly starts at a zero collected balance.

## Queries

`TaxRule(jurisdiction_code)`, `JurisdictionBalance(jurisdiction_code)`. Both are hand-written
CLI commands (`x/taxrail/client/cli/query.go`) rather than AutoCLI-generated, because AutoCLI's
default aminojson response encoder doesn't respect gogoproto's `customtype` annotation and
would print `TaxRule.rate` as its raw internal scaled integer instead of a decimal — the exact
issue found and fixed for `x/feemarket`/`x/emission` in Phase A; see those modules' SPEC.md.

## Events

`EventTaxRuleRegistered`, `EventTaxedTransfer` (sender, recipient, jurisdiction_code,
tax_amount, remainder_amount), `EventCollectedTaxSpent`.

## Invariants

`LedgerMatchesModuleBalanceInvariant` — the sum of every jurisdiction's tracked
`JurisdictionLedger.balance` must equal the taxrail module account's actual `ulcv` bank
balance. `TaxedTransfer` and `SpendCollectedTax` are the only two code paths that touch either
side of this equation, and both update the ledger and move real coins in the same handler
call, so any divergence would mean a bug let one happen without the other.

## Design decision: single denom (`ulcv`) only

`JurisdictionLedger` was initially designed with `denom` as part of its store key, but
`QueryJurisdictionBalanceRequest` only takes a `jurisdiction_code` — there was no way to
specify *which* denom's balance to fetch, making a multi-denom ledger unqueryable as
specified. Rather than redesign the query to take a denom parameter (adding complexity for a
use case not in the acceptance criteria), the module now hard-scopes itself to a single
`DefaultDenom = "ulcv"` constant and explicitly rejects any other denom in both
`MsgTaxedTransfer` and `MsgSpendCollectedTax`. See [DECISIONS.md](../../DECISIONS.md) if this
needs revisiting for a future multi-denom requirement.

## Limitations

- **Single denom only** (see above) — not a general-purpose multi-asset tax primitive.
- **No per-transfer tax-rule versioning/history.** `MsgRegisterTaxRule` upserts in place; there
  is no record of what rate applied to a transfer that already happened beyond what's in that
  transfer's own emitted event — reconstructing "what was the rate on date X" requires
  scanning event history, not a dedicated query.
- **Jurisdiction ledger balances are pooled in one module account**, distinguished only by the
  bookkeeping in `JurisdictionLedger` records, not by separate on-chain sub-accounts. This is
  fine given the invariant that keeps the ledger sum equal to the true pooled balance, but it
  means a jurisdiction's "treasury" is a purely logical, not physically segregated, balance.
