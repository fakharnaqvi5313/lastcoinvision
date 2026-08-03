import { DirectSecp256k1HdWallet } from "@cosmjs/proto-signing";
import { SigningStargateClient, type DeliverTxResponse } from "@cosmjs/stargate";

import { config } from "./config.js";

interface FaucetSigner {
  client: SigningStargateClient;
  address: string;
}

let signerPromise: Promise<FaucetSigner> | undefined;

function getSigner(): Promise<FaucetSigner> {
  signerPromise ??= (async () => {
    const wallet = await DirectSecp256k1HdWallet.fromMnemonic(config.mnemonic, { prefix: config.bech32Prefix });
    const [account] = await wallet.getAccounts();
    if (!account) {
      throw new Error("Faucet wallet produced no accounts.");
    }
    const client = await SigningStargateClient.connectWithSigner(config.rpcEndpoint, wallet);
    return { client, address: account.address };
  })();
  return signerPromise;
}

export async function getFaucetAddress(): Promise<string> {
  const { address } = await getSigner();
  return address;
}

export async function sendDrip(recipient: string): Promise<DeliverTxResponse> {
  const { client, address } = await getSigner();
  return client.sendTokens(
    address,
    recipient,
    [{ denom: config.denom, amount: config.dripAmount }],
    // A fixed fee, not "auto" — see the 2026-08-03 DECISIONS.md entry:
    // `--gas auto`/CosmJS "auto" fee estimation underestimates on every tx
    // on this chain right now (open bug, not specific to this message).
    { amount: [{ denom: config.denom, amount: "10000" }], gas: "300000" },
  );
}
