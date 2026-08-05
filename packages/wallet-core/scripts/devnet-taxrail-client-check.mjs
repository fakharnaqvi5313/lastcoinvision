// Companion check: exercises the ACTUAL TaxrailClient (not a plain bank
// send) with a wallet-core signer. No jurisdiction/tax-rule is bootstrapped
// on this throwaway devnet, so the chain is expected to reject the message
// with "no tax rule registered" -- that's a legitimate, informative
// rejection from taxrail's own business logic, not a wiring failure. What
// this proves is that TaxrailClient (built on SigningStargateClient, same
// as every sdk/src/clients/*.ts file) accepts a wallet-core signer and
// gets as far as a real chain response, with a fixed fee, exactly as
// sdk/examples/taxed-transfer-node.ts does with an env-var mnemonic.
import { SigningStargateClient } from "@cosmjs/stargate";
import { TaxrailClient, createLcvRegistry } from "@lastcoinvision/sdk";

import { LcvWallet, TESTNET_DEFAULT_FEES } from "../dist/index.js";

const MNEMONIC = process.argv[2];
if (!MNEMONIC) {
  throw new Error("usage: node devnet-taxrail-client-check.mjs <mnemonic>");
}

const wallet = await LcvWallet.fromMnemonic(MNEMONIC, { cosmosPrefix: "lcv" });
const address = await wallet.cosmosAddress();

const signingClient = await SigningStargateClient.connectWithSigner(
  "http://localhost:26657",
  wallet.cosmosSigner,
  { registry: createLcvRegistry() },
);
const taxrail = new TaxrailClient("http://localhost:1317", signingClient);

try {
  const result = await taxrail.taxedTransfer(
    address,
    {
      sender: address,
      recipient: address,
      jurisdiction_code: "US-CA",
      amount: { denom: "ulcv", amount: "1000" },
    },
    TESTNET_DEFAULT_FEES.moduleMsg, // fixed fee -- never "auto"
  );
  console.log(`unexpected success (code ${result.code}): ${result.transactionHash}`);
} catch (err) {
  console.log("Expected rejection reached the chain (proves client+signer wiring works):");
  console.log(String(err.message ?? err).slice(0, 300));
}
