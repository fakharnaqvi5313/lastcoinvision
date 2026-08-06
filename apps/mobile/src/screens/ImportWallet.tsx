import { isValidMnemonic } from "@lastcoinvision/wallet-core";
import { useState } from "react";
import { Text, TextInput, TouchableOpacity, View } from "react-native";

import { useSession } from "../lib/session";
import { styles } from "../lib/theme";

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
    <View style={styles.screen}>
      <View style={styles.card}>
        <Text style={styles.title}>Import existing wallet</Text>
        <Text style={styles.label}>Recovery phrase</Text>
        <TextInput
          style={[styles.input, styles.textArea]}
          multiline
          value={mnemonic}
          onChangeText={setMnemonic}
          placeholder="12 or 24 words, separated by spaces"
          placeholderTextColor="#5a7a9a"
          autoCapitalize="none"
          autoCorrect={false}
        />
        <Text style={styles.label}>New password (to encrypt this wallet on this device)</Text>
        <TextInput style={styles.input} secureTextEntry value={password} onChangeText={setPassword} />
        <Text style={styles.label}>Confirm password</Text>
        <TextInput style={styles.input} secureTextEntry value={confirmPassword} onChangeText={setConfirmPassword} />
        {error && <Text style={styles.error}>{error}</Text>}
        <View style={styles.buttonRow}>
          <TouchableOpacity
            style={[styles.primaryButton, busy && styles.primaryButtonDisabled]}
            onPress={handleImport}
            disabled={busy}
          >
            <Text style={styles.primaryButtonText}>{busy ? "Importing…" : "Import wallet"}</Text>
          </TouchableOpacity>
          <TouchableOpacity style={styles.secondaryButton} onPress={onBack} disabled={busy}>
            <Text style={styles.secondaryButtonText}>Back</Text>
          </TouchableOpacity>
        </View>
      </View>
    </View>
  );
}
