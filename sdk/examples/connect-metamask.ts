/**
 * Browser-only example. Not runnable standalone with tsx/node — it takes
 * the injected EIP-1193 provider as a parameter rather than reading
 * `window.ethereum` itself, so wire it up from your dApp like:
 *
 *   import { runWithMetaMask } from "./connect-metamask.js";
 *   if (window.ethereum) await runWithMetaMask(window.ethereum);
 */
import type { EIP1193Provider } from "viem";

import { connectMetaMask, type LcvEvmNetworkInfo } from "../src/index.js";

export async function runWithMetaMask(provider: EIP1193Provider): Promise<void> {
  const network: LcvEvmNetworkInfo = {
    evmChainId: 2611,
    chainName: "LastcoinVision Devnet",
    rpcUrl: "http://localhost:8545",
  };

  const { accounts, publicClient } = await connectMetaMask(provider, network);
  const address = accounts[0];
  if (!address) {
    throw new Error("MetaMask returned no accounts.");
  }

  const balance = await publicClient.getBalance({ address: address as `0x${string}` });
  console.log(`LCV balance for ${address}: ${balance} wei (18-decimal alcv precision)`);
}
