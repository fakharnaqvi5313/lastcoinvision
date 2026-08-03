import { fromBech32, fromHex, toBech32 } from "@cosmjs/encoding";

/**
 * Accepts either a bech32 address (`lcv1...`) or a 0x-hex address (the same
 * account, as MetaMask would display it, for eth_secp256k1 keys — see
 * modules/evm/SPEC.md) and returns the bech32 form the bank-send message
 * needs. Throws with a client-safe message on anything malformed.
 */
export function normalizeAddress(input: string, prefix: string): string {
  const trimmed = input.trim();

  if (trimmed.startsWith("0x") || trimmed.startsWith("0X")) {
    const bytes = fromHex(trimmed.slice(2));
    if (bytes.length !== 20) {
      throw new Error("Invalid 0x address: expected 20 bytes.");
    }
    return toBech32(prefix, bytes);
  }

  const { prefix: gotPrefix, data } = fromBech32(trimmed);
  if (gotPrefix !== prefix) {
    throw new Error(`Invalid address prefix: expected "${prefix}", got "${gotPrefix}".`);
  }
  if (data.length !== 20) {
    throw new Error("Invalid bech32 address: expected a 20-byte payload.");
  }
  return trimmed;
}
