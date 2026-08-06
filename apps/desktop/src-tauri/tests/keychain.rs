// Real integration test against this machine's actual OS keychain (macOS
// Keychain here) -- not mocked. Exercises the exact `keyring::Entry` calls
// `keychain_set`/`keychain_get`/`keychain_delete` in src/lib.rs wrap,
// proving the OS-native secret storage path genuinely works, since the
// GUI itself couldn't be driven in this environment to click through it.
use keyring::Entry;

const SERVICE: &str = "com.lastcoinvision.wallet.test";

#[test]
fn set_get_delete_round_trip() {
    let account = "integration-test-account";
    let entry = Entry::new(SERVICE, account).expect("failed to open keychain entry");

    // Start clean in case a previous failed run left an entry behind.
    let _ = entry.delete_credential();

    let secret = r#"{"version":1,"kdf":"argon2id","ciphertext":"not-a-real-mnemonic-just-a-test-blob"}"#;
    entry.set_password(secret).expect("failed to write to keychain");

    let read_back = entry.get_password().expect("failed to read from keychain");
    assert_eq!(read_back, secret);

    entry.delete_credential().expect("failed to delete from keychain");

    match entry.get_password() {
        Err(keyring::Error::NoEntry) => {} // expected -- fails closed after deletion
        other => panic!("expected NoEntry after deletion, got {:?}", other),
    }
}

#[test]
fn missing_entry_is_reported_distinctly_not_as_a_generic_error() {
    let entry = Entry::new(SERVICE, "an-account-that-was-never-created").unwrap();
    match entry.get_password() {
        Err(keyring::Error::NoEntry) => {}
        other => panic!("expected NoEntry for a never-created account, got {:?}", other),
    }
}
