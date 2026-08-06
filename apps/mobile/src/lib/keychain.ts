import * as SecureStore from "expo-secure-store";

/**
 * Thin wrapper over `expo-secure-store`, Expo's first-party wrapper around
 * iOS Keychain / Android Keystore (EncryptedSharedPreferences, hardware-
 * backed where the device supports it). Only ever stores/retrieves the
 * ENCRYPTED keystore JSON string produced by
 * `@lastcoinvision/wallet-core`'s `encryptKeystore` -- never a raw
 * mnemonic or private key. Same role as `apps/desktop/src/lib/keychain.ts`'s
 * Tauri/`keyring`-crate bridge, different platform API.
 */
const KEY = "lcv_wallet_keystore";

export async function saveEncryptedKeystore(json: string): Promise<void> {
  await SecureStore.setItemAsync(KEY, json);
}

export async function loadEncryptedKeystore(): Promise<string | null> {
  return SecureStore.getItemAsync(KEY);
}

export async function deleteEncryptedKeystore(): Promise<void> {
  await SecureStore.deleteItemAsync(KEY);
}
