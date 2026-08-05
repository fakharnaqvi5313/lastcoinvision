# Deploy scripts

## `deploy-web.sh` — lastcoinvision.com

One-command deploy of `web/` to the production VPS: builds locally, syncs the
Next.js standalone output, restarts the service, and verifies the live site.

```bash
cp scripts/deploy/config.example.sh scripts/deploy/config.local.sh
# edit config.local.sh with your real VPS host/key/paths (gitignored, never committed)
./scripts/deploy/deploy-web.sh
```

Assumes the one-time server setup already exists (systemd unit, nginx reverse proxy, SSL —
all in place as of the initial deploy): the app user (`VPS_USER` in your config) owns the
remote app directory, a systemd unit runs it on the configured port, and nginx/certbot
handle the domain. This script only ships new application code — it never touches nginx
config, the systemd unit, or SSL certs.

`config.local.sh` (gitignored — see `config.example.sh` for the template) holds this
server's actual host/key/path details, which don't belong in git history. It sets:
`VPS_HOST`, `VPS_USER` (rsync target), `VPS_ADMIN_USER` (for `systemctl restart` — needs
root or passwordless sudo), `SSH_KEY`, `REMOTE_APP_DIR`, `SERVICE_NAME`, `APP_PORT`,
`HEALTH_URL`.

## `deploy-docs.sh` — docs.lastcoinvision.com

One-command deploy of `docs/` (the Docusaurus site) to the production VPS: builds the
static site locally and syncs it straight into nginx's docroot. Unlike `deploy-web.sh`
there's no app process to restart — nginx serves the static files directly, so a
content-only deploy is just a build + rsync + health check.

```bash
./scripts/deploy/deploy-docs.sh
```

Uses the same `config.local.sh` as `deploy-web.sh`, plus two more vars: `REMOTE_DOCS_DIR`
(the static docroot, owned by `VPS_USER` — currently `/var/www/docs.lastcoinvision.com`)
and `DOCS_HEALTH_URL`. Assumes the one-time server setup already exists (nginx vhost +
certbot SSL for `docs.lastcoinvision.com`) — this script only ships new site content, it
never touches nginx config or SSL certs.

## `deploy-faucet.sh` — faucet.lastcoinvision.com

One-command deploy of `faucet/` to the production VPS: builds locally, syncs the build
output + package manifests, installs production dependencies **on the server** (Node
native deps shouldn't be rsynced cross-platform from macOS), and restarts the service.
Never touches the server's `.env` — that holds the faucet's funded wallet mnemonic and is
created once, by hand, on the server (never in this repo, never in `config.local.sh`).

```bash
./scripts/deploy/deploy-faucet.sh
```

Uses the same `config.local.sh`, plus `REMOTE_FAUCET_DIR` (currently `/home/deploy/faucet`)
and `FAUCET_HEALTH_URL` (`/health`, a dedicated endpoint — the faucet doesn't serve
anything at `/`). Assumes the one-time server setup already exists: nginx vhost + certbot
SSL for `faucet.lastcoinvision.com`, and `/home/deploy/faucet/.env` already populated with
`FAUCET_MNEMONIC` and the rest of `faucet/.env.example`'s settings.

## `deploy-explorer.sh` — explorer.lastcoinvision.com

One-command deploy of `explorer/` (ping.pub, pointed at `lcv-testnet-1` by default) to the
production VPS: builds the static output via Docker (matching `explorer/Dockerfile`'s build
stage exactly, `CHAIN_CONFIG` overridable via env var) and syncs it into nginx's docroot.
Like `deploy-docs.sh`, there's no app process to restart.

```bash
./scripts/deploy/deploy-explorer.sh
```

Uses the same `config.local.sh`, plus `REMOTE_EXPLORER_DIR` (currently
`/var/www/explorer.lastcoinvision.com`) and `EXPLORER_HEALTH_URL`. Assumes the one-time
server setup already exists (nginx vhost + certbot SSL for `explorer.lastcoinvision.com`,
set up 2026-08-05 — see DECISIONS.md).

The chain config in `explorer/config/*.json` needs two fields ping.pub requires at runtime
but doesn't validate at build time — `logo` (top-level, must be an absolute `http(s)` URL) and
`assets[].logo`/`assets[].coingecko_id` — omitting them throws inside a `forEach` that's
building the chain list and silently aborts every chain after the broken one, including
whichever chain is actually meant to be served. See the 2026-08-05 DECISIONS.md entry for the
full root-cause trace. The Dockerfile also strips ping.pub's ~7 bundled stock demo chains
before copying LCV's own config in, so this deployment serves exactly one chain, not eight.

## `lcvd.service` / `faucet.service` — systemd units

Not deploy scripts, but the systemd unit files installed once on the VPS for the
`lcv-testnet-1` validator node and the faucet, referenced by `deploy-faucet.sh` (which
restarts `faucet.service`) and by manual `systemctl restart lcvd.service` after shipping a
new `lcvd` binary. `lcvd` is a real Cosmos SDK binary linking CosmWasm's `wasmvm`, which
requires a matching native `libwasmvm.<arch>.so` on the target's linker path — see
DECISIONS.md for how that's obtained and why it isn't just a `go build` cross-compile.

`lcvd.service` runs the node under `cosmovisor` (`cosmovisor run start`), not `lcvd start`
directly — see `scripts/upgrade/README.md` for how cosmovisor's binary-swap-at-upgrade-height
mechanism works. `$DAEMON_HOME/cosmovisor/genesis/bin/lcvd` is cosmovisor's own copy of the
binary; shipping a new non-consensus-breaking build (e.g. a docs/UI fix) still means
replacing that file specifically, not `/home/deploy/lcv-testnet-1/lcvd` (which nothing reads
anymore once cosmovisor is in place — kept only as a convenient reference copy).

## `lcvd-monitor.sh` / `.service` / `.timer` — missed-block/signing-rate check

A systemd timer (every 5 minutes) that queries the validator's own `slashing signing-info`
and `staking validators` status and logs a `WARNING` line if missed blocks approach
x/slashing's 50-of-100 jailing threshold, or if the validator isn't `BOND_STATUS_BONDED`.
Not a real alerting stack (no external paging channel exists for this project yet) — logs to
`/home/deploy/lcv-testnet-1/monitor.log`, so `grep WARNING` there (or via `journalctl -u
lcvd-monitor.service`) is a manual but real early-warning check for now.
