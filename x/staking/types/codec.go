package types

import (
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types for codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgCreateValidator{}, "evoblockchain/staking/MsgCreateValidator", nil)
	cdc.RegisterConcrete(MsgEditValidator{}, "evoblockchain/staking/MsgEditValidator", nil)
	cdc.RegisterConcrete(MsgEditValidatorCommissionRate{}, "evoblockchain/staking/MsgEditValidatorCommissionRate", nil)
	cdc.RegisterConcrete(MsgDestroyValidator{}, "evoblockchain/staking/MsgDestroyValidator", nil)
	cdc.RegisterConcrete(MsgDeposit{}, "evoblockchain/staking/MsgDeposit", nil)
	cdc.RegisterConcrete(MsgWithdraw{}, "evoblockchain/staking/MsgWithdraw", nil)
	cdc.RegisterConcrete(MsgAddShares{}, "evoblockchain/staking/MsgAddShares", nil)
	cdc.RegisterConcrete(MsgRegProxy{}, "evoblockchain/staking/MsgRegProxy", nil)
	cdc.RegisterConcrete(MsgBindProxy{}, "evoblockchain/staking/MsgBindProxy", nil)
	cdc.RegisterConcrete(MsgUnbindProxy{}, "evoblockchain/staking/MsgUnbindProxy", nil)
}

// ModuleCdc is generic sealed codec to be used throughout this module
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
