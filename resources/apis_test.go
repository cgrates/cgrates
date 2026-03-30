/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package resources

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/rpcclient"

	"github.com/cgrates/cgrates/utils"
)

func TestResourceV1AuthorizeResourceMissingStruct(t *testing.T) {
	var dmRES *engine.DataManager
	cfg := config.NewDefaultCGRConfig()

	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dmRES = engine.NewDataManager(dbCM, cfg, nil)
	cfg.ResourceSCfg().StoreInterval = 1
	cfg.ResourceSCfg().StringIndexedFields = nil
	cfg.ResourceSCfg().PrefixIndexedFields = nil
	fltrs := engine.NewFilterS(cfg, nil, dmRES)
	resService := NewResourceService(dmRES, cfg,
		fltrs, nil)
	var reply *string
	argsMissingTenant := &utils.CGREvent{
		ID:    "id1",
		Event: map[string]any{},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID: "test1",
			utils.OptsResourcesUnits:   20,
		},
	}
	argsMissingUsageID := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "id1",
		Event:  map[string]any{},
		APIOpts: map[string]any{
			utils.OptsResourcesUnits: 20,
		},
	}
	if err := resService.V1AuthorizeResources(context.TODO(), argsMissingTenant, reply); err != nil && err.Error() != "MANDATORY_IE_MISSING: [Event]" {
		t.Error(err.Error())
	}
	if err := resService.V1AuthorizeResources(context.TODO(), argsMissingUsageID, reply); err != nil && err.Error() != "MANDATORY_IE_MISSING: [Event]" {
		t.Error(err.Error())
	}
}

func TestResourceAllocateResourceOtherDB(t *testing.T) {
	rProf := &utils.ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "RL_DB",
		FilterIDs: []string{"*string:~*opts.Resource:RL_DB"},
		Weights: utils.DynamicWeights{
			{
				Weight: 100,
			},
		},
		Limit:        2,
		ThresholdIDs: []string{utils.MetaNone},
		UsageTTL:     -time.Nanosecond,
	}

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	idb, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: idb}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	fltS := engine.NewFilterS(cfg, nil, dm)
	rs := NewResourceService(dm, cfg, fltS, nil)
	if err := dm.SetResourceProfile(context.TODO(), rProf, true); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetResource(context.TODO(), &utils.Resource{
		Tenant: "cgrates.org",
		ID:     "RL_DB",
		Usages: map[string]*utils.ResourceUsage{
			"RU1": { // the resource in DB is expired (should be cleaned when the next allocate is called)
				Tenant:     "cgrates.org",
				ID:         "RU1",
				ExpiryTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
				Units:      1,
			},
		},
		TTLIdx: []string{"RU1"},
	}); err != nil { // simulate how the resource is stored in redis or mongo(non-exported fields are not populated)
		t.Fatal(err)
	}
	var reply string
	exp := rProf.ID
	if err := rs.V1AllocateResources(context.TODO(), &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ef0f554",
		Event:  map[string]any{"": ""},
		APIOpts: map[string]any{
			"Resource":                 "RL_DB",
			utils.OptsResourcesUsageID: "56156434-2e44-4f16-a766-086f10b413cd",
			utils.OptsResourcesUnits:   1,
		},
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != exp {
		t.Errorf("Expected: %q, received: %q", exp, reply)
	}
}

func TestResourcesV1ResourcesForEventOK(t *testing.T) {
	rS, dm := newTestResourceSWithCache(t)
	rsPrf := &utils.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Limit:    10,
		UsageTTL: time.Minute,
	}
	rs := &utils.Resource{
		Tenant: "cgrates.org",
		ID:     "RES1",
		TTLIdx: []string{},
		Usages: map[string]*utils.ResourceUsage{
			"RU1": {
				Tenant: "cgrates.org",
				ID:     "RU1",
				Units:  10,
			},
		},
	}
	if err := dm.SetResourceProfile(context.Background(), rsPrf, true); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetResource(context.Background(), rs); err != nil {
		t.Fatal(err)
	}

	args := &utils.CGREvent{
		ID: "ResourcesForEventTest",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID: "RU_TEST1",
		},
	}

	exp := Resources{
		{
			Resource: &utils.Resource{
				Tenant: "cgrates.org",
				ID:     "RES1",
				TTLIdx: []string{},
				Usages: map[string]*utils.ResourceUsage{
					"RU1": {
						Tenant: "cgrates.org",
						ID:     "RU1",
						Units:  10,
					},
				},
			},
			profile: rsPrf,
			ttl:     utils.DurationPointer(72 * time.Hour),
		},
	}
	var reply Resources
	if err := rS.V1GetResourcesForEvent(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp), utils.ToJSON(reply))
	}
}

