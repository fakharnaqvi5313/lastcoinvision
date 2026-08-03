import type { Config } from "tailwindcss";

const config: Config = {
  darkMode: "class",
  content: ["./app/**/*.{ts,tsx,mdx}", "./components/**/*.{ts,tsx}", "./mdx-components.tsx"],
  theme: {
    extend: {
      colors: {
        // "Deep navy (#1F4E79) + white base, single accent" per the brand
        // spec, reconciled with the master logo's darker disc (#0E2640,
        // used here as navy-900) — see /brand and DECISIONS.md.
        navy: {
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
        },
        accent: {
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
        },
      },
      fontFamily: {
        sans: ["var(--font-sans)", "system-ui", "sans-serif"],
      },
    },
  },
  plugins: [require("@tailwindcss/typography")],
};

export default config;
