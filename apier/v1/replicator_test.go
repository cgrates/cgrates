// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1_test

import (
	"fmt"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"

	"github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type replicationRPCRequest struct {
	method string
	args   any
}

type replicationRPCMock chan *replicationRPCRequest

func (m replicationRPCMock) Call(_ *context.Context, method string, args, _ any) error {
	select {
	case m <- &replicationRPCRequest{method: method, args: args}:
		return nil
	default:
		return fmt.Errorf("unexpected second RPC call to %q", method)
	}
}

func TestReplicatorSetIndexesClear(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ApierCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	oldCfg := config.CgrConfig()
	config.SetCgrConfig(cfg)
	t.Cleanup(func() { config.SetCgrConfig(oldCfg) })

	idxItmType := utils.CacheAttributeFilterIndexes
	tntCtx := "cgrates.org:*cdrs"
	dataDB, err := engine.NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	if err := dataDB.SetIndexesDrv(idxItmType, tntCtx, map[string]utils.StringSet{
		"A": {"old": {}},
		"C": {"old": {}},
	}, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}

	cache := make(replicationRPCMock, 1)
	cacheConn := make(chan birpc.ClientConnector, 1)
	cacheConn <- cache
	connMgr := engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): cacheConn,
	})
	connMgr.Reload()
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	api := &v1.APIerSv1{Config: cfg, DataManager: dm, ConnMgr: connMgr}
	replicator := v1.NewReplicatorSv1(dm, api)
	args := &utils.SetIndexesArg{
		IdxItmType: idxItmType,
		TntCtx:     tntCtx,
		Indexes: map[string]utils.StringSet{
			"A": {"new": {}},
			"B": nil,
		},
		Clear:   true,
		Tenant:  "cgrates.org",
		APIOpts: map[string]any{utils.CacheOpt: utils.MetaLoad},
	}
	var reply string
	if err := replicator.SetIndexes(context.Background(), args, &reply); err != nil {
		t.Fatal(err)
	}
	if reply != utils.OK {
		t.Fatalf("reply = %q, want %q", reply, utils.OK)
	}
	got, err := dataDB.GetIndexesDrv(idxItmType, tntCtx)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]utils.StringSet{"A": {"new": {}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("stored indexes = %#v, want %#v", got, want)
	}
	if len(cache) != 1 {
		t.Fatalf("cache has %d calls, want 1", len(cache))
	}
	request := <-cache
	if request.method != utils.CacheSv1RemoveGroup {
		t.Fatalf("cache method = %q, want %q", request.method, utils.CacheSv1RemoveGroup)
	}
	cacheArgs, ok := request.args.(*utils.ArgsGetGroupWithAPIOpts)
	if !ok || cacheArgs.CacheID != idxItmType || cacheArgs.GroupID != tntCtx {
		t.Fatalf("cache args = %#v, want context group %q", request.args, tntCtx)
	}
}

func TestReplayGroupedIndexPatch(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	oldCfg := config.CgrConfig()
	config.SetCgrConfig(cfg)
	t.Cleanup(func() { config.SetCgrConfig(oldCfg) })

	const connID = "replicator-test"
	cfg.RPCConns()[connID] = config.NewDfltRPCConn()
	requests := make(replicationRPCMock, 1)
	connector := make(chan birpc.ClientConnector, 1)
	connector <- requests
	connMgr := engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{connID: connector})
	connMgr.Reload()

	tntCtx := "cgrates.org:*cdrs"
	indexes := map[string]utils.StringSet{
		"A": {"one": {}},
		"B": nil,
	}
	task := &engine.ReplicationTask{
		ConnIDs: []string{connID},
		ObjType: utils.CacheInstanceToPrefix[utils.CacheAttributeFilterIndexes],
		ObjID:   tntCtx,
		Method:  utils.ReplicatorSv1SetIndexes,
		Args: &utils.SetIndexesArg{
			IdxItmType: utils.CacheAttributeFilterIndexes,
			TntCtx:     tntCtx,
			Indexes:    indexes,
			Clear:      true,
		},
	}
	sourcePath := t.TempDir()
	path := filepath.Join(sourcePath, "index_patch"+utils.GOBSuffix)
	if err := task.WriteToFile(path); err != nil {
		t.Fatal(err)
	}
	api := &v1.APIerSv1{Config: cfg, ConnMgr: connMgr}
	var reply string
	if err := api.ReplayFailedReplications(context.Background(), v1.ReplayFailedReplicationsArgs{
		SourcePath: sourcePath,
		FailedPath: utils.MetaNone,
	}, &reply); err != nil {
		t.Fatal(err)
	}
	if reply != utils.OK {
		t.Fatalf("reply = %q, want %q", reply, utils.OK)
	}
	if len(requests) != 1 {
		t.Fatalf("replay made %d calls, want 1", len(requests))
	}
	request := <-requests
	got, ok := request.args.(*utils.SetIndexesArg)
	if request.method != utils.ReplicatorSv1SetIndexes || !ok || got.TntCtx != tntCtx ||
		!got.Clear || len(got.Indexes) != len(indexes) ||
		!reflect.DeepEqual(got.Indexes["A"], indexes["A"]) || len(got.Indexes["B"]) != 0 {
		t.Fatalf("replayed request = {Method:%q Args:%#v}", request.method, got)
	}
}

