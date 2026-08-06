# LCV Desktop Wallet (Tauri)

## Architecture

The frontend (`src/`, React + Vite, running in Tauri's native OS webview) imports
`@lastcoinvision/wallet-core` and `@lastcoinvision/tx-builders` directly and does all key
derivation, encryption, and transaction signing in JS. This is a deliberate change from the
original wallet plan's framing ("Rust shell means no bundled Node/Chromium in the process that
touches decrypted keys") — worth being explicit about the actual security model:

- `wallet-core`'s dependencies (`@cosmjs/*`, `viem`, `@noble/hashes`, `@noble/ciphers`) are all
  browser-compatible by design (the same code already runs in Keplr/MetaMask browser-extension
  contexts) — no Node.js runtime is required to run them.
- Tauri's webview, unlike Electron's renderer, has **no Node.js integration by default** — it
  cannot touch the filesystem, spawn processes, or do anything outside the browser sandbox
  except through the specific commands defined in `src-tauri/src/lib.rs`. Loading no remote
  content (the whole app is the local bundle) removes the "compromised web content gets OS
  access" risk Electron warnings are usually about.
- What Rust *does* own: the OS-native secret store (macOS Keychain / Windows Credential
  Manager / Linux Secret Service, via the `keyring` crate) through three commands —
  `keychain_set`/`keychain_get`/`keychain_delete`. These only ever handle the **encrypted**
  keystore JSON string produced by `wallet-core`'s `encryptKeystore` — never a raw mnemonic or
  private key.
- The decrypted `LcvWallet` (and hence the mnemonic, transiently, via `revealMnemonic()` during
  the one-time backup-display screen) lives only in the webview's JS heap for the session.
  Locking the wallet just drops that in-memory state — nothing sensitive was ever written to
  disk outside the OS keychain, so there's nothing to securely wipe.

This is a real tradeoff, not the original plan's ideal (which would keep decrypted key material
out of the webview's JS entirely via a full Rust reimplementation of `wallet-core` using crates
like `cosmrs`). That would eliminate any theoretical webview-JS exposure window during an active
signing operation, at the cost of a second, separately-maintained implementation of the same
crypto/derivation logic, duplicating (not reusing) `wallet-core`'s already-tested code. Revisit
if a future security review calls for it specifically.

## Running it

```bash
npm install                 # from the repo root, once
cd apps/desktop
npm run tauri:dev           # starts Vite + the Tauri window together
```

The default network is `lcv-devnet` (`src/lib/network.ts`) — start one with `make devnet` from
the repo root first. Switch to `lcv-testnet-1` from the in-app network dropdown; EVM actions are
disabled there until a public EVM JSON-RPC endpoint for the testnet is confirmed and added to
`NETWORKS.testnet.evmRpcUrl`.

## What's here (Phase 3 scope)

- Create a new wallet (generates a mnemonic, shows it once for backup, encrypts + stores it).
- Import an existing wallet from a mnemonic.
- Unlock with a password; "forget this device" (removes the local encrypted copy only).
- Dashboard: both addresses (Cosmos and EVM, always shown separately — see
  `wallet-core`'s dual-derivation design), balances, and a plain send on each chain.

Not in scope yet: transaction history, the custom-module actions `tx-builders` wraps
(identity/registry/voting/taxrail), and the app-icon/bundle-identifier polish needed before any
real distribution (Phase 5).
