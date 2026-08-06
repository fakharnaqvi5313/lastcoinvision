# App Store Compliance Checklist

Phase 5 of the wallet build (see [`DECISIONS.md`](DECISIONS.md) Phases 0-4 for `packages/wallet-core`,
`packages/tx-builders`, `apps/desktop`, and `apps/mobile`). This is a hard gate before any public
App Store / Play Store listing — nothing here should be read as "submitted" or "live" until every
box below is checked against the real store consoles, not this document.

**Status (2026-08-07):** the parts of this phase that don't require a legal business identity or
a payment are done (see §1). Everything else in §2-§4 requires action only the account owner can
take — an AI agent cannot enroll a developer program, verify a business identity, or pay an
account fee on someone's behalf. This document exists so that work is itemized and ready the
moment those steps happen, not so it can claim they're done.

## §1 — Done, no account needed

- [x] **Privacy policy page** — [`web/app/privacy/page.tsx`](web/app/privacy/page.tsx), linked
  from the site footer. States the non-custodial model plainly: keys never leave the device,
  no account/sign-up, no analytics/crash-reporting SDK in the wallet apps today, and what the
  RPC/REST network calls do expose (IP address to the node operator, same as any wallet).
  Both Apple's App Privacy questionnaire and Google Play's Data Safety form should be filled out
  consistent with this page — see §3.
- [x] **Bundle identifiers set** — `com.lastcoinvision.wallet` for both iOS and Android, already
  in [`apps/mobile/app.json`](apps/mobile/app.json) since Phase 4.
- [x] **Testnet-only framing exists in-app** — both apps default to `lcv-devnet`/`lcv-testnet-1`
  (`apps/desktop/src/lib/network.ts`, `apps/mobile/src/lib/network.ts`); no mainnet chain id
  exists yet for either app to point at (see Phase 6 gate).
- [x] **`apps/mobile/eas.json` build profiles written** — `development`/`preview`/`production`,
  with preview and production setting `EXPO_PUBLIC_LCV_NETWORK=testnet` so a real distributed
  build defaults to the public testnet instead of `localhost` (which no tester's device can
  reach) on first launch; `src/lib/network.ts` reads it, falling back to `devnet` for local
  `npx expo start`. Verified: `tsc`/`vitest` still clean, and `EXPO_PUBLIC_LCV_NETWORK=testnet
  npx expo export` still bundles correctly. This is config only — no `eas build` has been run
  (that needs an Expo/EAS account login, and for an installable iOS build, the Apple Developer
  account from §2).

## §2 — Blocked on the account owner (cannot be done by an agent)

These are account-creation and payment actions on real developer-program consoles, gated to
whoever controls the LastcoinVision business identity and payment method:

- [ ] **Apple Developer Program organization enrollment.** Requires a real legal business
  entity, a D-U-N-S number (Apple verifies against Dun & Bradstreet — can take 1-2 weeks if one
  doesn't already exist), and the $99/year program fee. Apple's crypto/wallet review (guideline
  3.1.5(b)) is historically slower and more scrutinized than typical app review — budget extra
  round-trips.
- [ ] **Google Play Console organization account.** Requires business verification and the
  one-time $25 registration fee. Play's Financial Services policy requires declaring the app as
  a "wallet" type and linking a privacy policy covering the custody model (§1's page is ready
  for this).
- [ ] **App Store Connect / Play Console app records created**, once both accounts above exist —
  bundle ID `com.lastcoinvision.wallet` registered in both, app name/subtitle reserved.

## §3 — Ready to fill in once §2 unblocks it

Draft copy, prepared now so submission isn't blocked on writing it later. Update the numbers/
dates before actually submitting — don't paste this verbatim without checking it's still current.

**App name:** LCV Wallet
**Subtitle / short description:** Non-custodial wallet for LastcoinVision (testnet)
**Full description (draft):**
> LCV Wallet is a non-custodial wallet for the LastcoinVision blockchain (testnet). Your
> recovery phrase is generated and encrypted on your device and never leaves it — LastcoinVision
> has no server that can see or recover your keys. Create or import a wallet, view your Cosmos
> and EVM addresses, and send transactions on LastcoinVision's public testnet. This app currently
> supports testnet only; testnet tokens have no monetary value.

**Screenshots:** must be captured from a real build once §2 unblocks a distributable one (see
§4's EAS note) — every screenshot must show the testnet network selector or otherwise make clear
this is not mainnet, per this phase's testnet-only framing requirement.

**Apple App Privacy / Google Play Data Safety forms:** answer "no" to collecting any personal
data category the app doesn't collect (see §1's privacy page for the exhaustive list of what is
and isn't collected/sent). Do not check boxes defensively — over-declaring collection that isn't
happening is its own compliance problem.

**Category:** Finance (both stores).

## §4 — Distribution mechanics (TestFlight / Play Internal Testing)

- [x] **`eas.json` build profiles** — done, see §1.
- [ ] **iOS build path.** This dev environment's system Ruby cannot build CocoaPods (see
  [`apps/mobile/README.md`](apps/mobile/README.md) and the 2026-08-07 `DECISIONS.md` entry) —
  there is no local `pod install`/Xcode archive path today. The realistic route once an Apple
  Developer account exists is an **EAS Build** (Expo's cloud build service) targeting a real
  bundle identifier, which sidesteps the local CocoaPods gap entirely. `eas.json` is ready;
  actually running `eas build` needs an Expo/EAS account login (an agent won't create one — see
  §2's reasoning) and, for anything beyond internal builds, the Apple Developer Program
  membership from §2 (EAS still needs Apple's signing credentials to produce an
  App-Store-installable build).
- [ ] **Android build path.** `expo build`/EAS Build for Android doesn't have the CocoaPods
  problem — should work once there's a Play Console account to upload to.
- [ ] **TestFlight internal testing track** — needs §2's Apple account plus a signed build.
- [ ] **Play Internal Testing track** — needs §2's Play Console account plus a signed build.
- [ ] Once either testing track has real testers, do the device-restore test called for in
  Phase 4's still-open item: reinstall, restore from mnemonic, confirm the same two addresses
  reappear — on a real device this time, not just the headless script.

## What this phase is not

Not a claim that the wallet is on either store, not a timeline commitment (Apple's org
verification alone can take 1-2 weeks and is outside anyone's direct control), and not a signal
to start distributing builds outside these official channels. See
[`DECISIONS.md`](DECISIONS.md)'s Phase 6 entry (once written) for the separate mainnet-readiness
gate — this phase stays testnet-only throughout.
