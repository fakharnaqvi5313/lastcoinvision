---
sidebar_position: 2
title: Issue a credential
---

# Issue a credential

`x/identity` never stores personal data — only a credential's hash, its type, and its
revocation status, anchored by a governance-authorized issuer. This walks through
authorizing an issuer, issuing a credential, and watching an account's tier flip to
`attested` — every command below was actually run against a live devnet.

## Prerequisites

A running devnet (see [Run a node](../getting-started/run-a-node.md)) and two funded
accounts: one to act as the issuer, one to receive the credential.

```bash
lcvd keys add issuer1 --home ~/.lcv-devnet/validator0 --keyring-backend test
lcvd keys add subject1 --home ~/.lcv-devnet/validator0 --keyring-backend test

lcvd tx bank send validator0 <issuer1-address> 5000000ulcv \
  --home ~/.lcv-devnet/validator0 --keyring-backend test --chain-id lcv-devnet \
  --gas 300000 --fees 10000ulcv -y
```

## Authorize the issuer

Issuers are added via governance (`MsgAddIssuer`'s `authority` is `x/gov`), not by any
individual account unilaterally. This devnet runs a fast governance config (30s voting
period) specifically so this kind of walkthrough doesn't take an hour:

```bash
cat > add-issuer-proposal.json <<'EOF'
{
  "messages": [
    {
      "@type": "/lcv.identity.MsgAddIssuer",
      "authority": "<gov-module-account-address>",
      "issuer_address": "<issuer1-address>",
      "name": "Tutorial Issuer"
    }
  ],
  "metadata": "",
  "deposit": "10000000ulcv",
  "title": "Authorize tutorial credential issuer",
  "summary": "Adds issuer1 as an authorized x/identity credential issuer."
}
EOF
```

Get the gov module account address (it's deterministic, not a key you generate):

```bash
lcvd query auth module-account gov --home ~/.lcv-devnet/validator0
```

Submit and vote:

```bash
lcvd tx gov submit-proposal add-issuer-proposal.json \
  --from validator0 --home ~/.lcv-devnet/validator0 --keyring-backend test \
  --chain-id lcv-devnet --gas 300000 --fees 10000ulcv -y

# vote yes from every validator (this devnet's 4 validators hold all the
# voting power — a real network needs the equivalent quorum/threshold from
# its actual delegator set)
for i in 0 1 2 3; do
  lcvd tx gov vote <proposal-id> yes --from validator$i --home ~/.lcv-devnet/validator$i \
    --keyring-backend test --chain-id lcv-devnet --gas 300000 --fees 10000ulcv -y
done
```

After the voting period ends:

```bash
curl -s http://localhost:1317/lastcoinvision/lcv/identity/issuers
# {"issuers":[{"address":"lcv19dehj6ta4e82mpehdq6k9nye25r207y38kfklm","name":"Tutorial Issuer","added_at":"..."}],"pagination":null}
```

## Issue the credential

```bash
CRED_HASH=$(echo -n "subject1-kyc-credential-doc-v1" | shasum -a 256 | cut -d' ' -f1)

lcvd tx identity issue-credential \
  --subject <subject1-address> \
  --credential-type kyc \
  --credential-hash "0x${CRED_HASH}" \
  --from issuer1 --home ~/.lcv-devnet/validator0 --keyring-backend test \
  --chain-id lcv-devnet --gas 300000 --fees 10000ulcv -y
```

Note what's on-chain here: a hash of an off-chain document, a type (`"kyc"`), and who issued
it. Never the document itself, never a name or date of birth.

## Check the tier

`attested_credential_type` is a module param (`"kyc"` by default) — an account is
automatically `attested` if and only if it holds a valid, unrevoked credential of that type.
Nothing sets the tier directly; it's derived:

```bash
curl -s http://localhost:1317/lastcoinvision/lcv/identity/tier/<subject1-address>
# {"tier":"ACCOUNT_TIER_ATTESTED","role_metadata":""}
```

Revoke it and the tier reverts automatically:

```bash
lcvd tx identity revoke-credential --credential-id <id> \
  --from issuer1 --home ~/.lcv-devnet/validator0 --keyring-backend test \
  --chain-id lcv-devnet --gas 300000 --fees 10000ulcv -y
```

## Next

This `attested` account can now do anything that requires it — for example, register to
vote in an election whose `eligibility_credential_type` is `"kyc"`. See [Build a voting
portal](./build-a-voting-portal.md).
