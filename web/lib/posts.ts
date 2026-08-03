export interface PostMeta {
  slug: string;
  title: string;
  description: string;
  date: string; // ISO date
}

// Kept as a small hand-maintained registry (matching each post.mdx's own
// `metadata` export) rather than pulling in a content-collection library
// for three posts.
export const posts: PostMeta[] = [
  {
    slug: "introducing-lastcoinvision",
    title: "Introducing LastcoinVision",
    description: "What's actually been built so far: chain, identity, government modules, smart contracts, and a developer platform.",
    date: "2026-08-03",
  },
  {
    slug: "why-commit-reveal-voting",
    title: "Why commit-reveal voting (for now)",
    description: "x/voting is verifiable today and not ballot-secret yet — an honest look at what that means and what a production system needs.",
    date: "2026-08-03",
  },
  {
    slug: "identity-without-the-data",
    title: "Identity without the data",
    description: "How x/identity anchors credential hashes and account tiers on-chain without ever storing personal data.",
    date: "2026-08-03",
  },
];
