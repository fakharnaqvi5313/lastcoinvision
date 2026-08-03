"use client";

import { useState } from "react";

import { registryClient } from "../lib/registry";

export default function Home() {
  const [jurisdictionCode, setJurisdictionCode] = useState("");
  const [jurisdictionResult, setJurisdictionResult] = useState<string>();
  const [jurisdictionError, setJurisdictionError] = useState<string>();

  const [assetId, setAssetId] = useState("");
  const [assetResult, setAssetResult] = useState<string>();
  const [assetError, setAssetError] = useState<string>();

  async function lookupJurisdiction() {
    setJurisdictionError(undefined);
    setJurisdictionResult(undefined);
    try {
      const result = await registryClient.jurisdictionAuthorityOf(jurisdictionCode);
      setJurisdictionResult(JSON.stringify(result, null, 2));
    } catch (err) {
      setJurisdictionError(err instanceof Error ? err.message : String(err));
    }
  }

  async function lookupAsset() {
    setAssetError(undefined);
    setAssetResult(undefined);
    try {
      const [asset, history] = await Promise.all([
        registryClient.assetByID(assetId),
        registryClient.transferHistory(assetId),
      ]);
      setAssetResult(JSON.stringify({ asset, history }, null, 2));
    } catch (err) {
      setAssetError(err instanceof Error ? err.message : String(err));
    }
  }

  return (
    <main style={{ maxWidth: 720, margin: "4rem auto", fontFamily: "system-ui, sans-serif", padding: "0 1rem" }}>
      <h1>LCV Registry UI</h1>
      <p>Read-only lookups against x/registry over its REST API — no wallet needed.</p>

      <section style={{ marginTop: "2rem" }}>
        <h2>Jurisdiction authority</h2>
        <input
          placeholder="Jurisdiction code, e.g. US-CA"
          value={jurisdictionCode}
          onChange={(e) => setJurisdictionCode(e.target.value)}
          style={{ padding: "0.4rem", marginRight: "0.5rem", width: 220 }}
        />
        <button onClick={lookupJurisdiction} disabled={!jurisdictionCode}>
          Look up
        </button>
        {jurisdictionError && <p style={{ color: "crimson" }}>{jurisdictionError}</p>}
        {jurisdictionResult && <pre>{jurisdictionResult}</pre>}
      </section>

      <section style={{ marginTop: "2rem" }}>
        <h2>Asset lookup</h2>
        <input
          placeholder="Asset ID"
          value={assetId}
          onChange={(e) => setAssetId(e.target.value)}
          style={{ padding: "0.4rem", marginRight: "0.5rem", width: 220 }}
        />
        <button onClick={lookupAsset} disabled={!assetId}>
          Look up
        </button>
        {assetError && <p style={{ color: "crimson" }}>{assetError}</p>}
        {assetResult && <pre>{assetResult}</pre>}
      </section>
    </main>
  );
}
