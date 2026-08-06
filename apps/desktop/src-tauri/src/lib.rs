// The only job this Rust layer has for the wallet is OS-native secret
// storage. It never sees a raw mnemonic or private key -- only the
// encrypted keystore JSON string produced by
// `@lastcoinvision/wallet-core`'s `encryptKeystore`. Encryption/decryption
// happens in the webview's JS (wallet-core is browser-compatible; Tauri's
// webview has no Node.js integration by default, unlike Electron's
// renderer, so this isn't the same risk profile that framing implies).

use keyring::Entry;

fn open_entry(service: &str, account: &str) -> Result<Entry, String> {
    Entry::new(service, account).map_err(|e| e.to_string())
}

#[tauri::command]
fn keychain_set(service: String, account: String, secret: String) -> Result<(), String> {
    let entry = open_entry(&service, &account)?;
    entry.set_password(&secret).map_err(|e| e.to_string())
}

#[tauri::command]
fn keychain_get(service: String, account: String) -> Result<Option<String>, String> {
    let entry = open_entry(&service, &account)?;
    match entry.get_password() {
        Ok(secret) => Ok(Some(secret)),
        Err(keyring::Error::NoEntry) => Ok(None),
        Err(e) => Err(e.to_string()),
    }
}

#[tauri::command]
fn keychain_delete(service: String, account: String) -> Result<(), String> {
    let entry = open_entry(&service, &account)?;
    match entry.delete_credential() {
        Ok(()) | Err(keyring::Error::NoEntry) => Ok(()),
        Err(e) => Err(e.to_string()),
    }
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .setup(|app| {
            if cfg!(debug_assertions) {
                app.handle().plugin(
                    tauri_plugin_log::Builder::default()
                        .level(log::LevelFilter::Info)
                        .build(),
                )?;
            }
            Ok(())
        })
        .invoke_handler(tauri::generate_handler![
            keychain_set,
            keychain_get,
            keychain_delete
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
