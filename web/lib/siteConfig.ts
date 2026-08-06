export const siteConfig = {
  name: "LastcoinVision",
  ticker: "LCV",
  domain: "lastcoinvision.com",
  url: "https://lastcoinvision.com",
  description:
    "LastcoinVision (LCV) is a government-grade, developer-first Layer-1 blockchain: auditable modules for voting, tax, land registry, and identity, plus an EVM- and CosmWasm-compatible platform for building on top of them.",
  nav: [
    { href: "/developers", label: "Developers" },
    { href: "/governments", label: "Governments" },
    { href: "/technology", label: "Technology" },
    { href: "/token", label: "Token" },
    { href: "/roadmap", label: "Roadmap" },
    { href: "/blog", label: "Blog" },
  ],
  footerLinks: {
    Product: [
      { href: "/developers", label: "Developers" },
      { href: "/governments", label: "Governments" },
      { href: "/technology", label: "Technology" },
      { href: "/token", label: "Token" },
    ],
    Resources: [
      { href: "/roadmap", label: "Roadmap" },
      { href: "/blog", label: "Blog" },
      { href: "/brand", label: "Brand assets" },
      { href: "https://docs.lastcoinvision.com", label: "Documentation" },
    ],
    Network: [
      { href: "https://faucet.lastcoinvision.com", label: "Testnet faucet" },
      { href: "https://explorer.lastcoinvision.com", label: "Explorer" },
    ],
    Legal: [{ href: "/privacy", label: "Privacy Policy" }],
  },
  chainIds: {
    mainnet: "lastcoinvision-1",
    testnet: "lcv-testnet-1",
    devnet: "lcv-devnet",
  },
  endpoints: {
    rpc: "http://localhost:26657",
    rest: "http://localhost:1317",
    grpc: "localhost:9090",
    evmRpc: "http://localhost:8545",
  },
  testnetEndpoints: {
    rpc: "https://rpc.lastcoinvision.com",
    rest: "https://api.lastcoinvision.com",
  },
} as const;
