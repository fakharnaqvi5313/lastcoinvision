package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/lastcoinvision/lcv/x/taxrail/types"
)

// GetQueryCmd builds hand-written "tax-rule" and "jurisdiction-balance"
// query commands instead of letting AutoCLI generate them. AutoCLI's
// response encoder doesn't render gogoproto's
// `(gogoproto.customtype) = "cosmossdk.io/math.LegacyDec"` field
// annotation specially, so it would print TaxRule.rate as its raw internal
// integer (LegacyDec scales by 1e18 internally) rather than a decimal —
// see modules/feemarket/SPEC.md for the full writeup of this AutoCLI gap,
// and x/feemarket/client/cli/query.go for the first fix of this kind.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the taxrail module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryTaxRule())
	cmd.AddCommand(CmdQueryJurisdictionBalance())

	return cmd
}

func CmdQueryTaxRule() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tax-rule [jurisdiction-code]",
		Short: "Shows the tax rule registered for a jurisdiction code",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.TaxRule(cmd.Context(), &types.QueryTaxRuleRequest{JurisdictionCode: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryJurisdictionBalance() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jurisdiction-balance [jurisdiction-code]",
		Short: "Shows how much tax a jurisdiction has collected and not yet spent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.JurisdictionBalance(cmd.Context(), &types.QueryJurisdictionBalanceRequest{JurisdictionCode: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
