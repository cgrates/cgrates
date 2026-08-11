// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"bytes"
	"encoding/gob"
	"errors"
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
		{set: utils.ReplicatorSv1SetDestination, remove: utils.ReplicatorSv1RemoveDestination},
		{set: utils.ReplicatorSv1SetThresholdProfile, remove: utils.ReplicatorSv1RemoveThresholdProfile},
		{set: utils.ReplicatorSv1SetThreshold, remove: utils.ReplicatorSv1RemoveThreshold},
		{set: utils.ReplicatorSv1SetStatQueueProfile, remove: utils.ReplicatorSv1RemoveStatQueueProfile},
		{set: utils.ReplicatorSv1SetStatQueue, remove: utils.ReplicatorSv1RemoveStatQueue},
		{set: utils.ReplicatorSv1SetFilter, remove: utils.ReplicatorSv1RemoveFilter},
		{set: utils.ReplicatorSv1SetRankingProfile, remove: utils.ReplicatorSv1RemoveRankingProfile},
		{set: utils.ReplicatorSv1SetRanking, remove: utils.ReplicatorSv1RemoveRanking},
		{set: utils.ReplicatorSv1SetTrendProfile, remove: utils.ReplicatorSv1RemoveTrendProfile},
		{set: utils.ReplicatorSv1SetTrend, remove: utils.ReplicatorSv1RemoveTrend},
		{set: utils.ReplicatorSv1SetTiming, remove: utils.ReplicatorSv1RemoveTiming},
		{set: utils.ReplicatorSv1SetResourceProfile, remove: utils.ReplicatorSv1RemoveResourceProfile},
		{set: utils.ReplicatorSv1SetResource, remove: utils.ReplicatorSv1RemoveResource},
		{set: utils.ReplicatorSv1SetIPProfile, remove: utils.ReplicatorSv1RemoveIPProfile},
		{set: utils.ReplicatorSv1SetIPAllocations, remove: utils.ReplicatorSv1RemoveIPAllocations},
		{set: utils.ReplicatorSv1SetActionTriggers, remove: utils.ReplicatorSv1RemoveActionTriggers},
		{set: utils.ReplicatorSv1SetSharedGroup, remove: utils.ReplicatorSv1RemoveSharedGroup},
		{set: utils.ReplicatorSv1SetActions, remove: utils.ReplicatorSv1RemoveActions},
		{set: utils.ReplicatorSv1SetActionPlan, remove: utils.ReplicatorSv1RemoveActionPlan},
		{set: utils.ReplicatorSv1SetAccountActionPlans, remove: utils.ReplicatorSv1RemAccountActionPlans},
		{set: utils.ReplicatorSv1SetRouteProfile, remove: utils.ReplicatorSv1RemoveRouteProfile},
		{set: utils.ReplicatorSv1SetAttributeProfile, remove: utils.ReplicatorSv1RemoveAttributeProfile},
		{set: utils.ReplicatorSv1SetChargerProfile, remove: utils.ReplicatorSv1RemoveChargerProfile},
		{set: utils.ReplicatorSv1SetDispatcherProfile, remove: utils.ReplicatorSv1RemoveDispatcherProfile},
		{set: utils.ReplicatorSv1SetDispatcherHost, remove: utils.ReplicatorSv1RemoveDispatcherHost},
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
		utils.ReplicatorSv1SetRatingPlan,
		utils.ReplicatorSv1RemoveRatingPlan,
		utils.ReplicatorSv1SetRatingProfile,
		utils.ReplicatorSv1RemoveRatingProfile,
		utils.ReplicatorSv1SetIndexes,
		utils.ReplicatorSv1RemoveIndexes,
		utils.ReplicatorSv1SetBackupSessions,
		utils.ReplicatorSv1RemoveSessionBackup,
		utils.ReplicatorSv1SetReverseDestination,
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

func TestFailedReplicationFileName(t *testing.T) {
	ordinaryKey := replicationKey(utils.DestinationPrefix, "destination",
		utils.ReplicatorSv1SetDestination)
	if got := failedReplicationFileName(ordinaryKey); got != ordinaryKey {
		t.Fatalf("failed filename = %q, want unchanged key %q", got, ordinaryKey)
	}
	patchKey := indexPatchKey("parent")
	if got := failedReplicationFileName(patchKey); got == patchKey || !strings.HasPrefix(got, "index_") {
		t.Fatalf("failed filename = %q, want hashed index name", got)
	}
}

