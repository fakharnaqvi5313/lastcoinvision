import type { Metadata } from "next";
import Image from "next/image";

export const metadata: Metadata = {
  title: "Brand",
  description: "Logo downloads and color palette.",
};

const navySwatches = [
  { name: "navy-950", hex: "#081627" },
  { name: "navy-900", hex: "#0E2640" },
  { name: "navy-700", hex: "#153453" },
  { name: "navy-500", hex: "#1F4E79" },
  { name: "navy-300", hex: "#7DABD5" },
  { name: "navy-100", hex: "#D4E3F1" },
];

const accentSwatches = [
  { name: "accent-700", hex: "#0899A5" },
  { name: "accent-500", hex: "#0EEBFD" },
  { name: "accent-300", hex: "#66F2FC" },
  { name: "accent-100", hex: "#CCFBFE" },
];

function Swatch({ name, hex }: { name: string; hex: string }) {
  return (
    <div className="overflow-hidden rounded-lg border border-navy-100 dark:border-navy-800">
      <div className="h-16 w-full" style={{ backgroundColor: hex }} />
      <div className="px-3 py-2 text-xs">
        <div className="font-medium text-navy-900 dark:text-white">{name}</div>
        <div className="font-mono text-navy-500 dark:text-navy-400">{hex}</div>
      </div>
    </div>
  );
}

export default function BrandPage() {
  return (
    <div className="container-page py-16 sm:py-20">
      <div className="max-w-2xl">
        <span className="eyebrow">Brand</span>
        <h1 className="mt-2 text-4xl font-bold text-navy-950 dark:text-white">Logo & palette</h1>
        <p className="mt-4 text-lg text-navy-500 dark:text-navy-300">
          Deep navy and white, with a single cyan accent — used consistently across the chain, SDK, docs, and this
          site.
        </p>
      </div>

      <section className="mt-14">
        <h2 className="text-xl font-semibold text-navy-900 dark:text-white">Logo</h2>
        <div className="card mt-4 flex flex-col items-center gap-6 sm:flex-row sm:items-center sm:justify-between">
          <Image src="/logo-512.png" alt="LastcoinVision logo" width={160} height={160} />
          <div className="flex flex-col gap-2 sm:items-end">
            <a href="/logo.png" download className="btn-secondary">
              Download PNG (2048×2048)
            </a>
            <a href="/logo-512.png" download className="text-sm text-navy-500 underline underline-offset-4 dark:text-navy-400">
              512×512
            </a>
            <a href="/logo-192.png" download className="text-sm text-navy-500 underline underline-offset-4 dark:text-navy-400">
              192×192
            </a>
          </div>
        </div>
        <p className="mt-3 text-sm text-navy-500 dark:text-navy-400">
          Keep clear space around the disc at least a quarter of its diameter. Don&apos;t recolor it, stretch it, or
          place it on a background that competes with the navy disc.
        </p>
      </section>

      <section className="mt-14">
        <h2 className="text-xl font-semibold text-navy-900 dark:text-white">Navy</h2>
        <p className="mt-2 text-sm text-navy-500 dark:text-navy-400">
          The top-level spec calls for deep navy at <code className="rounded bg-navy-100 px-1 dark:bg-navy-800">#1F4E79</code>{" "}
          (navy-500 here); the master logo&apos;s disc background is a darker{" "}
          <code className="rounded bg-navy-100 px-1 dark:bg-navy-800">#0E2640</code> (navy-900). Both are in the same
          family and used together — navy-500 for primary UI elements, navy-900 for dark backgrounds and the logo
          itself.
        </p>
        <div className="mt-4 grid grid-cols-3 gap-3 sm:grid-cols-6">
          {navySwatches.map((s) => (
            <Swatch key={s.name} {...s} />
          ))}
        </div>
      </section>

      <section className="mt-14">
        <h2 className="text-xl font-semibold text-navy-900 dark:text-white">Accent</h2>
        <p className="mt-2 text-sm text-navy-500 dark:text-navy-400">
          One accent, used consistently: the cyan from the logo&apos;s outer ring,{" "}
          <code className="rounded bg-navy-100 px-1 dark:bg-navy-800">#0EEBFD</code>.
        </p>
        <div className="mt-4 grid grid-cols-2 gap-3 sm:grid-cols-4">
          {accentSwatches.map((s) => (
            <Swatch key={s.name} {...s} />
          ))}
        </div>
      </section>
    </div>
  );
}
