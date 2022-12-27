package erc20_test

import (
	"testing"
	"time"

	"github.com/evoblockchain/evochain/app"
	sdk "github.com/evoblockchain/evochain/libs/cosmos-sdk/types"
	abci "github.com/evoblockchain/evochain/libs/tendermint/abci/types"
	"github.com/evoblockchain/evochain/x/erc20"
	"github.com/evoblockchain/evochain/x/erc20/types"
	"github.com/stretchr/testify/suite"
)

type Erc20TestSuite struct {
	suite.Suite

	ctx     sdk.Context
	handler sdk.Handler
	app     *app.EVOChainApp
}

func TestErc20TestSuite(t *testing.T) {
	suite.Run(t, new(Erc20TestSuite))
}

func (suite *Erc20TestSuite) SetupTest() {
	checkTx := false

	suite.app = app.Setup(false)
	suite.ctx = suite.app.BaseApp.NewContext(checkTx, abci.Header{Height: 1, ChainID: "ethermint-3", Time: time.Now().UTC()})
	suite.handler = erc20.NewHandler(suite.app.Erc20Keeper)
	suite.app.Erc20Keeper.SetParams(suite.ctx, types.DefaultParams())
}
