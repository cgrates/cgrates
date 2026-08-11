// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestReplicationKey(t *testing.T) {
	pairs := []struct {
		set    string
		remove string
	}{
		{set: utils.ReplicatorSv1SetAccount, remove: utils.ReplicatorSv1RemoveAccount},
		{set: utils.ReplicatorSv1SetThresholdProfile, remove: utils.ReplicatorSv1RemoveThresholdProfile},
		{set: utils.ReplicatorSv1SetThreshold, remove: utils.ReplicatorSv1RemoveThreshold},
		{set: utils.ReplicatorSv1SetStatQueueProfile, remove: utils.ReplicatorSv1RemoveStatQueueProfile},
		{set: utils.ReplicatorSv1SetStatQueue, remove: utils.ReplicatorSv1RemoveStatQueue},
		{set: utils.ReplicatorSv1SetFilter, remove: utils.ReplicatorSv1RemoveFilter},
		{set: utils.ReplicatorSv1SetRankingProfile, remove: utils.ReplicatorSv1RemoveRankingProfile},
		{set: utils.ReplicatorSv1SetTrendProfile, remove: utils.ReplicatorSv1RemoveTrendProfile},
		{set: utils.ReplicatorSv1SetTrend, remove: utils.ReplicatorSv1RemoveTrend},
		{set: utils.ReplicatorSv1SetResourceProfile, remove: utils.ReplicatorSv1RemoveResourceProfile},
		{set: utils.ReplicatorSv1SetResource, remove: utils.ReplicatorSv1RemoveResource},
		{set: utils.ReplicatorSv1SetIPProfile, remove: utils.ReplicatorSv1RemoveIPProfile},
		{set: utils.ReplicatorSv1SetIPAllocations, remove: utils.ReplicatorSv1RemoveIPAllocations},
		{set: utils.ReplicatorSv1SetActionProfile, remove: utils.ReplicatorSv1RemoveActionProfile},
		{set: utils.ReplicatorSv1SetRouteProfile, remove: utils.ReplicatorSv1RemoveRouteProfile},
		{set: utils.ReplicatorSv1SetAttributeProfile, remove: utils.ReplicatorSv1RemoveAttributeProfile},
		{set: utils.ReplicatorSv1SetChargerProfile, remove: utils.ReplicatorSv1RemoveChargerProfile},
	}
	want := "type_id"
	for _, pair := range pairs {
		for _, method := range []string{pair.set, pair.remove} {
			if got := replicationKey("type_", "id", method); got != want {
				t.Errorf("replicationKey(%q) = %q, want %q", method, got, want)
			}
		}
	}

	separate := []string{
		utils.ReplicatorSv1SetRanking,
		utils.ReplicatorSv1RemoveRanking,
		utils.ReplicatorSv1SetRateProfile,
		utils.ReplicatorSv1RemoveRateProfile,
		utils.ReplicatorSv1SetIndexes,
		utils.ReplicatorSv1RemoveIndexes,
		utils.ReplicatorSv1SetLoadIDs,
	}
	for _, method := range separate {
		_, operation, _ := strings.Cut(method, utils.NestingSep)
		want := operation + "_type_id"
		if got := replicationKey("type_", "id", method); got != want {
			t.Errorf("replicationKey(%q) = %q, want %q", method, got, want)
		}
	}
}

