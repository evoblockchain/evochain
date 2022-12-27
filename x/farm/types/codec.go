package types

import (
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgCreatePool{}, "evoblockchain/farm/MsgCreatePool", nil)
	cdc.RegisterConcrete(MsgDestroyPool{}, "evoblockchain/farm/MsgDestroyPool", nil)
	cdc.RegisterConcrete(MsgLock{}, "evoblockchain/farm/MsgLock", nil)
	cdc.RegisterConcrete(MsgUnlock{}, "evoblockchain/farm/MsgUnlock", nil)
	cdc.RegisterConcrete(MsgClaim{}, "evoblockchain/farm/MsgClaim", nil)
	cdc.RegisterConcrete(MsgProvide{}, "evoblockchain/farm/MsgProvide", nil)
	cdc.RegisterConcrete(ManageWhiteListProposal{}, "evoblockchain/farm/ManageWhiteListProposal", nil)
}

// ModuleCdc defines the module codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
