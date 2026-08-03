# LCV testnet faucet

Rate-limited faucet service (Express) for LastcoinVision devnets/testnets: per-address
cooldown plus a per-IP request-rate limit, funding requests from a configured mnemonic.

## Setup

```bash
cp .env.example .env
# edit .env — set FAUCET_MNEMONIC to a funded devnet/testnet account's mnemonic.
npm install
npm run dev
```

## API

`POST /faucet` with a JSON body `{ "address": "lcv1..." }` (a bech32 address also accepts a
`0x...` address — the same account, re-encoded, for eth_secp256k1/MetaMask-style keys; see
`modules/evm/SPEC.md`).

```bash
curl -X POST http://localhost:8088/faucet \
  -H 'Content-Type: application/json' \
  -d '{"address":"lcv1..."}'
```

Success:

```json
{ "address": "lcv1...", "amount": "10000000", "denom": "ulcv", "txHash": "..." }
```

Errors: `400` for a missing/malformed address, `429` for the per-address cooldown or the
per-IP rate limit, `500` if broadcasting failed.

`GET /health` returns `{ "status": "ok" }`.

## Limits

Configurable via `.env` (see `.env.example`): `FAUCET_ADDRESS_COOLDOWN_MS` (default 24h, one
drip per address per window, tracked in-memory) and `FAUCET_IP_WINDOW_MS`/
`FAUCET_IP_MAX_REQUESTS` (default 5 requests/hour per IP, via `express-rate-limit`). Both
reset if the process restarts — fine for a testnet faucet, not intended as a durable ledger.

## Notes

- Broadcasts use a fixed fee, not `"auto"` gas estimation — see the 2026-08-03 DECISIONS.md
  entry: this chain's `"auto"`/`--gas auto` fee estimation currently underestimates gas on
  every transaction, not just this one.
- Devnet/testnet only. Never point `FAUCET_MNEMONIC` at a mainnet-funded account.
