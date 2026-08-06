import { useState } from "react";

import { useSession } from "../lib/session.js";

type Step = "password" | "backup";

export default function CreateWallet({ onDone, onBack }: { onDone: () => void; onBack: () => void }) {
  const { createWallet, network } = useSession();
  const [step, setStep] = useState<Step>("password");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [mnemonic, setMnemonic] = useState<string | null>(null);
  const [backedUp, setBackedUp] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function handleCreate() {
    setError(null);
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
      const { mnemonic: newMnemonic } = await createWallet(password, network.cosmosPrefix);
      setMnemonic(newMnemonic);
      setStep("backup");
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setBusy(false);
    }
  }

  if (step === "backup" && mnemonic) {
    return (
      <div className="card">
        <h1>Back up your recovery phrase</h1>
        <p className="warning">
          Write these 12 words down and store them somewhere safe. Anyone with this phrase can control your wallet.
          It is never stored anywhere by this app except as an encrypted keystore behind your password — losing
          this phrase AND forgetting your password means permanent loss of access.
        </p>
        <div className="mnemonic-box">{mnemonic}</div>
        <div className="field" style={{ marginTop: "1rem" }}>
          <label>
            <input type="checkbox" checked={backedUp} onChange={(e) => setBackedUp(e.target.checked)} /> I've
            written this down and stored it safely
          </label>
        </div>
        <div className="btn-row">
          <button className="btn-primary" disabled={!backedUp} onClick={onDone}>
            Continue to wallet
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="card">
      <h1>Create a new wallet</h1>
      <p className="muted">Choose a password to encrypt your wallet on this device.</p>
      <div className="field">
        <label htmlFor="password">Password</label>
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
        <button className="btn-primary" onClick={handleCreate} disabled={busy}>
          {busy ? "Creating…" : "Create wallet"}
        </button>
        <button className="btn-secondary" onClick={onBack} disabled={busy}>
          Back
        </button>
      </div>
    </div>
  );
}
