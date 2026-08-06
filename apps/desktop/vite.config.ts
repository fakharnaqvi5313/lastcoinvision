import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

// Tauri-recommended Vite settings: fixed port (the Rust shell points at a
// specific devUrl), don't clear the terminal so cargo's own output stays
// visible, and ignore src-tauri/ in the watcher (Rust rebuilds are cargo's
// job, not Vite's).
export default defineConfig({
  plugins: [react()],
  clearScreen: false,
  server: {
    port: 5183,
    strictPort: true,
    watch: {
      ignored: ["**/src-tauri/**"],
    },
  },
});
