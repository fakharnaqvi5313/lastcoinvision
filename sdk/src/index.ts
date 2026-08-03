export { IdentityClient } from "./clients/identity.js";
export { RegistryClient } from "./clients/registry.js";
export { VotingClient } from "./clients/voting.js";
export { TaxrailClient } from "./clients/taxrail.js";

export { RestClient, pathSegment } from "./rest.js";
export { TypeUrls, lcvRegistryTypes, createLcvRegistry } from "./registry.js";

export {
  buildKeplrChainInfo,
  suggestLcvChain,
  getKeplrOfflineSigner,
  connectKeplr,
} from "./wallets/keplr.js";
export type { KeplrLike, KeplrChainInfo, LcvChainInfo } from "./wallets/keplr.js";

export {
  toHexChainId,
  buildLcvViemChain,
  addLcvNetworkToMetaMask,
  requestMetaMaskAccounts,
  connectMetaMask,
} from "./wallets/metamask.js";
export type { LcvEvmNetworkInfo } from "./wallets/metamask.js";

export * as identityTx from "./proto/lcv/identity/tx.js";
export * as identityQuery from "./proto/lcv/identity/query.js";
export * as registryTx from "./proto/lcv/registry/tx.js";
export * as registryQuery from "./proto/lcv/registry/query.js";
export * as votingTx from "./proto/lcv/voting/tx.js";
export * as votingQuery from "./proto/lcv/voting/query.js";
export * as taxrailTx from "./proto/lcv/taxrail/tx.js";
export * as taxrailQuery from "./proto/lcv/taxrail/query.js";
