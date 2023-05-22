package appstatus

import (
	"fmt"
	"path/filepath"

	bam "github.com/evoblockchain/evochain/libs/cosmos-sdk/baseapp"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/client/flags"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/store/mpt"
	sdk "github.com/evoblockchain/evochain/libs/cosmos-sdk/types"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/x/auth"
	capabilitytypes "github.com/evoblockchain/evochain/libs/cosmos-sdk/x/capability/types"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/x/mint"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/x/params"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/x/supply"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/x/upgrade"
	"github.com/evoblockchain/evochain/libs/iavl"
	ibctransfertypes "github.com/evoblockchain/evochain/libs/ibc-go/modules/apps/transfer/types"
	ibchost "github.com/evoblockchain/evochain/libs/ibc-go/modules/core/24-host"
	dbm "github.com/evoblockchain/evochain/libs/tm-db"
	"github.com/evoblockchain/evochain/x/ammswap"
	dex "github.com/evoblockchain/evochain/x/dex/types"
	distr "github.com/evoblockchain/evochain/x/distribution"
	"github.com/evoblockchain/evochain/x/evidence"
	"github.com/evoblockchain/evochain/x/evm"
	"github.com/evoblockchain/evochain/x/farm"
	"github.com/evoblockchain/evochain/x/feesplit"
	"github.com/evoblockchain/evochain/x/gov"
	"github.com/evoblockchain/evochain/x/order"
	"github.com/evoblockchain/evochain/x/slashing"
	staking "github.com/evoblockchain/evochain/x/staking/types"
	token "github.com/evoblockchain/evochain/x/token/types"
	"github.com/evoblockchain/evochain/x/wasm"
	"github.com/spf13/viper"
)

const (
	applicationDB = "application"
	dbFolder      = "data"
)

func GetAllStoreKeys() []string {
	return []string{
		bam.MainStoreKey, auth.StoreKey, staking.StoreKey,
		supply.StoreKey, mint.StoreKey, distr.StoreKey, slashing.StoreKey,
		gov.StoreKey, params.StoreKey, upgrade.StoreKey, evidence.StoreKey,
		evm.StoreKey, token.StoreKey, token.KeyLock, dex.StoreKey, dex.TokenPairStoreKey,
		order.OrderStoreKey, ammswap.StoreKey, farm.StoreKey, ibctransfertypes.StoreKey, capabilitytypes.StoreKey,
		ibchost.StoreKey,
		//erc20.StoreKey,
		mpt.StoreKey,
		wasm.StoreKey,
		feesplit.StoreKey,
	}
}

func IsFastStorageStrategy() bool {
	return checkFastStorageStrategy(GetAllStoreKeys())
}

func checkFastStorageStrategy(storeKeys []string) bool {
	home := viper.GetString(flags.FlagHome)
	dataDir := filepath.Join(home, dbFolder)
	db, err := sdk.NewDB(applicationDB, dataDir)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	for _, v := range storeKeys {
		if !isFss(db, v) {
			return false
		}
	}

	return true
}

func isFss(db dbm.DB, storeKey string) bool {
	prefix := fmt.Sprintf("s/k:%s/", storeKey)
	prefixDB := dbm.NewPrefixDB(db, []byte(prefix))

	return iavl.IsFastStorageStrategy(prefixDB)
}
