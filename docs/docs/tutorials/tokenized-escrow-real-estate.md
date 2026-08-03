---
sidebar_position: 4
title: Tokenized escrow for real estate
---

# Tokenized escrow for real estate

`LCVEscrow.sol` is a milestone escrow: a buyer deposits the full sale price up front, and a
neutral arbiter releases funds to the seller milestone by milestone as delivery progresses —
or refunds whatever's undisbursed if the deal falls through. This walks through a full
lifecycle (create, deposit, release) with three real accounts; every step below was actually
run against a live devnet.

## Prerequisites

A running devnet with EVM RPC reachable, Foundry installed, and `LCVEscrow` deployed — see
[Deploy a Solidity contract](./deploy-a-solidity-contract.md) if you haven't done that yet.
This tutorial uses three separate accounts: buyer, seller, arbiter.

```bash
cast wallet new   # buyer (fund it — see the previous tutorial)
cast wallet new   # seller (no funds needed — receiving only)
cast wallet new   # arbiter (fund it — pays gas to release milestones)
```

## Create the escrow

The buyer calls `createEscrow`, listing each milestone's description and amount (in wei —
18-decimal, same as any EVM native-currency amount):

```bash
cast send $CONTRACT \
  "createEscrow(address,address,string[],uint256[])" \
  $SELLER $ARBITER \
  '["Foundation complete","Final delivery"]' \
  '[3000000000000000000,2000000000000000000]' \
  --rpc-url http://localhost:8545 --private-key $BUYER_KEY
```

This is a 5 LCV total deal (3 + 2), across two milestones. `escrowId` starts at `0` and
increments per escrow — this call returns `escrowId 0`.

## Deposit

The buyer sends the full amount in one call (partial deposits across multiple calls also
work — the contract just rejects anything that would push `deposited` past `totalAmount`):

```bash
cast send $CONTRACT "deposit(uint256)" 0 \
  --value 5000000000000000000 \
  --rpc-url http://localhost:8545 --private-key $BUYER_KEY
```

Check the escrow state:

```bash
cast call $CONTRACT \
  "getEscrow(uint256)(address,address,address,uint256,uint256,uint256,bool)" 0 \
  --rpc-url http://localhost:8545
# buyer, seller, arbiter, totalAmount=5e18, deposited=5e18, releasedAmount=0, cancelled=false
```

## Release a milestone

Only the arbiter can do this, and only once the escrow is fully deposited:

```bash
cast balance $SELLER --rpc-url http://localhost:8545
# 0

cast send $CONTRACT "releaseMilestone(uint256,uint256)" 0 0 \
  --rpc-url http://localhost:8545 --private-key $ARBITER_KEY

cast balance $SELLER --rpc-url http://localhost:8545
# 3000000000000000000   — exactly the first milestone's amount
```

Release the second milestone (`releaseMilestone(0, 1)`) the same way once its deliverable is
confirmed. Each milestone can only be released once (`MilestoneAlreadyReleased`).

## If the deal falls through

The arbiter can cancel and refund whatever hasn't been released back to the buyer:

```bash
cast send $CONTRACT "cancelAndRefund(uint256)" 0 \
  --rpc-url http://localhost:8545 --private-key $ARBITER_KEY
```

This refunds `deposited - releasedAmount` to the buyer and marks the escrow `cancelled` —
further deposits or releases against it then revert with `EscrowIsCancelled`.

## What this doesn't handle

This is a reference contract, not a production-ready real-estate settlement system: there's
no dispute resolution beyond "trust the arbiter," no partial-milestone amounts, and no
integration with `x/registry` (LCV's on-chain asset registry) to actually transfer title
alongside the funds — those would need to be composed on top, and any real deployment needs
its own legal/regulatory review for the jurisdiction involved.
