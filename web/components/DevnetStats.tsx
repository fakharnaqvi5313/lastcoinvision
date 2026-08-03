"use client";

import { useEffect, useState } from "react";

import { fetchDevnetStats, type DevnetStats } from "@/lib/devnetStats";

type State = { status: "loading" } | { status: "ok"; stats: DevnetStats } | { status: "unavailable" };

export function DevnetStatsStrip() {
  const [state, setState] = useState<State>({ status: "loading" });

  useEffect(() => {
    let cancelled = false;
    fetchDevnetStats().then((stats) => {
      if (cancelled) return;
      setState(stats ? { status: "ok", stats } : { status: "unavailable" });
    });
    return () => {
      cancelled = true;
    };
  }, []);

  if (state.status === "unavailable") {
    return (
      <div className="rounded-lg border border-dashed border-navy-200 bg-navy-50/50 px-4 py-3 text-center text-sm text-navy-500 dark:border-navy-700 dark:bg-navy-900/50 dark:text-navy-400">
        Live devnet stats unavailable right now — run <code className="rounded bg-navy-100 px-1 dark:bg-navy-800">make devnet</code> locally to see this strip populate.
      </div>
    );
  }

  const stats = state.status === "ok" ? state.stats : undefined;

  const items = [
    { label: "Chain ID", value: stats?.chainId ?? "—" },
    { label: "Block height", value: stats ? stats.blockHeight.toLocaleString() : "—" },
    { label: "Validators", value: stats ? stats.validatorCount.toString() : "—" },
  ];

  return (
    <div className="grid grid-cols-3 divide-x divide-navy-100 rounded-lg border border-navy-100 bg-white/60 dark:divide-navy-800 dark:border-navy-800 dark:bg-navy-900/40">
      {items.map((item) => (
        <div key={item.label} className="px-4 py-4 text-center sm:px-6">
          <div className="text-[11px] font-medium uppercase tracking-wider text-navy-500 dark:text-navy-400">{item.label}</div>
          <div className="mt-1 font-mono text-lg font-semibold text-navy-900 dark:text-white">
            {state.status === "loading" ? <span className="animate-pulse text-navy-300 dark:text-navy-600">···</span> : item.value}
          </div>
        </div>
      ))}
    </div>
  );
}
