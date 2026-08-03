---
sidebar_position: 3
title: Deploy a Solidity contract
---

# Deploy a Solidity contract

LCV's EVM is Ethereum-JSON-RPC-compatible — standard tooling (Foundry, Hardhat) works
against it unmodified. This walks through deploying one of the reference contracts in
`contracts/solidity/` with Foundry; every command below was actually run against a live
devnet.

## Prerequisites

A running devnet with the EVM JSON-RPC endpoint reachable (default `http://localhost:8545`
— see [Run a node](../getting-started/run-a-node.md)) and [Foundry](https://getfoundry.sh)
installed (`forge`, `cast`).

## Fund a deployer account

Any EVM account works — generate a fresh one with Foundry itself rather than exporting a
key from `lcvd`'s keyring:

```bash
cast wallet new
# Address:     0xcfA91013e4dD9Ce35Ff0ca014F3b3170d944e9D9
# Private key: 0x...
```

Fund it — since the account was generated as a raw EVM key (not via `lcvd keys add
--algo eth_secp256k1`), send `ulcv` to its bech32 form via a bank transfer from a funded
account:

```bash
BECH32=$(lcvd query evm 0x-to-bech32 0xcfA91013e4dD9Ce35Ff0ca014F3b3170d944e9D9 --home ~/.lcv-devnet/validator0)

lcvd tx bank send validator0 "$BECH32" 20000000ulcv \
  --home ~/.lcv-devnet/validator0 --keyring-backend test --chain-id lcv-devnet \
  --gas 300000 --fees 10000ulcv -y
```

Confirm the balance shows up EVM-side, at 18-decimal precision (`x/precisebank` handles the
6-to-18-decimal bridge — see [`modules/evm/SPEC.md`](/modules/evm/SPEC)):

```bash
cast balance 0xcfA91013e4dD9Ce35Ff0ca014F3b3170d944e9D9 --rpc-url http://localhost:8545
# 20000000000000000000   (20 LCV, at 18 decimals)
```

## Build and deploy

```bash
cd contracts/solidity
forge build

forge create src/LCVEscrow.sol:LCVEscrow \
  --rpc-url http://localhost:8545 \
  --private-key <your-private-key> \
  --broadcast
```

```
Deployer: 0xcfA91013e4dD9Ce35Ff0ca014F3b3170d944e9D9
Deployed to: 0xA6508b03aD8232A3AB564c1Ff0C90eD666013da7
Transaction hash: 0x40526c9bc4e9f3b6eae70c919d66d2ab5f01bad0d73896ddf6892cb0fa1267f3
```

`LCVEscrow` has no constructor arguments — `createEscrow(...)` is a factory-style entry
point called per-escrow after deployment, not a per-deal constructor.

## Verify it deployed

```bash
cast code 0xA6508b03aD8232A3AB564c1Ff0C90eD666013da7 --rpc-url http://localhost:8545
# 0x60806040526004361061006f575f3560e01c80639f43811f1161004d...

cast call 0xA6508b03aD8232A3AB564c1Ff0C90eD666013da7 "nextEscrowId()(uint256)" \
  --rpc-url http://localhost:8545
# 0
```

## Using Hardhat instead

An unmodified Hardhat config works the same way — point its network config at
`http://localhost:8545` with chain ID `2611` and a funded private key, then `npx hardhat
run scripts/deploy.js --network lcv-devnet`. No LCV-specific plugin or config is required.

## Next

For a fuller walkthrough of actually using this contract (deposits, milestone releases,
refunds), see [Tokenized escrow for real estate](./tokenized-escrow-real-estate.md).
