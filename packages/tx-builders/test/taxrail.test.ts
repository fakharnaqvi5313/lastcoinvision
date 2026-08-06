import { describe, expect, it, vi } from "vitest";

import type { TaxrailClient } from "@lastcoinvision/sdk";
import { TESTNET_DEFAULT_FEES } from "@lastcoinvision/wallet-core";

import { TaxrailTxClient } from "../src/taxrail.js";

function fakeTaxrailClient() {
  return {
    registerTaxRule: vi.fn().mockResolvedValue({ code: 0, transactionHash: "AAA" }),
    taxedTransfer: vi.fn().mockResolvedValue({ code: 0, transactionHash: "BBB" }),
    spendCollectedTax: vi.fn().mockResolvedValue({ code: 0, transactionHash: "CCC" }),
    params: vi.fn().mockResolvedValue({ params: {} }),
    taxRule: vi.fn().mockResolvedValue({ rule: {} }),
    jurisdictionBalance: vi.fn().mockResolvedValue({ balance: {} }),
  } as unknown as TaxrailClient;
}

const ADDRESS = "lcv1abc";

describe("TaxrailTxClient", () => {
  it("injects the signer address and defaults to the fixed module-msg fee, never auto", async () => {
    const inner = fakeTaxrailClient();
    const client = new TaxrailTxClient(ADDRESS, inner);
    const msg = { authority: ADDRESS, jurisdiction_code: "US-CA", rate: "0.05" };

    await client.registerTaxRule(msg);

    expect(inner.registerTaxRule).toHaveBeenCalledWith(ADDRESS, msg, TESTNET_DEFAULT_FEES.moduleMsg, "");
  });

  it("passes through an explicit fee and memo override", async () => {
    const inner = fakeTaxrailClient();
    const client = new TaxrailTxClient(ADDRESS, inner);
    const msg = {
      sender: ADDRESS,
      recipient: "lcv1recipient",
      jurisdiction_code: "US-CA",
      amount: { denom: "ulcv", amount: "1000000" },
    };
    const customFee = { amount: [{ amount: "9999", denom: "ulcv" }], gas: "500000" };

    await client.taxedTransfer(msg, customFee, "a memo");

    expect(inner.taxedTransfer).toHaveBeenCalledWith(ADDRESS, msg, customFee, "a memo");
  });

  it("spendCollectedTax also defaults to the fixed fee", async () => {
    const inner = fakeTaxrailClient();
    const client = new TaxrailTxClient(ADDRESS, inner);
    const msg = {
      authority: ADDRESS,
      jurisdiction_code: "US-CA",
      recipient: "lcv1recipient",
      amount: { denom: "ulcv", amount: "1000" },
      reason: "budget allocation",
    };

    await client.spendCollectedTax(msg);

    expect(inner.spendCollectedTax).toHaveBeenCalledWith(ADDRESS, msg, TESTNET_DEFAULT_FEES.moduleMsg, "");
  });

  it("read-only queries pass through unmodified", async () => {
    const inner = fakeTaxrailClient();
    const client = new TaxrailTxClient(ADDRESS, inner);

    await client.params();
    await client.taxRule("US-CA");
    await client.jurisdictionBalance("US-CA");

    expect(inner.params).toHaveBeenCalled();
    expect(inner.taxRule).toHaveBeenCalledWith("US-CA");
    expect(inner.jurisdictionBalance).toHaveBeenCalledWith("US-CA");
  });
});
