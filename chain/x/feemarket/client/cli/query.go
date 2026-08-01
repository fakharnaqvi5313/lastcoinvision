package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/lastcoinvision/lcv/x/feemarket/types"
)

// GetQueryCmd builds a hand-written "params" query command instead of
// letting AutoCLI generate one. AutoCLI's response encoder
// (cosmossdk.io/client/v2/autocli's aminojson-based printer) only special-
// cases a couple of well-known proto message types (google.protobuf.Duration,
// cosmos.base.v1beta1.DecCoin) — it has no hook for gogoproto's
// `(gogoproto.customtype) = "cosmossdk.io/math.LegacyDec"` field annotation,
// so it prints those fields as their raw internal integer (LegacyDec scales
// by 1e18 internally), e.g. "500000000000000000" instead of "0.5". The
// classic `clientCtx.PrintProto` path (used here) goes through
// `codec.ProtoCodec.MarshalJSON`, which does respect customtype and renders
// correctly — see modules/feemarket/SPEC.md for the full writeup.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the feemarket module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryParams())

	return cmd
}

func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Shows the parameters of the module",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
