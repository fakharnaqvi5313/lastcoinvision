# Architecture & Tooling Decisions

Log of deviations from the original spec, forced by dependency/version realities, with reasoning. Newest first.

## 2026-08-03 — CosmWasm contracts need a `wasm-opt` post-processing step on this Rust toolchain

Completing Phase D's reference contracts (`contracts/cosmwasm/consent-registry`,
`contracts/cosmwasm/procurement-ledger`). Both compiled and unit-tested cleanly, but the
resulting `.wasm` bytecode was rejected by a real, live governance `MsgStoreCode` proposal on
a local devnet: `Error during static Wasm validation: ... "bulk memory support is not enabled"`.

Root cause: this toolchain's `wasm32-unknown-unknown` target ships prebuilt `core`/`std`
libraries that emit WASM bulk-memory instructions (`memory.copy`/`memory.fill`) by default —
wasmvm's static validator (bundled with wasmd) doesn't accept them. A per-crate
`-C target-feature=-bulk-memory` rustflag (`contracts/cosmwasm/.cargo/config.toml`) only
affects code the workspace's own crates emit, not the prebuilt standard-library code already
compiled into the shipped toolchain — so it alone wasn't enough; the same governance proposal
failed a second time with byte-for-byte identical output before this was understood.

**Decision:** `contracts/cosmwasm/build.sh` runs `wasm-opt --enable-bulk-memory
--llvm-memory-copy-fill-lowering` (binaryen) on the compiled output as a required last step,
lowering those instructions back to WASM-MVP-compatible equivalents. Confirmed working
end-to-end: the lowered bytecode passed a real `MsgStoreCode` governance proposal, both
contracts instantiated permissionlessly per the existing gov-gated-upload policy, and both
were exercised with real state-changing messages against a live local devnet. This is a
toolchain/build-environment issue, not a design choice specific to either contract — every
CosmWasm contract built on a similarly recent Rust toolchain needs this same step.

## 2026-08-03 — Solidity contracts (Foundry) and CosmWasm build tooling: gitignore additions

`contracts/solidity/lib/` (forge-std, fetched via `forge install`, ~1.3MB of vendored source)
and `contracts/solidity/cache/` follow the same convention as `node_modules/`/`target/`
elsewhere in this repo — installed on demand, not committed.
`contracts/cosmwasm/*/artifacts/` (the `wasm-opt`-processed build output, see above) is
likewise a build product, not source.

## 2026-08-02 — cosmos/evm integration: two fee markets, a module rename, and three genesis bugs found only by actually running a node

Completing Phase D's EVM half on top of the wasmd milestone (same day). Full design rationale
lives in [modules/evm/SPEC.md](modules/evm/SPEC.md); this entry covers what needed *fixing*,
found only by booting a real local node and sending a real signed transaction — none of these
surfaced from `go build`/`go vet`/unit tests, all of them from `lcvd start` plus `cast send`.

- **`x/feemarket` module-name collision.** cosmos/evm ships its own module literally named
  `"feemarket"` too (its EVM-side EIP-1559 market). Module names key the module manager,
  store, and genesis maps — two modules can't share one. Since cosmos/evm's name is
  upstream/fixed, this chain's own Phase A module was renamed `"lcvfeemarket"`
  (`x/feemarket/types/keys.go`); its proto package/REST namespace (`lcv.feemarket.v1`) was
  untouched, since that's independent of the Go-level `ModuleName` constant. Caught immediately
  by a Go compile error (duplicate map key in `maccPerms`) — the cheapest possible way to catch
  a collision like this.
- **Deliberately did not unify the two fee markets.** cosmos/evm's own feemarket must price EVM
  txs for MetaMask/JSON-RPC compatibility (the whole point of this phase); this chain's
  original feemarket must keep pricing everything else, unchanged since Phase A. `app/ante.go`
  was forked from cosmos/evm's own (unexported) branch-builder functions specifically so a
  tx's type — EVM vs. everything else — picks which fee market's decorator chain runs. See
  modules/evm/SPEC.md for the full decorator-chain diff.
