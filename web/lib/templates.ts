export interface TemplateInfo {
  id: string;
  label: string;
  stack: string;
  description: string;
}

// Mirrors cli/src/templates.ts — kept as a separate, small copy rather
// than a cross-package import so the website has no build-time dependency
// on the create-lcv-app CLI package.
export const templates: TemplateInfo[] = [
  {
    id: "dapp-evm",
    label: "dApp (EVM)",
    stack: "Next.js + wagmi/viem",
    description: "Connects MetaMask to LCV's EVM and displays the native balance.",
  },
  {
    id: "dapp-cosmos",
    label: "dApp (Cosmos)",
    stack: "Next.js + CosmJS",
    description: "Connects Keplr to LCV and displays the account balance.",
  },
  {
    id: "gov-registry",
    label: "Registry UI",
    stack: "Next.js (REST, no wallet)",
    description: "Look up jurisdiction authorities and registered assets on x/registry.",
  },
  {
    id: "voting-portal",
    label: "Voting portal",
    stack: "Next.js (REST, no wallet)",
    description: "Look up elections and their commit-reveal results on x/voting.",
  },
];
