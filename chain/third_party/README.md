# third_party/

Vendored copies of upstream dependencies that needed a local patch this repo maintains itself,
because no released upstream version has the fix. Referenced via a `replace` directive in
`chain/go.mod` — see the comment there for which patch applies to which vendored copy.

- **`cosmos-evm-patched/`** — `github.com/cosmos/evm@v0.6.1`, patched for
  [GHSA-mjfq-3qr2-6g84](https://github.com/cosmos/evm/security/advisories/GHSA-mjfq-3qr2-6g84)
  ("Partial Precompile State Writes", Critical). See `AUDIT_FINDINGS_RESPONSE.md` Finding #6 and
  `DECISIONS.md` (2026-08-07) for the full trace of why no upstream release fixes this and what
  the patch does. The only functional change from the real upstream `v0.6.1` is a single `defer`
  block in `precompiles/common/precompile.go`'s `runNativeAction`, marked with an `LCV PATCH`
  comment.

**Trimmed, not a full mirror of upstream.** The initial vendor pass copied the entire ~1,235-file
upstream repo, which included its own `tests/` integration-test suite (167 Go files, never
imported by `lcvd`'s build — verified by removing it and confirming `go build ./cmd/lcvd` still
succeeds) and over a dozen scattered `package.json`/`package-lock.json` files for that suite's
JS/npm tooling (hardhat, viem, web3.js). GitHub's dependency graph scans every manifest file it
finds in the repo regardless of Go module boundaries, so those files alone inflated this repo's
Dependabot alert count from 7 to 465 (26 critical) on push — see `DECISIONS.md` (2026-08-07).
Removed `tests/`, every JS/npm manifest, and the upstream repo's own CI/lint/tooling configs
(`.github/`, `docs/`, `Makefile`, `docker-compose.yml`, etc.) — none of it is needed to satisfy
the Go build. Kept: every `.go` source file actually reachable from `lcvd`'s import graph, plus
`go.mod`/`go.sum` (genuinely required for the `replace` directive to resolve), `LICENSE`, and
`README.md`.

**When updating**: if a future `cosmos/evm` release ships a real fix for the advisory above,
switch back to a normal versioned dependency and delete this vendored copy — don't keep
maintaining a local patch once upstream has one. If re-vendoring for any reason, trim it the same
way — don't copy the upstream repo's own tests/tooling wholesale again.
