import type { EncodeObject } from "@cosmjs/proto-signing";
import type { SigningStargateClient, StdFee } from "@cosmjs/stargate";

import { RestClient, pathSegment } from "../rest.js";
import { TypeUrls } from "../registry.js";
import type {
  MsgAddIssuer,
  MsgCreateDID,
  MsgDeactivateDID,
  MsgIssueCredential,
  MsgPresentCredential,
  MsgRemoveIssuer,
  MsgRevokeCredential,
  MsgUpdateDID,
} from "../proto/lcv/identity/tx.js";
import type {
  QueryCredentialStatusResponse,
  QueryDIDByAddressResponse,
  QueryIssuerListResponse,
  QueryParamsResponse,
  QueryTierOfResponse,
} from "../proto/lcv/identity/query.js";

const REST_PREFIX = "/lastcoinvision/lcv/identity";

/**
 * Typed client for `x/identity`: W3C DIDs, verifiable-credential hashes,
 * and account-tier lookups. See modules/identity/SPEC.md.
 *
 * Message-builder methods (`msgXxx`) are pure and return a CosmJS
 * `EncodeObject` ready for `SigningStargateClient.signAndBroadcast` — they
 * make no network calls, which is what this SDK's unit tests target.
 * Convenience `xxx` methods do the same encoding and then broadcast
 * immediately, for callers that don't need to batch multiple messages into
 * one transaction.
 */
export class IdentityClient {
  private readonly rest: RestClient;

  constructor(
    restEndpoint: string,
    private readonly signingClient?: SigningStargateClient,
  ) {
    this.rest = new RestClient(restEndpoint);
  }

  // ---- message builders (pure, no network) ----

  msgCreateDID(value: MsgCreateDID): EncodeObject {
    return { typeUrl: TypeUrls.identity.createDID, value };
  }

  msgUpdateDID(value: MsgUpdateDID): EncodeObject {
    return { typeUrl: TypeUrls.identity.updateDID, value };
  }

  msgDeactivateDID(value: MsgDeactivateDID): EncodeObject {
    return { typeUrl: TypeUrls.identity.deactivateDID, value };
  }

  msgAddIssuer(value: MsgAddIssuer): EncodeObject {
    return { typeUrl: TypeUrls.identity.addIssuer, value };
  }

  msgRemoveIssuer(value: MsgRemoveIssuer): EncodeObject {
    return { typeUrl: TypeUrls.identity.removeIssuer, value };
  }

  msgIssueCredential(value: MsgIssueCredential): EncodeObject {
    return { typeUrl: TypeUrls.identity.issueCredential, value };
  }

  msgRevokeCredential(value: MsgRevokeCredential): EncodeObject {
    return { typeUrl: TypeUrls.identity.revokeCredential, value };
  }

  msgPresentCredential(value: MsgPresentCredential): EncodeObject {
    return { typeUrl: TypeUrls.identity.presentCredential, value };
  }

  // ---- convenience: build + sign + broadcast ----

  private requireSigner(): SigningStargateClient {
    if (!this.signingClient) {
      throw new Error("IdentityClient was constructed without a SigningStargateClient; pass one to broadcast transactions.");
    }
    return this.signingClient;
  }

  async createDID(signer: string, msg: MsgCreateDID, fee: StdFee | "auto" | number = "auto", memo = "") {
    return this.requireSigner().signAndBroadcast(signer, [this.msgCreateDID(msg)], fee, memo);
  }

  async issueCredential(signer: string, msg: MsgIssueCredential, fee: StdFee | "auto" | number = "auto", memo = "") {
    return this.requireSigner().signAndBroadcast(signer, [this.msgIssueCredential(msg)], fee, memo);
  }

  async presentCredential(signer: string, msg: MsgPresentCredential, fee: StdFee | "auto" | number = "auto", memo = "") {
    return this.requireSigner().signAndBroadcast(signer, [this.msgPresentCredential(msg)], fee, memo);
  }

  // ---- queries (REST / gRPC-gateway) ----

  async params(): Promise<QueryParamsResponse> {
    return this.rest.get<QueryParamsResponse>(`${REST_PREFIX}/params`);
  }

  async didByAddress(address: string): Promise<QueryDIDByAddressResponse> {
    return this.rest.get<QueryDIDByAddressResponse>(`${REST_PREFIX}/did/${pathSegment(address)}`);
  }

  async credentialStatus(id: string): Promise<QueryCredentialStatusResponse> {
    return this.rest.get<QueryCredentialStatusResponse>(`${REST_PREFIX}/credential/${pathSegment(id)}`);
  }

  async issuerList(): Promise<QueryIssuerListResponse> {
    return this.rest.get<QueryIssuerListResponse>(`${REST_PREFIX}/issuers`);
  }

  async tierOf(address: string): Promise<QueryTierOfResponse> {
    return this.rest.get<QueryTierOfResponse>(`${REST_PREFIX}/tier/${pathSegment(address)}`);
  }
}
