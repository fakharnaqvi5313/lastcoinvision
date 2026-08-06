import { useState } from "react";

import { useSession } from "../lib/session.js";

export default function Unlock({ onForgotten }: { onForgotten: () => void }) {
  const { unlock, forgetWallet, network } = useSession();
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [confirmForget, setConfirmForget] = useState(false);

  async function handleUnlock() {
    setError(null);
    setBusy(true);
    try {
      await unlock(password, network.cosmosPrefix);
    } catch {
      setError("Wrong password.");
    } finally {
      setBusy(false);
    }
  }

  async function handleForget() {
    await forgetWallet();
    onForgotten();
  }

  return (
    <div className="card">
      <h1>Unlock wallet</h1>
      <div className="field">
        <label htmlFor="password">Password</label>
        <input
          id="password"
          type="password"
          autoFocus
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter") void handleUnlock();
          }}
        />
      </div>
      {error && <p className="error">{error}</p>}
      <div className="btn-row">
        <button className="btn-primary" onClick={handleUnlock} disabled={busy || !password}>
          {busy ? "Unlocking…" : "Unlock"}
        </button>
      </div>
      <p className="muted" style={{ marginTop: "1.5rem" }}>
        Forgot your password? The only way back in is your recovery phrase — there is no password reset.
      </p>
      {!confirmForget ? (
        <button className="btn-secondary" onClick={() => setConfirmForget(true)}>
          Remove this wallet from this device
        </button>
      ) : (
        <div>
          <p className="warning">
            This removes the encrypted wallet from this device only. You'll need your recovery phrase to bring it
            back — this doesn't affect funds on-chain.
          </p>
          <div className="btn-row">
            <button className="btn-primary" onClick={handleForget}>
              Yes, remove it
            </button>
            <button className="btn-secondary" onClick={() => setConfirmForget(false)}>
              Cancel
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
