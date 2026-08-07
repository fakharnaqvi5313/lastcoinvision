import { LcvTxClients } from "@lastcoinvision/tx-builders";
import { LcvWallet } from "@lastcoinvision/wallet-core";
import type { ReactNode } from "react";
import { createContext, useCallback, useContext, useMemo, useState } from "react";

import { deleteEncryptedKeystore, loadEncryptedKeystore, saveEncryptedKeystore } from "./keychain.js";
import type { NetworkConfig } from "./network.js";

/**
 * The decrypted `LcvWallet` and its connected `LcvTxClients` live ONLY in
 * this React state (JS heap, this session) -- never written to disk.
 * What's persisted (via keychain.ts, into the OS secret store) is only the
 * output of `wallet.toKeystore(password)`, an encrypted blob. Locking the
 * wallet just drops this state; nothing needs to be securely wiped from
 * disk because nothing sensitive was ever written there.
 */
interface SessionState {
  readonly hasStoredKeystore: boolean;
  readonly wallet: LcvWallet | null;
  readonly txClients: LcvTxClients | null;
  readonly cosmosAddress: string | null;
  readonly network: NetworkConfig;
  refreshHasStoredKeystore(): Promise<void>;
  createWallet(password: string, cosmosPrefix: string): Promise<{ wallet: LcvWallet; mnemonic: string }>;
  activateWallet(w: LcvWallet): Promise<void>;
  importWallet(mnemonic: string, password: string, cosmosPrefix: string): Promise<void>;
  unlock(password: string, cosmosPrefix: string): Promise<void>;
  revealMnemonicWithPassword(password: string, cosmosPrefix: string): Promise<string>;
  lock(): void;
  forgetWallet(): Promise<void>;
  connectTxClients(network: NetworkConfig): Promise<void>;
  setNetwork(network: NetworkConfig): void;
}

const SessionContext = createContext<SessionState | null>(null);

export function useSession(): SessionState {
  const ctx = useContext(SessionContext);
  if (!ctx) throw new Error("useSession must be used within SessionProvider");
  return ctx;
}

export function SessionProvider({
  children,
  defaultNetwork,
}: {
  children: ReactNode;
  defaultNetwork: NetworkConfig;
}) {
  const [hasStoredKeystore, setHasStoredKeystore] = useState(false);
  const [wallet, setWallet] = useState<LcvWallet | null>(null);
  const [txClients, setTxClients] = useState<LcvTxClients | null>(null);
  const [cosmosAddress, setCosmosAddress] = useState<string | null>(null);
  const [network, setNetworkState] = useState<NetworkConfig>(defaultNetwork);

  const refreshHasStoredKeystore = useCallback(async () => {
    const existing = await loadEncryptedKeystore();
    setHasStoredKeystore(existing !== null);
  }, []);

  const applyWallet = useCallback(async (w: LcvWallet) => {
    setWallet(w);
    setCosmosAddress(await w.cosmosAddress());
    setTxClients(null); // require an explicit connect once the caller knows which network
  }, []);

  const createWallet = useCallback(async (password: string, cosmosPrefix: string) => {
    const w = await LcvWallet.create({ cosmosPrefix });
    const keystore = await w.toKeystore(password);
    await saveEncryptedKeystore(JSON.stringify(keystore));
    setHasStoredKeystore(true);
    // Deliberately does NOT call applyWallet()/setWallet() here. Router
    // treats a non-null `wallet` as "show the dashboard," so activating the
    // session before the caller has shown+confirmed the mnemonic backup
    // screen would unmount that screen before it ever renders. The caller
    // (CreateWallet) must show this mnemonic once for backup, then call
    // activateWallet() only once the user has confirmed the backup.
    return { wallet: w, mnemonic: w.revealMnemonic() };
  }, []);

  const activateWallet = useCallback(
    async (w: LcvWallet) => {
      await applyWallet(w);
    },
    [applyWallet],
  );

  const importWallet = useCallback(
    async (mnemonic: string, password: string, cosmosPrefix: string) => {
      const w = await LcvWallet.fromMnemonic(mnemonic, { cosmosPrefix });
      const keystore = await w.toKeystore(password);
      await saveEncryptedKeystore(JSON.stringify(keystore));
      await applyWallet(w);
      setHasStoredKeystore(true);
    },
    [applyWallet],
  );

  const unlock = useCallback(
    async (password: string, cosmosPrefix: string) => {
      const raw = await loadEncryptedKeystore();
      if (!raw) throw new Error("No stored wallet to unlock");
      const keystore = JSON.parse(raw);
      const w = await LcvWallet.fromKeystore(keystore, password, { cosmosPrefix }); // throws on wrong password
      await applyWallet(w);
    },
    [applyWallet],
  );

  const revealMnemonicWithPassword = useCallback(async (password: string, cosmosPrefix: string) => {
    const raw = await loadEncryptedKeystore();
    if (!raw) throw new Error("No stored wallet on this device");
    const keystore = JSON.parse(raw);
    // Re-derives from the stored keystore with the just-entered password,
    // deliberately not trusting that the session is already unlocked --
    // this is the one re-authentication checkpoint before showing the
    // mnemonic again, so a walked-away unlocked device can't be used to
    // pull it up with no further proof of the password.
    const w = await LcvWallet.fromKeystore(keystore, password, { cosmosPrefix }); // throws on wrong password
    return w.revealMnemonic();
  }, []);

  const lock = useCallback(() => {
    setWallet(null);
    setTxClients(null);
    setCosmosAddress(null);
  }, []);

  const forgetWallet = useCallback(async () => {
    await deleteEncryptedKeystore();
    lock();
    setHasStoredKeystore(false);
  }, [lock]);

  const connectTxClients = useCallback(
    async (net: NetworkConfig) => {
      if (!wallet) throw new Error("Unlock the wallet before connecting");
      const clients = await LcvTxClients.connect(wallet, {
        rpcEndpoint: net.rpcEndpoint,
        restEndpoint: net.restEndpoint,
      });
      setTxClients(clients);
    },
    [wallet],
  );

  const setNetwork = useCallback((net: NetworkConfig) => {
    setNetworkState(net);
    setTxClients(null); // force reconnect on the new network
  }, []);

  const value = useMemo<SessionState>(
    () => ({
      hasStoredKeystore,
      wallet,
      txClients,
      cosmosAddress,
      network,
      refreshHasStoredKeystore,
      createWallet,
      activateWallet,
      importWallet,
      unlock,
      revealMnemonicWithPassword,
      lock,
      forgetWallet,
      connectTxClients,
      setNetwork,
    }),
    [
      hasStoredKeystore,
      wallet,
      txClients,
      cosmosAddress,
      network,
      refreshHasStoredKeystore,
      createWallet,
      activateWallet,
      importWallet,
      unlock,
      revealMnemonicWithPassword,
      lock,
      forgetWallet,
      connectTxClients,
      setNetwork,
    ],
  );

  return <SessionContext.Provider value={value}>{children}</SessionContext.Provider>;
}