func TestReplicatorFailedTaskReplacement(t *testing.T) {
	set := &utils.AttributeProfileWithAPIOpts{AttributeProfile: &utils.AttributeProfile{Tenant: "cgrates.org", ID: "destination"}}
	remove := &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "destination"}}
	sequences := []struct {
		name    string
		methods [2]string
		args    [2]any
	}{
		{
			name:    "SetRemove",
			methods: [2]string{utils.ReplicatorSv1SetAttributeProfile, utils.ReplicatorSv1RemoveAttributeProfile},
			args:    [2]any{set, remove},
		},
		{
			name:    "RemoveSet",
			methods: [2]string{utils.ReplicatorSv1RemoveAttributeProfile, utils.ReplicatorSv1SetAttributeProfile},
			args:    [2]any{remove, set},
		},
	}
	modes := []struct {
		name     string
		interval time.Duration
	}{
		{name: "Immediate"},
		{name: "Interval", interval: time.Hour},
	}
	for _, mode := range modes {
		for _, sequence := range sequences {
			t.Run(mode.name+"/"+sequence.name, func(t *testing.T) {
				failedDir := t.TempDir()
				failedErr := errors.New("replication failed")
				connector := &mockConnector{err: failedErr}
				r := newTestReplicator(t, mode.interval, failedDir, connector)
				item := &config.ItemOpts{Replicate: true}
				attempt := func(method string, args any) {
					r.replicate(context.Background(), utils.AttributeProfilePrefix, "cgrates.org:destination", method, args, item)
					if mode.interval > 0 {
						r.flush()
					}
				}

				for i, method := range sequence.methods {
					attempt(method, sequence.args[i])
				}
				entries, err := os.ReadDir(failedDir)
				if err != nil {
					t.Fatal(err)
				}
				if len(entries) != 1 {
					t.Fatalf("failed directory has %d files, want 1", len(entries))
				}
				wantName := utils.AttributeProfilePrefix + "cgrates.org:destination" + utils.GOBSuffix
				if entries[0].Name() != wantName {
					t.Errorf("failed filename = %q, want %q", entries[0].Name(), wantName)
				}
				taskBytes, err := os.ReadFile(filepath.Join(failedDir, entries[0].Name()))
				if err != nil {
					t.Fatal(err)
				}
				var task *ReplicationTask
				if err := gob.NewDecoder(bytes.NewReader(taskBytes)).Decode(&task); err != nil {
					t.Fatal(err)
				}
				if task.Method != sequence.methods[1] {
					t.Errorf("method = %q, want %q", task.Method, sequence.methods[1])
				}
				if !reflect.DeepEqual(task.Args, sequence.args[1]) {
					t.Errorf("args = %#v, want %#v", task.Args, sequence.args[1])
				}

				connector.err = nil
				attempt(sequence.methods[0], sequence.args[0])
				entries, err = os.ReadDir(failedDir)
				if err != nil {
					t.Fatal(err)
				}
				if len(entries) != 0 {
					t.Errorf("failed directory has %d files, want 0", len(entries))
				}
			})
		}
	}
}

func TestReplicatorIntervalFinalOperation(t *testing.T) {
	set1 := &utils.AttributeProfileWithAPIOpts{AttributeProfile: &utils.AttributeProfile{Tenant: "cgrates.org", ID: "destination"}}
	set2 := &utils.AttributeProfileWithAPIOpts{AttributeProfile: &utils.AttributeProfile{Tenant: "cgrates.org", ID: "destination"}}
	remove := &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "destination"}}
	tests := []struct {
		name       string
		methods    [2]string
		args       [2]any
		wantMethod string
		wantArgs   any
	}{
		{
			name:       "SetSet",
			methods:    [2]string{utils.ReplicatorSv1SetAttributeProfile, utils.ReplicatorSv1SetAttributeProfile},
			args:       [2]any{set1, set2},
			wantMethod: utils.ReplicatorSv1SetAttributeProfile,
			wantArgs:   set2,
		},
		{
			name:       "SetRemove",
			methods:    [2]string{utils.ReplicatorSv1SetAttributeProfile, utils.ReplicatorSv1RemoveAttributeProfile},
			args:       [2]any{set1, remove},
			wantMethod: utils.ReplicatorSv1RemoveAttributeProfile,
			wantArgs:   remove,
		},
		{
			name:       "RemoveSet",
			methods:    [2]string{utils.ReplicatorSv1RemoveAttributeProfile, utils.ReplicatorSv1SetAttributeProfile},
			args:       [2]any{remove, set2},
			wantMethod: utils.ReplicatorSv1SetAttributeProfile,
			wantArgs:   set2,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := newTestReplicator(t, time.Second, "", &mockConnector{})
			item := &config.ItemOpts{Replicate: true}
			for i, method := range test.methods {
				r.replicate(context.Background(), utils.AttributeProfilePrefix, "cgrates.org:destination", method, test.args[i], item)
			}
			if len(r.pending) != 1 {
				t.Fatalf("pending has %d items, want 1", len(r.pending))
			}
			got, ok := r.pending[utils.AttributeProfilePrefix+"cgrates.org:destination"]
			if !ok {
				t.Fatal("pending object key not found")
			}
			if got.method != test.wantMethod {
				t.Errorf("method = %q, want %q", got.method, test.wantMethod)
			}
			if got.args != test.wantArgs {
				t.Errorf("args = %#v, want %#v", got.args, test.wantArgs)
			}
		})
	}
}

