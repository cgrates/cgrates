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

	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dmRES = engine.NewDataManager(data, cfg, nil)
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
			}},
		Limit:        2,
		ThresholdIDs: []string{utils.MetaNone},
		UsageTTL:     -time.Nanosecond,
	}

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	idb, err := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dm := engine.NewDataManager(idb, cfg, nil)
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
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	rsPrf := &utils.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			}},
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
			rPrf: &resourceProfile{ResourceProfile: rsPrf},
			ttl:  utils.DurationPointer(72 * time.Hour),
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
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	rsPrf := &utils.ResourceProfile{
		Tenant:       "cgrates.org",
		ID:           "RES1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		ThresholdIDs: []string{utils.MetaNone},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			}},
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
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	rsPrf := &utils.ResourceProfile{
		Tenant:       "cgrates.org",
		ID:           "RES1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		ThresholdIDs: []string{utils.MetaNone},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			}},
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

	experr := `MANDATORY_IE_MISSING: [Event]`
	var reply Resources
	if err := rS.V1GetResourcesForEvent(context.Background(), nil, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID: "RU_TEST2",
		},
	}

	experr = `MANDATORY_IE_MISSING: [ID]`
	if err := rS.V1GetResourcesForEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	args = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ResourcesForEventTest",
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID: "RU_TEST3",
		},
	}

	experr = `MANDATORY_IE_MISSING: [Event]`
	if err := rS.V1GetResourcesForEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	args = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ResourcesForEventTest",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}

	experr = `MANDATORY_IE_MISSING: [UsageID]`
	if err := rS.V1GetResourcesForEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestResourcesV1ResourcesForEventCacheReplyExists(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	config.SetCgrConfig(cfg)
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	cacheKey := utils.ConcatenatedKey(utils.ResourceSv1GetResourcesForEvent,
		utils.ConcatenatedKey("cgrates.org", "ResourcesForEventTest"))
	rsPrf := &resourceProfile{
		ResourceProfile: &utils.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                "RES1",
			FilterIDs:         []string{"*string:~*req.Account:1001"},
			ThresholdIDs:      []string{utils.MetaNone},
			AllocationMessage: "Approved",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				}},
			Limit:    10,
			UsageTTL: time.Minute,
		},
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
	err := dm.SetResourceProfile(context.Background(), rsPrf.ResourceProfile, true)
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
			rPrf:   rsPrf,
			dirty:  utils.BoolPointer(false),
			tUsage: utils.Float64Pointer(10),
			ttl:    utils.DurationPointer(time.Minute),
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
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	config.SetCgrConfig(cfg)
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
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
			}},
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
			ttl: utils.DurationPointer(72 * time.Hour),
			rPrf: &resourceProfile{
				ResourceProfile: rsPrf,
			},
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

func TestResourcesV1GetResourceOK(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)

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
	err := dm.SetResource(context.Background(), rs)
	if err != nil {
		t.Error(err)
	}

	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	exp := utils.Resource{
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

	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "RES1",
		},
	}
	var reply utils.Resource
	if err := rS.V1GetResource(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp), utils.ToJSON(reply))
	}
}

func TestResourcesV1GetResourceNotFound(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)

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
	err := dm.SetResource(context.Background(), rs)
	if err != nil {
		t.Error(err)
	}

	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "RES2",
		},
	}
	var reply utils.Resource
	if err := rS.V1GetResource(context.Background(), args, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestResourcesV1GetResourceMissingParameters(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)

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
	err := dm.SetResource(context.Background(), rs)
	if err != nil {
		t.Error(err)
	}

	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{},
	}

	experr := `MANDATORY_IE_MISSING: [ID]`
	var reply utils.Resource
	if err := rS.V1GetResource(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestResourcesV1GetResourceWithConfigOK(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)

	rsPrf := &utils.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			}},
		Limit:    10,
		UsageTTL: time.Minute,
	}
	err := dm.SetResourceProfile(context.Background(), rsPrf, true)
	if err != nil {
		t.Error(err)
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
	err = dm.SetResource(context.Background(), rs)
	if err != nil {
		t.Error(err)
	}

	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	exp := utils.ResourceWithConfig{
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
		Config: rsPrf,
	}

	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "RES1",
		},
	}
	var reply utils.ResourceWithConfig
	if err := rS.V1GetResourceWithConfig(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp), utils.ToJSON(reply))
	}
}

