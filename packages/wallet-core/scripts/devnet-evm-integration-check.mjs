// Manual, live-devnet verification (EVM side) -- companion to
// devnet-integration-check.mjs. Proves wallet-core's evmSigner (viem
// account + buildEvmWalletClient) can really sign and broadcast a
// transaction against LCV's EVM JSON-RPC, independent of the Cosmos side.
import { parseEther } from "viem";

import { deriveEvmAccount, buildEvmWalletClient } from "../dist/index.js";

const MNEMONIC = process.argv[2];
if (!MNEMONIC) {
  throw new Error("usage: node devnet-evm-integration-check.mjs <mnemonic>");
}

const account = deriveEvmAccount(MNEMONIC);
const client = buildEvmWalletClient(account, {
  evmChainId: 2611,
  chainName: "LCV Devnet",
  rpcUrl: "http://localhost:8545",
});

// A second, freshly generated recipient -- doesn't need to be prefunded.
const recipient = deriveEvmAccount(
  "legal winner thank year wave sausage worth useful legal winner thank yellow",
);

console.log("sender:   ", account.address);
console.log("recipient:", recipient.address);

const hash = await client.sendTransaction({
  account,
  to: recipient.address,
  value: parseEther("0.1"),
  chain: client.chain,
});

console.log("broadcast tx hash:", hash);
