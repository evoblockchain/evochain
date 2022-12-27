package types

import (
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgAddLiquidity{}, "evoblockchain/ammswap/MsgAddLiquidity", nil)
	cdc.RegisterConcrete(MsgRemoveLiquidity{}, "evoblockchain/ammswap/MsgRemoveLiquidity", nil)
	cdc.RegisterConcrete(MsgCreateExchange{}, "evoblockchain/ammswap/MsgCreateExchange", nil)
	cdc.RegisterConcrete(MsgTokenToToken{}, "evoblockchain/ammswap/MsgSwapToken", nil)
}

// ModuleCdc defines the module codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
