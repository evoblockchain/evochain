package baseapp

import (
	"bytes"
	"runtime"
	"sync"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/store/types"
	sdk "github.com/evoblockchain/evochain/libs/cosmos-sdk/types"
	sdkerrors "github.com/evoblockchain/evochain/libs/cosmos-sdk/types/errors"
	abci "github.com/evoblockchain/evochain/libs/tendermint/abci/types"
	sm "github.com/evoblockchain/evochain/libs/tendermint/state"
	"github.com/spf13/viper"
)

var (
	maxTxNumberInParallelChan   = 20000
	whiteAcc                    = string(hexutil.MustDecode("0x01f1829676db577682e944fc3493d451b67ff3e29f")) //fee
	maxGoroutineNumberInParaTx  = runtime.NumCPU()
	multiCacheListClearInterval = int64(100)
)

type extraDataForTx struct {
	fee       sdk.Coins
	isEvm     bool
	from      string
	to        string
	stdTx     sdk.Tx
	decodeErr error
}

// getExtraDataByTxs preprocessing tx : verify tx, get sender, get toAddress, get txFee
func (app *BaseApp) getExtraDataByTxs(txs [][]byte) {
	para := app.parallelTxManage

	var wg sync.WaitGroup
	for index, txBytes := range txs {
		wg.Add(1)
		go func(index int, txBytes []byte) {
			defer wg.Done()

			var tx sdk.Tx
			var err error

			if mem := GetGlobalMempool(); mem != nil {
				tx, _ = mem.ReapEssentialTx(txBytes).(sdk.Tx)
			}
			if tx == nil {
				tx, err = app.txDecoder(txBytes)
				if err != nil {
					para.extraTxsInfo[index] = &extraDataForTx{
						decodeErr: err,
					}
					return
				}
			}
			if tx != nil {
				app.blockDataCache.SetTx(txBytes, tx)
			}

			coin, isEvm, s, toAddr, _ := app.getTxFeeAndFromHandler(app.getContextForTx(runTxModeDeliver, txBytes), tx)
			para.extraTxsInfo[index] = &extraDataForTx{
				fee:   coin,
				isEvm: isEvm,
				from:  s,
				to:    toAddr,
				stdTx: tx,
			}
		}(index, txBytes)
	}
	wg.Wait()
}

var (
	rootAddr = make(map[string]string, 0)
)

// Find father node
func Find(x string) string {
	if rootAddr[x] != x {
		rootAddr[x] = Find(rootAddr[x])
	}
	return rootAddr[x]
}

// Union from and to
func Union(x string, yString string) {
	if _, ok := rootAddr[x]; !ok {
		rootAddr[x] = x
	}
	if yString == "" {
		return
	}
	if _, ok := rootAddr[yString]; !ok {
		rootAddr[yString] = yString
	}
	fx := Find(x)
	fy := Find(yString)
	if fx != fy {
		rootAddr[fy] = fx
	}
}

// calGroup cal group by txs
func (app *BaseApp) calGroup() {

	para := app.parallelTxManage

	rootAddr = make(map[string]string, 0)
	for index, tx := range para.extraTxsInfo {
		if tx.isEvm { //evmTx
			Union(tx.from, tx.to)
		} else {
			para.haveCosmosTxInBlock = true
			app.parallelTxManage.txResultCollector.putResult(index, &executeResult{paraMsg: &sdk.ParaMsg{}})
		}
	}

	addrToID := make(map[string]int, 0)

	for index, txInfo := range para.extraTxsInfo {
		if !txInfo.isEvm {
			continue
		}
		rootAddr := Find(txInfo.from)
		id, exist := addrToID[rootAddr]
		if !exist {
			id = len(para.groupList)
			addrToID[rootAddr] = id

		}
		para.groupList[id] = append(para.groupList[id], index)
	}

	groupSize := len(para.groupList)
	for groupIndex := 0; groupIndex < groupSize; groupIndex++ {
		list := para.groupList[groupIndex]
		for index := 0; index < len(list); index++ {
			if index+1 <= len(list)-1 {
				app.parallelTxManage.nextTxInGroup[list[index]] = list[index+1]
			}
			if index-1 >= 0 {
				app.parallelTxManage.preTxInGroup[list[index]] = list[index-1]
			}
		}
	}
}