func TestResourcesV1GetResourceWithConfigNilrPrfProfileNotFound(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)

	rsPrf := &utils.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES2",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			}},
		Limit:    10,
		UsageTTL: time.Minute,
	}
	err := dm.SetResourceProfile(context.Background(), rsPrf, true)
	if err != nil {
		t.Error(err)
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
	err = dm.SetResource(context.Background(), rs)
	if err != nil {
		t.Error(err)
	}

	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "RES1",
		},
	}
	var reply utils.ResourceWithConfig
	if err := rS.V1GetResourceWithConfig(context.Background(), args, &reply); err == nil ||
		err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestResourcesV1GetResourceWithConfigResourceNotFound(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)

	rs := &utils.Resource{
		Tenant: "cgrates.org",
		ID:     "RES2",
		Usages: map[string]*utils.ResourceUsage{
			"RU1": {
				Tenant: "cgrates.org",
				ID:     "RU1",
				Units:  10,
			},
		},
		TTLIdx: []string{},
	}
	err := dm.SetResource(context.Background(), rs)
	if err != nil {
		t.Error(err)
	}

	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "RES1",
		},
	}
	var reply utils.ResourceWithConfig
	if err := rS.V1GetResourceWithConfig(context.Background(), args, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestResourcesV1GetResourceWithConfigMissingParameters(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)

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
	err := dm.SetResource(context.Background(), rs)
	if err != nil {
		t.Error(err)
	}

	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	experr := `MANDATORY_IE_MISSING: [ID]`
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{},
	}
	var reply utils.ResourceWithConfig
	if err := rS.V1GetResourceWithConfig(context.Background(), args, &reply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestResourcesV1AuthorizeResourcesOK(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)

	rsPrf := &resourceProfile{
		ResourceProfile: &utils.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                "RES1",
			FilterIDs:         []string{"*string:~*req.Account:1001"},
			ThresholdIDs:      []string{utils.MetaNone},
			AllocationMessage: "Approved",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				}},
			Limit:    10,
			UsageTTL: time.Minute,
		},
	}

	err := dm.SetResourceProfile(context.Background(), rsPrf.ResourceProfile, true)
	if err != nil {
		t.Error(err)
	}

	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

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

	if err := rS.V1AuthorizeResources(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != "Approved" {
		t.Errorf("Unexpected reply returned: %q", reply)
	}
}

func TestResourcesV1AuthorizeResourcesNotAuthorized(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)

	rsPrf := &resourceProfile{
		ResourceProfile: &utils.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                "RES1",
			FilterIDs:         []string{"*string:~*req.Account:1001"},
			ThresholdIDs:      []string{utils.MetaNone},
			AllocationMessage: "Approved",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				}},
			Limit:    0,
			UsageTTL: time.Minute,
		},
	}

	err := dm.SetResourceProfile(context.Background(), rsPrf.ResourceProfile, true)
	if err != nil {
		t.Error(err)
	}

	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EventAuthorizeResource",
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

	if err := rS.V1AuthorizeResources(context.Background(), args, &reply); err == nil ||
		err != utils.ErrResourceUnauthorized {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrResourceUnauthorized, err)
	}
}