func TestFailedReplicationFileName(t *testing.T) {
	ordinaryKey := replicationKey(utils.AttributeProfilePrefix, "destination",
		utils.ReplicatorSv1SetAttributeProfile)
	if got := failedReplicationFileName(ordinaryKey); got != ordinaryKey {
		t.Fatalf("failed filename = %q, want unchanged key %q", got, ordinaryKey)
	}
	patchKey := indexPatchKey("parent")
	if got := failedReplicationFileName(patchKey); got == patchKey || !strings.HasPrefix(got, "index_") {
		t.Fatalf("failed filename = %q, want hashed index name", got)
	}
}

type indexConnector struct {
	err   error
	calls []*utils.SetIndexesArg
}

func (c *indexConnector) Call(_ *context.Context, method string, args, _ any) error {
	if method != utils.ReplicatorSv1SetIndexes {
		return c.err
	}
	c.calls = append(c.calls, args.(*utils.SetIndexesArg))
	return c.err
}

func newIndexDataManager(tb testing.TB, interval time.Duration, failedDir string,
	connector birpc.ClientConnector) *DataManager {
	tb.Helper()
	cm := setupReplicator(tb, failedDir, connector)
	cfg := cm.cfg
	cfg.DbCfg().Items[utils.CacheAttributeFilterIndexes].Replicate = true
	db, err := NewInternalDB(nil, cfg.DbCfg().Items)
	if err != nil {
		tb.Fatal(err)
	}
	dbCM := NewDBConnManager(map[string]DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dm := NewDataManager(dbCM, cfg, cm, NewLocker(cfg))
	dm.SetCache(cm.cache)
	testReplicator(dm).interval = interval
	return dm
}

func readFailedIndexTask(t *testing.T, dir, key string) (*ReplicationTask, *utils.SetIndexesArg) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, failedReplicationFileName(key)+utils.GOBSuffix))
	if err != nil {
		t.Fatal(err)
	}
	var task *ReplicationTask
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&task); err != nil {
		t.Fatal(err)
	}
	args, ok := task.Args.(*utils.SetIndexesArg)
	if !ok || task.Method != utils.ReplicatorSv1SetIndexes || task.ObjID != args.TntCtx {
		t.Fatalf("unexpected failed task %#v", task)
	}
	return task, args
}

