package types

import (
	sdkerrors "github.com/evoblockchain/evochain/libs/cosmos-sdk/types/errors"
	host "github.com/evoblockchain/evochain/libs/ibc-go/modules/core/24-host"
)

var ErrIbcDisabled = sdkerrors.Register(host.ModuleName, 1, "IBC are disabled")
