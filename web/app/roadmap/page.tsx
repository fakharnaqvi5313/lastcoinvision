import type { Metadata } from "next";

import { RoadmapTimeline } from "@/components/RoadmapTimeline";

export const metadata: Metadata = {
  title: "Roadmap",
  description: "What's been built, what's in progress, and what's planned — no calendar-date promises.",
};

export default function RoadmapPage() {
  return (
    <div className="container-page py-16 sm:py-20">
      <div className="max-w-2xl">
        <span className="eyebrow">Roadmap</span>
        <h1 className="mt-2 text-4xl font-bold text-navy-950 dark:text-white">Built in the open, phase by phase</h1>
        <p className="mt-4 text-lg text-navy-500 dark:text-navy-300">
          This tracks the project&apos;s actual build phases, not a marketing timeline — status reflects what&apos;s
          been built and verified, not what&apos;s aspirational. No calendar dates, because those aren&apos;t ours to
          promise.
        </p>
      </div>

      <div className="mx-auto mt-14 max-w-2xl">
        <RoadmapTimeline detailed />
      </div>
    </div>
  );
}
