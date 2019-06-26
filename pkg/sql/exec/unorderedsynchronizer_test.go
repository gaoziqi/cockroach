// Copyright 2019 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package exec

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/cockroachdb/cockroach/pkg/sql/exec/coldata"
	"github.com/cockroachdb/cockroach/pkg/sql/exec/types"
	"github.com/cockroachdb/cockroach/pkg/testutils"
	"github.com/cockroachdb/cockroach/pkg/util/leaktest"
	"github.com/cockroachdb/cockroach/pkg/util/randutil"
	"github.com/stretchr/testify/require"
)

func TestUnorderedSynchronizer(t *testing.T) {
	defer leaktest.AfterTest(t)()

	const (
		maxInputs  = 16
		maxBatches = 16
	)

	var (
		rng, _     = randutil.NewPseudoRand()
		typs       = []types.T{types.Int64}
		numInputs  = rng.Intn(maxInputs) + 1
		numBatches = rng.Intn(maxBatches) + 1
	)

	inputs := make([]Operator, numInputs)
	for i := range inputs {
		source := NewRepeatableBatchSource(RandomBatch(rng, typs, coldata.BatchSize, rng.Float64()))
		source.ResetBatchesToReturn(numBatches)
		inputs[i] = source
	}

	var wg sync.WaitGroup
	s := NewUnorderedSynchronizer(inputs, typs, &wg)

	ctx, cancelFn := context.WithCancel(context.Background())
	var cancel bool
	if rng.Float64() < 0.5 {
		cancel = true
	}
	if cancel {
		wg.Add(1)
		sleepTime := time.Duration(rng.Intn(500)) * time.Microsecond
		go func() {
			time.Sleep(sleepTime)
			cancelFn()
			wg.Done()
		}()
	} else {
		// Appease the linter complaining about context leaks.
		defer cancelFn()
	}

	batchesReturned := 0
	for {
		var b coldata.Batch
		if err := CatchVectorizedRuntimeError(func() { b = s.Next(ctx) }); err != nil {
			if cancel {
				require.True(t, testutils.IsError(err, "context canceled"), err)
				break
			} else {
				t.Fatal(err)
			}
		}
		if b.Length() == 0 {
			break
		}
		batchesReturned++
	}
	if !cancel {
		require.Equal(t, numInputs*numBatches, batchesReturned)
	}
	wg.Wait()
}

func BenchmarkUnorderedSynchronizer(b *testing.B) {
	const numInputs = 6

	typs := []types.T{types.Int64}
	inputs := make([]Operator, numInputs)
	for i := range inputs {
		batch := coldata.NewMemBatchWithSize(typs, coldata.BatchSize)
		batch.SetLength(coldata.BatchSize)
		inputs[i] = NewRepeatableBatchSource(batch)
	}
	var wg sync.WaitGroup
	ctx, cancelFn := context.WithCancel(context.Background())
	s := NewUnorderedSynchronizer(inputs, typs, &wg)
	b.SetBytes(8 * coldata.BatchSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Next(ctx)
	}
	b.StopTimer()
	cancelFn()
	wg.Wait()
}
