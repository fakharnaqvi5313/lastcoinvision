# Validator HSM / Remote-Signer Runbook — DRAFT, NOT EXECUTED

Planning document for closing one specific item in
[`RELEASE_CHECKLIST.md`](RELEASE_CHECKLIST.md) §2 and
[`MAINNET_LAUNCH_RUNBOOK.md`](MAINNET_LAUNCH_RUNBOOK.md) §1: *"the validator set meets the
hardware/HSM/remote-signer-backed consensus key bar."* As of 2026-08-07, both `lcv-testnet-1`
validators keep `priv_validator_key.json` as a plain file on a shared VPS — functionally
adequate for a no-real-value testnet, not what this gate requires before mainnet.

**This document does not touch either live validator.** Migrating a consensus key is the single
highest-risk operation a validator operator performs — done wrong, it causes a double-sign and
gets the validator tombstoned (permanent, unrecoverable ejection, unlike a normal downtime
jail). That migration needs the actual operator's hands, on their actual infrastructure, with
their own judgment about timing — not something to run unattended. What this document provides
is the concrete plan so that work is well-understood before anyone touches production.

## Two real options, not a false choice

[`docs/docs/validator-handbook.md`](docs/docs/validator-handbook.md#key-management-and-hsm-guidance)
already names the standard tool (tmkms) and the general shape (a remote signer between the
validator process and the actual key material). There are two established ways to fill that in,
and they solve slightly different problems — pick based on how many independent operators are
willing to run infrastructure for this, not by default:

### Option A — tmkms + YubiHSM2 (single hardware-backed signer)

[tmkms](https://github.com/iqlusioninc/tmkms) runs as a separate process from `lcvd`, holds the
consensus key, and signs over a secure connection when `lcvd` asks it to — `lcvd` itself never
touches the raw key. Backed by a YubiHSM2 (~$650 hardware, industry-standard for this because
it's small, purpose-built for signing operations, and widely used across the Cosmos validator
ecosystem specifically for this role), the key is generated inside the HSM and is not
extractable even by someone with physical access to it.

- **What it protects against**: OS-level compromise of the validator host — an attacker who
  roots the box still cannot exfiltrate the key, only request signatures through tmkms while
  it's running (which is itself a smaller, more auditable attack surface than a general-purpose
  Linux host).
- **What it does not protect against**: a single point of failure. One YubiHSM2, one machine
  running tmkms — if that machine is unreachable, the validator can't sign until it's back.
- **Cost/complexity**: one piece of hardware per validator, one new always-on process
  (`tmkms start`) to operate and monitor in addition to `lcvd` itself.

### Option B — Horcrux (threshold signing across independent machines)

[Horcrux](https://github.com/strangelove-ventures/horcrux) splits the consensus key into `t`-of-`n`
shares held by independent cosigner processes (no single machine ever holds the whole key), and
uses threshold cryptography so any `t` of them jointly produce a valid signature. No HSM
hardware required.

- **What it protects against**: both key exfiltration *and* single-machine compromise/downtime —
  losing one cosigner (up to `n - t` of them) doesn't stop signing, and no single compromised
  machine can sign alone.
- **What it does not protect against**: a poorly-chosen threshold. `2-of-3` with all three
  cosigners run by the same operator on the same cloud account is a weaker guarantee than it
  looks — the independence has to be real (different providers/accounts/people), the same
  principle this repo already applied deciding the second `lcv-testnet-1` validator had to be a
  genuinely separate operator, not a second copy of the first (see `DECISIONS.md`, 2026-08-05).
- **Cost/complexity**: no special hardware, but 3+ always-on cosigner processes to run and keep
  mutually reachable instead of one signer — more moving parts, no hardware purchase.

**Recommendation for this project specifically**: Option A (tmkms + YubiHSM2) per validator is
the simpler starting point given `lcv-testnet-1` currently has exactly two independent
operators, each running their own single validator — there isn't yet a natural multi-machine
cosigner set per validator without each operator standing up 3 machines themselves. Revisit
Option B if a future validator is run by an operator with the infrastructure (and appetite) for
a proper threshold setup; it's the better long-term answer, just a bigger lift than this
project's current validator count justifies starting with.

## Migration procedure (general shape — operator executes on their own validator)

Standard Cosmos practice, not specific to this chain — every step here matters because the
failure mode (double-signing with the same key from two places at once) is catastrophic and
irreversible:

1. **Set up the new signer (tmkms or Horcrux) on separate hardware from the validator host**,
   fully configured and tested against a *test* key first — never the real
   `priv_validator_key.json` — to confirm the signer, the network path to `lcvd`, and the
   validator's `priv_validator_laddr` configuration all work end-to-end.
2. **Import (not copy) the real consensus key into the new signer**, following that tool's own
   documented import procedure, then treat the original `priv_validator_key.json` as sensitive
   material to be securely deleted once the new signer is confirmed working — not archived
   "just in case" somewhere it could later be started up in parallel by mistake.
3. **Never run both signers live at the same time against the same key.** Stop `lcvd`, point its
   `priv_validator_laddr` at the new remote signer (removing the local `priv_validator_key.json`
   from the path `lcvd` would otherwise use), start the new signer, then start `lcvd` — in that
   order, with the old key file already gone from the validator host by this point, not
   "disabled" by convention alone.
4. **Verify before declaring success**: confirm the validator is still signing blocks
   (`slashing signing-info`, the same check `lcvd-monitor.timer` already runs per
   `RELEASE_CHECKLIST.md` §2) for a real stretch of time — hours, not minutes — before
   considering the migration complete.
5. **Update `RELEASE_CHECKLIST.md` §2 and this gate in `MAINNET_LAUNCH_RUNBOOK.md`** to reflect
   the real, post-migration state, per-validator (this gate is per-validator, not "any one
   validator did it").

## Rehearsed locally with a software signer (2026-08-07)

Before this ever touches a live validator, ran the actual tmkms↔`lcvd` wiring end-to-end against
a disposable single-node local devnet, using tmkms's `softsign` provider (a plain-file software
key — dev/test only, gives none of a real HSM's protection, but speaks the exact same remote-
signer protocol a YubiHSM2-backed tmkms would). **Worked, and kept signing every block for the
full run** — this is empirical confirmation of the mechanics above, not just a plan:

- **Startup order matters and is unforgiving.** Starting `lcvd` before the remote signer is
  already up and retrying fails fast: `lcvd` calls `GetPubKey()` once at startup with a short
  timeout, and exits with `error with private validator socket client: can't get pubkey: send:
  endpoint connection timed out` if nothing answers — it does not wait indefinitely. Confirms
  step 3's ordering above is a hard requirement, not just a good habit: **start the remote signer
  first** (with `reconnect = true`, so it's actively dialing) and only then start `lcvd`.
- **Connection direction**: `lcvd`'s `priv_validator_laddr` is what it *listens* on; the remote
  signer is the one that *dials out* to that address, not the other way around (confirmed against
  `cometbft`'s own `node/setup.go` — `createAndStartPrivValidatorSocketClient` wraps a
  `SignerListener` on the `lcvd` side).
- **`protocol_version` must match the real CometBFT version** — tmkms's default generated config
  uses `v0.34` as a placeholder and logs a deprecation warning; this chain runs CometBFT v0.38.x,
  so `protocol_version = "v0.38"` is the correct value here, not the template default.
- **Real gap found, not yet closed**: tmkms logged `unverified validator peer ID!` on connect,
  because the rehearsal's `[[validator]].addr` omitted the `ID@` prefix its own generated
  template includes (`tcp://<ID>@host:port`). Without it, tmkms accepts the connection without
  verifying which validator it's actually talking to — fine for a disposable local rehearsal,
  **not fine for production**: a real migration must pin the actual validator peer ID in that
  field so tmkms refuses to sign for an unexpected/spoofed endpoint. Not resolved here since it
  needs the specific target validator's real identity, not a devnet throwaway's.

Rehearsal cleaned up afterward — no lasting state, and the standard local `make devnet` instance
(stopped to free the ports this needed) was restored to how it was found.

## What closes this gate

Both current `lcv-testnet-1` validators completing the migration above (or a future validator
set doing so from the start) — dedicated hardware, consensus key never a plain file on the
validator host, verified signing continuity after cutover. Not started; needs each operator's
own execution, hardware budget, and scheduled maintenance window.
