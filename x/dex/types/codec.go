package types

import "github.com/evoblockchain/evochain/libs/cosmos-sdk/codec"

// RegisterCodec registers concrete types on the Amino codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgList{}, "evoblockchain/dex/MsgList", nil)
	cdc.RegisterConcrete(MsgDeposit{}, "evoblockchain/dex/MsgDeposit", nil)
	cdc.RegisterConcrete(MsgWithdraw{}, "evoblockchain/dex/MsgWithdraw", nil)
	cdc.RegisterConcrete(MsgTransferOwnership{}, "evoblockchain/dex/MsgTransferTradingPairOwnership", nil)
	cdc.RegisterConcrete(MsgConfirmOwnership{}, "evoblockchain/dex/MsgConfirmOwnership", nil)
	cdc.RegisterConcrete(DelistProposal{}, "evoblockchain/dex/DelistProposal", nil)
	cdc.RegisterConcrete(MsgCreateOperator{}, "evoblockchain/dex/CreateOperator", nil)
	cdc.RegisterConcrete(MsgUpdateOperator{}, "evoblockchain/dex/UpdateOperator", nil)
}

// ModuleCdc represents generic sealed codec to be used throughout this module
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