func TestReplicatorRemoveCacheRouting(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ApierCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	oldCfg := config.CgrConfig()
	config.SetCgrConfig(cfg)
	t.Cleanup(func() { config.SetCgrConfig(oldCfg) })

	dataDB, err := engine.NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	cache := make(replicationRPCMock, 1)
	cacheConn := make(chan birpc.ClientConnector, 1)
	cacheConn <- cache
	connMgr := engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): cacheConn,
	})
	connMgr.Reload()
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	api := &v1.APIerSv1{Config: cfg, DataManager: dm, ConnMgr: connMgr}
	replicator := v1.NewReplicatorSv1(dm, api)
	tests := []struct {
		name             string
		remove           func(*context.Context, *utils.TenantIDWithAPIOpts, *string) error
		expectedCacheIDs []string
	}{
		{"Threshold", replicator.RemoveThreshold, []string{utils.CacheThresholds}},
		{"StatQueue", replicator.RemoveStatQueue, []string{utils.CacheStatQueues}},
		{"Filter", replicator.RemoveFilter, []string{utils.CacheFilters}},
		{"ThresholdProfile", replicator.RemoveThresholdProfile, []string{utils.CacheThresholdProfiles, utils.CacheThresholds}},
		{"StatQueueProfile", replicator.RemoveStatQueueProfile, []string{utils.CacheStatQueueProfiles, utils.CacheStatQueues}},
		{"RankingProfile", replicator.RemoveRankingProfile, []string{utils.CacheRankingProfiles}},
		{"Ranking", replicator.RemoveRanking, []string{utils.CacheRankings}},
		{"Trend", replicator.RemoveTrend, []string{utils.CacheTrends}},
		{"TrendProfile", replicator.RemoveTrendProfile, []string{utils.CacheTrendProfiles}},
		{"Resource", replicator.RemoveResource, []string{utils.CacheResources}},
		{"ResourceProfile", replicator.RemoveResourceProfile, []string{utils.CacheResourceProfiles, utils.CacheResources}},
		{"IPAllocations", replicator.RemoveIPAllocations, []string{utils.CacheIPAllocations}},
		{"IPProfile", replicator.RemoveIPProfile, []string{utils.CacheIPProfiles, utils.CacheIPAllocations}},
		{"RouteProfile", replicator.RemoveRouteProfile, []string{utils.CacheRouteProfiles}},
		{"AttributeProfile", replicator.RemoveAttributeProfile, []string{utils.CacheAttributeProfiles}},
		{"ChargerProfile", replicator.RemoveChargerProfile, []string{utils.CacheChargerProfiles}},
		{"DispatcherProfile", replicator.RemoveDispatcherProfile, []string{utils.CacheDispatcherProfiles}},
		{"DispatcherHost", replicator.RemoveDispatcherHost, []string{utils.CacheDispatcherHosts}},
	}
	tntID := utils.NewTenantID("cgrates.org:item")
	args := &utils.TenantIDWithAPIOpts{
		TenantID: tntID,
		APIOpts:  map[string]any{utils.CacheOpt: utils.MetaRemove},
	}
	wantIDs := []string{tntID.TenantID()}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var reply string
			if err := test.remove(context.Background(), args, &reply); err != nil {
				t.Fatal(err)
			}
			if reply != utils.OK {
				t.Fatalf("reply = %q, want %q", reply, utils.OK)
			}
			if len(cache) != 1 {
				t.Fatalf("cache has %d calls, want 1", len(cache))
			}
			request := <-cache
			if request.method != utils.CacheSv1RemoveItems {
				t.Fatalf("cache method = %q, want %q", request.method, utils.CacheSv1RemoveItems)
			}
			cacheArgs, ok := request.args.(*utils.AttrReloadCacheWithAPIOpts)
			if !ok {
				t.Fatalf("cache args = %#v, want *utils.AttrReloadCacheWithAPIOpts", request.args)
			}
			cacheIDs := cacheArgs.Map()
			for _, cacheID := range test.expectedCacheIDs {
				if ids := cacheIDs[cacheID]; !reflect.DeepEqual(ids, wantIDs) {
					t.Fatalf("cache %q IDs = %v, want %v", cacheID, ids, wantIDs)
				}
				delete(cacheIDs, cacheID)
			}
			for cacheID, ids := range cacheIDs {
				if len(ids) != 0 {
					t.Fatalf("cache %q IDs = %v, want none", cacheID, ids)
				}
			}
		})
	}
}
