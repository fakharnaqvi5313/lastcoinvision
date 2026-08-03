# Builds ping.pub (vendored as a git submodule pinned in vendor/ping-pub) with
# LCV's own chain config layered on top, then serves the static output via
# nginx. See README.md for why the submodule is never modified in place.
FROM node:20-alpine AS build
WORKDIR /app

COPY vendor/ping-pub/package.json vendor/ping-pub/yarn.lock ./
RUN corepack enable && yarn install --ignore-engines --frozen-lockfile

COPY vendor/ping-pub/. .

# ping.pub's dashboard defaults to its "mainnet" chain set unless the
# deployment hostname contains "testnet" (see src/stores/useDashboard.ts) —
# irrelevant upstream when picking a public chain to browse, but this build
# only ever serves one private chain, so the config is layered into both
# sets to make sure it loads regardless of hostname.
COPY config/lcv-devnet.json ./chains/testnet/lcv-devnet.json
COPY config/lcv-devnet.json ./chains/mainnet/lcv-devnet.json

RUN yarn build

FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
