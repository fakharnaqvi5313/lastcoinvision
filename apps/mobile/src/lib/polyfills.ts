/**
 * Must be imported FIRST, before anything that touches wallet-core
 * (index.ts does this). React Native's Hermes engine has neither
 * `crypto.getRandomValues` nor a global `Buffer` -- both of which
 * @cosmjs/*, viem, and @noble/hashes/@noble/ciphers expect to exist,
 * since they're written to run in a browser-like environment (the same
 * assumption that lets them run in the Tauri desktop app's webview
 * unmodified -- see apps/desktop/README.md). Hermes is close to but not
 * exactly that environment; this file closes the two gaps.
 *
 * Deliberately using `expo-crypto` (an Expo Go-compatible, first-party
 * module) instead of the more commonly-recommended
 * `react-native-get-random-values`, which requires native module linking
 * that Expo Go doesn't support -- see apps/mobile/README.md for why this
 * matters for how this app can be run/tested.
 */
import { Buffer } from "buffer";
import { getRandomValues as expoCryptoGetRandomValues } from "expo-crypto";

const g = globalThis as unknown as { Buffer?: typeof Buffer; crypto?: Crypto };

if (typeof g.Buffer === "undefined") {
  g.Buffer = Buffer;
}

if (!g.crypto || typeof g.crypto.getRandomValues !== "function") {
  g.crypto = {
    getRandomValues: expoCryptoGetRandomValues,
  } as Crypto;
}