func TestResourcesV1ResourcesForEventNotFound(t *testing.T) {
	rS, dm := newTestResourceSWithCache(t)
	rsPrf := &utils.ResourceProfile{
		Tenant:       "cgrates.org",
		ID:           "RES1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		ThresholdIDs: []string{utils.MetaNone},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Limit:    10,
		UsageTTL: time.Minute,
	}
	rs := &utils.Resource{
		Tenant: "cgrates.org",
		ID:     "RES1",
		Usages: map[string]*utils.ResourceUsage{
			"RU1": {
				Tenant: "cgrates.org",
				ID:     "RU1",
				Units:  10,
			},
		},
	}
	if err := dm.SetResourceProfile(context.Background(), rsPrf, true); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetResource(context.Background(), rs); err != nil {
		t.Fatal(err)
	}

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ResourcesForEventTest",
		Event: map[string]any{
			utils.AccountField: "1002",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID: "RU_TEST1",
		},
	}

	var reply Resources
	if err := rS.V1GetResourcesForEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestResourcesV1ResourcesForEventMissingParameters(t *testing.T) {
	rS, dm := newTestResourceSWithCache(t)

	rsPrf := &utils.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weights:           utils.DynamicWeights{{Weight: 10}},
		Limit:             10,
		UsageTTL:          time.Minute,
	}
	if err := dm.SetResourceProfile(context.Background(), rsPrf, true); err != nil {
		t.Fatal(err)
	}

	var reply Resources

	// missing UsageID
	args := &utils.CGREvent{
		ID: "EventAuthorizeResource",
		Event: map[string]any{
			utils.AccountField:          "1001",
			utils.OptsResourcesUnits:    5,
			utils.OptsResourcesUsageTTL: time.Minute,
		},
	}
	experr := `MANDATORY_IE_MISSING: [UsageID]`
	if err := rS.V1GetResourcesForEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	// missing Event
	args = &utils.CGREvent{
		ID: "EventAuthorizeResource",
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID:  "RU_Test",
			utils.OptsResourcesUnits:    5,
			utils.OptsResourcesUsageTTL: time.Minute,
		},
	}
	experr = `MANDATORY_IE_MISSING: [Event]`
	if err := rS.V1GetResourcesForEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	// missing ID
	args = &utils.CGREvent{
		Event: map[string]any{utils.AccountField: "1001"},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID:  "RU_Test",
			utils.OptsResourcesUnits:    5,
			utils.OptsResourcesUsageTTL: time.Minute,
		},
	}
	experr = `MANDATORY_IE_MISSING: [ID]`
	if err := rS.V1GetResourcesForEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	// nil args
	experr = `MANDATORY_IE_MISSING: [Event]`
	if err := rS.V1GetResourcesForEvent(context.Background(), nil, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestResourcesV1ResourcesForEventCacheReplyExists(t *testing.T) {
	tmp := engine.Cache
	t.Cleanup(func() { engine.Cache = tmp })

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	config.SetCgrConfig(cfg)
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	cacheKey := utils.ConcatenatedKey(utils.ResourceSv1GetResourcesForEvent,
		utils.ConcatenatedKey("cgrates.org", "ResourcesForEventTest"))
	rsPrf := &utils.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Limit:    10,
		UsageTTL: time.Minute,
	}
	rs := &utils.Resource{
		Tenant: "cgrates.org",
		ID:     "RES1",
		Usages: map[string]*utils.ResourceUsage{
			"RU1": {
				Tenant: "cgrates.org",
				ID:     "RU1",
				Units:  10,
			},
		},
		TTLIdx: []string{},
	}
	err := dm.SetResourceProfile(context.Background(), rsPrf, true)
	if err != nil {
		t.Error(err)
	}
	err = dm.SetResource(context.Background(), rs)
	if err != nil {
		t.Error(err)
	}

	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := &utils.CGREvent{
		ID: "ResourcesForEventTest",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID: "RU_TEST1",
		},
	}

	cacheReply := Resources{
		{
			Resource: &utils.Resource{
				Tenant: "cgrates.org",
				ID:     "RES1",
				Usages: map[string]*utils.ResourceUsage{
					"RU1": {
						Tenant: "cgrates.org",
						ID:     "RU1",
						Units:  10,
					},
				},
				TTLIdx: []string{},
			},
			profile:    rsPrf,
			dirty:      utils.BoolPointer(false),
			totalUsage: utils.Float64Pointer(10),
			ttl:        utils.DurationPointer(time.Minute),
		},
	}
	engine.Cache.Set(context.Background(), utils.CacheRPCResponses, cacheKey,
		&utils.CachedRPCResponse{Result: &cacheReply, Error: nil},
		nil, true, utils.NonTransactional)
	var reply Resources
	if err := rS.V1GetResourcesForEvent(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, cacheReply) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(cacheReply), utils.ToJSON(reply))
	}

	config.SetCgrConfig(config.NewDefaultCGRConfig())
}

func TestResourcesV1ResourcesForEventCacheReplySet(t *testing.T) {
	tmp := engine.Cache
	t.Cleanup(func() { engine.Cache = tmp })

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	config.SetCgrConfig(cfg)
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	cacheKey := utils.ConcatenatedKey(utils.ResourceSv1GetResourcesForEvent,
		utils.ConcatenatedKey("cgrates.org", "ResourcesForEventTest"))
	rsPrf := &utils.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Limit:    10,
		UsageTTL: time.Minute,
	}
	rs := &utils.Resource{
		Tenant: "cgrates.org",
		ID:     "RES1",
		Usages: map[string]*utils.ResourceUsage{
			"RU1": {
				Tenant: "cgrates.org",
				ID:     "RU1",
				Units:  10,
			},
		},
		TTLIdx: []string{},
	}
	err := dm.SetResourceProfile(context.Background(), rsPrf, true)
	if err != nil {
		t.Error(err)
	}
	err = dm.SetResource(context.Background(), rs)
	if err != nil {
		t.Error(err)
	}

	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := &utils.CGREvent{
		ID: "ResourcesForEventTest",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID: "RU_TEST1",
		},
	}

	exp := &Resources{
		{
			Resource: &utils.Resource{
				Tenant: "cgrates.org",
				ID:     "RES1",
				Usages: map[string]*utils.ResourceUsage{
					"RU1": {
						Tenant: "cgrates.org",
						ID:     "RU1",
						Units:  10,
					},
				},
				TTLIdx: []string{},
			},
			ttl:     utils.DurationPointer(72 * time.Hour),
			profile: rsPrf,
		},
	}
	var reply Resources
	if err := rS.V1GetResourcesForEvent(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, *exp) {
		t.Errorf("expected: <%v>, received: <%v>", exp, reply)
	}

	if itm, has := engine.Cache.Get(utils.CacheRPCResponses, cacheKey); has {
		resp := itm.(*utils.CachedRPCResponse)
		if !reflect.DeepEqual(resp.Result, exp) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(resp.Result))
		}
	}

	config.SetCgrConfig(config.NewDefaultCGRConfig())
}

