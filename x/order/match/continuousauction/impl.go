package continuousauction

import (
	sdk "github.com/evoblockchain/evochain/libs/cosmos-sdk/types"

	"github.com/evoblockchain/evochain/x/order/keeper"
)

// nolint
type CaEngine struct {
}

// nolint
func (e *CaEngine) Run(ctx sdk.Context, keeper keeper.Keeper) {
	// TODO
}
