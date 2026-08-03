import { createPublicClient, createWalletClient, custom, defineChain } from "viem";
import type { Chain, EIP1193Provider, PublicClient, WalletClient } from "viem";

export interface LcvEvmNetworkInfo {
  /** EIP-155 chain id, e.g. `2611` (see `app.EVMChainID`, modules/evm/SPEC.md). */
  evmChainId: number;
  chainName: string;
  /** JSON-RPC endpoint, e.g. `http://localhost:8545`. */
  rpcUrl: string;
  blockExplorerUrl?: string;
}

export function toHexChainId(evmChainId: number): `0x${string}` {
  return `0x${evmChainId.toString(16)}`;
}

/**
 * Builds a viem `Chain` for LCV's EVM. Native currency is always 18
 * decimals — that's fixed by the EVM/MetaMask side of the precisebank
 * bridge (`app.EVMExtendedDenom`, "alcv"), independent of the 6-decimal
 * `ulcv` bank denom Cosmos-side balances use.
 */
export function buildLcvViemChain(info: LcvEvmNetworkInfo): Chain {
  return defineChain({
    id: info.evmChainId,
    name: info.chainName,
    nativeCurrency: { name: "LCV", symbol: "LCV", decimals: 18 },
    rpcUrls: { default: { http: [info.rpcUrl] } },
    blockExplorers: info.blockExplorerUrl
      ? { default: { name: `${info.chainName} Explorer`, url: info.blockExplorerUrl } }
      : undefined,
  });
}

/** Prompts MetaMask (or any EIP-1193 wallet) to add/switch to the LCV EVM network. */
export async function addLcvNetworkToMetaMask(provider: EIP1193Provider, info: LcvEvmNetworkInfo): Promise<void> {
  await provider.request({
    method: "wallet_addEthereumChain",
    params: [
      {
        chainId: toHexChainId(info.evmChainId),
        chainName: info.chainName,
        nativeCurrency: { name: "LCV", symbol: "LCV", decimals: 18 },
        rpcUrls: [info.rpcUrl],
        blockExplorerUrls: info.blockExplorerUrl ? [info.blockExplorerUrl] : undefined,
      },
    ],
  });
}

/** Requests account access and returns the connected 0x addresses. */
export async function requestMetaMaskAccounts(provider: EIP1193Provider): Promise<string[]> {
  const accounts = await provider.request({ method: "eth_requestAccounts" });
  return accounts as string[];
}

/**
 * Full connect flow: add/switch the network, request accounts, and return
 * viem clients ready for reads (`publicClient`) and signed writes
 * (`walletClient`, e.g. for the Phase D reference contracts).
 */
export async function connectMetaMask(
  provider: EIP1193Provider,
  info: LcvEvmNetworkInfo,
): Promise<{ accounts: string[]; chain: Chain; publicClient: PublicClient; walletClient: WalletClient }> {
  await addLcvNetworkToMetaMask(provider, info);
  const accounts = await requestMetaMaskAccounts(provider);
  const chain = buildLcvViemChain(info);
  const transport = custom(provider);
  const publicClient = createPublicClient({ chain, transport });
  const walletClient = createWalletClient({ chain, transport });
  return { accounts, chain, publicClient, walletClient };
}
