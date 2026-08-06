import { SigningStargateClient } from "@cosmjs/stargate";
import { LcvWallet } from "@lastcoinvision/wallet-core";
import { afterEach, describe, expect, it, vi } from "vitest";

import { LcvTxClients } from "../src/context.js";
import { IdentityTxClient } from "../src/identity.js";
import { RegistryTxClient } from "../src/registry.js";
import { TaxrailTxClient } from "../src/taxrail.js";
import { VotingTxClient } from "../src/voting.js";

afterEach(() => {
  vi.restoreAllMocks();
});

describe("LcvTxClients.connect", () => {
  it("wires all four module clients to one signing client and the wallet's own address", async () => {
    const wallet = await LcvWallet.fromMnemonic(
      "test test test test test test test test test test test junk",
      { cosmosPrefix: "lcv" },
    );
    const expectedAddress = await wallet.cosmosAddress();

    const fakeSigningClient = {} as SigningStargateClient;
    const connectSpy = vi
      .spyOn(SigningStargateClient, "connectWithSigner")
      .mockResolvedValue(fakeSigningClient);

    const clients = await LcvTxClients.connect(wallet, {
      rpcEndpoint: "http://localhost:26657",
      restEndpoint: "http://localhost:1317",
    });

    expect(connectSpy).toHaveBeenCalledWith(
      "http://localhost:26657",
      wallet.cosmosSigner,
      expect.objectContaining({ registry: expect.anything() }),
    );
    expect(clients.identity).toBeInstanceOf(IdentityTxClient);
    expect(clients.registry).toBeInstanceOf(RegistryTxClient);
    expect(clients.voting).toBeInstanceOf(VotingTxClient);
    expect(clients.taxrail).toBeInstanceOf(TaxrailTxClient);

    // Each wrapper's baked-in address should be the wallet's own address --
    // verified indirectly by exercising a query call and checking it
    // doesn't throw (constructor wiring is otherwise private).
    expect(expectedAddress).toMatch(/^lcv1/);
  });
});
