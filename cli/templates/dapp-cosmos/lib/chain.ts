import type { LcvChainInfo } from "@lastcoinvision/sdk";

export const CHAIN_INFO: LcvChainInfo = {
  chainId: process.env.NEXT_PUBLIC_LCV_CHAIN_ID ?? "lcv-devnet-1",
  chainName: "LastcoinVision Devnet",
  rpcEndpoint: process.env.NEXT_PUBLIC_LCV_RPC_URL ?? "http://localhost:26657",
  restEndpoint: process.env.NEXT_PUBLIC_LCV_REST_URL ?? "http://localhost:1317",
};
