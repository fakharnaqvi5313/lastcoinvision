# Architecture & Tooling Decisions

Log of deviations from the original spec, forced by dependency/version realities, with reasoning. Newest first.

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
whatever time had actually elapsed. Out of scope for Phase D (no EVM/wasm code involved) —
flagged as a follow-up task rather than fixed here, to avoid scope-creeping unrelated Phase A
changes into this phase.

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

