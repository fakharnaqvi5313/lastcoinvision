import { RegistryClient } from "@lastcoinvision/sdk";
import { TESTNET_DEFAULT_FEES } from "@lastcoinvision/wallet-core";
import type { StdFee } from "@cosmjs/stargate";

// See identity.ts's comment on why these are derived structurally rather
// than imported by name from sdk's `registryTx`/`registryQuery` namespaces.
type RegisterAssetMsg = Parameters<RegistryClient["registerAsset"]>[1];
type TransferAssetMsg = Parameters<RegistryClient["transferAsset"]>[1];
type FreezeAssetMsg = Parameters<RegistryClient["freezeAsset"]>[1];

/** See identity.ts's doc comment -- same pattern for x/registry. */
export class RegistryTxClient {
  constructor(
    private readonly address: string,
    private readonly client: RegistryClient,
  ) {}

  async registerAsset(msg: RegisterAssetMsg, fee: StdFee = TESTNET_DEFAULT_FEES.moduleMsg, memo = "") {
    return this.client.registerAsset(this.address, msg, fee, memo);
  }

  async transferAsset(msg: TransferAssetMsg, fee: StdFee = TESTNET_DEFAULT_FEES.moduleMsg, memo = "") {
    return this.client.transferAsset(this.address, msg, fee, memo);
  }

  async freezeAsset(msg: FreezeAssetMsg, fee: StdFee = TESTNET_DEFAULT_FEES.moduleMsg, memo = "") {
    return this.client.freezeAsset(this.address, msg, fee, memo);
  }

  params(): ReturnType<RegistryClient["params"]> {
    return this.client.params();
  }

  assetByID(id: string): ReturnType<RegistryClient["assetByID"]> {
    return this.client.assetByID(id);
  }

  transferHistory(assetId: string): ReturnType<RegistryClient["transferHistory"]> {
    return this.client.transferHistory(assetId);
  }

  jurisdictionAuthorityOf(jurisdictionCode: string): ReturnType<RegistryClient["jurisdictionAuthorityOf"]> {
    return this.client.jurisdictionAuthorityOf(jurisdictionCode);
  }
}