func TestReplicatorIndexFields(t *testing.T) {
	idxItmType := utils.CacheAttributeFilterIndexes
	tntCtx := "cgrates.org"
	fieldA := "A"
	fieldB := "B"
	fieldC := "C"
	fieldD := "D"
	valueOne := utils.StringSet{"RL1": {}}
	valueTwo := utils.StringSet{"RL2": {}}

	type operation struct {
		indexes map[string]utils.StringSet
		remove  []string
		clear   bool
	}
	for _, test := range []struct {
		name       string
		operations []operation
		wantClear  bool
		want       map[string]utils.StringSet
	}{
		{
			name:       "Set A and Set B",
			operations: []operation{{indexes: map[string]utils.StringSet{fieldA: valueOne, fieldB: valueTwo}}},
			want:       map[string]utils.StringSet{fieldA: valueOne, fieldB: valueTwo},
		},
		{
			name: "repeated Set A",
			operations: []operation{
				{indexes: map[string]utils.StringSet{fieldA: valueOne}},
				{indexes: map[string]utils.StringSet{fieldA: valueTwo}},
			},
			want: map[string]utils.StringSet{fieldA: valueTwo},
		},
		{
			name: "mixed fields",
			operations: []operation{
				{indexes: map[string]utils.StringSet{fieldA: valueOne}},
				{remove: []string{fieldB}},
				{indexes: map[string]utils.StringSet{fieldC: valueOne}},
				{remove: []string{fieldC}},
				{indexes: map[string]utils.StringSet{fieldB: valueTwo}},
				{remove: []string{fieldD}},
			},
			want: map[string]utils.StringSet{
				fieldA: valueOne,
				fieldB: valueTwo,
				fieldC: nil,
				fieldD: nil,
			},
		},
		{
			name:       "Set A then remove all",
			operations: []operation{{indexes: map[string]utils.StringSet{fieldA: valueOne}}, {clear: true}},
			wantClear:  true,
			want:       map[string]utils.StringSet{},
		},
		{
			name: "remove all then Set A then Remove B",
			operations: []operation{
				{clear: true},
				{indexes: map[string]utils.StringSet{fieldA: valueOne}},
				{remove: []string{fieldB}},
			},
			wantClear: true,
			want: map[string]utils.StringSet{
				fieldA: valueOne,
				fieldB: nil,
			},
		},
	} {
		t.Run("interval/"+test.name, func(t *testing.T) {
			connector := &indexConnector{}
			dm := newIndexDataManager(t, time.Hour, "", connector)
			for _, operation := range test.operations {
				var err error
				switch {
				case operation.clear:
					err = dm.RemoveIndexes(context.Background(), idxItmType, tntCtx)
				case operation.remove != nil:
					err = dm.RemoveIndexes(context.Background(), idxItmType, tntCtx, operation.remove...)
				default:
					err = dm.SetIndexes(context.Background(), idxItmType, tntCtx, operation.indexes, true, "")
				}
				if err != nil {
					t.Fatal(err)
				}
			}
			if len(testReplicator(dm).pending) != 1 {
				t.Fatalf("pending has %d patches, want 1", len(testReplicator(dm).pending))
			}
			for _, pending := range testReplicator(dm).pending {
				if pending.objID != tntCtx {
					t.Fatalf("pending objID = %q, want %q", pending.objID, tntCtx)
				}
			}
			testReplicator(dm).flush()
			if len(connector.calls) != 1 {
				t.Fatalf("connector has %d calls, want 1", len(connector.calls))
			}
			got := connector.calls[0]
			if got.Clear != test.wantClear || !reflect.DeepEqual(got.Indexes, test.want) {
				t.Fatalf("replicated patch = {Clear:%t Indexes:%#v}, want {Clear:%t Indexes:%#v}",
					got.Clear, got.Indexes, test.wantClear, test.want)
			}
		})
	}
}

func TestReplicatorIndexesImmediate(t *testing.T) {
	idxItmType := utils.CacheAttributeFilterIndexes
	tntCtx := "cgrates.org"
	fieldA := "A"
	fieldB := "B"

	connector := &indexConnector{}
	dm := newIndexDataManager(t, 0, "", connector)
	calls := []func() error{
		func() error {
			return dm.SetIndexes(context.Background(), idxItmType, tntCtx, map[string]utils.StringSet{
				fieldA: {"RL1": {}},
				fieldB: {"RL2": {}},
			}, true, "")
		},
		func() error {
			return dm.RemoveIndexes(context.Background(), idxItmType, tntCtx, fieldA, fieldB)
		},
		func() error { return dm.RemoveIndexes(context.Background(), idxItmType, tntCtx) },
	}
	for i, call := range calls {
		before := len(connector.calls)
		if err := call(); err != nil {
			t.Fatal(err)
		}
		if got := len(connector.calls) - before; got != 1 {
			t.Fatalf("call %d sent %d RPCs, want 1", i, got)
		}
	}
	if got := connector.calls[1].Indexes; len(got) != 2 || len(got[fieldA]) != 0 || len(got[fieldB]) != 0 {
		t.Fatalf("selected removal patch = %#v", got)
	}
	if !connector.calls[2].Clear {
		t.Fatal("remove all patch does not clear context")
	}
}

