import { describe, expect, it } from "vitest";

import { DEFAULT_NETWORK, NETWORKS } from "./network.js";

describe("network config", () => {
  it("devnet points at localhost on the standard lcvd ports", () => {
    const { devnet } = NETWORKS;
    expect(devnet.rpcEndpoint).toBe("http://localhost:26657");
    expect(devnet.restEndpoint).toBe("http://localhost:1317");
    expect(devnet.evmRpcUrl).toBe("http://localhost:8545");
    expect(devnet.cosmosPrefix).toBe("lcv");
  });

  it("both networks share the same (placeholder) EVM chain id, per modules/evm/SPEC.md", () => {
    expect(NETWORKS.devnet.evmChainId).toBe(2611);
    expect(NETWORKS.testnet.evmChainId).toBe(2611);
  });

  it("testnet has no EVM RPC URL guessed -- disables EVM actions rather than pointing at an unconfirmed endpoint", () => {
    expect(NETWORKS.testnet.evmRpcUrl).toBeNull();
  });

  it("defaults to devnet for local development", () => {
    expect(DEFAULT_NETWORK).toBe("devnet");
    expect(NETWORKS[DEFAULT_NETWORK]).toBe(NETWORKS.devnet);
  });

  it("cosmos chain ids are distinct between networks (never point both at the same chain by accident)", () => {
    expect(NETWORKS.devnet.cosmosChainId).not.toBe(NETWORKS.testnet.cosmosChainId);
  });
});
