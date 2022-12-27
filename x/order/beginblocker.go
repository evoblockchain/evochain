package order

import (
	sdk "github.com/evoblockchain/evochain/libs/cosmos-sdk/types"

	"github.com/evoblockchain/evochain/x/common/perf"
	"github.com/evoblockchain/evochain/x/order/keeper"
	"github.com/evoblockchain/evochain/x/order/types"
	//"github.com/evoblockchain/evochain/x/common/version"
)

// BeginBlocker runs the logic of BeginBlocker with version 0.
// BeginBlocker resets keeper cache.
func BeginBlocker(ctx sdk.Context, keeper keeper.Keeper) {
	seq := perf.GetPerf().OnBeginBlockEnter(ctx, types.ModuleName)
	defer perf.GetPerf().OnBeginBlockExit(ctx, types.ModuleName, seq)

	keeper.ResetCache(ctx)
}
