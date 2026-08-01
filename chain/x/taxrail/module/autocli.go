package taxrail

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"github.com/spf13/cobra"

	modulev1 "github.com/lastcoinvision/lcv/api/lcv/taxrail"
	"github.com/lastcoinvision/lcv/x/taxrail/client/cli"
)

// GetQueryCmd implements the autocli.HasCustomQueryCommand interface. It
// takes over the "tax-rule" and "jurisdiction-balance" queries specifically
// so their output goes through codec.ProtoCodec.MarshalJSON (via
// clientCtx.PrintProto) instead of AutoCLI's default aminojson response
// encoder, which doesn't know how to render the gogoproto customtype
// LegacyDec/Int fields in TaxRule/JurisdictionLedger — see
// x/taxrail/client/cli/query.go.
func (am AppModule) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface. Query's
// EnhanceCustomCommand is set so AutoCLI still layers in any RPCs not
// already present in GetQueryCmd's hand-written tree (e.g. "params").
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service:              modulev1.Query_ServiceDesc.ServiceName,
			EnhanceCustomCommand: true,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Shows the parameters of the module",
				},
				{
					RpcMethod: "TaxRule",
					Skip:      true, // served by the hand-written command in GetQueryCmd
				},
				{
					RpcMethod: "JurisdictionBalance",
					Skip:      true, // served by the hand-written command in GetQueryCmd
				},
				// this line is used by ignite scaffolding # autocli/query
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              modulev1.Msg_ServiceDesc.ServiceName,
			EnhanceCustomCommand: true, // only required if you want to use the custom command
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "UpdateParams",
					Skip:      true, // skipped because authority gated
				},
				// this line is used by ignite scaffolding # autocli/tx
			},
		},
	}
}
