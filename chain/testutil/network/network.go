package network

import (
	"fmt"
	"testing"

	"cosmossdk.io/log"
	pruningtypes "cosmossdk.io/store/pruning/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil/network"

	"github.com/lastcoinvision/lcv/app"
)

type (
	Network = network.Network
	Config  = network.Config
)

// New creates instance with fully configured cosmos network.
// Accepts optional config, that will be used in place of the DefaultConfig() if provided.
func New(t *testing.T, configs ...Config) *Network {
	t.Helper()
	if len(configs) > 1 {
		panic("at most one config should be provided")
	}
	var cfg network.Config
	if len(configs) == 0 {
		cfg = DefaultConfig()
	} else {
		cfg = configs[0]
	}
	net, err := network.New(t, t.TempDir(), cfg)
	require.NoError(t, err)
	_, err = net.WaitForHeight(1)
	require.NoError(t, err)
	t.Cleanup(net.Cleanup)
	return net
}

// DefaultConfig will initialize config for the network with custom application,
// genesis and single validator. All other parameters are inherited from
// cosmos-sdk/testutil/network.DefaultConfig.
//
// The chain is no longer depinject/appconfig-wired (see DECISIONS.md), so
// unlike network.DefaultConfigWithAppConfig this builds its own
// TestFixtureFactory/AppConstructor directly against app.New.
func DefaultConfig() network.Config {
	cfg := network.DefaultConfig(func() network.TestFixture {
		tempApp, err := app.New(log.NewNopLogger(), dbm.NewMemDB(), nil, true, simtestutil.EmptyAppOptions{})
		if err != nil {
			panic(err)
		}

		return network.TestFixture{
			EncodingConfig: moduletestutil.TestEncodingConfig{
				Codec:             tempApp.AppCodec(),
				TxConfig:          tempApp.TxConfig(),
				Amino:             tempApp.LegacyAmino(),
				InterfaceRegistry: tempApp.InterfaceRegistry(),
			},
			GenesisState: tempApp.BasicModuleManager.DefaultGenesis(tempApp.AppCodec()),
		}
	})

	// AppConstructor builds one app instance per validator; it references
	// cfg.ChainID, so it's assigned after DefaultConfig has populated cfg's
	// other defaults rather than from inside the TestFixtureFactory above.
	cfg.AppConstructor = func(val network.ValidatorI) servertypes.Application {
		pruningOpts := pruningtypes.NewPruningOptionsFromString(val.GetAppConfig().Pruning)
		valApp, err := app.New(
			val.GetCtx().Logger, dbm.NewMemDB(), nil, true, simtestutil.EmptyAppOptions{},
			baseapp.SetPruning(pruningOpts),
			baseapp.SetMinGasPrices(val.GetAppConfig().MinGasPrices),
			baseapp.SetChainID(cfg.ChainID),
		)
		if err != nil {
			panic(err)
		}
		return valApp
	}

	ports, err := freePorts(3)
	if err != nil {
		panic(err)
	}
	if cfg.APIAddress == "" {
		cfg.APIAddress = fmt.Sprintf("tcp://0.0.0.0:%s", ports[0])
	}
	if cfg.RPCAddress == "" {
		cfg.RPCAddress = fmt.Sprintf("tcp://0.0.0.0:%s", ports[1])
	}
	if cfg.GRPCAddress == "" {
		cfg.GRPCAddress = fmt.Sprintf("0.0.0.0:%s", ports[2])
	}
	return cfg
}

// freePorts return the available ports based on the number of requested ports.
func freePorts(n int) ([]string, error) {
	closeFns := make([]func() error, n)
	ports := make([]string, n)
	for i := 0; i < n; i++ {
		_, port, closeFn, err := network.FreeTCPAddr()
		if err != nil {
			return nil, err
		}
		ports[i] = port
		closeFns[i] = closeFn
	}
	for _, closeFn := range closeFns {
		if err := closeFn(); err != nil {
			return nil, err
		}
	}
	return ports, nil
}
