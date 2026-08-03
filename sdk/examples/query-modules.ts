/**
 * Read-only smoke test: queries every custom module's `params` endpoint
 * over REST. No wallet or signing required — run this first to confirm
 * the SDK can reach a devnet before trying any of the signing examples.
 *
 * Usage:
 *   LCV_REST=http://localhost:1317 npx tsx examples/query-modules.ts
 */
import { IdentityClient, RegistryClient, TaxrailClient, VotingClient } from "../src/index.js";

const REST_ENDPOINT = process.env.LCV_REST ?? "http://localhost:1317";

async function main() {
  const identity = new IdentityClient(REST_ENDPOINT);
  const registry = new RegistryClient(REST_ENDPOINT);
  const voting = new VotingClient(REST_ENDPOINT);
  const taxrail = new TaxrailClient(REST_ENDPOINT);

  console.log("x/identity params:", await identity.params());
  console.log("x/registry params:", await registry.params());
  console.log("x/voting params:", await voting.params());
  console.log("x/taxrail params:", await taxrail.params());
}

main().catch((err) => {
  console.error(err);
  process.exitCode = 1;
});
