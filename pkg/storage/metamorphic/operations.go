// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package metamorphic

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/uint128"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
)

// opReference represents one operation; an opGenerator reference as well as
// bound arguments.
type opReference struct {
	generator *opGenerator
	args      []string
}

// opRun represents one operation run; a generated operation, and bound
// arguments.
type opRun struct {
	name string
	op   mvccOp
	// TODO(itsbilal): Instead of storing arguments separately and printing them,
	// have mvccOp be able to print and parse its arguments itself. This would
	// give us more freedom with printing of arguments and would easily correct
	// for things like endKey and key being swapped if endKey < key. It would also
	// let us avoid the opGenerator.generate() call completely when
	// parsing/checking an already-generated output file.
	args []string
	// The following fields are only used in the "check mode", in parseFileAndRun.
	lineNum        uint64
	expectedOutput string
}

// mvccOp represents an operation instance that can be run. It's generated by
// an instance of opGenerator.
type mvccOp interface {
	// run runs the operation. An output string is returned.
	run(ctx context.Context) string
}

// An opGenerator instance represents one type of an operation. The run and
// dependantOps commands should be stateless, with all state stored in the
// passed-in test runner or its operand generators.
type opGenerator struct {
	// Name of the operation. Used in file output and parsing.
	name string
	// Function to call to generate the operation runner.
	generate func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp
	// Returns a list of operation runs that must happen before this operation.
	// Note that openers for non-existent operands are handled separately and
	// don't need to be handled here.
	dependentOps func(m *metaTestRunner, args ...string) []opReference
	// Operands this operation expects. Passed in the same order to run and
	// dependentOps.
	operands []operandType
	// weight is used to denote frequency of this operation to the TPCC-style
	// deck.
	//
	// Note that the generator tends to bias towards opener operations; since
	// an opener can be generated outside of the deck shuffle, in resolveAndAddOp
	// to create an instance of an operand that does not exist. To counter this
	// bias, we try to keep the sum of opener operations to be less than half
	// the sum of "closer" operations for an operand type, to prevent too many
	// of that type of object from accumulating throughout the run.
	weight int
	// isOpener denotes whether this operation is an opener. Opener operations are
	// special, in that the last operand specified is generated from a getNew()
	// call to the matching operand generator, instead of the usual get().
	isOpener bool
}

// Helper function to generate iterator_close opRuns for all iterators on a
// passed-in Batch.
func closeItersOnBatch(m *metaTestRunner, reader readWriterID) (results []opReference) {
	// No need to close iters on non-batches (i.e. engines).
	if reader == "engine" {
		return
	}
	// Close all iterators for this batch first.
	for _, iter := range m.iterGenerator.readerToIter[reader] {
		results = append(results, opReference{
			generator: m.nameToGenerator["iterator_close"],
			args:      []string{string(iter)},
		})
	}
	return
}

// Helper function to run MVCCScan given a key range and a reader.
func generateMVCCScan(
	ctx context.Context, m *metaTestRunner, reverse bool, inconsistent bool, args []string,
) *mvccScanOp {
	key := m.keyGenerator.parse(args[0])
	endKey := m.keyGenerator.parse(args[1])
	if endKey.Less(key) {
		key, endKey = endKey, key
	}
	var ts hlc.Timestamp
	var txn txnID
	if inconsistent {
		ts = m.pastTSGenerator.parse(args[2])
	} else {
		txn = txnID(args[2])
	}
	return &mvccScanOp{
		m:            m,
		key:          key.Key,
		endKey:       endKey.Key,
		ts:           ts,
		txn:          txn,
		inconsistent: inconsistent,
		reverse:      reverse,
	}
}

// Prints the key where an iterator is positioned, or valid = false if invalid.
func printIterState(iter storage.Iterator) string {
	if ok, err := iter.Valid(); !ok || err != nil {
		if err != nil {
			return fmt.Sprintf("valid = %v, err = %s", ok, err.Error())
		}
		return "valid = false"
	}
	return fmt.Sprintf("key = %s", iter.UnsafeKey().String())
}

type mvccGetOp struct {
	m            *metaTestRunner
	reader       readWriterID
	key          roachpb.Key
	ts           hlc.Timestamp
	txn          txnID
	inconsistent bool
}

