export default function Welcome({ onCreate, onImport }: { onCreate: () => void; onImport: () => void }) {
  return (
    <div className="card">
      <h1>LastcoinVision Wallet</h1>
      <p className="muted">
        Set up a new wallet, or import one from an existing recovery phrase. Everything happens on this device —
        nothing is sent anywhere until you sign and broadcast a transaction yourself.
      </p>
      <div className="btn-row">
        <button className="btn-primary" onClick={onCreate}>
          Create a new wallet
        </button>
        <button className="btn-secondary" onClick={onImport}>
          Import existing
        </button>
      </div>
    </div>
  );
}
