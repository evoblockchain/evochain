package periodicauction

import (
	sdk "github.com/evoblockchain/evochain/libs/cosmos-sdk/types"

	"github.com/evoblockchain/evochain/x/order/keeper"
)

// PaEngine is the periodic auction match engine
type PaEngine struct {
}

// nolint
func (e *PaEngine) Run(ctx sdk.Context, keeper keeper.Keeper) {
	cleanupExpiredOrders(ctx, keeper)
	cleanupOrdersWhoseTokenPairHaveBeenDelisted(ctx, keeper)
	matchOrders(ctx, keeper)
}
