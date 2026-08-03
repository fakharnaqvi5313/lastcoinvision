"use client";

import { Area, AreaChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";

import { buildEmissionCurve } from "@/lib/emission";

export function EmissionChart() {
  const data = buildEmissionCurve(12);

  return (
    <div className="h-72 w-full">
      <ResponsiveContainer width="100%" height="100%">
        <AreaChart data={data} margin={{ top: 8, right: 16, bottom: 0, left: 0 }}>
          <defs>
            <linearGradient id="emissionFill" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor="#0EEBFD" stopOpacity={0.35} />
              <stop offset="100%" stopColor="#0EEBFD" stopOpacity={0} />
            </linearGradient>
          </defs>
          <CartesianGrid strokeDasharray="3 3" className="stroke-navy-100 dark:stroke-navy-800" />
          <XAxis
            dataKey="year"
            tickFormatter={(year: number) => `Y${year}`}
            className="text-xs fill-navy-400"
            tick={{ fill: "currentColor" }}
          />
          <YAxis
            tickFormatter={(value: number) => `${(value / 1_000_000).toFixed(0)}M`}
            className="text-xs fill-navy-400"
            tick={{ fill: "currentColor" }}
            width={56}
          />
          <Tooltip
            formatter={(value) => [`${Math.round(Number(value)).toLocaleString()} LCV`, "Cumulative emitted"]}
            labelFormatter={(year) => `Year ${year}`}
            contentStyle={{ borderRadius: 8, fontSize: 12 }}
          />
          <Area type="monotone" dataKey="cumulativeEmittedLcv" stroke="#0EEBFD" strokeWidth={2} fill="url(#emissionFill)" />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
}
