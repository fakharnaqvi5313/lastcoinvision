# lcvd links CosmWasm's wasmvm, which requires CGO and its own native
# libwasmvm shared library — the runtime image must match the builder's
# glibc/arch, so this stays on debian-slim rather than alpine/musl.
FROM golang:1.25-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -trimpath -o /out/lcvd ./cmd/lcvd
RUN CGO_ENABLED=1 go build -trimpath -o /out/gengenesis ./cmd/gengenesis

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl && rm -rf /var/lib/apt/lists/*
COPY --from=build /out/lcvd /usr/local/bin/lcvd
COPY --from=build /out/gengenesis /usr/local/bin/gengenesis
EXPOSE 26656 26657 1317 9090 8545 8546
ENTRYPOINT ["lcvd"]
