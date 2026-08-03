function requireEnv(name: string): string {
  const value = process.env[name];
  if (!value) {
    throw new Error(`Missing required environment variable: ${name}`);
  }
  return value;
}

export const config = {
  port: Number(process.env.PORT ?? "8088"),
  rpcEndpoint: process.env.LCV_RPC_URL ?? "http://localhost:26657",
  bech32Prefix: process.env.LCV_BECH32_PREFIX ?? "lcv",
  denom: process.env.FAUCET_DENOM ?? "ulcv",
  dripAmount: process.env.FAUCET_DRIP_AMOUNT ?? "10000000",
  mnemonic: requireEnv("FAUCET_MNEMONIC"),
  perAddressCooldownMs: Number(process.env.FAUCET_ADDRESS_COOLDOWN_MS ?? String(24 * 60 * 60 * 1000)),
  perIpWindowMs: Number(process.env.FAUCET_IP_WINDOW_MS ?? String(60 * 60 * 1000)),
  perIpMaxRequests: Number(process.env.FAUCET_IP_MAX_REQUESTS ?? "5"),
} as const;