func TestReplicatorIndexesImmediateFailure(t *testing.T) {
	idxItmType := utils.CacheAttributeFilterIndexes
	tntCtx := "cgrates.org"

	failedDir := t.TempDir()
	connector := &indexConnector{err: errors.New("replication failed")}
	dm := newIndexDataManager(t, 0, failedDir, connector)
	indexes := map[string]utils.StringSet{
		"A": {"RL1": {}},
		"B": {"RL2": {}},
	}
	if err := dm.SetIndexes(context.Background(), idxItmType, tntCtx, indexes, true, ""); err != nil {
		t.Fatalf("SetIndexes returned replication error: %v", err)
	}
	if len(connector.calls) != 1 {
		t.Fatalf("connector has %d calls, want 1", len(connector.calls))
	}
	stored, err := dm.dbConns.dbs[utils.MetaDefault].GetIndexesDrv(context.Background(), idxItmType, tntCtx, utils.NonTransactional)
	if err != nil || !reflect.DeepEqual(stored, indexes) {
		t.Fatalf("stored indexes = %#v, %v, want %#v", stored, err, indexes)
	}
	key := indexPatchKey(replicationKey(
		utils.CacheInstanceToPrefix[idxItmType], tntCtx, utils.ReplicatorSv1SetIndexes))
	task, args := readFailedIndexTask(t, failedDir, key)
	if task.ObjID != tntCtx || args.TntCtx != tntCtx || !reflect.DeepEqual(args.Indexes, indexes) {
		t.Fatalf("failed task = %#v", task)
	}
}

func TestReplicatorFailedIndexPatches(t *testing.T) {
	idxItmType := utils.CacheAttributeFilterIndexes
	tntCtx := "cgrates.org"
	fieldA := "A"
	fieldB := "B"
	valueOne := utils.StringSet{"RL1": {}}
	valueTwo := utils.StringSet{"RL2": {}}

	t.Run("merges newer fields", func(t *testing.T) {
		failedDir := t.TempDir()
		connector := &indexConnector{err: errors.New("replication failed")}
		dm := newIndexDataManager(t, time.Hour, failedDir, connector)
		if err := dm.RemoveIndexes(context.Background(), idxItmType, tntCtx); err != nil {
			t.Fatal(err)
		}
		if err := dm.SetIndexes(context.Background(), idxItmType, tntCtx,
			map[string]utils.StringSet{fieldA: valueOne}, true, ""); err != nil {
			t.Fatal(err)
		}
		testReplicator(dm).flush()
		entries, err := os.ReadDir(failedDir)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 1 {
			t.Fatalf("failed directory has %d patches, want 1", len(entries))
		}
		key := indexPatchKey(replicationKey(
			utils.CacheInstanceToPrefix[idxItmType], tntCtx, utils.ReplicatorSv1SetIndexes))
		_, failed := readFailedIndexTask(t, failedDir, key)
		if !failed.Clear || !reflect.DeepEqual(failed.Indexes,
			map[string]utils.StringSet{fieldA: valueOne}) {
			t.Fatalf("failed patch = %#v", failed)
		}

		connector.err = nil
		if err := dm.SetIndexes(context.Background(), idxItmType, tntCtx,
			map[string]utils.StringSet{fieldB: valueTwo}, true, ""); err != nil {
			t.Fatal(err)
		}
		testReplicator(dm).flush()
		got := connector.calls[len(connector.calls)-1]
		want := map[string]utils.StringSet{fieldA: valueOne, fieldB: valueTwo}
		if !got.Clear || !reflect.DeepEqual(got.Indexes, want) {
			t.Fatalf("retried patch = %#v, want Clear with %#v", got, want)
		}
		failedPath := filepath.Join(failedDir, failedReplicationFileName(key)+utils.GOBSuffix)
		if _, err := os.Stat(failedPath); !os.IsNotExist(err) {
			t.Fatalf("successful patch failed file still exists: %v", err)
		}
	})

	t.Run("newer clear discards failed fields", func(t *testing.T) {
		failedDir := t.TempDir()
		connector := &indexConnector{err: errors.New("replication failed")}
		dm := newIndexDataManager(t, time.Hour, failedDir, connector)
		if err := dm.SetIndexes(context.Background(), idxItmType, tntCtx,
			map[string]utils.StringSet{fieldA: valueOne}, true, ""); err != nil {
			t.Fatal(err)
		}
		testReplicator(dm).flush()
		connector.err = nil
		if err := dm.RemoveIndexes(context.Background(), idxItmType, tntCtx); err != nil {
			t.Fatal(err)
		}
		if err := dm.SetIndexes(context.Background(), idxItmType, tntCtx,
			map[string]utils.StringSet{fieldB: valueTwo}, true, ""); err != nil {
			t.Fatal(err)
		}
		testReplicator(dm).flush()
		got := connector.calls[len(connector.calls)-1]
		want := map[string]utils.StringSet{fieldB: valueTwo}
		if !got.Clear || !reflect.DeepEqual(got.Indexes, want) {
			t.Fatalf("retried patch = %#v, want Clear with %#v", got, want)
		}
	})
}

