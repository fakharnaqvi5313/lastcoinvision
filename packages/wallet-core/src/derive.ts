import { stringToPath } from "@cosmjs/crypto";
import { DirectSecp256k1HdWallet } from "@cosmjs/proto-signing";
import type { HDAccount } from "viem/accounts";
import { mnemonicToAccount } from "viem/accounts";

/**
 * One mnemonic, two independently-derived addresses — by design, not by
 * accident. Cosmos-side accounts here use the plain secp256k1 keyring
 * default (coin type 118, same as Keplr's default and `lcvd keys add`
 * without `--algo eth_secp256k1`); EVM-side accounts use the standard
 * Ethereum path (coin type 60, same as MetaMask). These will NEVER produce
 * the same address — see `modules/evm/SPEC.md`'s "Known limitation" and
 * the same note in the SDK's `wallets/keplr.ts`. Every UI built on this
 * package must show both addresses explicitly and never imply they're one
 * account.
 */
export const COSMOS_HD_PATH = "m/44'/118'/0'/0/0";
export const EVM_HD_PATH = "m/44'/60'/0'/0/0";

export interface DeriveOptions {
  /** bech32 address prefix, e.g. "lcv". */
  readonly cosmosPrefix: string;
}

/** Derives the Cosmos-side signer (secp256k1, coin type 118). */
export async function deriveCosmosSigner(
  mnemonic: string,
  options: DeriveOptions,
): Promise<DirectSecp256k1HdWallet> {
  return DirectSecp256k1HdWallet.fromMnemonic(mnemonic, {
    hdPaths: [stringToPath(COSMOS_HD_PATH)],
    prefix: options.cosmosPrefix,
  });
}

/** Derives the EVM-side account (secp256k1/Keccak256, coin type 60). */
export function deriveEvmAccount(mnemonic: string): HDAccount {
  return mnemonicToAccount(mnemonic, { path: EVM_HD_PATH });
}
