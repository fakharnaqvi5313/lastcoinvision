import type { Metadata } from "next";

import { EmissionChart } from "@/components/EmissionChart";
import { HALF_LIFE_YEARS, STAKING_REWARD_POOL_LCV, TOTAL_SUPPLY_LCV, supplyAllocation } from "@/lib/emission";

export const metadata: Metadata = {
  title: "Token",
  description: "LCV supply, allocation, and emission schedule.",
};

function Disclaimer() {
  return (
    <div className="rounded-lg border border-navy-300 bg-navy-50 p-4 text-sm text-navy-700 dark:border-navy-600 dark:bg-navy-900 dark:text-navy-300">
      <strong className="font-semibold text-navy-900 dark:text-white">Informational only.</strong> This page
      describes the design of a testnet/devnet token. It is not an offer or solicitation to buy or sell any
      security or financial instrument, and nothing here is investment advice. Availability of any token, network,
      or service may be restricted or unavailable depending on your jurisdiction.
    </div>
  );
}

export default function TokenPage() {
  return (
    <div className="container-page py-16 sm:py-20">
      <div className="max-w-2xl">
        <span className="eyebrow">Token</span>
        <h1 className="mt-2 text-4xl font-bold text-navy-950 dark:text-white">LCV</h1>
        <p className="mt-4 text-lg text-navy-500 dark:text-navy-300">
          A fixed-supply utility token: gas, staking, and governance for the network — nothing more implied.
        </p>
      </div>

      <div className="mt-8 max-w-2xl">
        <Disclaimer />
      </div>

      <section className="mt-14 grid gap-6 sm:grid-cols-3">
        <div className="card">
          <h2 className="font-semibold text-navy-900 dark:text-white">Gas</h2>
          <p className="mt-2 text-sm text-navy-500 dark:text-navy-400">
            Every transaction — Cosmos message or EVM call — pays fees in ulcv (1 LCV = 1,000,000 ulcv) under a
            governance-tunable, EIP-1559-style dynamic base fee.
          </p>
        </div>
        <div className="card">
          <h2 className="font-semibold text-navy-900 dark:text-white">Staking</h2>
          <p className="mt-2 text-sm text-navy-500 dark:text-navy-400">
            LCV secures the network via delegated proof-of-stake. Rewards are paid from a genesis-funded pool on a
            decaying emission curve, not open-ended inflation.
          </p>
        </div>
        <div className="card">
          <h2 className="font-semibold text-navy-900 dark:text-white">Governance</h2>
          <p className="mt-2 text-sm text-navy-500 dark:text-navy-400">
            Staked LCV votes on parameter changes, credential-issuer authorization, CosmWasm code uploads, and
            software upgrades.
          </p>
        </div>
      </section>

      <section className="mt-16">
        <span className="eyebrow">Supply</span>
        <h2 className="mt-2 text-3xl font-bold text-navy-950 dark:text-white">
          {TOTAL_SUPPLY_LCV.toLocaleString()} LCV, fixed
        </h2>
        <p className="mt-4 max-w-2xl text-navy-500 dark:text-navy-300">
          Total supply is fixed at genesis — there is no minting beyond the allocation table below. Team and
          investor allocations vest linearly with a one-year cliff.
        </p>
        <div className="mt-8 overflow-x-auto rounded-lg border border-navy-100 dark:border-navy-800">
          <table className="w-full text-left text-sm">
            <thead className="bg-navy-50 text-navy-500 dark:bg-navy-900 dark:text-navy-400">
              <tr>
                <th className="px-4 py-3 font-medium">Allocation</th>
                <th className="px-4 py-3 font-medium">Share</th>
                <th className="px-4 py-3 font-medium">LCV</th>
                <th className="px-4 py-3 font-medium">Notes</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-navy-100 dark:divide-navy-800">
              {supplyAllocation.map((row) => (
                <tr key={row.label}>
                  <td className="px-4 py-3 font-medium text-navy-900 dark:text-white">{row.label}</td>
                  <td className="px-4 py-3">{row.percent}%</td>
                  <td className="px-4 py-3 font-mono text-xs">{row.lcv.toLocaleString()}</td>
                  <td className="px-4 py-3 text-navy-500 dark:text-navy-400">{row.notes}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>

      <section className="mt-16">
        <span className="eyebrow">Emission</span>
        <h2 className="mt-2 text-3xl font-bold text-navy-950 dark:text-white">Staking rewards decay over 10 years</h2>
        <p className="mt-4 max-w-2xl text-navy-500 dark:text-navy-300">
          The {STAKING_REWARD_POOL_LCV.toLocaleString()} LCV staking-reward pool (12% of supply) is released on an
          exponential decay curve with a {HALF_LIFE_YEARS}-year half-life — roughly half is emitted in the first two
          years, and the curve asymptotically approaches the full pool rather than an open-ended issuance.
        </p>
        <div className="card mt-8">
          <EmissionChart />
        </div>
      </section>

      <div className="mt-16 max-w-2xl">
        <Disclaimer />
      </div>
    </div>
  );
}
