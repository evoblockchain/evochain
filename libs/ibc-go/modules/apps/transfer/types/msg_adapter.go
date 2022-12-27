package types

import (
	sdk "github.com/evoblockchain/evochain/libs/cosmos-sdk/types"
	sdkerrors "github.com/evoblockchain/evochain/libs/cosmos-sdk/types/errors"
)

//for denom convert wei to evo and reject evo direct
func (m *MsgTransfer) RulesFilter() (sdk.Msg, error) {
	if m.Token.Denom == sdk.DefaultBondDenom {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "ibc MsgTransfer not support evo denom")
	}
	return m, nil
}
