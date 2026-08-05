# Copy to config.local.sh (gitignored) and fill in your VPS's real values.
# Sourced by deploy-web.sh — plain shell variable assignments.

VPS_HOST="203.0.113.10"                          # your VPS's IP or hostname
VPS_USER="deploy"                                # owns the app's files on the VPS
VPS_ADMIN_USER="root"                             # needs root or passwordless sudo for systemctl
SSH_KEY="$HOME/.ssh/your_vps_key"
REMOTE_APP_DIR="/home/deploy/lastcoinvision-website"
SERVICE_NAME="lastcoinvision"
APP_PORT="3004"
HEALTH_URL="https://lastcoinvision.com/"

REMOTE_DOCS_DIR="/var/www/docs.lastcoinvision.com"       # static Docusaurus docroot, owned by deploy user
DOCS_HEALTH_URL="https://docs.lastcoinvision.com/"

REMOTE_FAUCET_DIR="/home/deploy/faucet"                  # faucet app dir; .env (FAUCET_MNEMONIC, etc.) lives here, created once by hand
FAUCET_HEALTH_URL="https://faucet.lastcoinvision.com/health"

REMOTE_EXPLORER_DIR="/var/www/explorer.lastcoinvision.com"  # static ping.pub docroot, owned by deploy user
EXPLORER_HEALTH_URL="https://explorer.lastcoinvision.com/"
