/**
 * Node-only example: signs and broadcasts a MsgCreateDID from a mnemonic,
 * then reads the resulting DID back. Uses `DirectSecp256k1HdWallet`
 * (standard secp256k1, coin type 118) — the same key type Keplr and
 * `lcvd keys add` produce by default.
 *
 * Usage:
 *   LCV_MNEMONIC="..." LCV_RPC=http://localhost:26657 LCV_REST=http://localhost:1317 \
 *     npx tsx examples/create-did-node.ts
 */
import { DirectSecp256k1HdWallet } from "@cosmjs/proto-signing";
import { GasPrice, SigningStargateClient } from "@cosmjs/stargate";

import { createLcvRegistry, IdentityClient } from "../src/index.js";

const RPC_ENDPOINT = process.env.LCV_RPC ?? "http://localhost:26657";
const REST_ENDPOINT = process.env.LCV_REST ?? "http://localhost:1317";
const MNEMONIC = process.env.LCV_MNEMONIC;

async function main() {
  if (!MNEMONIC) {
    throw new Error("Set LCV_MNEMONIC to a funded devnet account's mnemonic before running this example.");
  }

  const wallet = await DirectSecp256k1HdWallet.fromMnemonic(MNEMONIC, { prefix: "lcv" });
  const [account] = await wallet.getAccounts();
  if (!account) {
    throw new Error("Wallet produced no accounts.");
  }

  const signingClient = await SigningStargateClient.connectWithSigner(RPC_ENDPOINT, wallet, {
    registry: createLcvRegistry(),
    // Required because IdentityClient's convenience methods use "auto" fee
    // estimation, which needs a gas price to convert gas units into ulcv.
    gasPrice: GasPrice.fromString("0.025ulcv"),
  });
  const identity = new IdentityClient(REST_ENDPOINT, signingClient);

  const result = await identity.createDID(
    account.address,
    { creator: account.address, verification_method: [], service: [] },
    // A fixed fee, not "auto": this chain's feemarket ante handler's
    // ReadFlat/WriteFlat gas cost scales with the tx's own gas limit, so
    // raising the "auto" multiplier alone never converges — every increase
    // in gasWanted raises gasUsed by a comparable amount. See
    // modules/emission/SPEC.md and DECISIONS.md for tracking; a fixed
    // limit well above simulation sidesteps the feedback loop.
    { amount: [{ denom: "ulcv", amount: "10000" }], gas: "300000" },
  );
  console.log(`MsgCreateDID broadcast (code ${result.code}): ${result.transactionHash}`);

  const did = await identity.didByAddress(account.address);
  console.log("DID document:", JSON.stringify(did, null, 2));
}

main().catch((err) => {
  console.error(err);
  process.exitCode = 1;
});
