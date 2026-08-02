package app

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	icacontroller "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller"
	icacontrollertypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/types"
	icahost "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host"
	icahosttypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host/types"
	ibccallbacks "github.com/cosmos/ibc-go/v10/modules/apps/callbacks"
	ibctransfer "github.com/cosmos/ibc-go/v10/modules/apps/transfer"
	ibctransfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	transferv2 "github.com/cosmos/ibc-go/v10/modules/apps/transfer/v2"
	porttypes "github.com/cosmos/ibc-go/v10/modules/core/05-port/types"
	ibcapi "github.com/cosmos/ibc-go/v10/modules/core/api"
	ibctm "github.com/cosmos/ibc-go/v10/modules/light-clients/07-tendermint"
)

// registerIBCModules wires the IBC transfer, interchain-accounts, and
// CosmWasm IBC routes onto app.IBCKeeper's router, and registers the
// tendermint light client module. Called after every keeper the stack
// depends on (TransferKeeper, ICA keepers, WasmKeeper) already exists.
//
// ibc-go v10 dropped the x/capability module entirely (port/channel
// ownership no longer needs a separate capability system) — see
// DECISIONS.md for why this chain is on v10 at all (wasmd/cosmos-evm both
// require it), and note this does mean the ICS-29 fee middleware
// (`29-fee`), which v10 also dropped, is no longer available; incentivized
// packet relaying is not supported now.
func (app *App) registerIBCModules() error {
	wasmStackIBCHandler := wasm.NewIBCHandler(app.WasmKeeper, app.IBCKeeper.ChannelKeeper, app.TransferKeeper, app.IBCKeeper.ChannelKeeper)

	var icaControllerStack porttypes.IBCModule
	var noAuthzModule porttypes.IBCModule
	icaControllerStack = icacontroller.NewIBCMiddlewareWithAuth(noAuthzModule, app.ICAControllerKeeper)
	icaControllerStack = ibccallbacks.NewIBCMiddleware(icaControllerStack, app.IBCKeeper.ChannelKeeper, wasmStackIBCHandler, wasm.DefaultMaxIBCCallbackGas)
	icaICS4Wrapper := icaControllerStack.(porttypes.ICS4Wrapper)
	app.ICAControllerKeeper.WithICS4Wrapper(icaICS4Wrapper)

	icaHostStack := icahost.NewIBCModule(app.ICAHostKeeper)

	var transferStack porttypes.IBCModule
	transferStack = ibctransfer.NewIBCModule(app.TransferKeeper)
	transferStack = ibccallbacks.NewIBCMiddleware(transferStack, app.IBCKeeper.ChannelKeeper, wasmStackIBCHandler, wasm.DefaultMaxIBCCallbackGas)
	transferICS4Wrapper := transferStack.(porttypes.ICS4Wrapper)
	app.TransferKeeper.WithICS4Wrapper(transferICS4Wrapper)

	ibcRouter := porttypes.NewRouter().
		AddRoute(ibctransfertypes.ModuleName, transferStack).
		AddRoute(wasmtypes.ModuleName, wasmStackIBCHandler).
		AddRoute(icacontrollertypes.SubModuleName, icaControllerStack).
		AddRoute(icahosttypes.SubModuleName, icaHostStack)
	app.IBCKeeper.SetRouter(ibcRouter)

	ibcRouterV2 := ibcapi.NewRouter().
		AddRoute(ibctransfertypes.PortID, transferv2.NewIBCModule(app.TransferKeeper)).
		AddPrefixRoute(wasmkeeper.PortIDPrefixV2, wasmkeeper.NewIBC2Handler(app.WasmKeeper))
	app.IBCKeeper.SetRouterV2(ibcRouterV2)

	clientKeeper := app.IBCKeeper.ClientKeeper
	storeProvider := app.IBCKeeper.ClientKeeper.GetStoreProvider()
	app.tmLightClientModule = ibctm.NewLightClientModule(app.appCodec, storeProvider)
	clientKeeper.AddRoute(ibctm.ModuleName, &app.tmLightClientModule)

	return nil
}
