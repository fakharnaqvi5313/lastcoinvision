import { describe, expect, it, vi } from "vitest";

import type { RegistryClient } from "@lastcoinvision/sdk";
import { TESTNET_DEFAULT_FEES } from "@lastcoinvision/wallet-core";

import { RegistryTxClient } from "../src/registry.js";

function fakeRegistryClient() {
  return {
    registerAsset: vi.fn().mockResolvedValue({ code: 0, transactionHash: "AAA" }),
    transferAsset: vi.fn().mockResolvedValue({ code: 0, transactionHash: "BBB" }),
    freezeAsset: vi.fn().mockResolvedValue({ code: 0, transactionHash: "CCC" }),
    params: vi.fn().mockResolvedValue({ params: {} }),
    assetByID: vi.fn().mockResolvedValue({ asset: {} }),
    transferHistory: vi.fn().mockResolvedValue({ transfers: [] }),
    jurisdictionAuthorityOf: vi.fn().mockResolvedValue({ authority: "" }),
  } as unknown as RegistryClient;
}

const ADDRESS = "lcv1abc";

describe("RegistryTxClient", () => {
  it("injects the signer address and defaults to the fixed module-msg fee, never auto", async () => {
    const inner = fakeRegistryClient();
    const client = new RegistryTxClient(ADDRESS, inner);
    const msg = {
      creator: ADDRESS,
      asset_type: "deed",
      jurisdiction_code: "US-CA",
      metadata_hash: "0xabc",
      owner: ADDRESS,
    };

    await client.registerAsset(msg);

    expect(inner.registerAsset).toHaveBeenCalledWith(ADDRESS, msg, TESTNET_DEFAULT_FEES.moduleMsg, "");
  });

  it("passes through an explicit fee and memo override", async () => {
    const inner = fakeRegistryClient();
    const client = new RegistryTxClient(ADDRESS, inner);
    const msg = { creator: ADDRESS, asset_id: "asset-1", new_owner: "lcv1new" };
    const customFee = { amount: [{ amount: "9999", denom: "ulcv" }], gas: "500000" };

    await client.transferAsset(msg, customFee, "a memo");

    expect(inner.transferAsset).toHaveBeenCalledWith(ADDRESS, msg, customFee, "a memo");
  });

  it("freezeAsset also defaults to the fixed fee", async () => {
    const inner = fakeRegistryClient();
    const client = new RegistryTxClient(ADDRESS, inner);
    const msg = { creator: ADDRESS, asset_id: "asset-1", reason: "compliance hold" };

    await client.freezeAsset(msg);

    expect(inner.freezeAsset).toHaveBeenCalledWith(ADDRESS, msg, TESTNET_DEFAULT_FEES.moduleMsg, "");
  });

  it("read-only queries pass through unmodified", async () => {
    const inner = fakeRegistryClient();
    const client = new RegistryTxClient(ADDRESS, inner);

    await client.params();
    await client.assetByID("asset-1");
    await client.transferHistory("asset-1");
    await client.jurisdictionAuthorityOf("US-CA");

    expect(inner.params).toHaveBeenCalled();
    expect(inner.assetByID).toHaveBeenCalledWith("asset-1");
    expect(inner.transferHistory).toHaveBeenCalledWith("asset-1");
    expect(inner.jurisdictionAuthorityOf).toHaveBeenCalledWith("US-CA");
  });
});
