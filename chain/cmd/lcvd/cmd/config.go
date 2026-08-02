package cmd

import (
	cmtcfg "github.com/cometbft/cometbft/config"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"

	evmconfig "github.com/cosmos/evm/config"
	cosmosevmserverconfig "github.com/cosmos/evm/server/config"

	"github.com/lastcoinvision/lcv/app"
)

// initCometBFTConfig helps to override default CometBFT Config values.
// return cmtcfg.DefaultConfig if no custom configuration is required for the application.
func initCometBFTConfig() *cmtcfg.Config {
	cfg := cmtcfg.DefaultConfig()

	// these values put a higher strain on node memory
	// cfg.P2P.MaxNumInboundPeers = 100
	// cfg.P2P.MaxNumOutboundPeers = 40

	return cfg
}

// initAppConfig helps to override default appConfig template and configs.
//
// Combines the standard SDK app.toml template with cosmos/evm's own [evm],
// [json-rpc], and [tls] sections (EVMAppTemplate) so a fresh `lcvd init`
// already has JSON-RPC enabled by default at the chain's configured
// EVMChainID — see modules/evm/SPEC.md.
func initAppConfig() (string, interface{}) {
	srvCfg := serverconfig.DefaultConfig()
	// The SDK's default minimum gas price is "" (empty), which halts a
	// validator on startup if left unset. This chain's real fee floor is
	// enforced by x/feemarket's own ante decorator regardless, so 0 here
	// just satisfies the SDK's own startup check.
	srvCfg.MinGasPrices = "0" + app.EVMBaseDenom

	evmCfg := cosmosevmserverconfig.DefaultEVMConfig()
	evmCfg.EVMChainID = app.EVMChainID

	customAppConfig := evmconfig.EVMAppConfig{
		Config:  *srvCfg,
		EVM:     *evmCfg,
		JSONRPC: *cosmosevmserverconfig.DefaultJSONRPCConfig(),
		TLS:     *cosmosevmserverconfig.DefaultTLSConfig(),
	}

	customAppTemplate := serverconfig.DefaultConfigTemplate + cosmosevmserverconfig.DefaultEVMConfigTemplate

	return customAppTemplate, customAppConfig
}
