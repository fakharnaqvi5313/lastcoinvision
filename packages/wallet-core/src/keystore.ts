import { fromHex, fromUtf8, toHex, toUtf8 } from "@cosmjs/encoding";
import { Random } from "@cosmjs/crypto";
import { xchacha20poly1305 } from "@noble/ciphers/chacha.js";
import { argon2idAsync } from "@noble/hashes/argon2.js";

/**
 * Encrypted-at-rest container for a mnemonic (or, for imported raw keys,
 * any UTF-8 secret string). Deliberately NOT built on `@cosmjs/crypto`'s
 * `Argon2id`/`Xchacha20poly1305Ietf` — both are marked `@deprecated`
 * ("will likely be removed") in the installed `@cosmjs/crypto` version.
 * Calling `@noble/hashes`/`@noble/ciphers` directly (the same underlying,
 * actively-maintained libraries cosmjs's wrapper itself delegates to) gets
 * the identical primitives without depending on an API cosmjs may delete.
 *
 * This is the ONLY new crypto surface in this package — the Cosmos-side
 * signer (`derive.ts`) uses CosmJS's own `DirectSecp256k1HdWallet`
 * unmodified, including its own (separate, cosmjs-owned) serialize/
 * deserialize if a consumer prefers that format instead of this one. This
 * container exists because viem has no equivalent for the EVM-side key,
 * and using one format for both halves keeps there being exactly one
 * encryption scheme to review, not two.
 */

export interface Argon2idParams {
  readonly t: number;
  readonly m: number;
  readonly p: number;
  readonly dkLen: number;
}

/**
 * Roughly libsodium's "moderate" Argon2id preset area — deliberately
 * heavier than OWASP's bare minimum since this protects a wallet mnemonic,
 * not a login password. Mobile performance (Phase 4) may need its own
 * tuning pass; tracked there, not assumed solved here. `encryptKeystore`
 * accepts an override (tests use a much lighter one) — `decryptKeystore`
 * never does, since it must use whatever params are stored *in* the
 * keystore it's opening, not whatever this module's current default is.
 * That distinction matters: if this default is ever strengthened later,
 * old keystores encrypted under the old params must keep decrypting
 * correctly, not silently derive the wrong key.
 */
export const DEFAULT_KDF_PARAMS: Argon2idParams = {
  t: 3,
  m: 65536, // 64 MiB, in KiB
  p: 4,
  dkLen: 32, // matches xchacha20poly1305's 32-byte key requirement
};

const SALT_LENGTH = 16;

export interface EncryptedKeystoreV1 {
  readonly version: 1;
  readonly kdf: "argon2id";
  readonly kdfParams: Argon2idParams;
  /** hex-encoded */
  readonly salt: string;
  /** hex-encoded */
  readonly nonce: string;
  /** hex-encoded; includes the Poly1305 auth tag (appended by the cipher itself) */
  readonly ciphertext: string;
}

async function deriveKey(password: string, salt: Uint8Array, params: Argon2idParams): Promise<Uint8Array> {
  return argon2idAsync(password, salt, params);
}

/**
 * Encrypts `plaintext` (a mnemonic, or an imported raw private key string)
 * under `password`. A fresh salt and nonce are generated every call — never
 * reuse a nonce with a derived key, which is why both are regenerated here
 * rather than accepted as parameters. `kdfParams` defaults to
 * `DEFAULT_KDF_PARAMS`; the only reason to override it is a deliberately
 * lighter cost for automated tests, never to weaken a real keystore.
 */
export async function encryptKeystore(
  plaintext: string,
  password: string,
  kdfParams: Argon2idParams = DEFAULT_KDF_PARAMS,
): Promise<EncryptedKeystoreV1> {
  const salt = Random.getBytes(SALT_LENGTH);
  const key = await deriveKey(password, salt, kdfParams);
  const nonce = Random.getBytes(xchacha20poly1305.nonceLength);
  const cipher = xchacha20poly1305(key, nonce);
  const ciphertext = cipher.encrypt(toUtf8(plaintext));
  return {
    version: 1,
    kdf: "argon2id",
    kdfParams,
    salt: toHex(salt),
    nonce: toHex(nonce),
    ciphertext: toHex(ciphertext),
  };
}

/**
 * Decrypts a keystore produced by `encryptKeystore`. Fails closed: a wrong
 * password or any tampering with the ciphertext/nonce/salt causes the
 * Poly1305 tag check inside `cipher.decrypt` to throw — there is no code
 * path here that returns garbage plaintext instead of throwing. Always
 * uses the `kdfParams` stored in `keystore` itself, never this module's
 * current default (see `DEFAULT_KDF_PARAMS`'s doc comment).
 */
export async function decryptKeystore(keystore: EncryptedKeystoreV1, password: string): Promise<string> {
  if (keystore.version !== 1 || keystore.kdf !== "argon2id") {
    throw new Error(`Unsupported keystore version/kdf: ${keystore.version}/${keystore.kdf}`);
  }
  const salt = fromHex(keystore.salt);
  const key = await deriveKey(password, salt, keystore.kdfParams);
  const nonce = fromHex(keystore.nonce);
  const cipher = xchacha20poly1305(key, nonce);
  const plaintext = cipher.decrypt(fromHex(keystore.ciphertext)); // throws on wrong password or tampering
  return fromUtf8(plaintext);
}
