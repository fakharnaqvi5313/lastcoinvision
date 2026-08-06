# LCV Mobile Wallet (Expo / React Native)

## Architecture

Same design as [`apps/desktop`](../desktop/README.md): the app imports
`@lastcoinvision/wallet-core` and `@lastcoinvision/tx-builders` directly and does all key
derivation, encryption, and signing in JS, on top of Expo/React Native rather than Tauri's
webview.

This deviates from the original wallet plan in two ways, both driven by the same constraint —
this Mac's system Ruby (2.6.10) cannot build native gem extensions against the installed Xcode
SDK, so `pod install` (required by any bare React Native / Expo prebuild iOS build) is not
available in this environment:

- **Expo Go workflow, not bare/dev-client.** Every native capability the app needs is provided
  by a first-party Expo Go-compatible module instead of a community package that requires native
  linking:
  - `expo-secure-store` instead of `react-native-keychain` for at-rest encrypted-keystore
    storage (still backed by iOS Keychain / Android Keystore under the hood).
  - `expo-crypto`'s `getRandomValues` instead of `react-native-get-random-values` for the
    `crypto.getRandomValues` polyfill — same Web Crypto API signature, no native linking.
  - `expo-clipboard` for copy-to-clipboard on the dashboard.
  - Plain `buffer` (pure JS) for the global `Buffer` polyfill `@cosmjs/*`/viem/`@noble/*` expect.

  This means the app runs and can be tested entirely inside plain Expo Go — no `pod install`,
  no Xcode project, no native build step. Revisit if a future phase needs a capability Expo Go
  doesn't expose (e.g. hardware-backed biometric gating beyond what `expo-secure-store`'s
  `requireAuthentication` option offers).
- **`src/lib/polyfills.ts`, imported first in `index.ts`.** Hermes (React Native's JS engine)
  has neither `crypto.getRandomValues` nor a global `Buffer` — both of which `@cosmjs/*`, viem,
  and `@noble/hashes`/`@noble/ciphers` assume exist, since they're written for browser-like
  environments. This file closes both gaps before anything imports `wallet-core` transitively.

## Running it

```bash
npm install                 # from the repo root, once
cd apps/mobile
npx expo start               # scan the QR code in Expo Go, or press i/a for a simulator/emulator
```

The default network is `lcv-devnet` (`src/lib/network.ts`) — start one with `make devnet` from
the repo root first. A distributed build (EAS, see below) defaults to `lcv-testnet-1` instead,
since a real device can't reach `localhost`.

## Building for TestFlight / Play Internal Testing

`eas.json` has `development`/`preview`/`production` build profiles ready. Actually running them
(`eas build`) needs an Expo/EAS account login and, for an installable iOS build, an Apple
Developer Program membership — neither of which this repo's tooling sets up for you. See
[`STORE_COMPLIANCE.md`](../../STORE_COMPLIANCE.md) at the repo root for the full compliance
checklist and what's still outstanding.

## What's here (Phase 4 scope)

Same feature set as the desktop app: create/import wallet, unlock/forget device, dashboard with
both addresses (Cosmos and EVM, always shown separately) and a plain send on each chain.

## Verification status and known gaps

- `npx tsc --noEmit` and `npx vitest run` (network config tests) pass cleanly.
- `scripts/verify-session-flow.mjs` exercises the exact sequence `src/lib/session.tsx` drives
  (create wallet → both addresses differ → keystore round-trip → `LcvTxClients.connect()` →
  a live query) against the real local devnet, under plain Node — this is the same script and
  same purpose as `apps/desktop/scripts/verify-session-flow.mjs`.
- `npx expo export --platform ios` and `--platform android` both succeed (1400+ modules bundled
  to Hermes bytecode each), confirming the full import graph — including `wallet-core`,
  `tx-builders`, and every crypto dependency they pull in — resolves and compiles correctly
  under Metro/Hermes. This caught a real bug during development: Metro's resolver does not map a
  `.js` import specifier to a `.ts` file the way `tsc`'s `bundler`/`nodenext` module resolution
  does, so every relative import across this app must be extensionless.
- **Not yet verified: an actual on-device/simulator run of the UI.** This environment's iOS
  Simulator control tool is currently down (repeated crashes; the tool itself reports further
  retries won't help and a fresh panel needs to be reopened). Metro bundling success and the
  headless session-flow script give strong evidence the logic is correct, but neither exercises
  the polyfills' actual behavior inside a running Hermes VM, nor `expo-secure-store`'s real
  Keychain/Keystore integration, nor the rendered UI itself. Do a real Expo Go run (`npx expo
  start`, scan/simulator) before treating this phase as fully verified.
- Not yet done: biometric unlock gating (`expo-secure-store`'s `requireAuthentication` option),
  a device-restore test (reinstall, restore from mnemonic, confirm same addresses reappear), and
  transaction history / the `tx-builders` custom-module actions (identity/registry/voting/
  taxrail) — same "not in scope yet" list as desktop's Phase 3.
