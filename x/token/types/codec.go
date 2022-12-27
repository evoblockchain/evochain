package types

import (
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on the Amino codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgTokenIssue{}, "evoblockchain/token/MsgIssue", nil)
	cdc.RegisterConcrete(MsgTokenBurn{}, "evoblockchain/token/MsgBurn", nil)
	cdc.RegisterConcrete(MsgTokenMint{}, "evoblockchain/token/MsgMint", nil)
	cdc.RegisterConcrete(MsgMultiSend{}, "evoblockchain/token/MsgMultiTransfer", nil)
	cdc.RegisterConcrete(MsgSend{}, "evoblockchain/token/MsgTransfer", nil)
	cdc.RegisterConcrete(MsgTransferOwnership{}, "evoblockchain/token/MsgTransferOwnership", nil)
	cdc.RegisterConcrete(MsgConfirmOwnership{}, "evoblockchain/token/MsgConfirmOwnership", nil)
	cdc.RegisterConcrete(MsgTokenModify{}, "evoblockchain/token/MsgModify", nil)

	// for test
	//cdc.RegisterConcrete(MsgTokenDestroy{}, "evoblockchain/token/MsgDestroy", nil)
}

// generic sealed codec to be used throughout this module
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