func TestResourcesV1GetResource(t *testing.T) {
	tests := []struct {
		name        string
		storeRes    bool
		queryTenant string
		queryID     string
		wantErr     string
		wantReply   utils.Resource
	}{
		{
			name:     "OK",
			storeRes: true,
			queryID:  "RES1",
			wantReply: utils.Resource{
				Tenant: "cgrates.org",
				ID:     "RES1",
				TTLIdx: []string{},
				Usages: map[string]*utils.ResourceUsage{
					"RU1": {
						Tenant: "cgrates.org",
						ID:     "RU1",
						Units:  10,
					},
				},
			},
		},
		{
			name:        "NotFound",
			storeRes:    true,
			queryTenant: "cgrates.org",
			queryID:     "RES2",
			wantErr:     utils.ErrNotFound.Error(),
		},
		{
			name:    "MissingParameters",
			wantErr: `MANDATORY_IE_MISSING: [ID]`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rS, dm := newTestResourceSWithCache(t)
			if tc.storeRes {
				rs := &utils.Resource{
					Tenant: "cgrates.org",
					ID:     "RES1",
					Usages: map[string]*utils.ResourceUsage{
						"RU1": {
							Tenant: "cgrates.org",
							ID:     "RU1",
							Units:  10,
						},
					},
					TTLIdx: []string{},
				}
				if err := dm.SetResource(context.Background(), rs); err != nil {
					t.Fatal(err)
				}
			}

			args := &utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: tc.queryTenant,
					ID:     tc.queryID,
				},
			}
			var reply utils.Resource
			err := rS.V1GetResource(context.Background(), args, &reply)
			if tc.wantErr != "" {
				if err == nil || err.Error() != tc.wantErr {
					t.Errorf("expected: <%+v>, \nreceived: <%+v>", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(reply, tc.wantReply) {
				t.Errorf("expected: <%+v>, \nreceived: <%+v>",
					utils.ToJSON(tc.wantReply), utils.ToJSON(reply))
			}
		})
	}
}

func TestResourcesV1GetResourceWithConfig(t *testing.T) {
	rsPrf := &utils.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weights:           utils.DynamicWeights{{Weight: 10}},
		Limit:             10,
		UsageTTL:          time.Minute,
	}
	rs := &utils.Resource{
		Tenant: "cgrates.org",
		ID:     "RES1",
		Usages: map[string]*utils.ResourceUsage{
			"RU1": {
				Tenant: "cgrates.org",
				ID:     "RU1",
				Units:  10,
			},
		},
		TTLIdx: []string{},
	}

	tests := []struct {
		name       string
		profileID  string // store profile with this ID; empty = skip
		resourceID string // store resource with this ID
		queryID    string
		wantErr    string
	}{
		{name: "OK", profileID: "RES1", resourceID: "RES1", queryID: "RES1"},
		{name: "NilProfileNotFound", profileID: "RES2", resourceID: "RES1", queryID: "RES1",
			wantErr: utils.ErrNotFound.Error()},
		{name: "ResourceNotFound", resourceID: "RES2", queryID: "RES1",
			wantErr: utils.ErrNotFound.Error()},
		{name: "MissingParameters", resourceID: "RES1",
			wantErr: `MANDATORY_IE_MISSING: [ID]`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rS, dm := newTestResourceSWithCache(t)
			if tc.profileID != "" {
				prf := *rsPrf
				prf.ID = tc.profileID
				if err := dm.SetResourceProfile(context.Background(), &prf, true); err != nil {
					t.Fatal(err)
				}
			}
			res := *rs
			res.ID = tc.resourceID
			if err := dm.SetResource(context.Background(), &res); err != nil {
				t.Fatal(err)
			}

			args := &utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{ID: tc.queryID},
			}
			var reply utils.ResourceWithConfig
			err := rS.V1GetResourceWithConfig(context.Background(), args, &reply)
			if tc.wantErr != "" {
				if err == nil || err.Error() != tc.wantErr {
					t.Errorf("expected: <%+v>, \nreceived: <%+v>", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			exp := utils.ResourceWithConfig{
				Resource: &res,
				Config:   rsPrf,
			}
			if !reflect.DeepEqual(reply, exp) {
				t.Errorf("expected: <%+v>, \nreceived: <%+v>",
					utils.ToJSON(exp), utils.ToJSON(reply))
			}
		})
	}
}