func TestReplicatorIndexTransactions(t *testing.T) {
	idxItmType := utils.CacheAttributeFilterIndexes
	objType := utils.CacheInstanceToPrefix[idxItmType]
	tntCtx := "cgrates.org"
	fieldA := "A"
	valueOne := utils.StringSet{"RL1": {}}

	t.Run("interval", func(t *testing.T) {
		connector := &indexConnector{}
		dm := newIndexDataManager(t, time.Hour, "", connector)
		indexes := map[string]utils.StringSet{fieldA: valueOne}
		if err := dm.SetIndexes(context.Background(), idxItmType, tntCtx, indexes, false, "transaction"); err != nil {
			t.Fatal(err)
		}
		key := replicationKey(objType, tntCtx, utils.ReplicatorSv1SetIndexes)
		if _, has := testReplicator(dm).pending[key]; !has {
			t.Fatal("transaction call did not use the generic pending path")
		}
		if _, has := testReplicator(dm).pending[indexPatchKey(key)]; has {
			t.Fatal("transaction call entered the grouped index path")
		}
		testReplicator(dm).flush()
		if len(connector.calls) != 1 || !reflect.DeepEqual(connector.calls[0].Indexes, indexes) {
			t.Fatalf("connector calls = %#v, want one transaction call with %#v", connector.calls, indexes)
		}
	})

	t.Run("immediate transaction failure keeps API success", func(t *testing.T) {
		failedDir := t.TempDir()
		connector := &indexConnector{err: errors.New("replication failed")}
		dm := newIndexDataManager(t, 0, failedDir, connector)
		indexes := map[string]utils.StringSet{fieldA: valueOne}
		if err := dm.SetIndexes(context.Background(), idxItmType, tntCtx, indexes, false, "transaction"); err != nil {
			t.Fatalf("SetIndexes returned replication error: %v", err)
		}
		key := replicationKey(objType, tntCtx, utils.ReplicatorSv1SetIndexes)
		_, args := readFailedIndexTask(t, failedDir, key)
		if !reflect.DeepEqual(args.Indexes, indexes) {
			t.Fatalf("failed transaction indexes = %#v, want %#v", args.Indexes, indexes)
		}
	})
}

