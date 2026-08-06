import { useState } from "react";
import { Text, TextInput, TouchableOpacity, View } from "react-native";

import { useSession } from "../lib/session";
import { styles } from "../lib/theme";

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
    <View style={styles.screen}>
      <View style={styles.card}>
        <Text style={styles.title}>Unlock wallet</Text>
        <Text style={styles.label}>Password</Text>
        <TextInput
          style={styles.input}
          secureTextEntry
          autoFocus
          value={password}
          onChangeText={setPassword}
          onSubmitEditing={handleUnlock}
        />
        {error && <Text style={styles.error}>{error}</Text>}
        <View style={styles.buttonRow}>
          <TouchableOpacity
            style={[styles.primaryButton, (busy || !password) && styles.primaryButtonDisabled]}
            onPress={handleUnlock}
            disabled={busy || !password}
          >
            <Text style={styles.primaryButtonText}>{busy ? "Unlocking…" : "Unlock"}</Text>
          </TouchableOpacity>
        </View>

        <Text style={[styles.muted, { marginTop: 24, marginBottom: 8 }]}>
          Forgot your password? The only way back in is your recovery phrase — there is no password reset.
        </Text>
        {!confirmForget ? (
          <TouchableOpacity style={styles.secondaryButton} onPress={() => setConfirmForget(true)}>
            <Text style={styles.secondaryButtonText}>Remove this wallet from this device</Text>
          </TouchableOpacity>
        ) : (
          <View>
            <Text style={styles.warning}>
              This removes the encrypted wallet from this device only. You'll need your recovery phrase to bring it
              back — this doesn't affect funds on-chain.
            </Text>
            <View style={styles.buttonRow}>
              <TouchableOpacity style={styles.primaryButton} onPress={handleForget}>
                <Text style={styles.primaryButtonText}>Yes, remove it</Text>
              </TouchableOpacity>
              <TouchableOpacity style={styles.secondaryButton} onPress={() => setConfirmForget(false)}>
                <Text style={styles.secondaryButtonText}>Cancel</Text>
              </TouchableOpacity>
            </View>
          </View>
        )}
      </View>
    </View>
  );
}