func TestResourcesV1AuthorizeResources(t *testing.T) {
	tests := []struct {
		name      string
		limit     float64
		args      *utils.CGREvent
		wantErr   string
		wantReply string
	}{
		{
			name:  "OK",
			limit: 10,
			args: &utils.CGREvent{
				ID:    "EventAuthorizeResource",
				Event: map[string]any{utils.AccountField: "1001"},
				APIOpts: map[string]any{
					utils.OptsResourcesUsageID:  "RU_Test",
					utils.OptsResourcesUnits:    5,
					utils.OptsResourcesUsageTTL: time.Minute,
				},
			},
			wantReply: "Approved",
		},
		{
			name:  "NotAuthorized",
			limit: 0,
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "EventAuthorizeResource",
				Event:  map[string]any{utils.AccountField: "1001"},
				APIOpts: map[string]any{
					utils.OptsResourcesUsageID:  "RU_Test",
					utils.OptsResourcesUnits:    5,
					utils.OptsResourcesUsageTTL: time.Minute,
				},
			},
			wantErr: utils.ErrResourceUnauthorized.Error(),
		},
		{
			name:  "NoMatch",
			limit: 10,
			args: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "EventAuthorizeResource",
				Event:  map[string]any{utils.AccountField: "1002"},
				APIOpts: map[string]any{
					utils.OptsResourcesUsageID:  "RU_Test",
					utils.OptsResourcesUnits:    5,
					utils.OptsResourcesUsageTTL: time.Minute,
				},
			},
			wantErr: utils.ErrNotFound.Error(),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rS, dm := newTestResourceSWithCache(t)
			rsPrf := &utils.ResourceProfile{
				Tenant:            "cgrates.org",
				ID:                "RES1",
				FilterIDs:         []string{"*string:~*req.Account:1001"},
				ThresholdIDs:      []string{utils.MetaNone},
				AllocationMessage: "Approved",
				Weights:           utils.DynamicWeights{{Weight: 10}},
				Limit:             tc.limit,
				UsageTTL:          time.Minute,
			}
			if err := dm.SetResourceProfile(context.Background(), rsPrf, true); err != nil {
				t.Fatal(err)
			}

			var reply string
			err := rS.V1AuthorizeResources(context.Background(), tc.args, &reply)
			if tc.wantErr != "" {
				if err == nil || err.Error() != tc.wantErr {
					t.Errorf("expected: <%+v>, \nreceived: <%+v>", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if reply != tc.wantReply {
				t.Errorf("expected: %q, received: %q", tc.wantReply, reply)
			}
		})
	}
}

type callFunc = func(*ResourceS, *context.Context, *utils.CGREvent, *string) error

func TestResourcesV1CacheReplyExists(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		eventID    string
		limit      float64
		cacheReply string
		call       callFunc
	}{
		{
			name:       "Authorize",
			method:     utils.ResourceSv1AuthorizeResources,
			eventID:    "EventAuthorizeResource",
			limit:      10,
			cacheReply: "Approved",
			call: func(rS *ResourceS, ctx *context.Context, ev *utils.CGREvent, reply *string) error {
				return rS.V1AuthorizeResources(ctx, ev, reply)
			},
		},
		{
			name:       "Allocate",
			method:     utils.ResourceSv1AllocateResources,
			eventID:    "EventAllocateResource",
			limit:      -1,
			cacheReply: "cacheApproved",
			call: func(rS *ResourceS, ctx *context.Context, ev *utils.CGREvent, reply *string) error {
				return rS.V1AllocateResources(ctx, ev, reply)
			},
		},
		{
			name:       "Release",
			method:     utils.ResourceSv1ReleaseResources,
			eventID:    "EventReleaseResource",
			limit:      -1,
			cacheReply: "cacheReply",
			call: func(rS *ResourceS, ctx *context.Context, ev *utils.CGREvent, reply *string) error {
				return rS.V1ReleaseResources(ctx, ev, reply)
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmp := engine.Cache
			t.Cleanup(func() {
				engine.Cache = tmp
				config.SetCgrConfig(config.NewDefaultCGRConfig())
			})

			engine.Cache.Clear(nil)
			cfg := config.NewDefaultCGRConfig()
			cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
			config.SetCgrConfig(cfg)
			data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
			dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
			dm := engine.NewDataManager(dbCM, cfg, nil)
			engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)

			cacheKey := utils.ConcatenatedKey(tc.method,
				utils.ConcatenatedKey("cgrates.org", tc.eventID))

			rsPrf := &utils.ResourceProfile{
				Tenant:            "cgrates.org",
				ID:                "RES1",
				FilterIDs:         []string{"*string:~*req.Account:1001"},
				ThresholdIDs:      []string{utils.MetaNone},
				AllocationMessage: "Approved",
				Weights:           utils.DynamicWeights{{Weight: 10}},
				Limit:             tc.limit,
				UsageTTL:          time.Minute,
			}
			rs := &utils.Resource{
				Tenant: "cgrates.org",
				ID:     "RES1",
				Usages: map[string]*utils.ResourceUsage{
					"RU1": {
						Tenant: "cgrates.org",
						ID:     "RU1",
						Units:  10,
					},
				},
				TTLIdx: []string{},
			}

			if err := dm.SetResourceProfile(context.Background(), rsPrf, true); err != nil {
				t.Fatal(err)
			}
			if err := dm.SetResource(context.Background(), rs); err != nil {
				t.Fatal(err)
			}

			fltrs := engine.NewFilterS(cfg, nil, dm)
			rS := NewResourceService(dm, cfg, fltrs, nil)

			args := &utils.CGREvent{
				ID:    tc.eventID,
				Event: map[string]any{utils.AccountField: "1001"},
				APIOpts: map[string]any{
					utils.OptsResourcesUsageID:  "RU_Test",
					utils.OptsResourcesUnits:    5,
					utils.OptsResourcesUsageTTL: time.Minute,
				},
			}

			cachedVal := tc.cacheReply
			engine.Cache.Set(context.Background(), utils.CacheRPCResponses, cacheKey,
				&utils.CachedRPCResponse{Result: &cachedVal, Error: nil},
				nil, true, utils.NonTransactional)

			var reply string
			if err := tc.call(rS, context.Background(), args, &reply); err != nil {
				t.Error(err)
			} else if reply != tc.cacheReply {
				t.Errorf("Unexpected reply returned: %q", reply)
			}
		})
	}
}

