---
sidebar_position: 2
title: Connect MetaMask
---

# Connect MetaMask

LCV runs an Ethereum-JSON-RPC-compatible EVM alongside its Cosmos side, with a native
two-way bridge to the `ulcv` bank balance — see [`modules/evm/SPEC.md`](/modules/evm/SPEC).

## Add the network

With a devnet running (see [Run a node](./run-a-node.md)), add it to MetaMask manually
(Settings → Networks → Add network) or programmatically:

| Field | Value |
|---|---|
| Network name | LastcoinVision Devnet |
| RPC URL | `http://localhost:8545` |
| Chain ID | `2611` |
| Currency symbol | `LCV` |
| Currency decimals | `18` (EVM-side precision; see below) |

Confirm the RPC is actually reachable before adding it:

```bash
curl -s -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}'
# {"jsonrpc":"2.0","id":1,"result":"0xa33"}   (0xa33 = 2611)
```

Programmatically, from a dApp:

```ts
import { addLcvNetworkToMetaMask } from "@lastcoinvision/sdk";

await addLcvNetworkToMetaMask(window.ethereum, {
  evmChainId: 2611,
  chainName: "LastcoinVision Devnet",
  rpcUrl: "http://localhost:8545",
});
```

## Why 18 decimals for a 6-decimal token

`ulcv` (the Cosmos-side bank denom) has 6 decimals, but the EVM always represents native
currency at 18 decimals. `x/precisebank` is the bridge that reconciles this — it tracks the
extra precision internally (as `alcv`) so 1 LCV is worth the same on both sides without
minting any new supply. You don't need to do anything for this to work; it's why MetaMask's
balance display and `x/bank`'s balance query for the same account agree once you account for
the decimal difference.

## Key portability caveat

A MetaMask account created fresh (or imported by private key) works fine. What does **not**
work: taking a mnemonic from `lcvd keys add --algo eth_secp256k1` and importing it into
MetaMask by seed phrase — they derive different addresses, because this chain's BIP44 coin
type is still Cosmos's standard 118, not Ethereum's 60. See the [technology
page](https://lastcoinvision.com/technology) for the full explanation.

## Get test funds

```bash
curl -X POST https://faucet.lastcoinvision.com/faucet \
  -H "Content-Type: application/json" \
  -d '{"address":"0xYourMetaMaskAddress"}'
```

The faucet accepts either a bech32 (`lcv1...`) or `0x...` address.
