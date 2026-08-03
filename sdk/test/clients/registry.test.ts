import { afterEach, describe, expect, it, vi } from "vitest";

import { RegistryClient } from "../../src/clients/registry.js";
import { TypeUrls } from "../../src/registry.js";
import { fakeSigningClient, mockFetchJsonOnce } from "../testUtils.js";

const REST = "http://localhost:1317";

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe("RegistryClient message builders", () => {
  const client = new RegistryClient(REST);

  it("builds MsgSetJurisdictionAuthority", () => {
    const msg = { authority: "lcv1gov", jurisdiction_code: "US-CA", address: "lcv1authority" };
    expect(client.msgSetJurisdictionAuthority(msg)).toEqual({
      typeUrl: TypeUrls.registry.setJurisdictionAuthority,
      value: msg,
    });
  });

  it("builds MsgRegisterAsset", () => {
    const msg = {
      creator: "lcv1authority",
      asset_type: "real-estate",
      jurisdiction_code: "US-CA",
      metadata_hash: "0xabc",
      owner: "did:lcv:lcv1owner",
    };
    expect(client.msgRegisterAsset(msg)).toEqual({ typeUrl: TypeUrls.registry.registerAsset, value: msg });
  });

  it("builds MsgTransferAsset", () => {
    const msg = { creator: "lcv1owner", asset_id: "asset-1", new_owner: "did:lcv:lcv1newowner" };
    expect(client.msgTransferAsset(msg)).toEqual({ typeUrl: TypeUrls.registry.transferAsset, value: msg });
  });

  it("builds MsgFreezeAsset", () => {
    const msg = { creator: "lcv1authority", asset_id: "asset-1", reason: "fraud" };
    expect(client.msgFreezeAsset(msg)).toEqual({ typeUrl: TypeUrls.registry.freezeAsset, value: msg });
  });
});

describe("RegistryClient convenience broadcast methods", () => {
  it("throws when constructed without a signing client", async () => {
    const client = new RegistryClient(REST);
    await expect(
      client.registerAsset("lcv1authority", {
        creator: "lcv1authority",
        asset_type: "real-estate",
        jurisdiction_code: "US-CA",
        metadata_hash: "0xabc",
        owner: "did:lcv:lcv1owner",
      }),
    ).rejects.toThrow(/without a SigningStargateClient/);
  });

  it("registerAsset signs and broadcasts the built message", async () => {
    const { client: signingClient, signAndBroadcast } = fakeSigningClient();
    const client = new RegistryClient(REST, signingClient);
    const msg = {
      creator: "lcv1authority",
      asset_type: "real-estate",
      jurisdiction_code: "US-CA",
      metadata_hash: "0xabc",
      owner: "did:lcv:lcv1owner",
    };

    await client.registerAsset("lcv1authority", msg, "auto", "memo");

    expect(signAndBroadcast).toHaveBeenCalledWith(
      "lcv1authority",
      [{ typeUrl: TypeUrls.registry.registerAsset, value: msg }],
      "auto",
      "memo",
    );
  });

  it("transferAsset signs and broadcasts the built message", async () => {
    const { client: signingClient, signAndBroadcast } = fakeSigningClient();
    const client = new RegistryClient(REST, signingClient);
    const msg = { creator: "lcv1owner", asset_id: "asset-1", new_owner: "did:lcv:lcv1newowner" };

    await client.transferAsset("lcv1owner", msg);

    expect(signAndBroadcast).toHaveBeenCalledWith(
      "lcv1owner",
      [{ typeUrl: TypeUrls.registry.transferAsset, value: msg }],
      "auto",
      "",
    );
  });

  it("freezeAsset signs and broadcasts the built message", async () => {
    const { client: signingClient, signAndBroadcast } = fakeSigningClient();
    const client = new RegistryClient(REST, signingClient);
    const msg = { creator: "lcv1authority", asset_id: "asset-1", reason: "fraud" };

    await client.freezeAsset("lcv1authority", msg);

    expect(signAndBroadcast).toHaveBeenCalledWith(
      "lcv1authority",
      [{ typeUrl: TypeUrls.registry.freezeAsset, value: msg }],
      "auto",
      "",
    );
  });
});

describe("RegistryClient queries", () => {
  const client = new RegistryClient(REST);

  it("params()", async () => {
    const fetchMock = mockFetchJsonOnce({ params: {} });
    await client.params();
    expect(fetchMock).toHaveBeenCalledWith(`${REST}/lastcoinvision/lcv/registry/params`);
  });

  it("assetByID()", async () => {
    const fetchMock = mockFetchJsonOnce({ asset: { id: "asset-1" } });
    await client.assetByID("asset-1");
    expect(fetchMock).toHaveBeenCalledWith(`${REST}/lastcoinvision/lcv/registry/asset/asset-1`);
  });

  it("transferHistory()", async () => {
    const fetchMock = mockFetchJsonOnce({ transfers: [] });
    await client.transferHistory("asset-1");
    expect(fetchMock).toHaveBeenCalledWith(`${REST}/lastcoinvision/lcv/registry/asset/asset-1/history`);
  });

  it("jurisdictionAuthorityOf()", async () => {
    const fetchMock = mockFetchJsonOnce({ address: "lcv1authority" });
    await client.jurisdictionAuthorityOf("US-CA");
    expect(fetchMock).toHaveBeenCalledWith(`${REST}/lastcoinvision/lcv/registry/jurisdiction/US-CA`);
  });

  it("propagates REST errors", async () => {
    mockFetchJsonOnce({ message: "not found" }, false, 404);
    await expect(client.assetByID("missing")).rejects.toThrow(/404/);
  });
});
