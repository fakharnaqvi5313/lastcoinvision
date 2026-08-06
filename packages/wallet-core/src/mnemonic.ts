import { Bip39, EnglishMnemonic, Random } from "@cosmjs/crypto";

/**
 * 12-word (128-bit entropy) mnemonics are the default across this wallet —
 * matches Keplr/MetaMask's own default and is what most recovery-phrase UX
 * (and users) expect. 24-word support is exposed for import flows only,
 * since some other wallets default to it.
 */
const DEFAULT_ENTROPY_BYTES = 16; // 128 bits -> 12 words

/**
 * Generates a new mnemonic from cryptographically secure entropy
 * (`@cosmjs/crypto`'s `Random`, backed by `crypto.getRandomValues` — never
 * `Math.random` or any non-CSPRNG source).
 */
export function generateMnemonic(entropyBytes: number = DEFAULT_ENTROPY_BYTES): string {
  if (![16, 20, 24, 28, 32].includes(entropyBytes)) {
    throw new Error("entropyBytes must be one of 16, 20, 24, 28, 32 (12/15/18/21/24-word mnemonics)");
  }
  const entropy = Random.getBytes(entropyBytes);
  return Bip39.encode(entropy).toString();
}

/**
 * Validates a mnemonic's wordlist membership and checksum. Returns `false`
 * rather than throwing, so callers (e.g. an import-flow text field) can
 * drive UI validation state without a try/catch on every keystroke.
 */
export function isValidMnemonic(mnemonic: string): boolean {
  try {
    // Validation is the constructor's side effect -- discarding the
    // instance is intentional, we only care whether it throws.
    new EnglishMnemonic(mnemonic.trim().toLowerCase());
    return true;
  } catch {
    return false;
  }
}

/**
 * Throws with a specific reason if invalid; use this at the point a
 * mnemonic is about to be used (import confirm, keystore creation), not on
 * every keystroke — `isValidMnemonic` is the right check for that.
 */
export function assertValidMnemonic(mnemonic: string): void {
  const normalized = mnemonic.trim().toLowerCase();
  new EnglishMnemonic(normalized); // throws with a specific reason on bad checksum/wordlist
}
