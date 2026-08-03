import { buildLcvViemChain } from "@lastcoinvision/sdk";

const RPC_URL = process.env.NEXT_PUBLIC_LCV_EVM_RPC_URL ?? "http://localhost:8545";
const CHAIN_ID = Number(process.env.NEXT_PUBLIC_LCV_EVM_CHAIN_ID ?? "2611");
const CHAIN_NAME = process.env.NEXT_PUBLIC_LCV_CHAIN_NAME ?? "LastcoinVision Devnet";

export const lcvChain = buildLcvViemChain({ evmChainId: CHAIN_ID, chainName: CHAIN_NAME, rpcUrl: RPC_URL });
