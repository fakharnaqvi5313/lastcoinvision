import { IdentityClient } from "@lastcoinvision/sdk";
import { TESTNET_DEFAULT_FEES } from "@lastcoinvision/wallet-core";
import type { StdFee } from "@cosmjs/stargate";

// Msg/Query types are derived structurally from IdentityClient's own
// method signatures (Parameters/ReturnType) rather than imported by name
// from sdk's re-exported `identityTx`/`identityQuery` namespaces. Those
// namespaces mix a ts-proto-generated interface and a same-named runtime
// const (a deliberate declaration merge) -- referencing the type side
// through a renamed namespace re-export breaks tsup/rollup-plugin-dts's
// cross-package .d.ts generation (TS2749/TS2724). Deriving structurally
// from the class method sdk already exports avoids the whole problem and
// can't drift out of sync with the real method signature.
type CreateDIDMsg = Parameters<IdentityClient["createDID"]>[1];
type IssueCredentialMsg = Parameters<IdentityClient["issueCredential"]>[1];
type PresentCredentialMsg = Parameters<IdentityClient["presentCredential"]>[1];

/**
 * Wraps `sdk`'s `IdentityClient` with the signer address baked in and a
 * required fee that's never the CosmJS auto-fee-estimation sentinel.
 * `sdk/src/clients/identity.ts`'s own convenience methods default `fee` to
 * that sentinel -- broken on this chain (`DECISIONS.md`, 2026-08-03) -- so
 * every method here defaults instead to `TESTNET_DEFAULT_FEES.moduleMsg`,
 * a real fixed fee. There is no parameter shape that can reach the sdk's
 * own default.
 */
export class IdentityTxClient {
  constructor(
    private readonly address: string,
    private readonly client: IdentityClient,
  ) {}

  async createDID(msg: CreateDIDMsg, fee: StdFee = TESTNET_DEFAULT_FEES.moduleMsg, memo = "") {
    return this.client.createDID(this.address, msg, fee, memo);
  }

  async issueCredential(msg: IssueCredentialMsg, fee: StdFee = TESTNET_DEFAULT_FEES.moduleMsg, memo = "") {
    return this.client.issueCredential(this.address, msg, fee, memo);
  }

  async presentCredential(msg: PresentCredentialMsg, fee: StdFee = TESTNET_DEFAULT_FEES.moduleMsg, memo = "") {
    return this.client.presentCredential(this.address, msg, fee, memo);
  }

  // Read-only queries pass straight through -- no signer/fee involved.
  params(): ReturnType<IdentityClient["params"]> {
    return this.client.params();
  }

  didByAddress(address: string = this.address): ReturnType<IdentityClient["didByAddress"]> {
    return this.client.didByAddress(address);
  }

  credentialStatus(id: string): ReturnType<IdentityClient["credentialStatus"]> {
    return this.client.credentialStatus(id);
  }

  issuerList(): ReturnType<IdentityClient["issuerList"]> {
    return this.client.issuerList();
  }

  tierOf(address: string = this.address): ReturnType<IdentityClient["tierOf"]> {
    return this.client.tierOf(address);
  }
}
