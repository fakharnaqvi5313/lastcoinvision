import { describe, expect, it, vi } from "vitest";

import type { VotingClient } from "@lastcoinvision/sdk";
import { TESTNET_DEFAULT_FEES } from "@lastcoinvision/wallet-core";

import { VotingTxClient } from "../src/voting.js";

function fakeVotingClient() {
  return {
    createElection: vi.fn().mockResolvedValue({ code: 0, transactionHash: "AAA" }),
    registerToVote: vi.fn().mockResolvedValue({ code: 0, transactionHash: "BBB" }),
    commitVote: vi.fn().mockResolvedValue({ code: 0, transactionHash: "CCC" }),
    revealVote: vi.fn().mockResolvedValue({ code: 0, transactionHash: "DDD" }),
    params: vi.fn().mockResolvedValue({ params: {} }),
    election: vi.fn().mockResolvedValue({ election: {} }),
    electionResults: vi.fn().mockResolvedValue({ results: [] }),
    electoralAuthorityList: vi.fn().mockResolvedValue({ authorities: [] }),
  } as unknown as VotingClient;
}

const ADDRESS = "lcv1abc";

describe("VotingTxClient", () => {
  it("injects the signer address and defaults to the fixed module-msg fee, never auto", async () => {
    const inner = fakeVotingClient();
    const client = new VotingTxClient(ADDRESS, inner);
    const msg = {
      creator: ADDRESS,
      title: "Board election",
      description: "Annual board election",
      eligibility_credential_type: "member",
      options: ["Alice", "Bob"],
      registration_start: "2026-08-01T00:00:00Z",
      registration_end: "2026-08-07T00:00:00Z",
      voting_start: "2026-08-08T00:00:00Z",
      voting_end: "2026-08-09T00:00:00Z",
      reveal_end: "2026-08-10T00:00:00Z",
    };

    await client.createElection(msg);

    expect(inner.createElection).toHaveBeenCalledWith(ADDRESS, msg, TESTNET_DEFAULT_FEES.moduleMsg, "");
  });

  it("passes through an explicit fee and memo override", async () => {
    const inner = fakeVotingClient();
    const client = new VotingTxClient(ADDRESS, inner);
    const msg = { voter: ADDRESS, election_id: "election-1" };
    const customFee = { amount: [{ amount: "9999", denom: "ulcv" }], gas: "500000" };

    await client.registerToVote(msg, customFee, "a memo");

    expect(inner.registerToVote).toHaveBeenCalledWith(ADDRESS, msg, customFee, "a memo");
  });

  it("commitVote and revealVote also default to the fixed fee", async () => {
    const inner = fakeVotingClient();
    const client = new VotingTxClient(ADDRESS, inner);
    const commitMsg = { voter: ADDRESS, election_id: "election-1", commitment: "0xhash" };
    const revealMsg = { voter: ADDRESS, election_id: "election-1", choice: "Alice", salt: "0xsalt" };

    await client.commitVote(commitMsg);
    await client.revealVote(revealMsg);

    expect(inner.commitVote).toHaveBeenCalledWith(ADDRESS, commitMsg, TESTNET_DEFAULT_FEES.moduleMsg, "");
    expect(inner.revealVote).toHaveBeenCalledWith(ADDRESS, revealMsg, TESTNET_DEFAULT_FEES.moduleMsg, "");
  });

  it("read-only queries pass through unmodified", async () => {
    const inner = fakeVotingClient();
    const client = new VotingTxClient(ADDRESS, inner);

    await client.params();
    await client.election("election-1");
    await client.electionResults("election-1");
    await client.electoralAuthorityList();

    expect(inner.params).toHaveBeenCalled();
    expect(inner.election).toHaveBeenCalledWith("election-1");
    expect(inner.electionResults).toHaveBeenCalledWith("election-1");
    expect(inner.electoralAuthorityList).toHaveBeenCalled();
  });
});
