# CosmWasm (`x/wasm`)

Code: this module is the upstream [CosmWasm/wasmd](https://github.com/CosmWasm/wasmd) module,
vendored as a dependency rather than custom-written like `x/emission`..`x/taxrail`. This doc
covers only how it's *configured* for LastcoinVision — see wasmd's own docs for the module's
general behavior.

## Purpose

General-purpose smart-contract support via CosmWasm, alongside the EVM (see
[modules/evm/SPEC.md](../evm/SPEC.md)). Per the top-level spec: permissioned code upload,
permissionless instantiation of already-approved code.

## Version

`github.com/CosmWasm/wasmd v0.61.14`, pulled in specifically because it's the newest wasmd
release line that still targets `cosmos-sdk v0.53.6` without any `replace` directive pointing
at an unofficial fork — see [DECISIONS.md](../../DECISIONS.md) for the full version-selection
writeup (this is also why the chain's `cosmos-sdk` dependency moved from v0.50.14 to v0.53.6,
and `ibc-go` from v8 to v10, since wasmd v0.61.x and cosmos/evm both require ibc-go v10).

## Permissions

- **Code upload is governance-gated.** `Params.CodeUploadAccess` is set at genesis to
  `AccessTypeAnyOfAddresses` with the single address being the `x/gov` module account
  (`authtypes.NewModuleAddress(govtypes.ModuleName)`). Concretely: `MsgStoreCode` is only
  valid when its `sender` is the gov module account itself, which only happens when a
  governance proposal containing `MsgStoreCode` as its proposal message has passed and been
  executed. A directly-submitted `MsgStoreCode` from any regular account fails with
  `unauthorized` — verified against a live local node (`lcvd tx wasm store` from a funded,
  non-gov key returns `can not create code: unauthorized`).
- **Instantiation is permissionless.** `Params.InstantiateDefaultPermission =
  AccessTypeEverybody` — once code has been uploaded (via the governance path above), anyone
  may submit `MsgInstantiateContract` against that code ID without a further proposal. This
  matches the top-level spec's "permissionless instantiate of approved code IDs" exactly:
  approval happens once, at upload; instantiation afterward is open.
- This override lives in `app/app.go`'s `wasmAppModuleBasic` — a thin wrapper around wasmd's
  own `wasm.AppModuleBasic` that replaces only `DefaultGenesis` (which otherwise defaults to
  fully permissionless upload) with the params above. Verified live: `lcvd init` produces a
  `genesis.json` with `code_upload_access.permission: AnyOfAddresses` pointing at the gov
  address and `instantiate_default_permission: Everybody`.

## CLI / gRPC

wasmd ships its own hand-written CLI commands (`x/wasm/client/cli`), not AutoCLI-generated —
`lcvd tx wasm {store,instantiate,instantiate2,execute,migrate,...}` and
`lcvd query wasm {code,code-info,contract,contract-state,list-code,params,...}` all work
out of the box with no additional wiring needed on this chain's part; verified via
`lcvd tx wasm --help` / `lcvd query wasm --help` against the built binary.

## IBC

wasmd's own IBC handler (`wasm.NewIBCHandler`) is registered on the shared IBC router
alongside transfer and interchain-accounts (see `app/ibc.go`), so CosmWasm contracts can send
and receive IBC packets (CosmWasm's `IbcMsg`/`IbcQuery` extension points) using the same
underlying IBC core as the rest of the chain.

## Reference contracts

See [contracts/cosmwasm/](../../contracts/cosmwasm/) for `consent-registry` and
`procurement-ledger`, and [modules/evm/SPEC.md](../evm/SPEC.md) / the Solidity contracts in
[contracts/solidity/](../../contracts/solidity/) for the EVM side of Phase D.

## Limitations

- **Governance proposal turnaround for every new code upload.** There is no fast path (e.g. a
  pre-approved allowlist of uploader addresses) — every new contract *version* requires its
  own passed proposal, by design, matching "permissioned upload" literally. Migrating an
  already-instantiated contract to new code is likewise gated the same way at the code-upload
  step (the code must already exist on-chain before `MsgMigrateContract` can point at it).
- **No custom message/query plugins.** The keeper is wired with wasmd's stock message/query
  router only (`wasmkeeper.BuiltInCapabilities()`); there's no chain-specific custom module
  exposed to contracts (e.g. a "call x/registry from a CosmWasm contract" bridge). Could be
  added later via `wasmkeeper.Option` if a use case needs it.
