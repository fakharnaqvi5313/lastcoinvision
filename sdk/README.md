# @lastcoinvision/sdk

TypeScript SDK for LastcoinVision (LCV): typed message-builder clients for the four custom
Cosmos SDK modules (`x/identity`, `x/registry`, `x/voting`, `x/taxrail`), Keplr and MetaMask
wallet-connect helpers, and a CosmJS `Registry` pre-loaded with every custom message type.

## Install

Not yet published to npm — see `examples/README.md` for how `create-lcv-app` and the local
examples resolve this package from source within this monorepo in the meantime.

## Usage

```ts
import { IdentityClient, createLcvRegistry } from "@lastcoinvision/sdk";
import { DirectSecp256k1HdWallet } from "@cosmjs/proto-signing";
import { GasPrice, SigningStargateClient } from "@cosmjs/stargate";

const wallet = await DirectSecp256k1HdWallet.fromMnemonic(MNEMONIC, { prefix: "lcv" });
const signingClient = await SigningStargateClient.connectWithSigner(RPC_ENDPOINT, wallet, {
  registry: createLcvRegistry(),
  gasPrice: GasPrice.fromString("0.025ulcv"),
});

const identity = new IdentityClient(REST_ENDPOINT, signingClient);

// Pure message builder — no network call, returns a CosmJS EncodeObject.
const msg = identity.msgCreateDID({ creator: address, verification_method: [], service: [] });

// Convenience: build + sign + broadcast in one call. IMPORTANT — see the
// gas-estimation caveat below before using "auto" for the fee argument.
await identity.createDID(address, { creator: address, verification_method: [], service: [] });

// Query (REST / gRPC-gateway), no wallet needed.
const params = await identity.params();
```

`RegistryClient`, `VotingClient`, and `TaxrailClient` follow the same shape. See
`examples/` for complete, runnable scripts (Node) and browser wallet-connect snippets.

### ⚠️ Gas estimation

Every convenience broadcast method (`createDID`, `taxedTransfer`, etc.) takes an optional
`fee` argument that defaults to `"auto"`. **Do not rely on that default right now** — see the
2026-08-03 entry in the root `DECISIONS.md`: this chain's gas simulation currently
underestimates real usage for essentially any transaction. Pass a fixed `StdFee` instead
(e.g. `{ amount: [{ denom: "ulcv", amount: "10000" }], gas: "300000" }`) until that's fixed.

### Wallet helpers

```ts
import { connectKeplr } from "@lastcoinvision/sdk";
const { signer, client } = await connectKeplr(window.keplr, chainInfo);

import { connectMetaMask } from "@lastcoinvision/sdk";
const { accounts, publicClient, walletClient } = await connectMetaMask(window.ethereum, network);
```

See `examples/connect-keplr.ts` and `examples/connect-metamask.ts`.

## Development

```bash
npm install
npm run build          # tsup, dual ESM/CJS + .d.ts
npm test                # vitest
npm run test:coverage   # vitest --coverage (thresholds: 80% on src/clients/**)
npm run typecheck
npm run proto:gen       # regenerate src/proto/** from ../chain/proto (see buf.gen.yaml)
```

Proto codegen uses `ts-proto@^1.181` specifically (not v2.x) — v2 switched its generated
runtime to `@bufbuild/protobuf`, which isn't compatible with CosmJS's `Registry`/
`GeneratedType` interface. Don't bump past the 1.x line without re-checking that.
