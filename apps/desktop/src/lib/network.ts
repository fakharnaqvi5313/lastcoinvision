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

// EVM chain id 2611 is a documented placeholder shared by devnet and
// testnet (modules/evm/SPEC.md) -- it must be replaced with a real
// ethereum-lists-registered id before mainnet (see the wallet plan's
// Phase 6 gate). Nothing here should be pointed at real value yet.
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
    // No public EVM JSON-RPC endpoint for lcv-testnet-1 is documented
    // anywhere in this repo (checked modules/evm/SPEC.md, DECISIONS.md) --
    // the public testnet work exposed Cosmos RPC/REST, not necessarily the
    // EVM port. Leaving this null rather than guessing a plausible-looking
    // URL; EVM actions are disabled on testnet in the UI until this is
    // confirmed and filled in.
    evmRpcUrl: null,
  },
};

export const DEFAULT_NETWORK: keyof typeof NETWORKS = "devnet";
