import { describe, expect, it } from "vitest";

import { fixedFee, TESTNET_DEFAULT_FEES } from "../src/fees.js";

describe("fixed-fee helpers", () => {
  it("builds an explicit StdFee, never the auto-estimation sentinel", () => {
    const fee = fixedFee("1000", "ulcv", "200000");
    expect(fee).toEqual({ amount: [{ amount: "1000", denom: "ulcv" }], gas: "200000" });
  });

  it("exposes non-empty testnet defaults for both simple and module messages", () => {
    expect(TESTNET_DEFAULT_FEES.simple.gas).toBeTruthy();
    expect(TESTNET_DEFAULT_FEES.moduleMsg.gas).toBeTruthy();
    expect(TESTNET_DEFAULT_FEES.simple).not.toEqual(TESTNET_DEFAULT_FEES.moduleMsg);
  });
});
