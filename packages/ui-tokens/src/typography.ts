/**
 * web/tailwind.config.ts only defines a font-family override (via a CSS
 * var, resolved by Next.js' font loader) — there's no custom type/spacing
 * scale to port, just Tailwind's defaults. Non-Tailwind consumers (React
 * Native, Tauri) should fall back to their own platform system font unless
 * the brand later ships a real webfont, tracked as a follow-up rather than
 * invented here.
 */
export const fontFamilyFallback = ["system-ui", "sans-serif"] as const;
