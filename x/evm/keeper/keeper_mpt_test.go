package keeper_test

import (
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/store/mpt"
	"github.com/evoblockchain/evochain/libs/tendermint/types"
	"github.com/stretchr/testify/suite"
	"testing"
)

type KeeperMptTestSuite struct {
	KeeperTestSuite
}

func (suite *KeeperMptTestSuite) SetupTest() {
	mpt.TrieWriteAhead = true
	types.UnittestOnlySetMilestoneMarsHeight(1)

	suite.KeeperTestSuite.SetupTest()
}

func TestKeeperMptTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperMptTestSuite))
}
