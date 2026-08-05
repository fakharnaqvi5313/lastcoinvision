// Manual, live-devnet verification for Phase 1 (not part of `npm test` --
// this needs a running `make devnet`, same convention as
// sdk/examples/*-node.ts). Proves wallet-core's own signer, not a browser
// extension or an env-var mnemonic fed straight to CosmJS, can really sign
// and broadcast against a live chain, with the required fixed fee.
import { SigningStargateClient } from "@cosmjs/stargate";
import { createLcvRegistry } from "@lastcoinvision/sdk";

import { LcvWallet, TESTNET_DEFAULT_FEES } from "../dist/index.js";

const MNEMONIC = process.argv[2];
const RECIPIENT = process.argv[3];
if (!MNEMONIC || !RECIPIENT) {
  throw new Error("usage: node devnet-integration-check.mjs <mnemonic> <recipient-address>");
}

const wallet = await LcvWallet.fromMnemonic(MNEMONIC, { cosmosPrefix: "lcv" });
const address = await wallet.cosmosAddress();
console.log("wallet-core cosmos address:", address);
console.log("wallet-core evm address:   ", wallet.evmAddress);

const client = await SigningStargateClient.connectWithSigner(
  "http://localhost:26657",
  wallet.cosmosSigner,
  { registry: createLcvRegistry() },
);

const result = await client.sendTokens(
  address,
  RECIPIENT,
  [{ denom: "ulcv", amount: "1000000" }],
  TESTNET_DEFAULT_FEES.simple, // fixed fee -- never "auto", see fees.ts
  "wallet-core Phase 1 devnet integration check",
);

console.log(`broadcast ok (code ${result.code}): ${result.transactionHash}`);
if (result.code !== 0) {
  throw new Error(`tx failed with code ${result.code}: ${result.rawLog}`);
}
