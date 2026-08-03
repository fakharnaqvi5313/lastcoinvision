import type { Metadata } from "next";

import { ContactForm } from "@/components/ContactForm";
import { GovZoneDiagram } from "@/components/GovZoneDiagram";

export const metadata: Metadata = {
  title: "Governments",
  description: "Auditable, purpose-built modules for voting, tax, land registry, and identity.",
};

const modules = [
  {
    title: "Voting",
    outcome: "Elections that verify themselves",
    body: "Eligibility is gated on a credential type per election, and every vote is a commitment revealed and tallied on-chain — anyone can recompute the result from public data instead of trusting a black-box count.",
    caveat: "Current implementation is commit-reveal, not full ballot secrecy — see the honest limitations on the technology page before considering this for a real election.",
  },
  {
    title: "Land registry",
    outcome: "Title transfers in minutes, not weeks",
    body: "Assets are non-fungible on-chain records with a full, queryable transfer history. Transfers require both parties to hold a verified credential; a jurisdiction authority can freeze a record with an on-chain reason.",
  },
  {
    title: "Tax",
    outcome: "Tax collected automatically, transparently",
    body: "A single transfer atomically routes the jurisdiction's tax share to a collector treasury and the remainder to the recipient — no separate remittance step, and every flow is queryable by anyone.",
  },
  {
    title: "Identity",
    outcome: "Prove eligibility without exposing data",
    body: "Personal data never touches the chain — only credential hashes and revocation status, anchored by governance-authorized issuers. Account tiers let other modules gate access on verified status alone.",
  },
];

export default function GovernmentsPage() {
  return (
    <div className="container-page py-16 sm:py-20">
      <div className="max-w-2xl">
        <span className="eyebrow">For governments</span>
        <h1 className="mt-2 text-4xl font-bold text-navy-950 dark:text-white">Infrastructure, not a pilot project</h1>
        <p className="mt-4 text-lg text-navy-500 dark:text-navy-300">
          Four modules built for the workflows governments actually run — each with a published spec, on-chain
          invariants, and an honest account of what it doesn&apos;t do yet.
        </p>
      </div>

      <section className="mt-14 grid gap-6 sm:grid-cols-2">
        {modules.map((mod) => (
          <div key={mod.title} className="card">
            <span className="eyebrow">{mod.title}</span>
            <h2 className="mt-2 text-lg font-semibold text-navy-900 dark:text-white">{mod.outcome}</h2>
            <p className="mt-2 text-sm text-navy-500 dark:text-navy-400">{mod.body}</p>
            {mod.caveat && <p className="mt-3 text-xs text-navy-500 dark:text-navy-400">⚠ {mod.caveat}</p>}
          </div>
        ))}
      </section>

      <section className="mt-16">
        <div className="max-w-2xl">
          <span className="eyebrow">Architecture</span>
          <h2 className="mt-2 text-3xl font-bold text-navy-950 dark:text-white">The GovZone pattern</h2>
          <p className="mt-4 text-navy-500 dark:text-navy-300">
            Citizen- and agency-facing portals never talk to the chain directly — they go through the SDK, which
            talks to the relevant module, secured by the same validator set as everything else on the network.
          </p>
        </div>
        <div className="card mt-8">
          <GovZoneDiagram />
        </div>
      </section>

      <section className="mt-16 grid gap-10 lg:grid-cols-2">
        <div>
          <span className="eyebrow">Pilot programs</span>
          <h2 className="mt-2 text-3xl font-bold text-navy-950 dark:text-white">Tell us what you&apos;re trying to build</h2>
          <p className="mt-4 text-navy-500 dark:text-navy-300">
            This is testnet/devnet software. We&apos;re looking for agencies and integrators who want to pilot a
            specific workflow — not a general sales inquiry.
          </p>
          <p className="mt-4 text-sm text-navy-500 dark:text-navy-400">
            This form is a development-stage stub: it validates and logs your message server-side, but isn&apos;t
            wired to a live inbox yet.
          </p>
        </div>
        <ContactForm />
      </section>
    </div>
  );
}