func TestResourcesV1AuthorizeResourcesNoMatch(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)

	rsPrf := &resourceProfile{
		ResourceProfile: &utils.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                "RES1",
			FilterIDs:         []string{"*string:~*req.Account:1001"},
			ThresholdIDs:      []string{utils.MetaNone},
			AllocationMessage: "Approved",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				}},
			Limit:    10,
			UsageTTL: time.Minute,
		},
	}

	err := dm.SetResourceProfile(context.Background(), rsPrf.ResourceProfile, true)
	if err != nil {
		t.Error(err)
	}

	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EventAuthorizeResource",
		Event: map[string]any{
			utils.AccountField: "1002",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID:  "RU_Test",
			utils.OptsResourcesUnits:    5,
			utils.OptsResourcesUsageTTL: time.Minute,
		},
	}
	var reply string

	if err := rS.V1AuthorizeResources(context.Background(), args, &reply); err == nil ||
		err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestResourcesV1AuthorizeResourcesNilCGREvent(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)

	rsPrf := &resourceProfile{
		ResourceProfile: &utils.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                "RES1",
			FilterIDs:         []string{"*string:~*req.Account:1001"},
			ThresholdIDs:      []string{utils.MetaNone},
			AllocationMessage: "Approved",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				}},
			Limit:    10,
			UsageTTL: time.Minute,
		},
	}

	err := dm.SetResourceProfile(context.Background(), rsPrf.ResourceProfile, true)
	if err != nil {
		t.Error(err)
	}

	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	experr := `MANDATORY_IE_MISSING: [Event]`
	var reply string

	if err := rS.V1AuthorizeResources(context.Background(), nil, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestResourcesV1AuthorizeResourcesMissingUsageID(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)

	rsPrf := &resourceProfile{
		ResourceProfile: &utils.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                "RES1",
			FilterIDs:         []string{"*string:~*req.Account:1001"},
			ThresholdIDs:      []string{utils.MetaNone},
			AllocationMessage: "Approved",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				}},
			Limit:    10,
			UsageTTL: time.Minute,
		},
	}

	err := dm.SetResourceProfile(context.Background(), rsPrf.ResourceProfile, true)
	if err != nil {
		t.Error(err)
	}

	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := &utils.CGREvent{
		ID: "EventAuthorizeResource",
		Event: map[string]any{
			utils.AccountField:          "1001",
			utils.OptsResourcesUnits:    5,
			utils.OptsResourcesUsageTTL: time.Minute,
		},
	}
	experr := `MANDATORY_IE_MISSING: [UsageID]`
	var reply string

	if err := rS.V1AuthorizeResources(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestResourcesV1AuthorizeResourcesCacheReplyExists(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	config.SetCgrConfig(cfg)
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	cacheKey := utils.ConcatenatedKey(utils.ResourceSv1AuthorizeResources,
		utils.ConcatenatedKey("cgrates.org", "EventAuthorizeResource"))

	rsPrf := &utils.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			}},
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

	cacheReply := "Approved"
	engine.Cache.Set(context.Background(), utils.CacheRPCResponses, cacheKey,
		&utils.CachedRPCResponse{Result: &cacheReply, Error: nil},
		nil, true, utils.NonTransactional)

	var reply string
	if err := rS.V1AuthorizeResources(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != cacheReply {
		t.Errorf("Unexpected reply returned: %q", reply)
	}
	config.SetCgrConfig(config.NewDefaultCGRConfig())
}

func TestResourcesV1AuthorizeResourcesCacheReplySet(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	config.SetCgrConfig(cfg)
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	cacheKey := utils.ConcatenatedKey(utils.ResourceSv1AuthorizeResources,
		utils.ConcatenatedKey("cgrates.org", "EventAuthorizeResource"))

	rsPrf := &utils.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			}},
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
				Units:  4,
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
		ID: "EventAuthorizeResource",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID:  "RU_Test",
			utils.OptsResourcesUnits:    2,
			utils.OptsResourcesUsageTTL: time.Minute,
		},
	}

	var reply string
	if err := rS.V1AuthorizeResources(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != "Approved" {
		t.Errorf("Unexpected reply returned: %q", reply)
	}

	if itm, has := engine.Cache.Get(utils.CacheRPCResponses, cacheKey); has {
		resp := itm.(*utils.CachedRPCResponse)
		if *resp.Result.(*string) != "Approved" {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				"Approved", *resp.Result.(*string))
		}
	}

	config.SetCgrConfig(config.NewDefaultCGRConfig())
}

func TestResourcesV1AllocateResourcesOK(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)

	rsPrf := &resourceProfile{
		ResourceProfile: &utils.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                "RES1",
			FilterIDs:         []string{"*string:~*req.Account:1001"},
			ThresholdIDs:      []string{utils.MetaNone},
			AllocationMessage: "Approved",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				}},
			Limit:    10,
			UsageTTL: time.Minute,
		},
	}

	err := dm.SetResourceProfile(context.Background(), rsPrf.ResourceProfile, true)
	if err != nil {
		t.Error(err)
	}

	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

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

	if err := rS.V1AllocateResources(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != "Approved" {
		t.Errorf("Unexpected reply returned: %q", reply)
	}
}

