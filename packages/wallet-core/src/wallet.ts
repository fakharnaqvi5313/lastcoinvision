import type { DirectSecp256k1HdWallet } from "@cosmjs/proto-signing";
import type { HDAccount } from "viem/accounts";

import { deriveCosmosSigner, deriveEvmAccount } from "./derive.js";
import { DEFAULT_KDF_PARAMS, decryptKeystore, encryptKeystore } from "./keystore.js";
import type { Argon2idParams, EncryptedKeystoreV1 } from "./keystore.js";
import { generateMnemonic } from "./mnemonic.js";
import { getCosmosAddress } from "./signers/cosmosSigner.js";

export interface LcvWalletOptions {
  /** bech32 address prefix, e.g. "lcv". */
  readonly cosmosPrefix: string;
}

/**
 * One mnemonic, two independently-derived signers. See `derive.ts` for why
 * `cosmosAddress` and `evmAddress` are never the same string, and never
 * should be presented as if they were.
 */
export class LcvWallet {
  private constructor(
    private readonly mnemonicValue: string,
    readonly cosmosSigner: DirectSecp256k1HdWallet,
    readonly evmAccount: HDAccount,
  ) {}

  static async fromMnemonic(mnemonic: string, options: LcvWalletOptions): Promise<LcvWallet> {
    const cosmosSigner = await deriveCosmosSigner(mnemonic, options);
    const evmAccount = deriveEvmAccount(mnemonic);
    return new LcvWallet(mnemonic, cosmosSigner, evmAccount);
  }

  /** Generates a brand-new 12-word mnemonic and derives both signers from it. */
  static async create(options: LcvWalletOptions): Promise<LcvWallet> {
    return LcvWallet.fromMnemonic(generateMnemonic(), options);
  }

  static async fromKeystore(
    keystore: EncryptedKeystoreV1,
    password: string,
    options: LcvWalletOptions,
  ): Promise<LcvWallet> {
    const mnemonic = await decryptKeystore(keystore, password); // throws on wrong password/tampering
    return LcvWallet.fromMnemonic(mnemonic, options);
  }

  /**
   * Encrypts this wallet's mnemonic for at-rest storage. The RESULT is safe
   * to write to disk/OS-keychain storage; the mnemonic string itself never
   * should be. Callers are responsible for choosing where the returned
   * JSON-serializable object is persisted — this package has no opinion on
   * platform storage (that's `apps/desktop`'s Rust keyring bridge and
   * `apps/mobile`'s `react-native-keychain`, per the wallet plan).
   */
  async toKeystore(password: string, kdfParams: Argon2idParams = DEFAULT_KDF_PARAMS): Promise<EncryptedKeystoreV1> {
    return encryptKeystore(this.mnemonicValue, password, kdfParams);
  }

  async cosmosAddress(): Promise<string> {
    return getCosmosAddress(this.cosmosSigner);
  }

  /**
   * Returns the raw mnemonic. This exists for exactly one legitimate UI
   * flow: showing a brand-new wallet's recovery phrase to the user ONCE at
   * creation time so they can back it up. Callers must never log this
   * value, send it over any network call, or hold onto it longer than the
   * backup-display screen needs it — `toKeystore()` is the form meant for
   * anything at rest, including in-memory caches.
   */
  revealMnemonic(): string {
    return this.mnemonicValue;
  }

  get evmAddress(): string {
    return this.evmAccount.address;
  }
}
