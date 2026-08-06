import { IdentityClient, RegistryClient, TaxrailClient, VotingClient, createLcvRegistry } from "@lastcoinvision/sdk";
import type { LcvWallet } from "@lastcoinvision/wallet-core";
import { SigningStargateClient } from "@cosmjs/stargate";

import { IdentityTxClient } from "./identity.js";
import { RegistryTxClient } from "./registry.js";
import { TaxrailTxClient } from "./taxrail.js";
import { VotingTxClient } from "./voting.js";

export interface LcvTxEndpoints {
  /** Tendermint/CometBFT RPC endpoint, e.g. `http://localhost:26657`. */
  readonly rpcEndpoint: string;
  /** gRPC-gateway REST endpoint, e.g. `http://localhost:1317`. */
  readonly restEndpoint: string;
}

/**
 * One `SigningStargateClient` (built from a wallet-core `LcvWallet`'s
 * Cosmos signer, registered with every LCV custom-module message type via
 * `createLcvRegistry`), fanned out into all four fixed-fee module
 * wrappers. This is the thing a wallet UI actually holds onto after
 * "unlocking" a wallet -- one connect call, then every module's tx/query
 * methods are ready with the signer address already baked in.
 */
export class LcvTxClients {
  private constructor(
    readonly identity: IdentityTxClient,
    readonly registry: RegistryTxClient,
    readonly voting: VotingTxClient,
    readonly taxrail: TaxrailTxClient,
  ) {}

  static async connect(wallet: LcvWallet, endpoints: LcvTxEndpoints): Promise<LcvTxClients> {
    const address = await wallet.cosmosAddress();
    const signingClient = await SigningStargateClient.connectWithSigner(endpoints.rpcEndpoint, wallet.cosmosSigner, {
      registry: createLcvRegistry(),
    });

    return new LcvTxClients(
      new IdentityTxClient(address, new IdentityClient(endpoints.restEndpoint, signingClient)),
      new RegistryTxClient(address, new RegistryClient(endpoints.restEndpoint, signingClient)),
      new VotingTxClient(address, new VotingClient(endpoints.restEndpoint, signingClient)),
      new TaxrailTxClient(address, new TaxrailClient(endpoints.restEndpoint, signingClient)),
    );
  }
}
