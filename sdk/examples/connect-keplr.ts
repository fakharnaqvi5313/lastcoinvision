/**
 * Browser-only example. Not runnable standalone with tsx/node — it takes
 * the injected Keplr provider as a parameter rather than reading
 * `window.keplr` itself, so wire it up from your dApp like:
 *
 *   import { runWithKeplr } from "./connect-keplr.js";
 *   if (window.keplr) await runWithKeplr(window.keplr, "lcv-devnet-1");
 */
import { connectKeplr, IdentityClient, type KeplrLike, type LcvChainInfo } from "../src/index.js";

export async function runWithKeplr(keplr: KeplrLike, chainId: string): Promise<void> {
  const chainInfo: LcvChainInfo = {
    chainId,
    chainName: "LastcoinVision Devnet",
    rpcEndpoint: "http://localhost:26657",
    restEndpoint: "http://localhost:1317",
  };

  const { signer, client } = await connectKeplr(keplr, chainInfo);
  const [account] = await signer.getAccounts();
  if (!account) {
    throw new Error("Keplr returned no accounts for this chain.");
  }

  const identity = new IdentityClient(chainInfo.restEndpoint, client);
  const tier = await identity.tierOf(account.address);
  console.log(`Account tier for ${account.address}:`, tier);
}
