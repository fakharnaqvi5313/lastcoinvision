# Smart contracts

Reference contracts for Phase D's dual smart-contract layer — see
[modules/evm/SPEC.md](../modules/evm/SPEC.md) and [modules/wasm/SPEC.md](../modules/wasm/SPEC.md)
for the chain-side integration these exercise.

## `solidity/` — EVM reference contracts (Foundry)

- `LCVEscrow.sol` — milestone escrow for real-estate/off-plan sales.
- `InvoiceFactor.sol` — multi-party invoice financing (factoring).
- `CommerceEscrow.sol` — buyer-seller escrow with an arbiter, for general commerce.

```bash
cd contracts/solidity
forge install   # fetches forge-std into lib/, gitignored — see foundry.toml
forge test
```

Deploy to a running devnet (`json-rpc.enable = true` in `app.toml`, default `:8545`) with a
standard Hardhat or Foundry config pointed at the RPC URL — no chain-specific plugins needed;
see `modules/evm/SPEC.md`'s "Hardhat compatibility" section for a verified example.

## `cosmwasm/` — CosmWasm reference contracts

- `consent-registry` — healthcare consent grants + an append-only access log.
- `procurement-ledger` — multi-party procurement workflow state machine (Draft → Approved →
  OpenForBids → Awarded → Delivered → Paid, with Cancel before award).

```bash
cd contracts/cosmwasm
cargo test --workspace 2>/dev/null || (cd consent-registry && cargo test) && (cd procurement-ledger && cargo test)
./build.sh   # requires `wasm-opt` (binaryen) — see build.sh's own comment for why
```

Uploading either contract's `artifacts/*.wasm` requires a governance proposal
(`MsgStoreCode`) — code upload is gov-gated on this chain; instantiation of already-approved
code is permissionless. See `modules/wasm/SPEC.md`.
