# Builds ping.pub (vendored as a git submodule pinned in vendor/ping-pub) with
# LCV's own chain config layered on top, then serves the static output via
# nginx. See README.md for why the submodule is never modified in place.
FROM node:20-alpine AS build
WORKDIR /app

COPY vendor/ping-pub/package.json vendor/ping-pub/yarn.lock ./
RUN corepack enable && yarn install --ignore-engines --frozen-lockfile

COPY vendor/ping-pub/. .

COPY config/ ./lcv-config/

# Which chain this build serves — a filename under config/ (also becomes the
# dashboard's route, e.g. lcv-devnet.json -> /lcv-devnet). Defaults to the
# bundled local devnet; pass --build-arg CHAIN_CONFIG=lcv-testnet-1.json (or
# via docker-compose.yml's build.args) for a real-network build instead —
# this is a build-time choice, baked into the image, not runtime-configurable
# the way faucet's LCV_RPC_URL is.
ARG CHAIN_CONFIG=lcv-devnet.json

# ping.pub's dashboard defaults to its "mainnet" chain set unless the
# deployment hostname contains "testnet" (see src/stores/useDashboard.ts) —
# irrelevant upstream when picking a public chain to browse, but this build
# only ever serves one private chain, so the config is layered into both
# sets to make sure it loads regardless of hostname.
RUN cp "./lcv-config/${CHAIN_CONFIG}" "./chains/testnet/${CHAIN_CONFIG}" && \
    cp "./lcv-config/${CHAIN_CONFIG}" "./chains/mainnet/${CHAIN_CONFIG}"

RUN yarn build

FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
