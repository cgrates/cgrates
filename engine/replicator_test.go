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
		utils.ReplicatorSv1SetRanking,
		utils.ReplicatorSv1RemoveRanking,
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
		if err := r.replicate(utils.DestinationPrefix, "destination", method, nil, item); err != nil {
			t.Fatal(err)
		}
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
				attempt := func(method string, args any) error {
					err := r.replicate(utils.DestinationPrefix, "destination", method, args, item)
					if mode.interval > 0 {
						r.flush()
					}
					return err
				}

				for i, method := range sequence.methods {
					err := attempt(method, sequence.args[i])
					if mode.interval == 0 && !errors.Is(err, failedErr) {
						t.Fatalf("replicate returned %v, want %v", err, failedErr)
					}
					if mode.interval > 0 && err != nil {
						t.Fatal(err)
					}
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
				if err := attempt(sequence.methods[0], sequence.args[0]); err != nil {
					t.Fatal(err)
				}
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
				if err := r.replicate(utils.DestinationPrefix, "destination", method, test.args[i], item); err != nil {
					t.Fatal(err)
				}
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
