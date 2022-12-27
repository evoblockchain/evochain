package types

import (
	"testing"

	"github.com/evoblockchain/evochain/x/common"

	sdk "github.com/evoblockchain/evochain/libs/cosmos-sdk/types"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/x/auth"
	"github.com/evoblockchain/evochain/libs/tendermint/crypto/secp256k1"
	"github.com/stretchr/testify/require"
)

func TestAmountToCoins(t *testing.T) {
	coinStr := "2btc,1" + common.NativeToken
	coins, err := sdk.ParseDecCoins(coinStr)
	require.Nil(t, err)
	expectedCoins := sdk.SysCoins{
		sdk.NewDecCoinFromDec("btc", sdk.NewDec(2)),
		sdk.NewDecCoinFromDec(common.NativeToken, sdk.NewDec(1)),
	}
	require.EqualValues(t, expectedCoins, coins)
}

func TestStrToTransfers(t *testing.T) {
	common.InitConfig()
	//coinStr := `[{"to": "cosmos18ragjd23yv4ctjg3vadh43q5zf8z0hafm4qjrf", "amount": "1BNB,2BTC"},
	//{"to": "cosmos18ragjd23yv4ctjg3vadh43q5zf8z0hafm4qjrf", "amount": "1OKB,2BTC"}]`
	coinStr := `[{"to":"evo1ss60jyagpw8v224006cqdkw22m32ytz54av3h3","amount":"1` + common.NativeToken + `"}]`
	coinStrError := `[{"to":"veo16zra6y26jytss520xrffq96p4fp4mnantcvyp2","amount":"1` + common.NativeToken + `"}]`
	addr, err := sdk.AccAddressFromBech32("evo1ss60jyagpw8v224006cqdkw22m32ytz54av3h3")
	require.Nil(t, err)
	_, err = StrToTransfers(coinStrError)
	require.Error(t, err)
	transfers, err := StrToTransfers(coinStr)
	require.Nil(t, err)
	transfer := []TransferUnit{
		{
			To: addr,
			Coins: []sdk.SysCoin{
				sdk.NewDecCoinFromDec(common.NativeToken, sdk.NewDec(1)),
			},
		},
	}
	require.EqualValues(t, transfer, transfers)

	coinStr = `[{"to":"evo16zra6y26jytss520xrffq96p4fp4mnantcvyp2",amount":"1"}]`
	_, err = StrToTransfers(coinStr)
	require.Error(t, err)
}

func TestMergeCoinInfo(t *testing.T) {

	//availableCoins, freezeCoins, lockCoins sdk.SysCoins
	availableCoins := sdk.SysCoins{
		sdk.NewDecCoinFromDec(common.NativeToken, sdk.NewDec(100)),
		sdk.NewDecCoinFromDec("bnb", sdk.NewDec(100)),
		sdk.NewDecCoinFromDec("btc", sdk.NewDec(100)),
	}

	lockedCoins := sdk.SysCoins{
		sdk.NewDecCoinFromDec("btc", sdk.NewDec(100)),
		sdk.NewDecCoinFromDec("abc", sdk.NewDec(100)),
	}

	coinsInfo := MergeCoinInfo(availableCoins, lockedCoins)
	expectedCoinsInfo := CoinsInfo{
		CoinInfo{"abc", "0", "100.000000000000000000"},
		CoinInfo{"bnb", "100.000000000000000000", "0"},
		CoinInfo{"btc", "100.000000000000000000", "100.000000000000000000"},
		CoinInfo{common.NativeToken, "100.000000000000000000", "0"},
	}
	require.EqualValues(t, expectedCoinsInfo, coinsInfo)
}