// ParallelTxs run txs
func (app *BaseApp) ParallelTxs(txs [][]byte, onlyCalSender bool) []*abci.ResponseDeliverTx {
	pm := app.parallelTxManage

	txSize := len(txs)
	pm.txSize = txSize
	pm.haveCosmosTxInBlock = false
	pm.workgroup.txs = txs
	pm.isAsyncDeliverTx = true
	pm.cms = app.deliverState.ms.CacheMultiStore()
	pm.cms.DisableCacheReadList()
	app.deliverState.ms.DisableCacheReadList()
	pm.blockHeight = app.deliverState.ctx.BlockHeight()

	if txSize == 0 {
		return make([]*abci.ResponseDeliverTx, 0)
	}
	pm.init()

	app.getExtraDataByTxs(txs)

	app.calGroup()

	return app.runTxs()
}

func (app *BaseApp) fixFeeCollector() {
	ctx, _ := app.cacheTxContext(app.getContextForTx(runTxModeDeliver, []byte{}), []byte{})

	ctx.SetMultiStore(app.parallelTxManage.cms)
	// The feesplit is only processed at the endblock
	if err := app.updateFeeCollectorAccHandler(ctx, app.parallelTxManage.currTxFee, nil); err != nil {
		panic(err)
	}
}

// fixFeeCollector update fee account

func (app *BaseApp) runTxs() []*abci.ResponseDeliverTx {
	maxGas := app.getMaximumBlockGas()
	currentGas := uint64(0)
	overFlow := func(sumGas uint64, currGas int64, maxGas uint64) bool {
		if maxGas <= 0 {
			return false
		}
		if sumGas+uint64(currGas) >= maxGas { // TODO : fix later
			return true
		}
		return false
	}
	signal := make(chan int, 1)
	rerunIdx := 0
	txIndex := 0

	pm := app.parallelTxManage
	pm.workgroup.isReady = true
	app.parallelTxManage.workgroup.Start()

	deliverTxs := make([]*abci.ResponseDeliverTx, pm.txSize)

	asyncCb := func(execRes *executeResult) {

		if !pm.workgroup.isReady { // runTxs end
			return
		}
		if execRes.blockHeight != app.deliverState.ctx.BlockHeight() { // not excepted resp
			return
		}
		receiveTxIndex := int(execRes.counter)
		pm.workgroup.setTxStatus(receiveTxIndex, false)

		//skip old txIndex
		if receiveTxIndex < txIndex {
			return
		}
		pm.txResultCollector.putResult(receiveTxIndex, execRes)

		if pm.workgroup.isFailed(pm.workgroup.runningStats(receiveTxIndex)) {
			pm.txResultCollector.putResult(receiveTxIndex, nil)
			// reRun already failed tx
			pm.workgroup.AddTask(receiveTxIndex)
		} else {
			if nextTx, ok := pm.nextTxInGroup[receiveTxIndex]; ok {
				if !pm.workgroup.isRunning(nextTx) {
					pm.txResultCollector.putResult(nextTx, nil)
					// run next tx in this group
					pm.workgroup.AddTask(nextTx)
				}
			}
		}

		// not excepted tx
		if txIndex != receiveTxIndex {
			return
		}

		for true {
			res := pm.txResultCollector.getTxResult(txIndex)
			if res == nil {
				break
			}
			if pm.newIsConflict(res) || overFlow(currentGas, res.resp.GasUsed, maxGas) {
				rerunIdx++

				// conflict rerun tx
				if !pm.extraTxsInfo[txIndex].isEvm {
					app.fixFeeCollector()
				}
				res = app.deliverTxWithCache(txIndex)
				pm.txResultCollector.putResult(txIndex, res)

				if nextTx, ok := app.parallelTxManage.nextTxInGroup[txIndex]; ok {
					if !pm.workgroup.isRunning(nextTx) {
						pm.txResultCollector.putResult(nextTx, nil)
						pm.workgroup.AddTask(nextTx)
					}
				}

			}
			if pm.txResultCollector.getTxResult(txIndex).paraMsg.AnteErr != nil {
				res.ms = nil
			}

			txRs := res.resp
			deliverTxs[txIndex] = &txRs

			pm.blockGasMeterMu.Lock()
			// Note : don't take care of the case of ErrorGasOverflow
			app.deliverState.ctx.BlockGasMeter().ConsumeGas(sdk.Gas(res.resp.GasUsed), "unexpected error")
			pm.blockGasMeterMu.Unlock()
			// merge tx
			pm.SetCurrentIndex(txIndex, res)
			pm.finalResult[txIndex] = res

			currentGas += uint64(res.resp.GasUsed)
			txIndex++
			if txIndex == pm.txSize {
				app.logger.Info("Paralleled-tx", "blockHeight", app.deliverState.ctx.BlockHeight(), "len(txs)", pm.txSize,
					"Parallel run", pm.txSize-rerunIdx, "ReRun", rerunIdx, "len(group)", len(pm.groupList))
				signal <- 0
				return
			}
			if pm.txResultCollector.getTxResult(txIndex) == nil && !pm.workgroup.isRunning(txIndex) {
				pm.workgroup.AddTask(txIndex)
			}
		}
	}

	pm.workgroup.resultCb = asyncCb
	pm.workgroup.taskRun = app.asyncDeliverTx

	if len(pm.groupList) == 0 {
		pm.workgroup.AddTask(0)
	} else if pm.groupList[0][0] != 0 {
		pm.workgroup.AddTask(0)
	}

	for _, group := range pm.groupList {
		pm.workgroup.AddTask(group[0])
	}

	//waiting for call back
	<-signal

	// fix logs
	app.feeChanged = true
	app.feeCollector = app.parallelTxManage.currTxFee
	receiptsLogs := app.endParallelTxs()
	for index, v := range receiptsLogs {
		if len(v) != 0 { // only update evm tx result
			deliverTxs[index].Data = v
		}
	}

	pm.cms.Write()
	return deliverTxs
}

