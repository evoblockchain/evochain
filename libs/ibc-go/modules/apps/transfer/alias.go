package transfer

import (
	"github.com/evoblockchain/evochain/libs/ibc-go/modules/apps/transfer/keeper"
	"github.com/evoblockchain/evochain/libs/ibc-go/modules/apps/transfer/types"
)

var (
	NewKeeper  = keeper.NewKeeper
	ModuleCdc  = types.ModuleCdc
	SetMarshal = types.SetMarshal
)