func (m mvccGetOp) run(ctx context.Context) string {
	reader := m.m.getReadWriter(m.reader)
	var txn *roachpb.Transaction
	if !m.inconsistent {
		txn = m.m.getTxn(m.txn)
		m.ts = txn.ReadTimestamp
	}
	// TODO(itsbilal): Specify these bools as operands instead of having a
	// separate operation for inconsistent cases. This increases visibility for
	// anyone reading the output file.
	val, intent, err := storage.MVCCGet(ctx, reader, m.key, m.ts, storage.MVCCGetOptions{
		Inconsistent: m.inconsistent,
		Tombstones:   true,
		Txn:          txn,
	})
	if err != nil {
		return fmt.Sprintf("error: %s", err)
	}
	return fmt.Sprintf("val = %v, intent = %v", val, intent)
}

type mvccPutOp struct {
	m      *metaTestRunner
	writer readWriterID
	key    roachpb.Key
	value  roachpb.Value
	txn    txnID
}

func (m mvccPutOp) run(ctx context.Context) string {
	txn := m.m.getTxn(m.txn)
	txn.Sequence++
	writer := m.m.getReadWriter(m.writer)

	err := storage.MVCCPut(ctx, writer, nil, m.key, txn.WriteTimestamp, m.value, txn)
	if err != nil {
		return fmt.Sprintf("error: %s", err)
	}

	// Update the txn's lock spans to account for this intent being written.
	txn.LockSpans = append(txn.LockSpans, roachpb.Span{
		Key: m.key,
	})
	return "ok"
}

type mvccCPutOp struct {
	m      *metaTestRunner
	writer readWriterID
	key    roachpb.Key
	value  roachpb.Value
	expVal roachpb.Value
	txn    txnID
}

func (m mvccCPutOp) run(ctx context.Context) string {
	txn := m.m.getTxn(m.txn)
	writer := m.m.getReadWriter(m.writer)
	txn.Sequence++

	err := storage.MVCCConditionalPut(ctx, writer, nil, m.key, txn.WriteTimestamp, m.value, &m.expVal, true, txn)
	if err != nil {
		return fmt.Sprintf("error: %s", err)
	}

	// Update the txn's lock spans to account for this intent being written.
	txn.LockSpans = append(txn.LockSpans, roachpb.Span{
		Key: m.key,
	})
	return "ok"
}

type mvccInitPutOp struct {
	m      *metaTestRunner
	writer readWriterID
	key    roachpb.Key
	value  roachpb.Value
	txn    txnID
}

func (m mvccInitPutOp) run(ctx context.Context) string {
	txn := m.m.getTxn(m.txn)
	writer := m.m.getReadWriter(m.writer)
	txn.Sequence++

	err := storage.MVCCInitPut(ctx, writer, nil, m.key, txn.WriteTimestamp, m.value, false, txn)
	if err != nil {
		return fmt.Sprintf("error: %s", err)
	}

	// Update the txn's lock spans to account for this intent being written.
	txn.LockSpans = append(txn.LockSpans, roachpb.Span{
		Key: m.key,
	})
	return "ok"
}

type mvccDeleteRangeOp struct {
	m      *metaTestRunner
	writer readWriterID
	key    roachpb.Key
	endKey roachpb.Key
	txn    txnID
}

func (m mvccDeleteRangeOp) run(ctx context.Context) string {
	txn := m.m.getTxn(m.txn)
	writer := m.m.getReadWriter(m.writer)
	txn.Sequence++

	keys, _, _, err := storage.MVCCDeleteRange(ctx, writer, nil, m.key, m.endKey, 0, txn.WriteTimestamp, txn, true)
	if err != nil {
		return fmt.Sprintf("error: %s", err)
	}

	// Update the txn's lock spans to account for this intent being written.
	for _, key := range keys {
		txn.LockSpans = append(txn.LockSpans, roachpb.Span{
			Key: key,
		})
	}
	return "ok"
}

type mvccClearTimeRangeOp struct {
	m         *metaTestRunner
	writer    readWriterID
	key       roachpb.Key
	endKey    roachpb.Key
	startTime hlc.Timestamp
	endTime   hlc.Timestamp
}

