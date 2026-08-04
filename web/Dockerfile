FROM node:20-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
# Next.js inlines NEXT_PUBLIC_* vars into the client bundle at build time, not
# runtime — an ARG (rather than relying on a local .env.local incidentally
# being present in the build context) makes the target network an explicit,
# reproducible build input. Default matches the bundled local devnet; pass
# --build-arg NEXT_PUBLIC_LCV_RPC_URL=https://rpc.lastcoinvision.com (or via
# docker-compose.yml's build.args) to point a built image at a real network
# instead — this can't be changed after the image is built, unlike faucet's
# runtime-read LCV_RPC_URL.
ARG NEXT_PUBLIC_LCV_RPC_URL=http://localhost:26657
ENV NEXT_PUBLIC_LCV_RPC_URL=$NEXT_PUBLIC_LCV_RPC_URL
RUN npm run build

FROM node:20-alpine
WORKDIR /app
ENV NODE_ENV=production
COPY --from=build /app/public ./public
COPY --from=build /app/.next/standalone ./
COPY --from=build /app/.next/static ./.next/static
EXPOSE 3000
CMD ["node", "server.js"]
