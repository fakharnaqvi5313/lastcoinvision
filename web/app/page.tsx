import Link from "next/link";

import { CodeBlock } from "@/components/CodeBlock";
import { DevnetStatsStrip } from "@/components/DevnetStats";
import { RoadmapTimeline } from "@/components/RoadmapTimeline";
import { UseCaseGrid } from "@/components/UseCaseGrid";

export default function Home() {
  return (
    <>
      <section className="border-b border-navy-100 bg-gradient-to-b from-navy-50 to-white dark:border-navy-800 dark:from-navy-900 dark:to-navy-950">
        <div className="container-page py-20 text-center sm:py-28">
          <span className="eyebrow">Cosmos SDK · CometBFT · EVM · CosmWasm</span>
          <h1 className="mx-auto mt-4 max-w-3xl text-4xl font-bold tracking-tight text-navy-950 sm:text-5xl dark:text-white">
            The Layer-1 for Governments and the Developers Who Serve Them
          </h1>
          <p className="mx-auto mt-5 max-w-2xl text-lg text-navy-500 dark:text-navy-300">
            Pre-built, auditable modules for voting, tax, land registry, and identity — plus an EVM- and
            CosmWasm-compatible platform for building ERPs, commerce, and healthcare records on top of them.
          </p>
          <div className="mt-8 flex flex-wrap items-center justify-center gap-3">
            <Link href="/developers" className="btn-primary">
              Start building
            </Link>
            <Link href="/governments" className="btn-secondary">
              For governments
            </Link>
          </div>

          <div className="mx-auto mt-12 max-w-2xl">
            <DevnetStatsStrip />
          </div>
        </div>
      </section>

      <section className="container-page py-16 sm:py-20">
        <div className="mx-auto max-w-2xl text-center">
          <span className="eyebrow">Built for both sides of the network</span>
          <h2 className="mt-2 text-3xl font-bold text-navy-950 dark:text-white">One chain, purpose-built modules</h2>
        </div>
        <div className="mt-10">
          <UseCaseGrid />
        </div>
      </section>

      <section className="border-y border-navy-100 bg-navy-50/50 dark:border-navy-800 dark:bg-navy-900/40">
        <div className="container-page py-16 sm:py-20">
          <div className="grid gap-10 lg:grid-cols-2 lg:items-center">
            <div>
              <span className="eyebrow">Developer quick start</span>
              <h2 className="mt-2 text-3xl font-bold text-navy-950 dark:text-white">Running against a devnet in one command</h2>
              <p className="mt-4 text-navy-600 dark:text-navy-300">
                Four starter templates — EVM, Cosmos, a registry UI, and a voting portal — all connect to a local
                devnet out of the box.
              </p>
              <Link href="/developers" className="mt-6 inline-block text-sm font-semibold text-navy-700 underline underline-offset-4 dark:text-accent-400">
                See the full developer quick start →
              </Link>
            </div>
            <CodeBlock code={"npx create-lcv-app my-app --template dapp-evm\ncd my-app\nnpm run dev"} />
          </div>
        </div>
      </section>

      <section className="container-page py-16 sm:py-20">
        <div className="mx-auto max-w-2xl text-center">
          <span className="eyebrow">Roadmap</span>
          <h2 className="mt-2 text-3xl font-bold text-navy-950 dark:text-white">What&apos;s built, what&apos;s next</h2>
        </div>
        <div className="mx-auto mt-10 max-w-2xl">
          <RoadmapTimeline />
        </div>
        <div className="mt-8 text-center">
          <Link href="/roadmap" className="text-sm font-semibold text-navy-700 underline underline-offset-4 dark:text-accent-400">
            View the full roadmap →
          </Link>
        </div>
      </section>
    </>
  );
}
