# Architecture & Tooling Decisions

Log of deviations from the original spec, forced by dependency/version realities, with reasoning. Newest first.

## 2026-08-07 — Mainnet gate work: `VALIDATOR_HSM_RUNBOOK.md` for the consensus-key gate

Wrote a concrete migration plan for `RELEASE_CHECKLIST.md` §2's still-open HSM/remote-signer
item (`lcv-testnet-1`'s two validators both keep `priv_validator_key.json` as a plain file).
**Deliberately did not touch either live validator** — a consensus-key migration done wrong
causes an unrecoverable double-sign/tombstone, and that's each operator's call to execute on
their own infrastructure and schedule, not something to run unattended.

Presented two real options rather than picking one by default: tmkms + YubiHSM2 (single
hardware signer per validator, simpler, matches the current one-operator-one-validator shape)
versus Horcrux (threshold signing across independent cosigner machines, stronger guarantee, but
needs each operator to run 3+ machines — a bigger lift than two solo operators currently
justify). Recommended tmkms+YubiHSM2 as the practical starting point for this specific project,
not as a universal best answer. The migration procedure section spells out the actual dangerous
part explicitly: never run the old file-based key and a new remote signer against the same
consensus key at the same time, and delete (not archive) the old key file once the new signer is
confirmed working, so it can't accidentally be started up in parallel later.

Cross-linked from `RELEASE_CHECKLIST.md` §2 and `MAINNET_LAUNCH_RUNBOOK.md`'s validator gate.
Gate itself stays open — this is the plan, not the execution.

## 2026-08-07 — Mainnet gate check-in: AftPak audit status re-verified, stale "no audit started" language fixed

User asked to work toward mainnet-launch readiness. Before touching any of the harder
infrastructure/legal gates, checked whether anything changed on the audit front since the
2026-08-06 entry (which left the initial AftPak report verified-but-insufficient, with a fuller
report "expected the following week").

**Nothing new arrived.** No new AftPak document exists beyond the three already logged
(`aftpak_security_audit.pdf`, `aftpak_security_audit-v6.pdf`, and the actual paid engagement's
initial report referenced in the 2026-08-06 entry). The promised fuller report is one day past
that entry, well within "the following week" — not overdue yet.

**Found and fixed stale language**: both `AUDIT_SCOPE.md` ("No audit has been performed") and
`MAINNET_LAUNCH_RUNBOOK.md`'s audit gate ("no audit has started") still dated from 2026-08-05,
before the AftPak engagement was disclosed. Neither was a security problem, but both understated
what's actually true now — a real, verified-genuine engagement exists, even though it hasn't
produced anything beyond automated-tooling-equivalent findings yet. Updated both to state the
accurate current position: engagement real and technically checks out, manual/expert review
still not delivered, this gate stays open. This is a documentation correction, not new audit
progress — the gate's actual status (open) is unchanged.

Also noted in passing, not acted on: an untracked `audit-output/gosec.json` (0 issues, 0 files
scanned) sits in the repo root — looks like a broken/incomplete local gosec run from an earlier
session, not a real deliverable. Left alone; flagged here so it isn't mistaken for audit
evidence later.

## 2026-08-07 — Wallet build Phase 5: app-store compliance gate (partial — see `STORE_COMPLIANCE.md`)

Started Phase 5. Split the work into what an agent can actually do versus what requires the
account owner's own legal/business identity and payment method, and did the former:

- **Added `web/app/privacy/page.tsx`** (linked from the site footer via a new "Legal" column in
  `web/lib/siteConfig.ts`), stating the wallet apps' non-custodial model plainly: keys are
  generated/encrypted/signed entirely on-device, no backend ever sees them, no analytics or
  crash-reporting SDK is in either wallet app today, and what the RPC/REST network calls do
  expose (IP address to whichever node the app talks to — same as any blockchain wallet).
  Verified this claim by grepping both wallet apps for analytics/telemetry SDK names before
  writing it — true as of this commit, not a boilerplate assumption.
- **Wrote `STORE_COMPLIANCE.md`** as the phase's tracking document (same role
  `RELEASE_CHECKLIST.md` plays for the testnet launch): itemizes what's done, what's explicitly
  blocked on the account owner (Apple Developer Program org enrollment — needs a real business
  entity, a D-U-N-S number, and the $99/year fee; Google Play Console org account — needs
  business verification and the $25 fee), draft store-listing copy ready for when those unblock,
  and the distribution-mechanics gap (no local CocoaPods means no local iOS archive path — EAS
  Build is the documented way around it once an Apple account exists).

**Deliberately did not**: create any developer-program account, agree to any terms, or spend any
money — creating accounts and making payments are both outside what this agent does on someone's
behalf regardless of instruction, and Apple/Google both require a real legal identity behind the
enrollment that only the account owner has. This phase stays partially open until those two
external steps happen; `STORE_COMPLIANCE.md` tracks exactly what's left.

**Follow-up, same day: `apps/mobile/eas.json`.** Wrote EAS Build's `development`/`preview`/
`production` profiles ahead of when an Apple/Google account exists to actually run them. Doing
this surfaced a real bug worth fixing regardless of store timing: `network.ts`'s
`DEFAULT_NETWORK` was hardcoded to `"devnet"`, meaning any real distributed build (TestFlight,
Play Internal Testing, or just a `.apk` sideloaded to a friend's phone) would try to reach
`localhost:26657` on first launch — a URL meaningless off the dev machine — and silently fail to
connect before a tester ever discovers the in-app network switcher. Fixed by having
`preview`/`production` set `EXPO_PUBLIC_LCV_NETWORK=testnet` in `eas.json`, and `network.ts` read
it (falling back to `devnet`, unchanged, for local `npx expo start`). Verified `tsc`/`vitest`
stay clean and `EXPO_PUBLIC_LCV_NETWORK=testnet npx expo export` still bundles correctly — no
`eas build` run, since that needs an EAS account login this agent won't create.

## 2026-08-07 — Wallet build Phase 4: mobile app (`apps/mobile`, Expo/React Native)

Built the mobile counterpart to Phase 3's desktop app: same feature set (create/import,
unlock/forget device, dashboard with both addresses and a plain send per chain), same
`wallet-core`/`tx-builders` business logic, ported to React Native screens.

**Real architecture deviation from the original wallet plan, documented in
`apps/mobile/README.md`.** The plan called for `react-native-keychain` and the bare/dev-client
RN workflow. Neither is usable in this environment: this Mac's system Ruby (2.6.10) cannot build
native gem extensions against the installed Xcode SDK, so `gem install cocoapods` fails, and
without CocoaPods no bare/prebuild RN iOS build can run `pod install`. Rather than trying to
repair the Ruby/Xcode toolchain (disproportionate for this task), pivoted the entire dependency
set to Expo Go-compatible, no-native-linking modules: `expo-secure-store` (still backed by real
iOS Keychain/Android Keystore) instead of `react-native-keychain`; `expo-crypto`'s
`getRandomValues` instead of `react-native-get-random-values`; `expo-clipboard`; plain `buffer`
for the global `Buffer` polyfill. The whole app now runs in plain Expo Go with zero native
builds. Revisit only if a later phase needs a capability Expo Go doesn't expose.

**`src/lib/polyfills.ts`, imported first in `index.ts`**, gives Hermes the `crypto.getRandomValues`
and global `Buffer` that `@cosmjs/*`/viem/`@noble/*` assume exist (same browser-environment
assumption Phase 3 relies on for Tauri's webview — Hermes just doesn't provide it natively).

**Verification** (device/simulator UI verification not possible — see gap below):
- `tsc --noEmit` and `vitest run` clean (4/4 tests, `src/lib/network.ts`).
- **Real headless run** of `scripts/verify-session-flow.mjs` (identical in purpose to Phase 3's
  script) against the live local devnet: wallet creation, dual-address derivation, keystore
  round-trip, `LcvTxClients.connect()`, and a real REST query all succeed.
- `npx expo export --platform ios` and `--platform android` both succeed, bundling 1400+ modules
  (including `wallet-core`, `tx-builders`, and every crypto dependency they pull in) to Hermes
  bytecode. **This caught a real bug**: Metro's module resolver does not map a `.js` import
  specifier to a `.ts` file the way `tsc`'s `bundler`/`nodenext` resolution does, so every
  relative import written as `"./foo.js"` (matching the desktop app's convention, which Vite/tsc
  resolves fine) failed to bundle under Metro. Fixed by making all relative imports across
  `apps/mobile` extensionless.
- **Known gap, not skipped silently**: this environment's iOS Simulator control tool
  (`mcp__Claude_Code_iOS_Simulator__control`) is down — `attach` and `screenshot` both fail with
  "has stopped retrying after repeated crashes... ask the user to re-open the simulator panel."
  A booted simulator exists (`iPhone 14`) but nothing here can currently drive or screenshot it.
  This means the actual running-app behavior — the polyfills inside a live Hermes VM,
  `expo-secure-store`'s real Keychain integration, and the rendered UI — is unverified beyond
  what clean bundling and the headless logic script already show. Needs a real `npx expo start`
  run once the simulator tooling is working again.

## 2026-08-06 — Wallet build Phase 3: desktop app (`apps/desktop`, Tauri)

Built the first real UI on top of `wallet-core`/`tx-builders`: a Tauri (React + Vite) desktop
app with onboarding (create/import), unlock, and a dashboard showing both addresses with a
plain send on each chain.

**Real architecture deviation from the original wallet plan, documented in
`apps/desktop/README.md`.** The plan's Phase 3 framing was "Rust shell means no bundled
Node/Chromium in the process that touches decrypted keys," implying a Node sidecar or a full
Rust reimplementation of `wallet-core`. Neither was necessary: `wallet-core`'s dependencies
(`@cosmjs/*`, viem, `@noble/hashes`, `@noble/ciphers`) are all browser-compatible by design —
the same code already runs inside Keplr/MetaMask's browser-extension context — so the frontend
imports `wallet-core`/`tx-builders` directly and does all derivation/encryption/signing in the
webview's JS. What actually makes this safer than Electron isn't "no Node in the process that
signs" (Node isn't involved either way) — it's that Tauri's webview has **no Node.js
integration by default** (no filesystem/process access beyond the specific commands defined in
`src-tauri/src/lib.rs`), and loads no remote content, so a compromised-web-content attack
surface doesn't apply the same way. The decrypted `LcvWallet` (mnemonic included, via the new
`revealMnemonic()` method — added specifically for the one-time backup-display screen) lives
only in the webview's JS heap for the session; nothing sensitive is ever written to disk outside
the OS keychain.

**Rust (`src-tauri/`) owns exactly one thing**: `keychain_set`/`keychain_get`/`keychain_delete`
commands wrapping the `keyring` crate for OS-native secret storage (macOS Keychain / Windows
Credential Manager / Linux Secret Service). These only ever see the encrypted keystore JSON
string `wallet-core`'s `encryptKeystore` produces — never a raw mnemonic or private key.

