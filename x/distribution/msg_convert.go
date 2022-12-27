package distribution

import (
	"encoding/json"
	"errors"

	"github.com/evoblockchain/evochain/libs/cosmos-sdk/baseapp"
	sdk "github.com/evoblockchain/evochain/libs/cosmos-sdk/types"
	tmtypes "github.com/evoblockchain/evochain/libs/tendermint/types"
	"github.com/evoblockchain/evochain/x/common"
	"github.com/evoblockchain/evochain/x/distribution/types"
)

var (
	ErrCheckSignerFail = errors.New("check signer fail")
)

func init() {
	RegisterConvert()
}

func RegisterConvert() {
	enableHeight := tmtypes.GetVenus3Height()
	baseapp.RegisterCmHandle("evoblockchain/distribution/MsgWithdrawDelegatorAllRewards", baseapp.NewCMHandle(ConvertWithdrawDelegatorAllRewardsMsg, enableHeight))
}

func ConvertWithdrawDelegatorAllRewardsMsg(data []byte, signers []sdk.AccAddress) (sdk.Msg, error) {
	newMsg := types.MsgWithdrawDelegatorAllRewards{}
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
