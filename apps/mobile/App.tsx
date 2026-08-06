import { useEffect, useState } from "react";
import { StatusBar } from "expo-status-bar";
import { Text, View } from "react-native";

import CreateWallet from "./src/screens/CreateWallet";
import Dashboard from "./src/screens/Dashboard";
import ImportWallet from "./src/screens/ImportWallet";
import Unlock from "./src/screens/Unlock";
import Welcome from "./src/screens/Welcome";
import { DEFAULT_NETWORK, NETWORKS } from "./src/lib/network";
import { SessionProvider, useSession } from "./src/lib/session";
import { styles } from "./src/lib/theme";

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
      <View style={styles.screen}>
        <Text style={styles.muted}>Loading…</Text>
      </View>
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
      <StatusBar style="light" />
      <Router />
    </SessionProvider>
  );
}
