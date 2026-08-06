// Headless verification of the same session-flow sequence
// src/lib/session.tsx drives, against the real local devnet -- same
// script as apps/desktop/scripts/verify-session-flow.mjs, since both
// apps' session logic is identical (only the storage backend differs:
// Tauri keychain commands vs expo-secure-store). Does NOT exercise the
// RN/Hermes-specific polyfills (src/lib/polyfills.ts) -- that needs an
// actual device/simulator run, which this script can't substitute for.
import { LcvTxClients } from "@lastcoinvision/tx-builders";
import { LcvWallet } from "@lastcoinvision/wallet-core";

const NETWORK = {
  rpcEndpoint: "http://localhost:26657",
  restEndpoint: "http://localhost:1317",
  cosmosPrefix: "lcv",
};

const wallet = await LcvWallet.create({ cosmosPrefix: NETWORK.cosmosPrefix });
const cosmosAddress = await wallet.cosmosAddress();
console.log("Created wallet:");
console.log("  Cosmos address:", cosmosAddress);
console.log("  EVM address:   ", wallet.evmAddress);
if (cosmosAddress === wallet.evmAddress) throw new Error("Addresses must differ");

const keystore = await wallet.toKeystore("a-test-password", { t: 1, m: 8, p: 1, dkLen: 32 });
const restored = await LcvWallet.fromKeystore(keystore, "a-test-password", { cosmosPrefix: NETWORK.cosmosPrefix });
if ((await restored.cosmosAddress()) !== cosmosAddress) throw new Error("Keystore round-trip address mismatch");
console.log("Keystore round-trip OK");

const txClients = await LcvTxClients.connect(wallet, {
  rpcEndpoint: NETWORK.rpcEndpoint,
  restEndpoint: NETWORK.restEndpoint,
});
console.log("LcvTxClients.connect() succeeded against the real devnet");

const params = await txClients.identity.params();
console.log("identity.params() query result:", JSON.stringify(params));

console.log("\nAll session-flow steps verified against the real local devnet.");
