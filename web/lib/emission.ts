// Mirrors chain/x/emission/keeper/decay.go's actual default parameters and
// formula exactly (12% of the fixed 10B supply, 2-year half-life) — this
// chart is derived from the real on-chain default, not a separately
// invented curve.
export const TOTAL_SUPPLY_LCV = 10_000_000_000;
export const STAKING_REWARD_POOL_LCV = 1_200_000_000; // 12% of total supply
export const HALF_LIFE_YEARS = 2;

const LN2 = Math.log(2);
const LAMBDA = LN2 / HALF_LIFE_YEARS;

export interface EmissionPoint {
  year: number;
  cumulativeEmittedLcv: number;
  percentOfPoolEmitted: number;
}

/** Cumulative LCV emitted from the staking-rewards pool by `year` years in. */
export function cumulativeEmitted(year: number): number {
  return STAKING_REWARD_POOL_LCV * (1 - Math.exp(-LAMBDA * year));
}

export function buildEmissionCurve(maxYear = 12): EmissionPoint[] {
  const points: EmissionPoint[] = [];
  for (let year = 0; year <= maxYear; year++) {
    const emitted = cumulativeEmitted(year);
    points.push({
      year,
      cumulativeEmittedLcv: Math.round(emitted),
      percentOfPoolEmitted: (emitted / STAKING_REWARD_POOL_LCV) * 100,
    });
  }
  return points;
}

export const supplyAllocation = [
  { label: "Ecosystem / grants", percent: 25, lcv: 2_500_000_000, notes: "Governance-directed ecosystem grants." },
  { label: "Foundation treasury", percent: 20, lcv: 2_000_000_000, notes: "Long-term protocol treasury." },
  { label: "Team", percent: 15, lcv: 1_500_000_000, notes: "4-year vest, 1-year cliff." },
  { label: "Investors", percent: 15, lcv: 1_500_000_000, notes: "3-year vest, 1-year cliff." },
  { label: "Staking rewards", percent: 12, lcv: 1_200_000_000, notes: "10-year decaying emission (2-year half-life)." },
  { label: "Public", percent: 8, lcv: 800_000_000, notes: "" },
  { label: "Community / airdrop", percent: 5, lcv: 500_000_000, notes: "" },
] as const;
