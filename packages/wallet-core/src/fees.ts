import type { StdFee } from "@cosmjs/amino";

/**
 * CosmJS's auto-fee-estimation sentinel (passing the string `auto` as a fee, or
 * `SigningStargateClient`'s simulate-then-`GasPrice`-multiply path)
 * underestimates on every transaction on this chain today — a real, open
 * bug in `x/feemarket`'s gas-scaling ante handler, not specific to any one
 * message type (see `DECISIONS.md`, 2026-08-03, and the same note in
 * `sdk/examples/taxed-transfer-node.ts`). Every fee this package produces
 * must be an explicit `StdFee`. There is deliberately no function here
 * that accepts or returns the literal fee sentinel string — if a future
 * change needs it, that's a decision to make explicitly at the call site
 * once the underlying bug is fixed, not something this module should make
 * easy to reach for by accident.
 */
export function fixedFee(amount: string, denom: string, gas: string): StdFee {
  return { amount: [{ amount, denom }], gas };
}

/**
 * Starting points only — every real deployment (a new devnet reset, a
 * mainnet launch with different `feemarket` params) needs its own
 * calibration against `lcvd query feemarket params`, not these numbers
 * carried forward blindly. Values here match what's already verified
 * working in `sdk/examples/taxed-transfer-node.ts` against
 * `lcv-testnet-1`/`lcv-devnet`.
 */
export const TESTNET_DEFAULT_FEES = {
  /** A single bank-module MsgSend or similar simple message. */
  simple: fixedFee("5000", "ulcv", "200000"),
  /** A custom-module message (identity/registry/voting/taxrail), which does more work per tx. */
  moduleMsg: fixedFee("10000", "ulcv", "300000"),
} as const;
