package keeper_test

import (
	abci "github.com/evoblockchain/evochain/libs/tendermint/abci/types"

	"github.com/evoblockchain/evochain/libs/cosmos-sdk/simapp"
	sdk "github.com/evoblockchain/evochain/libs/cosmos-sdk/types"
	authtypes "github.com/evoblockchain/evochain/libs/cosmos-sdk/x/auth/types"
)

// returns context and app with params set on account keeper
func createTestApp(isCheckTx bool) (*simapp.SimApp, sdk.Context) {
	app := simapp.Setup(isCheckTx)
	ctx := app.BaseApp.NewContext(isCheckTx, abci.Header{})
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())

	return app, ctx
}

func createTestAppWithHeight(isCheckTx bool, height int64) (*simapp.SimApp, sdk.Context) {
	app := simapp.Setup(isCheckTx)
	ctx := app.BaseApp.NewContext(isCheckTx, abci.Header{Height: height})
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())

	return app, ctx
}
