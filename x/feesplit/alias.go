package feesplit

import (
	"github.com/evoblockchain/evochain/x/feesplit/keeper"
	"github.com/evoblockchain/evochain/x/feesplit/types"
)

const (
	ModuleName = types.ModuleName
	StoreKey   = types.StoreKey
	RouterKey  = types.RouterKey
)

var (
	NewKeeper           = keeper.NewKeeper
	SetParamsNeedUpdate = types.SetParamsNeedUpdate
)

type (
	Keeper = keeper.Keeper
)
