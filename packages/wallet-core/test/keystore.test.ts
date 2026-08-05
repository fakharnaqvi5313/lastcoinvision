import { describe, expect, it } from "vitest";

import { DEFAULT_KDF_PARAMS, decryptKeystore, encryptKeystore } from "../src/keystore.js";
import type { Argon2idParams } from "../src/keystore.js";

// Deliberately weak params, ONLY for fast test iteration -- never use
// anything like this for a real keystore. `encryptKeystore`'s real default
// (`DEFAULT_KDF_PARAMS`) is exercised separately below with a longer
// timeout, since that's the one that actually matters.
const FAST_TEST_PARAMS: Argon2idParams = { t: 1, m: 8, p: 1, dkLen: 32 };

describe("encrypted keystore", () => {
  it("round-trips a secret under the correct password", async () => {
    const secret = "test test test test test test test test test test test junk";
    const keystore = await encryptKeystore(secret, "correct horse battery staple", FAST_TEST_PARAMS);
    const recovered = await decryptKeystore(keystore, "correct horse battery staple");
    expect(recovered).toBe(secret);
  });

  it("fails closed on the wrong password (does not return garbage)", async () => {
    const secret = "some mnemonic phrase here";
    const keystore = await encryptKeystore(secret, "right-password", FAST_TEST_PARAMS);
    await expect(decryptKeystore(keystore, "wrong-password")).rejects.toThrow();
  });

  it("fails closed on tampered ciphertext", async () => {
    const secret = "some mnemonic phrase here";
    const keystore = await encryptKeystore(secret, "a-password", FAST_TEST_PARAMS);
    const tampered = {
      ...keystore,
      // Flip one hex character in the ciphertext -- must fail the Poly1305
      // auth-tag check, not silently decrypt to different plaintext.
      ciphertext: keystore.ciphertext.slice(0, -1) + (keystore.ciphertext.endsWith("0") ? "1" : "0"),
    };
    await expect(decryptKeystore(tampered, "a-password")).rejects.toThrow();
  });

  it("fails closed on a tampered nonce", async () => {
    const secret = "some mnemonic phrase here";
    const keystore = await encryptKeystore(secret, "a-password", FAST_TEST_PARAMS);
    const tampered = {
      ...keystore,
      nonce: keystore.nonce.slice(0, -1) + (keystore.nonce.endsWith("0") ? "1" : "0"),
    };
    await expect(decryptKeystore(tampered, "a-password")).rejects.toThrow();
  });

  it("rejects an unsupported keystore version", async () => {
    const secret = "some mnemonic phrase here";
    const keystore = await encryptKeystore(secret, "a-password", FAST_TEST_PARAMS);
    const badVersion = { ...keystore, version: 2 as unknown as 1 };
    await expect(decryptKeystore(badVersion, "a-password")).rejects.toThrow(/Unsupported keystore version/);
  });

  it("uses a fresh salt and nonce every call, even for the same secret+password", async () => {
    const secret = "same secret every time";
    const a = await encryptKeystore(secret, "same-password", FAST_TEST_PARAMS);
    const b = await encryptKeystore(secret, "same-password", FAST_TEST_PARAMS);
    expect(a.salt).not.toBe(b.salt);
    expect(a.nonce).not.toBe(b.nonce);
    expect(a.ciphertext).not.toBe(b.ciphertext);
    // Both must still independently decrypt correctly.
    expect(await decryptKeystore(a, "same-password")).toBe(secret);
    expect(await decryptKeystore(b, "same-password")).toBe(secret);
  });

  it(
    "round-trips under the REAL production KDF params (DEFAULT_KDF_PARAMS), not just the fast test params",
    async () => {
      const secret = "a real mnemonic would go here for this test only";
      const keystore = await encryptKeystore(secret, "a-real-password");
      expect(keystore.kdfParams).toEqual(DEFAULT_KDF_PARAMS);
      const recovered = await decryptKeystore(keystore, "a-real-password");
      expect(recovered).toBe(secret);
    },
    // 64 MiB argon2id in pure JS legitimately takes several seconds, and
    // longer still under v8 coverage instrumentation -- this is real work,
    // not a hang, so the generous timeout is intentional here.
    60_000,
  );
});
