import { defineConfig } from "tsup";

export default defineConfig({
  entry: ["src/index.ts"],
  format: ["esm", "cjs"],
  dts: true,
  splitting: false,
  sourcemap: true,
  clean: true,
  // Same reasoning as sdk/tsup.config.ts: protobufjs (a transitive dep via
  // @cosmjs/proto-signing) is imported extension-less by generated code,
  // which trips strict ESM resolvers unless bundled in directly.
  noExternal: ["protobufjs"],
});