func TestResourcesV1AllocateResourcesNoMatch(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)

	rsPrf := &resourceProfile{
		ResourceProfile: &utils.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                "RES1",
			FilterIDs:         []string{"*string:~*req.Account:1001"},
			ThresholdIDs:      []string{utils.MetaNone},
			AllocationMessage: "Approved",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				}},
			Limit:    10,
			UsageTTL: time.Minute,
		},
	}

	err := dm.SetResourceProfile(context.Background(), rsPrf.ResourceProfile, true)
	if err != nil {
		t.Error(err)
	}

	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := &utils.CGREvent{
		ID: "EventAuthorizeResource",
		Event: map[string]any{
			utils.AccountField: "1002",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID:  "RU_Test",
			utils.OptsResourcesUnits:    5,
			utils.OptsResourcesUsageTTL: time.Minute,
		},
	}
	var reply string

	if err := rS.V1AllocateResources(context.Background(), args, &reply); err == nil ||
		err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestResourcesV1AllocateResourcesMissingParameters(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)

	rsPrf := &utils.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			}},
		Limit:    10,
		UsageTTL: time.Minute,
	}

	err := dm.SetResourceProfile(context.Background(), rsPrf, true)
	if err != nil {
		t.Error(err)
	}

	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := &utils.CGREvent{
		ID: "EventAuthorizeResource",
		Event: map[string]any{
			utils.AccountField:          "1001",
			utils.OptsResourcesUnits:    5,
			utils.OptsResourcesUsageTTL: time.Minute,
		},
	}
	var reply string

	experr := `MANDATORY_IE_MISSING: [UsageID]`
	if err := rS.V1AllocateResources(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	args = &utils.CGREvent{
		ID: "EventAuthorizeResource",
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID:  "RU_Test",
			utils.OptsResourcesUnits:    5,
			utils.OptsResourcesUsageTTL: time.Minute,
		},
	}

	experr = `MANDATORY_IE_MISSING: [Event]`
	if err := rS.V1AllocateResources(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	args = &utils.CGREvent{
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID:  "RU_Test",
			utils.OptsResourcesUnits:    5,
			utils.OptsResourcesUsageTTL: time.Minute,
		},
	}

	experr = `MANDATORY_IE_MISSING: [ID]`
	if err := rS.V1AllocateResources(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	experr = `MANDATORY_IE_MISSING: [Event]`
	if err := rS.V1AllocateResources(context.Background(), nil, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestResourcesV1AllocateResourcesCacheReplyExists(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	config.SetCgrConfig(cfg)
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	cacheKey := utils.ConcatenatedKey(utils.ResourceSv1AllocateResources,
		utils.ConcatenatedKey("cgrates.org", "EventAllocateResource"))

	rsPrf := &utils.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			}},
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

	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := &utils.CGREvent{
		ID: "EventAllocateResource",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID:  "RU_Test",
			utils.OptsResourcesUnits:    5,
			utils.OptsResourcesUsageTTL: time.Minute,
		},
	}

	cacheReply := "cacheApproved"
	engine.Cache.Set(context.Background(), utils.CacheRPCResponses, cacheKey,
		&utils.CachedRPCResponse{Result: &cacheReply, Error: nil},
		nil, true, utils.NonTransactional)

	var reply string
	if err := rS.V1AllocateResources(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != cacheReply {
		t.Errorf("Unexpected reply returned: %q", reply)
	}
	config.SetCgrConfig(config.NewDefaultCGRConfig())
}

func TestResourcesV1AllocateResourcesCacheReplySet(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	config.SetCgrConfig(cfg)
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	cacheKey := utils.ConcatenatedKey(utils.ResourceSv1AllocateResources,
		utils.ConcatenatedKey("cgrates.org", "EventAllocateResource"))

	rsPrf := &utils.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			}},
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
				Units:  4,
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
		ID: "EventAllocateResource",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID:  "RU_Test",
			utils.OptsResourcesUnits:    2,
			utils.OptsResourcesUsageTTL: time.Minute,
		},
	}

	var reply string
	if err := rS.V1AllocateResources(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != "Approved" {
		t.Errorf("Unexpected reply returned: %q", reply)
	}

	if itm, has := engine.Cache.Get(utils.CacheRPCResponses, cacheKey); has {
		resp := itm.(*utils.CachedRPCResponse)
		if *resp.Result.(*string) != "Approved" {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				"Approved", *resp.Result.(*string))
		}
	}

	config.SetCgrConfig(config.NewDefaultCGRConfig())
}

