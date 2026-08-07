# third_party/

Vendored copies of upstream dependencies that needed a local patch this repo maintains itself,
because no released upstream version has the fix. Referenced via a `replace` directive in
`chain/go.mod` — see the comment there for which patch applies to which vendored copy.

- **`cosmos-evm-patched/`** — `github.com/cosmos/evm@v0.6.1`, patched for
  [GHSA-mjfq-3qr2-6g84](https://github.com/cosmos/evm/security/advisories/GHSA-mjfq-3qr2-6g84)
  ("Partial Precompile State Writes", Critical). See `AUDIT_FINDINGS_RESPONSE.md` Finding #6 and
  `DECISIONS.md` (2026-08-07) for the full trace of why no upstream release fixes this and what
  the patch does. The only change from the real upstream `v0.6.1` is a single `defer` block in
  `precompiles/common/precompile.go`'s `runNativeAction`, marked with an `LCV PATCH` comment.

**When updating**: if a future `cosmos/evm` release ships a real fix for the advisory above,
switch back to a normal versioned dependency and delete this vendored copy — don't keep
maintaining a local patch once upstream has one.
