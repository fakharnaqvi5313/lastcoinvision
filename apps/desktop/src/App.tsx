import { useEffect, useState } from "react";

import CreateWallet from "./components/CreateWallet.js";
import Dashboard from "./components/Dashboard.js";
import ImportWallet from "./components/ImportWallet.js";
import Unlock from "./components/Unlock.js";
import Welcome from "./components/Welcome.js";
import { DEFAULT_NETWORK, NETWORKS } from "./lib/network.js";
import { SessionProvider, useSession } from "./lib/session.js";

type Screen = "welcome" | "create" | "import" | "unlock" | "dashboard";

function Router() {
  const { hasStoredKeystore, wallet, refreshHasStoredKeystore } = useSession();
  const [checked, setChecked] = useState(false);
  const [screen, setScreen] = useState<Screen>("welcome");

  useEffect(() => {
    void refreshHasStoredKeystore().then(() => setChecked(true));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (!checked) return;
    setScreen(hasStoredKeystore ? "unlock" : "welcome");
  }, [checked, hasStoredKeystore]);

  if (wallet) {
    return <Dashboard />;
  }

  if (!checked) {
    return (
      <div className="card">
        <p className="muted">Loading…</p>
      </div>
    );
  }

  switch (screen) {
    case "welcome":
      return <Welcome onCreate={() => setScreen("create")} onImport={() => setScreen("import")} />;
    case "create":
      return <CreateWallet onDone={() => setScreen("dashboard")} onBack={() => setScreen("welcome")} />;
    case "import":
      return <ImportWallet onDone={() => setScreen("dashboard")} onBack={() => setScreen("welcome")} />;
    case "unlock":
      return <Unlock onForgotten={() => setScreen("welcome")} />;
    case "dashboard":
      return <Dashboard />;
  }
}

export default function App() {
  return (
    <SessionProvider defaultNetwork={NETWORKS[DEFAULT_NETWORK]}>
      <Router />
    </SessionProvider>
  );
}