func TestResourcesV1AllocateResourcesResAllocErr(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)

	rsPrf := &utils.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			}},
		Limit:    -1,
		UsageTTL: time.Minute,
	}
	err := dm.SetResourceProfile(context.Background(), rsPrf, true)
	if err != nil {
		t.Error(err)
	}

	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

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
		err != utils.ErrResourceUnavailable {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrResourceUnavailable, err)
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
	cfg.ResourceSCfg().ThresholdSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
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
			}},
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
	dm.DataDB().Flush(utils.EmptyString)
}

func TestResourcesV1ReleaseResourcesOK(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)

	rsPrf := &utils.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			}},
		Limit:    10,
		UsageTTL: time.Minute,
	}

	err := dm.SetResourceProfile(context.Background(), rsPrf, true)
	if err != nil {
		t.Error(err)
	}

	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

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
	if err := rS.V1AllocateResources(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != "Approved" {
		t.Errorf("Unexpected reply returned: %q", reply)
	}

	if err := rS.V1ReleaseResources(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned: %q", reply)
	}
}

func TestResourcesV1ReleaseResourcesUsageNotFound(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)

	rsPrf := &utils.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			}},
		Limit:    10,
		UsageTTL: 0,
	}

	err := dm.SetResourceProfile(context.Background(), rsPrf, true)
	if err != nil {
		t.Error(err)
	}

	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

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
	if err := rS.V1AllocateResources(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != "Approved" {
		t.Errorf("Unexpected reply returned: %q", reply)
	}

	args = &utils.CGREvent{
		ID: "EventAuthorizeResource",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID:  "RU_Test2",
			utils.OptsResourcesUnits:    5,
			utils.OptsResourcesUsageTTL: time.Minute,
		},
	}

	experr := `cannot find usage record with id: RU_Test2`
	if err := rS.V1ReleaseResources(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestResourcesV1ReleaseResourcesNoMatch(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)

	rsPrf := &utils.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			}},
		Limit:    10,
		UsageTTL: time.Minute,
	}

	err := dm.SetResourceProfile(context.Background(), rsPrf, true)
	if err != nil {
		t.Error(err)
	}

	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := &utils.CGREvent{
		ID: "EventAuthorizeResource",
		Event: map[string]any{
			utils.AccountField: "1002",
		},
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
}

func TestResourcesV1ReleaseResourcesMissingParameters(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)

	rsPrf := &utils.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			}},
		Limit:    10,
		UsageTTL: time.Minute,
	}

	err := dm.SetResourceProfile(context.Background(), rsPrf, true)
	if err != nil {
		t.Error(err)
	}

	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := &utils.CGREvent{
		ID: "EventAuthorizeResource",
		Event: map[string]any{
			utils.AccountField:          "1001",
			utils.OptsResourcesUnits:    5,
			utils.OptsResourcesUsageTTL: time.Minute,
		},
	}
	var reply string

	experr := `MANDATORY_IE_MISSING: [UsageID]`
	if err := rS.V1ReleaseResources(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	args = &utils.CGREvent{
		ID: "EventAuthorizeResource",
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID:  "RU_Test",
			utils.OptsResourcesUnits:    5,
			utils.OptsResourcesUsageTTL: time.Minute,
		},
	}

	experr = `MANDATORY_IE_MISSING: [Event]`
	if err := rS.V1ReleaseResources(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	args = &utils.CGREvent{
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID:  "RU_Test",
			utils.OptsResourcesUnits:    5,
			utils.OptsResourcesUsageTTL: time.Minute,
		},
	}

	experr = `MANDATORY_IE_MISSING: [ID]`
	if err := rS.V1ReleaseResources(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	experr = `MANDATORY_IE_MISSING: [Event]`
	if err := rS.V1ReleaseResources(context.Background(), nil, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestResourcesV1ReleaseResourcesCacheReplyExists(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	config.SetCgrConfig(cfg)
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	cacheKey := utils.ConcatenatedKey(utils.ResourceSv1ReleaseResources,
		utils.ConcatenatedKey("cgrates.org", "EventReleaseResource"))

	rsPrf := &utils.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			}},
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

	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EventReleaseResource",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID:  "RU_Test",
			utils.OptsResourcesUnits:    5,
			utils.OptsResourcesUsageTTL: time.Minute,
		},
	}
	cacheReply := "cacheReply"
	engine.Cache.Set(context.Background(), utils.CacheRPCResponses, cacheKey,
		&utils.CachedRPCResponse{Result: &cacheReply, Error: nil},
		nil, true, utils.NonTransactional)

	var reply string
	if err := rS.V1ReleaseResources(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != cacheReply {
		t.Errorf("Unexpected reply returned: %q", reply)
	}
	config.SetCgrConfig(config.NewDefaultCGRConfig())
}

