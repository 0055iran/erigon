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

package stagedsync

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"runtime"
	"slices"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/c2h5oh/datasize"

	"github.com/erigontech/erigon-lib/log/v3"

	libcommon "github.com/erigontech/erigon-lib/common"
	"github.com/erigontech/erigon-lib/common/dbg"
	"github.com/erigontech/erigon-lib/common/hexutility"
	"github.com/erigontech/erigon-lib/etl"
	"github.com/erigontech/erigon-lib/kv"
	"github.com/erigontech/erigon-lib/kv/bitmapdb"
	"github.com/erigontech/erigon-lib/kv/dbutils"

	"github.com/erigontech/erigon/core/types"
	"github.com/erigontech/erigon/ethdb/prune"
)

const (
	bitmapsBufLimit   = 256 * datasize.MB // limit how much memory can use bitmaps before flushing to DB
	bitmapsFlushEvery = 10 * time.Second
)

type LogIndexCfg struct {
	tmpdir     string
	db         kv.RwDB
	prune      prune.Mode
	bufLimit   datasize.ByteSize
	flushEvery time.Duration

	// For not pruning the logs of this contract since deposit contract logs are needed by CL to validate/produce blocks.
	// All logs should be available to a validating node through eth_getLogs
	depositContract *libcommon.Address
}

func StageLogIndexCfg(db kv.RwDB, prune prune.Mode, tmpDir string, depositContract *libcommon.Address) LogIndexCfg {
	return LogIndexCfg{
		db:              db,
		prune:           prune,
		bufLimit:        bitmapsBufLimit,
		flushEvery:      bitmapsFlushEvery,
		tmpdir:          tmpDir,
		depositContract: depositContract,
	}
}

