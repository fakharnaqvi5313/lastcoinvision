import { vi } from "vitest";
import type { SigningStargateClient } from "@cosmjs/stargate";

/** Stubs global `fetch` to resolve once with the given JSON body. */
export function mockFetchJsonOnce(body: unknown, ok = true, status = 200): ReturnType<typeof vi.fn> {
  const fetchMock = vi.fn().mockResolvedValue({
    ok,
    status,
    statusText: ok ? "OK" : "Error",
    json: async () => body,
    text: async () => JSON.stringify(body),
  });
  vi.stubGlobal("fetch", fetchMock);
  return fetchMock;
}

/** A minimal fake signing client whose `signAndBroadcast` calls can be asserted on. */
export function fakeSigningClient(): { client: SigningStargateClient; signAndBroadcast: ReturnType<typeof vi.fn> } {
  const signAndBroadcast = vi.fn().mockResolvedValue({ code: 0, transactionHash: "DEADBEEF" });
  return { client: { signAndBroadcast } as unknown as SigningStargateClient, signAndBroadcast };
}
