# `x/emission`

Code: [`chain/x/emission`](../../chain/x/emission) (see [DECISIONS.md](../../DECISIONS.md)
for why module code lives under `chain/x/` rather than here).

## Purpose

LCV has a fixed total supply of 10,000,000,000 LCV — nothing is ever minted after genesis.
The 12% "staking rewards" allocation (1,200,000,000 LCV) is pre-funded into a module
account (`staking_rewards`) at genesis. `x/emission` is the only thing that ever moves
coins out of that account: every block it computes how much reward is due under an
exponential decay curve and transfers that amount — never mints it — into the
`fee_collector` account, where `x/distribution`'s existing `BeginBlocker` picks it up and
allocates it to the block proposer, bonded validators/delegators, and the community pool
exactly as it does with transaction fees.

`x/mint` stays enabled (per the global spec) but is pinned to 0% inflation at genesis
(`InflationMin = InflationMax = InflationRateChange = 0`), so it mints exactly zero coins
every block — `x/emission` is the sole source of new-to-circulation staking rewards.

## Decay curve

Let `t` be seconds elapsed since `params.start_time` (the genesis time, unless a genesis
file explicitly overrides it), `P = params.total_reward_pool`, and
`H = params.half_life_seconds`. Define:

```
lambda = ln(2) / H
rate(t) = lambda * P * e^(-lambda*t)          // instantaneous reward rate at time t
emitted(t1, t2) = P * (e^(-lambda*t1) - e^(-lambda*t2))   // closed-form integral of rate(t) from t1 to t2
```

`emitted(0, ∞) = P`: the curve asymptotically emits the entire pool but never exceeds it,
and every block only ever transfers `emitted(t_prev, t_now)` for the actual elapsed wall
time since the previous block (`t_prev` is persisted as `LastEmissionTime`), so the exact
formula is used rather than a per-block-average approximation — no drift accumulates
regardless of block time variance.

Default `half_life_seconds` is 2 years, chosen so that by year 10 (5 half-lives),
`1 - 0.5^5 ≈ 96.9%` of the pool has been emitted, matching the "10-year decaying emission"
description in the top-level spec while leaving a long tail beyond that horizon rather than
a hard cliff to zero.

All arithmetic (`e^-x` via a Taylor series, the pool multiplication) is done in
`cosmossdk.io/math.LegacyDec` (arbitrary-precision fixed-point), never `float64`, so every
validator computes bit-identical results — see
[`chain/x/emission/keeper/decay.go`](../../chain/x/emission/keeper/decay.go).

## State

- `Params` (`0x00` key `p_emission`): `mint_denom`, `staking_rewards_account`,
  `total_reward_pool` (`math.Int`), `half_life_seconds`, `start_time`.
- `LastEmissionTime` (key `last_emission_time`): unix seconds up to which rewards have
  already been transferred; advances every block regardless of whether the computed amount
  was positive, so a chain halt/restart never double-pays for the downtime window once it
  resumes (the gap is paid out once, in the first block after restart).

## Messages

- `MsgUpdateParams` (governance/authority-gated): updates all params atomically. Changing
  `total_reward_pool` or `half_life_seconds` takes effect only for time elapsed *after* the
  change (the closed-form integral is always evaluated between `LastEmissionTime` and the
  current block using the *current* params, so a mid-course correction doesn't retroactively
  rewrite past emissions).

## BeginBlock

Runs immediately after `x/mint` and before `x/distribution` in `app/app_config.go`'s
`beginBlockers` order (mint stays first purely so both "issuance-shaped" modules are
grouped together; the only ordering constraint that actually matters is emission-before-
distribution). Each block:

1. Compute `due = emitted(LastEmissionTime - start_time, now - start_time)`.
2. Clamp `due` to the `staking_rewards` account's actual balance (never errors on a
   depleted pool — emissions just stop).
3. Advance `LastEmissionTime` to `now` unconditionally.
4. If `due > 0`, `SendCoinsFromModuleToModule(staking_rewards, fee_collector, due)`.

## Invariants

- The `staking_rewards` module account never receives a Minter/Burner permission
  (`app/app_config.go`); it can only be drawn down by this module's `SendCoinsFromModule`
  call, and its genesis balance is fixed by the allocation script
  (`scripts/genesis/`) — so total LCV supply is invariant across the lifetime of the
  decay curve regardless of how emission plays out.
- `emitted(t1, t2) < total_reward_pool` for any finite `t2` (strict asymptote), enforced by
  construction of the closed-form formula, not by a runtime check.

## CLI note: hand-written `params` query command

`lcvd query emission params` is served by a hand-written command
([`chain/x/emission/client/cli/query.go`](../../chain/x/emission/client/cli/query.go)) rather
than the one AutoCLI would otherwise generate. See
[`modules/feemarket/SPEC.md`](../feemarket/SPEC.md#cli-note-hand-written-params-query-command)
for the full explanation — the same AutoCLI/gogoproto-customtype gap applies here.

## Limitations

- The Taylor series in `expNeg` is capped at 300 terms; this comfortably covers any
  `lambda*t` value corresponding to centuries of chain uptime at the default 2-year
  half-life; SPEC intentionally doesn't try to make this unbounded, since a value that large
  would indicate a params or clock bug worth surfacing rather than silently absorbing.
- Governance can change `total_reward_pool` upward via `MsgUpdateParams` with no on-chain
  check against the `staking_rewards` account's actual balance — the BeginBlock balance
  clamp means this can't overpay, but it also means a misconfigured increase silently has
  no effect once the account is exhausted rather than erroring. A future revision could add
  an explicit invariant comparing `total_reward_pool` against the funded balance at
  `MsgUpdateParams` time.
