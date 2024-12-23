// Copyright 2024 The Erigon Authors
// This file is part of Erigon.
//
// Erigon is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Erigon is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with Erigon. If not, see <http://www.gnu.org/licenses/>.

package exec3

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/erigontech/erigon-lib/log/v3"

	"github.com/erigontech/erigon-lib/common/datadir"
	"github.com/erigontech/erigon/eth/consensuschain"

	"github.com/erigontech/erigon-lib/chain"
	"github.com/erigontech/erigon-lib/kv"

	"github.com/erigontech/erigon/consensus"
	"github.com/erigontech/erigon/core"
	"github.com/erigontech/erigon/core/exec"
	"github.com/erigontech/erigon/core/state"
	"github.com/erigontech/erigon/core/types"
	"github.com/erigontech/erigon/core/vm"
	"github.com/erigontech/erigon/core/vm/evmtypes"
	"github.com/erigontech/erigon/turbo/services"
	"github.com/erigontech/erigon/turbo/shards"
)

var noop = state.NewNoopWriter()

type Worker struct {
	lock        sync.Locker
	logger      log.Logger
	chainDb     kv.RoDB
	chainTx     kv.Tx
	background  bool // if true - worker does manage RoTx (begin/rollback) in .ResetTx()
	blockReader services.FullBlockReader
	in          *exec.QueueWithRetry
	rs          *state.StateV3Buffered
	stateWriter state.StateWriter
	stateReader state.ResettableStateReader
	historyMode bool // if true - stateReader is HistoryReaderV3, otherwise it's state reader
	chainConfig *chain.Config

	ctx      context.Context
	engine   consensus.Engine
	genesis  *types.Genesis
	resultCh *exec.ResultsQueue
	chain    consensus.ChainReader

	callTracer  *CallTracer
	taskGasPool *core.GasPool

	evm   *vm.EVM
	ibs   *state.IntraBlockState
	vmCfg vm.Config

	dirs datadir.Dirs

	isMining bool
}

func NewWorker(lock sync.Locker, logger log.Logger, ctx context.Context, background bool, chainDb kv.RoDB, in *exec.QueueWithRetry, blockReader services.FullBlockReader, chainConfig *chain.Config, genesis *types.Genesis, results *exec.ResultsQueue, engine consensus.Engine, dirs datadir.Dirs) *Worker {
	w := &Worker{
		lock:        lock,
		logger:      logger,
		chainDb:     chainDb,
		in:          in,
		background:  background,
		blockReader: blockReader,
		chainConfig: chainConfig,

		ctx:      ctx,
		genesis:  genesis,
		resultCh: results,
		engine:   engine,

		evm:         vm.NewEVM(evmtypes.BlockContext{}, evmtypes.TxContext{}, nil, chainConfig, vm.Config{}),
		callTracer:  NewCallTracer(),
		taskGasPool: new(core.GasPool),

		dirs: dirs,
	}
	w.taskGasPool.AddBlobGas(chainConfig.GetMaxBlobGasPerBlock())
	w.vmCfg = vm.Config{Debug: true, Tracer: w.callTracer}
	w.ibs = state.New(w.stateReader)
	return w
}

func (rw *Worker) LogLRUStats() { rw.evm.JumpDestCache.LogStats() }

func (rw *Worker) ResetState(rs *state.StateV3Buffered, stateWriter state.StateWriter, accumulator *shards.Accumulator) {
	rw.rs = rs
	if rw.background {
		rw.SetReader(state.NewReaderParallelV3(rs.Domains()))
	} else {
		rw.SetReader(state.NewReaderV3(rs.Domains()))
	}

	if stateWriter != nil {
		rw.stateWriter = stateWriter
	} else {
		rw.stateWriter = state.NewStateWriterV3(rs.StateV3, accumulator)
	}
}

func (rw *Worker) Tx() kv.Tx        { return rw.chainTx }
func (rw *Worker) DiscardReadList() { rw.stateReader.DiscardReadList() }
func (rw *Worker) ResetTx(chainTx kv.Tx) {
	if rw.background && rw.chainTx != nil {
		rw.chainTx.Rollback()
		rw.chainTx = nil
	}
	if chainTx != nil {
		rw.chainTx = chainTx
		rw.stateReader.SetTx(rw.chainTx)
		rw.chain = consensuschain.NewReader(rw.chainConfig, rw.chainTx, rw.blockReader, rw.logger)
	}
}

func (rw *Worker) Run() error {
	for txTask, ok := rw.in.Next(rw.ctx); ok; txTask, ok = rw.in.Next(rw.ctx) {
		//fmt.Println("RTX", txTask.Version().BlockNum, txTask.Version().TxIndex, txTask.Version().TxNum, txTask.IsBlockEnd())
		result := rw.RunTxTask(txTask)
		if err := rw.resultCh.Add(rw.ctx, result); err != nil {
			return err
		}
	}
	return nil
}

