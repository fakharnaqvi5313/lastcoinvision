import { afterEach, describe, expect, it, vi } from "vitest";

import { VotingClient } from "../../src/clients/voting.js";
import { TypeUrls } from "../../src/registry.js";
import { fakeSigningClient, mockFetchJsonOnce } from "../testUtils.js";

const REST = "http://localhost:1317";

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe("VotingClient message builders", () => {
  const client = new VotingClient(REST);

  it("builds MsgAddElectoralAuthority", () => {
    const msg = { authority: "lcv1gov", address: "lcv1authority", name: "County Elections Board" };
    expect(client.msgAddElectoralAuthority(msg)).toEqual({
      typeUrl: TypeUrls.voting.addElectoralAuthority,
      value: msg,
    });
  });

  it("builds MsgRemoveElectoralAuthority", () => {
    const msg = { authority: "lcv1gov", address: "lcv1authority" };
    expect(client.msgRemoveElectoralAuthority(msg)).toEqual({
      typeUrl: TypeUrls.voting.removeElectoralAuthority,
      value: msg,
    });
  });

  it("builds MsgCreateElection", () => {
    const msg = {
      creator: "lcv1authority",
      title: "2026 County Referendum",
      description: "desc",
      eligibility_credential_type: "kyc",
      options: ["yes", "no"],
      registration_start: "2026-01-01T00:00:00Z",
      registration_end: "2026-01-15T00:00:00Z",
      voting_start: "2026-02-01T00:00:00Z",
      voting_end: "2026-02-10T00:00:00Z",
      reveal_end: "2026-02-12T00:00:00Z",
    };
    expect(client.msgCreateElection(msg)).toEqual({ typeUrl: TypeUrls.voting.createElection, value: msg });
  });

  it("builds MsgRegisterToVote", () => {
    const msg = { voter: "lcv1voter", election_id: "election-1" };
    expect(client.msgRegisterToVote(msg)).toEqual({ typeUrl: TypeUrls.voting.registerToVote, value: msg });
  });

  it("builds MsgCommitVote", () => {
    const msg = { voter: "lcv1voter", election_id: "election-1", commitment: "0xcommit" };
    expect(client.msgCommitVote(msg)).toEqual({ typeUrl: TypeUrls.voting.commitVote, value: msg });
  });

  it("builds MsgRevealVote", () => {
    const msg = { voter: "lcv1voter", election_id: "election-1", choice: "yes", salt: "0xsalt" };
    expect(client.msgRevealVote(msg)).toEqual({ typeUrl: TypeUrls.voting.revealVote, value: msg });
  });
});

describe("VotingClient convenience broadcast methods", () => {
  it("throws when constructed without a signing client", async () => {
    const client = new VotingClient(REST);
    await expect(client.registerToVote("lcv1voter", { voter: "lcv1voter", election_id: "election-1" })).rejects.toThrow(
      /without a SigningStargateClient/,
    );
  });

  it("createElection signs and broadcasts the built message", async () => {
    const { client: signingClient, signAndBroadcast } = fakeSigningClient();
    const client = new VotingClient(REST, signingClient);
    const msg = {
      creator: "lcv1authority",
      title: "2026 County Referendum",
      description: "desc",
      eligibility_credential_type: "kyc",
      options: ["yes", "no"],
      registration_start: "2026-01-01T00:00:00Z",
      registration_end: "2026-01-15T00:00:00Z",
      voting_start: "2026-02-01T00:00:00Z",
      voting_end: "2026-02-10T00:00:00Z",
      reveal_end: "2026-02-12T00:00:00Z",
    };

    await client.createElection("lcv1authority", msg, "auto", "memo");

    expect(signAndBroadcast).toHaveBeenCalledWith(
      "lcv1authority",
      [{ typeUrl: TypeUrls.voting.createElection, value: msg }],
      "auto",
      "memo",
    );
  });

  it("registerToVote signs and broadcasts the built message", async () => {
    const { client: signingClient, signAndBroadcast } = fakeSigningClient();
    const client = new VotingClient(REST, signingClient);
    const msg = { voter: "lcv1voter", election_id: "election-1" };

    await client.registerToVote("lcv1voter", msg);

    expect(signAndBroadcast).toHaveBeenCalledWith(
      "lcv1voter",
      [{ typeUrl: TypeUrls.voting.registerToVote, value: msg }],
      "auto",
      "",
    );
  });

  it("commitVote signs and broadcasts the built message", async () => {
    const { client: signingClient, signAndBroadcast } = fakeSigningClient();
    const client = new VotingClient(REST, signingClient);
    const msg = { voter: "lcv1voter", election_id: "election-1", commitment: "0xcommit" };

    await client.commitVote("lcv1voter", msg);

    expect(signAndBroadcast).toHaveBeenCalledWith(
      "lcv1voter",
      [{ typeUrl: TypeUrls.voting.commitVote, value: msg }],
      "auto",
      "",
    );
  });

  it("revealVote signs and broadcasts the built message", async () => {
    const { client: signingClient, signAndBroadcast } = fakeSigningClient();
    const client = new VotingClient(REST, signingClient);
    const msg = { voter: "lcv1voter", election_id: "election-1", choice: "yes", salt: "0xsalt" };

    await client.revealVote("lcv1voter", msg);

    expect(signAndBroadcast).toHaveBeenCalledWith(
      "lcv1voter",
      [{ typeUrl: TypeUrls.voting.revealVote, value: msg }],
      "auto",
      "",
    );
  });
});

describe("VotingClient queries", () => {
  const client = new VotingClient(REST);

  it("params()", async () => {
    const fetchMock = mockFetchJsonOnce({ params: {} });
    await client.params();
    expect(fetchMock).toHaveBeenCalledWith(`${REST}/lastcoinvision/lcv/voting/params`);
  });

  it("election()", async () => {
    const fetchMock = mockFetchJsonOnce({ election: { id: "election-1" } });
    await client.election("election-1");
    expect(fetchMock).toHaveBeenCalledWith(`${REST}/lastcoinvision/lcv/voting/election/election-1`);
  });

  it("electionResults()", async () => {
    const fetchMock = mockFetchJsonOnce({ tallies: [] });
    await client.electionResults("election-1");
    expect(fetchMock).toHaveBeenCalledWith(`${REST}/lastcoinvision/lcv/voting/election/election-1/results`);
  });

  it("electoralAuthorityList()", async () => {
    const fetchMock = mockFetchJsonOnce({ authorities: [] });
    await client.electoralAuthorityList();
    expect(fetchMock).toHaveBeenCalledWith(`${REST}/lastcoinvision/lcv/voting/electoral-authorities`);
  });

  it("propagates REST errors", async () => {
    mockFetchJsonOnce({ message: "not found" }, false, 404);
    await expect(client.election("missing")).rejects.toThrow(/404/);
  });
});