func (m mvccClearTimeRangeOp) run(ctx context.Context) string {
	writer := m.m.getReadWriter(m.writer)
	span, err := storage.MVCCClearTimeRange(ctx, writer, nil, m.key, m.endKey, m.startTime, m.endTime, math.MaxInt64)
	if err != nil {
		return fmt.Sprintf("error: %s", err)
	}
	return fmt.Sprintf("ok, span = %v", span)
}

type mvccDeleteOp struct {
	m      *metaTestRunner
	writer readWriterID
	key    roachpb.Key
	txn    txnID
}

func (m mvccDeleteOp) run(ctx context.Context) string {
	txn := m.m.getTxn(m.txn)
	writer := m.m.getReadWriter(m.writer)
	txn.Sequence++

	err := storage.MVCCDelete(ctx, writer, nil, m.key, txn.WriteTimestamp, txn)
	if err != nil {
		return fmt.Sprintf("error: %s", err)
	}

	// Update the txn's lock spans to account for this intent being written.
	txn.LockSpans = append(txn.LockSpans, roachpb.Span{
		Key: m.key,
	})
	return "ok"
}

type mvccFindSplitKeyOp struct {
	m         *metaTestRunner
	key       roachpb.RKey
	endKey    roachpb.RKey
	splitSize int64
}

func (m mvccFindSplitKeyOp) run(ctx context.Context) string {
	splitKey, err := storage.MVCCFindSplitKey(ctx, m.m.engine, m.key, m.endKey, m.splitSize)
	if err != nil {
		return fmt.Sprintf("error: %s", err)
	}

	return fmt.Sprintf("ok, splitSize = %d, splitKey = %v", m.splitSize, splitKey)
}

type mvccScanOp struct {
	m            *metaTestRunner
	key          roachpb.Key
	endKey       roachpb.Key
	ts           hlc.Timestamp
	txn          txnID
	inconsistent bool
	reverse      bool
}

func (m mvccScanOp) run(ctx context.Context) string {
	var txn *roachpb.Transaction
	if !m.inconsistent {
		txn = m.m.getTxn(m.txn)
		m.ts = txn.ReadTimestamp
	}
	// While MVCCScanning on a batch works in Pebble, it does not in rocksdb.
	// This is due to batch iterators not supporting SeekForPrev. For now, use
	// m.engine instead of a readWriterGenerator-generated engine.Reader, otherwise
	// we will try MVCCScanning on batches and produce diffs between runs on
	// different engines that don't point to an actual issue.
	result, err := storage.MVCCScan(ctx, m.m.engine, m.key, m.endKey, m.ts, storage.MVCCScanOptions{
		Inconsistent: m.inconsistent,
		Tombstones:   true,
		Reverse:      m.reverse,
		Txn:          txn,
	})
	if err != nil {
		return fmt.Sprintf("error: %s", err)
	}
	return fmt.Sprintf("kvs = %v, intents = %v", result.KVs, result.Intents)
}

type txnOpenOp struct {
	m  *metaTestRunner
	id txnID
	ts hlc.Timestamp
}

func (t txnOpenOp) run(ctx context.Context) string {
	var id uint64
	if _, err := fmt.Sscanf(string(t.id), "t%d", &id); err != nil {
		panic(err)
	}
	txn := &roachpb.Transaction{
		TxnMeta: enginepb.TxnMeta{
			ID:             uuid.FromUint128(uint128.FromInts(0, id)),
			Key:            roachpb.KeyMin,
			WriteTimestamp: t.ts,
			Sequence:       0,
		},
		Name:                    string(t.id),
		DeprecatedOrigTimestamp: t.ts,
		ReadTimestamp:           t.ts,
		Status:                  roachpb.PENDING,
	}
	t.m.setTxn(t.id, txn)
	return txn.Name
}

type txnCommitOp struct {
	m  *metaTestRunner
	id txnID
}

func (t txnCommitOp) run(ctx context.Context) string {
	txn := t.m.getTxn(t.id)
	txn.Status = roachpb.COMMITTED

	for _, span := range txn.LockSpans {
		intent := roachpb.MakeLockUpdate(txn, span)
		intent.Status = roachpb.COMMITTED
		_, err := storage.MVCCResolveWriteIntent(context.TODO(), t.m.engine, nil, intent)
		if err != nil {
			panic(err)
		}
	}
	delete(t.m.openTxns, t.id)

	return "ok"
}

