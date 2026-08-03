# Block explorer

[ping.pub](https://github.com/ping-pub/explorer) (GPLv2), vendored as a git submodule pinned
at `vendor/ping-pub` (commit `20e65d6`) with LCV's chain config layered on top at build time —
see `Dockerfile`. The submodule itself is never modified; `config/lcv-devnet.json` is copied
into its `chains/` directory during the Docker build instead, so `vendor/ping-pub` stays a
clean, unmodified checkout of upstream.

## Files

- `vendor/ping-pub/` — the submodule (run `git submodule update --init` if it's empty after
  cloning this repo)
- `config/lcv-devnet.json` — LCV devnet's chain config, in ping.pub's own schema (RPC/API
  URLs, bech32 prefix, coin type, asset metadata) — values taken directly from
  `chain/app/app.go` and `chain/go.mod`, not guessed
- `Dockerfile` — multi-stage build: `yarn build` the submodule with our config layered in,
  then serve the static output via nginx
- `nginx.conf` — adds the SPA fallback (`try_files ... /index.html`) ping.pub's Vue Router
  needs for direct navigation to `/lcv-devnet`; without it nginx 404s on anything but `/`

## Why the config is copied into both `chains/testnet/` and `chains/mainnet/`

ping.pub's dashboard defaults to loading its `chains/mainnet/*.json` set unless the deployment
hostname contains the literal substring `"testnet"` (see
`vendor/ping-pub/src/stores/useDashboard.ts`). That heuristic is meant for ping.pub's own
public multi-chain instance; a build that only ever serves one private devnet has no use for
it, so `Dockerfile` copies `lcv-devnet.json` into both directories to make sure it loads
regardless of what hostname it's served from.

## Build and run standalone

```bash
docker build -t lcv-explorer .
docker run -p 8090:80 lcv-explorer
# http://localhost:8090/lcv-devnet
```

(Confirmed during Phase H: the image builds, nginx serves the SPA correctly, and the
`/lcv-devnet` route loads with our chain config bundled in — checked without a devnet running,
so it showed "Not Connected." The dashboard fetches `localhost:1317`/`localhost:26657` from
the *browser*, not the container, so those ports need to be reachable from wherever you open
it.)

Normally run as part of the full stack instead — see `../deploy/docker-compose/`.
