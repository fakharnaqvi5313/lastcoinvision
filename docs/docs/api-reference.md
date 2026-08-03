---
sidebar_position: 5
title: API Reference
---

# API reference

Three separate APIs, all live on a running devnet by default (see [Run a
node](./getting-started/run-a-node.md)):

| API | Default port | Protocol |
|---|---|---|
| Tendermint/CometBFT RPC | `26657` | JSON-RPC over HTTP + WebSocket |
| gRPC-gateway REST | `1317` | REST/JSON |
| gRPC | `9090` | Protobuf/gRPC |
| EVM JSON-RPC | `8545` | Ethereum JSON-RPC |

On a 4-validator `make devnet`, each validator's ports are offset by 10 from the previous —
validator1's REST is `1327`, validator2's is `1337`, and so on (see the banner `make devnet`
prints on startup for the exact list).

## Tendermint/CometBFT RPC (`:26657`)

The consensus-layer API — block/tx queries, broadcasting, subscriptions. Not
LCV-module-aware; this is standard CometBFT.

```bash
curl -s http://localhost:26657/status
curl -s http://localhost:26657/block?height=100
curl -s -X POST http://localhost:26657/broadcast_tx_sync -d '{"tx":"<base64-encoded-tx>"}'
```

## gRPC-gateway REST (`:1317`)

Every module's queries, plus each of the four custom modules. Field names are **snake_case**,
matching the `.proto` declarations directly — not the camelCase of standard protobuf3 JSON
(confirmed empirically; see the SDK's own design notes).

Standard modules (`x/bank`, `x/staking`, `x/gov`, ...) follow the usual
`/cosmos/<module>/v1beta1/...` paths. The four custom modules:

### `x/identity`

| Method | Path |
|---|---|
| Params | `GET /lastcoinvision/lcv/identity/params` |
| DID by address | `GET /lastcoinvision/lcv/identity/did/{address}` |
| Credential status | `GET /lastcoinvision/lcv/identity/credential/{id}` |
| Issuer list | `GET /lastcoinvision/lcv/identity/issuers` |
| Tier of an address | `GET /lastcoinvision/lcv/identity/tier/{address}` |

### `x/registry`

| Method | Path |
|---|---|
| Params | `GET /lastcoinvision/lcv/registry/params` |
| Asset by ID | `GET /lastcoinvision/lcv/registry/asset/{id}` |
| Asset transfer history | `GET /lastcoinvision/lcv/registry/asset/{asset_id}/history` |
| Jurisdiction authority | `GET /lastcoinvision/lcv/registry/jurisdiction/{jurisdiction_code}` |

### `x/voting`

| Method | Path |
|---|---|
| Params | `GET /lastcoinvision/lcv/voting/params` |
| Election by ID | `GET /lastcoinvision/lcv/voting/election/{id}` |
| Election results | `GET /lastcoinvision/lcv/voting/election/{election_id}/results` |
| Electoral authority list | `GET /lastcoinvision/lcv/voting/electoral-authorities` |

### `x/taxrail`

| Method | Path |
|---|---|
| Params | `GET /lastcoinvision/lcv/taxrail/params` |
| Tax rule by jurisdiction | `GET /lastcoinvision/lcv/taxrail/rule/{jurisdiction_code}` |
| Jurisdiction collected-tax balance | `GET /lastcoinvision/lcv/taxrail/balance/{jurisdiction_code}` |

```bash
curl -s http://localhost:1317/lastcoinvision/lcv/identity/params
# {"params":{"attested_credential_type":"kyc"}}
```

Transactions (any `Msg`) go through gRPC or the Tendermint RPC's `broadcast_tx_*` endpoints,
not a REST POST — this gateway is query-only. Use [`@lastcoinvision/sdk`](https://lastcoinvision.com/developers)
or `lcvd tx` for signing and broadcasting.

## gRPC (`:9090`)

The canonical, typed interface every query and message type is defined against — what the
REST gateway above and the SDK's clients both compile down to. Use
[grpcurl](https://github.com/fullstorydev/grpcurl) or a generated gRPC client directly if
REST/the SDK don't fit your use case:

```bash
grpcurl -plaintext localhost:9090 lcv.identity.Query/Params
```

## EVM JSON-RPC (`:8545`)

Standard Ethereum JSON-RPC — `eth_*`, `net_*`, `web3_*` methods, MetaMask- and
Hardhat/Foundry-compatible. See [Connect MetaMask](./getting-started/connect-metamask.md)
and [Deploy a Solidity contract](./tutorials/deploy-a-solidity-contract.md).

```bash
curl -s -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}'
# {"jsonrpc":"2.0","id":1,"result":"0xa33"}   (2611)
```

CORS is disabled by default on a fresh `lcvd init`/hand-configured node — `make devnet`
enables `enabled-unsafe-cors` for local development specifically so browser-based dApps can
call the REST API directly (see the root `DECISIONS.md`); a real testnet/mainnet deployment
should scope CORS to known origins rather than leave it wide open.
