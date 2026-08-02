package cmd

import (
	"os"
	"strings"

	"cosmossdk.io/log"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtxconfig "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"

	"github.com/lastcoinvision/lcv/app"
)

// NewRootCmd creates a new root command for lcvd. It is called once in the main function.
//
// The chain is no longer wired via depinject/appconfig (see DECISIONS.md):
// wasmd and cosmos/evm both ship as traditionally-wired modules, so app.New
// builds a plain *baseapp.BaseApp directly. A throwaway in-memory app
// instance is built here purely to read back its already-configured codec,
// tx config, and AutoCLI options, matching the pattern every wasmd/cosmos-evm
// chain uses for this same reason.
func NewRootCmd() *cobra.Command {
	// wasmd's keeper always opens a real wasmvm cache directory on disk
	// (even for a throwaway app never used to run anything), so — exactly
	// like wasmd's own cmd/wasmd/root.go — this needs a real temp home
	// directory, not simtestutil.EmptyAppOptions{}'s empty one, which
	// wasmkeeper would try to use to create a directory at the filesystem
	// root. See https://github.com/CosmWasm/wasmd/issues/2017.
	tempDir, err := os.MkdirTemp("", app.Name+"-tempapp")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)

	tempApp, err := app.New(log.NewNopLogger(), dbm.NewMemDB(), nil, false, simtestutil.NewAppOptionsWithFlagHome(tempDir))
	if err != nil {
		panic(err)
	}

	clientCtx := client.Context{}.
		WithCodec(tempApp.AppCodec()).
		WithInterfaceRegistry(tempApp.InterfaceRegistry()).
		WithTxConfig(tempApp.TxConfig()).
		WithLegacyAmino(tempApp.LegacyAmino()).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithHomeDir(app.DefaultNodeHome).
		WithViper(app.Name) // env variable prefix

	rootCmd := &cobra.Command{
		Use:           app.Name + "d",
		Short:         "Start lcv node",
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// set the default command outputs
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())

			clientCtx := clientCtx.WithCmdContext(cmd.Context())
			clientCtx, err := client.ReadPersistentCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			clientCtx, err = config.ReadFromClientConfig(clientCtx)
			if err != nil {
				return err
			}

			if !clientCtx.Offline {
				txConfigOpts := authtx.ConfigOptions{
					TextualCoinMetadataQueryFn: authtxconfig.NewGRPCCoinMetadataQueryFn(clientCtx),
				}
				txConfig, err := authtx.NewTxConfigWithOptions(clientCtx.Codec, txConfigOpts)
				if err != nil {
					return err
				}
				clientCtx = clientCtx.WithTxConfig(txConfig)
			}

			if err := client.SetCmdClientContextHandler(clientCtx, cmd); err != nil {
				return err
			}

			customAppTemplate, customAppConfig := initAppConfig()
			customCMTConfig := initCometBFTConfig()

			return server.InterceptConfigsPreRunHandler(cmd, customAppTemplate, customAppConfig, customCMTConfig)
		},
	}

	initRootCmd(rootCmd, tempApp.TxConfig(), tempApp.BasicModuleManager)

	overwriteFlagDefaults(rootCmd, map[string]string{
		flags.FlagChainID:        strings.ReplaceAll(app.Name, "-", ""),
		flags.FlagKeyringBackend: "test",
	})

	autoCliOpts := tempApp.AutoCliOpts()
	autoCliOpts.ClientCtx = clientCtx
	if err := autoCliOpts.EnhanceRootCommand(rootCmd); err != nil {
		panic(err)
	}

	return rootCmd
}

func overwriteFlagDefaults(c *cobra.Command, defaults map[string]string) {
	set := func(s *pflag.FlagSet, key, val string) {
		if f := s.Lookup(key); f != nil {
			f.DefValue = val
			_ = f.Value.Set(val)
		}
	}
	for key, val := range defaults {
		set(c.Flags(), key, val)
		set(c.PersistentFlags(), key, val)
	}
	for _, c := range c.Commands() {
		overwriteFlagDefaults(c, defaults)
	}
}
