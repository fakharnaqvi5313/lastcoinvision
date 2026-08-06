import { invoke } from "@tauri-apps/api/core";

/**
 * Thin wrapper over the Rust `keychain_*` commands (src-tauri/src/lib.rs).
 * Only ever stores/retrieves the ENCRYPTED keystore JSON string produced
 * by `@lastcoinvision/wallet-core`'s `encryptKeystore` -- never a raw
 * mnemonic or private key. Backed by the OS-native secret store (macOS
 * Keychain / Windows Credential Manager / Linux Secret Service) via the
 * Rust `keyring` crate, not a plain file on disk.
 */
const SERVICE = "com.lastcoinvision.wallet";
const ACCOUNT = "primary";

export async function saveEncryptedKeystore(json: string): Promise<void> {
  await invoke("keychain_set", { service: SERVICE, account: ACCOUNT, secret: json });
}

export async function loadEncryptedKeystore(): Promise<string | null> {
  return invoke<string | null>("keychain_get", { service: SERVICE, account: ACCOUNT });
}

export async function deleteEncryptedKeystore(): Promise<void> {
  await invoke("keychain_delete", { service: SERVICE, account: ACCOUNT });
}
