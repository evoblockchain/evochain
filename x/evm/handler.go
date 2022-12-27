package evm

import (
	bam "github.com/evoblockchain/evochain/libs/cosmos-sdk/baseapp"
	sdk "github.com/evoblockchain/evochain/libs/cosmos-sdk/types"
	sdkerrors "github.com/evoblockchain/evochain/libs/cosmos-sdk/types/errors"
	cfg "github.com/evoblockchain/evochain/libs/tendermint/config"
	common2 "github.com/evoblockchain/evochain/x/common"
	"github.com/evoblockchain/evochain/x/evm/txs"
	"github.com/evoblockchain/evochain/x/evm/txs/base"
	"github.com/evoblockchain/evochain/x/evm/types"
	"github.com/evoblockchain/evochain/x/evm/watcher"
)

// NewHandler returns a handler for Ethermint type messages.
func NewHandler(k *Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (result *sdk.Result, err error) {
		ctx.SetEventManager(sdk.NewEventManager())

		if ctx.IsDeliver() {
			k.EvmStateDb.WithContext(ctx).MarkUpdatedAcc(k.UpdatedAccount)
			k.UpdatedAccount = k.UpdatedAccount[:0]
		}

		evmTx, ok := msg.(*types.MsgEthereumTx)
		if ok {
			if watcher.IsWatcherEnabled() {
				ctx.SetWatcher(watcher.NewTxWatcher())
			}
			result, err = handleMsgEthereumTx(ctx, k, evmTx)
			if err != nil {
				err = sdkerrors.New(types.ModuleName, types.CodeSpaceEvmCallFailed, err.Error())
			}
		} else {
			err = sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", ModuleName, msg)
		}

		return result, err
	}
}

func updateHGU(ctx sdk.Context, msg sdk.Msg) {
	if cfg.DynamicConfig.GetMaxGasUsedPerBlock() <= 0 {
		return
	}

	db := bam.InstanceOfHistoryGasUsedRecordDB()
	msgFnSignature, toDeployContractSize := getMsgCallFnSignature(msg)

	if msgFnSignature == nil {
		return
	}

	hisGu, err := db.Get(msgFnSignature)
	if err != nil {
		return
	}

	gc := int64(ctx.GasMeter().GasConsumed())
	if toDeployContractSize > 0 {
		// calculate average gas consume for deploy contract case
		gc = gc / int64(toDeployContractSize)
	}

	var avgGas int64
	if hisGu != nil {
		hgu := common2.BytesToInt64(hisGu)
		avgGas = int64(bam.GasUsedFactor*float64(gc) + (1.0-bam.GasUsedFactor)*float64(hgu))
	} else {
		avgGas = gc
	}

	err = db.Set(msgFnSignature, common2.Int64ToBytes(avgGas))
	if err != nil {
		return
	}
}

func getMsgCallFnSignature(msg sdk.Msg) ([]byte, int) {
	switch msg := msg.(type) {
	case *types.MsgEthereumTx:
		return msg.GetTxFnSignatureInfo()
	default:
		return nil, 0
	}
}

// handleMsgEthereumTx handles an Ethereum specific tx
// 1. txs can be divided into TraceTxLog/CheckTx/DeliverTx
func handleMsgEthereumTx(ctx sdk.Context, k *Keeper, msg *types.MsgEthereumTx) (*sdk.Result, error) {
	txFactory := txs.NewFactory(base.Config{
		Ctx:    ctx,
		Keeper: k,
	})
	tx, err := txFactory.CreateTx()
	if err != nil {
		return nil, err
	}
	defer tx.Dispose()

	// core logical to handle ethereum tx
	rst, err := txs.TransitionEvmTx(tx, msg)
	if err == nil && !ctx.IsCheckTx() {
		updateHGU(ctx, msg)
	}

	return rst, err
}
