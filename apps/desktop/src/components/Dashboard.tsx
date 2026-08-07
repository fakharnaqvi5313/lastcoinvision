import { buildEvmWalletClient } from "@lastcoinvision/wallet-core";
import { SigningStargateClient } from "@cosmjs/stargate";
import { useCallback, useEffect, useState } from "react";
import { createPublicClient, formatEther, http, parseEther } from "viem";

import { NETWORKS } from "../lib/network.js";
import type { NetworkConfig } from "../lib/network.js";
import { useSession } from "../lib/session.js";

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

  const [pubkeyJson, setPubkeyJson] = useState<string | null>(null);
  const [pubkeyCopied, setPubkeyCopied] = useState(false);

  async function togglePubkey() {
    if (!wallet) return;
    if (pubkeyJson) {
      setPubkeyJson(null);
      return;
    }
    setPubkeyJson(await wallet.cosmosPubkeyJson());
    setPubkeyCopied(false);
  }

  async function copyPubkey() {
    if (!pubkeyJson) return;
    await navigator.clipboard.writeText(pubkeyJson);
    setPubkeyCopied(true);
  }

  const refreshBalances = useCallback(async () => {
    if (!wallet || !cosmosAddress) return;
    setBalanceError(null);
    try {
      const res = await fetch(`${network.restEndpoint}/cosmos/bank/v1beta1/balances/${cosmosAddress}/by_denom?denom=ulcv`);
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
    <div className="dashboard">
      <div className="card">
        <h1>Wallet</h1>

        <div className="field">
          <label htmlFor="network">Network</label>
          <select
            id="network"
            value={network.name}
            onChange={(e) => {
              const next = Object.values(NETWORKS).find((n: NetworkConfig) => n.name === e.target.value);
              if (next) setNetwork(next);
            }}
          >
            {Object.values(NETWORKS).map((n: NetworkConfig) => (
              <option key={n.name} value={n.name}>
                {n.name}
              </option>
            ))}
          </select>
        </div>

        <div className="address-row">
          <span className="label">Cosmos address</span>
          <span className="value">{cosmosAddress}</span>
          <span className="muted">{cosmosBalance !== null ? `${cosmosBalance} ulcv` : "—"}</span>
          <button className="link-btn" onClick={() => void togglePubkey()}>
            {pubkeyJson ? "Hide public key" : "Show public key"}
          </button>
          {pubkeyJson && (
            <>
              <span className="muted">
                Safe to share — e.g. with whoever is constructing a multisig you're one signer of. Never share your
                recovery phrase.
              </span>
              <span className="value pubkey-json">{pubkeyJson}</span>
              <button className="link-btn" onClick={() => void copyPubkey()}>
                {pubkeyCopied ? "Copied" : "Copy"}
              </button>
            </>
          )}
        </div>
        <div className="address-row">
          <span className="label">EVM address</span>
          <span className="value">{wallet.evmAddress}</span>
          <span className="muted">
            {network.evmRpcUrl ? (evmBalance !== null ? `${evmBalance} LCV` : "—") : "EVM RPC not available on this network"}
          </span>
        </div>
        {balanceError && <p className="error">{balanceError}</p>}

        <div className="tabs">
          <button className={tab === "cosmos" ? "active" : ""} onClick={() => setTab("cosmos")}>
            Send (Cosmos)
          </button>
          <button className={tab === "evm" ? "active" : ""} onClick={() => setTab("evm")} disabled={!network.evmRpcUrl}>
            Send (EVM)
          </button>
        </div>

        <div className="field">
          <label htmlFor="to">{tab === "cosmos" ? "Recipient (lcv1...)" : "Recipient (0x...)"}</label>
          <input id="to" value={sendTo} onChange={(e) => setSendTo(e.target.value)} />
        </div>
        <div className="field">
          <label htmlFor="amount">{tab === "cosmos" ? "Amount (ulcv)" : "Amount (LCV)"}</label>
          <input id="amount" value={sendAmount} onChange={(e) => setSendAmount(e.target.value)} />
        </div>

        {sendError && <p className="error">{sendError}</p>}
        {sendResult && <div className="tx-result">Broadcast: {sendResult}</div>}

        <div className="btn-row">
          <button
            className="btn-primary"
            disabled={sending || !sendTo || !sendAmount}
            onClick={tab === "cosmos" ? handleSendCosmos : handleSendEvm}
          >
            {sending ? "Sending…" : "Send"}
          </button>
          <button className="btn-secondary" onClick={lock}>
            Lock
          </button>
        </div>
      </div>
    </div>
  );
}