func TestResourcesV1CacheReplySet(t *testing.T) {
	tests := []struct {
		name            string
		method          string
		eventID         string
		units           int
		wantErr         string
		wantReply       string
		wantCachedReply string
		call            callFunc
	}{
		{
			name:            "Authorize",
			method:          utils.ResourceSv1AuthorizeResources,
			eventID:         "EventAuthorizeResource",
			units:           2,
			wantReply:       "Approved",
			wantCachedReply: "Approved",
			call: func(rS *ResourceS, ctx *context.Context, ev *utils.CGREvent, reply *string) error {
				return rS.V1AuthorizeResources(ctx, ev, reply)
			},
		},
		{
			name:            "Allocate",
			method:          utils.ResourceSv1AllocateResources,
			eventID:         "EventAllocateResource",
			units:           2,
			wantReply:       "Approved",
			wantCachedReply: "Approved",
			call: func(rS *ResourceS, ctx *context.Context, ev *utils.CGREvent, reply *string) error {
				return rS.V1AllocateResources(ctx, ev, reply)
			},
		},
		{
			name:            "Release",
			method:          utils.ResourceSv1ReleaseResources,
			eventID:         "EventReleaseResource",
			units:           2,
			wantErr:         "cannot find usage record with id: RU_Test",
			wantCachedReply: "",
			call: func(rS *ResourceS, ctx *context.Context, ev *utils.CGREvent, reply *string) error {
				return rS.V1ReleaseResources(ctx, ev, reply)
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmp := engine.Cache
			t.Cleanup(func() {
				engine.Cache = tmp
				config.SetCgrConfig(config.NewDefaultCGRConfig())
			})

			engine.Cache.Clear(nil)
			cfg := config.NewDefaultCGRConfig()
			cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
			config.SetCgrConfig(cfg)
			data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
			dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
			dm := engine.NewDataManager(dbCM, cfg, nil)
			engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)

			cacheKey := utils.ConcatenatedKey(tc.method,
				utils.ConcatenatedKey("cgrates.org", tc.eventID))

			rsPrf := &utils.ResourceProfile{
				Tenant:            "cgrates.org",
				ID:                "RES1",
				FilterIDs:         []string{"*string:~*req.Account:1001"},
				ThresholdIDs:      []string{utils.MetaNone},
				AllocationMessage: "Approved",
				Weights:           utils.DynamicWeights{{Weight: 10}},
				Limit:             -1,
				UsageTTL:          time.Minute,
			}
			rs := &utils.Resource{
				Tenant: "cgrates.org",
				ID:     "RES1",
				Usages: map[string]*utils.ResourceUsage{
					"RU1": {
						Tenant: "cgrates.org",
						ID:     "RU1",
						Units:  4,
					},
				},
				TTLIdx: []string{},
			}

			if err := dm.SetResourceProfile(context.Background(), rsPrf, true); err != nil {
				t.Fatal(err)
			}
			if err := dm.SetResource(context.Background(), rs); err != nil {
				t.Fatal(err)
			}

			fltrs := engine.NewFilterS(cfg, nil, dm)
			rS := NewResourceService(dm, cfg, fltrs, nil)

			args := &utils.CGREvent{
				ID:    tc.eventID,
				Event: map[string]any{utils.AccountField: "1001"},
				APIOpts: map[string]any{
					utils.OptsResourcesUsageID:  "RU_Test",
					utils.OptsResourcesUnits:    tc.units,
					utils.OptsResourcesUsageTTL: time.Minute,
				},
			}

			var reply string
			err := tc.call(rS, context.Background(), args, &reply)
			if tc.wantErr != "" {
				if err == nil || err.Error() != tc.wantErr {
					t.Errorf("expected error: <%+v>, got: <%+v>", tc.wantErr, err)
				}
			} else {
				if err != nil {
					t.Error(err)
				} else if reply != tc.wantReply {
					t.Errorf("Unexpected reply returned: %q", reply)
				}
			}

			if itm, has := engine.Cache.Get(utils.CacheRPCResponses, cacheKey); has {
				resp := itm.(*utils.CachedRPCResponse)
				if *resp.Result.(*string) != tc.wantCachedReply {
					t.Errorf("expected cached: <%+v>, received: <%+v>",
						tc.wantCachedReply, *resp.Result.(*string))
				}
			}
		})
	}
}

func TestResourcesV1MissingParameters(t *testing.T) {
	methods := []struct {
		name string
		call callFunc
	}{
		{"AuthorizeResources", func(rS *ResourceS, ctx *context.Context, ev *utils.CGREvent, reply *string) error {
			return rS.V1AuthorizeResources(ctx, ev, reply)
		}},
		{"AllocateResources", func(rS *ResourceS, ctx *context.Context, ev *utils.CGREvent, reply *string) error {
			return rS.V1AllocateResources(ctx, ev, reply)
		}},
		{"ReleaseResources", func(rS *ResourceS, ctx *context.Context, ev *utils.CGREvent, reply *string) error {
			return rS.V1ReleaseResources(ctx, ev, reply)
		}},
	}
	for _, m := range methods {
		t.Run(m.name, func(t *testing.T) {
			rS, dm := newTestResourceSWithCache(t)

			rsPrf := &utils.ResourceProfile{
				Tenant:            "cgrates.org",
				ID:                "RES1",
				FilterIDs:         []string{"*string:~*req.Account:1001"},
				ThresholdIDs:      []string{utils.MetaNone},
				AllocationMessage: "Approved",
				Weights:           utils.DynamicWeights{{Weight: 10}},
				Limit:             10,
				UsageTTL:          time.Minute,
			}
			if err := dm.SetResourceProfile(context.Background(), rsPrf, true); err != nil {
				t.Fatal(err)
			}

			var reply string

			// missing UsageID
			args := &utils.CGREvent{
				ID: "EventAuthorizeResource",
				Event: map[string]any{
					utils.AccountField:          "1001",
					utils.OptsResourcesUnits:    5,
					utils.OptsResourcesUsageTTL: time.Minute,
				},
			}
			experr := `MANDATORY_IE_MISSING: [UsageID]`
			if err := m.call(rS, context.Background(), args, &reply); err == nil ||
				err.Error() != experr {
				t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
			}

			// missing Event
			args = &utils.CGREvent{
				ID: "EventAuthorizeResource",
				APIOpts: map[string]any{
					utils.OptsResourcesUsageID:  "RU_Test",
					utils.OptsResourcesUnits:    5,
					utils.OptsResourcesUsageTTL: time.Minute,
				},
			}
			experr = `MANDATORY_IE_MISSING: [Event]`
			if err := m.call(rS, context.Background(), args, &reply); err == nil ||
				err.Error() != experr {
				t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
			}

			// missing ID
			args = &utils.CGREvent{
				Event: map[string]any{utils.AccountField: "1001"},
				APIOpts: map[string]any{
					utils.OptsResourcesUsageID:  "RU_Test",
					utils.OptsResourcesUnits:    5,
					utils.OptsResourcesUsageTTL: time.Minute,
				},
			}
			experr = `MANDATORY_IE_MISSING: [ID]`
			if err := m.call(rS, context.Background(), args, &reply); err == nil ||
				err.Error() != experr {
				t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
			}

			// nil args
			experr = `MANDATORY_IE_MISSING: [Event]`
			if err := m.call(rS, context.Background(), nil, &reply); err == nil ||
				err.Error() != experr {
				t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
			}
		})
	}
}

