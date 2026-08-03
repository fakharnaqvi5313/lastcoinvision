---
sidebar_position: 1
title: Build a voting portal
---

# Build a voting portal

This walks through the full `x/voting` lifecycle — authorize an electoral authority, create
an election, register, commit, and reveal a vote — then scaffolds a portal to view the
result. Every command below was actually run against a live devnet, including the exact
commit-reveal hash format the chain expects.

## Prerequisites

A running devnet (see [Run a node](../getting-started/run-a-node.md)) and a voter account
holding a `kyc` credential — see [Issue a credential](./issue-a-credential.md) first if you
haven't. This tutorial assumes a `voter1` account already `ACCOUNT_TIER_ATTESTED`.

## Authorize an electoral authority

Like issuers, electoral authorities are added via governance:

```bash
GOV_ADDR=$(lcvd query auth module-account gov --home ~/.lcv-devnet/validator0 | grep address | awk '{print $2}')

cat > add-authority-proposal.json <<EOF
{
  "messages": [{
    "@type": "/lcv.voting.MsgAddElectoralAuthority",
    "authority": "$GOV_ADDR",
    "address": "<validator0-address>",
    "name": "Tutorial Authority"
  }],
  "metadata": "",
  "deposit": "10000000ulcv",
  "title": "Authorize tutorial electoral authority",
  "summary": "Adds validator0 as an authorized x/voting electoral authority."
}
EOF

lcvd tx gov submit-proposal add-authority-proposal.json \
  --from validator0 --home ~/.lcv-devnet/validator0 --keyring-backend test \
  --chain-id lcv-devnet --gas 300000 --fees 10000ulcv -y

# vote yes from every validator, same as in Issue a credential
```

## Create an election

Only an authorized address may create one. Timestamps are Unix seconds; give yourself
realistic windows — this devnet's fast-gov config doesn't shorten election windows, and a
window that's too tight is the easiest way to have a vote reveal too late to count:

```bash
NOW=$(date +%s)
lcvd tx voting create-election \
  --title "Tutorial Referendum" --description "..." \
  --eligibility-credential-type kyc --options "yes,no" \
  --registration-start $((NOW+5)) --registration-end $((NOW+60)) \
  --voting-start $((NOW+60)) --voting-end $((NOW+120)) --reveal-end $((NOW+200)) \
  --from validator0 --home ~/.lcv-devnet/validator0 --keyring-backend test \
  --chain-id lcv-devnet --gas 300000 --fees 10000ulcv -y
```

The election ID isn't the proposal/tx sequence number — it's a string like `election-1`,
visible in the tx's response data or via:

```bash
curl -s http://localhost:1317/lastcoinvision/lcv/voting/election/election-1
```

## Register, commit, reveal

**Register** (only during the registration window, only if `voter1` holds a `kyc`
credential):

```bash
lcvd tx voting register-to-vote --election-id election-1 \
  --from voter1 --home ~/.lcv-devnet/validator0 --keyring-backend test \
  --chain-id lcv-devnet --gas 300000 --fees 10000ulcv -y
```

**Commit** a hash of your choice and a secret salt — not the choice itself. The exact
formula the chain checks at reveal time is `sha256(choice + "|" + salt)`, hex-encoded, no
`0x` prefix:

```bash
CHOICE="yes"
SALT="some-secret-salt"
COMMITMENT=$(echo -n "${CHOICE}|${SALT}" | shasum -a 256 | cut -d' ' -f1)

lcvd tx voting commit-vote --election-id election-1 --commitment "$COMMITMENT" \
  --from voter1 --home ~/.lcv-devnet/validator0 --keyring-backend test \
  --chain-id lcv-devnet --gas 300000 --fees 10000ulcv -y
```

**Reveal**, once the voting window ends — resubmit the same choice and salt, and the chain
recomputes the hash to verify it matches your commitment:

```bash
lcvd tx voting reveal-vote --election-id election-1 --choice "$CHOICE" --salt "$SALT" \
  --from voter1 --home ~/.lcv-devnet/validator0 --keyring-backend test \
  --chain-id lcv-devnet --gas 300000 --fees 10000ulcv -y
```

:::caution
`lcvd tx ... -y` in sync broadcast mode returns as soon as `CheckTx` passes — a `code: 0`
here does **not** guarantee the message executed successfully on-chain. Query the tx hash
afterward (`lcvd query tx <hash>`) if you need to confirm inclusion, especially for a reveal
where a hash mismatch fails silently from the broadcast command's own output.
:::

Check the tally — anyone can recompute this from public chain data, that's the point:

```bash
curl -s http://localhost:1317/lastcoinvision/lcv/voting/election/election-1/results
# {"tallies":[{"option":"yes","count":"1"},{"option":"no","count":"0"}],"total_revealed":"1","total_registered":"1"}
```

## Scaffold the portal

```bash
npx create-lcv-app my-voting-portal --template voting-portal
cd my-voting-portal
npm install
npm run dev
```

Open http://localhost:3000, enter `election-1` in the election lookup field — no wallet
needed, it's a read-only REST query — and you'll see the same election detail and tally
JSON you just queried with `curl`.

## Limitations

This is verifiable, not ballot-secret — see the [honest limitations on the technology
page](https://lastcoinvision.com/technology) before considering this design for a real
election.
