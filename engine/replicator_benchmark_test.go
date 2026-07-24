// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"bytes"
	"encoding/gob"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

const replicationBenchmarkBatchSize = 1000

type replicationCall struct {
	objID  string
	method string
	args   any
}

type countingConnector struct {
	calls atomic.Uint64
	err   error
}

func (c *countingConnector) Call(_ *context.Context, _ string, _, _ any) error {
	c.calls.Add(1)
	return c.err
}

func newReplicationCall(objID string) replicationCall {
	return replicationCall{
		objID:  objID,
		method: utils.ReplicatorSv1SetIndexes,
		args: &utils.SetIndexesArg{
			IdxItmType: utils.CacheAttributeProfiles,
			TntCtx:     objID,
			Indexes: map[string]utils.StringSet{
				"tenant": {"cgrates.org": {}},
			},
		},
	}
}

func uniqueReplicationCalls(size int) []replicationCall {
	calls := make([]replicationCall, size)
	for i := range calls {
		calls[i] = newReplicationCall("cgrates.org:benchmark-" + strconv.Itoa(i))
	}
	return calls
}

func sameObjectSetCalls(size int) []replicationCall {
	calls := make([]replicationCall, size)
	for i := range calls {
		calls[i] = newReplicationCall("cgrates.org:benchmark")
	}
	return calls
}

func sameObjectSetRemoveCalls(size int) []replicationCall {
	objID := "cgrates.org:benchmark"
	calls := make([]replicationCall, size)
	for i := range calls {
		if i%2 == 0 {
			calls[i] = newReplicationCall(objID)
		} else {
			calls[i] = replicationCall{
				objID:  objID,
				method: utils.ReplicatorSv1RemoveIndexes,
				args: &utils.GetIndexesArg{
					IdxItmType: utils.CacheAttributeProfiles,
					TntCtx:     objID,
					IdxKeys:    []string{"tenant"},
				},
			}
		}
	}
	return calls
}

func newReplicationTask() *ReplicationTask {
	call := newReplicationCall("cgrates.org:benchmark")
	return &ReplicationTask{
		ConnIDs: []string{"replicator-test"},
		ObjType: utils.CacheInstanceToPrefix[utils.CacheAttributeProfiles],
		ObjID:   call.objID,
		Method:  call.method,
		Args:    call.args,
	}
}

