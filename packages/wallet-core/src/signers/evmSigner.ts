import { buildLcvViemChain } from "@lastcoinvision/sdk";
import type { LcvEvmNetworkInfo } from "@lastcoinvision/sdk";
import { createWalletClient, http } from "viem";
import type { HDAccount, WalletClient } from "viem";

/**
 * Builds a viem `WalletClient` from a wallet-core-derived EVM account.
 * Reuses `buildLcvViemChain` from `@lastcoinvision/sdk` unmodified — chain
 * metadata (native currency, RPC URL) doesn't care whether the account
 * came from a browser extension or a local encrypted keystore. This
 * replaces `sdk/src/wallets/metamask.ts`'s `custom(provider)` transport
 * (which assumes `window.ethereum`) with viem's plain `http(rpcUrl)`
 * transport plus a locally-held signing account — no extension involved.
 */
export function buildEvmWalletClient(account: HDAccount, info: LcvEvmNetworkInfo): WalletClient {
  const chain = buildLcvViemChain(info);
  return createWalletClient({ account, chain, transport: http(info.rpcUrl) });
}
