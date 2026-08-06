// Headless verification of the EXACT logic src/lib/session.tsx drives,
// against the real local devnet -- not a mock. The React/Tauri UI wraps
// this same sequence; this proves the sequence itself works end to end,
// independent of whether a GUI screenshot can be taken of this specific
// unregistered dev binary.
import { LcvTxClients } from "@lastcoinvision/tx-builders";
import { LcvWallet } from "@lastcoinvision/wallet-core";

const NETWORK = {
  rpcEndpoint: "http://localhost:26657",
  restEndpoint: "http://localhost:1317",
  cosmosPrefix: "lcv",
};

// 1. createWallet() equivalent
const wallet = await LcvWallet.create({ cosmosPrefix: NETWORK.cosmosPrefix });
const cosmosAddress = await wallet.cosmosAddress();
console.log("Created wallet:");
console.log("  Cosmos address:", cosmosAddress);
console.log("  EVM address:   ", wallet.evmAddress);
if (cosmosAddress === wallet.evmAddress) throw new Error("Addresses must differ");

// 2. toKeystore()/fromKeystore() round trip equivalent (skip real disk/keychain -- proven in wallet-core's own tests)
const keystore = await wallet.toKeystore("a-test-password", { t: 1, m: 8, p: 1, dkLen: 32 });
const restored = await LcvWallet.fromKeystore(keystore, "a-test-password", { cosmosPrefix: NETWORK.cosmosPrefix });
if ((await restored.cosmosAddress()) !== cosmosAddress) throw new Error("Keystore round-trip address mismatch");
console.log("Keystore round-trip OK");

// 3. connectTxClients() equivalent -- NOT previously tested against a real chain (only mocked in tx-builders' own test suite)
const txClients = await LcvTxClients.connect(wallet, {
  rpcEndpoint: NETWORK.rpcEndpoint,
  restEndpoint: NETWORK.restEndpoint,
});
console.log("LcvTxClients.connect() succeeded against the real devnet");

// 4. A real read-only query through the connected client, proving the REST wiring works too
const params = await txClients.identity.params();
console.log("identity.params() query result:", JSON.stringify(params));

console.log("\nAll session-flow steps verified against the real local devnet.");
