# Local docker-compose stack

4 validators + faucet + explorer + website, all containerized, wired together on one
docker-compose network. This is a local devnet for development/demoing, not a production
deployment — see `../k8s/` for the public-testnet topology.

## Quick start

```bash
./generate.sh   # one-time: builds lcvd, generates a 4-validator genesis, writes .env
docker compose up --build
```

Then:

- RPC: http://localhost:26657 (validator1-3 on 26667/26677/26687)
- REST: http://localhost:1317 (validator1-3 on 1327/1337/1347)
- gRPC: localhost:9090 (validator1-3 on 9100/9110/9120)
- EVM JSON-RPC: http://localhost:8545 (validator1-3 on 8555/8565/8575)
- Faucet: http://localhost:8088 (`POST /faucet {"address":"lcv1..."}`)
- Explorer: http://localhost:8090/lcv-devnet
- Website: http://localhost:3000

## Why `generate.sh` exists

`scripts/genesis/init-genesis.sh` (used by `make devnet`) is designed for one host running 4
`lcvd` processes on `127.0.0.1`, distinguished by port offsets (26657/26667/26677/26687, ...).
Docker Compose runs 4 *separate containers* instead, each with its own loopback — a
`persistent_peers` entry pointing at `localhost:26667` from inside validator0's container
would try to dial validator0 itself, not validator1.

`generate.sh` calls `init-genesis.sh` unmodified (same keys, same gentxs, same token
allocation table — zero duplication of that logic), then rewrites each validator's
`config.toml`/`app.toml` so every container binds the *same* default port and reaches its
peers by Compose service name (`lcvd0`..`lcvd3`) instead of by port offset. `docker-compose.yml`
then does the differentiation on the host side, mapping a distinct host port to each
container's fixed internal port.

It also extracts the `public` allocation-bucket account's mnemonic (already funded with 8% of
genesis supply — see `modules/emission/SPEC.md`) into a generated `.env`, so the faucet
container starts already funded without a manual funding step.

`.devnet/` and `.env` are gitignored — regenerate with `./generate.sh` any time you want a
fresh chain state.

## Rebuilding after a code change

```bash
docker compose build lcvd0 lcvd1 lcvd2 lcvd3   # chain/ changed
docker compose build faucet                     # faucet/ changed
docker compose build web                        # web/ changed
docker compose build explorer                   # explorer/ changed
docker compose up
```

(All four `lcvd*` services share one image definition — `docker compose build` only compiles
it once and reuses the result for all four.)

## Resetting

```bash
docker compose down
rm -rf .devnet .env
./generate.sh
```