func TestReplicatorIndexesDisabled(t *testing.T) {
	idxItmType := utils.CacheAttributeFilterIndexes
	tntCtx := "cgrates.org"
	fieldA := "A"

	connector := &indexConnector{}
	dm := newIndexDataManager(t, time.Hour, "", connector)
	dm.cfg.DbCfg().Items[idxItmType].Replicate = false
	if err := dm.SetIndexes(context.Background(), idxItmType, tntCtx,
		map[string]utils.StringSet{fieldA: {"RL1": {}}}, true, ""); err != nil {
		t.Fatal(err)
	}
	if err := dm.RemoveIndexes(context.Background(), idxItmType, tntCtx, fieldA); err != nil {
		t.Fatal(err)
	}
	testReplicator(dm).flush()
	if len(connector.calls) != 0 || len(testReplicator(dm).pending) != 0 {
		t.Fatalf("disabled replication made %d calls with %d pending patches",
			len(connector.calls), len(testReplicator(dm).pending))
	}
}

func TestReplicationTaskGobRoundTrip(t *testing.T) {
	// types passed to rpl.replicate() must be gob-registered
	argTypes := []any{
		new(FilterWithAPIOpts),
		new(utils.StatQueueProfileWithAPIOpts),
		new(utils.StatQueueWithAPIOpts),
		new(utils.ThresholdProfileWithAPIOpts),
		new(utils.ThresholdWithAPIOpts),
		new(utils.AccountWithAPIOpts),
		new(utils.ActionProfileWithAPIOpts),
		new(utils.AttributeProfileWithAPIOpts),
		new(utils.ChargerProfileWithAPIOpts),
		new(utils.GetIndexesArg),
		new(utils.IPAllocationsWithAPIOpts),
		new(utils.IPProfileWithAPIOpts),
		new(utils.RankingProfileWithAPIOpts),
		new(utils.RankingWithAPIOpts),
		new(utils.RateProfileWithAPIOpts),
		new(utils.ResourceProfileWithAPIOpts),
		new(utils.ResourceWithAPIOpts),
		new(utils.RouteProfileWithAPIOpts),
		new(utils.SetIndexesArg),
		new(utils.TenantIDWithAPIOpts),
		new(utils.TrendProfileWithAPIOpts),
		new(utils.TrendWithAPIOpts),
	}
	for _, args := range argTypes {
		t.Run(fmt.Sprintf("%T", args), func(t *testing.T) {
			task := &ReplicationTask{
				ConnIDs: []string{"conn1"},
				Method:  "ReplicatorSv1.Test",
				Args:    args,
			}
			var buf bytes.Buffer
			if err := gob.NewEncoder(&buf).Encode(task); err != nil {
				t.Fatalf("encode: %v", err)
			}
			var decoded ReplicationTask
			if err := gob.NewDecoder(&buf).Decode(&decoded); err != nil {
				t.Fatalf("decode: %v", err)
			}
		})
	}
}

type warningLogger struct {
	*utils.StdLogger
	warnings []string
}

func (l *warningLogger) Warning(msg string) error {
	l.warnings = append(l.warnings, msg)
	return nil
}

func TestReplicatorImmediateFailureWarns(t *testing.T) {
	logger := &warningLogger{}
	oldLogger := utils.Logger
	utils.Logger = logger
	defer func() { utils.Logger = oldLogger }()

	r := newTestReplicator(t, 0, "", &mockConnector{err: errors.New("replication failed")})
	item := &config.ItemOpts{Replicate: true}
	r.replicate(context.Background(), utils.AttributeProfilePrefix, "destination",
		utils.ReplicatorSv1SetAttributeProfile,
		&utils.AttributeProfileWithAPIOpts{AttributeProfile: &utils.AttributeProfile{Tenant: "cgrates.org", ID: "destination"}}, item)
	if len(logger.warnings) != 1 || !strings.Contains(logger.warnings[0], "failed to replicate") {
		t.Fatalf("expected one replication warning, got %v", logger.warnings)
	}
}
