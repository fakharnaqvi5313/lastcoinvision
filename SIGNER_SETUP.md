# Setting Up as a Multisig Signer

You've been asked to be one of the signers on a LastcoinVision multisig account (see
`MULTISIG_CEREMONY_RUNBOOK.md` for why this exists and how it works). This page is everything
you need to do your one step: **generate your own key, on your own device, and send back only
your public key.**

## Before you start — the two things that matter most

1. **Never share your recovery phrase (mnemonic) with anyone — not the person coordinating
   this, not the other signers, not in any chat or email.** The whole point of a multisig is
   that no single person (including you, alone) can move funds. If your mnemonic ever leaves
   your device, that guarantee is broken.
2. **You will send back exactly one thing: your public key.** It's printed as a line starting
   with `pubkey:` when you run the command below. It looks like this and is safe to share:
   `{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"Ag5Ta1GRjwZ1EWwe30GWVt1Kngh/lAbhGZgqKALed03L"}`

## Prerequisite: repository access

You'll need read access to the GitHub repository this key-generation tool lives in. If you
don't have it yet, ask whoever sent you this page to add you as a collaborator first.

## Option A: the desktop wallet app (recommended if you're comfortable running a couple of
terminal commands to build it once)

```bash
git clone https://github.com/fakharnaqvi5313/lastcoinvision.git
cd lastcoinvision
npm install
npm run build
cd apps/desktop
npm run tauri:dev
```

The `npm run build` step (from the repo root, before `cd apps/desktop`) is required — it builds
the shared internal packages the desktop app depends on. Skipping it causes a
`Failed to resolve entry for package "@lastcoinvision/ui-tokens"` error when the app tries to
start.

The first run takes a few minutes (it's compiling a real desktop app, not downloading a
pre-built one — this project doesn't have a packaged installer yet). Once the window opens:

1. Click **Create new wallet**.
2. You'll see a recovery phrase — write it down somewhere safe and private. This is the one
   secret you must never share (see above).
3. On the dashboard, find the **Cosmos address** row and click **Show public key**.
4. Click **Copy**, then send that text back — that's it.

## Option B: the command line

If Option A's build toolchain isn't already on your machine (it needs Rust in addition to what's
below), this is usually faster:

```bash
git clone https://github.com/fakharnaqvi5313/lastcoinvision.git
cd lastcoinvision/chain
go build -o lcvd ./cmd/lcvd    # needs Go 1.25+: https://go.dev/dl/ (the exact patch version is
                                # fetched automatically, just install any recent 1.25.x or newer)
./lcvd keys add treasury-signer --keyring-backend os
```

`--keyring-backend os` stores your key in your operating system's own secure keychain (macOS
Keychain, Windows Credential Manager, or Linux Secret Service) — **don't** omit this flag or use
the default, which stores keys in an unencrypted plain file.

This prints three things: an `address`, a `pubkey` (send this one back), and a mnemonic phrase
(never share this one). If the build fails because of a missing C compiler, on macOS run
`xcode-select --install` first; on Linux, install `build-essential` (Debian/Ubuntu) or your
distro's equivalent, then retry.

## What happens next

Once everyone's public key is collected, the multisig account gets constructed from just those
public keys — nothing about this step ever needs your private key or mnemonic again. You'll keep
using your own key (via whichever tool you used above) whenever a transaction from the multisig
needs your signature.
