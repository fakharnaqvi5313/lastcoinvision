export interface DevnetStats {
  chainId: string;
  blockHeight: number;
  latestBlockTime: string;
  validatorCount: number;
}

const RPC_URL = process.env.NEXT_PUBLIC_LCV_RPC_URL ?? "http://localhost:26657";
const FETCH_TIMEOUT_MS = 3000;

async function fetchJson(path: string): Promise<unknown> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), FETCH_TIMEOUT_MS);
  try {
    const res = await fetch(`${RPC_URL.replace(/\/$/, "")}${path}`, { signal: controller.signal });
    if (!res.ok) {
      throw new Error(`${path} responded ${res.status}`);
    }
    return await res.json();
  } finally {
    clearTimeout(timeout);
  }
}

interface StatusResponse {
  result: {
    node_info: { network: string };
    sync_info: { latest_block_height: string; latest_block_time: string };
  };
}

interface ValidatorsResponse {
  result: { total: string };
}

function isStatusResponse(value: unknown): value is StatusResponse {
  const v = value as Partial<StatusResponse> | null;
  return typeof v?.result?.node_info?.network === "string" && typeof v.result.sync_info?.latest_block_height === "string";
}

function isValidatorsResponse(value: unknown): value is ValidatorsResponse {
  const v = value as Partial<ValidatorsResponse> | null;
  return typeof v?.result?.total === "string";
}

/**
 * Fetches a snapshot of live devnet stats for the homepage strip. Returns
 * `null` on any failure (no RPC reachable, timeout, unexpected shape) so
 * callers can render a graceful fallback rather than a broken page — this
 * is expected on a static deploy with no devnet running behind it.
 */
export async function fetchDevnetStats(): Promise<DevnetStats | null> {
  try {
    const [status, validators] = await Promise.all([fetchJson("/status"), fetchJson("/validators")]);
    if (!isStatusResponse(status) || !isValidatorsResponse(validators)) {
      return null;
    }
    return {
      chainId: status.result.node_info.network,
      blockHeight: Number(status.result.sync_info.latest_block_height),
      latestBlockTime: status.result.sync_info.latest_block_time,
      validatorCount: Number(validators.result.total),
    };
  } catch {
    return null;
  }
}