type batchOpenOp struct {
	m  *metaTestRunner
	id readWriterID
}

func (b batchOpenOp) run(ctx context.Context) string {
	batch := b.m.engine.NewBatch()
	b.m.setReadWriter(b.id, batch)
	return string(b.id)
}

type batchCommitOp struct {
	m  *metaTestRunner
	id readWriterID
}

func (b batchCommitOp) run(ctx context.Context) string {
	if b.id == "engine" {
		return "noop"
	}
	batch := b.m.getReadWriter(b.id).(storage.Batch)
	if err := batch.Commit(true); err != nil {
		return err.Error()
	}
	batch.Close()
	delete(b.m.openBatches, b.id)
	return "ok"
}

type iterOpenOp struct {
	m      *metaTestRunner
	rw     readWriterID
	key    roachpb.Key
	endKey roachpb.Key
	id     iteratorID
}

func (i iterOpenOp) run(ctx context.Context) string {
	rw := i.m.getReadWriter(i.rw)
	iter := rw.NewIterator(storage.IterOptions{
		Prefix:     false,
		LowerBound: i.key,
		UpperBound: i.endKey.Next(),
	})

	i.m.setIterInfo(i.id, iteratorInfo{
		id:          i.id,
		lowerBound:  i.key,
		iter:        iter,
		isBatchIter: i.rw != "engine",
	})

	if _, ok := rw.(storage.Batch); ok {
		// When Next()-ing on a newly initialized batch iter without a key,
		// pebble's iterator stays invalid while RocksDB's finds the key after
		// the first key. This is a known difference. For now seek the iterator
		// to standardize behavior for this test.
		iter.SeekGE(storage.MakeMVCCMetadataKey(i.key))
	}

	return string(i.id)
}

type iterCloseOp struct {
	m  *metaTestRunner
	id iteratorID
}

func (i iterCloseOp) run(ctx context.Context) string {
	iterInfo := i.m.getIterInfo(i.id)
	iterInfo.iter.Close()
	delete(i.m.openIters, i.id)
	return "ok"
}

type iterSeekOp struct {
	m      *metaTestRunner
	iter   iteratorID
	key    storage.MVCCKey
	seekLT bool
}

func (i iterSeekOp) run(ctx context.Context) string {
	iterInfo := i.m.getIterInfo(i.iter)
	iter := iterInfo.iter
	if iterInfo.isBatchIter {
		if i.seekLT {
			return "noop due to missing seekLT support in rocksdb batch iterators"
		}
		// RocksDB batch iterators do not account for lower bounds consistently:
		// https://github.com/cockroachdb/cockroach/issues/44512
		// In the meantime, ensure the SeekGE key >= lower bound.
		lowerBound := iterInfo.lowerBound
		if i.key.Key.Compare(lowerBound) < 0 {
			i.key.Key = lowerBound
		}
	}
	if i.seekLT {
		iter.SeekLT(i.key)
	} else {
		iter.SeekGE(i.key)
	}

	return printIterState(iter)
}

type iterNextOp struct {
	m       *metaTestRunner
	iter    iteratorID
	nextKey bool
}

func (i iterNextOp) run(ctx context.Context) string {
	iter := i.m.getIterInfo(i.iter).iter
	// The rocksdb iterator does not treat kindly to a Next() if it is already
	// invalid. Don't run next if that is the case.
	if ok, err := iter.Valid(); !ok || err != nil {
		if err != nil {
			return fmt.Sprintf("valid = %v, err = %s", ok, err.Error())
		}
		return "valid = false"
	}
	if i.nextKey {
		iter.NextKey()
	} else {
		iter.Next()
	}

	return printIterState(iter)
}

type iterPrevOp struct {
	m    *metaTestRunner
	iter iteratorID
}

