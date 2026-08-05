import type { DirectSecp256k1HdWallet } from "@cosmjs/proto-signing";

/**
 * `DirectSecp256k1HdWallet` already implements CosmJS's `OfflineDirectSigner`
 * interface natively — this file doesn't wrap or reimplement anything, it
 * just documents the contract: this is the exact object type
 * `SigningStargateClient.connectWithSigner` and every method on
 * `sdk/src/clients/{identity,registry,voting,taxrail}.ts` already expect,
 * so those files work with a wallet-core-derived signer completely
 * unmodified — no adapter layer needed.
 */
export type CosmosSigner = DirectSecp256k1HdWallet;

/** First (and, in this wallet, only) account's bech32 address. */
export async function getCosmosAddress(signer: CosmosSigner): Promise<string> {
  const [account] = await signer.getAccounts();
  if (!account) {
    throw new Error("Cosmos signer produced no accounts");
  }
  return account.address;
}
