# Architecture & Tooling Decisions

Log of deviations from the original spec, forced by dependency/version realities, with reasoning. Newest first.

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

