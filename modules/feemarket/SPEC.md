# `x/feemarket`

Code: [`chain/x/feemarket`](../../chain/x/feemarket) and [`chain/app/ante.go`](../../chain/app/ante.go)
(see [DECISIONS.md](../../DECISIONS.md) for why module code lives under `chain/x/` rather
than here).

## Purpose

An EIP-1559-style dynamic base fee for standard Cosmos transactions: a per-block
`base_fee` (in `ulcv` per unit of gas) that rises when blocks run above a target gas usage
and falls when they run below it, replacing a flat governance-set minimum gas price. Of
every transaction's base-fee payment, 30% is burned and 70% goes to the fee collector for
`x/distribution` to allocate to the block proposer, bonded validators/delegators, and the
community pool exactly as with any other fee — both the burn percentage and every other
parameter are governance-adjustable.

## Params

| Field | Meaning |
|---|---|
| `fee_denom` | denom the base fee is charged in (`ulcv`) |
| `base_fee` | current base fee, `ulcv` per unit gas (`math.LegacyDec`) |
| `min_base_fee` | floor `base_fee` can never adjust below |
| `target_gas` | per-block gas usage the market targets |
| `max_base_fee_change_denominator` | caps `base_fee` movement to `1/denominator` per block (8 ⇒ max 12.5%, matching Ethereum's EIP-1559 default) |
| `burn_percentage` | fraction of the base-fee payment that is burned (default 0.30) |

## Base fee adjustment (`EndBlock`)

Standard EIP-1559 formula, using the gas the block just consumed
(`ctx.BlockGasMeter()`):

```
delta      = base_fee * (gasUsed - target_gas) / target_gas / max_base_fee_change_denominator
newBaseFee = max(base_fee + delta, min_base_fee)
```

A block exactly at target leaves `base_fee` unchanged; a full block (2x target, the
implicit gas limit) raises it by the maximum step; an empty block lowers it by the same
step, floored so it can never reach zero (a zero base fee would make spam transactions
free). See [`chain/x/feemarket/keeper/abci.go`](../../chain/x/feemarket/keeper/abci.go).

## Fee enforcement and collection (ante handler)

`chain/x/feemarket/ante.Decorator` replaces the standard `x/auth/ante.DeductFeeDecorator`
in the ante chain built by [`chain/app/ante.go`](../../chain/app/ante.go) (the default
ante handler auto-wired by the `"tx"` module config is disabled via `SkipAnteHandler: true`
in `app/app_config.go` specifically so this swap can happen). Every other decorator
(signature verification, sequence checks, memo/timeout validation, etc.) is identical to
the cosmos-sdk default — only fee handling changes.

Per transaction:

1. Reject if any part of the fee is paid in a denom other than `fee_denom`.
2. Compute `required = base_fee * gasWanted`, rounded up. Reject if the fee paid is less.
3. Move the full fee paid from the payer into a transient `feemarket` module account
   (registered with `Burner` permission in `app/app_config.go`, otherwise unprivileged).
4. Burn `required * burn_percentage` from that account.
5. Forward everything else — the un-burned base-fee remainder, plus any amount paid above
   `required` (a validator tip) — to the fee collector.

## State

- `Params` (key `p_feemarket`): see table above.

## Invariants

- `base_fee >= min_base_fee` always (enforced both by `Validate()` on `MsgUpdateParams`/genesis
  and by the floor clamp in `AdjustBaseFee`).
- Every fee collected is fully accounted for: `burn + forwarded == fee paid` exactly (no
  rounding remainder can be silently lost — see `TestRequiredFeeAndBurnSplit` in
  `x/feemarket/keeper/abci_test.go`).

## CLI note: hand-written `params` query command

`lcvd query feemarket params` is served by a hand-written command
([`chain/x/feemarket/client/cli/query.go`](../../chain/x/feemarket/client/cli/query.go))
rather than the one AutoCLI would otherwise generate from `autocli.go`. AutoCLI's response
encoder (`cosmossdk.io/client/v2/autocli`'s aminojson-based printer) only special-cases a
couple of well-known proto message types (`google.protobuf.Duration`,
`cosmos.base.v1beta1.DecCoin`) — it has no hook for gogoproto's
`(gogoproto.customtype) = "cosmossdk.io/math.LegacyDec"` field annotation, so it printed
`base_fee`/`min_base_fee`/`burn_percentage` as their raw internal scaled integers (e.g.
`"500000000000000000"` instead of `"0.5"`, since `LegacyDec` scales by `1e18` internally).
The hand-written command goes through `clientCtx.PrintProto` → `codec.ProtoCodec.MarshalJSON`
instead, which does respect customtype and renders correctly. `AutoCLIOptions` sets
`EnhanceCustomCommand: true` so any future query RPCs not covered by a custom command are
still auto-generated as normal.

## Limitations

- `target_gas` is a fixed governance parameter, not derived from the chain's actual
  consensus-params gas limit; operators changing the block gas limit should update
  `target_gas` via governance to match, or the market will chase a stale target.
- The ante decorator requires single-denom fees; multi-denom fee payments (e.g. paying part
  in `ulcv` and part in an IBC-bridged token) are rejected outright rather than partially
  honored, favoring predictability over flexibility.
- A validator's local `app.toml` `minimum-gas-prices` no longer does anything: the standard
  ante decorator that reads it was removed along with `DeductFeeDecorator`. The network-wide
  `base_fee` is the only floor, by design (a single unified fee market rather than
  per-validator overrides) — but it means raising `minimum-gas-prices` is not a way for an
  individual validator to opt out of low-fee spam the way it is on an unmodified SDK chain.
