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
				attempt := func(method string, args any) error {
					err := r.replicate(context.Background(), utils.AttributeProfilePrefix, "cgrates.org:destination", method, args, item)
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
				if err := r.replicate(context.Background(), utils.AttributeProfilePrefix, "cgrates.org:destination", method, test.args[i], item); err != nil {
					t.Fatal(err)
				}
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