func (app *BaseApp) endParallelTxs() [][]byte {

	// handle receipt's logs
	logIndex := make([]int, app.parallelTxManage.txSize)
	errs := make([]error, app.parallelTxManage.txSize)
	hasEnterEvmTx := make([]bool, app.parallelTxManage.txSize)
	resp := make([]abci.ResponseDeliverTx, app.parallelTxManage.txSize)
	watchers := make([]sdk.IWatcher, app.parallelTxManage.txSize)
	txs := make([]sdk.Tx, app.parallelTxManage.txSize)
	app.FeeSplitCollector = make([]*sdk.FeeSplitInfo, 0)
	for index := 0; index < app.parallelTxManage.txSize; index++ {
		txRes := app.parallelTxManage.finalResult[index]
		logIndex[index] = txRes.paraMsg.LogIndex
		errs[index] = txRes.paraMsg.AnteErr
		hasEnterEvmTx[index] = txRes.paraMsg.HasRunEvmTx
		resp[index] = txRes.resp
		watchers[index] = txRes.watcher
		txs[index] = app.parallelTxManage.extraTxsInfo[index].stdTx
		if txRes.FeeSpiltInfo.HasFee {
			app.FeeSplitCollector = append(app.FeeSplitCollector, txRes.FeeSpiltInfo)
		}

	}
	app.watcherCollector(watchers...)
	app.parallelTxManage.clear()
	return app.logFix(txs, logIndex, hasEnterEvmTx, errs, resp)
}

//we reuse the nonce that changed by the last async call
//if last ante handler has been failed, we need rerun it ? or not?
func (app *BaseApp) deliverTxWithCache(txIndex int) *executeResult {
	app.parallelTxManage.workgroup.setTxStatus(txIndex, true)
	txStatus := app.parallelTxManage.extraTxsInfo[txIndex]

	if txStatus.stdTx == nil {
		asyncExe := newExecuteResult(sdkerrors.ResponseDeliverTx(txStatus.decodeErr,
			0, 0, app.trace), nil, uint32(txIndex), nil, 0, sdk.EmptyWatcher{}, nil, nil)
		return asyncExe
	}
	var (
		resp abci.ResponseDeliverTx
		mode runTxMode
	)
	mode = runTxModeDeliverInAsync
	info, errM := app.runTxWithIndex(txIndex, mode, app.parallelTxManage.workgroup.txs[txIndex], txStatus.stdTx, LatestSimulateTxHeight)
	if errM != nil {
		resp = sdkerrors.ResponseDeliverTx(errM, info.gInfo.GasWanted, info.gInfo.GasUsed, app.trace)
	} else {
		resp = abci.ResponseDeliverTx{
			GasWanted: int64(info.gInfo.GasWanted), // TODO: Should type accept unsigned ints?
			GasUsed:   int64(info.gInfo.GasUsed),   // TODO: Should type accept unsigned ints?
			Log:       info.result.Log,
			Data:      info.result.Data,
			Events:    info.result.Events.ToABCIEvents(),
		}
	}

	asyncExe := newExecuteResult(resp, info.msCacheAnte, uint32(txIndex), info.ctx.ParaMsg(),
		0, info.runMsgCtx.GetWatcher(), info.tx.GetMsgs(), info.ctx.GetFeeSplitInfo())
	app.parallelTxManage.addMultiCache(info.msCacheAnte, info.msCache)
	return asyncExe
}

