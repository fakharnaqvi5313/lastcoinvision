import { defineConfig } from "tsup";

export default defineConfig({
  entry: ["src/index.ts"],
  format: ["esm", "cjs"],
  dts: true,
  splitting: false,
  sourcemap: true,
  clean: true,
  // Same reasoning as sdk/tsup.config.ts and wallet-core/tsup.config.ts.
  noExternal: ["protobufjs"],
});
