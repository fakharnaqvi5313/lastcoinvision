import { accent, navy } from "@lastcoinvision/ui-tokens";
import React from "react";
import ReactDOM from "react-dom/client";

import App from "./App.js";
import "./styles.css";

// Real reuse of ui-tokens, not hand-copied hex values: publish every shade
// as a CSS custom property so styles.css can reference `var(--navy-500)`
// etc. and stay in sync with the shared palette automatically.
for (const [shade, hex] of Object.entries(navy)) {
  document.documentElement.style.setProperty(`--navy-${shade}`, hex);
}
for (const [shade, hex] of Object.entries(accent)) {
  document.documentElement.style.setProperty(`--accent-${shade}`, hex);
}

const rootEl = document.getElementById("root");
if (!rootEl) throw new Error("#root element not found");

ReactDOM.createRoot(rootEl).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
