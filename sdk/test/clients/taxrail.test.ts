import { afterEach, describe, expect, it, vi } from "vitest";

import { TaxrailClient } from "../../src/clients/taxrail.js";
import { TypeUrls } from "../../src/registry.js";
import { fakeSigningClient, mockFetchJsonOnce } from "../testUtils.js";

const REST = "http://localhost:1317";

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe("TaxrailClient message builders", () => {
  const client = new TaxrailClient(REST);

  it("builds MsgRegisterTaxRule", () => {
    const msg = { authority: "lcv1gov", jurisdiction_code: "US-CA", rate: "0.075000000000000000" };
    expect(client.msgRegisterTaxRule(msg)).toEqual({ typeUrl: TypeUrls.taxrail.registerTaxRule, value: msg });
  });

  it("builds MsgTaxedTransfer", () => {
    const msg = {
      sender: "lcv1sender",
      recipient: "lcv1recipient",
      jurisdiction_code: "US-CA",
      amount: { denom: "ulcv", amount: "1000000" },
    };
    expect(client.msgTaxedTransfer(msg)).toEqual({ typeUrl: TypeUrls.taxrail.taxedTransfer, value: msg });
  });

  it("builds MsgSpendCollectedTax", () => {
    const msg = {
      authority: "lcv1gov",
      jurisdiction_code: "US-CA",
      recipient: "lcv1treasury",
      amount: { denom: "ulcv", amount: "500000" },
      reason: "infrastructure",
    };
    expect(client.msgSpendCollectedTax(msg)).toEqual({ typeUrl: TypeUrls.taxrail.spendCollectedTax, value: msg });
  });
});

describe("TaxrailClient convenience broadcast methods", () => {
  it("throws when constructed without a signing client", async () => {
    const client = new TaxrailClient(REST);
    await expect(
      client.taxedTransfer("lcv1sender", {
        sender: "lcv1sender",
        recipient: "lcv1recipient",
        jurisdiction_code: "US-CA",
        amount: { denom: "ulcv", amount: "1000000" },
      }),
    ).rejects.toThrow(/without a SigningStargateClient/);
  });

  it("registerTaxRule signs and broadcasts the built message", async () => {
    const { client: signingClient, signAndBroadcast } = fakeSigningClient();
    const client = new TaxrailClient(REST, signingClient);
    const msg = { authority: "lcv1gov", jurisdiction_code: "US-CA", rate: "0.075000000000000000" };

    await client.registerTaxRule("lcv1gov", msg, "auto", "memo");

    expect(signAndBroadcast).toHaveBeenCalledWith(
      "lcv1gov",
      [{ typeUrl: TypeUrls.taxrail.registerTaxRule, value: msg }],
      "auto",
      "memo",
    );
  });

  it("taxedTransfer signs and broadcasts the built message", async () => {
    const { client: signingClient, signAndBroadcast } = fakeSigningClient();
    const client = new TaxrailClient(REST, signingClient);
    const msg = {
      sender: "lcv1sender",
      recipient: "lcv1recipient",
      jurisdiction_code: "US-CA",
      amount: { denom: "ulcv", amount: "1000000" },
    };

    await client.taxedTransfer("lcv1sender", msg);

    expect(signAndBroadcast).toHaveBeenCalledWith(
      "lcv1sender",
      [{ typeUrl: TypeUrls.taxrail.taxedTransfer, value: msg }],
      "auto",
      "",
    );
  });

  it("spendCollectedTax signs and broadcasts the built message", async () => {
    const { client: signingClient, signAndBroadcast } = fakeSigningClient();
    const client = new TaxrailClient(REST, signingClient);
    const msg = {
      authority: "lcv1gov",
      jurisdiction_code: "US-CA",
      recipient: "lcv1treasury",
      amount: { denom: "ulcv", amount: "500000" },
      reason: "infrastructure",
    };

    await client.spendCollectedTax("lcv1gov", msg);

    expect(signAndBroadcast).toHaveBeenCalledWith(
      "lcv1gov",
      [{ typeUrl: TypeUrls.taxrail.spendCollectedTax, value: msg }],
      "auto",
      "",
    );
  });
});

describe("TaxrailClient queries", () => {
  const client = new TaxrailClient(REST);

  it("params()", async () => {
    const fetchMock = mockFetchJsonOnce({ params: {} });
    await client.params();
    expect(fetchMock).toHaveBeenCalledWith(`${REST}/lastcoinvision/lcv/taxrail/params`);
  });

  it("taxRule()", async () => {
    const fetchMock = mockFetchJsonOnce({ rule: { jurisdiction_code: "US-CA", rate: "0.075" } });
    await client.taxRule("US-CA");
    expect(fetchMock).toHaveBeenCalledWith(`${REST}/lastcoinvision/lcv/taxrail/rule/US-CA`);
  });

  it("jurisdictionBalance()", async () => {
    const fetchMock = mockFetchJsonOnce({ balance: { denom: "ulcv", amount: "0" } });
    await client.jurisdictionBalance("US-CA");
    expect(fetchMock).toHaveBeenCalledWith(`${REST}/lastcoinvision/lcv/taxrail/balance/US-CA`);
  });

  it("propagates REST errors", async () => {
    mockFetchJsonOnce({ message: "not found" }, false, 404);
    await expect(client.taxRule("missing")).rejects.toThrow(/404/);
  });
});
