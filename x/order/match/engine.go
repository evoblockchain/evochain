package match

import (
	"sync"

	sdk "github.com/evoblockchain/evochain/libs/cosmos-sdk/types"

	"github.com/evoblockchain/evochain/x/order/keeper"
	"github.com/evoblockchain/evochain/x/order/match/continuousauction"
	"github.com/evoblockchain/evochain/x/order/match/periodicauction"
)

// nolint
const DefaultAuctionType = "periodicauction"

// nolint
var (
	once        sync.Once
	engine      Engine
	auctionType = DefaultAuctionType
)

// GetEngine : periodic auction only today
func GetEngine() Engine {
	once.Do(func() {
		if auctionType == DefaultAuctionType {
			engine = &periodicauction.PaEngine{}
		} else {
			engine = &continuousauction.CaEngine{}
		}
	})
	return engine
}

// nolint
type Engine interface {
	Run(ctx sdk.Context, keeper keeper.Keeper)
}
