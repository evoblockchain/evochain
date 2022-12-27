package feesplit

import (
	"encoding/json"
	"errors"

	"github.com/evoblockchain/evochain/libs/cosmos-sdk/baseapp"
	sdk "github.com/evoblockchain/evochain/libs/cosmos-sdk/types"
	tmtypes "github.com/evoblockchain/evochain/libs/tendermint/types"
	"github.com/evoblockchain/evochain/x/common"
	"github.com/evoblockchain/evochain/x/feesplit/types"
)

var (
	ErrCheckSignerFail = errors.New("check signer fail")
)

func init() {
	RegisterConvert()
}

func RegisterConvert() {
	enableHeight := tmtypes.GetVenus3Height()
	baseapp.RegisterCmHandle("evoblockchain/MsgRegisterFeeSplit", baseapp.NewCMHandle(ConvertRegisterFeeSplitMsg, enableHeight))
	baseapp.RegisterCmHandle("evoblockchain/MsgUpdateFeeSplit", baseapp.NewCMHandle(ConvertUpdateFeeSplitMsg, enableHeight))
	baseapp.RegisterCmHandle("evoblockchain/MsgCancelFeeSplit", baseapp.NewCMHandle(ConvertCancelFeeSplitMsg, enableHeight))
}

func ConvertRegisterFeeSplitMsg(data []byte, signers []sdk.AccAddress) (sdk.Msg, error) {
	newMsg := types.MsgRegisterFeeSplit{}
	err := json.Unmarshal(data, &newMsg)
	if err != nil {
		return nil, err
	}
	err = newMsg.ValidateBasic()
	if err != nil {
		return nil, err
	}
	if ok := common.CheckSignerAddress(signers, newMsg.GetSigners()); !ok {
		return nil, ErrCheckSignerFail
	}
	return newMsg, nil
}

func ConvertUpdateFeeSplitMsg(data []byte, signers []sdk.AccAddress) (sdk.Msg, error) {
	newMsg := types.MsgUpdateFeeSplit{}
	err := json.Unmarshal(data, &newMsg)
	if err != nil {
		return nil, err
	}
	err = newMsg.ValidateBasic()
	if err != nil {
		return nil, err
	}
	if ok := common.CheckSignerAddress(signers, newMsg.GetSigners()); !ok {
		return nil, ErrCheckSignerFail
	}
	return newMsg, nil
}

func ConvertCancelFeeSplitMsg(data []byte, signers []sdk.AccAddress) (sdk.Msg, error) {
	newMsg := types.MsgCancelFeeSplit{}
	err := json.Unmarshal(data, &newMsg)
	if err != nil {
		return nil, err
	}
	err = newMsg.ValidateBasic()
	if err != nil {
		return nil, err
	}
	if ok := common.CheckSignerAddress(signers, newMsg.GetSigners()); !ok {
		return nil, ErrCheckSignerFail
	}
	return newMsg, nil
}
