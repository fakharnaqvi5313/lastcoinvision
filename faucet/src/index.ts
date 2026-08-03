import express from "express";
import rateLimit from "express-rate-limit";

import { normalizeAddress } from "./address.js";
import { getAddressCooldownRemaining, recordAddressClaim } from "./addressCooldown.js";
import { config } from "./config.js";
import { getFaucetAddress, sendDrip } from "./faucetClient.js";

const app = express();
app.use(express.json());

const perIpLimiter = rateLimit({
  windowMs: config.perIpWindowMs,
  limit: config.perIpMaxRequests,
  standardHeaders: true,
  legacyHeaders: false,
  message: { error: "Too many requests from this IP. Try again later." },
});

app.get("/health", (_req, res) => {
  res.json({ status: "ok" });
});

app.post("/faucet", perIpLimiter, async (req, res) => {
  const rawAddress: unknown = req.body?.address;
  if (typeof rawAddress !== "string" || rawAddress.length === 0) {
    res.status(400).json({ error: 'Request body must include a non-empty "address" string.' });
    return;
  }

  let address: string;
  try {
    address = normalizeAddress(rawAddress, config.bech32Prefix);
  } catch (err) {
    res.status(400).json({ error: err instanceof Error ? err.message : "Invalid address." });
    return;
  }

  const cooldownMs = getAddressCooldownRemaining(address, config.perAddressCooldownMs);
  if (cooldownMs > 0) {
    res.status(429).json({ error: `Address already funded recently. Try again in ${Math.ceil(cooldownMs / 1000)}s.` });
    return;
  }

  try {
    const result = await sendDrip(address);
    recordAddressClaim(address);
    res.json({ address, amount: config.dripAmount, denom: config.denom, txHash: result.transactionHash });
  } catch (err) {
    res.status(500).json({ error: err instanceof Error ? err.message : "Broadcast failed." });
  }
});

async function main(): Promise<void> {
  const faucetAddress = await getFaucetAddress();
  app.listen(config.port, () => {
    console.log(`LCV faucet listening on :${config.port}`);
    console.log(`Faucet address: ${faucetAddress}`);
    console.log(`RPC endpoint: ${config.rpcEndpoint}`);
  });
}

main().catch((err) => {
  console.error("Failed to start faucet:", err);
  process.exitCode = 1;
});