type executeResult struct {
	resp         abci.ResponseDeliverTx
	ms           sdk.CacheMultiStore
	counter      uint32
	paraMsg      *sdk.ParaMsg
	blockHeight  int64
	watcher      sdk.IWatcher
	msgs         []sdk.Msg
	FeeSpiltInfo *sdk.FeeSplitInfo
}

func newExecuteResult(r abci.ResponseDeliverTx, ms sdk.CacheMultiStore, counter uint32,
	paraMsg *sdk.ParaMsg, height int64, watcher sdk.IWatcher, msgs []sdk.Msg, feeSpiltInfo *sdk.FeeSplitInfo) *executeResult {

	if feeSpiltInfo == nil {
		feeSpiltInfo = &sdk.FeeSplitInfo{}
	}
	ans := &executeResult{
		resp:         r,
		ms:           ms,
		counter:      counter,
		paraMsg:      paraMsg,
		blockHeight:  height,
		watcher:      watcher,
		msgs:         msgs,
		FeeSpiltInfo: feeSpiltInfo,
	}

	if paraMsg == nil {
		ans.paraMsg = &sdk.ParaMsg{}
	}
	return ans
}

type asyncWorkGroup struct {
	txs     [][]byte
	isReady bool

	runningStatus map[int]int
	isrunning     map[int]bool

	markFailedStats map[int]bool

	indexInAll int
	runningMu  sync.RWMutex

	resultCh chan *executeResult
	resultCb func(*executeResult)

	taskCh  chan int
	taskRun func(int)

	stopChan chan struct{}
}

func newAsyncWorkGroup() *asyncWorkGroup {
	return &asyncWorkGroup{
		runningStatus:   make(map[int]int, 0),
		isrunning:       make(map[int]bool, 0),
		markFailedStats: make(map[int]bool),

		resultCh: make(chan *executeResult, maxTxNumberInParallelChan),
		resultCb: nil,

		taskCh:  make(chan int, maxTxNumberInParallelChan),
		taskRun: nil,

		stopChan: make(chan struct{}),
	}
}

func (a *asyncWorkGroup) markFailed(txIndexAll int) {
	a.runningMu.Lock()
	defer a.runningMu.Unlock()
	a.markFailedStats[txIndexAll] = true
}

func (a *asyncWorkGroup) isFailed(txIndexAll int) bool {
	a.runningMu.Lock()
	defer a.runningMu.Unlock()
	return a.markFailedStats[txIndexAll]
}

func (a *asyncWorkGroup) setTxStatus(txIndex int, status bool) {
	a.runningMu.Lock()
	defer a.runningMu.Unlock()
	if status == true {
		a.runningStatus[txIndex] = a.indexInAll
		a.indexInAll++
	}
	a.isrunning[txIndex] = status
}

func (a *asyncWorkGroup) runningStats(txIndex int) int {
	a.runningMu.RLock()
	defer a.runningMu.RUnlock()
	return a.runningStatus[txIndex]
}

func (a *asyncWorkGroup) isRunning(txIndex int) bool {
	a.runningMu.RLock()
	defer a.runningMu.RUnlock()
	return a.isrunning[txIndex]
}

func (a *asyncWorkGroup) Push(item *executeResult) {
	a.resultCh <- item
}

func (a *asyncWorkGroup) AddTask(index int) {
	a.setTxStatus(index, true)
	a.taskCh <- index
}

func (a *asyncWorkGroup) Close() {
	for index := 0; index <= maxGoroutineNumberInParaTx; index++ {
		a.stopChan <- struct{}{}
	}
}

func (a *asyncWorkGroup) Start() {
	for index := 0; index < maxGoroutineNumberInParaTx; index++ {
		go func() {
			for true {
				select {
				case task := <-a.taskCh:
					a.taskRun(task)
				case <-a.stopChan:
					return
				}
			}
		}()

	}

	go func() {
		for {
			select {
			case exec := <-a.resultCh:
				a.resultCb(exec)
			case <-a.stopChan:
				return
			}
		}
	}()
}

