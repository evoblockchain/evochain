package infura

import evm "github.com/evoblockchain/evochain/x/evm/watcher"

type EvmKeeper interface {
	SetObserverKeeper(keeper evm.InfuraKeeper)
}