func TestDecAccount_String(t *testing.T) {
	priKey := secp256k1.GenPrivKey()
	pubKey := priKey.PubKey()
	addr := sdk.AccAddress(pubKey.Address())
	dec := sdk.MustNewDecFromStr("0.2")
	decCoins := sdk.SysCoins{
		sdk.NewDecCoinFromDec(common.NativeToken, dec),
	}
	decAccount := DecAccount{
		Address:       addr,
		Coins:         decCoins,
		PubKey:        pubKey,
		AccountNumber: 1,
		Sequence:      1,
	}

	expectedStr := `Account:
 Address:       ` + addr.String() + `
 Pubkey:        ` + sdk.MustBech32ifyAccPub(pubKey) + `
 Coins:         0.200000000000000000` + common.NativeToken + `
 AccountNumber: 1
 Sequence:      1`

	decStr := decAccount.String()
	require.EqualValues(t, decStr, expectedStr)
}

func TestBaseAccountToDecAccount(t *testing.T) {
	priKey := secp256k1.GenPrivKey()
	pubKey := priKey.PubKey()
	addr := sdk.AccAddress(pubKey.Address())

	coins := sdk.SysCoins{
		sdk.NewDecCoinFromDec(common.NativeToken, sdk.NewDec(100)),
	}

	baseAccount := auth.BaseAccount{
		Address:       addr,
		Coins:         coins,
		PubKey:        pubKey,
		AccountNumber: 1,
		Sequence:      1,
	}

	dec := sdk.MustNewDecFromStr("100.00000000")
	decCoins := sdk.SysCoins{
		sdk.NewDecCoinFromDec(common.NativeToken, dec),
	}

	expectedDecAccount := DecAccount{
		Address:       addr,
		Coins:         decCoins,
		PubKey:        pubKey,
		AccountNumber: 1,
		Sequence:      1,
	}

	decAccount := BaseAccountToDecAccount(baseAccount)
	require.EqualValues(t, decAccount, expectedDecAccount)
}

func TestValidCoinName(t *testing.T) {
	coinName := "abf.s0fa"
	valid := sdk.ValidateDenom(coinName)
	require.Error(t, valid)
}

func TestValidOriginalSymbol(t *testing.T) {
	name := "abc"
	require.True(t, ValidOriginalSymbol(name))
	name = notAllowedPrefix
	require.False(t, ValidOriginalSymbol(name))
	name = notAllowedPrefix + "abc"
	require.False(t, ValidOriginalSymbol(name))
	name = notAllowedPrefix + "-abc"
	require.False(t, ValidOriginalSymbol(name))
	name = notAllowedPrefix + "-abc-1af"
	require.False(t, ValidOriginalSymbol(name))
	name = notAllowedPrefix + "1"
	require.False(t, ValidOriginalSymbol(name))
	name = "abc-1fa"
	require.False(t, ValidOriginalSymbol(name))
}

func TestValidateDenom(t *testing.T) {
	name := "abc"
	require.Nil(t, sdk.ValidateDenom(name))

	name = "abc-123"
	require.Nil(t, sdk.ValidateDenom(name))

	name = "abc-ts3"
	require.NotNil(t, sdk.ValidateDenom(name))

	name = notAllowedPrefix + "_abc"
	require.NotNil(t, sdk.ValidateDenom(name))

	name = notAllowedPrefix + "_abc-1af"
	require.NotNil(t, sdk.ValidateDenom(name))

	name = notAllowedPrefix + "_abcde-aaa_abtc-e12"
	require.Nil(t, sdk.ValidateDenom(name))

	name = notAllowedPrefix + "_f-abcde-aaa"
	require.NotNil(t, sdk.ValidateDenom(name))

	name = "pool-abcde-aaa"
	require.NotNil(t, sdk.ValidateDenom(name))

	name = "token-abcde-aaa"
	require.NotNil(t, sdk.ValidateDenom(name))

	name = "pool-abc"
	require.Nil(t, sdk.ValidateDenom(name))

	name = notAllowedPrefix + "_abc"
	require.NotNil(t, sdk.ValidateDenom(name))

	name = notAllowedPrefix + "_abc" + "-bcd"
	require.NotNil(t, sdk.ValidateDenom(name))

	name = notAllowedPrefix + "_abc" + "_bcd-1234"
	require.NotNil(t, sdk.ValidateDenom(name))

	name = notAllowedPrefix + "_abc-123" + "_bcd"
	require.Nil(t, sdk.ValidateDenom(name))
}
