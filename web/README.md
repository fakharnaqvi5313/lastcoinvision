# lastcoinvision.com

Next.js 14 (App Router) + Tailwind marketing/informational site.

## Pages

`/` (hero, live devnet stats, use cases, roadmap), `/developers`, `/governments` (with a
pilot-program contact form), `/technology` (including an honest limitations section),
`/token` (supply table + emission chart), `/roadmap`, `/blog` (MDX), `/brand`.

## Develop

```bash
cp .env.local.example .env.local
npm install
npm run dev
```

The homepage's live devnet stats strip polls `NEXT_PUBLIC_LCV_RPC_URL` (default
`http://localhost:26657`) and falls back gracefully if nothing is reachable there.

## Build

```bash
npm run build
```

Every page is static except `/api/contact` (a stub handler for the governments pilot-program
form — see DECISIONS.md for why this isn't a pure `output: "export"` build).

## Notes

- Brand colors, MDX blog posts, and the roadmap's phase-based (not calendar-dated) status are
  all explained inline on `/brand` and `/roadmap`, and in DECISIONS.md.
- Verified with a real headless-Chrome Lighthouse run (not assumed): every page scores 100 on
  accessibility and SEO, ≥94 on performance and best-practices.
