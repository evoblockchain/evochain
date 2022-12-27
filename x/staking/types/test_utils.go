package types

import (
	"github.com/evoblockchain/evochain/libs/tendermint/crypto"
	"github.com/evoblockchain/evochain/libs/tendermint/crypto/ed25519"

	sdk "github.com/evoblockchain/evochain/libs/cosmos-sdk/types"
)

// nolint: deadcode unused
var (
	pk1   = ed25519.GenPrivKey().PubKey()
	pk2   = ed25519.GenPrivKey().PubKey()
	addr1 = pk1.Address()
	addr2 = pk2.Address()

	valAddr1 = sdk.ValAddress(addr1)
	valAddr2 = sdk.ValAddress(addr2)

	dlgAddr1 = sdk.AccAddress(addr1)
	dlgAddr2 = sdk.AccAddress(addr2)

	emptyAddr   sdk.ValAddress
	emptyPubkey crypto.PubKey
)
