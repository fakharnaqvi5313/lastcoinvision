import type { EncodeObject } from "@cosmjs/proto-signing";
import type { SigningStargateClient, StdFee } from "@cosmjs/stargate";

import { RestClient, pathSegment } from "../rest.js";
import { TypeUrls } from "../registry.js";
import type {
  MsgAddElectoralAuthority,
  MsgCommitVote,
  MsgCreateElection,
  MsgRegisterToVote,
  MsgRemoveElectoralAuthority,
  MsgRevealVote,
} from "../proto/lcv/voting/tx.js";
import type {
  QueryElectionResponse,
  QueryElectionResultsResponse,
  QueryElectoralAuthorityListResponse,
  QueryParamsResponse,
} from "../proto/lcv/voting/query.js";

const REST_PREFIX = "/lastcoinvision/lcv/voting";

/**
 * Typed client for `x/voting`: electoral authorities and commit-reveal
 * elections. See modules/voting/SPEC.md.
 */
export class VotingClient {
  private readonly rest: RestClient;

  constructor(
    restEndpoint: string,
    private readonly signingClient?: SigningStargateClient,
  ) {
    this.rest = new RestClient(restEndpoint);
  }

  // ---- message builders (pure, no network) ----

  msgAddElectoralAuthority(value: MsgAddElectoralAuthority): EncodeObject {
    return { typeUrl: TypeUrls.voting.addElectoralAuthority, value };
  }

  msgRemoveElectoralAuthority(value: MsgRemoveElectoralAuthority): EncodeObject {
    return { typeUrl: TypeUrls.voting.removeElectoralAuthority, value };
  }

  msgCreateElection(value: MsgCreateElection): EncodeObject {
    return { typeUrl: TypeUrls.voting.createElection, value };
  }

  msgRegisterToVote(value: MsgRegisterToVote): EncodeObject {
    return { typeUrl: TypeUrls.voting.registerToVote, value };
  }

  msgCommitVote(value: MsgCommitVote): EncodeObject {
    return { typeUrl: TypeUrls.voting.commitVote, value };
  }

  msgRevealVote(value: MsgRevealVote): EncodeObject {
    return { typeUrl: TypeUrls.voting.revealVote, value };
  }

  // ---- convenience: build + sign + broadcast ----

  private requireSigner(): SigningStargateClient {
    if (!this.signingClient) {
      throw new Error("VotingClient was constructed without a SigningStargateClient; pass one to broadcast transactions.");
    }
    return this.signingClient;
  }

  async createElection(signer: string, msg: MsgCreateElection, fee: StdFee | "auto" | number = "auto", memo = "") {
    return this.requireSigner().signAndBroadcast(signer, [this.msgCreateElection(msg)], fee, memo);
  }

  async registerToVote(signer: string, msg: MsgRegisterToVote, fee: StdFee | "auto" | number = "auto", memo = "") {
    return this.requireSigner().signAndBroadcast(signer, [this.msgRegisterToVote(msg)], fee, memo);
  }

  async commitVote(signer: string, msg: MsgCommitVote, fee: StdFee | "auto" | number = "auto", memo = "") {
    return this.requireSigner().signAndBroadcast(signer, [this.msgCommitVote(msg)], fee, memo);
  }

  async revealVote(signer: string, msg: MsgRevealVote, fee: StdFee | "auto" | number = "auto", memo = "") {
    return this.requireSigner().signAndBroadcast(signer, [this.msgRevealVote(msg)], fee, memo);
  }

  // ---- queries (REST / gRPC-gateway) ----

  async params(): Promise<QueryParamsResponse> {
    return this.rest.get<QueryParamsResponse>(`${REST_PREFIX}/params`);
  }

  async election(id: string): Promise<QueryElectionResponse> {
    return this.rest.get<QueryElectionResponse>(`${REST_PREFIX}/election/${pathSegment(id)}`);
  }

  async electionResults(electionId: string): Promise<QueryElectionResultsResponse> {
    return this.rest.get<QueryElectionResultsResponse>(`${REST_PREFIX}/election/${pathSegment(electionId)}/results`);
  }

  async electoralAuthorityList(): Promise<QueryElectoralAuthorityListResponse> {
    return this.rest.get<QueryElectoralAuthorityListResponse>(`${REST_PREFIX}/electoral-authorities`);
  }
}
