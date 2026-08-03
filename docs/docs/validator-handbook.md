---
sidebar_position: 4
title: Validator Handbook
---

# Validator handbook

This is written for the devnet/testnet stage — there is no public LastcoinVision network to
validate yet (see the [roadmap](https://lastcoinvision.com/roadmap)). Treat the guidance
below as what running `lcvd` as a validator will require, based on this being a standard
Cosmos SDK + CometBFT chain, not as instructions for joining a network that exists today.

## Slashing conditions

These are the chain's actual configured `x/slashing` and `x/staking` parameters (queryable
with `lcvd query slashing params` / `lcvd query staking params`), not estimates:

| Condition | Parameter | Value |
|---|---|---|
| Double-signing | `slash_fraction_double_sign` | 5% of bonded stake slashed, validator tombstoned permanently |
| Downtime | `slash_fraction_downtime` | 1% of bonded stake slashed |
| Downtime window | `signed_blocks_window` / `min_signed_per_window` | Miss more than 50% of the last 100 blocks |
| Downtime jail | `downtime_jail_duration` | 10 minutes, then self-unjail with `lcvd tx slashing unjail` |
| Unbonding period | `unbonding_time` | 21 days |

Double-signing is unrecoverable — a tombstoned validator can never rejoin the active set
with the same consensus key. This is the entire reason for the sentry-node architecture and
single-instance discipline below: **the only way to double-sign is to run two processes with
the same consensus key against the same or conflicting chain state**, which is an
operational failure, not a protocol risk you can otherwise be exposed to.

## Hardware

No load testing has been run to produce chain-specific numbers — this is standard Cosmos SDK
validator guidance, which the chain's resource profile (nothing unusually heavy — CosmWasm
and the EVM layer are both optional workloads, not required for consensus participation)
doesn't currently give a reason to deviate from:

- **CPU**: 4+ cores, prefer high single-core clock speed (block execution is largely
  single-threaded)
- **RAM**: 32 GB minimum, 64 GB if running the EVM JSON-RPC and/or CosmWasm alongside
  consensus
- **Disk**: NVMe SSD, 500 GB+ and growing — IAVL tree I/O is the most common validator
  bottleneck, not CPU or RAM
- **Network**: reliable, low-latency connectivity; symmetric bandwidth, not just fast
  downloads

## Key management and HSM guidance

Two keys matter for a validator, with very different exposure:

- **Consensus key** (`priv_validator_key.json`) signs every block/vote your validator
  participates in. It must never exist in more than one running process at a time — that's
  precisely what a double-sign is. Use a remote signer (e.g.
  [tmkms](https://github.com/iqlusioninc/tmkms)) backed by an HSM or a cloud KMS in
  production; at minimum, keep it off any host that isn't the single active consensus
  process, and never copy it to a backup/failover node that might run concurrently with the
  primary.
- **Operator key** (in the `test`/`os`/`file` keyring) signs `MsgCreateValidator`,
  `MsgEditValidator`, and governance/staking transactions. This can and should live
  separately from the consensus key — ideally on a machine that isn't your validator node at
  all, since it doesn't need to be online for consensus to keep running.

## Sentry node architecture

Standard Cosmos practice: run your validator's P2P layer behind one or more sentry nodes,
with the validator itself configured with `--p2p.private-peer-ids` / `persistent_peers`
pointed only at your own sentries, `pex = false`, and no public P2P exposure. This limits
DDoS surface against the validator itself — sentries can be replaced or scaled without
touching the consensus-critical process.

## Before running a validator

- [ ] Consensus key on a single, dedicated signer (ideally HSM-backed) — never duplicated
- [ ] Operator key kept separate from the validator host
- [ ] Sentry nodes in front of the validator's P2P port
- [ ] Monitoring on missed blocks / signing rate (the 50%-of-100-blocks downtime threshold
      above is not a lot of margin)
- [ ] A tested unjail procedure, since a first downtime infraction is recoverable
- [ ] A clear understanding that a second infraction (or any double-sign) is not

## Software upgrades

`cosmovisor` is the standard Cosmos SDK upgrade-automation tool — not yet wired into this
repo's `deploy/` tooling (see the [roadmap](https://lastcoinvision.com/roadmap), Phase H).
Until then, coordinated upgrades are manual: stop at the announced upgrade height, swap the
`lcvd` binary, restart.
