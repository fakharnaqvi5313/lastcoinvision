import { buildEvmWalletClient } from "@lastcoinvision/wallet-core";
import { SigningStargateClient } from "@cosmjs/stargate";
import * as Clipboard from "expo-clipboard";
import { useCallback, useEffect, useState } from "react";
import { ScrollView, Text, TextInput, TouchableOpacity, View } from "react-native";
import { createPublicClient, formatEther, http, parseEther } from "viem";

import { NETWORKS } from "../lib/network";
import type { NetworkConfig } from "../lib/network";
import { useSession } from "../lib/session";
import { styles } from "../lib/theme";

type Tab = "cosmos" | "evm";

export default function Dashboard() {
  const { wallet, cosmosAddress, network, setNetwork, lock } = useSession();
  const [tab, setTab] = useState<Tab>("cosmos");
  const [cosmosBalance, setCosmosBalance] = useState<string | null>(null);
  const [evmBalance, setEvmBalance] = useState<string | null>(null);
  const [balanceError, setBalanceError] = useState<string | null>(null);

  const [sendTo, setSendTo] = useState("");
  const [sendAmount, setSendAmount] = useState("");
  const [sending, setSending] = useState(false);
  const [sendResult, setSendResult] = useState<string | null>(null);
  const [sendError, setSendError] = useState<string | null>(null);

  const refreshBalances = useCallback(async () => {
    if (!wallet || !cosmosAddress) return;
    setBalanceError(null);
    try {
      const res = await fetch(
        `${network.restEndpoint}/cosmos/bank/v1beta1/balances/${cosmosAddress}/by_denom?denom=ulcv`,
      );
      if (res.ok) {
        const data = (await res.json()) as { balance?: { amount: string } };
        setCosmosBalance(data.balance?.amount ?? "0");
      } else {
        setCosmosBalance("0");
      }
    } catch (err) {
      setBalanceError(err instanceof Error ? err.message : String(err));
    }

    if (network.evmRpcUrl) {
      try {
        const publicClient = createPublicClient({ transport: http(network.evmRpcUrl) });
        const wei = await publicClient.getBalance({ address: wallet.evmAddress as `0x${string}` });
        setEvmBalance(formatEther(wei));
      } catch (err) {
        setBalanceError(err instanceof Error ? err.message : String(err));
      }
    } else {
      setEvmBalance(null);
    }
  }, [wallet, cosmosAddress, network]);

  useEffect(() => {
    void refreshBalances();
  }, [refreshBalances]);

  async function handleSendCosmos() {
    if (!wallet || !cosmosAddress) return;
    setSending(true);
    setSendError(null);
    setSendResult(null);
    try {
      const client = await SigningStargateClient.connectWithSigner(network.rpcEndpoint, wallet.cosmosSigner);
      const result = await client.sendTokens(
        cosmosAddress,
        sendTo.trim(),
        [{ denom: "ulcv", amount: sendAmount.trim() }],
        { amount: [{ amount: "5000", denom: "ulcv" }], gas: "200000" }, // fixed fee -- never "auto", see wallet-core/src/fees.ts
        "",
      );
      if (result.code !== 0) {
        setSendError(`Transaction failed (code ${result.code}): ${result.rawLog ?? "unknown error"}`);
      } else {
        setSendResult(result.transactionHash);
        void refreshBalances();
      }
    } catch (err) {
      setSendError(err instanceof Error ? err.message : String(err));
    } finally {
      setSending(false);
    }
  }

  async function handleSendEvm() {
    if (!wallet || !network.evmRpcUrl) return;
    setSending(true);
    setSendError(null);
    setSendResult(null);
    try {
      const client = buildEvmWalletClient(wallet.evmAccount, {
        evmChainId: network.evmChainId,
        chainName: network.name,
        rpcUrl: network.evmRpcUrl,
      });
      const hash = await client.sendTransaction({
        account: wallet.evmAccount,
        to: sendTo.trim() as `0x${string}`,
        value: parseEther(sendAmount.trim()),
        chain: client.chain,
      });
      setSendResult(hash);
      void refreshBalances();
    } catch (err) {
      setSendError(err instanceof Error ? err.message : String(err));
    } finally {
      setSending(false);
    }
  }

  if (!wallet || !cosmosAddress) return null;

  return (
    <ScrollView style={styles.screen} contentContainerStyle={{ justifyContent: "flex-start" }}>
      <View style={styles.card}>
        <Text style={styles.title}>Wallet</Text>

        <Text style={styles.label}>Network</Text>
        <View style={styles.tabs}>
          {Object.values(NETWORKS).map((n: NetworkConfig) => (
            <TouchableOpacity
              key={n.name}
              style={[styles.tab, network.name === n.name && styles.tabActive]}
              onPress={() => setNetwork(n)}
            >
              <Text style={[styles.tabText, network.name === n.name && styles.tabTextActive]}>{n.name}</Text>
            </TouchableOpacity>
          ))}
        </View>

        <TouchableOpacity style={styles.addressRow} onPress={() => Clipboard.setStringAsync(cosmosAddress)}>
          <Text style={styles.addressLabel}>Cosmos address (tap to copy)</Text>
          <Text style={styles.addressValue}>{cosmosAddress}</Text>
          <Text style={styles.muted}>{cosmosBalance !== null ? `${cosmosBalance} ulcv` : "—"}</Text>
        </TouchableOpacity>
        <TouchableOpacity style={styles.addressRow} onPress={() => Clipboard.setStringAsync(wallet.evmAddress)}>
          <Text style={styles.addressLabel}>EVM address (tap to copy)</Text>
          <Text style={styles.addressValue}>{wallet.evmAddress}</Text>
          <Text style={styles.muted}>
            {network.evmRpcUrl
              ? evmBalance !== null
                ? `${evmBalance} LCV`
                : "—"
              : "EVM RPC not available on this network"}
          </Text>
        </TouchableOpacity>
        {balanceError && <Text style={styles.error}>{balanceError}</Text>}

        <View style={styles.tabs}>
          <TouchableOpacity
            style={[styles.tab, tab === "cosmos" && styles.tabActive]}
            onPress={() => setTab("cosmos")}
          >
            <Text style={[styles.tabText, tab === "cosmos" && styles.tabTextActive]}>Send (Cosmos)</Text>
          </TouchableOpacity>
          <TouchableOpacity
            style={[styles.tab, tab === "evm" && styles.tabActive]}
            onPress={() => setTab("evm")}
            disabled={!network.evmRpcUrl}
          >
            <Text style={[styles.tabText, tab === "evm" && styles.tabTextActive]}>Send (EVM)</Text>
          </TouchableOpacity>
        </View>

        <Text style={styles.label}>{tab === "cosmos" ? "Recipient (lcv1...)" : "Recipient (0x...)"}</Text>
        <TextInput style={styles.input} value={sendTo} onChangeText={setSendTo} autoCapitalize="none" />
        <Text style={styles.label}>{tab === "cosmos" ? "Amount (ulcv)" : "Amount (LCV)"}</Text>
        <TextInput style={styles.input} value={sendAmount} onChangeText={setSendAmount} keyboardType="numeric" />

        {sendError && <Text style={styles.error}>{sendError}</Text>}
        {sendResult && (
          <View style={styles.txResult}>
            <Text style={styles.txResultText}>Broadcast: {sendResult}</Text>
          </View>
        )}

        <View style={styles.buttonRow}>
          <TouchableOpacity
            style={[styles.primaryButton, (sending || !sendTo || !sendAmount) && styles.primaryButtonDisabled]}
            disabled={sending || !sendTo || !sendAmount}
            onPress={tab === "cosmos" ? handleSendCosmos : handleSendEvm}
          >
            <Text style={styles.primaryButtonText}>{sending ? "Sending…" : "Send"}</Text>
          </TouchableOpacity>
          <TouchableOpacity style={styles.secondaryButton} onPress={lock}>
            <Text style={styles.secondaryButtonText}>Lock</Text>
          </TouchableOpacity>
        </View>
      </View>
    </ScrollView>
  );
}
