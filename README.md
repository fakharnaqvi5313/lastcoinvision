<p align="center"><img src="brand/lcv-logo.png" alt="LastcoinVision logo" width="120"></p>

# LastcoinVision (LCV)

A government-grade, developer-first Layer-1 blockchain built on Cosmos SDK + CometBFT,
with CosmWasm and EVM smart-contract support, native IBC, and purpose-built modules for
public-sector use cases (identity, land/asset registry, voting, tax) alongside a
developer platform for building dApps.

**Status: testnet/devnet software under active development. Not for production use.
Nothing in this repository or its generated documentation/website constitutes a token
sale, investment offer, or legal/regulatory claim.**

- Ticker: `LCV` · Base denom: `ulcv` (1 LCV = 1,000,000 ulcv)
- Chain IDs: `lastcoinvision-1` (mainnet, not yet launched), `lcv-testnet-1` (testnet),
  `lcv-devnet` (local)
- Bech32 prefix: `lcv` (accounts `lcv1...`, validator operators `lcvvaloper1...`)

## Monorepo layout

| Path | Contents |
|---|---|
| [chain/](chain/) | Cosmos SDK application (Go), daemon `lcvd` |
| [modules/](modules/) | Custom `x/` modules (identity, registry, voting, taxrail, emission, feemarket) |
| [contracts/](contracts/) | CosmWasm (Rust) and Solidity reference contracts |
| [sdk/](sdk/) | TypeScript SDK, `@lastcoinvision/sdk` |
| [cli/](cli/) | `create-lcv-app` scaffolder |
| [web/](web/) | lastcoinvision.com (Next.js) |
| [docs/](docs/) | Documentation site (Docusaurus) |
| [explorer/](explorer/) | Block explorer configuration |
| [faucet/](faucet/) | Testnet faucet service |
| [deploy/](deploy/) | docker-compose devnet, k8s manifests, CI |
| [scripts/](scripts/) | Genesis generation, key management, upgrade tooling |
| [brand/](brand/) | Logo and brand assets |

See [DECISIONS.md](DECISIONS.md) for deviations from the original spec and why.

## Quick start (devnet)

```bash
make devnet
```

Boots a local 4-validator `lcv-devnet` network. See [chain/README.md](chain/README.md)
for details once the chain is scaffolded, and `make help` for all targets.

## Development

Requires Go 1.22+, Rust (stable), Node 20+.

```bash
make lint   # lint all sub-projects
make test   # unit + module tests
make build  # build lcvd and CLI tooling
```
