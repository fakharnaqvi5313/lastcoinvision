import { afterEach, describe, expect, it, vi } from "vitest";

import { IdentityClient } from "../../src/clients/identity.js";
import { TypeUrls } from "../../src/registry.js";
import { fakeSigningClient, mockFetchJsonOnce } from "../testUtils.js";

const REST = "http://localhost:1317";

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe("IdentityClient message builders", () => {
  const client = new IdentityClient(REST);

  it("builds MsgCreateDID", () => {
    const msg = { creator: "lcv1abc", verification_method: [], service: [] };
    expect(client.msgCreateDID(msg)).toEqual({ typeUrl: TypeUrls.identity.createDID, value: msg });
  });

  it("builds MsgUpdateDID", () => {
    const msg = { creator: "lcv1abc", verification_method: [], service: [] };
    expect(client.msgUpdateDID(msg)).toEqual({ typeUrl: TypeUrls.identity.updateDID, value: msg });
  });

  it("builds MsgDeactivateDID", () => {
    const msg = { creator: "lcv1abc" };
    expect(client.msgDeactivateDID(msg)).toEqual({ typeUrl: TypeUrls.identity.deactivateDID, value: msg });
  });

  it("builds MsgAddIssuer", () => {
    const msg = { authority: "lcv1gov", issuer_address: "lcv1issuer", name: "DMV" };
    expect(client.msgAddIssuer(msg)).toEqual({ typeUrl: TypeUrls.identity.addIssuer, value: msg });
  });

  it("builds MsgRemoveIssuer", () => {
    const msg = { authority: "lcv1gov", issuer_address: "lcv1issuer" };
    expect(client.msgRemoveIssuer(msg)).toEqual({ typeUrl: TypeUrls.identity.removeIssuer, value: msg });
  });

  it("builds MsgIssueCredential", () => {
    const msg = { issuer: "lcv1issuer", subject: "lcv1subject", credential_type: "kyc", credential_hash: "0xabc" };
    expect(client.msgIssueCredential(msg)).toEqual({ typeUrl: TypeUrls.identity.issueCredential, value: msg });
  });

  it("builds MsgRevokeCredential", () => {
    const msg = { issuer: "lcv1issuer", credential_id: "cred-1" };
    expect(client.msgRevokeCredential(msg)).toEqual({ typeUrl: TypeUrls.identity.revokeCredential, value: msg });
  });

  it("builds MsgPresentCredential", () => {
    const msg = { presenter: "lcv1presenter", credential_id: "cred-1", proof: new Uint8Array([1, 2, 3]) };
    expect(client.msgPresentCredential(msg)).toEqual({ typeUrl: TypeUrls.identity.presentCredential, value: msg });
  });
});

describe("IdentityClient convenience broadcast methods", () => {
  it("throws when constructed without a signing client", async () => {
    const client = new IdentityClient(REST);
    await expect(client.createDID("lcv1abc", { creator: "lcv1abc", verification_method: [], service: [] })).rejects.toThrow(
      /without a SigningStargateClient/,
    );
  });

  it("createDID signs and broadcasts the built message", async () => {
    const { client: signingClient, signAndBroadcast } = fakeSigningClient();
    const client = new IdentityClient(REST, signingClient);
    const msg = { creator: "lcv1abc", verification_method: [], service: [] };

    const result = await client.createDID("lcv1abc", msg, "auto", "memo");

    expect(signAndBroadcast).toHaveBeenCalledWith(
      "lcv1abc",
      [{ typeUrl: TypeUrls.identity.createDID, value: msg }],
      "auto",
      "memo",
    );
    expect(result).toEqual({ code: 0, transactionHash: "DEADBEEF" });
  });

  it("issueCredential signs and broadcasts the built message", async () => {
    const { client: signingClient, signAndBroadcast } = fakeSigningClient();
    const client = new IdentityClient(REST, signingClient);
    const msg = { issuer: "lcv1issuer", subject: "lcv1subject", credential_type: "kyc", credential_hash: "0xabc" };

    await client.issueCredential("lcv1issuer", msg);

    expect(signAndBroadcast).toHaveBeenCalledWith(
      "lcv1issuer",
      [{ typeUrl: TypeUrls.identity.issueCredential, value: msg }],
      "auto",
      "",
    );
  });

  it("presentCredential signs and broadcasts the built message", async () => {
    const { client: signingClient, signAndBroadcast } = fakeSigningClient();
    const client = new IdentityClient(REST, signingClient);
    const msg = { presenter: "lcv1presenter", credential_id: "cred-1", proof: new Uint8Array() };

    await client.presentCredential("lcv1presenter", msg);

    expect(signAndBroadcast).toHaveBeenCalledWith(
      "lcv1presenter",
      [{ typeUrl: TypeUrls.identity.presentCredential, value: msg }],
      "auto",
      "",
    );
  });
});

describe("IdentityClient queries", () => {
  const client = new IdentityClient(REST);

  it("params()", async () => {
    const fetchMock = mockFetchJsonOnce({ params: { attested_credential_type: "kyc" } });
    const result = await client.params();
    expect(fetchMock).toHaveBeenCalledWith(`${REST}/lastcoinvision/lcv/identity/params`);
    expect(result).toEqual({ params: { attested_credential_type: "kyc" } });
  });

  it("didByAddress() percent-encodes the address", async () => {
    const fetchMock = mockFetchJsonOnce({ did: { id: "did:lcv:lcv1abc" } });
    await client.didByAddress("lcv1abc/weird");
    expect(fetchMock).toHaveBeenCalledWith(`${REST}/lastcoinvision/lcv/identity/did/lcv1abc%2Fweird`);
  });

  it("credentialStatus()", async () => {
    const fetchMock = mockFetchJsonOnce({ revoked: false });
    await client.credentialStatus("cred-1");
    expect(fetchMock).toHaveBeenCalledWith(`${REST}/lastcoinvision/lcv/identity/credential/cred-1`);
  });

  it("issuerList()", async () => {
    const fetchMock = mockFetchJsonOnce({ issuers: [] });
    await client.issuerList();
    expect(fetchMock).toHaveBeenCalledWith(`${REST}/lastcoinvision/lcv/identity/issuers`);
  });

  it("tierOf()", async () => {
    const fetchMock = mockFetchJsonOnce({ tier: "TIER_VERIFIED" });
    await client.tierOf("lcv1abc");
    expect(fetchMock).toHaveBeenCalledWith(`${REST}/lastcoinvision/lcv/identity/tier/lcv1abc`);
  });

  it("propagates REST errors", async () => {
    mockFetchJsonOnce({ message: "not found" }, false, 404);
    await expect(client.tierOf("lcv1missing")).rejects.toThrow(/404/);
  });
});
