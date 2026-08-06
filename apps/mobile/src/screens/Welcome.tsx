import { Text, TouchableOpacity, View } from "react-native";

import { styles } from "../lib/theme";

export default function Welcome({ onCreate, onImport }: { onCreate: () => void; onImport: () => void }) {
  return (
    <View style={styles.screen}>
      <View style={styles.card}>
        <Text style={styles.title}>LastcoinVision Wallet</Text>
        <Text style={styles.body}>
          Set up a new wallet, or import one from an existing recovery phrase. Everything happens on this device —
          nothing is sent anywhere until you sign and broadcast a transaction yourself.
        </Text>
        <View style={styles.buttonRow}>
          <TouchableOpacity style={styles.primaryButton} onPress={onCreate}>
            <Text style={styles.primaryButtonText}>Create a new wallet</Text>
          </TouchableOpacity>
          <TouchableOpacity style={styles.secondaryButton} onPress={onImport}>
            <Text style={styles.secondaryButtonText}>Import existing</Text>
          </TouchableOpacity>
        </View>
      </View>
    </View>
  );
}
