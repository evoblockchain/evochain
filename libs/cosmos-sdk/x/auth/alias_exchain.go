package auth

import (
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/x/auth/exported"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/x/auth/keeper"
)

type (
	Account   = exported.Account
	ObserverI = keeper.ObserverI
)