func TestResourcesV1AllocateResources(t *testing.T) {
	tests := []struct {
		name              string
		profileID         string
		limit             float64
		allocationMessage string
		eventAccount      string
		wantErr           string
		wantReply         string
	}{
		{
			name:              "OK",
			profileID:         "RES1",
			limit:             10,
			allocationMessage: "Approved",
			eventAccount:      "1001",
			wantReply:         "Approved",
		},
		{
			name:              "NoMatch",
			profileID:         "RES1",
			limit:             10,
			allocationMessage: "Approved",
			eventAccount:      "1002",
			wantErr:           utils.ErrNotFound.Error(),
		},
		{
			name:         "ResAllocErr",
			limit:        -1,
			eventAccount: "1001",
			wantErr:      utils.ErrResourceUnavailable.Error(),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rS, dm := newTestResourceSWithCache(t)
			rsPrf := &utils.ResourceProfile{
				Tenant:            "cgrates.org",
				ID:                tc.profileID,
				FilterIDs:         []string{"*string:~*req.Account:1001"},
				ThresholdIDs:      []string{utils.MetaNone},
				AllocationMessage: tc.allocationMessage,
				Weights:           utils.DynamicWeights{{Weight: 10}},
				Limit:             tc.limit,
				UsageTTL:          time.Minute,
			}
			if err := dm.SetResourceProfile(context.Background(), rsPrf, true); err != nil {
				t.Fatal(err)
			}

			args := &utils.CGREvent{
				ID:    "EventAuthorizeResource",
				Event: map[string]any{utils.AccountField: tc.eventAccount},
				APIOpts: map[string]any{
					utils.OptsResourcesUsageID:  "RU_Test",
					utils.OptsResourcesUnits:    5,
					utils.OptsResourcesUsageTTL: time.Minute,
				},
			}
			var reply string
			err := rS.V1AllocateResources(context.Background(), args, &reply)
			if tc.wantErr != "" {
				if err == nil || err.Error() != tc.wantErr {
					t.Errorf("expected: <%+v>, \nreceived: <%+v>", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if reply != tc.wantReply {
				t.Errorf("expected: %q, received: %q", tc.wantReply, reply)
			}
		})
	}
}

type ccMock struct {
	calls map[string]func(ctx *context.Context, args any, reply any) error
}

func (ccM *ccMock) Call(ctx *context.Context, serviceMethod string, args any, reply any) (err error) {
	if call, has := ccM.calls[serviceMethod]; !has {
		return rpcclient.ErrUnsupporteServiceMethod
	} else {
		return call(ctx, args, reply)
	}
}

func TestResourcesV1AllocateResourcesProcessThErr(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.ResourceSCfg().StoreInterval = 2
	cfg.ResourceSCfg().Conns[utils.MetaThresholds] = []*config.DynamicConns{{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}}}
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)

	rsPrf := &utils.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)},
		AllocationMessage: "Approved",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Limit:    -1,
		UsageTTL: time.Minute,
	}
	rs := &utils.Resource{
		Tenant: "cgrates.org",
		ID:     "RES1",
		Usages: map[string]*utils.ResourceUsage{
			"RU1": {
				Tenant: "cgrates.org",
				ID:     "RU1",
				Units:  10,
			},
		},
		TTLIdx: []string{},
	}

	err := dm.SetResourceProfile(context.Background(), rsPrf, true)
	if err != nil {
		t.Error(err)
	}
	err = dm.SetResource(context.Background(), rs)
	if err != nil {
		t.Error(err)
	}

	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ThresholdSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				return utils.ErrExists
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	cM := engine.NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), utils.ThresholdSv1, rpcInternal)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, cM)

	args := &utils.CGREvent{
		ID: "EventAuthorizeResource",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID:  "RU_Test",
			utils.OptsResourcesUnits:    5,
			utils.OptsResourcesUsageTTL: time.Minute,
		},
	}
	var reply string

	if err := rS.V1AllocateResources(context.Background(), args, &reply); err == nil ||
		err != utils.ErrPartiallyExecuted {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrPartiallyExecuted, err)
	}
	dm.DataDB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestResourcesV1ReleaseResources(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		rS, dm := newTestResourceSWithCache(t)
		rsPrf := &utils.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                "RES1",
			FilterIDs:         []string{"*string:~*req.Account:1001"},
			ThresholdIDs:      []string{utils.MetaNone},
			AllocationMessage: "Approved",
			Weights:           utils.DynamicWeights{{Weight: 10}},
			Limit:             10,
			UsageTTL:          time.Minute,
		}
		if err := dm.SetResourceProfile(context.Background(), rsPrf, true); err != nil {
			t.Fatal(err)
		}

		args := &utils.CGREvent{
			ID:    "EventAuthorizeResource",
			Event: map[string]any{utils.AccountField: "1001"},
			APIOpts: map[string]any{
				utils.OptsResourcesUsageID:  "RU_Test",
				utils.OptsResourcesUnits:    5,
				utils.OptsResourcesUsageTTL: time.Minute,
			},
		}
		var reply string
		if err := rS.V1AllocateResources(context.Background(), args, &reply); err != nil {
			t.Fatal(err)
		}
		if err := rS.V1ReleaseResources(context.Background(), args, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("expected: %q, received: %q", utils.OK, reply)
		}
	})

	t.Run("UsageNotFound", func(t *testing.T) {
		rS, dm := newTestResourceSWithCache(t)
		rsPrf := &utils.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                "RES1",
			FilterIDs:         []string{"*string:~*req.Account:1001"},
			ThresholdIDs:      []string{utils.MetaNone},
			AllocationMessage: "Approved",
			Weights:           utils.DynamicWeights{{Weight: 10}},
			Limit:             10,
			UsageTTL:          0,
		}
		if err := dm.SetResourceProfile(context.Background(), rsPrf, true); err != nil {
			t.Fatal(err)
		}

		allocArgs := &utils.CGREvent{
			ID:    "EventAuthorizeResource",
			Event: map[string]any{utils.AccountField: "1001"},
			APIOpts: map[string]any{
				utils.OptsResourcesUsageID:  "RU_Test",
				utils.OptsResourcesUnits:    5,
				utils.OptsResourcesUsageTTL: time.Minute,
			},
		}
		var reply string
		if err := rS.V1AllocateResources(context.Background(), allocArgs, &reply); err != nil {
			t.Fatal(err)
		}

		releaseArgs := &utils.CGREvent{
			ID:    "EventAuthorizeResource",
			Event: map[string]any{utils.AccountField: "1001"},
			APIOpts: map[string]any{
				utils.OptsResourcesUsageID:  "RU_Test2",
				utils.OptsResourcesUnits:    5,
				utils.OptsResourcesUsageTTL: time.Minute,
			},
		}
		experr := `cannot find usage record with id: RU_Test2`
		if err := rS.V1ReleaseResources(context.Background(), releaseArgs, &reply); err == nil ||
			err.Error() != experr {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
		}
	})

	t.Run("NoMatch", func(t *testing.T) {
		rS, dm := newTestResourceSWithCache(t)
		rsPrf := &utils.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                "RES1",
			FilterIDs:         []string{"*string:~*req.Account:1001"},
			ThresholdIDs:      []string{utils.MetaNone},
			AllocationMessage: "Approved",
			Weights:           utils.DynamicWeights{{Weight: 10}},
			Limit:             10,
			UsageTTL:          time.Minute,
		}
		if err := dm.SetResourceProfile(context.Background(), rsPrf, true); err != nil {
			t.Fatal(err)
		}

		args := &utils.CGREvent{
			ID:    "EventAuthorizeResource",
			Event: map[string]any{utils.AccountField: "1002"},
			APIOpts: map[string]any{
				utils.OptsResourcesUsageID:  "RU_Test",
				utils.OptsResourcesUnits:    5,
				utils.OptsResourcesUsageTTL: time.Minute,
			},
		}
		var reply string
		if err := rS.V1ReleaseResources(context.Background(), args, &reply); err == nil ||
			err != utils.ErrNotFound {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
		}
	})
}