func TestReplicatorDestinationKeys(t *testing.T) {
	r := &replicator{
		pending:  make(map[string]*replicationData),
		interval: time.Second,
	}
	item := &config.ItemOpt{Replicate: true}
	for _, method := range []string{
		utils.ReplicatorSv1SetDestination,
		utils.ReplicatorSv1SetReverseDestination,
	} {
		r.replicate(utils.DestinationPrefix, "destination", method, nil, item)
	}
	if len(r.pending) != 2 {
		t.Fatalf("pending has %d items, want 2", len(r.pending))
	}
}

func TestReplicatorFailedTaskReplacement(t *testing.T) {
	set := &DestinationWithAPIOpts{Destination: &Destination{Id: "destination"}}
	remove := &utils.StringWithAPIOpts{Arg: "destination"}
	sequences := []struct {
		name    string
		methods [2]string
		args    [2]any
	}{
		{
			name:    "SetRemove",
			methods: [2]string{utils.ReplicatorSv1SetDestination, utils.ReplicatorSv1RemoveDestination},
			args:    [2]any{set, remove},
		},
		{
			name:    "RemoveSet",
			methods: [2]string{utils.ReplicatorSv1RemoveDestination, utils.ReplicatorSv1SetDestination},
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
				item := &config.ItemOpt{Replicate: true}
				attempt := func(method string, args any) {
					r.replicate(utils.DestinationPrefix, "destination", method, args, item)
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
				wantName := utils.DestinationPrefix + "destination" + utils.GOBSuffix
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
	set1 := &DestinationWithAPIOpts{Destination: &Destination{Id: "destination"}}
	set2 := &DestinationWithAPIOpts{Destination: &Destination{Id: "destination"}}
	remove := &utils.StringWithAPIOpts{Arg: "destination"}
	tests := []struct {
		name       string
		methods    [2]string
		args       [2]any
		wantMethod string
		wantArgs   any
	}{
		{
			name:       "SetSet",
			methods:    [2]string{utils.ReplicatorSv1SetDestination, utils.ReplicatorSv1SetDestination},
			args:       [2]any{set1, set2},
			wantMethod: utils.ReplicatorSv1SetDestination,
			wantArgs:   set2,
		},
		{
			name:       "SetRemove",
			methods:    [2]string{utils.ReplicatorSv1SetDestination, utils.ReplicatorSv1RemoveDestination},
			args:       [2]any{set1, remove},
			wantMethod: utils.ReplicatorSv1RemoveDestination,
			wantArgs:   remove,
		},
		{
			name:       "RemoveSet",
			methods:    [2]string{utils.ReplicatorSv1RemoveDestination, utils.ReplicatorSv1SetDestination},
			args:       [2]any{remove, set2},
			wantMethod: utils.ReplicatorSv1SetDestination,
			wantArgs:   set2,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := &replicator{
				pending:  make(map[string]*replicationData),
				interval: time.Second,
			}
			item := &config.ItemOpt{Replicate: true}
			for i, method := range test.methods {
				r.replicate(utils.DestinationPrefix, "destination", method, test.args[i], item)
			}
			if len(r.pending) != 1 {
				t.Fatalf("pending has %d items, want 1", len(r.pending))
			}
			got, ok := r.pending[utils.DestinationPrefix+"destination"]
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
	cfg := config.CgrConfig()
	cfg.DataDbCfg().Items[utils.CacheAttributeFilterIndexes].Replicate = true
	db, err := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if err != nil {
		tb.Fatal(err)
	}
	dm := NewDataManager(db, cfg.CacheCfg(), cm)
	dm.replicator.interval = interval
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
					err = dm.RemoveIndexes(idxItmType, tntCtx)
				case operation.remove != nil:
					err = dm.RemoveIndexes(idxItmType, tntCtx, operation.remove...)
				default:
					err = dm.SetIndexes(idxItmType, tntCtx, operation.indexes, true, "")
				}
				if err != nil {
					t.Fatal(err)
				}
			}
			if len(dm.replicator.pending) != 1 {
				t.Fatalf("pending has %d patches, want 1", len(dm.replicator.pending))
			}
			for _, pending := range dm.replicator.pending {
				if pending.objID != tntCtx {
					t.Fatalf("pending objID = %q, want %q", pending.objID, tntCtx)
				}
			}
			dm.replicator.flush()
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
			return dm.SetIndexes(idxItmType, tntCtx, map[string]utils.StringSet{
				fieldA: {"RL1": {}},
				fieldB: {"RL2": {}},
			}, true, "")
		},
		func() error { return dm.RemoveIndexes(idxItmType, tntCtx, fieldA, fieldB) },
		func() error { return dm.RemoveIndexes(idxItmType, tntCtx) },
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
	if err := dm.SetIndexes(idxItmType, tntCtx, indexes, true, ""); err != nil {
		t.Fatalf("SetIndexes returned replication error: %v", err)
	}
	if len(connector.calls) != 1 {
		t.Fatalf("connector has %d calls, want 1", len(connector.calls))
	}
	stored, err := dm.DataDB().GetIndexesDrv(idxItmType, tntCtx)
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
		if err := dm.RemoveIndexes(idxItmType, tntCtx); err != nil {
			t.Fatal(err)
		}
		if err := dm.SetIndexes(idxItmType, tntCtx,
			map[string]utils.StringSet{fieldA: valueOne}, true, ""); err != nil {
			t.Fatal(err)
		}
		dm.replicator.flush()
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
		if err := dm.SetIndexes(idxItmType, tntCtx,
			map[string]utils.StringSet{fieldB: valueTwo}, true, ""); err != nil {
			t.Fatal(err)
		}
		dm.replicator.flush()
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
		if err := dm.SetIndexes(idxItmType, tntCtx,
			map[string]utils.StringSet{fieldA: valueOne}, true, ""); err != nil {
			t.Fatal(err)
		}
		dm.replicator.flush()
		connector.err = nil
		if err := dm.RemoveIndexes(idxItmType, tntCtx); err != nil {
			t.Fatal(err)
		}
		if err := dm.SetIndexes(idxItmType, tntCtx,
			map[string]utils.StringSet{fieldB: valueTwo}, true, ""); err != nil {
			t.Fatal(err)
		}
		dm.replicator.flush()
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
		if err := dm.SetIndexes(idxItmType, tntCtx, indexes, false, "transaction"); err != nil {
			t.Fatal(err)
		}
		key := replicationKey(objType, tntCtx, utils.ReplicatorSv1SetIndexes)
		if _, has := dm.replicator.pending[key]; !has {
			t.Fatal("transaction call did not use the generic pending path")
		}
		if _, has := dm.replicator.pending[indexPatchKey(key)]; has {
			t.Fatal("transaction call entered the grouped index path")
		}
		dm.replicator.flush()
		if len(connector.calls) != 1 || !reflect.DeepEqual(connector.calls[0].Indexes, indexes) {
			t.Fatalf("connector calls = %#v, want one transaction call with %#v", connector.calls, indexes)
		}
	})

	t.Run("immediate transaction failure keeps API success", func(t *testing.T) {
		failedDir := t.TempDir()
		connector := &indexConnector{err: errors.New("replication failed")}
		dm := newIndexDataManager(t, 0, failedDir, connector)
		indexes := map[string]utils.StringSet{fieldA: valueOne}
		if err := dm.SetIndexes(idxItmType, tntCtx, indexes, false, "transaction"); err != nil {
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
	config.CgrConfig().DataDbCfg().Items[idxItmType].Replicate = false
	if err := dm.SetIndexes(idxItmType, tntCtx,
		map[string]utils.StringSet{fieldA: {"RL1": {}}}, true, ""); err != nil {
		t.Fatal(err)
	}
	if err := dm.RemoveIndexes(idxItmType, tntCtx, fieldA); err != nil {
		t.Fatal(err)
	}
	dm.replicator.flush()
	if len(connector.calls) != 0 || len(dm.replicator.pending) != 0 {
		t.Fatalf("disabled replication made %d calls with %d pending patches",
			len(connector.calls), len(dm.replicator.pending))
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
	item := &config.ItemOpt{Replicate: true}
	r.replicate(utils.DestinationPrefix, "destination",
		utils.ReplicatorSv1SetDestination,
		&DestinationWithAPIOpts{Destination: &Destination{Id: "destination"}}, item)
	if len(logger.warnings) != 1 || !strings.Contains(logger.warnings[0], "failed to replicate") {
		t.Fatalf("expected one replication warning, got %v", logger.warnings)
	}
}
