import { describe, expect, it } from "vitest";

import { assertValidMnemonic, generateMnemonic, isValidMnemonic } from "../src/mnemonic.js";

describe("mnemonic generation and validation", () => {
  it("generates a valid 12-word mnemonic by default", () => {
    const mnemonic = generateMnemonic();
    expect(mnemonic.split(" ")).toHaveLength(12);
    expect(isValidMnemonic(mnemonic)).toBe(true);
  });

  it("generates different mnemonics on each call (real entropy, not a fixture)", () => {
    const a = generateMnemonic();
    const b = generateMnemonic();
    expect(a).not.toBe(b);
  });

  it("supports 24-word generation for import-flow parity with other wallets", () => {
    const mnemonic = generateMnemonic(32);
    expect(mnemonic.split(" ")).toHaveLength(24);
    expect(isValidMnemonic(mnemonic)).toBe(true);
  });

  it("rejects an invalid entropy length", () => {
    expect(() => generateMnemonic(17)).toThrow();
  });

  it("rejects a mnemonic with a bad checksum", () => {
    // Same words as a valid test vector, in the wrong order -- one valid
    // permutation, valid wordlist membership, but wrong checksum.
    const scrambled = "junk test test test test test test test test test test test";
    expect(isValidMnemonic(scrambled)).toBe(false);
    expect(() => assertValidMnemonic(scrambled)).toThrow();
  });

  it("rejects a word not in the BIP-39 wordlist", () => {
    const notAWord = "notaword test test test test test test test test test test junk";
    expect(isValidMnemonic(notAWord)).toBe(false);
  });

  it("accepts a known-good mnemonic", () => {
    const known = "test test test test test test test test test test test junk";
    expect(isValidMnemonic(known)).toBe(true);
    expect(() => assertValidMnemonic(known)).not.toThrow();
  });
});
