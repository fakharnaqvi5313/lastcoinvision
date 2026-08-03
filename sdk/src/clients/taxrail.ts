import type { EncodeObject } from "@cosmjs/proto-signing";
import type { SigningStargateClient, StdFee } from "@cosmjs/stargate";

import { RestClient, pathSegment } from "../rest.js";
import { TypeUrls } from "../registry.js";
import type {
  MsgRegisterTaxRule,
  MsgSpendCollectedTax,
  MsgTaxedTransfer,
} from "../proto/lcv/taxrail/tx.js";
import type {
  QueryJurisdictionBalanceResponse,
  QueryParamsResponse,
  QueryTaxRuleResponse,
} from "../proto/lcv/taxrail/query.js";

const REST_PREFIX = "/lastcoinvision/lcv/taxrail";

/**
 * Typed client for `x/taxrail`: per-jurisdiction tax rules, taxed
 * transfers, and collected-tax spending. See modules/taxrail/SPEC.md.
 */
export class TaxrailClient {
  private readonly rest: RestClient;

  constructor(
    restEndpoint: string,
    private readonly signingClient?: SigningStargateClient,
  ) {
    this.rest = new RestClient(restEndpoint);
  }

  // ---- message builders (pure, no network) ----

  msgRegisterTaxRule(value: MsgRegisterTaxRule): EncodeObject {
    return { typeUrl: TypeUrls.taxrail.registerTaxRule, value };
  }

  msgTaxedTransfer(value: MsgTaxedTransfer): EncodeObject {
    return { typeUrl: TypeUrls.taxrail.taxedTransfer, value };
  }

  msgSpendCollectedTax(value: MsgSpendCollectedTax): EncodeObject {
    return { typeUrl: TypeUrls.taxrail.spendCollectedTax, value };
  }

  // ---- convenience: build + sign + broadcast ----

  private requireSigner(): SigningStargateClient {
    if (!this.signingClient) {
      throw new Error("TaxrailClient was constructed without a SigningStargateClient; pass one to broadcast transactions.");
    }
    return this.signingClient;
  }

  async registerTaxRule(signer: string, msg: MsgRegisterTaxRule, fee: StdFee | "auto" | number = "auto", memo = "") {
    return this.requireSigner().signAndBroadcast(signer, [this.msgRegisterTaxRule(msg)], fee, memo);
  }

  async taxedTransfer(signer: string, msg: MsgTaxedTransfer, fee: StdFee | "auto" | number = "auto", memo = "") {
    return this.requireSigner().signAndBroadcast(signer, [this.msgTaxedTransfer(msg)], fee, memo);
  }

  async spendCollectedTax(signer: string, msg: MsgSpendCollectedTax, fee: StdFee | "auto" | number = "auto", memo = "") {
    return this.requireSigner().signAndBroadcast(signer, [this.msgSpendCollectedTax(msg)], fee, memo);
  }

  // ---- queries (REST / gRPC-gateway) ----

  async params(): Promise<QueryParamsResponse> {
    return this.rest.get<QueryParamsResponse>(`${REST_PREFIX}/params`);
  }

  async taxRule(jurisdictionCode: string): Promise<QueryTaxRuleResponse> {
    return this.rest.get<QueryTaxRuleResponse>(`${REST_PREFIX}/rule/${pathSegment(jurisdictionCode)}`);
  }

  async jurisdictionBalance(jurisdictionCode: string): Promise<QueryJurisdictionBalanceResponse> {
    return this.rest.get<QueryJurisdictionBalanceResponse>(`${REST_PREFIX}/balance/${pathSegment(jurisdictionCode)}`);
  }
}
