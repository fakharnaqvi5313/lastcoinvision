import type { Metadata } from "next";
import Link from "next/link";

import { CodeBlock } from "@/components/CodeBlock";
import { siteConfig } from "@/lib/siteConfig";
import { templates } from "@/lib/templates";

export const metadata: Metadata = {
  title: "Developers",
  description: "SDK, CLI templates, faucet, and RPC endpoints for building on LastcoinVision.",
};

export default function DevelopersPage() {
  return (
    <div className="container-page py-16 sm:py-20">
      <div className="max-w-2xl">
        <span className="eyebrow">Developers</span>
        <h1 className="mt-2 text-4xl font-bold text-navy-950 dark:text-white">Build on LastcoinVision</h1>
        <p className="mt-4 text-lg text-navy-500 dark:text-navy-300">
          A typed SDK, four starter templates, and a testnet faucet — everything runs against a local devnet out of
          the box.
        </p>
      </div>

      <section className="mt-12">
        <h2 className="text-xl font-semibold text-navy-900 dark:text-white">Quick start</h2>
        <p className="mt-2 text-sm text-navy-500 dark:text-navy-400">
          Scaffold a template, install, and run — the app connects to a local devnet by default.
        </p>
        <div className="mt-4">
          <CodeBlock code={"npx create-lcv-app my-app --template dapp-evm\ncd my-app\nnpm run dev"} />
        </div>
      </section>

      <section className="mt-14">
        <h2 className="text-xl font-semibold text-navy-900 dark:text-white">Templates</h2>
        <div className="mt-4 grid gap-4 sm:grid-cols-2">
          {templates.map((template) => (
            <div key={template.id} className="card">
              <span className="eyebrow">{template.stack}</span>
              <h3 className="mt-2 font-semibold text-navy-900 dark:text-white">{template.label}</h3>
              <p className="mt-2 text-sm text-navy-500 dark:text-navy-400">{template.description}</p>
              <code className="mt-3 block rounded bg-navy-100 px-2 py-1 text-xs text-navy-700 dark:bg-navy-800 dark:text-navy-200">
                --template {template.id}
              </code>
            </div>
          ))}
        </div>
      </section>

      <section className="mt-14">
        <h2 className="text-xl font-semibold text-navy-900 dark:text-white">SDK</h2>
        <p className="mt-2 max-w-2xl text-sm text-navy-500 dark:text-navy-400">
          <code className="rounded bg-navy-100 px-1 dark:bg-navy-800">@lastcoinvision/sdk</code> ships typed
          message-builder clients for every custom module (identity, registry, voting, taxrail), Keplr and MetaMask
          wallet helpers, and a CosmJS registry pre-loaded with every custom message type.
        </p>
        <div className="mt-4">
          <CodeBlock
            code={`import { IdentityClient, createLcvRegistry } from "@lastcoinvision/sdk";

const identity = new IdentityClient(REST_ENDPOINT, signingClient);
const params = await identity.params();`}
          />
        </div>
        <Link
          href="https://docs.lastcoinvision.com"
          className="mt-4 inline-block text-sm font-semibold text-navy-700 underline underline-offset-4 dark:text-accent-400"
        >
          Full SDK reference →
        </Link>
      </section>

      <section className="mt-14">
        <h2 className="text-xl font-semibold text-navy-900 dark:text-white">Testnet faucet</h2>
        <p className="mt-2 max-w-2xl text-sm text-navy-500 dark:text-navy-400">
          Rate-limited per address and per IP. Fund a devnet/testnet account for development.
        </p>
        <Link href="https://faucet.lastcoinvision.com" className="btn-secondary mt-4">
          Open the faucet
        </Link>
      </section>

      <section className="mt-14">
        <h2 className="text-xl font-semibold text-navy-900 dark:text-white">RPC endpoints</h2>
        <div className="mt-4 overflow-x-auto rounded-lg border border-navy-100 dark:border-navy-800">
          <table className="w-full text-left text-sm">
            <thead className="bg-navy-50 text-navy-500 dark:bg-navy-900 dark:text-navy-400">
              <tr>
                <th className="px-4 py-3 font-medium">Network</th>
                <th className="px-4 py-3 font-medium">Chain ID</th>
                <th className="px-4 py-3 font-medium">Tendermint RPC</th>
                <th className="px-4 py-3 font-medium">REST</th>
                <th className="px-4 py-3 font-medium">EVM JSON-RPC</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-navy-100 dark:divide-navy-800">
              <tr>
                <td className="px-4 py-3 font-medium text-navy-900 dark:text-white">Devnet (local)</td>
                <td className="px-4 py-3 font-mono text-xs">{siteConfig.chainIds.devnet}</td>
                <td className="px-4 py-3 font-mono text-xs">{siteConfig.endpoints.rpc}</td>
                <td className="px-4 py-3 font-mono text-xs">{siteConfig.endpoints.rest}</td>
                <td className="px-4 py-3 font-mono text-xs">{siteConfig.endpoints.evmRpc}</td>
              </tr>
              <tr>
                <td className="px-4 py-3 font-medium text-navy-900 dark:text-white">Testnet</td>
                <td className="px-4 py-3 font-mono text-xs">{siteConfig.chainIds.testnet}</td>
                <td className="px-4 py-3 font-mono text-xs">{siteConfig.testnetEndpoints.rpc}</td>
                <td className="px-4 py-3 font-mono text-xs">{siteConfig.testnetEndpoints.rest}</td>
                <td className="px-4 py-3 text-navy-500 dark:text-navy-400">Not yet public</td>
              </tr>
              <tr>
                <td className="px-4 py-3 font-medium text-navy-900 dark:text-white">Mainnet</td>
                <td className="px-4 py-3 font-mono text-xs">{siteConfig.chainIds.mainnet}</td>
                <td className="px-4 py-3 text-navy-500 dark:text-navy-400" colSpan={3}>
                  Not launched.
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </div>
  );
}
