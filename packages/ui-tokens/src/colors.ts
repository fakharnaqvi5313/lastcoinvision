/**
 * Ported from web/tailwind.config.ts — that file stays the source of truth
 * for the marketing site's own Tailwind build; keep these two in sync by
 * hand (there are only two color scales, so drift is easy to catch in
 * review) rather than generating one from the other.
 */
export const navy = {
  50: "#EAF1F8",
  100: "#D4E3F1",
  200: "#A9C7E3",
  300: "#7DABD5",
  400: "#528FC7",
  500: "#1F4E79",
  600: "#1A4166",
  700: "#153453",
  800: "#102740",
  900: "#0E2640",
  950: "#081627",
} as const;

export const accent = {
  50: "#E6FDFF",
  100: "#CCFBFE",
  200: "#99F7FD",
  300: "#66F2FC",
  400: "#33EEFB",
  500: "#0EEBFD",
  600: "#0BC2D1",
  700: "#0899A5",
  800: "#067179",
  900: "#03484C",
} as const;

export type ColorScale = typeof navy;
export type ColorShade = keyof ColorScale;
