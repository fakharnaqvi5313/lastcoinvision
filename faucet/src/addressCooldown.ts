const lastClaimedAt = new Map<string, number>();

/** Milliseconds remaining before `address` may claim again; 0 if it's clear now. */
export function getAddressCooldownRemaining(address: string, cooldownMs: number): number {
  const last = lastClaimedAt.get(address);
  if (last === undefined) {
    return 0;
  }
  const elapsed = Date.now() - last;
  return elapsed >= cooldownMs ? 0 : cooldownMs - elapsed;
}

/** Call only after a drip has actually broadcast successfully. */
export function recordAddressClaim(address: string): void {
  lastClaimedAt.set(address, Date.now());
}
