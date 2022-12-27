package app

import (
	"runtime"
	"time"

	appconfig "github.com/evoblockchain/evochain/app/config"
	"github.com/evoblockchain/evochain/app/types"
	sdk "github.com/evoblockchain/evochain/libs/cosmos-sdk/types"
	"github.com/evoblockchain/evochain/libs/system/trace"
	abci "github.com/evoblockchain/evochain/libs/tendermint/abci/types"
	"github.com/evoblockchain/evochain/x/evm"
	"github.com/evoblockchain/evochain/x/wasm/watcher"
)

// BeginBlock implements the Application interface
func (app *EVOChainApp) BeginBlock(req abci.RequestBeginBlock) (res abci.ResponseBeginBlock) {
	trace.OnAppBeginBlockEnter(app.LastBlockHeight() + 1)
	app.EvmKeeper.Watcher.DelayEraseKey()
	return app.BaseApp.BeginBlock(req)
}

func (app *EVOChainApp) DeliverTx(req abci.RequestDeliverTx) (res abci.ResponseDeliverTx) {

	trace.OnAppDeliverTxEnter()

	resp := app.BaseApp.DeliverTx(req)

	if appconfig.GetOecConfig().GetDynamicGpMode() != types.CloseMode {
		tx, err := evm.TxDecoder(app.marshal)(req.Tx)
		if err == nil {
			//optimize get tx gas price can not get value from verifySign method
			app.gpo.CurrentBlockGPs.Update(tx.GetGasPrice(), uint64(resp.GasUsed))
		}
	}

	return resp
}

func (app *EVOChainApp) PreDeliverRealTx(req []byte) (res abci.TxEssentials) {
	return app.BaseApp.PreDeliverRealTx(req)
}

func (app *EVOChainApp) DeliverRealTx(req abci.TxEssentials) (res abci.ResponseDeliverTx) {
	trace.OnAppDeliverTxEnter()
	resp := app.BaseApp.DeliverRealTx(req)
	app.EvmKeeper.Watcher.RecordTxAndFailedReceipt(req, &resp, app.GetTxDecoder())

	var err error
	if appconfig.GetOecConfig().GetDynamicGpMode() != types.CloseMode {
		tx, _ := req.(sdk.Tx)
		if tx == nil {
			tx, err = evm.TxDecoder(app.Codec())(req.GetRaw())
		}
		if err == nil {
			//optimize get tx gas price can not get value from verifySign method
			app.gpo.CurrentBlockGPs.Update(tx.GetGasPrice(), uint64(resp.GasUsed))
		}
	}

	return resp
}

// EndBlock implements the Application interface
func (app *EVOChainApp) EndBlock(req abci.RequestEndBlock) (res abci.ResponseEndBlock) {
	return app.BaseApp.EndBlock(req)
}

// Commit implements the Application interface
func (app *EVOChainApp) Commit(req abci.RequestCommit) abci.ResponseCommit {
	if gcInterval := appconfig.GetOecConfig().GetGcInterval(); gcInterval > 0 {
		if (app.BaseApp.LastBlockHeight()+1)%int64(gcInterval) == 0 {
			startTime := time.Now()
			runtime.GC()
			elapsed := time.Now().Sub(startTime).Milliseconds()
			app.Logger().Info("force gc for debug", "height", app.BaseApp.LastBlockHeight()+1,
				"elapsed(ms)", elapsed)
		}
	}
	//defer trace.GetTraceSummary().Dump()
	defer trace.OnCommitDone()

	tasks := app.heightTasks[app.BaseApp.LastBlockHeight()+1]
	if tasks != nil {
		ctx := app.BaseApp.GetDeliverStateCtx()
		for _, t := range *tasks {
			if err := t.Execute(ctx); nil != err {
				panic("bad things")
			}
		}
	}
	res := app.BaseApp.Commit(req)

	// we call watch#Commit here ,because
	// 1. this round commit a valid block
	// 2. before commit the block,State#updateToState hasent not called yet,so the proposalBlockPart is not nil which means we wont
	// 	  call the prerun during commit step(edge case)
	app.EvmKeeper.Watcher.Commit()
	watcher.Commit()

	return res
}
