import { defineConfig } from "tsup";

export default defineConfig({
  entry: ["src/index.ts"],
  format: ["esm", "cjs"],
  dts: true,
  splitting: false,
  sourcemap: true,
  clean: true,
  // ts-proto's generated code imports "protobufjs/minimal" with no
  // extension. tsup/esbuild leaves third-party deps as external bare
  // imports by default, which CJS require() tolerates but Node's (and
  // Turbopack's) strict ESM resolver rejects — surfaced when a consumer
  // actually imports the ESM build under a strict resolver, not before.
  // Bundling protobufjs in removes the external specifier entirely.
  noExternal: ["protobufjs"],
});
