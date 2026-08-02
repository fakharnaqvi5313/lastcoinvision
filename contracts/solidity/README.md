# LastcoinVision Solidity reference contracts

Foundry project for the EVM reference contracts described in [../README.md](../README.md).

```bash
forge install   # fetches forge-std into lib/ (gitignored)
forge build
forge test
```

Deploy to a running devnet with a standard Hardhat or Foundry config pointed at its JSON-RPC
endpoint (`:8545` by default) — see [../../modules/evm/SPEC.md](../../modules/evm/SPEC.md).
