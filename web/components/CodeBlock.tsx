"use client";

import { useState } from "react";

export function CodeBlock({ code }: { code: string }) {
  const [copied, setCopied] = useState(false);

  async function handleCopy() {
    try {
      await navigator.clipboard.writeText(code);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {
      // Clipboard API can be unavailable (e.g. insecure context); the code
      // is still selectable/readable, so just skip the copy affordance.
    }
  }

  return (
    <div className="relative overflow-x-auto rounded-lg bg-navy-950 px-4 py-3 font-mono text-sm text-accent-300">
      <button
        type="button"
        onClick={handleCopy}
        className="absolute right-2 top-2 rounded border border-navy-700 bg-navy-900 px-2 py-1 text-xs text-navy-300 hover:text-white"
      >
        {copied ? "Copied" : "Copy"}
      </button>
      <pre className="pr-16">
        <code>{code}</code>
      </pre>
    </div>
  );
}
