import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Technology",
  description: "Architecture, consensus, smart contracts, and an honest account of current limitations.",
};

const architecture = [
  {
    title: "Consensus",
    body: "CometBFT (Tendermint-lineage) BFT consensus with a delegated validator set. Staking rewards come from a genesis-funded pool on a 10-year exponential decay curve, not open-ended inflation.",
  },
  {
    title: "Fee market",
    body: "An EIP-1559-style dynamic base fee, with a governance-tunable split — a portion burned, the remainder to validators and treasury.",
  },
  {
    title: "Interchain",
    body: "IBC-enabled (transfer and interchain accounts) — LCV is a Cosmos SDK chain first, so it inherits the IBC ecosystem rather than requiring a bespoke bridge.",
  },
  {
    title: "Smart contracts",
    body: "Two environments, one chain: CosmWasm (governance-gated code upload, permissionless instantiate) and an Ethereum-JSON-RPC-compatible EVM with a native two-way ulcv bridge. Standard tooling works unmodified — Hardhat, Foundry, wagmi, CosmJS.",
  },
  {
    title: "Identity & privacy",
    body: "W3C-style DIDs (did:lcv:<address>) and credential-hash anchoring by governance-authorized issuers. Personal data never touches the chain — only hashes and revocation status, plus a queryable account tier derived from them.",
  },
];

const limitations = [
  {
    title: "Voting is commit-reveal, not ballot-secret",
    body: "x/voting's commit-reveal scheme lets anyone recompute a tally from public data, but a vote's contents become visible at reveal time to anyone watching the chain during that window. Ballot secrecy in the sense a real election needs — where no one, including validators, can link a revealed vote back to a voter — requires threshold encryption and ZK nullifiers, neither of which is implemented. Do not use this for a real election.",
  },
  {
    title: "Credential verification is signature-based, not zero-knowledge",
    body: "MsgPresentCredential checks a signature against the credential registry today. There's a defined ZKVerifier interface for selective-disclosure proofs, but its current implementation is an explicitly marked stub — no ZK proof system is wired in yet.",
  },
  {
    title: "MetaMask and CLI-generated keys aren't portable by mnemonic",
    body: "The chain's BIP44 coin type is Cosmos's standard 118, unchanged for compatibility with existing Cosmos tooling. An eth_secp256k1 key from lcvd keys add and a MetaMask account don't derive the same address from the same seed phrase — each wallet is its own independent key, not two views of one account.",
  },
  {
    title: "This is testnet/devnet software",
    body: "An independent security audit is in progress but not yet complete — LastcoinVision has engaged AftPak.com (Al Mahdi Fast Technologies Pvt. Ltd.) for an ongoing security-audit engagement, with an initial assessment received and a fuller report expected soon. No production mainnet exists. Nothing on this site is an offer or solicitation — see the token page's disclaimer.",
  },
];

export default function TechnologyPage() {
  return (
    <div className="container-page py-16 sm:py-20">
      <div className="max-w-2xl">
        <span className="eyebrow">Technology</span>
        <h1 className="mt-2 text-4xl font-bold text-navy-950 dark:text-white">How LastcoinVision is built</h1>
        <p className="mt-4 text-lg text-navy-500 dark:text-navy-300">
          A Cosmos SDK chain with a fixed-supply, decaying-emission economy, dual smart-contract environments, and a
          privacy-first identity layer.
        </p>
      </div>

      <section className="mt-14 grid gap-6 sm:grid-cols-2">
        {architecture.map((item) => (
          <div key={item.title} className="card">
            <h2 className="font-semibold text-navy-900 dark:text-white">{item.title}</h2>
            <p className="mt-2 text-sm text-navy-500 dark:text-navy-400">{item.body}</p>
          </div>
        ))}
      </section>

      <section className="mt-16">
        <div className="max-w-2xl">
          <span className="eyebrow">Limitations</span>
          <h2 className="mt-2 text-3xl font-bold text-navy-950 dark:text-white">What this doesn&apos;t do yet</h2>
          <p className="mt-4 text-navy-500 dark:text-navy-300">
            Every module ships a SPEC.md with its own limitations section. The ones that matter most if you&apos;re
            evaluating this for a real deployment:
          </p>
        </div>
        <div className="mt-8 space-y-4">
          {limitations.map((item) => (
            <div key={item.title} className="rounded-lg border border-navy-200 bg-navy-50/40 p-5 dark:border-navy-700 dark:bg-navy-900/40">
              <h3 className="font-semibold text-navy-900 dark:text-white">{item.title}</h3>
              <p className="mt-2 text-sm text-navy-600 dark:text-navy-300">{item.body}</p>
            </div>
          ))}
        </div>
      </section>
    </div>
  );
}
