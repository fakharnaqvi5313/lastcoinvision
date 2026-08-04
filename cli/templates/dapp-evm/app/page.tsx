"use client";

import { formatUnits } from "viem";
import { useAccount, useBalance, useConnect, useDisconnect } from "wagmi";
import { injected } from "wagmi/connectors";

export default function Home() {
  const { address, isConnected } = useAccount();
  const { connect, isPending: isConnecting, error: connectError } = useConnect();
  const { disconnect } = useDisconnect();
  const { data: balance, isLoading: isBalanceLoading } = useBalance({
    address,
    query: { enabled: Boolean(address) },
  });

  return (
    <main style={{ maxWidth: 560, margin: "4rem auto", fontFamily: "system-ui, sans-serif", padding: "0 1rem" }}>
      <h1>LCV dApp (EVM)</h1>
      <p>Connects MetaMask to LastcoinVision&apos;s EVM and displays the native LCV balance.</p>

      {!isConnected ? (
        <button onClick={() => connect({ connector: injected() })} disabled={isConnecting}>
          {isConnecting ? "Connecting..." : "Connect MetaMask"}
        </button>
      ) : (
        <div>
          <p>
            Address: <code>{address}</code>
          </p>
          <p>
            Balance:{" "}
            {isBalanceLoading ? "loading..." : balance ? `${formatUnits(balance.value, balance.decimals)} ${balance.symbol}` : "—"}
          </p>
          <button onClick={() => disconnect()}>Disconnect</button>
        </div>
      )}

      {connectError && <p style={{ color: "crimson" }}>{connectError.message}</p>}
    </main>
  );
}
