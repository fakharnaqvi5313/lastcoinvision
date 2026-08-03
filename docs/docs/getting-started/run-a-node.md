---
sidebar_position: 1
title: Run a node
---

# Run a node

This walks through building `lcvd` and running a local single-node devnet from a clean
checkout — every command on this page has been run against a live devnet, not just written.

## Prerequisites

- Go 1.22+
- This repository, cloned locally

## Build

From the repository root:

```bash
make build
```

This builds `chain/build/lcvd` (the node daemon) and `chain/build/gengenesis` (the genesis
allocation-patching tool used below).

## Run a local devnet

The fastest path is the root `Makefile`, which generates a fresh 4-validator genesis (fixed
10B LCV supply, full allocation table) and starts all 4 validators as local background
processes:

```bash
make devnet
```

```
==> Devnet running. Endpoints:
  validator0: RPC http://localhost:26657  API http://localhost:1317  gRPC localhost:9090
  validator1: RPC http://localhost:26667  API http://localhost:1327  gRPC localhost:9100
  validator2: RPC http://localhost:26677  API http://localhost:1337  gRPC localhost:9110
  validator3: RPC http://localhost:26687  API http://localhost:1347  gRPC localhost:9120
```

Check it's producing blocks:

```bash
curl -s http://localhost:26657/status | jq '.result.sync_info.latest_block_height'
```

Stop it (keeps chain state) or wipe it entirely:

```bash
make localnet-stop    # stop, keep state — `make devnet` resumes it
make localnet-reset   # stop and delete all state
```

### Running a single node manually

If you want one node with a custom chain-id instead of the 4-validator devnet (useful for
SDK/dApp development against a throwaway chain):

```bash
lcvd init mynode --chain-id my-chain-1 --home /tmp/mychain
lcvd keys add validator --home /tmp/mychain --keyring-backend test
lcvd genesis add-genesis-account validator 1000000000000ulcv --home /tmp/mychain --keyring-backend test
lcvd genesis gentx validator 100000000ulcv --chain-id my-chain-1 --home /tmp/mychain --keyring-backend test
lcvd genesis collect-gentxs --home /tmp/mychain
lcvd start --home /tmp/mychain --chain-id my-chain-1 --minimum-gas-prices 0ulcv
```

:::tip
Always pass `--chain-id` explicitly to `lcvd start` when scripting a single custom node —
`lcvd`'s other subcommands (`tx`, `query`, `keys`) default `--chain-id` to `lcv` for
convenience, but `start` correctly falls back to whatever `genesis.json` actually says once
that default isn't forced on it.
:::

## Joining a public testnet

There is no public LastcoinVision testnet yet — see the [roadmap](https://lastcoinvision.com/roadmap)
for what's shipped versus planned. Everything above (devnet, single custom node) is
local-only. When a public testnet launches, this page will be updated with its chain-id,
seed nodes, and genesis file — check back here or the [faucet](https://faucet.lastcoinvision.com)
page rather than assuming a testnet exists.

## Next steps

- [Connect MetaMask](./connect-metamask.md)
- [Connect Keplr](./connect-keplr.md)
- [Module reference](/modules/)