func BenchmarkReplicatorImmediate(b *testing.B) {
	item := &config.ItemOpt{Replicate: true}
	call := newReplicationCall("cgrates.org:benchmark")
	objType := utils.CacheInstanceToPrefix[utils.CacheAttributeProfiles]

	b.Run("FailedDirDisabled", func(b *testing.B) {
		r := newTestReplicator(b, 0, "", &mockConnector{})
		for b.Loop() {
			if err := r.replicate(objType, call.objID, call.method, call.args, item); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("FailedDirEnabled", func(b *testing.B) {
		r := newTestReplicator(b, 0, b.TempDir(), &mockConnector{})
		for b.Loop() {
			if err := r.replicate(objType, call.objID, call.method, call.args, item); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkReplicatorImmediateParallel(b *testing.B) {
	item := &config.ItemOpt{Replicate: true}
	call := newReplicationCall("cgrates.org:benchmark")
	objType := utils.CacheInstanceToPrefix[utils.CacheAttributeProfiles]

	b.Run("SameObject", func(b *testing.B) {
		r := newTestReplicator(b, 0, "", &mockConnector{})
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if err := r.replicate(objType, call.objID, call.method, call.args, item); err != nil {
					b.Error(err)
					return
				}
			}
		})
	})
	b.Run("DifferentObjects", func(b *testing.B) {
		r := newTestReplicator(b, 0, "", &mockConnector{})
		calls := uniqueReplicationCalls(runtime.GOMAXPROCS(0))
		var workerID atomic.Uint64
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			call := calls[workerID.Add(1)-1]
			for pb.Next() {
				if err := r.replicate(objType, call.objID, call.method, call.args, item); err != nil {
					b.Error(err)
					return
				}
			}
		})
	})
}

func BenchmarkReplicatorIntervalQueue(b *testing.B) {
	item := &config.ItemOpt{Replicate: true}
	benchmarks := []struct {
		name  string
		calls []replicationCall
	}{
		{name: "UniqueObjects/1000", calls: uniqueReplicationCalls(replicationBenchmarkBatchSize)},
		{name: "SameObjectSet/1000", calls: sameObjectSetCalls(replicationBenchmarkBatchSize)},
		{name: "SameObjectSetRemove/1000", calls: sameObjectSetRemoveCalls(replicationBenchmarkBatchSize)},
	}
	for _, benchmark := range benchmarks {
		b.Run(benchmark.name, func(b *testing.B) {
			r := newTestReplicator(b, time.Hour, "", &mockConnector{})
			objType := utils.CacheInstanceToPrefix[utils.CacheAttributeProfiles]
			next := 0
			for b.Loop() {
				call := benchmark.calls[next]
				if err := r.replicate(objType, call.objID, call.method, call.args, item); err != nil {
					b.Fatal(err)
				}
				next++
				if next == len(benchmark.calls) {
					// Draining outside the timer keeps pending memory bounded without measuring RPC calls.
					b.StopTimer()
					r.flush()
					next = 0
					b.StartTimer()
				}
			}
			r.flush()
		})
	}
}

func BenchmarkReplicatorIntervalBatch(b *testing.B) {
	item := &config.ItemOpt{Replicate: true}
	benchmarks := []struct {
		name  string
		calls []replicationCall
	}{
		{name: "UniqueObjects/1000", calls: uniqueReplicationCalls(replicationBenchmarkBatchSize)},
		{name: "SameObjectSet/1000", calls: sameObjectSetCalls(replicationBenchmarkBatchSize)},
		{name: "SameObjectSetRemove/1000", calls: sameObjectSetRemoveCalls(replicationBenchmarkBatchSize)},
	}
	for _, benchmark := range benchmarks {
		b.Run(benchmark.name, func(b *testing.B) {
			connector := &countingConnector{}
			r := newTestReplicator(b, time.Hour, "", connector)
			for b.Loop() {
				objType := utils.CacheInstanceToPrefix[utils.CacheAttributeProfiles]
				for _, call := range benchmark.calls {
					if err := r.replicate(objType, call.objID, call.method, call.args, item); err != nil {
						b.Fatal(err)
					}
				}
				r.flush()
			}
			b.ReportMetric(float64(connector.calls.Load())/float64(b.N), "rpc/op")
		})
	}
}

func BenchmarkReplicatorFailedWrite(b *testing.B) {
	wantErr := errors.New("replication failed")
	r := newTestReplicator(b, 0, b.TempDir(),
		&mockConnector{err: wantErr})
	item := &config.ItemOpt{Replicate: true}
	call := newReplicationCall("cgrates.org:benchmark")
	objType := utils.CacheInstanceToPrefix[utils.CacheAttributeProfiles]
	for b.Loop() {
		if err := r.replicate(objType, call.objID, call.method, call.args, item); !errors.Is(err, wantErr) {
			b.Fatalf("replicate returned %v, want %v", err, wantErr)
		}
	}
}

func BenchmarkIndexReplication(b *testing.B) {
	idxItmType := utils.CacheAttributeFilterIndexes
	objType := utils.CacheInstanceToPrefix[idxItmType]
	tntCtx := "cgrates.org:*cdrs"
	for _, fields := range []int{1, 16, 256} {
		indexes := make(map[string]utils.StringSet, fields)
		for i := 0; i < fields; i++ {
			indexes["field-"+strconv.Itoa(i)] = utils.StringSet{"value": {}}
		}
		name := "Fields" + strconv.Itoa(fields)

		b.Run("Immediate/"+name, func(b *testing.B) {
			connector := &countingConnector{}
			dm := newIndexDataManager(b, 0, "", connector)
			b.ReportAllocs()
			for b.Loop() {
				if err := dm.SetIndexes(idxItmType, tntCtx, indexes, true, ""); err != nil {
					b.Fatal(err)
				}
			}
			b.ReportMetric(float64(connector.calls.Load())/float64(b.N), "rpc/op")
		})

		b.Run("IntervalAdmission/"+name, func(b *testing.B) {
			connector := &countingConnector{}
			dm := newIndexDataManager(b, time.Hour, "", connector)
			b.ReportAllocs()
			for b.Loop() {
				if err := dm.SetIndexes(idxItmType, tntCtx, indexes, true, ""); err != nil {
					b.Fatal(err)
				}
			}
			b.ReportMetric(float64(connector.calls.Load())/float64(b.N), "rpc/op")
		})

		b.Run("IntervalFlush/"+name, func(b *testing.B) {
			connector := &countingConnector{}
			dm := newIndexDataManager(b, time.Hour, "", connector)
			b.ReportAllocs()
			for b.Loop() {
				b.StopTimer()
				if err := dm.SetIndexes(idxItmType, tntCtx, indexes, true, ""); err != nil {
					b.Fatal(err)
				}
				b.StartTimer()
				dm.replicator.flush()
			}
			b.ReportMetric(float64(connector.calls.Load())/float64(b.N), "rpc/op")
		})

		b.Run("FailedRetry/"+name, func(b *testing.B) {
			wantErr := errors.New("replication failed")
			connector := &countingConnector{err: wantErr}
			r := newTestReplicator(b, 0, b.TempDir(), connector)
			args := &utils.SetIndexesArg{
				IdxItmType: idxItmType,
				TntCtx:     tntCtx,
				Indexes:    indexes,
			}
			b.ReportAllocs()
			for b.Loop() {
				if err := r.replicateIndexes(objType, tntCtx, args); !errors.Is(err, wantErr) {
					b.Fatalf("replicateIndexes returned %v, want %v", err, wantErr)
				}
			}
			b.ReportMetric(float64(connector.calls.Load())/float64(b.N), "rpc/op")
		})
	}
}

func BenchmarkReplicationTaskWriteToFile(b *testing.B) {
	_ = setupReplicator(b, "", &mockConnector{})
	path := filepath.Join(b.TempDir(), "replication.gob")
	task := newReplicationTask()
	for b.Loop() {
		if err := task.WriteToFile(path); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNewReplicationTaskFromFile(b *testing.B) {
	_ = setupReplicator(b, "", &mockConnector{})
	path := filepath.Join(b.TempDir(), "replication.gob")
	var taskBytes bytes.Buffer
	if err := gob.NewEncoder(&taskBytes).Encode(newReplicationTask()); err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		b.StopTimer()
		if err := os.WriteFile(path, taskBytes.Bytes(), 0600); err != nil {
			b.Fatal(err)
		}
		b.StartTimer()
		if _, err := NewReplicationTaskFromFile(path); err != nil {
			b.Fatal(err)
		}
	}
}
