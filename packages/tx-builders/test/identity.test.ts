import { describe, expect, it, vi } from "vitest";

import type { IdentityClient } from "@lastcoinvision/sdk";
import { TESTNET_DEFAULT_FEES } from "@lastcoinvision/wallet-core";

import { IdentityTxClient } from "../src/identity.js";

function fakeIdentityClient() {
  return {
    createDID: vi.fn().mockResolvedValue({ code: 0, transactionHash: "AAA" }),
    issueCredential: vi.fn().mockResolvedValue({ code: 0, transactionHash: "BBB" }),
    presentCredential: vi.fn().mockResolvedValue({ code: 0, transactionHash: "CCC" }),
    params: vi.fn().mockResolvedValue({ params: {} }),
    didByAddress: vi.fn().mockResolvedValue({ document: {} }),
    credentialStatus: vi.fn().mockResolvedValue({ status: "active" }),
    issuerList: vi.fn().mockResolvedValue({ issuers: [] }),
    tierOf: vi.fn().mockResolvedValue({ tier: 1 }),
  } as unknown as IdentityClient;
}

const ADDRESS = "lcv1abc";

describe("IdentityTxClient", () => {
  it("injects the signer address and defaults to the fixed module-msg fee, never auto", async () => {
    const inner = fakeIdentityClient();
    const client = new IdentityTxClient(ADDRESS, inner);
    const msg = { creator: ADDRESS, verification_method: [], service: [] };

    await client.createDID(msg);

    expect(inner.createDID).toHaveBeenCalledWith(ADDRESS, msg, TESTNET_DEFAULT_FEES.moduleMsg, "");
  });

  it("passes through an explicit fee and memo override", async () => {
    const inner = fakeIdentityClient();
    const client = new IdentityTxClient(ADDRESS, inner);
    const msg = { issuer: ADDRESS, subject: "lcv1subject", credential_type: "kyc", credential_hash: "0xabc" };
    const customFee = { amount: [{ amount: "9999", denom: "ulcv" }], gas: "500000" };

    await client.issueCredential(msg, customFee, "a memo");

    expect(inner.issueCredential).toHaveBeenCalledWith(ADDRESS, msg, customFee, "a memo");
  });

  it("presentCredential also defaults to the fixed fee", async () => {
    const inner = fakeIdentityClient();
    const client = new IdentityTxClient(ADDRESS, inner);
    const msg = { presenter: ADDRESS, credential_id: "cred-1", proof: new Uint8Array([1, 2, 3]) };

    await client.presentCredential(msg);

    expect(inner.presentCredential).toHaveBeenCalledWith(ADDRESS, msg, TESTNET_DEFAULT_FEES.moduleMsg, "");
  });

  it("defaults read-only queries' address argument to the wrapped signer's own address", async () => {
    const inner = fakeIdentityClient();
    const client = new IdentityTxClient(ADDRESS, inner);

    await client.didByAddress();
    await client.tierOf();

    expect(inner.didByAddress).toHaveBeenCalledWith(ADDRESS);
    expect(inner.tierOf).toHaveBeenCalledWith(ADDRESS);
  });

  it("read-only queries pass through unmodified", async () => {
    const inner = fakeIdentityClient();
    const client = new IdentityTxClient(ADDRESS, inner);

    await client.params();
    await client.credentialStatus("cred-1");
    await client.issuerList();

    expect(inner.params).toHaveBeenCalled();
    expect(inner.credentialStatus).toHaveBeenCalledWith("cred-1");
    expect(inner.issuerList).toHaveBeenCalled();
  });
});