- **wasmd's per-process wasmvm cache-directory lock (found in the wasmd milestone) has an EVM
  analogue that's much stricter: cosmos/evm's chain config is a process-global, set-once
  value** (`evmtypes.SetChainConfig`), not just a per-app-instance file lock. `cmd/lcvd/cmd/root.go`
  builds a *second* `app.New(...)` in the same process purely to read back codec/AutoCLI info
  before the real app is built for `start` — fine for wasmd (each app instance gets its own
  temp directory), fatal for cosmos/evm (`panic: chainConfig already set`) unless the first,
  throwaway call is made to resolve to `evmtypes.DefaultEVMChainID` specifically, which is the
  one value `SetChainConfig`'s guard allows a second, real call to override. Fixed by resolving
  the EVM chain id from `appOpts` inside `app.New(...)` (empty options → the sentinel; real
  `app.toml` config → the chain's actual id `2611`) instead of hardcoding the constant at every
  call site. See modules/evm/SPEC.md's "EVM chain ID" section.
- **The standard SDK's `SigVerificationDecorator` doesn't recognize eth_secp256k1 pubkeys.**
  Every gentx (and every tx generally) signed by an `eth_secp256k1`-algorithm account failed
  with `unrecognized public key type: *ethsecp256k1.PubKey` during `InitGenesis`, because the
  ante chain's `SigGasConsumer` defaulted to the SDK's own `DefaultSigVerificationGasConsumer`,
  whose pubkey-type switch has no case for it. Fixed by defaulting to cosmos/evm's own
  `SigVerificationGasConsumer` instead, which does.
- **The EVM feemarket's stock default base fee (1,000,000,000) is calibrated for an
  18-decimal-native EVM chain — 1,000 LCV per gas unit on this 6-decimal one.** Every
  `cast send` failed with `insufficient funds for gas * price + value` before this was caught,
  since no funded test account could afford even a plain 21,000-gas transfer at that price.
  Fixed by scaling the genesis default down by `10^(18-6)`, giving the same real-world gas
  price the upstream default intends. See modules/evm/SPEC.md's "Base fee" section for the
  live before/after numbers.

**`TestAppImportExport`/`TestAppSimulationAfterImport`/`TestAppStateDeterminism` (`app/sim_test.go`)
build multiple app instances within one test process**, which collides with cosmos/evm's
process-global chain config/coin-info the same way described above — fixed by resetting that
global state (`evmtypes.NewEVMConfigurator().ResetTestConfig()`) before every app instance after
the first, which only compiles under `go test -tags=test` (cosmos/evm ships two build-tagged
variants of this reset machinery; the production, non-`test` one panics if called at all — by
design, since a real running chain should never need to reset its own chain config). This has
no effect on production builds (`go build` never passes `-tags=test`) and no effect on the
default `go test ./...` run either, since these three tests are gated behind
`simcli.FlagEnabledValue` and self-skip unless invoked with `-Enabled=true` explicitly — exactly
the same opt-in simulation workflow already established in Phase A/C. Running them now requires
both flags together, e.g. `go test -tags=test -run TestAppSimulationAfterImport -Enabled=true
-NumBlocks=100 ./app`.

Running `TestAppImportExport` this way surfaced an unrelated, genuinely pre-existing bug in
Phase A's `x/emission` module: `ExportGenesis` never exports the keeper's `LastEmissionTime`
state (only `Params`), and `InitGenesis` always resets `LastEmissionTime` back to
`params.StartTime` — so a real chain export/import (upgrade, state-sync) would silently rewind
the decaying emission curve's bookkeeping to its original genesis start time, discarding
whatever time had actually elapsed. Fixed directly (small, well-contained, and the user asked for it explicitly): added
`last_emission_time` to `GenesisState` (`proto/lcv/emission/genesis.proto`), `ExportGenesis`
now populates it from `k.GetLastEmissionTime`, and `InitGenesis` uses it when non-zero
(falling back to `params.StartTime` only for a genuinely fresh genesis) — see
`x/emission/module/genesis.go` and its new round-trip tests.

**`TestAppImportExport` still fails after this fix**, on a different, unrelated mismatch —
root-caused (but not fixed; it lives entirely inside the vendored `cosmos/evm` dependency, not
this repo's own code) as follows: the diff is in the `evm` store (`x/vm`), on keys of the form
`KeyPrefixStorage (0x02) + <20-byte address> + <32-byte storage slot>`, with sequential slot
numbers (1, 2, 3, ... 15) under one address. The pattern is consistent: the key is *missing
entirely* from the original app's store but present with an *explicit zero value* after
export/import. Ruled out: `x/vm` and `x/erc20` both register zero `WeightedOperations` (no
`MsgEthereumTx`/ERC20-conversion messages are ever simulated), so this isn't real contract
storage from a simulated EVM call; `evmmonoante.NewGasWantedDecorator` (run for every tx,
Cosmos or EVM) only touches the feemarket keeper's gas bookkeeping, not `x/vm` storage. That
leaves cosmos/evm's own `x/vm.ExportGenesis`/`Keeper.GetAccountStorage`, which exports
*every* key found via a raw prefix-store iteration with no zero-value filtering
(`SetState` itself never calls `DeleteState` for a zero write either) — some other internal
code path (most likely an ante-level or precompile-adjacent StateDB touch, not yet pinned down
without instrumenting the vendored dependency itself) appears to mark a slot dirty long enough
for it to reach a zero-valued `SetState` call during import's `InitGenesis`, without the
original run ever having durably committed that same write. Not chased further: it's entirely
upstream (cosmos/evm) behavior, not something in this repo's own modules; the test is opt-in
(`-Enabled=true -tags=test`), not part of any acceptance criteria; and it doesn't reproduce in
the primary 100-block `BenchmarkSimulation` or in real devnet usage (JSON-RPC, MetaMask
transfers, and CosmWasm all verified working in Phase D). Worth reporting upstream to
cosmos/evm if genesis export/import fidelity becomes a hard requirement later.

**Net result, verified against a live local node (not just unit tests):** `eth_chainId`,
`net_version`, `eth_gasPrice` all return correct values; a real signed EVM transfer (`cast
send ... --value 10ether`, functionally identical to what MetaMask's own "Send" does — both
speak the same `eth_sendRawTransaction` JSON-RPC call) moved LCV between two accounts, with the
resulting `eth_getBalance` and `lcvd query bank balances` for both accounts agreeing exactly
(down to `ulcv`'s 6-decimal precision) — confirming the precisebank bridge is genuinely
bidirectional. Full build, `go vet`, and the full test suite are clean on top of this.

## 2026-08-02 — Phase D forced a cosmos-sdk v0.50.14 → v0.53.6 and ibc-go v8 → v10 upgrade, and a full app.go rewrite off depinject

Phase D asks for CosmWasm (wasmd) and EVM (cosmos/evm) support. Neither is compatible with
this chain's stack as it stood at the end of Phase C:

- **No compatible wasmd/cosmos/evm version exists for cosmos-sdk v0.50.14.** wasmd's own
  v0.5x series targets v0.50.x, but the newest wasmd release line that also lines up with an
  *official, non-forked* cosmos/evm release is wasmd v0.61.x + cosmos/evm v0.6.1, both of
  which require **cosmos-sdk v0.53.6**. The only wasmd/cosmos-evm combination compatible with
  v0.50.14 is cosmos/evm v0.1.0/v0.2.0, which pulls in cosmos-sdk via a `replace` directive
  pointing at an **unofficial forked commit** (`github.com/cosmos/cosmos-sdk
  v0.50.13-0.20250319183239-...`), not a tagged release — a materially worse choice for a
  chain marketed as government-grade, since it replaces cosmos-sdk *everywhere in the binary*,
  not just for the EVM module. Decision, confirmed with the user given the size of this
  tradeoff: upgrade to the official v0.53.6 release and pair it with official wasmd v0.61.14 /
  cosmos/evm v0.6.1, at the cost of a larger migration.
- **wasmd v0.61.x and cosmos/evm v0.6.1 both require ibc-go v10**, which dropped the
  `x/capability` module entirely (port/channel ownership no longer needs a separate capability
  system) and dropped the ICS-29 fee middleware (`29-fee`) outright — incentivized packet
  relaying is no longer available on this chain. ibc-go v8 and v10 are different Go modules
  (different major-version import paths), so nothing *forces* removing v8, but wasmd's and
  cosmos/evm's keepers are typed against v10's `ibckeeper.Keeper`, and a chain can only
  meaningfully run one IBC core (shared store keys, shared client/channel state) — so the
  whole app's IBC wiring (transfer, interchain-accounts) moved to v10 together with wasm/evm.
- **Neither wasmd nor cosmos/evm ship a depinject/appconfig module.** Both follow the older
  "manually construct every keeper in `app.go`" pattern (as does their own reference chains,
  `wasmd/app` and `cosmos/evm/evmd`) — there is no `appmodule.Provide`-based wiring available
  for either. Since a chain can only have one `app.go` construction style, `app/app_config.go`
  (the old `depinject.Configs`/`appconfig.Compose` setup) was deleted and `app/app.go` was
  rewritten to manually construct every keeper — all 6 custom modules included — directly,
  following the pattern in `wasmd/app/app.go` and `cosmos/evm/evmd/app.go`. `cmd/lcvd/cmd/root.go`
  was rewritten the same way: it builds a throwaway `app.New(...)` instance to read back the
  already-configured codec/tx-config/AutoCLI options, exactly as `cosmos/evm/evmd`'s and
  `wasmd`'s own root commands do.
- **`x/params` was dropped.** It existed solely so `app/ibc.go` could register legacy param
  subspaces for `ibcexported`/`ibctransfertypes`/ICA — none of which ibc-go v10's modules use
  anymore (they're on the current collections-based params pattern). Neither wasmd's nor
  cosmos/evm's own reference apps register `x/params` either.
- **A latent SDK test-helper bug became a real failure under v0.53.** `app/sim_test.go`'s
  `simtestutil.SimulationOperations(bApp, bApp.AppCodec(), config)` internally builds its own
  disposable tx config via `moduletestutil.MakeTestTxConfig()`, which defaults to a
  **hardcoded `"cosmos"` bech32 prefix** or `SigningOptions.AccAddressPrefix` unset — this has
  been true since at least v0.50.14, but only became a real "hrp does not match bech32 prefix:
  expected 'cosmos' got 'lcv'" failure once SDK v0.53 started resolving *every* message's
  signers exclusively through that address-codec path (v0.50.x's signer resolution had another
  path that didn't always need it). Fixed by switching every call site to
  `simtestutil.BuildSimulationOperations(bApp, bApp.AppCodec(), config, bApp.TxConfig())`,
  which takes the app's own (correctly "lcv"-prefixed) tx config instead of building a throwaway
  one. Confirmed clean on `TestAppSimulationAfterImport` (50 blocks) and a 100-block
  `BenchmarkSimulation`, both exercising wasmd's own `MsgStoreCode`/`MsgInstantiateContract`/
  `MsgExecuteContract` simulation operations successfully.
- **wasmd's keeper needs its own on-disk wasmvm cache directory per running app instance** —
  it takes an exclusive file lock on `<home>/wasm`. Two app instances in the same test process
  sharing one `FlagHome` (a pre-existing pattern in `sim_test.go`'s import/export and
  determinism tests, harmless before wasmd) now deadlock on that lock; fixed by giving each
  app instance in those tests its own unique temp directory, matching wasmd's own `sim_test.go`
  (which documents the exact same issue: https://github.com/CosmWasm/wasmd/issues/2017). The
  same fix was needed in `cmd/lcvd/cmd/root.go`'s throwaway app instance (used only to read
  back AutoCLI/codec info) — it now uses `os.MkdirTemp` + `simtestutil.NewAppOptionsWithFlagHome`
  instead of `simtestutil.EmptyAppOptions{}`, which resolves to no home directory at all and
  crashed wasmd's keeper trying to create `/wasm` at the filesystem root.

**Net effect:** `x/emission`, `x/feemarket`, `x/identity`, `x/registry`, `x/voting`,
`x/taxrail` — every custom module from Phases A–C — needed no logic changes at all; only their
manual keeper-construction call sites moved from `ModuleInputs`/depinject to plain arguments in
`app.go`. Full build, `go vet`, the full test suite, and a 100-block simulation are all
confirmed clean on the new stack before wiring in wasm's actual params (see below) or starting
the EVM half of Phase D.

## 2026-08-02 — CosmWasm code upload is gov-gated via a custom `DefaultGenesis` override

wasmd's stock `x/wasm` `AppModuleBasic.DefaultGenesis` defaults to fully permissionless code
upload (`AccessTypeEverybody`), which doesn't match the top-level spec's "permissioned upload
at genesis (gov-gated code upload), permissionless instantiate of approved code IDs." Since
`genutilcli.InitCmd` (used by `lcvd init`) calls `DefaultGenesis` on whatever
`module.BasicManager` it's given, and wasmd doesn't expose a way to pass different default
params into its `AppModuleBasic`, `app/app.go` defines `wasmAppModuleBasic` — a thin wrapper
embedding `wasm.AppModuleBasic` that overrides only `DefaultGenesis` to set
`CodeUploadAccess: AccessTypeAnyOfAddresses.With(govAddress)` and
`InstantiateDefaultPermission: AccessTypeEverybody`, then registers that wrapper in
`BasicModuleManager`'s override map (the same mechanism already used for `genutil`/`gov`).
Verified against a live local node: `lcvd init` produces the expected `genesis.json`, and
`lcvd tx wasm store` from a funded non-gov key fails with `can not create code: unauthorized`.
See [modules/wasm/SPEC.md](modules/wasm/SPEC.md).

## 2026-08-01 — Feemarket's mandatory base fee needed a simulation-only genesis override to unblock Phase C's 100-block acceptance run

Verifying Phase C's "simulations run clean for 100 blocks" acceptance criterion via
`go test -bench ^BenchmarkSimulation ./app` (the full randomized-tx block simulator) failed
immediately on ordinary `x/staking`/`x/feegrant` operations with `insufficient fee`, before
any Phase C module code ran at all — `x/registry`, `x/voting`, and `x/taxrail` register zero
`WeightedOperations` (matching every other custom module's existing pattern from Phases A/B),
so they contribute no simulated transactions and were not the cause.

The actual cause: `x/feemarket/ante.Decorator` (built in Phase A) enforces
`base_fee * gasWanted` on every delivered tx, not just at CheckTx/mempool time — that's the
point of an EIP-1559-style base fee that also gets burned. But the SDK's generic simulation
operations (`x/staking`, `x/bank`, `x/gov`, `x/feegrant`, etc. — none of which know about our
custom fee market) generate fees via `simulation.RandomFees`, which (a) has no way to target
an arbitrary external minimum, and (b) explicitly returns a **zero-value fee, not an error**,
for any freshly-created simulation account with no spendable balance yet — an assumption baked
into nearly every module's operations, which is fine for the standard SDK's un-gated
`DeliverTx`, but incompatible with our decorator's unconditional minimum.

**Decision:** `x/feemarket/module/simulation.go`'s `GenerateGenesisState` (used only by the
randomized block simulator, never by a real chain's genesis) now sets both `BaseFee` and
`MinBaseFee` to exactly zero, which `Params.Validate()` already allows (it only requires
`MinBaseFee` to be non-negative). This makes the ante decorator's required-fee check trivially
satisfied for every simulated tx regardless of the submitting account's balance, without
touching `types.DefaultParams()` — what a real chain's genesis actually uses — at all. This
mirrors how the SDK's own staking/slashing/gov modules already generate non-default,
simulation-only randomized params for this same harness. Confirmed clean with
`BenchmarkSimulation` (100 blocks) and `TestAppSimulationAfterImport` (50 blocks, full
import/export + invariant checks) afterward.

## 2026-08-01 — `x/registry`/`x/voting`/`x/taxrail` authority registries follow two different shapes depending on cardinality

`x/registry`'s jurisdiction authority (`MsgSetJurisdictionAuthority`) and `x/taxrail`'s tax
rule (`MsgRegisterTaxRule`) are each a single governance-set record keyed by jurisdiction code
— naturally 1:1, so each is a plain upsert rather than a list. `x/voting`'s electoral
authority (`MsgAddElectoralAuthority`/`MsgRemoveElectoralAuthority`) is a flat authorized set
instead, mirroring `x/identity`'s issuer-registry pattern, because electoral authorities
aren't jurisdiction-scoped the way land/tax records are. Documented per-module in each
`modules/<name>/SPEC.md` rather than unifying into one shared registry abstraction, since the
two shapes have genuinely different semantics (exclusive-per-key vs. open membership set).

## 2026-08-01 — `x/taxrail` is scoped to a single denom (`ulcv`)

`JurisdictionLedger` was initially keyed by `(jurisdiction_code, denom)`, but
`QueryJurisdictionBalanceRequest` only takes a `jurisdiction_code` — there was no way to
specify which denom's balance to fetch, making a multi-denom ledger unqueryable as specified.
Rather than extend the query surface for a use case outside the acceptance criteria, `x/taxrail`
now hard-scopes to `types.DefaultDenom = "ulcv"` and rejects any other denom outright in both
`MsgTaxedTransfer` and `MsgSpendCollectedTax` (`ErrUnsupportedDenom`) rather than silently
mishandling it. See `modules/taxrail/SPEC.md` for the full writeup.

## 2026-08-01 — `x/voting`'s commit-reveal scheme does not provide ballot secrecy against on-chain observers

Per the top-level spec's own requirement to document this honestly: `MsgRevealVote` is a
normal signed, broadcast transaction naming the voter's `choice` in the clear, so anyone
watching block history can trivially link the revealing address to its vote — the commitment
only hides the choice between commit and reveal, not after reveal, and this is not a
zero-knowledge or anonymous-voting scheme. See `modules/voting/SPEC.md`'s "Limitations"
section for the complete list of what this design does and does not guarantee (also covers
coercion/vote-selling resistance, which is likewise not addressed).

## 2026-08-01 — Docker deferred for local devnet bring-up

Docker Desktop is not installed on the dev machine (macOS 12.7.6, Intel). Docker Desktop
installation requires a privileged GUI installer that couldn't be run in this automated environment. The user is
installing it separately.

**Decision:** `make devnet` initially boots a 4-validator network as local OS processes
(separate `lcvd` home directories, ports offset per node) instead of docker-compose. The
docker-compose version specified in the Phase A acceptance criteria and built out fully in
Phase H is added as `make devnet-docker` once Docker is available, without changing the
process-based target. Both will be kept and tested going forward.

**Update 2026-08-01:** Docker Desktop is now installed and verified working (`docker run
hello-world` succeeds) after a clean reinstall. The process-based `make devnet` remains the
primary Phase A target since it's already built and tested; `make devnet-docker` can be
added opportunistically but isn't required to unblock Phase A anymore.

## 2026-08-01 — Pin Go toolchain to 1.25.0, not the system's 1.26.5

The dev machine has Go 1.26.5 installed. `github.com/bytedance/sonic` v1.12.3 (pulled in
transitively by the Cosmos SDK's gin-based HTTP stack) uses a `//go:linkname` hack into
`encoding/json`'s unexported `unquoteBytes`, whose signature/existence changed in the Go
1.26 standard library, so `cmd/lcvd` fails to link with "invalid reference to
encoding/json.unquoteBytes" against 1.26.5. Sonic hasn't published a fix compatible with
1.26.x as of this writing.

**Decision:** run `go env -w GOTOOLCHAIN=go1.25.0+auto` machine-wide, matching `chain/go.mod`'s
`go 1.25.0` directive. Go transparently downloads and uses the 1.25.0 toolchain for this
(and any other) module instead of the system-default 1.26.5, and still auto-upgrades further
for any module that explicitly requires newer. An explicit `toolchain go1.25.0` line in
`go.mod` was tried first but `go mod tidy` strips it as redundant with the `go` line, so the
env-level pin is the durable fix.

## 2026-08-01 — Custom module code lives at `chain/x/<name>`, not top-level `modules/<name>/`

The spec's monorepo layout lists `modules/` as a top-level directory sibling to `chain/`.
But Go module boundaries are directory-subtree based: a `go.mod` at `chain/` cannot include
packages from a sibling `../modules/` directory without either (a) making `modules/` its own
Go module wired to `chain/` via a `go.work` workspace and `replace` directives, or (b) just
keeping module code inside `chain/`'s own tree. Ignite's tooling (`ignite scaffold module`,
`app_config.go` wiring, proto codegen paths) all assume option (b) and fight you at every step
under (a), for no real benefit at this stage (nothing needs `modules/` independently
versioned or published yet).

**Decision:** custom module Go code (`emission`, `feemarket`, `identity`, `registry`,
`voting`, `taxrail`) lives at `chain/x/<name>/`, the standard Cosmos SDK location. The
top-level `modules/<name>/SPEC.md` files required by the global rules still exist as
specified — they just document code that physically lives under `chain/x/<name>/` — so
`modules/` becomes a specs-and-docs directory rather than a code directory. Revisit if a
module ever needs to be an independently published Go package.

## 2026-08-01 — Devnet validator self-bond funding sits outside the 10B allocation table

`scripts/genesis/init-genesis.sh` uses `lcvd multi-node` to generate the N validator
homes/keys/gentxs, and that command has a hardcoded (not flag-configurable) step that funds
each validator's own account with 500 LCV so it has something to self-delegate and pay gas
with (`chain/cmd/lcvd/cmd/testnet_multi_node.go`). `gengenesis` then adds the real
25/20/15/15/12/8/5% allocation table on top via separate accounts. The result: a 4-validator
devnet's *actual* total supply is 10,000,002,000 LCV, not exactly 10,000,000,000 — a
0.00002% overage that only exists because this is bootstrap tooling for a local test
network, not mainnet genesis construction.

**Decision:** accept this for `lcv-devnet`/`lcv-testnet-1` purposes; it is not a violation of
the "10B fixed supply" design intent, which is a mainnet property. A real mainnet launch's
genesis (see the Phase I `RELEASE_CHECKLIST.md`, once written) must have validators self-fund
their bonds from the public/ecosystem allocation through a real distribution process, not
`multi-node`'s devnet auto-funding — do not reuse this script unmodified for a production
genesis.

## 2026-08-01 — Fixed `lcvd multi-node`'s hardcoded 100% validator commission

Ignite's scaffolded `chain/cmd/lcvd/cmd/testnet_multi_node.go` hardcodes every generated
validator's commission (rate, max_rate, max_change_rate) to `math.LegacyOneDec()` — 100%.
That's not a devnet-only quirk that can be patched out of the genesis file afterward:
commission is a field of the signed `MsgCreateValidator` embedded in each gentx, so changing
it post-hoc invalidates the transaction's signature. Discovered while verifying Phase A's
"can delegate, undelegate, and see rewards accrue" acceptance criterion — with 100%
commission, every delegator's *personal* reward share is exactly zero (all rewards go to
the validator's commission pool), which made it initially look like x/emission wasn't
working at all, when the actual emission → fee_collector → distribution pipeline was fine
the whole time (confirmed via `query distribution validator-outstanding-rewards`).

**Decision:** since this file is scaffolded into our own repo (not a vendored dependency),
edit the hardcoded rate directly to 10%/20%/1% (rate/max_rate/max_change_rate) — a plausible
mainnet-ish default — rather than working around it downstream.

## 2026-08-03 — OPEN BUG: `--gas auto` / CosmJS `"auto"` fee underestimates on every tx, chain-wide

Discovered while building Phase E's SDK examples (a plain `MsgCreateDID` broadcast kept
failing with `out of gas`). Root-caused down to: this is **not** specific to `x/identity` or
to the SDK — a bare `lcvd tx bank send ... --gas auto --gas-adjustment 1.4` against a fresh
single-node devnet also fails with code 11 (`out of gas in location: ReadFlat`/`WriteFlat`),
`gas_used` exceeding `gas_wanted` despite the 1.4x adjustment already applied by the CLI
itself. Raising `--gas-adjustment` does not converge the way it should: at 1.4x, real usage
was ~41% over the raw (pre-adjustment) simulated estimate; at 3.0x, real usage was ~66% over
the same raw estimate for the same tx. Real gas usage appears to scale with the tx's own
requested gas limit rather than being fixed by the work the tx actually does — i.e. the CLI's
`simulate` step (which runs before a real gas limit exists) is measuring something structurally
different from what deliverTx charges, in a way that grows with whatever `--gas`/multiplier is
picked. Not reproduced as a total failure — a large enough fixed multiplier (~3x) or a
generously fixed manual `gas` value both work — but the *default* `--gas auto` (1.4x) flow that
every `lcvd tx` invocation and every CosmJS `"auto"` fee relies on by default is unreliable
for essentially any transaction on this chain, standard `x/bank` sends included. Not yet
root-caused to a specific ante decorator; `ReadFlat`/`WriteFlat` are generic Cosmos SDK
KVStore gas-config labels (from `store/types/gas.go`), not proof of which handler is
responsible — `x/feemarket`'s partial-burn ante decorator (Phase A) is the prime suspect
given the timing, but this needs an actual profiling pass, not just this observation, to
confirm.

**Decision:** do not attempt to root-cause or fix this from within Phase E — it is a Phase A/
`x/feemarket` ante-handler concern and out of scope for the SDK/CLI/faucet work. Phase E's
SDK examples, `create-lcv-app` templates, and faucet all use a fixed `StdFee` (manual `gas`
+ `amount`) instead of `"auto"`/`fee: "auto"` for any tx that goes through this chain's
ante handlers, with a comment pointing back to this entry. This is a workaround, not a fix:
real dApp developers using this SDK's default `"auto"` fee path will hit the same failure,
so this should be treated as a release blocker for testnet, not a cosmetic issue — flagged
here for prioritization rather than silently patched over.

## 2026-08-03 — Enabled `enabled-unsafe-cors` in the devnet's `app.toml`

Discovered while verifying Phase E's `gov-registry` and `voting-portal` `create-lcv-app`
templates against a live devnet: browser `fetch()` calls to the REST/gRPC-gateway API failed
with a generic "Failed to fetch" (no `Access-Control-Allow-Origin` header at all in the
response), even though the exact same URL worked fine from `curl`. `scripts/genesis/
init-genesis.sh` never touched CORS, so every devnet home was left at the Cosmos SDK's
default `enabled-unsafe-cors = false` in `app.toml` — safe for a real network, but it also
silently blocks every browser-based dApp (any of Phase E's Next.js templates, or any other
CosmJS/fetch-based frontend) from calling the REST API directly, which is the normal,
expected way these dApps talk to a chain.

**Decision:** `scripts/genesis/init-genesis.sh` now flips `enabled-unsafe-cors` to `true` for
every validator home it generates (chain/app/config.go's own account/API code is untouched —
this is purely an `app.toml` value). This only affects `make devnet`-generated local
networks; it is not something to carry into a real testnet or mainnet `app.toml`, where CORS
should be scoped to specific known origins (or left disabled) rather than wide open. Anyone
who already has a devnet home generated before this fix should manually flip the same line in
their existing `config/app.toml` and restart, since the script only patches new homes.

## 2026-08-03 — `web/` ships one dynamic API route, not a pure static export

The top-level spec calls the website "static-exportable," but the `/governments` page's
pilot-program contact form needs a real server-side handler to validate and log submissions
(`app/api/contact/route.ts`) — Next.js's `output: "export"` mode disallows POST route handlers
entirely, since static export has no server to run them on.

**Decision:** `next.config.mjs` does not set `output: "export"`. Every other page is fully
static (confirmed via `next build`'s per-route output — everything except `/api/contact` is
prerendered), so the site is static-exportable in the sense that matters: it deploys cleanly
to any Next.js-capable host (Vercel, a Node server, etc.) with only the one contact-form route
needing actual server execution. A deployment that truly requires 100% static hosting (e.g.
plain S3/CDN with no compute) would need to swap that one route for an external form
endpoint — not attempted here, since no such constraint was given.

## 2026-08-03 — Navy hex: spec's `#1F4E79` vs. the master logo's `#0E2640`

The top-level spec's brand direction gives an explicit hex for "deep navy" (`#1F4E79`), but
the master logo committed earlier (`brand/lcv-logo.png`, see the 2026-08-01 entry) uses a
darker navy (`#0E2640`) for its disc background, and its own README explicitly deferred
reconciling this to Phase F.

**Decision:** both hexes are kept, as adjacent steps in one navy scale (`#1F4E79` as
`navy-500`, `#0E2640` as `navy-900`) rather than picking one over the other — `navy-500` for
primary UI elements (buttons, links), `navy-900` for dark backgrounds and the logo itself.
This is stated explicitly on the `/brand` page rather than silently choosing one hex and
hoping no one notices the logo doesn't match it.

## 2026-08-03 — Roadmap uses phase status, not calendar dates

The spec asks for a "roadmap timeline." Real ship/target dates for phases G–I aren't something
this session can commit to on the project's behalf.

**Decision:** `/roadmap` and the homepage's roadmap section track the project's actual build
phases (A–I) with a status of shipped / in-progress / planned, derived from what's actually
been built and verified in this repo — no calendar dates. Revisit if real target dates are
ever decided outside of this session.

## 2026-08-03 — Verified `web/` with a real Lighthouse run, not an assumed score

Ran `lighthouse` (real Chrome, headless) against every page of the built site locally rather
than assuming it would pass the spec's "Lighthouse ≥90 on all categories" bar. First pass
found two genuine issues: several muted-text color combinations (`.eyebrow`, footer captions,
a few small labels) had light-mode-only or backwards light/dark contrast pairings that failed
WCAG's 4.5:1 for normal text (as low as 2.62:1), and the roadmap timeline's item headings used
`<h3>` where the surrounding page had no `<h2>`, tripping axe's heading-order check on
`/roadmap`. Both were fixed (contrast: darkened the affected light-mode text colors;
heading-order: item titles are now `<h2>`) and re-verified — every page now scores 100 on
accessibility and SEO, and ≥94 on performance and best-practices.

## 2026-08-03 — FIXED: `make devnet` was fundamentally broken (four bugs)

Discovered while writing Phase G's "Run a node" tutorial and needing it to actually work
against the project's own primary documented entry point. `make devnet` — the 4-validator
local devnet every phase's acceptance criteria and the whole `README.md`/Makefile point to —
has not actually come up successfully since some point in Phase D, across three independent
bugs stacked on top of each other:

1. **`scripts/genesis/init-genesis.sh` called `lcvd tendermint show-node-id`.** The Phase D
   cosmos-sdk v0.50.14 → v0.53.6 upgrade (see the 2026-08-02 entry) renamed that subcommand to
   `comet`. The call was piped through `2>/dev/null`, so it failed silently — `make devnet`
   exited nonzero with no visible error at all, not even a stack trace, just "Error 1" from
   Make itself.
2. **`lcvd start` rejected every real chain-id.** `chain/cmd/lcvd/cmd/root.go` forced
   `--chain-id`'s default to `"lcv"` (`app.Name`) across the *entire* command tree via
   `overwriteFlagDefaults`, intending it as a convenience default for `tx`/`query`/`keys`. But
   `--chain-id` is a single flag object shared by inheritance across the whole tree — there's
   no way to default it only for client commands — and `start` needs it **empty** so
   cosmos-sdk's server bootstrap falls back to reading the real chain-id from `genesis.json`.
   With `"lcv"` forced, every `lcvd start` failed with `invalid chain-id on InitChain;
   expected: lcv, got: lcv-devnet`, regardless of genesis being perfectly valid. This also
   meant the "convenience default" was subtly wrong for `tx`/`query` too — it could never be
   correct for any chain-id actually used anywhere in this project (`lcv-devnet`,
   `lcv-testnet-1`, `lastcoinvision-1`).
3. **`lcvd multi-node`-generated homes had no working EVM RPC or REST API.** `testnet_multi_node.go`
   built its `app.toml` from bare `serverconfig.DefaultConfig()` and wrote it via the
   package's default (non-EVM) template — entirely bypassing `initAppConfig()`'s EVM-aware
   config that a plain `lcvd init` gets. The generated `app.toml` had no `[json-rpc]` section
   at all, and explicitly set `API.Enable = false`. Every 4-validator devnet came up with no
   REST API and no `:8545`, even though `scripts/devnet/start.sh`'s own banner claimed
   `API http://localhost:1317` was live. Separately, cosmos/evm's own
   `DefaultJSONRPCConfig()` ships `Enable: false` despite its doc comment claiming the
   opposite — so even a plain `lcvd init` single node needed this fixed too.
4. **5-second block times made this devnet's fast-gov config (30s voting period, from
   `gengenesis --devnet-fast-gov`) unreliable in practice.** CometBFT's default
   `timeout_commit` (5s — meant to give real, geographically distributed validators time to
   gossip precommits) applies just as much to 4 local processes on localhost. A realistic
   governance workflow (submit proposal, poll for inclusion, vote from 4 keys) easily used up
   more than 30 seconds of wall-clock time at ~5s/block, causing votes to land after
   `voting_end_time` and get silently discarded (`code: 0` from `lcvd tx gov vote`'s sync
   broadcast reflects `CheckTx` passing, not `DeliverTx` succeeding — the actual failure only
   shows up via `lcvd query tx <hash>`, as `1: inactive proposal`).

**Decision:** fixed all four. (1) `comet show-node-id`, with the `2>/dev/null` also removed
so a future regression fails loudly. (2) Removed the chain-id override entirely rather than
trying to special-case it per-command (found and reverted one attempt at that — the flag is
inherited by reference, not copied, so a subcommand-scoped reset was silently a no-op).
`lcvd tx`/`query`/`keys` now require `--chain-id` explicitly, same as most cosmos-sdk chains'
documented CLI usage — this is correct, not a regression. (3) `testnet_multi_node.go` now
builds the same `evmconfig.EVMAppConfig` `initAppConfig()` does, calls
`srvconfig.SetConfigTemplate` with the matching EVM template before writing, sets
`API.Enable = true`, and assigns each validator a unique JSON-RPC port
(`scripts/genesis/init-genesis.sh`'s per-validator port-patching loop now also shifts
`[json-rpc]`'s `address`/`ws-address`, `scripts/devnet/start.sh`'s banner now prints the real
EVM port per validator); `initAppConfig()` itself now also forces `JSONRPCConfig.Enable =
true` for the single-node `lcvd init` path. (4) `scripts/genesis/init-genesis.sh`'s
per-validator `config.toml` patching now also sets `timeout_commit = "1s"`, cutting block
time roughly 5x for local iteration — not something to carry into a real testnet/mainnet
`config.toml`, where the default exists for a real reason.

Verified end-to-end after all four fixes: fresh `make devnet` produces 4 validators that
reach consensus in ~1-2s blocks, and every Phase G tutorial's commands were run against it
for real — a full `x/identity` issuer-authorization → credential-issuance → tier-check
cycle, a full `x/voting` electoral-authority-authorization → election → register → commit →
reveal → tally cycle (including discovering the exact `sha256(choice + "|" + salt)`
commitment format the chain expects, undocumented anywhere until now), and a full
`LCVEscrow.sol` deploy → fund → create → deposit → release cycle via Foundry against the EVM
JSON-RPC endpoint.

This should be treated as a significant finding, not a cosmetic fix: every prior phase's
"verified against a live devnet" claims that used `make devnet` specifically (rather than a
manually `lcvd init`'d single node with `--chain-id` passed explicitly, which sidesteps bug
#2) would have hit at least one of these. Spot-checking suggests most Phase D/E verification
used manual single-node setups instead, which is presumably why this went uncaught for so
long.

## 2026-08-03 — Docs site's module reference tolerates a few unresolvable links

`docs/`'s Docusaurus config renders each `modules/*/SPEC.md` directly at `/modules/<name>`
via a second `docs` plugin instance pointed at `../modules` (see the module-reference task in
this phase) — the whole point being zero duplication, so the reference and the source can
never drift apart. Those SPEC.md files also use GitHub-relative links like
`../../DECISIONS.md` and `../../chain/x/identity`, which resolve correctly when viewed on
GitHub (their primary context) but point outside `docs/`'s content root and can't resolve to
any Docusaurus route.

**Decision:** set `onBrokenLinks: "warn"` (Docusaurus's own default is `"throw"`, which would
fail the build) rather than rewriting those links or duplicating the SPEC.md content with
docs-site-specific links. The build output calls out exactly which links are unresolvable on
every affected page — accepted as a known, cosmetic gap rather than something worth breaking
the zero-duplication design for. Also had to fix one real MDX-compatibility bug found along
the way: a multi-line inline code span containing `<`/`<=` in `modules/voting/SPEC.md`
(`registration_start < registration_end <= ...`, split across a line wrap) confused MDX's
JSX-aware tokenizer into a hard parse failure — reflowed so the span stays on one line, no
content change.

## 2026-08-03 — FIXED: root `make lint`/`make test` were silently broken (CI would have been green on red code)

While wiring up `.github/workflows/ci.yml` for Phase H, running the exact `make lint` /
`make test` / `make build` commands the workflow calls (to confirm CI would actually pass)
surfaced two real, previously-undiscovered bugs in the root `Makefile`:

1. **`make test` failed outright, every time.** `go vet ./...` against the whole `chain`
   module trips "unreachable code" on generated `protoc-gen-go-pulsar` output under
   `chain/api/**` (confirmed these are 100% `// Code generated ... DO NOT EDIT` files, and
   `go vet` on every non-generated package is clean). `make test` ran vet as an unconditional
   prerequisite line, so it never even reached `go test`. Fixed by scoping vet to
   `$(go list ./... | grep -v '/api/')` in both `test` and `lint`.
2. **`make lint`'s fallback path (no `golangci-lint` installed, which is this environment's
   actual state) could never fail**, regardless of real vet or gofmt problems. The recipe's
   `cd chain && go vet ./...;` used a bare `;` instead of `&&`/`|| exit 1`, discarding vet's
   exit code. Worse, the *next* line's `gofmt -l .` ran inside `$(cd chain && ...)` — but the
   shell's cwd was already `chain/` from the previous line (recipe lines joined by `\` share
   one shell), so it silently tried `cd chain/chain`, failed, and produced an empty (i.e.
   "everything's formatted") result no matter what. Fixed by `cd`-ing into `chain/` exactly
   once at the top of the recipe, and propagating vet's exit code with `|| exit 1`.

Running the corrected `make lint` for the first time immediately caught one real
`gofmt`-dirty file, `chain/cmd/lcvd/cmd/testnet_multi_node.go` (import-block ordering, from
the Phase G devnet fix) — reformatted with `gofmt -w`, no logic change.

This should be treated the same way as the Phase G `make devnet` findings: a genuinely broken
piece of the documented workflow that would have made Phase H's "CI green" acceptance
criterion meaningless (CI would have reported success on the `test` job never running any
tests, and the `lint` job never being able to fail). Caught only because CI's exact commands
were run locally before trusting the workflow, not because of any change in scope.

## 2026-08-03 — Software-upgrade handler discovery: a binary can't schedule an upgrade it already contains

Building `scripts/upgrade/`'s governance walkthrough for the sample "v2" upgrade
(`chain/app/upgrades/v2`, wired via `chain/app/upgrades.go`'s `setupUpgradeHandlers()`)
surfaced a real cosmos-sdk behavior that isn't obvious from the docs: **a binary that already
has a named upgrade's handler registered cannot schedule that same upgrade via governance —
it panics immediately** (`BINARY UPDATED BEFORE TRIGGER! UPGRADE "v2" - in binary but not
executed on chain. Downgrade your binary`), not just at the target height. `x/upgrade`'s ABCI
handling checks this on every block once a plan is scheduled, and refuses to let history show
a plan as "passed but not yet executed" if the running binary is already capable of executing
it — a real safeguard against operators running tomorrow's binary today.

This meant the walkthrough's first design (submit the "v2" upgrade against a normal `make
devnet`, using the same always-registered binary throughout) doesn't work: `chain/build/lcvd`
has the `v2` handler registered unconditionally, so it hit the panic the instant the proposal
passed. Confirmed by testing on a fresh devnet with the current binary — same panic.

**Fix:** `scripts/upgrade/build-pre-upgrade-binary.sh` builds a temporary "pre-v2" binary
(the one `setupUpgradeHandlers()` line commented out via `sed`, restored immediately after
via a trap) purely for the walkthrough to run its "before" validators on —
`submit-upgrade-proposal.sh` starts a fresh devnet on that binary, submits+votes+passes the
proposal, confirms the pre-v2 binary halts at the target height with the expected `UPGRADE
"v2" NEEDED` error (the *correct*, designed outcome — this is exactly the moment a real
operator's cosmovisor would swap in the new binary), then swaps in the real
`chain/build/lcvd` and confirms the chain resumes and keeps producing blocks past that
height. Verified end-to-end, including watching the expected panic in the pre-v2 binary's log
before the swap.

This should be treated as a durable lesson for any *future* named upgrade added to this repo,
not just a one-off script quirk: an upgrade's handler must never ship in a binary that's
already running before governance schedules that exact upgrade, or the chain halts
immediately instead of at the intended height.

## 2026-08-03 — `deploy/docker-compose/`'s `lcvd` image could not be built end-to-end in this environment (disk)

This development sandbox's host filesystem has very little free space (repeatedly measured
between 0.5–5 GB free, fluctuating, on what's nominally a 234 GB volume — most of it is
consumed by other things on the host outside this repo). `chain/Dockerfile`'s builder stage
(`golang:1.25-bookworm`, then `go mod download` + compiling the full cosmos-sdk/wasmvm/evm
dependency tree with CGO) needs several GB of headroom to complete; three separate attempts
here failed with `no space left on device` mid-build, twice during the base image's own layer
extraction (before compilation even started). Freeing space between attempts (clearing
`go-build`/Homebrew/Yarn/pip caches, pruning Docker's build cache) helped but wasn't durably
enough — this appears to be a hard constraint of this particular sandbox, not something more
cache-clearing fixes.

**What was and wasn't verified as a result:**

- `docker-compose.yml`: valid — `docker compose config` resolves the full 7-service stack
  (including `.env`-sourced `FAUCET_MNEMONIC`) with no errors.
- `deploy/docker-compose/generate.sh`: verified for real — generates a working 4-validator
  genesis and rewrites each validator's `config.toml`/`app.toml` for container-to-container
  networking. Caught and fixed one real bug this way (see below).
- `explorer/Dockerfile`: **built successfully and run** (`docker build` + `docker run`),
  confirmed serving ping.pub correctly including the SPA-routing nginx fix and the bundled
  `lcv-devnet` chain config.
- `faucet/Dockerfile` and `web/Dockerfile`: standard multi-stage Node patterns (already
  proven — `npm run build` for both packages was verified working locally in this same
  environment, including `web`'s new `output: "standalone"` config added specifically for
  this Dockerfile). Not run inside Docker itself due to the same disk constraint, but nothing
  about them is unusual or untested at the build-tool level.
- `chain/Dockerfile`: **not verified end-to-end** — correct, standard multi-stage Go+CGO
  design (matches how any wasmvm-linked cosmos chain must be built: native compile per
  target, no cross-compilation), but this environment couldn't complete the build to confirm
  it. Recommend running `docker compose build lcvd0` (or `docker compose up --build`) on a
  machine with at least ~10 GB free disk to complete this specific verification — nothing in
  the design depends on this environment's constraint, it's purely a resource limit here.

**Real bug found and fixed via `generate.sh` regardless:** the P2P (`26656`) and RPC
(`26657`) `laddr` rewrite sed commands ran in an order where the RPC line's fresh
`tcp://0.0.0.0:...` value collided with the P2P rewrite's own match pattern, silently
clobbering both onto the same port. Fixed by reordering the two `sed` calls (P2P first, while
RPC's line still starts with `127.0.0.1:` and can't be matched by the P2P pattern) — see the
comment left in `generate.sh` at the fix site.