func TestResourcesV1ReleaseResourcesProcessThErr(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.ResourceSCfg().StoreInterval = 2
	cfg.ResourceSCfg().Conns[utils.MetaThresholds] = []*config.DynamicConns{{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}}}
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ThresholdSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				return utils.ErrExists
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	cM := engine.NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), utils.ThresholdSv1, rpcInternal)

	rsPrf := &utils.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)},
		AllocationMessage: "Approved",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Limit:    -1,
		UsageTTL: time.Minute,
	}
	rs := &matchedResource{
		Resource: &utils.Resource{
			Tenant: "cgrates.org",
			ID:     "RES1",
			Usages: map[string]*utils.ResourceUsage{
				"RU_Test": {
					Tenant: "cgrates.org",
					ID:     "RU_Test",
					Units:  4,
				},
			},
			TTLIdx: []string{},
		},
		dirty:      utils.BoolPointer(false),
		totalUsage: utils.Float64Pointer(10),
		ttl:        utils.DurationPointer(time.Minute),
		profile:    rsPrf,
	}

	err := dm.SetResourceProfile(context.Background(), rsPrf, true)
	if err != nil {
		t.Error(err)
	}
	err = dm.SetResource(context.Background(), rs.Resource)
	if err != nil {
		t.Error(err)
	}

	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, cM)

	args := &utils.CGREvent{
		ID: "EventAuthorizeResource",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID:  "RU_Test",
			utils.OptsResourcesUnits:    5,
			utils.OptsResourcesUsageTTL: time.Minute,
		},
	}
	var reply string
	var resources Resources
	resources = append(resources, rs)
	if _, err := resources.allocateResource(&utils.ResourceUsage{
		Tenant: "cgrates.org",
		ID:     "RU_ID",
		Units:  1,
	}, true); err != nil {
		t.Error(err)
	}

	if err := rS.V1ReleaseResources(context.Background(), args, &reply); err == nil ||
		err != utils.ErrPartiallyExecuted {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrPartiallyExecuted, err)
	}

	dm.DataDB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestResourcesStoreResourceError(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.ResourceSCfg().StoreInterval = -1
	cfg.RPCConns()["test"] = &config.RPCConn{
		Conns: []*config.RemoteHost{{}},
	}
	cfg.DbCfg().DBConns[utils.MetaDefault].RplConns = []string{"test"}
	dft := config.CgrConfig()
	config.SetCgrConfig(cfg)
	defer config.SetCgrConfig(dft)

	db, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, engine.NewConnManager(cfg))

	rS := NewResourceService(dm, cfg, engine.NewFilterS(cfg, nil, dm), nil)

	rsPrf := &utils.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Limit:    10,
		UsageTTL: time.Minute,
		Stored:   true,
	}

	err := dm.SetResourceProfile(context.Background(), rsPrf, true)
	if err != nil {
		t.Fatal(err)
	}

	args := &utils.CGREvent{
		ID: "EventAuthorizeResource",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID:  "RU_Test",
			utils.OptsResourcesUnits:    5,
			utils.OptsResourcesUsageTTL: time.Minute,
		},
	}
	cfg.DbCfg().Items[utils.MetaResources].Replicate = true
	var reply string
	if err := rS.V1AllocateResources(context.Background(), args, &reply); err != utils.ErrDisconnected {
		t.Error(err)
	}
	cfg.DbCfg().Items[utils.MetaResources].Replicate = false

	if err := rS.V1AllocateResources(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != "Approved" {
		t.Errorf("Unexpected reply returned: %q", reply)
	}

	cfg.DbCfg().Items[utils.MetaResources].Replicate = true
	if err := rS.V1ReleaseResources(context.Background(), args, &reply); err != utils.ErrDisconnected {
		t.Error(err)
	}
}

