export interface UseCase {
  id: string;
  title: string;
  description: string;
  tag: string;
}

export const useCases: UseCase[] = [
  {
    id: "voting",
    title: "Voting",
    description: "Eligibility-gated elections with commit-reveal ballots and a publicly recomputable tally.",
    tag: "x/voting",
  },
  {
    id: "tax",
    title: "Tax",
    description: "Split-payment transfers that route a jurisdiction's tax share atomically, with a transparent treasury.",
    tag: "x/taxrail",
  },
  {
    id: "land-registry",
    title: "Land registry",
    description: "Non-fungible asset records with full transfer history and jurisdiction-authority freeze powers.",
    tag: "x/registry",
  },
  {
    id: "identity",
    title: "Identity",
    description: "W3C-style DIDs and credential-hash anchoring — personal data never touches the chain.",
    tag: "x/identity",
  },
  {
    id: "aml",
    title: "AML / compliance",
    description: "Account tiers (pseudonymous / attested / institutional) gate access to credential-verified counterparties.",
    tag: "x/identity",
  },
  {
    id: "erp-crm",
    title: "ERP / CRM",
    description: "Build on CosmWasm or the EVM layer with standard tooling — Hardhat, Foundry, wagmi, CosmJS all work.",
    tag: "EVM + CosmWasm",
  },
  {
    id: "commerce",
    title: "Commerce",
    description: "Buyer-seller and milestone escrow, invoice factoring — reference contracts included, Foundry-tested.",
    tag: "Reference contracts",
  },
  {
    id: "healthcare",
    title: "Healthcare records",
    description: "Consent-registry and access-log patterns for CosmWasm — permissions and audit trail on-chain, records off.",
    tag: "CosmWasm",
  },
];
