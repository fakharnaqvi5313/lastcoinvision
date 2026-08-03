---
sidebar_position: 3
title: Connect Keplr
---

# Connect Keplr

## Add the chain

With a devnet running (see [Run a node](./run-a-node.md)), suggest the chain to Keplr
programmatically — there's no need to hand-author the `ChainInfo` object yourself:

```ts
import { connectKeplr } from "@lastcoinvision/sdk";

const { signer, client } = await connectKeplr(window.keplr, {
  chainId: "lcv-devnet",
  chainName: "LastcoinVision Devnet",
  rpcEndpoint: "http://localhost:26657",
  restEndpoint: "http://localhost:1317",
});

const [account] = await signer.getAccounts();
console.log(account.address); // lcv1...
```

This suggests the chain (bech32 prefix `lcv`, denom `ulcv`, 6 decimals), enables it, and
returns a ready `SigningStargateClient` registered with every custom module's message types
(see [`@lastcoinvision/sdk`](https://lastcoinvision.com/developers)).

## Check the connection

```bash
curl -s http://localhost:1317/cosmos/bank/v1beta1/balances/lcv1youraddress
```

## Key portability caveat

Keplr-side accounts on this chain use the keyring's default `secp256k1` (not the
`eth_secp256k1` opt-in `lcvd keys add --algo` path), so a Keplr account and a MetaMask
account derived from the *same mnemonic* are two independent keys, not two views of one
account — see [Connect MetaMask](./connect-metamask.md) and the [technology
page](https://lastcoinvision.com/technology) for why.

## Get test funds

```bash
curl -X POST https://faucet.lastcoinvision.com/faucet \
  -H "Content-Type: application/json" \
  -d '{"address":"lcv1youraddress"}'
```
