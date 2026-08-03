# Security Policy

LastcoinVision (LCV) is currently **devnet/testnet-stage software**. There is no mainnet, no
real value at stake, and no bug bounty program. This policy exists anyway because the
practices that matter for a production chain — a private disclosure channel, a documented
response process — are worth establishing now, not retrofitting later under pressure.

## Scope

This policy covers vulnerabilities in:

- `chain/` — the `lcvd` Cosmos SDK application, including all custom modules
  (`x/emission`, `x/feemarket`, `x/identity`, `x/registry`, `x/voting`, `x/taxrail`) and the
  CosmWasm/EVM integration layers
- `contracts/` — the reference Solidity and CosmWasm contracts
- `sdk/`, `cli/`, `faucet/` — the TypeScript SDK, `create-lcv-app` scaffolder, and testnet
  faucet
- `deploy/` — the docker-compose and Kubernetes deployment manifests

**Out of scope:** `web/` and `docs/` (the marketing/documentation sites — a defacement or
content issue there is a bug report, not a security disclosure) and anything already listed
as a known limitation in [`AUDIT_SCOPE.md`](AUDIT_SCOPE.md) (most notably, the commit-reveal
voting scheme's lack of ballot secrecy — that's a documented design limitation of this stage,
not something to report as a new finding).

## Reporting a vulnerability

**Do not open a public GitHub issue for a security vulnerability.** Use
[GitHub's private security advisory feature](https://docs.github.com/en/code-security/security-advisories/guidance-on-reporting-and-writing/privately-reporting-a-security-vulnerability)
on this repository instead (Security tab → "Report a vulnerability"). If that's unavailable to
you, email the maintainers directly rather than filing a public issue — check the repository's
GitHub organization page for a current contact address, since none is fabricated here.

Please include:

- The affected component (module name, contract, package) and, if known, the specific
  function or code path
- Steps to reproduce, or a proof-of-concept transaction/script against a local devnet
  (`make devnet` — see the [README](README.md))
- What you believe the impact is (funds at risk, consensus safety, availability, information
  disclosure, etc.) and any preconditions required to exploit it

## What to expect

This is a small project without a formal security team or SLA. In good faith, expect:

- Acknowledgment of a report within a reasonable time
- An honest assessment of severity and whether it's already a known/accepted limitation
  (check `AUDIT_SCOPE.md` and `DECISIONS.md` first — several are already documented)
- A fix, or a documented reason a fix isn't planned, before any public disclosure

Please give us a chance to ship a fix before disclosing publicly. Once mainnet exists, this
policy will be revisited with real coordinated-disclosure timelines and (if funded) a bug
bounty scope — neither exists today, and this document will not claim otherwise.

## Recognition

There is no bug bounty program at this stage. Reporters of genuine, previously-unknown
vulnerabilities will be credited (with permission) in `DECISIONS.md` or release notes, which
is the only form of recognition currently available.
