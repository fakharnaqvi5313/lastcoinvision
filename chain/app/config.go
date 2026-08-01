package app

import sdk "github.com/cosmos/cosmos-sdk/types"

func init() {
	// Without this, sdk.DefaultBondDenom stays "stake", which leaks into
	// every tool that references it instead of an explicit denom flag:
	// `lcvd multi-node`'s validator self-bond amounts and default
	// min-gas-prices, `lcvd testnet` defaults, etc. Overwriting it is the
	// SDK-documented way to change that (see cosmos-sdk types/staking.go).
	sdk.DefaultBondDenom = "ulcv"

	// Set prefixes
	accountPubKeyPrefix := AccountAddressPrefix + "pub"
	validatorAddressPrefix := AccountAddressPrefix + "valoper"
	validatorPubKeyPrefix := AccountAddressPrefix + "valoperpub"
	consNodeAddressPrefix := AccountAddressPrefix + "valcons"
	consNodePubKeyPrefix := AccountAddressPrefix + "valconspub"

	// Set and seal config
	config := sdk.GetConfig()
	config.SetCoinType(ChainCoinType)
	config.SetBech32PrefixForAccount(AccountAddressPrefix, accountPubKeyPrefix)
	config.SetBech32PrefixForValidator(validatorAddressPrefix, validatorPubKeyPrefix)
	config.SetBech32PrefixForConsensusNode(consNodeAddressPrefix, consNodePubKeyPrefix)
	config.Seal()
}
