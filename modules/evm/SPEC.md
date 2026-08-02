# EVM (`x/vm`, `x/feemarket`, `x/erc20`, `x/precisebank`)

Code: these four modules are the upstream [cosmos/evm](https://github.com/cosmos/evm)
stack, vendored as a dependency (like `x/wasm`, see [modules/wasm/SPEC.md](../wasm/SPEC.md)),
not custom-written. This doc covers how they're *configured* for LastcoinVision.

## Purpose

Full Ethereum JSON-RPC / EVM smart-contract support alongside CosmWasm, with the chain's
existing `ulcv` bond denom (unchanged since Phase A, 6 decimals) usable directly as the EVM's
native gas/value currency — no separate wrapped token, no bridge transaction required.

## Version and why it forced a chain-wide upgrade

`github.com/cosmos/evm v0.6.1`, paired with `wasmd v0.61.14` — the newest wasmd release still
on an *official* (non-forked) `cosmos-sdk` release. No version of wasmd or cosmos/evm targets
the `cosmos-sdk v0.50.14` this chain was on through Phase C; the only combination that does
pulls `cosmos-sdk` via a `replace` directive to an unofficial fork commit, rejected as a
materially worse choice for a chain marketed as government-grade. This is why `cosmos-sdk`
moved to v0.53.6 and `ibc-go` to v10 — see [DECISIONS.md](../../DECISIONS.md) for the full
writeup, which also covers the resulting `app.go` rewrite off depinject to manual wiring
(neither wasmd nor cosmos/evm ship a depinject module).

## EVM chain ID

`app.EVMChainID = 2611` — an arbitrary placeholder for devnet/testnet. **A real mainnet
launch must first reserve a permanent id via the
[ethereum-lists/chains](https://github.com/ethereum-lists/chains) registry** (chainlist.org)
before this value is treated as final; reusing an unregistered id risks silent replay
collisions with any other chain that happens to pick the same number.

`app.New(...)` resolves the EVM chain id from `appOpts` (the `evm.evm-chain-id` app.toml key,
defaulting to `2611` via `cmd/lcvd/cmd/config.go`'s `initAppConfig`) rather than hardcoding the
constant directly in the keeper-construction call — this is load-bearing, not just tidiness:
`evmtypes.SetChainConfig` is a **process-global, set-once** value, and `cmd/lcvd/cmd/root.go`
builds a throwaway `app.New(...)` (with no app.toml, so `evm.evm-chain-id` resolves to
`evmtypes.DefaultEVMChainID`) purely to read back codec/AutoCLI info before the real app is
built for `start`. `SetChainConfig`'s guard specifically allows exactly one subsequent call to
override a chain id still equal to that sentinel — so if the throwaway app used this chain's
real id instead, the *real* app's later call would panic with "chainConfig already set".
Confirmed against a live local node.

## Precision bridge: `ulcv` (6 decimals) as the EVM's native currency

The EVM (and Ethereum tooling generally) assumes 18-decimal arithmetic internally. This chain
keeps `ulcv` at 6 decimals (unchanged since Phase A — no token migration, no re-denomination):

- `x/precisebank` is cosmos/evm's own module built exactly for this: it tracks an
  `EVMExtendedDenom = "alcv"` (atto-`lcv`, 18 decimals) internally, backed 1:1 by real `ulcv`
  balances, so every EVM state transition, gas calculation, and `eth_getBalance` result is a
  full-precision 18-decimal number *without minting a single unit of new supply* — the
  extended denom is purely an internal accounting representation, never a real, independently
  transferable token.
- Bank denom metadata for `ulcv`/`lcv` (`bankAppModuleBasic` in `app/app.go`) is required, not
  optional: `x/vm`'s `InitGenesis` looks up `ulcv`'s decimal exponent from this metadata to
  compute the bridge; without it, genesis fails with "denom metadata could not be found".
- **Verified against a live local node:** a plain EVM transfer (`cast send ... --value 10ether`
  from one account to another) left the sender's `eth_getBalance` at `989.999992896959454`
  LCV and the recipient's at exactly `10.0` LCV, while `lcvd query bank balances` on the same
  two bech32 addresses showed `989999992 ulcv` and `10000000 ulcv` respectively — the two
  views of the same balance agree down to `ulcv`'s own 6-decimal precision, confirming the
  bridge is genuinely bidirectional and not a one-way snapshot.

## Base fee (EVM feemarket) had to be rescaled for 6 decimals

cosmos/evm's `x/feemarket` (its own EIP-1559 base-fee market for EVM transactions — see
"Two separate fee markets" below) ships a stock default `BaseFee = 1_000_000_000`, calibrated
for an 18-decimal-native chain (i.e. "1 gwei" in real terms). Left unmodified on this 6-decimal
chain, that same raw number means **1,000,000,000 ulcv per gas unit — 1,000 LCV per gas**,
found live: every `cast send` failed with `insufficient funds for gas * price + value` because
no test account could afford even a 21,000-gas plain transfer. `evmFeemarketAppModuleBasic` in
`app/app.go` scales the genesis default down by `10^(18-6) = 10^12`, giving the same
real-world gas price the upstream default intends (confirmed live: `eth_gasPrice` after the
fix reads in the hundreds-of-millions-of-wei range, consistent with a sub-cent gas price, not
the ~10^19-wei value observed before the fix).

## Two separate fee markets, deliberately not unified

This chain now has **two independent EIP-1559-style fee markets**:

- `x/lcvfeemarket` (Phase A's own module, renamed from `x/feemarket` — see below) prices and
  burns fees for every non-EVM transaction: our registry/voting/taxrail/wasm messages, and
  standard SDK messages (staking, gov, bank, etc.).
- cosmos/evm's own `x/feemarket` prices EVM transactions (`MsgEthereumTx`) exactly the way any
  other cosmos/evm chain does, which is required for JSON-RPC compatibility — MetaMask and
  every Ethereum tool assumes the EIP-1559 `baseFeePerGas`/`maxFeePerGas`/`maxPriorityFeePerGas`
  semantics cosmos/evm's feemarket implements; substituting a bespoke mechanism here would
  break MetaMask compatibility, the core Phase D requirement.

`app/ante.go` was forked from cosmos/evm's own `ante.NewAnteHandler` (not imported — its two
branch-builder functions are unexported) specifically so a transaction's *type* decides which
fee market applies: a `MsgEthereumTx` (identified by its `ExtensionOptionsEthereumTx` extension
option) goes through cosmos/evm's own decorator chain unmodified; everything else goes through
this chain's original Phase A decorator chain, with `feemarketante.NewDecorator` in the exact
position cosmos/evm's own `cosmosante.NewMinGasPriceDecorator` + `ante.NewDeductFeeDecorator`
occupy in its reference chain. Both fee markets' burn/collection logic run independently and
neither is aware of the other. `evmmonoante.NewGasWantedDecorator` is kept on the Cosmos branch
too (not swapped out), since cosmos/evm's own base-fee adjustment needs total per-block gas
usage regardless of which fee mechanism actually charged for a given tx.

## Module name collision: `x/feemarket` → `x/lcvfeemarket`

cosmos/evm's own module is *also* named `"feemarket"` — module names must be unique per app
(they key the module manager, store, and genesis state maps) and that name is upstream/fixed.
This chain's own Phase A fee-market module's `ModuleName`/`StoreKey` were renamed to
`"lcvfeemarket"` (`x/feemarket/types/keys.go`); its proto package (`lcv.feemarket.v1`, and
therefore its gRPC/REST namespace) was untouched, since Go module names and proto package
names are independent — only the Go-level identity needed to change. `lcvd query lcvfeemarket
params` / `lcvd tx lcvfeemarket ...` name our own module going forward; `lcvd query feemarket
params` now refers to cosmos/evm's.

## Keys: eth_secp256k1

`cmd/lcvd/cmd/root.go` registers `hd.EthSecp256k1Option()` as an available keyring algorithm,
so `lcvd keys add <name> --algo eth_secp256k1` derives a Keccak256-based address — the same
20-byte address space a MetaMask account occupies, verifiable via `lcvd query evm
bech32-to-0x`/`0x-to-bech32`. This is **not** the keyring's default algorithm (still plain
`secp256k1`, matching every other Cosmos chain's convention) — pass `--algo eth_secp256k1`
explicitly for any account meant to be usable from both `lcvd` and an EVM wallet with the same
key.

**Known limitation:** the chain's BIP44 coin type (`ChainCoinType = 118`, Cosmos's standard,
unchanged since Phase A) is not overridden to `60` (Ethereum's standard) for `eth_secp256k1`
keys. This means a *mnemonic* generated by `lcvd keys add --algo eth_secp256k1` will **not**
reproduce the same address if imported into MetaMask by seed phrase (MetaMask always derives
along `m/44'/60'/...`) — confirmed live: `lcvd`'s derived address and `cast wallet address
--mnemonic "..." --mnemonic-derivation-path "m/44'/60'/0'/0/0"` on the same mnemonic did not
match. This does **not** block MetaMask usage itself — a MetaMask account (created fresh, or
imported by raw private key rather than seed phrase) works identically regardless, and every
JSON-RPC / eth_sendRawTransaction path is unaffected — but full seed-phrase portability between
`lcvd` and MetaMask is not provided. Revisit if that portability becomes a real requirement.

## JSON-RPC

Enabled via `[json-rpc] enable = true` in `app.toml` (`false` by default, matching cosmos/evm's
own conservative default — a fresh `lcvd init` does not expose the RPC surface until an
operator opts in). Address `127.0.0.1:8545` (HTTP) / `127.0.0.1:8546` (WS), both cosmos/evm
defaults, unmodified. `cmd/lcvd/cmd/commands.go` wires `cosmosevmserver.AddCommands` (which
layers JSON-RPC server startup into `start`) in place of the plain SDK `server.AddCommands`;
`pruning`/`snapshot` keep using the plain `servertypes.Application` interface via a thin
adapter (`newAppForServerTypes`), since those SDK commands were never updated upstream to know
about cosmos/evm's wider `Application` interface even though every concrete app value here
satisfies both.

**Verified against a live local node:** `eth_chainId` → `0xa33` (2611), `net_version` →
`"2611"`, `eth_gasPrice` → a sane sub-gwei-equivalent value post-fix, and a full
`eth_sendRawTransaction`-equivalent transfer (via `cast send`, which speaks the identical
JSON-RPC protocol MetaMask uses) succeeded end-to-end as described above.

## Precompiles

`precompiletypes.DefaultStaticPrecompiles(...)` — every precompile cosmos/evm ships by default
(P256, bech32 address conversion, staking, distribution, ICS-20, bank, gov, slashing) is
active from genesis (`Params.ActiveStaticPrecompiles = evmtypes.AvailableStaticPrecompiles`).
The **bank precompile** is what most directly satisfies "bidirectional bank bridge": it lets
any EVM contract read (and, via the ICS-20/bank precompiles' write paths, move) real Cosmos
bank balances for any denom, `ulcv` included, without needing a separate wrapped-token
contract. `evmtypes.DefaultPreinstalls` (e.g. a standard `Create2` deployer) are enabled too.

## Hardhat compatibility

Verified live: an unmodified, standard Hardhat config (`solidity`, a `networks.devnet.url`
pointed at `:8545`, and an `accounts` private key — no chain-specific plugins or patches)
successfully deployed `contracts/solidity/src/LCVEscrow.sol` via `npx hardhat run
scripts/deploy.js --network devnet` and called a state-changing method on it afterward,
against a fresh local devnet. Deployed bytecode was independently confirmed present on-chain
via `eth_getCode`.

## Limitations

- **No WERC20 (wrapped-native-token) precompile registered.** cosmos/evm supports deploying a
  fixed-address ERC-20-interface wrapper over the native denom (the way WETH wraps ETH) via
  `erc20types.NativePrecompiles` in genesis — not configured here. The bank precompile already
  gives EVM contracts full native-denom access; a WERC20 wrapper is an ERC-20-interface
  convenience for Solidity code that specifically expects that interface (e.g. some DeFi
  patterns), not a functional requirement, and wasn't required for this phase's acceptance
  criteria. Can be added later without any migration, since it's a genesis-params addition.
- **No `erc20.NewIBCMiddleware` or EVM-side IBC callbacks.** cosmos/evm's reference app stacks
  an ERC-20 auto-token-pair-registration IBC middleware and its own IBC-callbacks contract
  keeper onto the transfer/ICA channels; this chain's transfer/ICA stack instead keeps only
  `x/wasm`'s own IBC callback support (wired in the wasmd milestone) — ibc-go v10's callbacks
  middleware supports exactly one contract-keeper implementation per stack layer, and wasm and
  EVM each ship their own, mutually exclusive one. Foreign IBC-bridged denoms won't
  automatically appear as ERC-20 token pairs, and EVM contracts don't get IBC packet lifecycle
  callbacks; CosmWasm contracts still do. Revisit with a multiplexing contract-keeper wrapper
  if EVM-side IBC callback support becomes a real requirement.
- **EVM chain id (2611) is a devnet/testnet placeholder** — see "EVM chain ID" above.
- **The experimental EVM mempool (`evmmempool.NewExperimentalEVMMempool`) may stall block
  production after a rejected broadcast.** Observed once live: after a transaction was
  rejected for a nonce mismatch, the devnet's block height stopped advancing entirely (while
  the process kept running and consuming CPU) until restarted with fresh state. A fresh devnet
  instance was unaffected and is what the Hardhat/MetaMask verifications above were run
  against. **Two rounds of follow-up reproduction attempts, all unsuccessful** (blocks kept
  advancing normally in every case):
  1. Both nonce-error variants individually — "tx nonce is higher than account nonce" (future
     gap) and "tx nonce is lower than account nonce" (stale replay, the exact error text from
     the original incident) — against a fresh devnet.
  2. The same stale-nonce error combined with a large (~12KB, contract-deployment-sized)
     payload submitted multiple times.
  3. A stale-nonce replay of an already-mined nonce, immediately followed by a burst of
     *valid, higher* nonces from the same account (testing the hypothesis that
     `mempool.InsertInvalidNonce` — called for **both** `ErrNonceGap` and `ErrNonceLow`, per
     `mempool/check_tx.go`, despite its doc comment only describing the gap case — might
     retain a permanently-unfillable low-nonce entry that blocks the account's later,
     otherwise-valid transactions behind it, since EVM mempools require gapless per-account
     nonces). All three higher-nonce transactions were mined normally; the underlying
     `legacypool`'s own `Add`/`validateTx` path apparently still rejects a truly-stale nonce
     outright rather than retaining it, so this plausible-looking code path did not reproduce
     the stall either.
  4. Firing 30 concurrent read-only RPC calls, and separately 10 concurrent broadcasts of the
     *exact same* transaction (identical nonce), to probe for a locking/race condition in the
     mempool's `Add`/`Remove`/`Select` methods. No effect in either case.

  The original incident involved several successful transactions from a *retrying* Hardhat
  deployment script (which resubmits after a client-side timeout, potentially reading a stale
  nonce — see the `eth_getTransactionCount` "latest" vs "pending" discrepancy noted below)
  before the rejected broadcast that immediately preceded the stall. The trigger is most
  likely some multi-step interaction between repeated client-side timeouts/retries and mempool
  state that these single-shot, deliberately-constructed scenarios don't capture — it remains
  unconfirmed beyond the original single observation despite substantial, systematic
  reproduction effort.
- **`eth_getTransactionCount(addr, "latest")` can return a stale value.** Observed once on the
  same broken instance as the mempool stall above: `"latest"` returned a lower count than
  `"pending"` even though the discrepancy was checked before the stall was confirmed. Also not
  reproduced on demand in the follow-up session (a fresh devnet's `"latest"` and `"pending"`
  always agreed after every scenario tried above) — plausibly a symptom of the same underlying,
  not-yet-understood condition rather than an independent bug.
