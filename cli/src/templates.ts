export interface TemplateInfo {
  id: string;
  label: string;
  description: string;
}

export const TEMPLATES: readonly TemplateInfo[] = [
  {
    id: "dapp-evm",
    label: "dApp (EVM)",
    description: "Next.js + wagmi/viem, connects MetaMask to LCV's EVM and displays the native balance.",
  },
  {
    id: "dapp-cosmos",
    label: "dApp (Cosmos)",
    description: "Next.js + CosmJS, connects Keplr to LCV and displays the account balance.",
  },
  {
    id: "gov-registry",
    label: "Registry UI",
    description: "Next.js UI over x/registry: look up jurisdiction authorities and registered assets.",
  },
  {
    id: "voting-portal",
    label: "Voting portal",
    description: "Next.js UI over x/voting: look up elections and their commit-reveal results.",
  },
];

export const TEMPLATE_IDS: readonly string[] = TEMPLATES.map((t) => t.id);

export function isValidTemplateId(id: string): boolean {
  return TEMPLATE_IDS.includes(id);
}