type valueWithIndex struct {
	value   []byte
	txIndex int
}

type conflictCheck struct {
	items map[string]valueWithIndex
}

func newConflictCheck() *conflictCheck {
	return &conflictCheck{
		items: make(map[string]valueWithIndex),
	}
}

func (c *conflictCheck) update(key string, value []byte, txIndex int) {
	c.items[key] = valueWithIndex{
		value:   value,
		txIndex: txIndex,
	}
}

func (c *conflictCheck) deleteFeeAccount() {
	delete(c.items, whiteAcc)
}

func (c *conflictCheck) clear() {
	for key := range c.items {
		delete(c.items, key)
	}
}

type txResultCollector struct {
	mu     sync.RWMutex
	txReps []*executeResult
}

func newExecResult() *txResultCollector {
	return &txResultCollector{
		mu:     sync.RWMutex{},
		txReps: make([]*executeResult, 0),
	}
}

func (e *txResultCollector) clear() {
	e.mu.Lock()
	e.txReps = nil
	e.mu.Unlock()
}

func (e *txResultCollector) init(txSize int) {
	txRepsCap := cap(e.txReps)
	if e.txReps == nil || txRepsCap < txSize {
		e.txReps = make([]*executeResult, txSize)
	} else if txRepsCap >= txSize {
		e.txReps = e.txReps[0:txSize:txRepsCap]
		// https://github.com/golang/go/issues/5373
		for i := range e.txReps {
			e.txReps[i] = nil
		}
	}
}

func (e *txResultCollector) putResult(index int, txResult *executeResult) {
	e.mu.Lock()
	e.txReps[index] = txResult
	e.mu.Unlock()
}

func (e *txResultCollector) getTxResult(index int) *executeResult {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.txReps[index]
}

type parallelTxManager struct {
	blockGasMeterMu     sync.Mutex
	haveCosmosTxInBlock bool
	isAsyncDeliverTx    bool
	workgroup           *asyncWorkGroup

	extraTxsInfo      []*extraDataForTx
	txResultCollector *txResultCollector
	finalResult       []*executeResult

	groupList     map[int][]int
	nextTxInGroup map[int]int
	preTxInGroup  map[int]int

	mu          sync.RWMutex
	cms         sdk.CacheMultiStore
	blockHeight int64

	txSize    int
	cc        *conflictCheck
	currIndex int
	currTxFee sdk.Coins

	blockMultiStores *cacheMultiStoreList
	chainMultiStores *cacheMultiStoreList
}

func newParallelTxManager() *parallelTxManager {
	isAsync := sm.DeliverTxsExecMode(viper.GetInt(sm.FlagDeliverTxsExecMode)) == sm.DeliverTxsExecModeParallel
	return &parallelTxManager{
		blockGasMeterMu:  sync.Mutex{},
		isAsyncDeliverTx: isAsync,
		workgroup:        newAsyncWorkGroup(),

		groupList:     make(map[int][]int),
		nextTxInGroup: make(map[int]int),
		preTxInGroup:  make(map[int]int),

		txResultCollector: newExecResult(),
		cc:                newConflictCheck(),
		currIndex:         -1,
		currTxFee:         sdk.Coins{},

		blockMultiStores: newCacheMultiStoreList(),
		chainMultiStores: newCacheMultiStoreList(),
	}
}

func (f *parallelTxManager) addMultiCache(msAnte types.CacheMultiStore, msCache types.CacheMultiStore) {
	if msAnte != nil {
		f.blockMultiStores.PushStore(msAnte)
	}

	if msCache != nil {
		f.blockMultiStores.PushStore(msCache)
	}
}

func shouldCleanChainCache(height int64) bool {
	return height%multiCacheListClearInterval == 0
}

func (f *parallelTxManager) addBlockCacheToChainCache() {

	if shouldCleanChainCache(f.blockHeight) {
		f.chainMultiStores.Clear()
	} else {
		jobChan := make(chan types.CacheMultiStore, f.blockMultiStores.Len())
		for index := 0; index < maxGoroutineNumberInParaTx; index++ {
			go func(ch chan types.CacheMultiStore) {
				for j := range ch {
					j.Clear()
					f.chainMultiStores.PushStore(j)
				}
			}(jobChan)
		}

		f.blockMultiStores.Range(func(c types.CacheMultiStore) {
			jobChan <- c
		})
		close(jobChan)
	}

	f.blockMultiStores.Clear()
}