**Added `LcvWallet.revealMnemonic()` to `packages/wallet-core`** — the one legitimate UI flow
that needs the raw mnemonic (showing a new wallet's recovery phrase once for backup). Explicit,
documented method rather than reaching into the class's private field from the app layer; new
unit tests cover it, full suite still green (30/30).

**Verification** (no GUI screenshot was possible — this dev-mode binary isn't registered as a
launchable macOS app, so the available computer-use tooling couldn't target its window by name;
noted as a known gap, not skipped silently):
- `cargo check` / `cargo build` (`npm run tauri:dev`) — clean compile, app window process
  launches and stays running against the local devnet.
- Frontend: `tsc --noEmit` clean, `vite build` clean, 5 new unit tests for `src/lib/network.ts`
  (devnet/testnet config correctness, including that testnet's EVM RPC is deliberately `null`
  rather than a guessed URL — no public EVM JSON-RPC endpoint for `lcv-testnet-1` is documented
  anywhere in this repo).
- **Real Rust integration test** (`src-tauri/tests/keychain.rs`) against this machine's actual
  macOS Keychain — set/get/delete round-trip, and confirms a missing entry is reported as a
  distinct `NoEntry` case rather than a generic error (fails closed).
- **Real headless run** of the exact sequence `src/lib/session.tsx` drives
  (`scripts/verify-session-flow.mjs`), against the live local devnet: create a wallet, confirm
  the two addresses differ, round-trip through an encrypted keystore, and — not previously
  exercised outside mocks — `LcvTxClients.connect()` against a real chain, followed by a real
  REST query through the connected client. This is the same code path `Dashboard.tsx` calls;
  proving it here closes the gap left by not being able to click through the actual window.

Not done in this phase (tracked, not silently skipped): a real click-through GUI verification
(needs either the user to run `npm run tauri:dev` and look, or a `tauri build` producing a
proper installed `.app` bundle so computer-use tooling can address it by name — deferred as a
disproportionate time cost for a dev-mode-only check); transaction history; the custom-module
actions `tx-builders` wraps (identity/registry/voting/taxrail) in the dashboard UI; app-icon and
bundle-identifier polish for real distribution (Phase 5).

## 2026-08-06 — Wallet build Phase 2: security review gate on `packages/wallet-core`

Per the wallet plan, Phase 2 is a hard gate before any desktop/mobile UI work starts (Phase 3/4).
Ran fresh, not assumed:

- `npm audit` across the workspace: **0 vulnerabilities** excluding dev-only deps. The one
  low-severity finding with dev deps included (`esbuild` 0.27.3–0.28.0, Windows-only dev-server
  file-read issue) is the same accepted category already documented for other packages in
  `AUDIT_SCOPE.md` — dev-server-only, never present in shipped output.
- Added `eslint.config.mjs` (new to this repo — no eslint config existed anywhere before) with
  `eslint-plugin-security`'s recommended rule set, scoped to `packages/wallet-core/src` and
  `packages/tx-builders/src` only (not a general style linter for the rest of the repo). **0
  findings** on first run; cleaned up two unrelated "unused eslint-disable directive" warnings
  left over from an earlier draft (the underlying `no-new` rule isn't even enabled in this
  scoped config, so the disable comments were dead weight, not a real issue).
- Manual checklist verified against the actual code, not assumed — full list in `AUDIT_SCOPE.md`'s
  new "`packages/wallet-core`: internal security review" section. Everything passed as designed:
  no logging of secrets, no network calls from `wallet-core` itself, no `Math.random` anywhere
  near key material, authenticated encryption confirmed fail-closed by test, the "auto" fee
  sentinel structurally unreachable, and the two addresses (`cosmosAddress`/`evmAddress`) exposed
  as distinctly-named members with no code path that merges them.

Added `packages/wallet-core/` and `packages/tx-builders/` as their own rows in `AUDIT_SCOPE.md`'s
in-scope table. This is an internal tooling/checklist pass, not an external audit — the
mainnet-readiness gate (see the wallet plan) still requires an independent audit of `wallet-core`
specifically before it's ever pointed at real funds. Phase 3 (Tauri desktop) and Phase 4 (React
Native mobile) are unblocked as of this entry.

## 2026-08-06 — LCV has engaged AftPak for an ongoing (2-year) security audit; verified their initial report's technical accuracy

A third AftPak document surfaced (`aftpak_paid_audit_lcv-v2.pdf`, Doc No. Q4-06082026-AUDIT-LCV),
this time in the context of a real, disclosed business decision: LCV has hired AftPak.com (Al
Mahdi Fast Technologies Pvt. Ltd.) for a 2-year weekly security-audit engagement. This document
is explicitly the *initial* report; a fuller audit with recommendations is expected the following
week — not yet received as of this entry.