func (rw *Worker) RunTxTask(txTask exec.Task) *exec.Result {
	rw.lock.Lock()
	defer rw.lock.Unlock()
	return rw.RunTxTaskNoLock(txTask)
}

// Needed to set history reader when need to offset few txs from block beginning and does not break processing,
// like compute gas used for block and then to set state reader to continue processing on latest data.
func (rw *Worker) SetReader(reader state.ResettableStateReader) {
	rw.stateReader = reader
	rw.stateReader.SetTx(rw.Tx())
	rw.ibs.Reset()
	rw.ibs = state.New(rw.stateReader)

	switch reader.(type) {
	case *state.HistoryReaderV3:
		rw.historyMode = true
	case *state.ReaderV3:
		rw.historyMode = false
	default:
		rw.historyMode = false
		//fmt.Printf("[worker] unknown reader %T: historyMode is set to disabled\n", reader)
	}
}

func (rw *Worker) RunTxTaskNoLock(txTask exec.Task) *exec.Result {
	if txTask.IsHistoric() && !rw.historyMode {
		// in case if we cancelled execution and commitment happened in the middle of the block, we have to process block
		// from the beginning until committed txNum and only then disable history mode.
		// Needed to correctly evaluate spent gas and other things.
		rw.SetReader(state.NewHistoryReaderV3())
	} else if !txTask.IsHistoric() && rw.historyMode {
		if rw.background {
			rw.SetReader(state.NewBufferedReader(rw.rs, state.NewReaderParallelV3(rw.rs.Domains())))
		} else {
			rw.SetReader(state.NewBufferedReader(rw.rs, state.NewReaderV3(rw.rs.Domains())))
		}
	}
	if rw.background && rw.chainTx == nil {
		var err error
		if rw.chainTx, err = rw.chainDb.BeginRo(rw.ctx); err != nil {
			panic(err)
		}
		rw.stateReader.SetTx(rw.chainTx)
		rw.chain = consensuschain.NewReader(rw.chainConfig, rw.chainTx, rw.blockReader, rw.logger)
	}

	txIndex := txTask.Version().TxIndex

	if txIndex != -1 && !txTask.IsBlockEnd() {
		rw.callTracer.Reset()
	}

	txTask.Reset(rw.ibs)

	if txIndex >= 0 {
		rw.ibs.SetTxContext(txIndex)
	}

	result := txTask.Execute(rw.evm, rw.vmCfg, rw.engine, rw.genesis, rw.taskGasPool, rw.rs, rw.ibs,
		rw.stateWriter, rw.stateReader, rw.chainConfig, rw.chain, rw.dirs, true)

	if result.Task == nil {
		result.Task = txTask
	}

	if txIndex != -1 && !txTask.IsBlockEnd() {
		result.TraceFroms = rw.callTracer.Froms()
		result.TraceTos = rw.callTracer.Tos()
	}

	return result
}

func NewWorkersPool(lock sync.Locker, accumulator *shards.Accumulator, logger log.Logger, ctx context.Context, background bool, chainDb kv.RoDB,
	rs *state.StateV3Buffered, stateWriter state.StateWriter, in *exec.QueueWithRetry, blockReader services.FullBlockReader, chainConfig *chain.Config, genesis *types.Genesis,
	engine consensus.Engine, workerCount int, dirs datadir.Dirs, isMining bool) (reconWorkers []*Worker, applyWorker *Worker, rws *exec.ResultsQueue, clear func(), wait func()) {
	reconWorkers = make([]*Worker, workerCount)

	resultChSize := workerCount * 8
	rws = exec.NewResultsQueue(resultChSize, workerCount) // workerCount * 4
	{
		// we all errors in background workers (except ctx.Cancel), because applyLoop will detect this error anyway.
		// and in applyLoop all errors are critical
		ctx, cancel := context.WithCancel(ctx)
		g, ctx := errgroup.WithContext(ctx)
		for i := 0; i < workerCount; i++ {
			reconWorkers[i] = NewWorker(lock, logger, ctx, background, chainDb, in, blockReader, chainConfig, genesis, rws, engine, dirs)
			reconWorkers[i].ResetState(rs, stateWriter, accumulator)
		}
		if background {
			for i := 0; i < workerCount; i++ {
				i := i
				g.Go(func() error {
					return reconWorkers[i].Run()
				})
			}
			wait = func() { g.Wait() }
		}

		var clearDone bool
		clear = func() {
			if clearDone {
				return
			}
			clearDone = true
			cancel()
			g.Wait()
			for _, w := range reconWorkers {
				w.ResetTx(nil)
			}
			//applyWorker.ResetTx(nil)
		}
	}
	applyWorker = NewWorker(lock, logger, ctx, false, chainDb, in, blockReader, chainConfig, genesis, rws, engine, dirs)

	return reconWorkers, applyWorker, rws, clear, wait
}