func TestResourcesV1ReleaseResourcesCacheReplySet(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	config.SetCgrConfig(cfg)
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	cacheKey := utils.ConcatenatedKey(utils.ResourceSv1ReleaseResources,
		utils.ConcatenatedKey("cgrates.org", "EventReleaseResource"))

	rsPrf := &utils.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			}},
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
				Units:  4,
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
		ID: "EventReleaseResource",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID:  "RU_Test",
			utils.OptsResourcesUnits:    2,
			utils.OptsResourcesUsageTTL: time.Minute,
		},
	}

	var reply string
	experr := `cannot find usage record with id: RU_Test`
	if err := rS.V1ReleaseResources(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if itm, has := engine.Cache.Get(utils.CacheRPCResponses, cacheKey); has {
		resp := itm.(*utils.CachedRPCResponse)
		if *resp.Result.(*string) != "" {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				"", *resp.Result.(*string))
		}
	}

	config.SetCgrConfig(config.NewDefaultCGRConfig())
}

func TestResourcesV1ReleaseResourcesProcessThErr(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.ResourceSCfg().StoreInterval = 2
	cfg.ResourceSCfg().ThresholdSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
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

	rsPrf := &resourceProfile{
		ResourceProfile: &utils.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                "RES1",
			FilterIDs:         []string{"*string:~*req.Account:1001"},
			ThresholdIDs:      []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)},
			AllocationMessage: "Approved",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				}},
			Limit:    -1,
			UsageTTL: time.Minute,
		},
	}
	rs := &resource{
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
		dirty:  utils.BoolPointer(false),
		tUsage: utils.Float64Pointer(10),
		ttl:    utils.DurationPointer(time.Minute),
		rPrf:   rsPrf,
	}

	err := dm.SetResourceProfile(context.Background(), rsPrf.ResourceProfile, true)
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
		Units:  1}, true); err != nil {
		t.Error(err)
	}

	if err := rS.V1ReleaseResources(context.Background(), args, &reply); err == nil ||
		err != utils.ErrPartiallyExecuted {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrPartiallyExecuted, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestResourcesStoreResourceError(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.ResourceSCfg().StoreInterval = -1
	cfg.RPCConns()["test"] = &config.RPCConn{
		Conns: []*config.RemoteHost{{}},
	}
	cfg.DataDbCfg().RplConns = []string{"test"}
	dft := config.CgrConfig()
	config.SetCgrConfig(cfg)
	defer config.SetCgrConfig(dft)

	db, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(db, cfg, engine.NewConnManager(cfg))

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
			}},
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
	cfg.DataDbCfg().Items[utils.MetaResources].Replicate = true
	var reply string
	if err := rS.V1AllocateResources(context.Background(), args, &reply); err != utils.ErrDisconnected {
		t.Error(err)
	}
	cfg.DataDbCfg().Items[utils.MetaResources].Replicate = false

	if err := rS.V1AllocateResources(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != "Approved" {
		t.Errorf("Unexpected reply returned: %q", reply)
	}

	cfg.DataDbCfg().Items[utils.MetaResources].Replicate = true
	if err := rS.V1ReleaseResources(context.Background(), args, &reply); err != utils.ErrDisconnected {
		t.Error(err)
	}
}

