package store

import (
	dbm "github.com/evoblockchain/evochain/libs/tm-db"

	"github.com/evoblockchain/evochain/libs/cosmos-sdk/store/cache"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/store/rootmulti"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/store/types"
)

func NewCommitMultiStore(db dbm.DB) types.CommitMultiStore {
	return rootmulti.NewStore(db)
}

func NewCommitKVStoreCacheManager() types.MultiStorePersistentCache {
	return cache.NewCommitKVStoreCacheManager(cache.DefaultCommitKVStoreCacheSize)
}
