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

## Threshold test — mechanics proof (throwaway keys, completed)

Before asking real signers to do anything, proved the exact sign → combine → broadcast command
sequence against real `lcv-testnet-1` using 3 disposable throwaway keys (never used again,
deleted after): constructed a throwaway 2-of-3 multisig, funded it from the real faucet
(confirmed on-chain), generated an unsigned send, signed with 2 of the 3 throwaway keys
independently, combined into one valid tx via `lcvd tx multisign`, broadcast, and confirmed
inclusion — **block 47825, `code: 0`**, destination balance matched the amount sent exactly.
This confirms the command sequence below is correct before any real signer runs it for real.

## Threshold test — real ceremony (in progress)

**Funded 2026-08-07**: the real treasury multisig (`lcv146lq4cqup54a78nqjpg38sl420da4a40uzr3tt`)
received 10 LCV testnet tokens from the faucet (confirmed on-chain).

**Unsigned test transaction** (send 1 LCV from the multisig back to Fakhar's own personal
address `lcv1w95xt69arjdjyhpdf09uwwzzzj65fa7cs60r89`, derived from his pubkey above — no private
key needed to generate this):

```json
{"body":{"messages":[{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"lcv146lq4cqup54a78nqjpg38sl420da4a40uzr3tt","to_address":"lcv1w95xt69arjdjyhpdf09uwwzzzj65fa7cs60r89","amount":[{"denom":"ulcv","amount":"1000000"}]}],"memo":"","timeout_height":"0","unordered":false,"timeout_timestamp":null,"extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[{"denom":"ulcv","amount":"5000"}],"gas_limit":"200000","payer":"","granter":""},"tip":null},"signatures":[]}
```

**What's still needed**: Fakhar signs this locally with his real key (`--keyring-backend os`),
one other real signer (Hafiza or Muhammad) does the same, then Fakhar combines both signatures
with `lcvd tx multisign` and broadcasts. Not yet done — this is the actual real-world proof the
runbook requires; the throwaway-key run above only proves the mechanics are correct.

## Next steps

1. Complete the real ceremony test above with 2 real signers.
2. Once that's verified, this becomes the real `treasury` bucket address for
   `MAINNET_LAUNCH_RUNBOOK.md` §1 (replacing the placeholder single-key address). The `ecosystem`
   bucket still needs its own, separate ceremony — not yet started.