func TestErrRetrieveOpts(t *testing.T) {
	args := &utils.CGREvent{
		ID:      "ResourcesForEventTest",
		Event:   map[string]any{utils.AccountField: "1001"},
		APIOpts: map[string]any{},
	}

	setUsageID := func(cfg *config.CGRConfig) {
		cfg.ResourceSCfg().Opts.UsageID = []*config.DynamicStringOpt{
			config.NewDynamicStringOpt([]string{"FLTR_Invalid"}, "*any", "value", nil),
		}
	}
	setUsageTTL := func(cfg *config.CGRConfig) {
		cfg.ResourceSCfg().Opts.UsageTTL = []*config.DynamicDurationOpt{
			config.NewDynamicDurationOpt([]string{"FLTR_Invalid"}, "*any", time.Minute, nil),
		}
	}
	setUnits := func(cfg *config.CGRConfig) {
		cfg.ResourceSCfg().Opts.Units = []*config.DynamicFloat64Opt{
			config.NewDynamicFloat64Opt([]string{"FLTR_Invalid"}, "*any", 3, nil),
		}
	}

	tests := []struct {
		name   string
		setCfg func(*config.CGRConfig)
		call   func(*ResourceS) error
	}{
		{"ResourcesForEvent/UsageID", setUsageID, func(rS *ResourceS) error {
			var reply Resources
			return rS.V1GetResourcesForEvent(context.Background(), args, &reply)
		}},
		{"ResourcesForEvent/UsageTTL", setUsageTTL, func(rS *ResourceS) error {
			var reply Resources
			return rS.V1GetResourcesForEvent(context.Background(), args, &reply)
		}},
		{"AuthorizeResources/UsageID", setUsageID, func(rS *ResourceS) error {
			var reply string
			return rS.V1AuthorizeResources(context.Background(), args, &reply)
		}},
		{"AuthorizeResources/Units", setUnits, func(rS *ResourceS) error {
			var reply string
			return rS.V1AuthorizeResources(context.Background(), args, &reply)
		}},
		{"AuthorizeResources/UsageTTL", setUsageTTL, func(rS *ResourceS) error {
			var reply string
			return rS.V1AuthorizeResources(context.Background(), args, &reply)
		}},
		{"AllocateResources/UsageID", setUsageID, func(rS *ResourceS) error {
			var reply string
			return rS.V1AllocateResources(context.Background(), args, &reply)
		}},
		{"AllocateResources/UsageTTL", setUsageTTL, func(rS *ResourceS) error {
			var reply string
			return rS.V1AllocateResources(context.Background(), args, &reply)
		}},
		{"AllocateResources/Units", setUnits, func(rS *ResourceS) error {
			var reply string
			return rS.V1AllocateResources(context.Background(), args, &reply)
		}},
		{"ReleaseResources/UsageID", setUsageID, func(rS *ResourceS) error {
			var reply string
			return rS.V1ReleaseResources(context.Background(), args, &reply)
		}},
		{"ReleaseResources/UsageTTL", setUsageTTL, func(rS *ResourceS) error {
			var reply string
			return rS.V1ReleaseResources(context.Background(), args, &reply)
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmp := engine.Cache
			t.Cleanup(func() { engine.Cache = tmp })
			engine.Cache.Clear(nil)
			cfg := config.NewDefaultCGRConfig()
			tc.setCfg(cfg)
			data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
			dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
			dm := engine.NewDataManager(dbCM, cfg, nil)
			engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
			fltrs := engine.NewFilterS(cfg, nil, dm)
			rS := NewResourceService(dm, cfg, fltrs, nil)

			experr := `NOT_FOUND:FLTR_Invalid`
			if err := tc.call(rS); err == nil || err.Error() != experr {
				t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
			}
		})
	}
}
