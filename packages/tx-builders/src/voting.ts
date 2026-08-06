import { VotingClient } from "@lastcoinvision/sdk";
import { TESTNET_DEFAULT_FEES } from "@lastcoinvision/wallet-core";
import type { StdFee } from "@cosmjs/stargate";

// See identity.ts's comment on why these are derived structurally rather
// than imported by name from sdk's `votingTx`/`votingQuery` namespaces.
type CreateElectionMsg = Parameters<VotingClient["createElection"]>[1];
type RegisterToVoteMsg = Parameters<VotingClient["registerToVote"]>[1];
type CommitVoteMsg = Parameters<VotingClient["commitVote"]>[1];
type RevealVoteMsg = Parameters<VotingClient["revealVote"]>[1];

/** See identity.ts's doc comment -- same pattern for x/voting. */
export class VotingTxClient {
  constructor(
    private readonly address: string,
    private readonly client: VotingClient,
  ) {}

  async createElection(msg: CreateElectionMsg, fee: StdFee = TESTNET_DEFAULT_FEES.moduleMsg, memo = "") {
    return this.client.createElection(this.address, msg, fee, memo);
  }

  async registerToVote(msg: RegisterToVoteMsg, fee: StdFee = TESTNET_DEFAULT_FEES.moduleMsg, memo = "") {
    return this.client.registerToVote(this.address, msg, fee, memo);
  }

  async commitVote(msg: CommitVoteMsg, fee: StdFee = TESTNET_DEFAULT_FEES.moduleMsg, memo = "") {
    return this.client.commitVote(this.address, msg, fee, memo);
  }

  async revealVote(msg: RevealVoteMsg, fee: StdFee = TESTNET_DEFAULT_FEES.moduleMsg, memo = "") {
    return this.client.revealVote(this.address, msg, fee, memo);
  }

  params(): ReturnType<VotingClient["params"]> {
    return this.client.params();
  }

  election(id: string): ReturnType<VotingClient["election"]> {
    return this.client.election(id);
  }

  electionResults(electionId: string): ReturnType<VotingClient["electionResults"]> {
    return this.client.electionResults(electionId);
  }

  electoralAuthorityList(): ReturnType<VotingClient["electoralAuthorityList"]> {
    return this.client.electoralAuthorityList();
  }
}
