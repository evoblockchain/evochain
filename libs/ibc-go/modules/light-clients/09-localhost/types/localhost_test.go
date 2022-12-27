package types_test

import (
	"testing"

	tmproto "github.com/evoblockchain/evochain/libs/tendermint/abci/types"
	"github.com/stretchr/testify/suite"

	"github.com/evoblockchain/evochain/libs/cosmos-sdk/codec"
	sdk "github.com/evoblockchain/evochain/libs/cosmos-sdk/types"
	clienttypes "github.com/evoblockchain/evochain/libs/ibc-go/modules/core/02-client/types"
	"github.com/evoblockchain/evochain/libs/ibc-go/modules/core/exported"
	"github.com/evoblockchain/evochain/libs/ibc-go/testing/simapp"
)

const (
	height = 4
)

var (
	clientHeight = clienttypes.NewHeight(0, 10)
)

type LocalhostTestSuite struct {
	suite.Suite

	cdc   *codec.CodecProxy
	ctx   sdk.Context
	store sdk.KVStore
}

func (suite *LocalhostTestSuite) SetupTest() {
	isCheckTx := false
	app := simapp.Setup(isCheckTx)

	suite.cdc = app.AppCodec()
	suite.ctx = app.BaseApp.NewContext(isCheckTx, tmproto.Header{Height: 1, ChainID: "ibc-chain"})
	suite.store = app.IBCKeeper.ClientKeeper.ClientStore(suite.ctx, exported.Localhost)
}

func TestLocalhostTestSuite(t *testing.T) {
	suite.Run(t, new(LocalhostTestSuite))
}
