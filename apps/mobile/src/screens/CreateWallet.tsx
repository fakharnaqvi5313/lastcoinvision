import { useState } from "react";
import { Switch, Text, TextInput, TouchableOpacity, View } from "react-native";

import { useSession } from "../lib/session";
import { styles } from "../lib/theme";

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
      <View style={styles.screen}>
        <View style={styles.card}>
          <Text style={styles.title}>Back up your recovery phrase</Text>
          <Text style={styles.warning}>
            Write these 12 words down and store them somewhere safe. Anyone with this phrase can control your
            wallet. It is never stored anywhere by this app except as an encrypted keystore behind your password —
            losing this phrase AND forgetting your password means permanent loss of access.
          </Text>
          <View style={styles.mnemonicBox}>
            <Text style={styles.mnemonicText}>{mnemonic}</Text>
          </View>
          <View style={styles.checkboxRow}>
            <Switch value={backedUp} onValueChange={setBackedUp} />
            <Text style={styles.body}>I've written this down and stored it safely</Text>
          </View>
          <View style={styles.buttonRow}>
            <TouchableOpacity
              style={[styles.primaryButton, !backedUp && styles.primaryButtonDisabled]}
              disabled={!backedUp}
              onPress={onDone}
            >
              <Text style={styles.primaryButtonText}>Continue to wallet</Text>
            </TouchableOpacity>
          </View>
        </View>
      </View>
    );
  }

  return (
    <View style={styles.screen}>
      <View style={styles.card}>
        <Text style={styles.title}>Create a new wallet</Text>
        <Text style={styles.body}>Choose a password to encrypt your wallet on this device.</Text>
        <Text style={styles.label}>Password</Text>
        <TextInput style={styles.input} secureTextEntry value={password} onChangeText={setPassword} />
        <Text style={styles.label}>Confirm password</Text>
        <TextInput style={styles.input} secureTextEntry value={confirmPassword} onChangeText={setConfirmPassword} />
        {error && <Text style={styles.error}>{error}</Text>}
        <View style={styles.buttonRow}>
          <TouchableOpacity
            style={[styles.primaryButton, busy && styles.primaryButtonDisabled]}
            onPress={handleCreate}
            disabled={busy}
          >
            <Text style={styles.primaryButtonText}>{busy ? "Creating…" : "Create wallet"}</Text>
          </TouchableOpacity>
          <TouchableOpacity style={styles.secondaryButton} onPress={onBack} disabled={busy}>
            <Text style={styles.secondaryButtonText}>Back</Text>
          </TouchableOpacity>
        </View>
      </View>
    </View>
  );
}
