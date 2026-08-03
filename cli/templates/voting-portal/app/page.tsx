"use client";

import { useState } from "react";

import { votingClient } from "../lib/voting";

export default function Home() {
  const [electionId, setElectionId] = useState("");
  const [result, setResult] = useState<string>();
  const [error, setError] = useState<string>();

  const [authorities, setAuthorities] = useState<string>();
  const [authoritiesError, setAuthoritiesError] = useState<string>();

  async function lookupElection() {
    setError(undefined);
    setResult(undefined);
    try {
      const [election, results] = await Promise.all([
        votingClient.election(electionId),
        votingClient.electionResults(electionId),
      ]);
      setResult(JSON.stringify({ election, results }, null, 2));
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    }
  }

  async function loadAuthorities() {
    setAuthoritiesError(undefined);
    setAuthorities(undefined);
    try {
      const list = await votingClient.electoralAuthorityList();
      setAuthorities(JSON.stringify(list, null, 2));
    } catch (err) {
      setAuthoritiesError(err instanceof Error ? err.message : String(err));
    }
  }

  return (
    <main style={{ maxWidth: 720, margin: "4rem auto", fontFamily: "system-ui, sans-serif", padding: "0 1rem" }}>
      <h1>LCV Voting Portal</h1>
      <p>Read-only lookups against x/voting over its REST API — no wallet needed.</p>

      <section style={{ marginTop: "2rem" }}>
        <h2>Election lookup</h2>
        <input
          placeholder="Election ID"
          value={electionId}
          onChange={(e) => setElectionId(e.target.value)}
          style={{ padding: "0.4rem", marginRight: "0.5rem", width: 220 }}
        />
        <button onClick={lookupElection} disabled={!electionId}>
          Look up
        </button>
        {error && <p style={{ color: "crimson" }}>{error}</p>}
        {result && <pre>{result}</pre>}
      </section>

      <section style={{ marginTop: "2rem" }}>
        <h2>Electoral authorities</h2>
        <button onClick={loadAuthorities}>List authorities</button>
        {authoritiesError && <p style={{ color: "crimson" }}>{authoritiesError}</p>}
        {authorities && <pre>{authorities}</pre>}
      </section>
    </main>
  );
}
