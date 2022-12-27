package keeper

import (
	sdk "github.com/evoblockchain/evochain/libs/cosmos-sdk/types"
	evmtypes "github.com/evoblockchain/evochain/x/evm/types"
)

type EvmKeeper interface {
	GetChainConfig(ctx sdk.Context) (evmtypes.ChainConfig, bool)
	GenerateCSDBParams() evmtypes.CommitStateDBParams
	GetParams(ctx sdk.Context) evmtypes.Params
	AddInnerTx(...interface{})
	AddContract(...interface{})
}