func (i iterPrevOp) run(ctx context.Context) string {
	iterInfo := i.m.getIterInfo(i.iter)
	iter := iterInfo.iter
	if iterInfo.isBatchIter {
		return "noop due to missing Prev support in rocksdb batch iterators"
	}
	// The rocksdb iterator does not treat kindly to a Prev() if it is already
	// invalid. Don't run prev if that is the case.
	if ok, err := iter.Valid(); !ok || err != nil {
		if err != nil {
			return fmt.Sprintf("valid = %v, err = %s", ok, err.Error())
		}
		return "valid = false"
	}
	iter.Prev()

	return printIterState(iter)
}

type clearRangeOp struct {
	m      *metaTestRunner
	key    roachpb.Key
	endKey roachpb.Key
}

func (c clearRangeOp) run(ctx context.Context) string {
	// All ClearRange calls in Cockroach usually happen with metadata keys, so
	// mimic the same behavior here.
	err := c.m.engine.ClearRange(storage.MakeMVCCMetadataKey(c.key), storage.MakeMVCCMetadataKey(c.endKey))
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	return "ok"
}

type compactOp struct {
	m      *metaTestRunner
	key    roachpb.Key
	endKey roachpb.Key
}

func (c compactOp) run(ctx context.Context) string {
	err := c.m.engine.CompactRange(c.key, c.endKey, false)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	return "ok"
}

type ingestOp struct {
	m    *metaTestRunner
	keys []storage.MVCCKey
}

func (i ingestOp) run(ctx context.Context) string {
	sstPath := filepath.Join(i.m.path, "ingest.sst")
	f, err := os.Create(sstPath)
	if err != nil {
		return fmt.Sprintf("error = %s", err.Error())
	}
	defer f.Close()

	sstWriter := storage.MakeIngestionSSTWriter(f)
	for _, key := range i.keys {
		_ = sstWriter.Put(key, []byte("ingested"))
	}
	if err := sstWriter.Finish(); err != nil {
		return fmt.Sprintf("error = %s", err.Error())
	}
	sstWriter.Close()

	if err := i.m.engine.IngestExternalFiles(ctx, []string{sstPath}); err != nil {
		return fmt.Sprintf("error = %s", err.Error())
	}

	return "ok"
}

type restartOp struct {
	m *metaTestRunner
}

func (r restartOp) run(ctx context.Context) string {
	if !r.m.restarts {
		r.m.printComment("no-op due to restarts being disabled")
		return "ok"
	}

	oldEngineName, newEngineName := r.m.restart()
	r.m.printComment(fmt.Sprintf("restarting: %s -> %s", oldEngineName, newEngineName))
	return "ok"
}

