# SDK examples

Node examples run directly against a local devnet with [tsx](https://github.com/privatenumber/tsx)
(installed as a devDependency):

```bash
npm run devnet:query   # examples/query-modules.ts    — read-only, no wallet needed
npm run devnet:did     # examples/create-did-node.ts  — signs + broadcasts MsgCreateDID
npm run devnet:tax     # examples/taxed-transfer-node.ts — signs + broadcasts MsgTaxedTransfer
```

or invoke `tsx` directly, e.g.:

```bash
LCV_MNEMONIC="..." npx tsx examples/create-did-node.ts
```

All Node examples read connection details from environment variables and default to a
local single-node devnet (`http://localhost:26657` RPC, `http://localhost:1317` REST). See
the top-of-file comment in each script for the variables it uses.

`connect-keplr.ts` and `connect-metamask.ts` are **browser-only** — they export a function
that takes the injected wallet provider (`window.keplr` / `window.ethereum`) as a parameter
rather than reading `window` themselves, so they typecheck standalone and drop directly into
a dApp (see the `dapp-cosmos` and `dapp-evm` `create-lcv-app` templates, which do exactly
this).
