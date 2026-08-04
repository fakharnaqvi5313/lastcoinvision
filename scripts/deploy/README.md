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
