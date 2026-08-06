export type RoadmapStatus = "shipped" | "in-progress" | "planned";

export interface RoadmapItem {
  id: string;
  title: string;
  summary: string;
  detail: string;
  status: RoadmapStatus;
}

// Mirrors this project's actual build phases — no calendar-date
// commitments, since those aren't ours to promise. Status reflects what's
// actually been built and verified, not aspiration.
export const roadmap: RoadmapItem[] = [
  {
    id: "phase-a",
    title: "Chain foundation",
    summary: "Cosmos SDK base chain, staking, EIP-1559-style fee market with partial burn.",
    detail:
      "lcvd daemon on Cosmos SDK + CometBFT, IBC-enabled, with a 10-year decaying emission schedule funding staking rewards and a governance-tunable dynamic base fee.",
    status: "shipped",
  },
  {
    id: "phase-b",
    title: "Identity (x/identity)",
    summary: "W3C-style DIDs and verifiable-credential hashes at the protocol level.",
    detail:
      "did:lcv:<address> identifiers, governance-authorized credential issuers, and account tiers (pseudonymous / attested / institutional) — personal data never touches the chain, only hashes and revocation status.",
    status: "shipped",
  },
  {
    id: "phase-c",
    title: "Government modules",
    summary: "Asset registry, commit-reveal elections, and split-payment tax routing.",
    detail:
      "x/registry (land/asset records), x/voting (eligibility-gated commit-reveal elections), and x/taxrail (atomic tax-and-remainder transfers) — each with on-chain invariants and a published SPEC.md, including honest limitations.",
    status: "shipped",
  },
  {
    id: "phase-d",
    title: "Smart contract layers",
    summary: "CosmWasm and EVM compatibility, with reference contracts for both.",
    detail:
      "Governance-gated CosmWasm code upload, an Ethereum-JSON-RPC-compatible EVM with a native ulcv bridge, and reference contracts (escrow, invoice factoring, consent registry, procurement workflows).",
    status: "shipped",
  },
  {
    id: "phase-e",
    title: "Developer platform",
    summary: "TypeScript SDK, create-lcv-app scaffolder, and a rate-limited faucet.",
    detail:
      "Typed clients for every custom module, Keplr/MetaMask wallet helpers, four starter templates that run against a devnet out of the box, and a testnet faucet with per-address and per-IP limits.",
    status: "shipped",
  },
  {
    id: "phase-f",
    title: "Public site",
    summary: "This website — institutional-facing documentation of what's been built.",
    detail: "Developer and government-facing overviews, live devnet stats, token/supply transparency, and brand assets.",
    status: "shipped",
  },
  {
    id: "phase-g",
    title: "Documentation site",
    summary: "A dedicated docs.lastcoinvision.com with tutorials verified against a live devnet.",
    detail: "Module reference generated from each SPEC.md, a validator handbook, and end-to-end tutorials that are actually run, not just written.",
    status: "shipped",
  },
  {
    id: "phase-h",
    title: "Public infrastructure",
    summary: "CI/CD, a full docker-compose stack, and a live public testnet.",
    detail: "Automated builds and nightly devnet runs, a block explorer, and a running testnet (lcv-testnet-1) with public RPC/REST, a faucet, and a verified sentry/validator topology ready for real operators.",
    status: "shipped",
  },
  {
    id: "phase-i",
    title: "Security review",
    summary: "Static analysis, fuzzing, and a documented audit scope ahead of any public launch.",
    detail: "gosec/staticcheck/cargo-audit/npm-audit/Slither passes, fuzz tests on the highest-risk message handlers, and a public SECURITY.md and AUDIT_SCOPE.md.",
    status: "shipped",
  },
  {
    id: "phase-j",
    title: "Mainnet preparation",
    summary: "A professional security audit (in progress), a real independent validator set, and real custody — the gates before any launch beyond testnet.",
    detail:
      "lcv-testnet-1 is live and producing blocks with two independently-operated validators, public RPC/REST, a faucet, and a verified sentry-node topology ready to onboard more real operators. A paid external security audit is underway: LastcoinVision has engaged AftPak.com (Al Mahdi Fast Technologies Pvt. Ltd.) for an ongoing security-audit engagement; an initial assessment has been received, with a fuller audit report and recommendations expected shortly. The audit isn't complete yet, so it isn't a closed gate — alongside it, a larger and more economically diverse validator set, a real multisig ceremony for the largest token allocations, and legal/regulatory review remain, none of which have a launch date yet.",
    status: "in-progress",
  },
];