func SpawnLogIndex(s *StageState, tx kv.RwTx, cfg LogIndexCfg, ctx context.Context, prematureEndBlock uint64, logger log.Logger) error {
	useExternalTx := tx != nil
	if !useExternalTx {
		var err error
		tx, err = cfg.db.BeginRw(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback()
	}

	endBlock, err := s.ExecutionAt(tx)
	if err != nil {
		return fmt.Errorf("getting last executed block: %w", err)
	}
	if s.BlockNumber > endBlock { // Erigon will self-heal (download missed blocks) eventually
		return nil
	}
	logPrefix := s.LogPrefix()
	// if prematureEndBlock is nonzero and less than the latest executed block,
	// then we only run the log index stage until prematureEndBlock
	if prematureEndBlock != 0 && prematureEndBlock < endBlock {
		endBlock = prematureEndBlock
	}
	// It is possible that prematureEndBlock < s.BlockNumber,
	// in which case it is important that we skip this stage,
	// or else we could overwrite stage_at with prematureEndBlock
	if endBlock <= s.BlockNumber {
		return nil
	}

	startBlock := s.BlockNumber
	pruneTo := cfg.prune.History.PruneTo(endBlock) //endBlock - prune.r.older
	// if startBlock < pruneTo {
	// 	startBlock = pruneTo
	// }
	if startBlock > 0 {
		startBlock++
	}
	if err = promoteLogIndex(logPrefix, tx, startBlock, endBlock, pruneTo, cfg, ctx, logger); err != nil {
		return err
	}
	if err = s.Update(tx, endBlock); err != nil {
		return err
	}

	if !useExternalTx {
		if err = tx.Commit(); err != nil {
			return err
		}
	}

	return nil
}

// Add the topics and address index for logs, if not in prune range or addr is the deposit contract
func promoteLogIndex(logPrefix string, tx kv.RwTx, start uint64, endBlock uint64, pruneBlock uint64, cfg LogIndexCfg, ctx context.Context, logger log.Logger) error {
	quit := ctx.Done()
	logEvery := time.NewTicker(30 * time.Second)
	defer logEvery.Stop()

	topics := map[string]*roaring.Bitmap{}
	addresses := map[string]*roaring.Bitmap{}
	logs, err := tx.Cursor(kv.Log)
	if err != nil {
		return err
	}
	defer logs.Close()
	checkFlushEvery := time.NewTicker(cfg.flushEvery)
	defer checkFlushEvery.Stop()

	collectorTopics := etl.NewCollector(logPrefix, cfg.tmpdir, etl.NewSortableBuffer(etl.BufferOptimalSize), logger)
	defer collectorTopics.Close()
	collectorAddrs := etl.NewCollector(logPrefix, cfg.tmpdir, etl.NewSortableBuffer(etl.BufferOptimalSize), logger)
	defer collectorAddrs.Close()

	reader := bytes.NewReader(nil)

	if endBlock != 0 && endBlock-start > 100 {
		logger.Info(fmt.Sprintf("[%s] processing", logPrefix), "from", start, "to", endBlock, "pruneTo", pruneBlock)
	}

	for k, v, err := logs.Seek(dbutils.LogKey(start, 0)); k != nil; k, v, err = logs.Next() {
		if err != nil {
			return err
		}

		if err := libcommon.Stopped(quit); err != nil {
			return err
		}
		blockNum := binary.BigEndian.Uint64(k[:8])

		// if endBlock is positive, we only run the stage up until endBlock
		// if endBlock is zero, we run the stage for all available blocks
		if endBlock != 0 && blockNum > endBlock {
			logger.Info(fmt.Sprintf("[%s] Reached user-specified end block", logPrefix), "endBlock", endBlock)
			break
		}

		select {
		default:
		case <-logEvery.C:
			var m runtime.MemStats
			dbg.ReadMemStats(&m)
			logger.Info(fmt.Sprintf("[%s] Progress", logPrefix), "number", blockNum, "alloc", libcommon.ByteCount(m.Alloc), "sys", libcommon.ByteCount(m.Sys))
		case <-checkFlushEvery.C:
			if needFlush(topics, cfg.bufLimit) {
				if err := flushBitmaps(collectorTopics, topics); err != nil {
					return err
				}
				topics = map[string]*roaring.Bitmap{}
			}

			if needFlush(addresses, cfg.bufLimit) {
				if err := flushBitmaps(collectorAddrs, addresses); err != nil {
					return err
				}
				addresses = map[string]*roaring.Bitmap{}
			}
		}

		var ll types.Logs
		reader.Reset(v)
		//if err := cbor.Unmarshal(&ll, reader); err != nil {
		//	return fmt.Errorf("receipt unmarshal failed: %w, blocl=%d", err, blockNum)
		//}

		toStore := true
		// if pruning is enabled, and depositContract isn't configured for the chain, don't index
		if blockNum < pruneBlock {
			toStore = false
			if cfg.depositContract == nil {
				continue
			}
			for _, l := range ll {
				// if any of the log address is in noPrune, store and index all logs for this txId
				if *cfg.depositContract == l.Address {
					toStore = true
					break
				}
			}
		}

		if !toStore {
			continue
		}
		for _, l := range ll {
			for _, topic := range l.Topics {
				topicStr := string(topic.Bytes())
				m, ok := topics[topicStr]
				if !ok {
					m = roaring.New()
					topics[topicStr] = m
				}
				m.Add(uint32(blockNum))
			}

			accStr := string(l.Address.Bytes())
			m, ok := addresses[accStr]
			if !ok {
				m = roaring.New()
				addresses[accStr] = m
			}
			m.Add(uint32(blockNum))
		}
	}

	if err := flushBitmaps(collectorTopics, topics); err != nil {
		return err
	}
	if err := flushBitmaps(collectorAddrs, addresses); err != nil {
		return err
	}

	var currentBitmap = roaring.New()
	var buf = bytes.NewBuffer(nil)

	lastChunkKey := make([]byte, 128)
	var loaderFunc = func(k []byte, v []byte, table etl.CurrentTableReader, next etl.LoadNextFunc) error {
		lastChunkKey = lastChunkKey[:len(k)+4]
		copy(lastChunkKey, k)
		binary.BigEndian.PutUint32(lastChunkKey[len(k):], ^uint32(0))
		lastChunkBytes, err := table.Get(lastChunkKey)
		if err != nil {
			return fmt.Errorf("find last chunk: %w", err)
		}

		lastChunk := roaring.New()
		if len(lastChunkBytes) > 0 {
			_, err = lastChunk.FromBuffer(lastChunkBytes)
			if err != nil {
				return fmt.Errorf("couldn't read last log index chunk: %w, len(lastChunkBytes)=%d", err, len(lastChunkBytes))
			}
		}

		if _, err := currentBitmap.FromBuffer(v); err != nil {
			return err
		}
		currentBitmap.Or(lastChunk) // merge last existing chunk from db - next loop will overwrite it
		return bitmapdb.WalkChunkWithKeys(k, currentBitmap, bitmapdb.ChunkLimit, func(chunkKey []byte, chunk *roaring.Bitmap) error {
			buf.Reset()
			if _, err := chunk.WriteTo(buf); err != nil {
				return err
			}
			return next(k, chunkKey, buf.Bytes())
		})
	}

	if err := collectorTopics.Load(tx, kv.LogTopicIndex, loaderFunc, etl.TransformArgs{Quit: quit}); err != nil {
		return err
	}

	if err := collectorAddrs.Load(tx, kv.LogAddressIndex, loaderFunc, etl.TransformArgs{Quit: quit}); err != nil {
		return err
	}

	return nil
}

func UnwindLogIndex(u *UnwindState, s *StageState, tx kv.RwTx, cfg LogIndexCfg, ctx context.Context) (err error) {
	quitCh := ctx.Done()
	useExternalTx := tx != nil
	if !useExternalTx {
		tx, err = cfg.db.BeginRw(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback()
	}

	logPrefix := s.LogPrefix()
	if err := unwindLogIndex(logPrefix, tx, u.UnwindPoint, cfg, quitCh); err != nil {
		return err
	}

	if err := u.Done(tx); err != nil {
		return fmt.Errorf("%w", err)
	}
	if !useExternalTx {
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func unwindLogIndex(logPrefix string, db kv.RwTx, to uint64, cfg LogIndexCfg, quitCh <-chan struct{}) error {
	topics := map[string]struct{}{}
	addrs := map[string]struct{}{}

	reader := bytes.NewReader(nil)
	c, err := db.Cursor(kv.Log)
	if err != nil {
		return err
	}
	defer c.Close()
	for k, v, err := c.Seek(hexutility.EncodeTs(to + 1)); k != nil; k, v, err = c.Next() {
		if err != nil {
			return err
		}

		if err := libcommon.Stopped(quitCh); err != nil {
			return err
		}
		var logs types.Logs
		reader.Reset(v)
		//if err := cbor.Unmarshal(&logs, reader); err != nil {
		//	return fmt.Errorf("receipt unmarshal: %w, block=%d", err, binary.BigEndian.Uint64(k))
		//}

		for _, l := range logs {
			for _, topic := range l.Topics {
				topics[string(topic.Bytes())] = struct{}{}
			}
			addrs[string(l.Address.Bytes())] = struct{}{}
		}
	}

	if err := truncateBitmaps(db, kv.LogTopicIndex, topics, to); err != nil {
		return err
	}
	if err := truncateBitmaps(db, kv.LogAddressIndex, addrs, to); err != nil {
		return err
	}
	return nil
}

func needFlush(bitmaps map[string]*roaring.Bitmap, memLimit datasize.ByteSize) bool {
	sz := uint64(0)
	for _, m := range bitmaps {
		sz += m.GetSizeInBytes() * 2 // for golang's overhead
	}
	const memoryNeedsForKey = 32 * 2 * 2 //  len(key) * (string and bytes) overhead * go's map overhead
	return uint64(len(bitmaps)*memoryNeedsForKey)+sz > uint64(memLimit)
}

func flushBitmaps(c *etl.Collector, inMem map[string]*roaring.Bitmap) error {
	for k, v := range inMem {
		v.RunOptimize()
		if v.GetCardinality() == 0 {
			continue
		}
		newV := bytes.NewBuffer(make([]byte, 0, v.GetSerializedSizeInBytes()))
		if _, err := v.WriteTo(newV); err != nil {
			return err
		}
		if err := c.Collect([]byte(k), newV.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

func truncateBitmaps(tx kv.RwTx, bucket string, inMem map[string]struct{}, to uint64) error {
	keys := make([]string, 0, len(inMem))
	for k := range inMem {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		if err := bitmapdb.TruncateRange(tx, bucket, []byte(k), uint32(to+1)); err != nil {
			return fmt.Errorf("fail TruncateRange: bucket=%s, %w", bucket, err)
		}
	}

	return nil
}

func pruneOldLogChunks(tx kv.RwTx, bucket string, inMem *etl.Collector, pruneTo uint64, ctx context.Context) error {
	logEvery := time.NewTicker(logInterval)
	defer logEvery.Stop()

	c, err := tx.RwCursor(bucket)
	if err != nil {
		return err
	}
	defer c.Close()

	if err := inMem.Load(tx, bucket, func(key, v []byte, table etl.CurrentTableReader, next etl.LoadNextFunc) error {
		for k, _, err := c.Seek(key); k != nil; k, _, err = c.Next() {
			if err != nil {
				return err
			}
			var blockNum uint64
			blockNum = uint64(binary.BigEndian.Uint32(k[len(key):]))

			if !bytes.HasPrefix(k, key) || blockNum >= pruneTo {
				break
			}

			if err = c.DeleteCurrent(); err != nil {
				return fmt.Errorf("failed delete log/index, bucket=%v block=%d: %w", bucket, blockNum, err)
			}
		}
		return nil
	}, etl.TransformArgs{
		Quit: ctx.Done(),
	}); err != nil {
		return err
	}
	return nil
}

// Call pruneLogIndex with the current sync progresses and commit the data to db
func PruneLogIndex(s *PruneState, tx kv.RwTx, cfg LogIndexCfg, ctx context.Context, logger log.Logger) (err error) {
	if !cfg.prune.History.Enabled() {
		return nil
	}
	logPrefix := s.LogPrefix()

	useExternalTx := tx != nil
	if !useExternalTx {
		tx, err = cfg.db.BeginRw(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback()
	}

	pruneTo := cfg.prune.History.PruneTo(s.ForwardProgress)
	if err = pruneLogIndex(logPrefix, tx, cfg.tmpdir, s.PruneProgress, pruneTo, ctx, logger, cfg.depositContract); err != nil {
		return err
	}
	if err = s.DoneAt(tx, pruneTo); err != nil {
		return err
	}

	if !useExternalTx {
		if err = tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

// Prune log indexes as well as logs within the prune range
func pruneLogIndex(logPrefix string, tx kv.RwTx, tmpDir string, pruneFrom, pruneTo uint64, ctx context.Context, logger log.Logger, depositContract *libcommon.Address) error {
	logEvery := time.NewTicker(logInterval)
	defer logEvery.Stop()

	bufferSize := etl.BufferOptimalSize
	topics := etl.NewCollector(logPrefix, tmpDir, etl.NewOldestEntryBuffer(bufferSize), logger)
	defer topics.Close()
	addrs := etl.NewCollector(logPrefix, tmpDir, etl.NewOldestEntryBuffer(bufferSize), logger)
	defer addrs.Close()

	reader := bytes.NewReader(nil)
	{
		c, err := tx.Cursor(kv.Log)
		if err != nil {
			return err
		}
		defer c.Close()

		for k, v, err := c.Seek(dbutils.LogKey(pruneFrom, 0)); k != nil; k, v, err = c.Next() {
			if err != nil {
				return err
			}
			blockNum := binary.BigEndian.Uint64(k)
			if blockNum >= pruneTo {
				break
			}
			select {
			case <-logEvery.C:
				logger.Info(fmt.Sprintf("[%s]", logPrefix), "table", kv.Log, "block", blockNum, "pruneFrom", pruneFrom, "pruneTo", pruneTo)
			case <-ctx.Done():
				return libcommon.ErrStopped
			default:
			}

			var logs types.Logs
			reader.Reset(v)
			//if err := cbor.Unmarshal(&logs, reader); err != nil {
			//	return fmt.Errorf("receipt unmarshal failed: %w, block=%d", err, binary.BigEndian.Uint64(k))
			//}

			toPrune := true
			for _, l := range logs {
				// No logs (or sublogs) for this txId should be pruned
				// if one of the logs belongs to the deposit contract
				if depositContract != nil && *depositContract == l.Address {
					toPrune = false
					break
				}
			}

			if toPrune {
				for _, l := range logs {
					for _, topic := range l.Topics {
						if err := topics.Collect(topic.Bytes(), nil); err != nil {
							return err
						}
					}
					if err := addrs.Collect(l.Address.Bytes(), nil); err != nil {
						return err
					}
				}
				if err := tx.Delete(kv.Log, k); err != nil {
					return err
				}
			}
		}
	}

	if err := pruneOldLogChunks(tx, kv.LogTopicIndex, topics, pruneTo, ctx); err != nil {
		return err
	}
	if err := pruneOldLogChunks(tx, kv.LogAddressIndex, addrs, pruneTo, ctx); err != nil {
		return err
	}
	return nil
}