/**
 * Node-only example: broadcasts a MsgTaxedTransfer from a mnemonic and
 * prints the tax/remainder split from the emitted response, then reads
 * back the jurisdiction's collected-tax balance.
 *
 * Usage:
 *   LCV_MNEMONIC="..." LCV_RECIPIENT=lcv1... LCV_JURISDICTION=US-CA \
 *     LCV_AMOUNT=1000000 npx tsx examples/taxed-transfer-node.ts
 */
import { DirectSecp256k1HdWallet } from "@cosmjs/proto-signing";
import { GasPrice, SigningStargateClient } from "@cosmjs/stargate";

import { createLcvRegistry, TaxrailClient } from "../src/index.js";

const RPC_ENDPOINT = process.env.LCV_RPC ?? "http://localhost:26657";
const REST_ENDPOINT = process.env.LCV_REST ?? "http://localhost:1317";
const MNEMONIC = process.env.LCV_MNEMONIC;
const RECIPIENT = process.env.LCV_RECIPIENT;
const JURISDICTION_CODE = process.env.LCV_JURISDICTION ?? "US-CA";
const AMOUNT = process.env.LCV_AMOUNT ?? "1000000";

async function main() {
  if (!MNEMONIC || !RECIPIENT) {
    throw new Error("Set LCV_MNEMONIC and LCV_RECIPIENT before running this example.");
  }

  const wallet = await DirectSecp256k1HdWallet.fromMnemonic(MNEMONIC, { prefix: "lcv" });
  const [account] = await wallet.getAccounts();
  if (!account) {
    throw new Error("Wallet produced no accounts.");
  }

  const signingClient = await SigningStargateClient.connectWithSigner(RPC_ENDPOINT, wallet, {
    registry: createLcvRegistry(),
    gasPrice: GasPrice.fromString("0.025ulcv"),
  });
  const taxrail = new TaxrailClient(REST_ENDPOINT, signingClient);

  const result = await taxrail.taxedTransfer(
    account.address,
    {
      sender: account.address,
      recipient: RECIPIENT,
      jurisdiction_code: JURISDICTION_CODE,
      amount: { denom: "ulcv", amount: AMOUNT },
    },
    // A fixed fee, not "auto" — see the 2026-08-03 DECISIONS.md entry:
    // `--gas auto`/CosmJS "auto" fee estimation underestimates on every tx
    // on this chain right now (open bug, not specific to this message).
    { amount: [{ denom: "ulcv", amount: "10000" }], gas: "300000" },
  );
  console.log(`MsgTaxedTransfer broadcast (code ${result.code}): ${result.transactionHash}`);

  const balance = await taxrail.jurisdictionBalance(JURISDICTION_CODE);
  console.log(`Collected tax balance for ${JURISDICTION_CODE}:`, balance);
}

main().catch((err) => {
  console.error(err);
  process.exitCode = 1;
});
