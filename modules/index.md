---
slug: /
title: Module Reference
sidebar_position: 1
---

# Module reference

Each module's page below is its actual `SPEC.md` from this repository, rendered directly —
not copied, so this reference and the source can't drift apart.

## Custom modules

- [`x/identity`](./identity/SPEC.md) — DIDs, verifiable-credential hashes, account tiers
- [`x/registry`](./registry/SPEC.md) — non-fungible asset/land registry
- [`x/voting`](./voting/SPEC.md) — credential-gated commit-reveal elections
- [`x/taxrail`](./taxrail/SPEC.md) — split-payment tax routing
- [`x/emission`](./emission/SPEC.md) — decaying staking-reward emission
- [`x/feemarket`](./feemarket/SPEC.md) — EIP-1559-style dynamic base fee

## Vendored modules (configuration notes)

- [CosmWasm (`x/wasm`)](./wasm/SPEC.md) — how wasmd is configured for this chain
- [EVM (`x/vm`, `x/feemarket`, `x/erc20`, `x/precisebank`)](./evm/SPEC.md) — how the
  cosmos/evm stack is configured for this chain
