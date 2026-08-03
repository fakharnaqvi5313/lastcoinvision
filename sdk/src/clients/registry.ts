import type { EncodeObject } from "@cosmjs/proto-signing";
import type { SigningStargateClient, StdFee } from "@cosmjs/stargate";

import { RestClient, pathSegment } from "../rest.js";
import { TypeUrls } from "../registry.js";
import type {
  MsgFreezeAsset,
  MsgRegisterAsset,
  MsgSetJurisdictionAuthority,
  MsgTransferAsset,
} from "../proto/lcv/registry/tx.js";
import type {
  QueryAssetByIDResponse,
  QueryJurisdictionAuthorityOfResponse,
  QueryParamsResponse,
  QueryTransferHistoryResponse,
} from "../proto/lcv/registry/query.js";

const REST_PREFIX = "/lastcoinvision/lcv/registry";

/**
 * Typed client for `x/registry`: jurisdiction authorities and the
 * registered-asset ledger (register/transfer/freeze). See
 * modules/registry/SPEC.md.
 */
export class RegistryClient {
  private readonly rest: RestClient;

  constructor(
    restEndpoint: string,
    private readonly signingClient?: SigningStargateClient,
  ) {
    this.rest = new RestClient(restEndpoint);
  }

  // ---- message builders (pure, no network) ----

  msgSetJurisdictionAuthority(value: MsgSetJurisdictionAuthority): EncodeObject {
    return { typeUrl: TypeUrls.registry.setJurisdictionAuthority, value };
  }

  msgRegisterAsset(value: MsgRegisterAsset): EncodeObject {
    return { typeUrl: TypeUrls.registry.registerAsset, value };
  }

  msgTransferAsset(value: MsgTransferAsset): EncodeObject {
    return { typeUrl: TypeUrls.registry.transferAsset, value };
  }

  msgFreezeAsset(value: MsgFreezeAsset): EncodeObject {
    return { typeUrl: TypeUrls.registry.freezeAsset, value };
  }

  // ---- convenience: build + sign + broadcast ----

  private requireSigner(): SigningStargateClient {
    if (!this.signingClient) {
      throw new Error("RegistryClient was constructed without a SigningStargateClient; pass one to broadcast transactions.");
    }
    return this.signingClient;
  }

  async registerAsset(signer: string, msg: MsgRegisterAsset, fee: StdFee | "auto" | number = "auto", memo = "") {
    return this.requireSigner().signAndBroadcast(signer, [this.msgRegisterAsset(msg)], fee, memo);
  }

  async transferAsset(signer: string, msg: MsgTransferAsset, fee: StdFee | "auto" | number = "auto", memo = "") {
    return this.requireSigner().signAndBroadcast(signer, [this.msgTransferAsset(msg)], fee, memo);
  }

  async freezeAsset(signer: string, msg: MsgFreezeAsset, fee: StdFee | "auto" | number = "auto", memo = "") {
    return this.requireSigner().signAndBroadcast(signer, [this.msgFreezeAsset(msg)], fee, memo);
  }

  // ---- queries (REST / gRPC-gateway) ----

  async params(): Promise<QueryParamsResponse> {
    return this.rest.get<QueryParamsResponse>(`${REST_PREFIX}/params`);
  }

  async assetByID(id: string): Promise<QueryAssetByIDResponse> {
    return this.rest.get<QueryAssetByIDResponse>(`${REST_PREFIX}/asset/${pathSegment(id)}`);
  }

  async transferHistory(assetId: string): Promise<QueryTransferHistoryResponse> {
    return this.rest.get<QueryTransferHistoryResponse>(`${REST_PREFIX}/asset/${pathSegment(assetId)}/history`);
  }

  async jurisdictionAuthorityOf(jurisdictionCode: string): Promise<QueryJurisdictionAuthorityOfResponse> {
    return this.rest.get<QueryJurisdictionAuthorityOfResponse>(
      `${REST_PREFIX}/jurisdiction/${pathSegment(jurisdictionCode)}`,
    );
  }
}
