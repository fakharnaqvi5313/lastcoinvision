/**
 * Thin fetch wrapper for the chain's gRPC-gateway REST API (default
 * `http://localhost:1317`). Every custom module's queries this SDK exposes
 * go through here rather than gRPC-web, since it needs zero extra
 * transport dependencies in either Node or the browser.
 */
export class RestClient {
  constructor(private readonly baseUrl: string) {}

  async get<T>(path: string): Promise<T> {
    const url = `${this.baseUrl.replace(/\/$/, "")}${path}`;
    const res = await fetch(url);
    if (!res.ok) {
      const body = await res.text().catch(() => "");
      throw new Error(`GET ${url} failed: ${res.status} ${res.statusText} ${body}`);
    }
    return (await res.json()) as T;
  }
}

/** Percent-encodes a single path segment (e.g. an address or asset id). */
export function pathSegment(value: string): string {
  return encodeURIComponent(value);
}