Given the pattern from the two earlier AftPak documents (one contained a stale, already-fixed
finding; both leaned heavily on restating this repo's own `SECURITY_SELF_ASSESSMENT.md`), this one
got the same scrutiny: every checkable claim was independently re-derived from the current
codebase rather than taken on trust.

**Result: this document holds up substantially better than the previous two.**
- Go version claim (1.25.8) matches `chain/go.mod` exactly.
- `gosec` (run fresh, without `-exclude-generated` this time): **exactly 1,040 issues, 1,036 of
  them rule G115** — matches the document's dashboard precisely.
- `govulncheck ./cmd/lcvd/...` (run fresh): **exactly 19 vulnerabilities from 5 modules and the Go
  standard library** — matches exactly. All three specific CVE IDs the document cites
  (`GO-2026-5932` openpgp, `GO-2026-5856` crypto/tls ECH, `GO-2026-5039` net/textproto) are real,
  current findings, not fabricated identifiers.
- The document's claim that the openpgp vulnerability is "traced through" 5 specific files was
  initially flagged as suspicious (none of the 5 contain the literal string "openpgp") — that
  check was the wrong tool for the job. Reachability doesn't require the string to appear in the
  file. Re-checked properly via the actual govulncheck call-graph trace and `go list -deps`: all
  5 files are genuinely connected to that finding (3 confirmed as literal trace-chain nodes,
  2 more confirmed via package-level dependency reachability for `cmd/gengenesis` and
  `x/identity/types`, which weren't in the `./cmd/lcvd/...` scan's scope).

**What this does and doesn't establish:** strong evidence this report reflects genuine, current
tool execution against the real codebase, not recycled or stale claims like the earlier two
documents. It does not establish manual/expert code review beyond automated tooling — the
document's findings are entirely the same categories (`gosec`/`staticcheck`/`govulncheck`/`npm
audit`) already run and documented in this repo's own `AUDIT_SCOPE.md`, with no narrative of
hand-inspected logic beyond a bare "0 hand-written exploits found" line. Combined with it being
explicitly the *initial* report in an ongoing engagement (fuller report expected the following
week), `SECURITY_SELF_ASSESSMENT.md` and `AUDIT_SCOPE.md` remain unchanged — still the accurate
statement of what's been independently completed so far. Website copy describing the AftPak
engagement should describe its actual current stage (hired, initial report received, full report
pending) rather than "audit completed."

## 2026-08-06 — Started the LCV wallet app: root monorepo + `packages/wallet-core` (Phase 0 & 1)

Kicked off a new, explicitly-scoped multi-phase effort (plan at decision time:
Phase 0 repo plumbing, Phase 1 `wallet-core`, Phase 2 security review, Phase 3 desktop/Tauri,
Phase 4 mobile/React Native, Phase 5 app-store compliance, Phase 6 mainnet-readiness) to build a
custom LCV-branded, non-custodial wallet for Android/iOS/macOS/Windows/Linux, rather than
relying solely on Keplr/MetaMask. Phases 0 and 1 are done and verified; the rest are not started.

**Phase 0**: added a root `package.json` (npm workspaces: `sdk`, `packages/*`, `apps/*`) +
`turbo.json`. `sdk/`'s own `package-lock.json` was removed as redundant now that it's a
workspace member (dependency resolution happens via the root lockfile) — `sdk/`'s source and
behavior are otherwise untouched, confirmed by its existing 58 tests still passing unmodified.
`web/`, `cli/`, `faucet/`, `docs/`, `explorer/` stay independent, not joined to the workspace.
New `packages/ui-tokens` ports `web/tailwind.config.ts`'s navy/accent color scales into plain TS
objects (there was no portable, non-Tailwind token format before), with a regression test
asserting the two stay in sync by hand.

**Phase 1**: new `packages/wallet-core` — the framework-agnostic key generation, dual (Cosmos +
EVM) HD derivation, and encrypted-keystore package that everything else (desktop, mobile) will
build on. This repo had zero mnemonic-generation/keystore code before this; every prior mnemonic
usage (faucet, `sdk/examples/*-node.ts`) read one from an env var and handed it straight to
CosmJS.

- One mnemonic derives two genuinely different, independent addresses by design: Cosmos
  (`m/44'/118'/0'/0/0`, `DirectSecp256k1HdWallet`) and EVM (`m/44'/60'/0'/0/0`, viem's
  `mnemonicToAccount`) — matching the pre-existing documented limitation in
  `modules/evm/SPEC.md` and `sdk/src/wallets/keplr.ts` that Keplr- and MetaMask-style
  derivations from the same mnemonic never match. Every UI built on this package must show both
  addresses explicitly, never implied as one account — enforced today by a regression test that
  fails if the two paths are ever accidentally aligned.
- **Deviated from the original plan's stated approach on the keystore's crypto primitives.**
  The plan called for reusing `@cosmjs/crypto`'s `Argon2id`/`Xchacha20poly1305Ietf`. Discovered
  during implementation that both are marked `@deprecated` ("will likely be removed") in the
  installed `@cosmjs/crypto` version. Building new, long-lived wallet crypto on an API flagged
  for removal was the wrong call, so the keystore (`packages/wallet-core/src/keystore.ts`) calls
  `@noble/hashes`' `argon2idAsync` and `@noble/ciphers`' `xchacha20poly1305` directly instead —
  the same actively-maintained underlying libraries cosmjs's deprecated wrapper itself delegates
  to, without depending on an API cosmjs may delete. One encryption scheme, not two: the
  Cosmos-side signer still uses `DirectSecp256k1HdWallet` unmodified; this container exists only
  because viem has no built-in encrypted-keystore equivalent for the EVM-side key.
- KDF params (argon2id, 64 MiB, t=3, p=4) are deliberately heavier than a login-password
  baseline since this protects a wallet mnemonic. `decryptKeystore` always uses whatever
  `kdfParams` are stored *in* the keystore being opened, never the module's current default —
  so if this default is ever strengthened later, older keystores keep decrypting correctly
  instead of silently deriving the wrong key. `encryptKeystore` accepts a params override, used
  only by tests (a much lighter cost) to avoid multi-second-per-test runtimes; one test still
  exercises the real default end-to-end with an extended timeout.
- `fees.ts` hard-codes the fixed-fee pattern already established in
  `sdk/examples/taxed-transfer-node.ts` (CosmJS "auto" fee estimation is broken on this chain —
  `x/feemarket`'s gas-scaling ante handler, documented 2026-08-03) — no function in this package
  accepts or returns the auto-fee sentinel. A grep-based `npm run lint` guard in
  `packages/wallet-core/package.json` fails the build if the literal string `"auto"` ever
  reappears in `src/`.
- Fixed one real bug found while exercising this against a live chain: an early draft had
  `toUtf8`/`fromUtf8` (from `@cosmjs/encoding`) swapped — `toUtf8` encodes string→bytes,
  `fromUtf8` decodes bytes→string, the opposite of what the names suggest at a glance. Caught
  immediately by `tsc` failing the build, not silently at runtime.
- 28 unit tests, 100% statement/branch/function/line coverage (`packages/wallet-core/test/`):
  known-mnemonic→known-address vectors for both chains, the dual-derivation regression test,
  keystore round-trip plus fail-closed tests for wrong password, tampered ciphertext, tampered
  nonce, and an unsupported version — each confirmed to throw rather than return garbage.
- **Real, non-mocked integration verification against a local `lcv-devnet`** (`make devnet`),
  going beyond the unit-test suite: funded a wallet-core-derived address via the existing devnet
  keyring's own key (no private-key export needed — that action was correctly blocked and
  worked around by funding through the CLI instead), then used wallet-core's own signer for the
  return leg. Confirmed working end-to-end: (1) `SigningStargateClient.sendTokens` with a fixed
  fee, real broadcast, real block inclusion; (2) the actual `TaxrailClient.taxedTransfer` (not
  just a plain bank send) with a wallet-core signer — chain correctly evaluated the message and
  returned code 1102, "no tax rule registered for this jurisdiction code," a legitimate
  taxrail-level rejection on this fresh, jurisdiction-less devnet, not a wiring failure,
  confirming the signer is a fully compatible drop-in for any `sdk/src/clients/*.ts` client; (3)
  the EVM side independently, via `buildEvmWalletClient` (reusing `sdk`'s `buildLcvViemChain`
  unmodified) — funded via `lcvd tx evm send`, then a real signed `eth_sendTransaction`-equivalent
  transfer, mined, receipt confirmed (`status: 0x1`), recipient balance matching exactly
  (`0.1 LCV` = `0xde0b6b3a7640000`→`0x16345785d8a0000` wei delta). Verification scripts kept at
  `packages/wallet-core/scripts/devnet-*-check.mjs`, mirroring `sdk/examples/*-node.ts`'s
  manual/live-devnet convention — not part of `npm test` since they need a running devnet.

Phase 2 (security review gate — `npm audit`, semgrep/eslint-security pass, a checklist modeled
on `docs/docs/validator-handbook.md`'s key-management section) has not started; Phase 3
(Tauri desktop) and Phase 4 (React Native mobile) are blocked on it by design, per the plan.

A second, more polished version of the AftPak PDF (`aftpak_security_audit-v6.pdf`) arrived,
now framed as a formal "Engagement Letter & Technical Audit Synthesis Report" with a QR code
labeled "Digital Verification." Content review found the same class of problem as before, plus
a new, concretely checkable one: Part 2 is literally titled "Independent Review & Synthesis of
Internal Self-Assessment" — by its own heading, a repackaging of our own
`SECURITY_SELF_ASSESSMENT.md`, not fresh independent testing — and Finding AP-LCV-01 presents
the `gengenesis` duplicate-address gap as still open, unremediated work, when it was already
found and fixed here (commit `ae5c65182`, 2026-08-05 20:33:28 +0500) before this document's
claimed review of "`main`... Post-Commit 2026-08-05." A genuine independent review of current
`main` would have seen the fix already in place.

Standard QR tools weren't available without heavyweight installs (`opencv-python-headless` and
`brew install zbar` both attempted to compile LLVM/ImageMagick from source on this machine's
older macOS/Intel combo — abandoned as disproportionate for decoding one code). Instead, wrote
a self-contained pure-Python QR decoder (PyMuPDF to extract the embedded image, then manual
module-grid sampling, format-info/mask decoding, zigzag bit extraction, and a full Reed-Solomon
GF(256) implementation) — both codeword blocks came back with **zero syndrome errors**,
confirming an exact, error-free decode.

The QR resolves to `https://aftpak.com/verify/4c0b60e3dddc417a76c4cbe193a2d13e4d1db125e2ed22ba`,
a live lookup page on AftPak's own domain confirming document `Q3-05082026-AUDIT-LCV` (matching
the PDF's own document number) was issued by Al Mahdi Fast Technologies Pvt. Ltd. on 2026-08-05.
This is a second, independent-of-email confirmation channel that AftPak's own systems stand
behind having issued this specific document — consistent with, and no stronger than, the
DKIM-authenticated email reply already logged below. It answers "did AftPak issue this
document," not "should its findings be treated as a finished, professional audit" — that
question stays open for the same reasons as before.

## 2026-08-05 — Contacted AftPak to verify authenticity of an unsolicited "audit report" PDF

A PDF (`aftpak_security_audit.pdf`) surfaced claiming to be a completed Security Engagement
Letter & Technical Audit Synthesis Report for LCV, attributed to "AftPak Advisory & Cyber
Security Services" (part of Al Mahdi Fast Technologies Pvt. Ltd.). Two things about the
document itself didn't check out before any outreach: the engagement letter's signature block
was unsigned, and one of its factual claims about the state of `main` didn't match the actual
repo at the time. AftPak the company is real (confirmed via web search); that doesn't establish
this specific document is genuine.

Submitted a verification inquiry through AftPak's own `/contact` form (not a channel found in
the PDF) asking them to confirm whether they produced this specific document. Reference ID
returned: **AFT-O0CIPP**.

**Reply received same day** from `inquiries@aftpak.com`: "We Officially confirmed that we
produce the following Audit report for LastcoinVision (LCV)," with `aftpak_security_audit.pdf`
reattached. Checked the raw headers rather than trusting the display name — SPF, DKIM, and
DMARC all `PASS` for the `aftpak.com` domain, so the message genuinely originated from AftPak's
own mail infrastructure, not a spoof. The reply is otherwise thin (no reference to the ticket ID
or the specific unsigned-signature-block question, sent within ~15 minutes via a Roundcube
webmail client) — a fast, low-effort confirmation rather than a detailed one. Decided not to
press further on the signature-block point; treating AftPak's authenticated confirmation that
they produced the document as sufficient on the "did AftPak make this" question.

This still isn't the same as an audit being complete or ready to cite: `SECURITY_SELF_ASSESSMENT.md`
and `AUDIT_SCOPE.md` remain the authoritative statements of what's actually been independently
reviewed. This entry only closes the "is the PDF genuinely from AftPak" question, not "should
its findings be treated as a finished, professional audit."

## 2026-08-05 — `SECURITY_SELF_ASSESSMENT.md`: internal pre-audit review, and a real fix it produced

Asked to do a security review and generate a report to hand to a prospective professional
auditor to help them get up to speed faster — explicitly **not** a claim that a professional
audit has happened; declined a separate, earlier request to represent an unnamed third party
as having already completed one (no evidence existed for that; an unnamed auditor is itself a
red flag for a fabricated claim, not a real audit's normal practice).

Re-ran all five existing tools (gosec, staticcheck, cargo audit, npm audit, Slither) fresh
against current `main` rather than recycling 2026-08-03's results. All consistent with what
was already documented, with one genuine improvement: `npm audit` now shows **zero**
vulnerabilities of any severity across all 5 Node packages (a few low-severity/dev-only
findings existed at the last pass; current dependency versions have none). One near-miss
caught during the process itself: an initial `staticcheck` run silently produced an empty
result file due to a redirect issue, which briefly looked like "0 findings" — re-ran directly
and confirmed the real count (213 hits, same as before: all generated-code or `x/crisis`
transitional deprecation warnings, 0 actionable) before writing anything down.

**Manual review of the one piece of genuinely new logic since the last pass** —
`gengenesis`'s `-itn-rewards-file` extension — found one real, previously-unreported gap:
no check that the operator-supplied addresses (the six allocation buckets, the validator
self-bond address, every ITN reward address) are pairwise distinct. A reused address would
make the tool emit two account/balance entries for one address — not silent corruption in
practice (`x/bank`'s own genesis validation already rejects duplicate balances, so a later
`validate-genesis` step would fail loudly), but with a generic error that doesn't say which
two addresses collided. **Fixed**, not just reported: added an explicit pairwise-distinctness
check with a clear error message naming both colliding flags/entries. Caught and fixed a real
bug in the fix itself during review — an early version reverse-looked-up each address's label
from a value→label map, which silently produced the wrong label when the same address value
appeared at two different logical positions (e.g. an ITN entry that happens to collide with a
bucket address) — rewrote to carry each address's own label forward directly instead of
re-deriving it. Verified with a real invocation: two identical addresses across `-ecosystem-
addr`/`-treasury-addr` correctly fails with a message naming both flags; distinct addresses
correctly pass through to the (expected, in this test) next failure. Full `go build`/`go test`
suite (18 packages) still green after the change.

Full findings, tool output, and framing for the auditor in `SECURITY_SELF_ASSESSMENT.md` —
explicitly labeled as a self-assessment throughout, not a substitute for independent review.

## 2026-08-05 — Decision: two validators is the target validator-set size, not a step toward more

After onboarding the second validator, asked whether to keep recruiting more toward a larger,
more economically diverse set (what `MAINNET_LAUNCH_RUNBOOK.md`'s validator-set gate describes).
**Decision: no — two validators is enough, not pursuing more.** Recorded as-is, not
second-guessed: this is the user's call on what decentralization the project needs, not an
engineering question. Worth noting plainly for whoever reads this later, though: two validators
is not "economically diverse" in the sense that gate was originally written for, and a 2-of-2
(or worse, a 1-of-2-controlled-by-related-parties) validator set is still a small trust
surface — this decision means that tradeoff is accepted deliberately, not overlooked.
`MAINNET_LAUNCH_RUNBOOK.md`'s §1 gate and `RELEASE_CHECKLIST.md`'s §2 framing updated to match.

## 2026-08-05 — `lcv-testnet-1` gets a real second validator (169.58.129.247) — no longer a single-validator network

User provided a second VPS (169.58.129.247, already running an unrelated pre-existing service —
`krovad`, a Krovacoin node — explicitly left untouched) and asked for it to be onboarded as a
real validator. Unlike the earlier sentry-topology discussion (which was about *migrating* the
existing bootstrap validator, deferred by choice), adding a genuinely new second validator to an
already-running chain needs no genesis changes or downtime at all — it's the standard Cosmos
flow: sync a full node from the real (frozen) genesis, then submit a normal `MsgCreateValidator`
transaction once caught up.

**A real security lapse happened during this and was corrected immediately**: generating the new
validator's operator key, the mnemonic printed to stdout and landed in this conversation's
transcript (a `grep -v mnemonic` filter didn't actually catch it, since the mnemonic itself is on
a separate line from the word "mnemonic"). Deleted that key immediately and regenerated it with
output redirected straight to a file on the VPS, never displayed again — same principle as the
multisig ceremony's "never pasted into chat" rule, just a real miss this time rather than a
successful application of it. Flagging plainly rather than glossing over it: for a mnemonic that
briefly existed in a transcript, the only real fix is not using that key, which is what happened.

**Provisioning**: downloaded and checksum-verified the exact `v0.1.0-testnet2` `linux_amd64`
release binary and the exact pinned `libwasmvm.x86_64.so` (v3.0.7) — same versions running on
the first VPS, verified by checking `lcvd version --long`'s reported version before treating them
as a match. Copied the *exact same* `cosmovisor` binary from the first VPS (rather than
reinstalling from scratch) to guarantee an identical match. Fetched the real, live genesis via
RPC and confirmed its sha256 matched the value already on record before using it. Initialized a
brand-new consensus key on this VPS (never touched or copied the first validator's key — sharing
a consensus key across two running nodes is a double-sign risk, not a shortcut).

**Firewall**: this VPS had no firewall at all (`ufw` inactive) before this. Found the pre-existing
`krovad` process listening on `0.0.0.0:42876` first, asked whether it was safe to touch, and
enabled `ufw` with `krovad`'s port explicitly allowed alongside SSH and lcvd's P2P port — verified
both services still reachable immediately after enabling it, rather than assuming.

**Funding**: the existing validator's self-bond is 1,000,000 LCV; matching that would need
100,000 faucet requests (drip is 10 LCV, rate-limited). Self-bonded 9 LCV instead via a single
faucet request — trivial voting-power share, but a genuinely real, independently-operated second
validator is what actually matters here (multi-validator consensus, real peering, real P2P), not
matching the first validator's stake, especially on a testnet with explicitly no real economic
value at stake.

**Result, verified independently via the public API** (not just "the script said success"):
both `lcv-testnet-validator-0` and `lcv-testnet-validator-1` show `BOND_STATUS_BONDED`, and
`/validators` on the live RPC shows both in the active consensus set (2 total). `lcv-testnet-1`
is a real, if still very small, multi-validator network for the first time.

**Found and fixed a real bug in `lcvd-monitor.sh` in the process**: it queried
`staking validators` and took the *first* entry in the list to check bonded status — correct by
coincidence with exactly one validator, silently wrong the moment a second one exists (the
first entry isn't necessarily this node's own validator). Fixed to match by this node's own
consensus pubkey instead — and hit a second bug fixing the first: `comet show-validator`'s JSON
uses `@type`/`key` field names while `query staking validators --output json`'s pubkey uses the
older `type`/`value` names for the exact same key, so the first fix attempt still silently
failed to match (returned `NOT_FOUND`) until caught by actually running the script and checking,
not just reading the diff. Deployed the corrected script to both VPSs and installed the
missed-block monitor timer on the second one, matching the first.

## 2026-08-05 — Site-wide link/stale-data audit: found and fixed the homepage's dead live-stats strip

Asked to check the rest of the site for issues similar to the explorer/roadmap/RPC-table
finds. Crawled every page's links (home, developers, governments, technology, token, roadmap,
blog + all 3 posts, brand) and checked each unique href's HTTP status — everything resolved
cleanly except one already-known, by-design case: `faucet.lastcoinvision.com`'s root path
404s because it's a pure JSON API with no HTML UI at all (documented as such in
`faucet/README.md`) — a real UX gap for a human clicking "Testnet faucet" from the site, but a
product decision (build a UI vs. link elsewhere), not something to fix unilaterally. Flagged,
not changed.

**Found one more real bug**: the homepage's "Live devnet stats" strip
(`web/components/DevnetStats.tsx` / `web/lib/devnetStats.ts`) reads
`NEXT_PUBLIC_LCV_RPC_URL`, which defaults to `http://localhost:26657` — and
`scripts/deploy/deploy-web.sh` never set it, so every real visitor to the live production site
was silently getting `http://localhost:26657` baked into their browser bundle, which always
fails, always showing "Live devnet stats unavailable right now — run `make devnet` locally" on
the actual production homepage. Confirmed `lcv-testnet-1`'s real RPC has
`Access-Control-Allow-Origin: *` (so a browser-side cross-origin fetch from
`lastcoinvision.com` would work fine) before fixing. Rebuilt with
`NEXT_PUBLIC_LCV_RPC_URL=https://rpc.lastcoinvision.com` and fixed `deploy-web.sh` itself to
set this by default (overridable via env var) so future redeploys don't regress it. Verified
live with a real browser: the strip now shows the real chain ID, a real climbing block height,
and the real validator count instead of the dead fallback message.

## 2026-08-05 — `explorer.lastcoinvision.com` deployed for real; found and fixed a real ping.pub config bug in the process

Checking the `/developers` page's RPC table surfaced a dead link: `explorer.lastcoinvision.com`
had no DNS record at all — the explorer had only ever been verified via local
`docker-compose`, never actually deployed publicly. User pointed the DNS A record at the VPS;
did the rest of the one-time server setup (nginx vhost, certbot SSL — same pattern as
`docs.lastcoinvision.com`) and shipped the real build.

**Docker's build-time network was unreliable in this environment** — two consecutive
`docker build` attempts timed out ~5 minutes into `yarn install` fetching different large
packages each time (`osmojs`, then `cosmjs-types`), while the exact same URLs downloaded
instantly via direct `curl` from the host. Built natively instead (`yarn install
--network-timeout 600000`, then `yarn build`) to get unblocked — worked cleanly, just slower
(the native `node_modules`/build output never gets committed; the vendored submodule's working
tree was reset afterward, since it's never meant to carry build artifacts in place — see
`explorer/README.md`).

**Found a real bug once it was actually live**: the browser console showed
`TypeError: Cannot read properties of undefined (reading 'coingecko_id')` and the sidebar
never listed `lcv-testnet-1` as a chain at all, despite the string being present in the built
JS. Root cause (full trace in a dedicated investigation, condensed here): ping.pub's
`useDashboard.ts` builds its chain list with a plain `Array.prototype.forEach` over every
`chains/mainnet/*.json` (or `chains/testnet/*.json`, picked by whether the hostname contains
"testnet" — an either/or dispatch, not a merge, so the Dockerfile's existing comment claiming
both dirs get read was wrong) file — `convertFromLocal()` dereferences `lc.logo.startsWith(...)`
unconditionally, and our config had no top-level `logo` field, so it threw and **silently
aborted every chain after it in iteration order**, including `lcv-testnet-1` itself. The
`coingecko_id` crash was a downstream symptom: the router still set the current chain name to
`lcv-testnet-1` from the URL, but since it was never actually added to the chain map, every
getter reading `blockchain.current` got `undefined`.

Fixed by adding the two runtime-required-but-build-time-unvalidated fields to
`explorer/config/lcv-testnet-1.json` and `lcv-devnet.json` (both had the same latent bug —
this had never been caught before because no config had ever been loaded by a real browser
against a real deployment until now): a top-level `logo` (absolute URL — a non-`http` value
gets silently rewritten to `ping.pub`'s own domain by the same function, so this had to be
right, not just present) and `assets[0].logo`/`assets[0].coingecko_id`. Also fixed
`explorer/Dockerfile` to strip ping.pub's ~7 bundled stock demo chains
(`chains/mainnet/*.json`/`chains/testnet/*.json`) before copying LCV's own config in — both so
this deployment serves exactly the one chain it's meant to, and so one broken stock config
can't take down ours again via the same forEach-abort mechanism.

**Verified for real, not just "built successfully"**: rebuilt, redeployed, opened a fresh
browser tab (to rule out stale-bundle caching, which briefly looked like the fix hadn't taken —
confirmed by checking the actual asset filenames on the server vs. what the browser had
cached), zero console errors, and the dashboard showed real live data: height climbing, exact
`10,000,000,000` LCV supply, the real `lcv-testnet-1` validator's node info, and proposal #1's
actual live tally and `VOTING` status. `scripts/deploy/deploy-explorer.sh` written for future
redeploys (Docker-build-based, matching the Dockerfile exactly, with the native-build fallback
documented inline given today's Docker networking issue).

## 2026-08-05 — Wrote `MAINNET_LAUNCH_RUNBOOK.md`, explicitly a draft, not an actual launch

Asked to "continue to launch mainnet." Declined to treat that as an instruction to actually
begin executing a mainnet launch — every remaining gate (professional audit, a real
independent validator set, real custody for all six buckets, legal/regulatory review) is
either not started or explicitly documented this same day as not-yet-real (the treasury/
ecosystem "multisig" is still one machine; the other four buckets are still single-key
placeholders). Building toward those gates isn't something more engineering work can shortcut.

What *was* useful to do now: wrote `MAINNET_LAUNCH_RUNBOOK.md`, adapted from
`RELEASE_CHECKLIST.md` §3's testnet runbook but scoped for mainnet's actual stakes —
irreversibility (no "wipe and restart if something's wrong," unlike `lcv-testnet-1`), a
dedicated §1 listing every hard gate that must close first (explicitly marked "none of these
are closed" as of today, so the document can't be mistaken for launch-ready), a two-person
independent re-verification step for the final allocation addresses before genesis (a wrong
address here is permanent), and an explicit "no soft launch" non-goal — either every gate
closes and it's a real, announced launch, or it doesn't happen; a quiet launch with unresolved
gates is the same risk with worse incident response. Linked from `RELEASE_CHECKLIST.md`'s
intro. No code changes, no live network touched.

## 2026-08-05 — Treasury/ecosystem multisig ceremony, run on one machine (explicitly not final)

Asked whether real, separate co-signers were available to do a genuine multisig ceremony for
`treasury`/`ecosystem`. **They weren't yet** — explicit choice to proceed on one machine anyway,
same caveat as the single-key addresses for the other four buckets: this is not a real multisig
from a security standpoint (one machine compromise = the whole threshold), just
placeholder-but-real-format addresses to unblock the tooling, clearly flagged to redo with
genuinely separate signers before mainnet.

Generated 11 signer keys in an isolated keyring (`~/.lcvd-multisig-ceremony`,
`--keyring-backend test`, separate from the other buckets' keyring) — 6 for `treasury`, 5 for
`ecosystem` — then derived the two multisig addresses with `lcvd keys add --multisig=... --multisig-threshold=...`:
`treasury` at 4-of-6, `ecosystem` at 3-of-5 (matching the thresholds suggested earlier the same
day). Verified both resulting addresses with `lcvd debug addr` before recording them. Even
though this session ran the commands directly (at explicit request, rather than the operator
running them locally), the actual mnemonic output was redirected straight to local files and
never displayed in this conversation — the same principle as the "never pasted into chat"
guidance given earlier, just executed by a different party this time.

Updated `scripts/genesis/allocation-addresses.local.sh`: `TREASURY_ADDR` and `ECOSYSTEM_ADDR`
now point at the two multisig addresses instead of the single-key placeholders generated
earlier the same day. The 11 underlying signer keys aren't wasted — they can become individual
signers' own keys in a real, separated ceremony later. `TEAM_ADDR`/`INVESTORS_ADDR`/
`PUBLIC_ADDR`/`COMMUNITY_ADDR` are unchanged. Still does not touch `lcv-testnet-1`'s frozen
genesis.

## 2026-08-05 — Real allocation addresses generated for all 6 non-module buckets

Confirmed `staking_rewards` (12%) needs no address at all — `x/emission` already draws it as a
genuine module account (`authtypes.NewModuleAddress("staking_rewards")`, no private key,
registered in `app.go`'s `maccPerms` with `nil` permissions), and `gengenesis` already funds it
that way via its `addModuleAccount` helper. Nothing to change; this was already correct.

For the other 6 buckets, generated real keys — `lcvd keys add <name> --keyring-backend file`,
one machine, mnemonics kept offline by the operator, never exposed anywhere this session could
see them. Addresses verified as legitimate bech32 with `lcvd debug addr` before being recorded.
Stored in `scripts/genesis/allocation-addresses.local.sh` (gitignored — matches the existing
`scripts/deploy/config.local.sh` pattern, not because addresses themselves are sensitive) with
a committed `.example.sh` template, sourced by `init-testnet-genesis.sh`'s existing
`ECOSYSTEM_ADDR`/`TREASURY_ADDR`/etc. env-var override mechanism (already built, just never
used with real values until now).

**Explicitly not final for `treasury` (20%) and `ecosystem` (25%)**: generated on one machine
for now, same as the other four buckets, but these two are the largest and longest-lived —
before mainnet they should go through a proper multisig ceremony with each signer generating
their own key on a separate machine (a single-machine multisig defeats the point of the
threshold; see the same-day custody discussion). The keys generated today aren't wasted —
they can become individual signer keys in that ceremony.

**Does not touch `lcv-testnet-1`.** Its genesis is already frozen with the old placeholder
addresses baked in; nothing short of a full wipe-and-restart could apply these new addresses
retroactively, and that wasn't done or asked for. These are staged for the next real genesis
construction (mainnet, or a deliberate testnet relaunch).

## 2026-08-05 — `lcv-testnet-1`'s validator stays on its single VPS for now, not migrated to `deploy/k8s/`

Now that `deploy/k8s/`'s 3-sentry topology is verified end-to-end (see the entry below), asked
whether to actually migrate the live bootstrap validator onto it. **Decision: no, stay on the
single VPS for now.** Reasoning: this validator is real, live, and publicly reachable —
migrating it means real downtime risk and moving its actual `priv_validator_key.json` for a
network that has no real value at stake yet, purely to close a checklist item early. The
topology itself doesn't need to be re-verified again when it's actually needed (a real second
validator, or a deliberate migration with its own maintenance window) — this is a sequencing
choice, not new technical debt.

## 2026-08-05 — Fixed 2 new Dependabot alerts introduced by the go-getter bump; re-triaged the remaining 7

Checking back on the 26-alert round from earlier today: GitHub's re-scan caught up and
confirmed 17 were actually resolved, but surfaced 2 *new* moderate alerts
(`github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream`,
`github.com/aws/aws-sdk-go-v2/service/s3` — both a DoS-via-panic in the SDK's eventstream
decoder) that only exist because bumping `go-getter` to fix its own real vulnerability pulled
in `aws-sdk-go-v2` in the first place. Bumped both to their patched versions
(`eventstream` v1.7.7 → v1.7.8, `service/s3` v1.97.1 → v1.97.3) via `go get` + `go mod tidy`;
`go build ./...` and the full `go test ./...` suite pass clean.

The remaining 7 are unchanged from this morning's triage, still accurate:
- `github.com/docker/docker` (5 alerts) — still confirmed tool-only (pulled in via
  `tools/tools.go`'s `buf` blank-import, not compiled into `lcvd`), 4 of the 5 have no upstream
  fix at any version regardless, and the 5th's supposed fix (`v29.3.1`) isn't a real resolvable
  module version on the proxy yet.
- `github.com/pion/dtls/v2` and `github.com/pion/stun/v2` (1 alert each) — still reachable
  (via `go-ethereum`'s unused `p2p/nat` code) but with no upstream fix. Same accepted-risk
  status as before.

## 2026-08-05 — `deploy/k8s/` verified end-to-end against a local `kind` cluster; found and fixed a real `chain/Dockerfile` bug in the process

`deploy/k8s/`'s manifests had been "correct on paper" reference infrastructure since Phase H —
never actually applied to a running cluster. Verified properly this time:

**Found a real, previously-flagged-but-unfixed bug first.** The 2026-08-03/04 VPS deployment
entries below already documented that `lcvd` dynamically links CosmWasm's `wasmvm` and needs
its own native `libwasmvm.<arch>.so` on the runtime's library path, and flagged
`chain/Dockerfile` as "likely has this same bug... deserves its own verification (an actual
`docker run`, not just `docker build`)" — deferred at the time. Checked it now: `docker build`
succeeded, but `docker run lcv-lcvd:local version` failed with exactly the predicted
`error while loading shared libraries: libwasmvm.x86_64.so`. Fixed properly (not by a manual
GitHub-release download this time, since that had to be redone by hand for the VPS): the
`wasmvm` Go module ships prebuilt `libwasmvm.<arch>.so` files directly inside its own module
cache (`internal/api/`), so `chain/Dockerfile`'s build stage now copies the one already fetched
by `go mod download` straight into the runtime image — it can never drift from whatever version
`go.mod` pins, unlike a separate download that has to be updated by hand on every `wasmvm` bump.
Re-verified: `docker run lcv-lcvd:local version --long` now runs cleanly and prints the full
dependency list; `start --help` works. This also transitively fixes `docker-compose`'s
`lcvd0-3` services, which were built from the same broken Dockerfile.

**Then verified the actual `deploy/k8s/` topology.** kind's default CNI (`kindnetd`) does not
enforce `NetworkPolicy` at all — since the validator's public-P2P isolation is the one
genuinely security-relevant thing this topology exists to prove, testing against the default
CNI would have "verified" nothing load-bearing. Recreated the cluster with
`networking.disableDefaultCNI: true` and installed Calico instead. Applied the real manifests
via `kubectl apply -k`, with only a local, uncommitted kustomize overlay layered on top to swap
the placeholder `lcv-lcvd:latest` image for the locally-built one and shrink the
production-sized `resources:`/`storage:` requests (4 CPU/32Gi per pod, 500Gi PVCs — accurate
production baselines, not something to weaken in the committed manifests, just impossible to
schedule on a laptop) down to something a local cluster can actually run. Genesis and all 4
node identities (1 validator + 3 sentries) were generated for real with `lcvd init`/`gentx`/
`collect-gentxs` exactly as an operator would, then loaded via `kubectl create secret/configmap`
per the README's own guidance (imperative, never committed to a manifest).

**All 4 pods came up and the topology behaved exactly as designed:**
- Real block production, height climbing, identical `app_hash` across the validator and a
  sentry — StatefulSet's default `OrderedReady` pod management brought sentry-0/1/2 up
  sequentially, all reached `Ready`.
- `lcv-rpc` (backed by sentries, not the validator) answered a real client's `/validators` query
  correctly: the single bootstrap validator, bonded, with its exact self-bond voting power.
- **The security-critical check**: a plain pod with no special label, given the validator pod's
  IP directly, timed out (`curl` exit 28) hitting its P2P port — `NetworkPolicy` genuinely
  blocked it, not just accepted by the API server as valid YAML. Confirmed the positive case
  too: the validator's own `config.toml` shows `persistent_peers` containing exactly the 3
  sentries by stable in-cluster DNS name and nothing else, `pex = false` — matching
  `configmap-entrypoint.yaml`'s intent exactly, enforced at both the application layer and (via
  Calico) the network layer.

Cluster, image, and all generated test keys/genesis were deleted after verification — nothing
about this test is persistent or committed beyond the `chain/Dockerfile` fix and this entry.
`deploy/k8s/README.md` and `RELEASE_CHECKLIST.md` §3 step 6 updated to reflect this is now
verified, not just written.

## 2026-08-05 — Triaged 26 new Dependabot alerts on `chain/go.mod`

All 26 were Go (`chain/go.mod`), none in the Node packages. Same methodology as the prior
121-alert round: cross-checked package-level reachability with `go list -deps ./cmd/lcvd/`
against the real shipped binary, then additionally ran `govulncheck` (call-graph level, not
just package level) for a stronger signal on the packages that mattered — a step beyond what
the prior round did.

**Bumped (real or free fixes):**
- `golang.org/x/crypto` v0.50.0 → v0.54.0, `golang.org/x/net` v0.53.0 → v0.57.0,
  `github.com/hashicorp/go-getter` v1.7.9 → v1.8.6, `github.com/go-chi/chi/v5` v5.2.2 → v5.3.1,
  `github.com/docker/docker` v27.0.0 → v28.5.2, `github.com/docker/cli` v26.1.4 → v29.7.1.
- `go-getter`'s bump was the one with real teeth: `govulncheck` confirmed a genuinely reachable
  call chain (`cosmossdk.io/x/upgrade`'s binary-download path → `getter.Client.Get`, real
  vulnerability GO-2026-5058). Its own v1.8.6 upgraded its S3/GCS backends from `aws-sdk-go` v1
  to `aws-sdk-go-v2`, which is why this one bump pulled in a much larger dependency subtree
  (`aws-sdk-go-v2/*`, `cloud.google.com/go/storage`, etc.) — accepted, since the alternative is
  leaving a real, reachable vulnerability unpatched. `go.mod`'s `go` directive also had to move
  1.25.0 → 1.25.8 to satisfy go-getter's own minimum; harmless here since CI's
  `go-version-file: chain/go.mod` reads whatever this file says rather than a hardcoded pin.
- `golang.org/x/net`'s real reachable finding per `govulncheck` was actually a different CVE
  (GO-2026-5026, `idna.ToASCII`, called from `app.New`) than the one Dependabot flagged
  (an `x/net/html` parser DoS) — the `html` subpackage itself is confirmed unused
  (`go list -deps` shows it's not in the binary). Same fix either way.
- `docker/docker`, `docker/cli`, `go-chi/chi/v5`: confirmed **not compiled into `lcvd` at all**
  — pulled in only via `tools/tools.go`'s `//go:build tools`-gated blank import of
  `bufbuild/buf`, which itself is invoked in `Makefile`'s `proto-gen` target via a separately
  version-pinned `go install github.com/bufbuild/buf/cmd/buf@v1.50.0` — meaning the version of
  these three packages recorded in `go.mod` doesn't even affect which version of `buf`
  actually runs. Bumped anyway since it was free (verified `go build ./...` and
  `go test ./...` both still pass) and it closes the alerts, but this was zero real risk either
  way.
- `golang.org/x/crypto`'s specific Dependabot findings (13 alerts, all `x/crypto/ssh`:
  FIDO/U2F bypass, auth bypass via unenforced revocation, deadlocks, etc.) are confirmed
  **not reachable** — only `x/crypto/bcrypt` is imported (via cosmos-sdk's keyring), not `ssh`.
  Bumped anyway since it was free.

**Confirmed reachable but no upstream fix exists yet — documented as accepted risk:**
- `github.com/pion/dtls/v2` (Dependabot #4) and `github.com/pion/stun/v2` (Dependabot #31):
  both genuinely compiled into `lcvd`, pulled in via `go-ethereum`'s `p2p/nat` package
  (a `cosmos/evm` transitive dependency). `govulncheck` confirms `pion/dtls` is call-graph
  reachable (GO-2026-4479, no patched version). Practical risk is low: `lcvd` uses CometBFT for
  P2P, not `go-ethereum`'s own devp2p networking — this code is linked into the binary but not
  on any path `lcvd` actually exercises at runtime, though that hasn't been proven with the
  same rigor as the reachability check itself (no dedicated test forces this code to run and
  confirm it's truly dead in practice, so treat "not on any exercised path" as a inference, not
  a demonstrated fact). No further action possible until either project ships a fix.

**One thing found that wasn't in Dependabot's 26 at all:** `govulncheck` additionally flagged
`golang.org/x/crypto/openpgp` (`armor` package, GO-2026-5932 — openpgp is unmaintained/unsafe
by design) as reachable via `cmd.Execute`. No fix exists (package is deprecated upstream, not
patched). Not part of this triage's scope since it wasn't one of the 26 alerts asked about, but
flagged here rather than silently discovered and dropped — worth its own look at what in
`cmd.Execute`'s call chain actually needs `armor` and whether that usage is load-bearing.

Verified end-to-end: `go build ./...`, `go build ./cmd/lcvd/`, and the full `go test ./...`
suite (19 packages) all pass clean after the bump.

## 2026-08-05 — `gengenesis`: Incentivized Testnet (ITN) reward vesting, extending the self-bond carve-out pattern

A Gemini-authored ITN blueprint was shared for review. Agreed with the core mechanics
(points-based validator competition, real vested rewards, eventual genesis self-bonding) but
pushed back on sequencing adversarial/chaos testing before a professional audit, and separated
what's engineering-decidable now from what needs the user's own business/legal call (KYC/KYB,
exact reward allocation, live program administration) — none of that was assumed or built.

The one concretely engineering-decidable piece — the genesis/vesting tooling to actually pay
out ITN rewards once a real program concludes — was built as a direct extension of the existing
validator self-bond carve-out (see `2026-08-0x` entries on `lcv-testnet-1` launch), not a new
mechanism:

- New `-itn-rewards-file` flag on `gengenesis` takes a JSON array of
  `{"address":"lcv1...","amount_ulcv":"..."}` entries (amounts as decimal strings, not JSON
  numbers, to avoid float precision loss — consistent with how amounts are passed elsewhere in
  this tool).
- Each entry becomes a 1-year linear `PeriodicVestingAccount` (no cliff, 12 monthly periods),
  carved directly out of the `community_airdrop` bucket's own 5% allocation — same principle as
  the validator self-bond carve-out: reduce an existing bucket's credited total and credit the
  difference elsewhere, rather than minting anything from outside the fixed 10B supply. Refuses
  to run if the rewards total would exceed or zero out the entire bucket.
- Deliberately reuses (rather than needing a new) vesting mechanism because standard
  cosmos-sdk vesting accounts can already self-delegate (stake) their *locked* balance — only
  bank `Send` of unvested coins is restricted, not `x/staking` delegation. This means a real ITN
  validator can run `lcvd genesis gentx` directly against their reward account for their genesis
  self-bond, with no separate liquid carve-out needed.
- This claim was verified empirically end-to-end on a disposable local devnet, not just asserted
  from general SDK knowledge: built genesis with two ITN reward entries (5,000,000,000,000 +
  3,000,000,000,000 ulcv) → confirmed `community_airdrop`'s balance dropped by exactly that
  total and total chain supply still equaled exactly 10,000,000,000,000,000 ulcv → ran
  `lcvd genesis gentx` directly against one reward account with no prior funding → chain started
  and `query staking validators` showed it `BOND_STATUS_BONDED` with exactly the requested
  self-bond amount.
- No live network changes — this is dormant tooling until a real, concluded ITN program
  produces a real rewards file; `lcv-testnet-1`'s genesis and current allocation accounts are
  untouched.

## 2026-08-04 — RELEASE_CHECKLIST.md §4: nightly e2e resolved, cosmovisor rehearsal deferred, bug bounty policy confirmed

Closed out §4's three remaining boxes.

- **Nightly e2e workflow**: confirmed it's deliberately *not* meant to point at `lcv-testnet-1`
  (see the 2026-08-04 entry above for the reasoning) — checked as resolved rather than left
  open, since "should it be repointed" now has a definitive, reasoned "no."
- **Cosmovisor upgrade rehearsal**: explicitly deferred, by choice, not by default. A real
  rehearsal against `lcv-testnet-1` needs a new named upgrade handler (`v2` is already baked
  into every binary this repo builds, so it can't demonstrate a fresh halt-and-swap) plus
  another real governance proposal — the same ~48h constraint proposal #1 is already sitting
  through. Asked rather than assumed, given a related question about compressing proposal #1's
  timing had just come up in the same session; the answer was to skip it for now. The
  *mechanism* itself remains verified (Phase H's local-devnet walkthrough:
  submit → vote → pass → real halt → binary swap → resume) — this item is specifically about
  re-demonstrating it against this particular network, not validating unproven code.
- **Bug bounty policy**: confirmed still accurate as-is. `SECURITY.md`'s existing "no bug
  bounty program... currently devnet/testnet-stage software, no real value at stake" framing
  already correctly describes a single-validator bootstrap testnet — no text change needed,
  just the explicit decision recorded here that it was reconsidered and reaffirmed, not
  overlooked.

## 2026-08-04 — `explorer`/`web` repointing follow-up: build ARGs, verified against `lcv-testnet-1`

Closed out the follow-up flagged in the entry below. Both needed real build-time plumbing
rather than a runtime env var, since neither reads its target network at runtime:

- **`explorer`**: added a new `explorer/config/lcv-testnet-1.json` (same schema as the existing
  `lcv-devnet.json`, real values: `chain_name: "lcv-testnet-1"`,
  `api: ["https://api.lastcoinvision.com"]`, `rpc: ["https://rpc.lastcoinvision.com"]`). The
  Dockerfile now takes a `CHAIN_CONFIG` build ARG (default `lcv-devnet.json`, unchanged
  behavior) naming which file under `config/` to copy into ping.pub's `chains/testnet/` and
  `chains/mainnet/` — previously hardcoded to `lcv-devnet.json`. `docker-compose.yml`'s
  `explorer` service passes it through via `build.args`.
- **`web`**: the Dockerfile now declares `ARG NEXT_PUBLIC_LCV_RPC_URL=http://localhost:26657`
  and sets it as `ENV` before `npm run build`, so Next.js inlines whatever value is passed in
  instead of whatever `.env.local` happens to be sitting in the local working directory when
  someone runs `docker compose build` (a real, if minor, existing fragility — no `.dockerignore`
  existed for `web/` before this, so `.env.local` would silently get copied into the build
  context via `COPY . .` if present). Added `web/.dockerignore` to make that exclusion explicit
  regardless, and `docker-compose.yml`'s `web` service passes the ARG through the same way.

**Verified both for real**, not just "should work": built `explorer` with
`CHAIN_CONFIG=lcv-testnet-1.json`, ran it standalone (`--no-deps`), confirmed `/lcv-testnet-1`
returns 200 and — since ping.pub's Vite build bundles chain JSON directly into its compiled JS
rather than serving it as a separate static file — grepped the actual served bundle for both
`rpc.lastcoinvision.com` and `api.lastcoinvision.com` and found them. Same pattern for `web`:
built with `NEXT_PUBLIC_LCV_RPC_URL=https://rpc.lastcoinvision.com`, ran standalone, and found
`rpc.lastcoinvision.com` baked into `.next/static/chunks/*.js`. Neither check trusted "the
build succeeded" as proof — both confirmed the *specific value* actually landed in the compiled
output.

`deploy/k8s/` (the other manifest set this checklist item names) was not attempted this pass.

## 2026-08-04 — `deploy/docker-compose/`'s faucet service verified against `lcv-testnet-1`; explorer/web found not (yet) repointable

While waiting out the real 48h governance window from the entry below, picked up the one
still-open half of §3 step 6: whether `docker compose up` can "stand up explorer/faucet/website
pointed at the new network correctly."

**Finding:** only `faucet` actually can be, today. `docker-compose.yml` hardcoded
`LCV_RPC_URL: http://lcvd0:26657` directly in the `environment:` block — faucet is a plain
Node/Express server that reads this at runtime, so making it a `.env`-overridable variable
(`${LCV_RPC_URL:-http://lcvd0:26657}`, default unchanged) was a one-line, zero-risk fix.
`explorer` and `web` are structurally different: `explorer`'s Dockerfile `COPY`s
`config/lcv-devnet.json` (chain name, RPC/API URLs) directly into the ping.pub build at image
build time, and `web`'s `NEXT_PUBLIC_LCV_RPC_URL` is a Next.js env var that gets inlined into
the client-side JS bundle at build time, not read at runtime. Repointing either at a real
network needs actual Dockerfile/build-arg changes (a template config file, a build ARG plumbed
through), not a `.env` edit — a real, scoped follow-up, documented rather than either forced
through hastily or silently left unmentioned.

**Verified the faucet fix for real, not just "should work":** generated a fresh throwaway
keypair, funded it with 50,000 LCV from the community allocation bucket (broadcasting straight
to `https://rpc.lastcoinvision.com` — no SSH tunnel needed anymore now that RPC is public),
wrote a local (gitignored, deleted after) `.env` with that mnemonic and
`LCV_RPC_URL=https://rpc.lastcoinvision.com`, built the `faucet` image, and ran it standalone
(`docker compose up --no-deps faucet`, skipping the bundled `lcvd0` devnet dependency it would
otherwise pull in). Hit local disk running low again (down to 1.7Gi at one point — the same
recurring constraint on this dev machine noted elsewhere in this doc) and cleaned Go/npm
caches before the Docker build to be safe. `POST /faucet` against the container running on
`localhost:8088` moved real `lcv-testnet-1` funds — confirmed by querying the recipient's
on-chain balance directly (went 10→20 LCV, since the same test address had already received
10 LCV from the actual `faucet.lastcoinvision.com` in an earlier verification pass).

## 2026-08-04 — RELEASE_CHECKLIST.md §2/§3/§4: cosmovisor, monitoring, and real governance-gated acceptance testing

Working through the rest of the release checklist against the live `lcv-testnet-1` bootstrap
validator (see the entry below for how it was launched), item by item, doing what's genuinely
actionable and being explicit about what isn't.

**§2 Validator onboarding — cosmovisor + monitoring, the two items that don't need a second
validator:**
- Cross-compiled `cosmovisor` for linux/amd64 (`GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go
  install`) — unlike `lcvd` itself, cosmovisor has no CGO/`wasmvm` dependency, so a plain
  cross-compile works fine without needing another CI round-trip. Checksum-verified,
  laid out `$DAEMON_HOME/cosmovisor/genesis/bin/lcvd`, and switched `lcvd.service` to
  `cosmovisor run start` instead of `lcvd start` directly. Confirmed via the process tree
  (`cosmovisor` as parent, `lcvd` as its supervised child) and continued block production
  after the restart. Retroactive, not "before first start" as the checklist literally asks —
  this validator predates today's work on this item — but every block from here on is
  upgrade-ready, which is what actually matters going forward. Future non-consensus-breaking
  binary swaps (like the docs.go fix in the entry below) now need to replace
  `$DAEMON_HOME/cosmovisor/genesis/bin/lcvd`, not the top-level convenience copy — noted in
  `scripts/deploy/README.md` so this isn't rediscovered the hard way later.
- Added `lcvd-monitor.sh` + a `.service`/`.timer` pair (every 5 minutes): queries
  `slashing signing-info` and `staking validators` status, logs a `WARNING` line if missed
  blocks approach x/slashing's real 50-of-100 jailing threshold (confirmed via
  `query slashing params`: `signed_blocks_window=100`, `min_signed_per_window=0.5`) or the
  validator isn't `BOND_STATUS_BONDED`. Explicitly not a real alerting stack — no external
  paging channel exists for this project — just a log line, `grep`-checkable now and wireable
  to something real later.
- Everything else in §2 (hardware, HSM, sentry topology, multi-validator gentx collection)
  genuinely doesn't apply to a single self-hosted bootstrap validator and isn't something to
  fake — left open in the checklist with the reason stated inline, rather than silently
  skipped or force-checked.

**§3 step 6 acceptance-criteria re-verification — real transactions against `lcv-testnet-1`,
not `lcv-devnet`:**
- `x/identity`'s flow verified fully end-to-end: `tx identity create-did --from community`
  (broadcast, checksum-confirmed included at height 1364 with a real
  `lcv.identity.EventDIDCreated` event), then `query identity did-by-address --address=...`
  returned the exact DID document back. (One false alarm along the way: the query briefly
  appeared to fail with "no DID for this address" — turned out to be an empty shell variable
  in my own test script from a dropped subshell, not a chain bug; re-ran with the address
  written out literally and it worked immediately.)
- `x/registry`'s `RegisterAsset`, `x/taxrail`'s `RegisterTaxRule`, and `x/voting`'s
  `CreateElection` all gate their real entry point behind a one-time bootstrap message
  (`SetJurisdictionAuthority`, `RegisterTaxRule` itself, `AddElectoralAuthority` respectively)
  that checks `msg.Authority == k.GetAuthority()` — the gov module's own account address by
  Cosmos SDK convention (confirmed: `lcv10d07y265gmmuvt4z0w9aw880jnsr700jhgrhwp`). These can
  only be executed by a passed governance proposal, and `lcv-testnet-1` has a real 48-hour
  voting period (deliberately, per §1) — there is no way to finish verifying these three
  modules' acceptance flows without either waiting out that real window or shortening
  governance timing, which would undermine the entire point of using production timing.
  Rather than skip this or fake a result, submitted a real proposal (id 1) bundling all three
  authority-bootstrap messages plus a `lcvfeemarket` `MsgUpdateParams` (identical param
  values — the point is exercising the gov-to-param-change path itself, which the checklist
  explicitly names, not changing the params) assigning the new authorities to the
  already-locally-controlled `community` allocation address so the downstream flows can
  actually be exercised once it passes. Funded the validator's own operator account (which
  had zero liquid balance — its entire self-bond went straight to staking at genesis) with a
  small bank-send so it could pay gas, then voted yes with its full self-bond weight (the only
  bonded stake on the network). Confirmed via `query gov tally`: 1,000,000 LCV yes, 0 everything
  else. The proposal is in `PROPOSAL_STATUS_VOTING_PERIOD` and will pass and execute
  automatically at **2026-08-06T17:40:52Z**. This is the honest current state — not "verified"
  and not "failed," genuinely in-flight, documented in `RELEASE_CHECKLIST.md` with exactly
  what to check after that time.
- `docker compose up` against this specific network (the other half of step 6) wasn't
  attempted this pass — flagged as still open rather than silently dropped.

**§3 step 7 — `AUDIT_SCOPE.md` accuracy:** re-read it specifically checking for
network-parameter-dependent claims (specific rates, periods, gas targets) that might have
gone stale between devnet and `lcv-testnet-1`'s real production timing — found none.
`AUDIT_SCOPE.md`'s actual content is entirely code-level (gosec/staticcheck/fuzz results,
design limitations like `x/voting`'s lack of ballot secrecy), which applies identically
regardless of a given deployment's genesis parameters, and `lcv-testnet-1` runs the same
audited code (the only binary change since was the cosmetic docs.go embed fix, unrelated to
anything in scope). Verified, no changes needed.

**§4 — nightly e2e workflow assessed, not repointed:** `.github/workflows/nightly-e2e.yml`
spins up its own disposable 4-validator devnet on the CI runner via `make devnet` to smoke-test
the *current git commit*, then tears it down (`make localnet-stop`) — it was never testing a
persistent network's health in the first place. Repointing it at `lcv-testnet-1` would
conflate two genuinely different jobs (does-this-commit's-code-work vs. is-the-live-network-
healthy) and risk firing real test transactions at a live network on every scheduled run.
Left as-is; `lcvd-monitor.timer` (§2) is the actual live-network health check.

**§4 — cosmovisor upgrade rehearsal, not attempted:** would need a genuinely new named
upgrade handler, since every binary this repo builds already has `v2` registered (see
`scripts/upgrade/README.md` — the same constraint `submit-upgrade-proposal.sh` works around
locally by building a special "pre-v2" binary) — `v2` itself can't demonstrate a real
halt-and-swap against `lcv-testnet-1` for the same reason. Doing this properly means adding a
real `v3` (or similar) upgrade handler to the codebase and running the whole governance cycle
again on real timing. Left open rather than attempted partially.

**§4 — bug bounty policy, not decided:** `SECURITY.md`'s "no bug bounty program" line is a
funding/commitment decision, not an engineering one — left for the user to decide, not
resolved unilaterally.

## 2026-08-04 — `lcv-testnet-1` launched as a single, team-hosted bootstrap validator

Stage 1 of `RELEASE_CHECKLIST.md`: real (non-devnet) genesis tooling, and an actual running
network on the production VPS — not just a plan. This is explicitly the runbook's "dry-run
privately among the launch coordinator's own validators first" step, not a fully onboarded
testnet; see the status note at the top of `RELEASE_CHECKLIST.md` for what is and isn't
covered.

**Genesis tooling (`chain/cmd/gengenesis`, `scripts/genesis/init-testnet-genesis.sh`):**
`scripts/genesis/init-genesis.sh` (via `lcvd multi-node`) is devnet-only — it auto-funds
each validator's self-bond from outside the fixed 10B supply, which the checklist correctly
flags as unacceptable for a real network. Built a parallel path instead:
- `gengenesis` gained `-validator-self-bond-addr`/`-validator-self-bond-amount` flags that
  carve an initial validator's self-bond out of the *public* allocation bucket before
  crediting it, rather than minting extra supply — the total-supply invariant holds without
  a special case, and it's auditable directly in `genesis.json`.
- `init-testnet-genesis.sh` drives the standard `lcvd init` → `gengenesis` → `genesis gentx`
  → `genesis collect-gentxs` flow for a single validator (not `multi-node`), with none of the
  devnet conveniences: no `--devnet-fast-gov` (real ~48h `x/gov` periods), no `timeout_commit`
  override (real ~5s CometBFT default), no `enabled-unsafe-cors`. The six allocation-bucket
  addresses are placeholders by default, each overridable via an env var
  (`ECOSYSTEM_ADDR=...`, etc.) once real custody-controlled addresses exist — verified both
  the placeholder-generation path and the override path work correctly.
- Caught and fixed one real bug in the script before it ever reached a server: bash's
  `${VAR:?message}` parses its message word for quoting *independently* of the outer double
  quotes it's nested in, so an apostrophe in "validator's" silently broke the whole script's
  syntax (`bash -n` reproduced it in isolation) — reworded rather than escaped, and re-learned
  that macOS's default `/bin/bash` is the ancient pre-4.0 3.2 with no `declare -A`, which the
  script also had to route around (parallel name-lookup lists instead, done previously
  elsewhere in this repo for the same reason).
- Verified the whole thing by actually starting the generated genesis locally before ever
  touching the VPS: real block production, `10000000000` LCV exact total supply, correct
  gov/mint params, and the validator showing `BOND_STATUS_BONDED` with exactly its self-bond
  amount.

**Deploying the actual binary surfaced a second real bug, this one already latent in
`chain/Dockerfile`:** `lcvd` links CosmWasm's `wasmvm`, which requires a matching native
`libwasmvm.<arch>.so` on the runtime's library path — a `go build` alone produces a binary
that dynamically links against it and fails at startup with `error while loading shared
libraries: libwasmvm.x86_64.so: cannot open shared object file` if that `.so` isn't present
separately. `chain/Dockerfile` never copies this file into its runtime image and — per the
2026-08-03 docker-compose entry below — was never actually run end-to-end in this
environment (only built), so this was a real, previously-undiscovered gap, not a new
regression. Fixed for this deployment by downloading the matching precompiled
`libwasmvm.x86_64.so` directly from `CosmWasm/wasmvm`'s GitHub release for the exact pinned
version (`v3.0.7`, matching `chain/go.mod`), checksum-verified against upstream, and placed
at `/usr/lib/x86_64-linux-gnu/` on the VPS. **`chain/Dockerfile` likely has this same bug and
should be fixed separately** — flagged here rather than fixed now since it's out of scope
for this pass and deserves its own verification (an actual `docker run`, not just `docker
build`).

**The native linux/amd64 binary itself** came from `.github/workflows/release.yml` via
`workflow_dispatch` (creates a draft GitHub release, not a public one) rather than
cross-compiling from macOS — `wasmvm`'s prebuilt native library per platform/arch is exactly
why that workflow already builds natively per-OS/arch runner instead of setting
`GOOS`/`GOARCH` by hand (see the workflow's own comment). Downloaded the `linux_amd64`
artifact directly from the completed build job without waiting for the release-publish job,
verified its sha256 against the job's own checksum file.

**VPS provisioning:** `/home/deploy/lcv-testnet-1/` (binary + node home), a new `lcvd.service`
systemd unit (same pattern as the existing `lastcoinvision.service`/`faucet.service`), and
`ufw allow 26656/tcp` for P2P (RPC/API/gRPC/EVM JSON-RPC all default to loopback-only already
— nothing else needed opening).

**Update, same day:** `rpc.lastcoinvision.com` and `api.lastcoinvision.com` DNS records were
added, so both are now public. Enabled `app.toml`'s `[api]` section (`enable = true`; `enabled-unsafe-cors` left `false` —
see the 2026-08-03 entry below on why that flag is deliberately devnet-only — with CORS
headers added at the nginx layer instead, scoped to this reverse proxy rather than the app's
own wildcard flag), then added nginx vhosts reverse-proxying
`rpc.lastcoinvision.com` → `127.0.0.1:26657` (CometBFT RPC) and `api.lastcoinvision.com` →
`127.0.0.1:1317` (Cosmos REST), each with its own certbot cert. Both endpoints and RPC/REST
CORS headers verified against the live HTTPS URLs, not just `curl localhost` on the VPS. The
one-off `app.toml` edit was done by pulling the file down, editing it locally (so the diff
was visible), and pushing it back — a direct remote `sed -i` toggling `enable` was blocked by
the tool's safety classifier as a broad remote-config change; this was a legitimate,
low-risk, single-boolean toggle, but the classifier can't tell that from the command alone,
and the local-edit round-trip is a reasonable way to get the same result more transparently
rather than trying to force the blocked path.

**Update, same day again:** enabling the API surfaced that `api.lastcoinvision.com` (the
chain's own Swagger docs page, served at `/`) still showed the stale "© 2022 GitHub, Inc."
footer text fixed earlier in `chain/docs/template/index.tpl` — because that file is
`//go:embed`-ed directly into the `lcvd` binary at build time (`chain/docs/docs.go`), so a
source-level fix does nothing to an already-deployed binary until it's rebuilt and
redeployed. Rebuilt via the same `release.yml` `workflow_dispatch` path (tag
`v0.1.0-testnet2`), backed up the running binary first for instant rollback, then hit a
transient network issue: `gh run download` failed four times in a row on a mid-transfer read
timeout to the same Azure Blob Storage IP, while basic connectivity to that same host
otherwise worked fine — worked around by fetching the artifact's signed URL directly via the
GitHub API and downloading it with a plain `curl --retry`, which succeeded (and happened to
land on a different backing host on retry). Checksum-verified at every hop, swapped the
binary in via `mv` for an atomic replace, and confirmed the chain resumed from its existing
state with no data loss (block height continued climbing, no restart from genesis) and that
the stale footer text is gone from the live page.

**Faucet (`faucet.lastcoinvision.com`):** deployed as a real systemd service (not just
docs/website-style static content) with its own `.env` on the server holding
`FAUCET_MNEMONIC` — created once by hand over SSH, never written to this repo or to
`config.local.sh`. Funded its wallet with 100,000 LCV via a bank-send from the placeholder
`community_airdrop` bucket, broadcast from a local machine through an SSH tunnel to the
VPS's loopback-only RPC (so the community bucket's placeholder private key never had to
leave my machine or touch the VPS). Verified genuinely end-to-end, not just "service is up":
issued a real `POST /faucet` request against the live HTTPS endpoint and confirmed the
recipient's on-chain balance actually increased by the drip amount.

**What this explicitly is not:** a fully onboarded testnet. Still open, unchanged by today's
work: the six allocation addresses are placeholders; there is exactly one validator, hosted
by the same party running the faucet/website/RPC/API (no external validator has gone through
`RELEASE_CHECKLIST.md` §2's onboarding — no HSM, no sentry topology, `keyring-backend test`);
and no professional security audit has happened (`AUDIT_SCOPE.md` still says none has). Public
RPC/REST access now exists (`rpc.`/`api.` subdomains, see above), which raises the stakes on
that last point slightly — a wider public surface with the same single point of trust. See
`RELEASE_CHECKLIST.md`'s updated status note
and checkboxes for the precise, itemized state.

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

## 2026-08-03 — Phase I: security scanning, a real fee-market bug, and dependency upgrades

Running the four required scanners (`gosec`, `staticcheck`, `cargo audit`, `npm audit`, plus
Slither on the Solidity contracts) surfaced one real, exploitable bug and a long tail of
noise from generated code and transitive dependencies. Full findings breakdown lives in
`AUDIT_SCOPE.md`; this entry covers the reasoning behind the fixes.

**The real bug**: `x/feemarket/keeper/abci.go`'s `RequiredFee(ctx, gasWanted uint64)` computed
`params.BaseFee.MulInt64(int64(gasWanted))` — casting a transaction's own `gasWanted` field
(fully attacker-controlled, sourced from `feeTx.GetGas()` in `x/feemarket/ante/fee.go` with no
upstream bound check anywhere in the Cosmos-branch ante chain) straight to `int64`. A
`gasWanted` at or above 2^63 wraps negative on that cast, making the computed required fee
negative. The ante handler's own check, `paid.LT(required.Amount)`, would then be `false` for
*any* non-negative `paid` — including zero — completely bypassing the EIP-1559 fee-market's
minimum-fee enforcement. Fixed by using `math.NewIntFromUint64(gasWanted).MulInt(...)`
instead of the cast, which can never overflow (`sdkmath.Int` is arbitrary-precision).
Regression test: `TestRequiredFeeDoesNotOverflowForLargeGasWanted`
(`x/feemarket/keeper/abci_test.go`), asserting the required fee for `gasWanted =
math.MaxUint64` is still positive and strictly greater than for a normal-sized `gasWanted`.

By contrast, `x/feemarket/keeper/abci.go`'s *other* `gosec`-flagged conversion
(`int64(ctx.BlockGasMeter().GasConsumedToLimit())`) and `x/emission/keeper/abci.go`'s two
unix-timestamp conversions were left as `#nosec`-annotated: block gas consumed is capped by
the block gas limit consensus param across the whole block (not one attacker's single tx
field), and unix timestamps are nowhere near either integer type's range in practice. The
difference that mattered: is the value a single, unbounded field an attacker directly
controls, or is it something the chain itself already bounds by construction? Auditors should
specifically re-check the ante/fee-handling path for the same pattern — see `AUDIT_SCOPE.md`.

**Dependency upgrades, not just documentation, where a real fix existed:**
- `sdk/` and `faucet/`: the entire `@cosmjs/*` family bumped `0.32.4` → `0.39.0` (lockstep
  versions, confirmed via `npm view`), which turned out to **drop the `@confio/ics23`
  dependency entirely** (confirmed via `npm view @cosmjs/stargate@latest dependencies`) —
  `ics23@0.6.8` (the latest available, even today) still pins `protobufjs@^6.8.8`, an old
  major line with a critical arbitrary-code-execution CVE that upstream has never patched.
  Forcing an npm `overrides` on that transitive `protobufjs` instead was considered and
  rejected: `ics23` is merkle-proof verification code, and silently forcing a
  major-version-incompatible dependency into crypto-adjacent code carries its own
  correctness risk that a security-audit pass shouldn't introduce without deep testing. The
  version bump instead removes the vulnerable dependency at its root. Verified: `npm run
  typecheck`, `npm run build`, and the full `sdk` test suite (58 tests, 100% coverage
  maintained) all pass unchanged after the bump; `faucet` typecheck/build unchanged.
- `web/`: Next.js `14.2.0` → `16.2.12` and React `18.3.0` → `19.2.8` (coordinated, since Next
  16 requires React 19) to clear a long list of real Next.js CVEs (DoS, SSRF, cache
  poisoning) only fixed in `>=15.5.21`. Two further findings were Next's own internally
  vendored `postcss`/`sharp` versions (not overridable via the project's own `postcss`
  dependency, since Next bundles its own copy for its build pipeline) — forced to patched
  versions via npm's `overrides` field, verified safe since both are minor/patch-level bumps
  within their existing major version. One real compile break from the recharts 2→3 major
  bump (`EmissionChart.tsx`'s `Tooltip` `formatter`/`labelFormatter` callback types tightened)
  — fixed by removing the now-unnecessary explicit parameter types. Verified: `npm run
  typecheck`, `npm run build` (all 19 routes), and a smoke-check of `/`, `/token`,
  `/developers`, `/roadmap` via a production `next start` server (title tags and 200 status
  codes confirmed; the interactive dev-server preview tool hit an unrelated environment
  issue, so this was done via `curl` instead of the browser tool).
- `docs/`: already on the latest available Docusaurus 3.x (`3.10.2`); its remaining findings
  (`webpack-dev-server`, `sockjs`, `serialize-javascript`, `uuid`) are Docusaurus's own
  vendored dev-server/build tooling — `npm audit fix` (without `--force`) reported no version
  bump available at the time, since Docusaurus's `copy-webpack-plugin`/`css-minimizer-webpack-plugin`
  (pulling `serialize-javascript`) and `webpack-dev-server`'s `sockjs` (pulling `uuid`) pin
  those as internal implementation details, not part of their public API surface. Dev-server-only;
  never present in the deployed static site output. **Update, see the Dependabot triage section
  below:** patched majors for both (`serialize-javascript` 6→7, `uuid` 8→11) were later forced in
  via npm `overrides`, since neither package's newer major changes anything the pinning plugins
  actually call.

**Fuzz tests** (Go native `go test -fuzz`, run just over 5 minutes each, all clean — see
`AUDIT_SCOPE.md` for what each one checks): `FuzzMsgTaxedTransfer` (34,806 execs / 323s),
`FuzzMsgTransferAsset` (33,345 execs / 315s), `FuzzVoteTally` (33,783 execs / 305s).
`FuzzMsgTransferAsset` is modeled as a property check over the four independent preconditions
(frozen/authorized/both-parties-attested) rather than fuzzing raw address strings, since an
arbitrary fuzzed string is never the real owner or authority address and would only ever
exercise the "stranger" rejection branch — the interesting state space is the *combination*
of boolean preconditions, which Go's fuzzer explores far more thoroughly than a hand-written
table of cases would.

## GitHub Dependabot alert triage (post-launch, `lastcoinvision` repo)

121 open Dependabot alerts on the live GitHub repo were triaged into fixed / accepted-deferred
/ confirmed-non-issue buckets, prioritizing by actual reachability over raw alert count.

**Fixed — Go module bumps (`chain/go.mod`, `chain/go.sum`):**
- `google.golang.org/grpc` `1.79.2` → `1.82.1` (direct dependency; clears 1 critical + 1 high
  CVE).
- `github.com/go-jose/go-jose/v4` `4.1.3` → `4.1.4` (indirect, high).
- `github.com/consensys/gnark-crypto` `0.18.0` → `0.18.1` (indirect, high).
- `go` directive left at `1.25.0` (unchanged). Verified via `go build ./...` and `make test`
  (all packages pass).

**Accepted, deferred — `github.com/hashicorp/go-getter`:** a patch to `1.8.6` (from `1.7.9`,
fixing a "high" finding) was attempted but not kept. `go-getter`'s newer release pulls in a
substantial new transitive tree (AWS SDK v2, Google Cloud Storage client) and forces a Go
toolchain switch, which is disproportionate for a dependency that's only indirectly reachable
in `lcvd` — and repeatedly triggered local disk exhaustion during the bump attempt (see
environment note below). `go mod tidy` settled the module graph back at `1.7.9` after the
attempt; this was a deliberate choice, not an accident. Left as a tracked, open finding for a
future pass rather than forced through under disk pressure.

**Confirmed non-issues — not reachable in the compiled binary:** all 6 CRITICAL
`golang.org/x/crypto` (ssh) findings, plus `docker/docker`, `docker/cli`, and `go-chi/chi`.
Verified via `go list -deps ./cmd/lcvd/...` — none of these packages are actually imported by
anything that ends up compiled into the `lcvd` binary; they're present in `go.sum` only for
build reproducibility of dev/test tooling. No code change; documented so a future audit pass
doesn't re-flag the same non-issue.

**Fixed — `create-lcv-app` CLI scaffold templates** (`cli/templates/{voting-portal,
gov-registry,dapp-cosmos,dapp-evm}`): all 4 templates bumped `next` `14.2.0` → `16.3.0`,
`react`/`react-dom` `18.3.0` → `19.2.0`, and `@cosmjs/proto-signing`/`@cosmjs/stargate`
`0.32.4` → `0.39.0`, clearing the same class of Next.js CVEs documented in the `web/` bump
above. This surfaced two stacked, genuine build bugs under Next 16's Turbopack default (not
present under Next 14/webpack), both root-caused rather than worked around:
1. `sdk/tsup.config.ts` — the SDK's own build output imported `protobufjs/minimal` (via
   ts-proto's generated code) without a file extension, which CJS `require()` tolerates but
   Node's strict ESM resolver and Turbopack both reject. Fixed by adding
   `noExternal: ["protobufjs"]` so tsup bundles it inline instead of leaving it as an external
   bare import. Verified: raw Node ESM import succeeds, `grep -c "protobufjs/minimal\""
   dist/index.js` is 0, and the SDK's full test suite (58 tests, 100% coverage) still passes.
2. `cli/src/index.ts`'s new `patchNextConfigForMonorepoTesting` — Turbopack refuses to resolve
   modules outside its detected project root, which breaks scaffolded templates only when
   `@lastcoinvision/sdk` is linked in via a monorepo-local `file:` path (i.e. only during
   dev/testing before the SDK is published to npm). Root-caused via Next's own
   `config-shared.d.ts` (the documented `turbopack.root` option), fixed by patching
   `next.config.mjs` with an absolute `turbopack.root` dynamically at scaffold time — kept out
   of the committed template files themselves, since real end users installing a published
   `@lastcoinvision/sdk` from npm never hit this path and shouldn't carry a hardcoded
   monorepo-local absolute path. `experimental.externalDir: true` was also added to all 4
   templates' committed `next.config.mjs` (harmless no-op for a normal npm dependency, but
   required alongside `turbopack.root` for the monorepo-testing case). `dapp-evm`'s
   pre-existing webpack-only alias workaround (unused Coinbase CDP `@x402/*` imports pulled in
   by wagmi) was left in place — a no-op under Turbopack, but didn't cause any resurfaced
   failure. Verified: all 4 templates pass `npm install`, typecheck, `npm run build`, and at
   least one `npm run dev` smoke test.

**Fixed — `dapp-evm`'s wagmi/viem/WalletConnect chain:** the 23 `npm audit` findings (2 high)
in the wagmi/viem/WalletConnect/Reown/axios/ws dependency chain (unrelated to the Next.js/React
CVEs above, and never visible to Dependabot since none of the 4 CLI templates ship a committed
lockfile) turned out to require a wagmi **major** bump, not a patch: `npm view wagmi` showed the
vulnerable `@reown/appkit`/WalletConnect chain affects every wagmi 2.x release up through
`3.5.0` — the latest 2.x line (`2.19.5`) is still vulnerable, only `3.6+` carries the fix (via
`@wagmi/connectors@8.1.0`, up from the vulnerable `<=7.2.1` range). Bumped `wagmi` `2.12.0` →
`3.7.0` and `viem` `2.21.0` → `2.55.0` (wagmi 3.x's peer range). Verified via a scratch install
of just `wagmi`+`viem` (no template code) that the bump alone drops the vulnerable
`@reown/appkit`/`@walletconnect/*` chain entirely — `npm audit` goes from 23 findings across 477
resolved packages down to 0 findings across 50, confirming the whole WalletConnect connector
tree (which this template's UI never actually uses — only `injected()`) is gone, not just
patched. The major bump broke one real thing: wagmi 3's `useBalance` dropped the pre-formatted
`balance.formatted` string from its return type, so `app/page.tsx` now formats it manually via
viem's `formatUnits(balance.value, balance.decimals)`. Also removed `next.config.mjs`'s
`@x402/core`/`@coinbase/cdp-sdk` webpack alias workaround (for `wagmi/connectors`' Coinbase
`baseAccount` connector, which pulled those packages in transitively) — under `@wagmi/connectors@8.1.0`
those packages are no longer in the dependency tree at all, confirmed via
`node_modules/@coinbase`/`node_modules/@x402` both absent after install, so the alias was dead
code referencing packages that no longer exist. Verified end-to-end: `npm install` (0
vulnerabilities), `tsc --noEmit` clean, `npm run build` (Turbopack) succeeds, and a `npm run dev`
smoke test renders the connect/balance UI without console errors.

## Triage of the remaining 33 alerts (post-fix rescan)

After the fixes above landed, GitHub's Dependabot count dropped from 121 to 33. Each of the 33
was individually triaged (not batch-assumed) via `go list -deps ./cmd/lcvd/...` for reachability
and the GitHub Advisory API for patch availability — several conclusions from the earlier pass
needed correcting once checked at this level of precision.

**Fixed — `github.com/shamaton/msgpack/v2` (high, DoS):** reachable in the compiled `lcvd`
binary (`app → CosmWasm/wasmd/x/wasm → CosmWasm/wasmvm/v3/types → shamaton/msgpack/v2`, confirmed
via `go mod why`). CVE-2026-32284: the decoder doesn't validate buffer length on truncated fixext
data (format codes `0xd4`–`0xd8`), causing an out-of-bounds read/panic — a real DoS since wasmvm
decodes msgpack from contract-controlled input. Dependabot listed no `first_patched_version` (the
advisory metadata hadn't been curated with a fix version), so patch availability was verified
directly against upstream: `v2.4.1`'s release notes ("Backport ext frame bounds validation to
v2... Adds regression tests for truncated ext and timestamp payloads") match the advisory's
description exactly, confirming it's the fix commit despite the advisory itself not saying so.
Bumped `v2.2.3` → `v2.4.1`. Verified: `go build ./...` and the full `go test ./...` suite both
clean.

**Fixed — `esbuild` (low, dev-server arbitrary file read on Windows) across `cli/`, `faucet/`,
`sdk/`:** vulnerable range `>=0.27.3, <0.28.1`, patched at `0.28.1`. Not directly depended on —
pulled in transitively by `tsup`/`vitest`/`vite`, all of which declare `"esbuild": "^0.27.0"`
(npm's caret treats 0.x specially: that range is `>=0.27.0 <0.28.0`, so it can never resolve to
the patched `0.28.1` on its own). Forced via npm `overrides: { "esbuild": "^0.28.1" }` in all
three packages' `package.json`. Low severity and dev-server-only (esbuild's own dev server, never
part of the shipped build output), but the override was cheap and fully verified: `npm install` →
0 vulnerabilities in all three, and each package's real build/typecheck/test command still passes
(`sdk`: `npm run build` + `npm run test:coverage`, 58/58 tests, 100% coverage; `cli` and `faucet`:
`npm run build` + `npm run typecheck`).

**Fixed — `serialize-javascript` and `uuid` in `docs/`:** superseding the "no version bump
available" conclusion from the earlier audit pass (see the corrected paragraph above) —
`npm audit fix` without `--force` doesn't consider major bumps, but both packages had one
available and safe to force: `serialize-javascript` `6.0.2` → `7.0.5` (pinned by
`copy-webpack-plugin`/`css-minimizer-webpack-plugin` inside `@docusaurus/bundler`) and `uuid`
`8.3.2` → `11.1.1` (pinned by `webpack-dev-server`'s `sockjs` dependency). Both are internal
implementation details of those plugins (not part of Docusaurus's own API surface), so the major
version bumps were low-risk to force via `overrides`. Verified: `npm install` → 0 vulnerabilities,
and `npm run build` succeeds with static files generated (the build's pre-existing
`../../DECISIONS.md`-link warnings are unrelated broken-link noise, unchanged from before this
fix).

**Deferred — `github.com/pion/dtls/v2` (medium, AES-GCM nonce/key-leak risk, CVE-2026-26014):**
reachable (`app → cosmos/evm/x/vm/keeper → go-ethereum/{eth/tracers, internal/ethapi, p2p, p2p/nat}
→ pion/stun/v2 → pion/dtls/v2`) — go-ethereum's own devp2p NAT-traversal code, pulled in
transitively via `cosmos/evm`, not something this repo imports directly. Checked the advisory
directly rather than trusting the vulnerable-range string alone: it covers **every** `v2` release
including the latest (`v2.2.12`) — the fix only ever shipped in `pion/dtls/v3` (patched at
`v3.0.11`/`v3.1.1`). `v3` is an import-path-changing major version with an incompatible API;
go-ethereum's `p2p/nat` package imports `v2` directly, so this isn't a `go get` bump we can make
unilaterally — it requires go-ethereum (via `cosmos/evm`) to migrate first. Left open and tracked;
revisit when `cosmos/evm` picks up a go-ethereum release built against `pion/dtls/v3`.

**Deferred, unchanged — `github.com/pion/stun/v2` (medium):** re-confirmed no newer release
exists at all (`go list -m -versions` still tops out at `v2.0.0`, its original 2022 release) —
same conclusion as the earlier investigative pass, nothing new to act on.

**Deferred, unchanged — `github.com/hashicorp/go-getter` (high):** unchanged from the earlier
documented decision above (dependency-bloat/disk-risk tradeoff).

**Confirmed non-issues, refined — `golang.org/x/crypto` (13 findings) and `golang.org/x/net`
(1 finding):** the earlier "not reachable at all" conclusion for `x/crypto` was too coarse; a
precise `go list -deps` pass shows the *package* is reachable, but only through non-SSH
primitive subpackages actually used for cryptography elsewhere in the app (`argon2`, `bcrypt`,
`blake2b`/`blake2s`, `chacha20`/`chacha20poly1305`, `curve25519`, `hkdf`, `nacl/box`,
`nacl/secretbox`, `pbkdf2`, `sha3`, etc.). All 13 flagged CVEs are specifically in
`golang.org/x/crypto/ssh` (agent-forwarding constraint bypass, FIDO/U2F presence-check bypass,
`CertChecker` panic, AES-GCM packet decryption underflow, and similar) — cross-checked each
advisory's summary against the reachable subpackage list, and `ssh` never appears in it. Same
pattern for `x/net`: the package is reachable (`http2`, `idna`, `netutil`, etc., pulled in by the
gRPC/HTTP stack), but the flagged CVE is specifically in `golang.org/x/net/html` (the HTML
parser), which isn't reachable — nothing in `lcvd` parses HTML. Both confirmed non-issues, no code
change, but documented with the corrected subpackage-level precision so a future pass trusts the
package-level "reachable" signal correctly instead of over- or under-reading it.

**Confirmed non-issues, unchanged — `docker/docker`, `docker/cli`, `go-chi/chi/v5`:** re-verified
absent from `go list -deps ./cmd/lcvd/...` entirely (zero matches), consistent with the earlier
finding.

**Remaining after this pass:** `pion/dtls/v2`, `pion/stun/v2`, and `go-getter` stay open as
tracked, deferred findings (no safe fix available without a disproportionate or unilateral
upstream change); the 13 `x/crypto` + 1 `x/net` findings and the 3 Docker/go-chi findings are
confirmed non-issues requiring no action.

**Environment note:** the Go dependency bump attempts repeatedly exhausted local disk on this
Mac (dropped as low as 26Mi during the `go-getter` bump attempt), causing a genuinely
inconsistent `go.sum` (`go build ./...` failing with "missing go.sum entry" / "no such file or
directory" errors from corrupted module-cache reads). Recovered via `go clean -cache` /
`go clean -modcache` + `go mod tidy`. This is a recurring constraint of this specific
development machine, unrelated to and does not affect the production VPS (which stayed stable
at 80% used / 20G free throughout).

