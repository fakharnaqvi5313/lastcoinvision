# LastcoinVision documentation site

Docusaurus 3 (TypeScript) site for docs.lastcoinvision.com.

## Structure

- `docs/` — Getting Started, Tutorials, Validator Handbook, API Reference (the default docs
  plugin instance, served at `/`)
- A second docs plugin instance, configured in `docusaurus.config.ts`, serves `../modules`
  directly at `/modules/` — every module's `SPEC.md` is rendered from its real location in
  this repo, not copied, so the reference and the source code can't drift apart. See the
  2026-08-03 DECISIONS.md entry for the trade-off this implies (a few GitHub-relative links
  inside those SPEC.md files, like `../../DECISIONS.md`, don't resolve to a Docusaurus route
  and are left as warnings, not fixed).

## Develop

```bash
npm install
npm run dev
```

## Build

```bash
npm run build
npm run serve   # preview the production build locally
```

## Notes

Every tutorial under `docs/tutorials/` was actually run against a live `make devnet`
while being written — see the 2026-08-03 DECISIONS.md entries for what that testing found
and fixed in the devnet tooling itself (several bugs in `make devnet` that predated this
phase).
