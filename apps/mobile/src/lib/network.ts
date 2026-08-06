export interface NetworkConfig {
  readonly name: string;
  readonly cosmosChainId: string;
  readonly cosmosPrefix: string;
  readonly rpcEndpoint: string;
  readonly restEndpoint: string;
  readonly evmChainId: number;
  /** null where no public EVM JSON-RPC endpoint is documented/confirmed yet -- EVM features disable rather than guess a URL. */
  readonly evmRpcUrl: string | null;
}

// See apps/desktop/src/lib/network.ts -- same reasoning, duplicated rather
// than shared because each app's "what's the right localhost host" answer
// differs (desktop: localhost; iOS simulator: localhost works too, but a
// physical device or Android emulator needs the host machine's LAN IP or
// 10.0.2.2 respectively -- see README.md for how to override this).
export const NETWORKS: Record<"devnet" | "testnet", NetworkConfig> = {
  devnet: {
    name: "lcv-devnet (local)",
    cosmosChainId: "lcv-devnet",
    cosmosPrefix: "lcv",
    rpcEndpoint: "http://localhost:26657",
    restEndpoint: "http://localhost:1317",
    evmChainId: 2611,
    evmRpcUrl: "http://localhost:8545",
  },
  testnet: {
    name: "lcv-testnet-1",
    cosmosChainId: "lcv-testnet-1",
    cosmosPrefix: "lcv",
    rpcEndpoint: "https://rpc.lastcoinvision.com",
    restEndpoint: "https://api.lastcoinvision.com",
    evmChainId: 2611,
    evmRpcUrl: null,
  },
};

export const DEFAULT_NETWORK: keyof typeof NETWORKS = "devnet";
