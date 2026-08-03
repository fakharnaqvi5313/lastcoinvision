import { GeneratedType, Registry } from "@cosmjs/proto-signing";
import { defaultRegistryTypes } from "@cosmjs/stargate";

import * as identityTx from "./proto/lcv/identity/tx.js";
import * as registryTx from "./proto/lcv/registry/tx.js";
import * as taxrailTx from "./proto/lcv/taxrail/tx.js";
import * as votingTx from "./proto/lcv/voting/tx.js";

/**
 * Type URLs for every custom-module Msg this SDK knows how to encode. Kept
 * as named constants (rather than inlined "/lcv.xxx.MsgYyy" string
 * literals scattered across the client classes) so a typo shows up as a
 * missing export at compile time instead of a silent broadcast failure.
 */
export const TypeUrls = {
  identity: {
    createDID: "/lcv.identity.MsgCreateDID",
    updateDID: "/lcv.identity.MsgUpdateDID",
    deactivateDID: "/lcv.identity.MsgDeactivateDID",
    addIssuer: "/lcv.identity.MsgAddIssuer",
    removeIssuer: "/lcv.identity.MsgRemoveIssuer",
    issueCredential: "/lcv.identity.MsgIssueCredential",
    revokeCredential: "/lcv.identity.MsgRevokeCredential",
    presentCredential: "/lcv.identity.MsgPresentCredential",
  },
  registry: {
    setJurisdictionAuthority: "/lcv.registry.MsgSetJurisdictionAuthority",
    registerAsset: "/lcv.registry.MsgRegisterAsset",
    transferAsset: "/lcv.registry.MsgTransferAsset",
    freezeAsset: "/lcv.registry.MsgFreezeAsset",
  },
  voting: {
    addElectoralAuthority: "/lcv.voting.MsgAddElectoralAuthority",
    removeElectoralAuthority: "/lcv.voting.MsgRemoveElectoralAuthority",
    createElection: "/lcv.voting.MsgCreateElection",
    registerToVote: "/lcv.voting.MsgRegisterToVote",
    commitVote: "/lcv.voting.MsgCommitVote",
    revealVote: "/lcv.voting.MsgRevealVote",
  },
  taxrail: {
    registerTaxRule: "/lcv.taxrail.MsgRegisterTaxRule",
    taxedTransfer: "/lcv.taxrail.MsgTaxedTransfer",
    spendCollectedTax: "/lcv.taxrail.MsgSpendCollectedTax",
  },
} as const;

/** All (typeUrl, GeneratedType) pairs this SDK can encode/decode. */
export function lcvRegistryTypes(): ReadonlyArray<[string, GeneratedType]> {
  return [
    [TypeUrls.identity.createDID, identityTx.MsgCreateDID as unknown as GeneratedType],
    [TypeUrls.identity.updateDID, identityTx.MsgUpdateDID as unknown as GeneratedType],
    [TypeUrls.identity.deactivateDID, identityTx.MsgDeactivateDID as unknown as GeneratedType],
    [TypeUrls.identity.addIssuer, identityTx.MsgAddIssuer as unknown as GeneratedType],
    [TypeUrls.identity.removeIssuer, identityTx.MsgRemoveIssuer as unknown as GeneratedType],
    [TypeUrls.identity.issueCredential, identityTx.MsgIssueCredential as unknown as GeneratedType],
    [TypeUrls.identity.revokeCredential, identityTx.MsgRevokeCredential as unknown as GeneratedType],
    [TypeUrls.identity.presentCredential, identityTx.MsgPresentCredential as unknown as GeneratedType],

    [TypeUrls.registry.setJurisdictionAuthority, registryTx.MsgSetJurisdictionAuthority as unknown as GeneratedType],
    [TypeUrls.registry.registerAsset, registryTx.MsgRegisterAsset as unknown as GeneratedType],
    [TypeUrls.registry.transferAsset, registryTx.MsgTransferAsset as unknown as GeneratedType],
    [TypeUrls.registry.freezeAsset, registryTx.MsgFreezeAsset as unknown as GeneratedType],

    [TypeUrls.voting.addElectoralAuthority, votingTx.MsgAddElectoralAuthority as unknown as GeneratedType],
    [TypeUrls.voting.removeElectoralAuthority, votingTx.MsgRemoveElectoralAuthority as unknown as GeneratedType],
    [TypeUrls.voting.createElection, votingTx.MsgCreateElection as unknown as GeneratedType],
    [TypeUrls.voting.registerToVote, votingTx.MsgRegisterToVote as unknown as GeneratedType],
    [TypeUrls.voting.commitVote, votingTx.MsgCommitVote as unknown as GeneratedType],
    [TypeUrls.voting.revealVote, votingTx.MsgRevealVote as unknown as GeneratedType],

    [TypeUrls.taxrail.registerTaxRule, taxrailTx.MsgRegisterTaxRule as unknown as GeneratedType],
    [TypeUrls.taxrail.taxedTransfer, taxrailTx.MsgTaxedTransfer as unknown as GeneratedType],
    [TypeUrls.taxrail.spendCollectedTax, taxrailTx.MsgSpendCollectedTax as unknown as GeneratedType],
  ];
}

/**
 * Builds a CosmJS Registry containing the standard Cosmos SDK message types
 * plus every LCV custom-module message this SDK supports. Pass the result
 * to `SigningStargateClient.connectWithSigner(rpcEndpoint, signer, { registry })`.
 */
export function createLcvRegistry(): Registry {
  return new Registry([...defaultRegistryTypes, ...lcvRegistryTypes()]);
}