func TestResourcesV1ResourcesForEventErrRetrieveUsageID(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.ResourceSCfg().Opts.UsageID = []*config.DynamicStringOpt{
		config.NewDynamicStringOpt([]string{"FLTR_Invalid"}, "*any", "value", nil),
	}
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := &utils.CGREvent{
		ID: "ResourcesForEventTest",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{},
	}

	experr := `NOT_FOUND:FLTR_Invalid`
	var reply Resources
	if err := rS.V1GetResourcesForEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestResourcesV1ResourcesForEventErrRetrieveUsageTTL(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.ResourceSCfg().Opts.UsageTTL = []*config.DynamicDurationOpt{
		config.NewDynamicDurationOpt([]string{"FLTR_Invalid"}, "*any", time.Minute, nil),
	}
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := &utils.CGREvent{
		ID: "ResourcesForEventTest",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{},
	}

	experr := `NOT_FOUND:FLTR_Invalid`
	var reply Resources
	if err := rS.V1GetResourcesForEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestResourcesV1AuthorizeResourcesErrRetrieveUsageID(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.ResourceSCfg().Opts.UsageID = []*config.DynamicStringOpt{
		config.NewDynamicStringOpt([]string{"FLTR_Invalid"}, "*any", "value", nil),
	}
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := &utils.CGREvent{
		ID: "ResourcesForEventTest",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{},
	}

	experr := `NOT_FOUND:FLTR_Invalid`
	var reply string
	if err := rS.V1AuthorizeResources(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestResourcesV1AuthorizeResourcesErrRetrieveUnits(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.ResourceSCfg().Opts.Units = []*config.DynamicFloat64Opt{
		config.NewDynamicFloat64Opt([]string{"FLTR_Invalid"}, "*any", 3, nil),
	}
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := &utils.CGREvent{
		ID: "ResourcesForEventTest",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{},
	}

	experr := `NOT_FOUND:FLTR_Invalid`
	var reply string
	if err := rS.V1AuthorizeResources(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestResourcesV1AuthorizeResourcesErrRetrieveUsageTTL(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.ResourceSCfg().Opts.UsageTTL = []*config.DynamicDurationOpt{
		config.NewDynamicDurationOpt([]string{"FLTR_Invalid"}, "*any", time.Minute, nil),
	}
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := &utils.CGREvent{
		ID: "ResourcesForEventTest",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{},
	}

	experr := `NOT_FOUND:FLTR_Invalid`
	var reply string
	if err := rS.V1AuthorizeResources(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestResourcesV1AllocateResourcesErrRetrieveUsageID(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.ResourceSCfg().Opts.UsageID = []*config.DynamicStringOpt{
		config.NewDynamicStringOpt([]string{"FLTR_Invalid"}, "*any", "value", nil),
	}
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := &utils.CGREvent{
		ID: "ResourcesForEventTest",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{},
	}

	experr := `NOT_FOUND:FLTR_Invalid`
	var reply string
	if err := rS.V1AllocateResources(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestResourcesV1AllocateResourcesErrRetrieveUsageTTL(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.ResourceSCfg().Opts.UsageTTL = []*config.DynamicDurationOpt{
		config.NewDynamicDurationOpt([]string{"FLTR_Invalid"}, "*any", time.Minute, nil),
	}
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := &utils.CGREvent{
		ID: "ResourcesForEventTest",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{},
	}

	experr := `NOT_FOUND:FLTR_Invalid`
	var reply string
	if err := rS.V1AllocateResources(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestResourcesV1AllocateResourcesErrRetrieveUnits(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.ResourceSCfg().Opts.Units = []*config.DynamicFloat64Opt{
		config.NewDynamicFloat64Opt([]string{"FLTR_Invalid"}, "*any", 3, nil),
	}
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := &utils.CGREvent{
		ID: "ResourcesForEventTest",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{},
	}

	experr := `NOT_FOUND:FLTR_Invalid`
	var reply string
	if err := rS.V1AllocateResources(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestResourcesV1ReleaseResourcesErrRetrieveUsageID(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.ResourceSCfg().Opts.UsageID = []*config.DynamicStringOpt{
		config.NewDynamicStringOpt([]string{"FLTR_Invalid"}, "*any", "value", nil),
	}
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := &utils.CGREvent{
		ID: "ResourcesForEventTest",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{},
	}

	experr := `NOT_FOUND:FLTR_Invalid`
	var reply string
	if err := rS.V1ReleaseResources(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestResourcesV1ReleaseResourcesErrRetrieveUsageTTL(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.ResourceSCfg().Opts.UsageTTL = []*config.DynamicDurationOpt{
		config.NewDynamicDurationOpt([]string{"FLTR_Invalid"}, "*any", time.Minute, nil),
	}
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := &utils.CGREvent{
		ID: "ResourcesForEventTest",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{},
	}

	experr := `NOT_FOUND:FLTR_Invalid`
	var reply string
	if err := rS.V1ReleaseResources(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}
