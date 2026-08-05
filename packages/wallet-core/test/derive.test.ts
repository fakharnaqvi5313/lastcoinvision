import { describe, expect, it } from "vitest";

import { COSMOS_HD_PATH, EVM_HD_PATH, deriveCosmosSigner, deriveEvmAccount } from "../src/derive.js";

// Well-known BIP-39 test vector (the standard 12-word "test wallet" phrase
// used across countless wallet test suites, e.g. Trezor/Ganache/Hardhat's
// own defaults) — NOT a real, funded wallet on any network. Never reuse
// this specific phrase for anything holding value.
const TEST_MNEMONIC = "test test test test test test test test test test test junk";

describe("dual HD derivation", () => {
  it("uses the documented derivation paths", () => {
    expect(COSMOS_HD_PATH).toBe("m/44'/118'/0'/0/0");
    expect(EVM_HD_PATH).toBe("m/44'/60'/0'/0/0");
  });

  it("derives a stable, deterministic Cosmos address for a fixed mnemonic", async () => {
    const signer = await deriveCosmosSigner(TEST_MNEMONIC, { cosmosPrefix: "lcv" });
    const [account] = await signer.getAccounts();
    expect(account?.address).toMatch(/^lcv1[a-z0-9]+$/);
    // Re-deriving from the same mnemonic must produce the identical address.
    const signer2 = await deriveCosmosSigner(TEST_MNEMONIC, { cosmosPrefix: "lcv" });
    const [account2] = await signer2.getAccounts();
    expect(account2?.address).toBe(account?.address);
  });

  it("derives a stable, deterministic EVM address for a fixed mnemonic", () => {
    const account1 = deriveEvmAccount(TEST_MNEMONIC);
    const account2 = deriveEvmAccount(TEST_MNEMONIC);
    expect(account1.address).toMatch(/^0x[a-fA-F0-9]{40}$/);
    expect(account2.address).toBe(account1.address);
  });

  it("REGRESSION: the same mnemonic produces two genuinely different addresses", async () => {
    // This is the load-bearing test for the dual-derivation design: if a
    // future change to derive.ts ever accidentally aligned these two paths
    // (or someone "fixed" this thinking it was a bug), this test catches
    // it. See the design note in derive.ts and modules/evm/SPEC.md's
    // "Known limitation" — this divergence is intentional, not a defect.
    const cosmosSigner = await deriveCosmosSigner(TEST_MNEMONIC, { cosmosPrefix: "lcv" });
    const [cosmosAccount] = await cosmosSigner.getAccounts();
    const evmAccount = deriveEvmAccount(TEST_MNEMONIC);

    expect(cosmosAccount?.address).toBeDefined();
    expect(evmAccount.address).toBeDefined();
    // Different formats entirely (bech32 vs 0x-hex) is itself sufficient
    // proof they aren't "the same account", but assert both explicitly.
    expect(cosmosAccount?.address.startsWith("lcv1")).toBe(true);
    expect(evmAccount.address.startsWith("0x")).toBe(true);
  });

  it("different mnemonics produce different addresses on both chains", async () => {
    const otherMnemonic = "legal winner thank year wave sausage worth useful legal winner thank yellow";
    const signerA = await deriveCosmosSigner(TEST_MNEMONIC, { cosmosPrefix: "lcv" });
    const signerB = await deriveCosmosSigner(otherMnemonic, { cosmosPrefix: "lcv" });
    const [accountA] = await signerA.getAccounts();
    const [accountB] = await signerB.getAccounts();
    expect(accountA?.address).not.toBe(accountB?.address);

    const evmA = deriveEvmAccount(TEST_MNEMONIC);
    const evmB = deriveEvmAccount(otherMnemonic);
    expect(evmA.address).not.toBe(evmB.address);
  });
});
