import { describe, expect, it } from "vitest";

import type { Argon2idParams } from "../src/keystore.js";
import { LcvWallet } from "../src/wallet.js";

const OPTIONS = { cosmosPrefix: "lcv" };
// Fast KDF params for test speed only -- see test/keystore.test.ts for why
// (and for a separate test that exercises the real default params).
const FAST_TEST_PARAMS: Argon2idParams = { t: 1, m: 8, p: 1, dkLen: 32 };

describe("LcvWallet facade", () => {
  it("create() produces a wallet with both a Cosmos and an EVM address", async () => {
    const wallet = await LcvWallet.create(OPTIONS);
    const cosmosAddress = await wallet.cosmosAddress();
    expect(cosmosAddress).toMatch(/^lcv1[a-z0-9]+$/);
    expect(wallet.evmAddress).toMatch(/^0x[a-fA-F0-9]{40}$/);
    expect(cosmosAddress).not.toBe(wallet.evmAddress);
  });

  it("fromMnemonic() is deterministic", async () => {
    const mnemonic = "test test test test test test test test test test test junk";
    const walletA = await LcvWallet.fromMnemonic(mnemonic, OPTIONS);
    const walletB = await LcvWallet.fromMnemonic(mnemonic, OPTIONS);
    expect(await walletA.cosmosAddress()).toBe(await walletB.cosmosAddress());
    expect(walletA.evmAddress).toBe(walletB.evmAddress);
  });

  it("round-trips through an encrypted keystore", async () => {
    const original = await LcvWallet.create(OPTIONS);
    const originalCosmosAddress = await original.cosmosAddress();
    const originalEvmAddress = original.evmAddress;

    const keystore = await original.toKeystore("a-strong-password", FAST_TEST_PARAMS);
    const restored = await LcvWallet.fromKeystore(keystore, "a-strong-password", OPTIONS);

    expect(await restored.cosmosAddress()).toBe(originalCosmosAddress);
    expect(restored.evmAddress).toBe(originalEvmAddress);
  });

  it("refuses to restore from a keystore with the wrong password", async () => {
    const original = await LcvWallet.create(OPTIONS);
    const keystore = await original.toKeystore("correct-password", FAST_TEST_PARAMS);
    await expect(LcvWallet.fromKeystore(keystore, "wrong-password", OPTIONS)).rejects.toThrow();
  });
});