// List of operation generators, where each operation is defined as one instance of opGenerator.
//
// TODO(itsbilal): Add more missing MVCC operations, such as:
//  - MVCCBlindPut
//  - MVCCMerge
//  - MVCCIncrement
//  - MVCCResolveWriteIntent in the aborted case
//  - and any others that would be important to test.
var opGenerators = []opGenerator{
	{
		name: "mvcc_inconsistent_get",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			reader := readWriterID(args[0])
			key := m.keyGenerator.parse(args[1])
			ts := m.pastTSGenerator.parse(args[2])
			return &mvccGetOp{
				m:            m,
				reader:       reader,
				key:          key.Key,
				ts:           ts,
				inconsistent: true,
			}
		},
		operands: []operandType{
			operandReadWriter,
			operandMVCCKey,
			operandPastTS,
		},
		weight: 100,
	},
	{
		name: "mvcc_get",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			reader := readWriterID(args[0])
			key := m.keyGenerator.parse(args[1])
			txn := txnID(args[2])
			return &mvccGetOp{
				m:      m,
				reader: reader,
				key:    key.Key,
				txn:    txn,
			}
		},
		operands: []operandType{
			operandReadWriter,
			operandMVCCKey,
			operandTransaction,
		},
		weight: 100,
	},
	{
		name: "mvcc_put",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			writer := readWriterID(args[0])
			key := m.keyGenerator.parse(args[1])
			value := roachpb.MakeValueFromBytes(m.valueGenerator.parse(args[2]))
			txn := txnID(args[3])

			// Track this write in the txn generator. This ensures the batch will be
			// committed before the transaction is committed
			m.txnGenerator.trackWriteOnBatch(writer, txn)
			return &mvccPutOp{
				m:      m,
				writer: writer,
				key:    key.Key,
				value:  value,
				txn:    txn,
			}
		},
		operands: []operandType{
			operandReadWriter,
			operandMVCCKey,
			operandValue,
			operandTransaction,
		},
		weight: 500,
	},
	{
		name: "mvcc_conditional_put",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			writer := readWriterID(args[0])
			key := m.keyGenerator.parse(args[1])
			value := roachpb.MakeValueFromBytes(m.valueGenerator.parse(args[2]))
			expVal := roachpb.MakeValueFromBytes(m.valueGenerator.parse(args[3]))
			txn := txnID(args[4])

			// Track this write in the txn generator. This ensures the batch will be
			// committed before the transaction is committed
			m.txnGenerator.trackWriteOnBatch(writer, txn)
			return &mvccCPutOp{
				m:      m,
				writer: writer,
				key:    key.Key,
				value:  value,
				expVal: expVal,
				txn:    txn,
			}
		},
		operands: []operandType{
			operandReadWriter,
			operandMVCCKey,
			operandValue,
			operandValue,
			operandTransaction,
		},
		weight: 50,
	},
	{
		name: "mvcc_init_put",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			writer := readWriterID(args[0])
			key := m.keyGenerator.parse(args[1])
			value := roachpb.MakeValueFromBytes(m.valueGenerator.parse(args[2]))
			txn := txnID(args[3])

			// Track this write in the txn generator. This ensures the batch will be
			// committed before the transaction is committed
			m.txnGenerator.trackWriteOnBatch(writer, txn)
			return &mvccInitPutOp{
				m:      m,
				writer: writer,
				key:    key.Key,
				value:  value,
				txn:    txn,
			}
		},
		operands: []operandType{
			operandReadWriter,
			operandMVCCKey,
			operandValue,
			operandTransaction,
		},
		weight: 50,
	},
	{
		name: "mvcc_delete_range",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			writer := readWriterID(args[0])
			key := m.keyGenerator.parse(args[1]).Key
			endKey := m.keyGenerator.parse(args[2]).Key
			txn := txnID(args[3])

			if endKey.Compare(key) < 0 {
				key, endKey = endKey, key
			}

			// Track this write in the txn generator. This ensures the batch will be
			// committed before the transaction is committed
			m.txnGenerator.trackWriteOnBatch(writer, txn)
			return &mvccDeleteRangeOp{
				m:      m,
				writer: writer,
				key:    key,
				endKey: endKey,
				txn:    txn,
			}
		},
		dependentOps: func(m *metaTestRunner, args ...string) (results []opReference) {
			return closeItersOnBatch(m, readWriterID(args[0]))
		},
		operands: []operandType{
			operandReadWriter,
			operandMVCCKey,
			operandMVCCKey,
			operandTransaction,
		},
		weight: 20,
	},
	{
		name: "mvcc_clear_time_range",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			writer := readWriterID(args[0])
			key := m.keyGenerator.parse(args[1]).Key
			endKey := m.keyGenerator.parse(args[2]).Key
			startTime := m.pastTSGenerator.parse(args[3])
			endTime := m.pastTSGenerator.parse(args[3])

			if endKey.Compare(key) < 0 {
				key, endKey = endKey, key
			}
			if endTime.Less(startTime) {
				startTime, endTime = endTime, startTime
			}
			return &mvccClearTimeRangeOp{
				m:         m,
				writer:    writer,
				key:       key,
				endKey:    endKey,
				startTime: startTime,
				endTime:   endTime,
			}
		},
		dependentOps: func(m *metaTestRunner, args ...string) (results []opReference) {
			return closeItersOnBatch(m, readWriterID(args[0]))
		},
		operands: []operandType{
			operandReadWriter,
			operandMVCCKey,
			operandMVCCKey,
			operandPastTS,
			operandPastTS,
		},
		weight: 20,
	},
	{
		name: "mvcc_delete",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			writer := readWriterID(args[0])
			key := m.keyGenerator.parse(args[1])
			txn := txnID(args[2])

			// Track this write in the txn generator. This ensures the batch will be
			// committed before the transaction is committed
			m.txnGenerator.trackWriteOnBatch(writer, txn)
			return &mvccDeleteOp{
				m:      m,
				writer: writer,
				key:    key.Key,
				txn:    txn,
			}
		},
		operands: []operandType{
			operandReadWriter,
			operandMVCCKey,
			operandTransaction,
		},
		weight: 100,
	},
	{
		name: "mvcc_find_split_key",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			key, _ := keys.Addr(m.keyGenerator.parse(args[0]).Key)
			endKey, _ := keys.Addr(m.keyGenerator.parse(args[1]).Key)
			splitSize := int64(1024)

			return &mvccFindSplitKeyOp{
				m:         m,
				key:       key,
				endKey:    endKey,
				splitSize: splitSize,
			}
		},
		operands: []operandType{
			operandMVCCKey,
			operandMVCCKey,
		},
		weight: 20,
	},
	{
		name: "mvcc_scan",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			return generateMVCCScan(ctx, m, false, false, args)
		},
		operands: []operandType{
			operandMVCCKey,
			operandMVCCKey,
			operandTransaction,
		},
		weight: 100,
	},
	{
		name: "mvcc_inconsistent_scan",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			return generateMVCCScan(ctx, m, false, true, args)
		},
		operands: []operandType{
			operandMVCCKey,
			operandMVCCKey,
			operandPastTS,
		},
		weight: 100,
	},
	{
		name: "mvcc_reverse_scan",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			return generateMVCCScan(ctx, m, true, false, args)
		},
		operands: []operandType{
			operandMVCCKey,
			operandMVCCKey,
			operandTransaction,
		},
		weight: 100,
	},
	{
		name: "txn_open",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			return &txnOpenOp{
				m:  m,
				id: txnID(args[1]),
				ts: m.nextTSGenerator.parse(args[0]),
			}
		},
		operands: []operandType{
			operandNextTS,
			operandTransaction,
		},
		weight:   40,
		isOpener: true,
	},
	{
		name: "txn_commit",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			m.txnGenerator.generateClose(txnID(args[0]))
			return &txnCommitOp{
				m:  m,
				id: txnID(args[0]),
			}
		},
		dependentOps: func(m *metaTestRunner, args ...string) (result []opReference) {
			txn := txnID(args[0])

			// A transaction could have in-flight writes in some batches. Get a list
			// of all those batches, and dispatch batch_commit operations for them.
			for batch := range m.txnGenerator.openBatches[txn] {
				result = append(result, opReference{
					generator: m.nameToGenerator["batch_commit"],
					args:      []string{string(batch)},
				})
			}
			return
		},
		operands: []operandType{
			operandTransaction,
		},
		weight: 100,
	},
	{
		name: "batch_open",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			batchID := readWriterID(args[0])
			return &batchOpenOp{
				m:  m,
				id: batchID,
			}
		},
		operands: []operandType{
			operandReadWriter,
		},
		weight:   40,
		isOpener: true,
	},
	{
		name: "batch_commit",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			batchID := readWriterID(args[0])
			m.rwGenerator.generateClose(batchID)

			return &batchCommitOp{
				m:  m,
				id: batchID,
			}
		},
		dependentOps: func(m *metaTestRunner, args ...string) (results []opReference) {
			return closeItersOnBatch(m, readWriterID(args[0]))
		},
		operands: []operandType{
			operandReadWriter,
		},
		weight: 100,
	},
	{
		name: "iterator_open",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			key := m.keyGenerator.parse(args[1])
			endKey := m.keyGenerator.parse(args[2])
			iterID := iteratorID(args[3])
			if endKey.Less(key) {
				key, endKey = endKey, key
			}
			rw := readWriterID(args[0])
			m.iterGenerator.generateOpen(rw, iterID)
			return &iterOpenOp{
				m:      m,
				rw:     rw,
				key:    key.Key,
				endKey: endKey.Key,
				id:     iterID,
			}
		},
		dependentOps: func(m *metaTestRunner, args ...string) (results []opReference) {
			return closeItersOnBatch(m, readWriterID(args[0]))
		},
		operands: []operandType{
			operandReadWriter,
			operandMVCCKey,
			operandMVCCKey,
			operandIterator,
		},
		weight:   20,
		isOpener: true,
	},
	{
		name: "iterator_close",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			iter := iteratorID(args[0])
			m.iterGenerator.generateClose(iter)
			return &iterCloseOp{
				m:  m,
				id: iter,
			}
		},
		operands: []operandType{
			operandIterator,
		},
		weight: 50,
	},
	{
		name: "iterator_seekge",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			iter := iteratorID(args[0])
			key := m.keyGenerator.parse(args[1])
			return &iterSeekOp{
				m:      m,
				iter:   iter,
				key:    key,
				seekLT: false,
			}
		},
		operands: []operandType{
			operandIterator,
			operandMVCCKey,
		},
		weight: 50,
	},
	{
		name: "iterator_seeklt",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			iter := iteratorID(args[0])
			key := m.keyGenerator.parse(args[1])

			return &iterSeekOp{
				m:      m,
				iter:   iter,
				key:    key,
				seekLT: true,
			}
		},
		operands: []operandType{
			operandIterator,
			operandMVCCKey,
		},
		weight: 50,
	},
	{
		name: "iterator_next",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			iter := iteratorID(args[0])

			return &iterNextOp{
				m:       m,
				iter:    iter,
				nextKey: false,
			}
		},
		operands: []operandType{
			operandIterator,
		},
		weight: 100,
	},
	{
		name: "iterator_nextkey",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			iter := iteratorID(args[0])
			return &iterNextOp{
				m:       m,
				iter:    iter,
				nextKey: false,
			}
		},
		operands: []operandType{
			operandIterator,
		},
		weight: 100,
	},
	{
		name: "iterator_prev",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			iter := iteratorID(args[0])
			return &iterPrevOp{
				m:    m,
				iter: iter,
			}
		},
		operands: []operandType{
			operandIterator,
		},
		weight: 100,
	},
	{
		// Note that this is not an MVCC* operation; unlike MVCC{Put,Get,Scan}, etc,
		// it does not respect transactions. This often yields interesting
		// behavior.
		name: "delete_range",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			key := m.keyGenerator.parse(args[0]).Key
			endKey := m.keyGenerator.parse(args[1]).Key
			if endKey.Compare(key) < 0 {
				key, endKey = endKey, key
			} else if endKey.Equal(key) {
				// Range tombstones where start = end can exhibit different behavior on
				// different engines; rocks treats it as a point delete, while pebble
				// treats it as a nonexistent tombstone. For the purposes of this test,
				// standardize behavior.
				endKey = endKey.Next()
			}

			return &clearRangeOp{
				m:      m,
				key:    key,
				endKey: endKey,
			}
		},
		operands: []operandType{
			operandMVCCKey,
			operandMVCCKey,
		},
		weight: 20,
	},
	{
		name: "compact",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			key := m.keyGenerator.parse(args[0]).Key
			endKey := m.keyGenerator.parse(args[1]).Key
			if endKey.Compare(key) < 0 {
				key, endKey = endKey, key
			}
			return &compactOp{
				m:      m,
				key:    key,
				endKey: endKey,
			}
		},
		operands: []operandType{
			operandMVCCKey,
			operandMVCCKey,
		},
		weight: 10,
	},
	{
		name: "ingest",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			var keys []storage.MVCCKey
			for _, arg := range args {
				key := m.keyGenerator.parse(arg)
				// Don't put anything at the 0 timestamp; the MVCC code expects
				// MVCCMetadata at those values.
				if key.Timestamp == (hlc.Timestamp{}) {
					key.Timestamp = key.Timestamp.Next()
				}
				keys = append(keys, key)
			}
			// SST Writer expects keys in sorted order, so sort them first.
			sort.Slice(keys, func(i, j int) bool {
				return keys[i].Less(keys[j])
			})

			return &ingestOp{
				m:    m,
				keys: keys,
			}
		},
		operands: []operandType{
			operandMVCCKey,
			operandMVCCKey,
			operandMVCCKey,
			operandMVCCKey,
			operandMVCCKey,
		},
		weight: 10,
	},
	{
		name: "restart",
		generate: func(ctx context.Context, m *metaTestRunner, args ...string) mvccOp {
			// Simulate a restart in all generators tracking open objects. This should
			// ensure all IDs generated beyond this point will not reference anything
			// created before a restart.
			m.closeGenerators()
			return &restartOp{
				m: m,
			}
		},
		operands: nil,
		weight:   4,
	},
}