func (f *parallelTxManager) newIsConflict(e *executeResult) bool {
	if e.ms == nil {
		return true //TODO fix later
	}
	conflict := false

	e.ms.IteratorCache(false, func(key string, value []byte, isDirty bool, isDelete bool, storeKey types.StoreKey) bool {
		if data, ok := f.cc.items[storeKey.Name()+key]; ok {
			if !bytes.Equal(data.value, value) {
				conflict = true
				return false
			}
		}
		return true
	}, nil)

	return conflict
}

func (f *parallelTxManager) clear() {
	f.addBlockCacheToChainCache()
	f.workgroup.Close()
	f.workgroup.isReady = false
	f.workgroup.indexInAll = 0

	for key := range f.workgroup.markFailedStats {
		delete(f.workgroup.markFailedStats, key)
	}

	f.extraTxsInfo = nil
	f.txResultCollector.clear()

	for key := range f.groupList {
		delete(f.groupList, key)
	}
	for key := range f.nextTxInGroup {
		delete(f.nextTxInGroup, key)
	}
	for key := range f.preTxInGroup {
		delete(f.preTxInGroup, key)
	}

	f.cc.clear()
	f.currIndex = -1
	f.currTxFee = sdk.Coins{}
}

func (f *parallelTxManager) init() {
	txSize := f.txSize

	f.txResultCollector.init(txSize)
	f.finalResult = make([]*executeResult, txSize)

	txsInfoCap := cap(f.extraTxsInfo)
	if f.extraTxsInfo == nil || txsInfoCap < txSize {
		f.extraTxsInfo = make([]*extraDataForTx, txSize)
	} else if txsInfoCap >= txSize {
		f.extraTxsInfo = f.extraTxsInfo[0:txSize:txsInfoCap]
		for i := range f.extraTxsInfo {
			f.extraTxsInfo[i] = nil
		}
	}

	for key := range f.workgroup.runningStatus {
		delete(f.workgroup.runningStatus, key)
	}
	for key := range f.workgroup.isrunning {
		delete(f.workgroup.isrunning, key)
	}
}

func (f *parallelTxManager) getTxResult(index int) sdk.CacheMultiStore {
	f.mu.Lock()
	defer f.mu.Unlock()

	if index <= f.currIndex {
		return nil
	}

	var ms types.CacheMultiStore
	preIndexInGroup, ok := f.preTxInGroup[index]
	if ok && preIndexInGroup > f.currIndex {
		// get parent tx ms
		preResp := f.txResultCollector.getTxResult(preIndexInGroup)

		if preResp != nil && preResp.paraMsg.AnteErr == nil {
			if preResp.ms == nil {
				return nil
			}

			preResp.ms.DisableCacheReadList()
			ms = f.chainMultiStores.GetStoreWithParent(preResp.ms)
		}
	}

	if ms == nil {
		ms = f.chainMultiStores.GetStoreWithParent(f.cms)
	}

	if next, ok := f.nextTxInGroup[index]; ok {
		if f.workgroup.isRunning(next) {
			// mark failed if running
			f.workgroup.markFailed(f.workgroup.runningStats(next))
		} else {
			f.txResultCollector.putResult(next, nil)
		}
	}

	return ms
}

func (f *parallelTxManager) SetCurrentIndex(txIndex int, res *executeResult) {
	if res.ms == nil {
		return
	}

	f.mu.Lock()
	res.ms.IteratorCache(true, func(key string, value []byte, isDirty bool, isdelete bool, storeKey sdk.StoreKey) bool {
		f.cc.update(storeKey.Name()+key, value, txIndex)
		if isdelete {
			f.cms.GetKVStore(storeKey).Delete([]byte(key))
		} else if value != nil {
			f.cms.GetKVStore(storeKey).Set([]byte(key), value)
		}

		return true
	}, nil)
	f.currIndex = txIndex
	f.mu.Unlock()
	//f.cc.deleteFeeAccount()

	if res.paraMsg.AnteErr != nil {
		return
	}
	f.currTxFee = f.currTxFee.Add(f.extraTxsInfo[txIndex].fee.Sub(f.txResultCollector.getTxResult(txIndex).paraMsg.RefundFee)...)
}
