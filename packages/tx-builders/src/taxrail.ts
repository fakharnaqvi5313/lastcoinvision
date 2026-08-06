import { TaxrailClient } from "@lastcoinvision/sdk";
import { TESTNET_DEFAULT_FEES } from "@lastcoinvision/wallet-core";
import type { StdFee } from "@cosmjs/stargate";

// See identity.ts's comment on why these are derived structurally rather
// than imported by name from sdk's `taxrailTx`/`taxrailQuery` namespaces.
type RegisterTaxRuleMsg = Parameters<TaxrailClient["registerTaxRule"]>[1];
type TaxedTransferMsg = Parameters<TaxrailClient["taxedTransfer"]>[1];
type SpendCollectedTaxMsg = Parameters<TaxrailClient["spendCollectedTax"]>[1];

/** See identity.ts's doc comment -- same pattern for x/taxrail. */
export class TaxrailTxClient {
  constructor(
    private readonly address: string,
    private readonly client: TaxrailClient,
  ) {}

  async registerTaxRule(msg: RegisterTaxRuleMsg, fee: StdFee = TESTNET_DEFAULT_FEES.moduleMsg, memo = "") {
    return this.client.registerTaxRule(this.address, msg, fee, memo);
  }

  async taxedTransfer(msg: TaxedTransferMsg, fee: StdFee = TESTNET_DEFAULT_FEES.moduleMsg, memo = "") {
    return this.client.taxedTransfer(this.address, msg, fee, memo);
  }

  async spendCollectedTax(msg: SpendCollectedTaxMsg, fee: StdFee = TESTNET_DEFAULT_FEES.moduleMsg, memo = "") {
    return this.client.spendCollectedTax(this.address, msg, fee, memo);
  }

  params(): ReturnType<TaxrailClient["params"]> {
    return this.client.params();
  }

  taxRule(jurisdictionCode: string): ReturnType<TaxrailClient["taxRule"]> {
    return this.client.taxRule(jurisdictionCode);
  }

  jurisdictionBalance(jurisdictionCode: string): ReturnType<TaxrailClient["jurisdictionBalance"]> {
    return this.client.jurisdictionBalance(jurisdictionCode);
  }
}
