"use client";

import { connectKeplr } from "@lastcoinvision/sdk";
import { useState } from "react";

import { CHAIN_INFO } from "../lib/chain";

export default function Home() {
  const [address, setAddress] = useState<string>();
  const [balance, setBalance] = useState<string>();
  const [error, setError] = useState<string>();
  const [connecting, setConnecting] = useState(false);

  async function handleConnect() {
    setError(undefined);
    if (!window.keplr) {
      setError("Keplr extension not found. Install it from https://www.keplr.app/.");
      return;
    }
    setConnecting(true);
    try {
      const { signer, client } = await connectKeplr(window.keplr, CHAIN_INFO);
      const [account] = await signer.getAccounts();
      if (!account) {
        throw new Error("Keplr returned no accounts.");
      }
      setAddress(account.address);
      const coin = await client.getBalance(account.address, "ulcv");
      setBalance(`${coin.amount} ${coin.denom}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setConnecting(false);
    }
  }

  return (
    <main style={{ maxWidth: 560, margin: "4rem auto", fontFamily: "system-ui, sans-serif", padding: "0 1rem" }}>
      <h1>LCV dApp (Cosmos)</h1>
      <p>Connects Keplr to LastcoinVision and displays the account balance.</p>

      {!address ? (
        <button onClick={handleConnect} disabled={connecting}>
          {connecting ? "Connecting..." : "Connect Keplr"}
        </button>
      ) : (
        <div>
          <p>
            Address: <code>{address}</code>
          </p>
          <p>Balance: {balance ?? "loading..."}</p>
        </div>
      )}

      {error && <p style={{ color: "crimson" }}>{error}</p>}
    </main>
  );
}
