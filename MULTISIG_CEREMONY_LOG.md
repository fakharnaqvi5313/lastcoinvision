# Treasury Multisig Ceremony — Live Record

Execution record for the ceremony described in `MULTISIG_CEREMONY_RUNBOOK.md`. Updated as each
signer's public key is collected. This is the actual evidence `MAINNET_LAUNCH_RUNBOOK.md` §1
needs once the ceremony completes — not a placeholder.

## Signers (3 total, per `DECISIONS.md`)

| # | Name | Role | Pubkey | Status |
|---|---|---|---|---|
| 1 | Fakhar Naqvi | Founder | `{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A9g5qfLEFZ6DNgvJB0F55DlZD+HSUJbqEX7WrRw2Le2E"}` | Received 2026-08-07 |
| 2 | Hafiza Hoor | Co-founder | `{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AxO0gkAt5abuQBfCu+c7iAANjmeNll0JZ3BcLkmr7wqS"}` | Received 2026-08-07 |
| 3 | Muhammad Muslim | Business Partner | `{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AzLeYVr1AI88BVTR08vO4KZeZtmbuC05AROW8vq6OwpS"}` | Received 2026-08-07 |

## Threshold

**2-of-3.** Confirmed 2026-08-07.

## Constructed account

**Address**: `lcv146lq4cqup54a78nqjpg38sl420da4a40uzr3tt`

Derived 2026-08-07 from the three pubkeys above via:

```bash
lcvd keys add fakhar --pubkey '<Fakhar's pubkey JSON>' --keyring-backend os
lcvd keys add hafiza --pubkey '<Hafiza's pubkey JSON>' --keyring-backend os
lcvd keys add muhammad --pubkey '<Muhammad's pubkey JSON>' --keyring-backend os
lcvd keys add treasury-multisig --multisig "fakhar,hafiza,muhammad" --multisig-threshold 2 --keyring-backend os
```

(`--pubkey`, not `--pubkey-base64` — confirmed against `lcvd keys add --help`; it expects the
full JSON blob `cosmosPubkeyJson()` produces, which is exactly what all three signers sent back.)

Verified via a real dry run against a throwaway `--keyring-backend test` keyring using these
exact three pubkeys before treating the address as final — this is a deterministic derivation
(no private key material involved, per the runbook), so re-running the same commands on any
machine with these three pubkeys reproduces the identical address. Dry-run keyring deleted after
verification.

Multisig pubkey (needed for future signing/co-signing flows, not secret):

```json
{"@type":"/cosmos.crypto.multisig.LegacyAminoPubKey","threshold":2,"public_keys":[{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AzLeYVr1AI88BVTR08vO4KZeZtmbuC05AROW8vq6OwpS"},{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A9g5qfLEFZ6DNgvJB0F55DlZD+HSUJbqEX7WrRw2Le2E"},{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AxO0gkAt5abuQBfCu+c7iAANjmeNll0JZ3BcLkmr7wqS"}]}
```

## Next steps

1. **Not done yet — test the threshold for real** before this address is trusted for anything:
   send it a small, inconsequential amount on `lcv-testnet-1`, then have any 2 of the 3 signers
   actually produce a valid combined signature moving funds back out. This is the step that turns
   "we derived an address" into "we confirmed this multisig actually works."
2. Once that's verified, this becomes the real `treasury` bucket address for
   `MAINNET_LAUNCH_RUNBOOK.md` §1 (replacing the placeholder single-key address). The `ecosystem`
   bucket still needs its own, separate ceremony — not yet started.
