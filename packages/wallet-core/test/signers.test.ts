import { describe, expect, it } from "vitest";

import { deriveEvmAccount } from "../src/derive.js";
import { getCosmosAddress } from "../src/signers/cosmosSigner.js";
import { buildEvmWalletClient } from "../src/signers/evmSigner.js";
import { LcvWallet } from "../src/wallet.js";

describe("signers", () => {
  it("buildEvmWalletClient wires a derived account into a viem WalletClient targeting LCV's EVM chain", () => {
    const account = deriveEvmAccount("test test test test test test test test test test test junk");
    const client = buildEvmWalletClient(account, {
      evmChainId: 2611,
      chainName: "LCV Testnet",
      rpcUrl: "http://localhost:8545",
    });
    expect(client.account?.address).toBe(account.address);
    expect(client.chain?.id).toBe(2611);
  });

  it("getCosmosAddress throws if a signer somehow produces no accounts", async () => {
    const emptySigner = { getAccounts: async () => [] } as unknown as Parameters<typeof getCosmosAddress>[0];
    await expect(getCosmosAddress(emptySigner)).rejects.toThrow("produced no accounts");
  });

  it("getCosmosAddress returns the real derived address for a normal signer", async () => {
    const wallet = await LcvWallet.fromMnemonic("test test test test test test test test test test test junk", {
      cosmosPrefix: "lcv",
    });
    const address = await getCosmosAddress(wallet.cosmosSigner);
    expect(address).toBe(await wallet.cosmosAddress());
  });
});
