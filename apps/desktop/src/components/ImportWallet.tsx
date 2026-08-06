import { isValidMnemonic } from "@lastcoinvision/wallet-core";
import { useState } from "react";

import { useSession } from "../lib/session.js";

export default function ImportWallet({ onDone, onBack }: { onDone: () => void; onBack: () => void }) {
  const { importWallet, network } = useSession();
  const [mnemonic, setMnemonic] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function handleImport() {
    setError(null);
    const trimmed = mnemonic.trim().toLowerCase();
    if (!isValidMnemonic(trimmed)) {
      setError("That doesn't look like a valid recovery phrase — check the words and try again.");
      return;
    }
    if (password.length < 8) {
      setError("Use at least 8 characters for the password.");
      return;
    }
    if (password !== confirmPassword) {
      setError("Passwords don't match.");
      return;
    }
    setBusy(true);
    try {
      await importWallet(trimmed, password, network.cosmosPrefix);
      onDone();
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="card">
      <h1>Import existing wallet</h1>
      <div className="field">
        <label htmlFor="mnemonic">Recovery phrase</label>
        <textarea
          id="mnemonic"
          value={mnemonic}
          onChange={(e) => setMnemonic(e.target.value)}
          placeholder="12 or 24 words, separated by spaces"
        />
      </div>
      <div className="field">
        <label htmlFor="password">New password (to encrypt this wallet on this device)</label>
        <input id="password" type="password" value={password} onChange={(e) => setPassword(e.target.value)} />
      </div>
      <div className="field">
        <label htmlFor="confirm">Confirm password</label>
        <input
          id="confirm"
          type="password"
          value={confirmPassword}
          onChange={(e) => setConfirmPassword(e.target.value)}
        />
      </div>
      {error && <p className="error">{error}</p>}
      <div className="btn-row">
        <button className="btn-primary" onClick={handleImport} disabled={busy}>
          {busy ? "Importing…" : "Import wallet"}
        </button>
        <button className="btn-secondary" onClick={onBack} disabled={busy}>
          Back
        </button>
      </div>
    </div>
  );
}
