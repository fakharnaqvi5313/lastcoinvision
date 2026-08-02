#!/usr/bin/env bash
# Builds both CosmWasm contracts to wasm32-unknown-unknown, then post-processes
# the output with wasm-opt to disable the WASM bulk-memory feature.
#
# Why: this toolchain's wasm32-unknown-unknown target enables bulk-memory
# (memory.copy/memory.fill) by default in the prebuilt std/core libraries that
# ship with rustc — a per-crate `-C target-feature=-bulk-memory` rustflag
# (already set in .cargo/config.toml) only affects code this workspace's own
# crates emit, not those prebuilt libraries, so it isn't enough on its own.
# wasmvm's static validator (bundled with wasmd) rejects bytecode using
# bulk-memory outright ("bulk memory support is not enabled") — confirmed
# live via a failed MsgStoreCode governance proposal on a local devnet before
# this script's lowering pass was added. Requires `wasm-opt` (binaryen):
# `brew install binaryen` / `apt install binaryen`.
set -euo pipefail

cd "$(dirname "$0")"

for crate in consent-registry procurement-ledger; do
  echo "==> building $crate"
  (cd "$crate" && cargo build --release --target wasm32-unknown-unknown)

  in_path="$crate/target/wasm32-unknown-unknown/release/$(echo "$crate" | tr '-' '_').wasm"
  out_path="$crate/artifacts"
  mkdir -p "$out_path"

  echo "==> wasm-opt lowering bulk-memory for $crate"
  wasm-opt --enable-bulk-memory --llvm-memory-copy-fill-lowering \
    -o "$out_path/$(basename "$in_path")" "$in_path"

  echo "==> wrote $out_path/$(basename "$in_path")"
done
