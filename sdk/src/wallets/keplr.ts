import type { OfflineSigner } from "@cosmjs/proto-signing";
import { SigningStargateClient } from "@cosmjs/stargate";

import { createLcvRegistry } from "../registry.js";

/**
 * Minimal shape of `window.keplr` this SDK depends on, kept local instead
 * of pulling in `@keplr-wallet/types` as a dependency for two functions'
 * worth of typing.
 */
export interface KeplrLike {
  experimentalSuggestChain(chainInfo: KeplrChainInfo): Promise<void>;
  enable(chainId: string): Promise<void>;
  getOfflineSigner(chainId: string): OfflineSigner;
}

export interface KeplrChainInfo {
  chainId: string;
  chainName: string;
  rpc: string;
  rest: string;
  bip44: { coinType: number };
  bech32Config: {
    bech32PrefixAccAddr: string;
    bech32PrefixAccPub: string;
    bech32PrefixValAddr: string;
    bech32PrefixValPub: string;
    bech32PrefixConsAddr: string;
    bech32PrefixConsPub: string;
  };
  currencies: Array<{ coinDenom: string; coinMinimalDenom: string; coinDecimals: number }>;
  feeCurrencies: Array<{
    coinDenom: string;
    coinMinimalDenom: string;
    coinDecimals: number;
    gasPriceStep: { low: number; average: number; high: number };
  }>;
  stakeCurrency: { coinDenom: string; coinMinimalDenom: string; coinDecimals: number };
}

export interface LcvChainInfo {
  chainId: string;
  chainName: string;
  /** Tendermint RPC endpoint, e.g. `http://localhost:26657`. */
  rpcEndpoint: string;
  /** gRPC-gateway REST endpoint, e.g. `http://localhost:1317`. */
  restEndpoint: string;
  bech32Prefix?: string;
  coinDenom?: string;
  coinMinimalDenom?: string;
  coinDecimals?: number;
  gasPriceStep?: { low: number; average: number; high: number };
}

const DEFAULTS = {
  bech32Prefix: "lcv",
  coinDenom: "LCV",
  coinMinimalDenom: "ulcv",
  coinDecimals: 6,
  // Cosmos SDK default (unchanged from Phase A); Keplr-side accounts on
  // this chain use the keyring's default secp256k1, not the eth_secp256k1
  // opt-in `lcvd keys add --algo` path, so they will NOT share an address
  // with a MetaMask account derived from the same mnemonic — see the
  // "Known limitation" in modules/evm/SPEC.md.
  coinType: 118,
  gasPriceStep: { low: 0, average: 0.025, high: 0.05 },
} as const;

/** Builds the `ChainInfo` object Keplr's `experimentalSuggestChain` expects. */
export function buildKeplrChainInfo(info: LcvChainInfo): KeplrChainInfo {
  const bech32Prefix = info.bech32Prefix ?? DEFAULTS.bech32Prefix;
  const coinDenom = info.coinDenom ?? DEFAULTS.coinDenom;
  const coinMinimalDenom = info.coinMinimalDenom ?? DEFAULTS.coinMinimalDenom;
  const coinDecimals = info.coinDecimals ?? DEFAULTS.coinDecimals;
  const gasPriceStep = info.gasPriceStep ?? DEFAULTS.gasPriceStep;

  return {
    chainId: info.chainId,
    chainName: info.chainName,
    rpc: info.rpcEndpoint,
    rest: info.restEndpoint,
    bip44: { coinType: DEFAULTS.coinType },
    bech32Config: {
      bech32PrefixAccAddr: bech32Prefix,
      bech32PrefixAccPub: `${bech32Prefix}pub`,
      bech32PrefixValAddr: `${bech32Prefix}valoper`,
      bech32PrefixValPub: `${bech32Prefix}valoperpub`,
      bech32PrefixConsAddr: `${bech32Prefix}valcons`,
      bech32PrefixConsPub: `${bech32Prefix}valconspub`,
    },
    currencies: [{ coinDenom, coinMinimalDenom, coinDecimals }],
    feeCurrencies: [{ coinDenom, coinMinimalDenom, coinDecimals, gasPriceStep }],
    stakeCurrency: { coinDenom, coinMinimalDenom, coinDecimals },
  };
}

/** Prompts Keplr to add LCV as a known chain, if it doesn't already have it. */
export async function suggestLcvChain(keplr: KeplrLike, info: LcvChainInfo): Promise<void> {
  await keplr.experimentalSuggestChain(buildKeplrChainInfo(info));
}

/** Enables the chain in Keplr and returns its `OfflineSigner` (amino + direct). */
export async function getKeplrOfflineSigner(keplr: KeplrLike, chainId: string): Promise<OfflineSigner> {
  await keplr.enable(chainId);
  return keplr.getOfflineSigner(chainId);
}

/**
 * Full connect flow: suggest the chain, enable it, and return a ready
 * `SigningStargateClient` (registered with every LCV custom-module message
 * type) plus the underlying signer.
 */
export async function connectKeplr(
  keplr: KeplrLike,
  info: LcvChainInfo,
): Promise<{ signer: OfflineSigner; client: SigningStargateClient }> {
  await suggestLcvChain(keplr, info);
  const signer = await getKeplrOfflineSigner(keplr, info.chainId);
  const client = await SigningStargateClient.connectWithSigner(info.rpcEndpoint, signer, {
    registry: createLcvRegistry(),
  });
  return { signer, client };
}
