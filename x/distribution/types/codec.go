package types

import (
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on codec codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgWithdrawValidatorCommission{}, "evoblockchain/distribution/MsgWithdrawReward", nil)
	cdc.RegisterConcrete(MsgWithdrawDelegatorReward{}, "evoblockchain/distribution/MsgWithdrawDelegatorReward", nil)
	cdc.RegisterConcrete(MsgSetWithdrawAddress{}, "evoblockchain/distribution/MsgModifyWithdrawAddress", nil)
	cdc.RegisterConcrete(CommunityPoolSpendProposal{}, "evoblockchain/distribution/CommunityPoolSpendProposal", nil)
	cdc.RegisterConcrete(ChangeDistributionTypeProposal{}, "evoblockchain/distribution/ChangeDistributionTypeProposal", nil)
	cdc.RegisterConcrete(WithdrawRewardEnabledProposal{}, "evoblockchain/distribution/WithdrawRewardEnabledProposal", nil)
	cdc.RegisterConcrete(RewardTruncatePrecisionProposal{}, "evoblockchain/distribution/RewardTruncatePrecisionProposal", nil)
	cdc.RegisterConcrete(MsgWithdrawDelegatorAllRewards{}, "evoblockchain/distribution/MsgWithdrawDelegatorAllRewards", nil)
}

// ModuleCdc generic sealed codec to be used throughout module
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
