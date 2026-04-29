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
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/rpcclient"

	"github.com/cgrates/cgrates/utils"
)

func newTestResourceS(t *testing.T) (*ResourceS, *engine.DataManager) {
	t.Helper()
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cfg.ResourceSCfg().StoreInterval = 1
	cfg.ResourceSCfg().StringIndexedFields = nil
	cfg.ResourceSCfg().PrefixIndexedFields = nil
	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(cfg, dm, fltrs, nil)
	return rS, dm
}

func newTestResourceSWithCache(t *testing.T) (*ResourceS, *engine.DataManager) {
	t.Helper()
	tmp := engine.Cache
	t.Cleanup(func() { engine.Cache = tmp })
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(cfg, dm, fltrs, nil)
	return rS, dm
}

func TestResourcesRecordUsage(t *testing.T) {
	testStruct := &matchedResource{
		Resource: &utils.Resource{
			Tenant: "test_tenant",
			ID:     "test_id",
			Usages: map[string]*utils.ResourceUsage{
				"test_id2": {
					Tenant:     "test_tenant2",
					ID:         "test_id2",
					ExpiryTime: time.Date(2015, 1, 14, 0, 0, 0, 0, time.UTC),
					Units:      1,
				},
			},
		},
	}
	recordStruct := &utils.ResourceUsage{
		Tenant:     "test_tenant3",
		ID:         "test_id3",
		ExpiryTime: time.Date(2016, 1, 14, 0, 0, 0, 0, time.UTC),
		Units:      1,
	}
	expStruct := matchedResource{
		Resource: &utils.Resource{
			Tenant: "test_tenant",
			ID:     "test_id",
			Usages: map[string]*utils.ResourceUsage{
				"test_id2": {
					Tenant:     "test_tenant2",
					ID:         "test_id2",
					ExpiryTime: time.Date(2015, 1, 14, 0, 0, 0, 0, time.UTC),
					Units:      1,
				},
				"test_id3": {
					Tenant:     "test_tenant3",
					ID:         "test_id3",
					ExpiryTime: time.Date(2016, 1, 14, 0, 0, 0, 0, time.UTC),
					Units:      1,
				},
			},
		},
	}
	err := testStruct.recordUsage(recordStruct)
	if err != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	if reflect.DeepEqual(testStruct, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expStruct, testStruct)
	}
}

func TestResourcesClearUsage(t *testing.T) {
	testStruct := &matchedResource{
		Resource: &utils.Resource{
			Tenant: "test_tenant",
			ID:     "test_id",
			Usages: map[string]*utils.ResourceUsage{
				"test_id2": {
					Tenant:     "test_tenant2",
					ID:         "test_id2",
					ExpiryTime: time.Date(2015, 1, 14, 0, 0, 0, 0, time.UTC),
					Units:      1,
				},
				"test_id3": {
					Tenant:     "test_tenant3",
					ID:         "test_id3",
					ExpiryTime: time.Date(2016, 1, 14, 0, 0, 0, 0, time.UTC),
					Units:      1,
				},
			},
		},
	}
	expStruct := matchedResource{
		Resource: &utils.Resource{
			Tenant: "test_tenant",
			ID:     "test_id",
			Usages: map[string]*utils.ResourceUsage{
				"test_id2": {
					Tenant:     "test_tenant2",
					ID:         "test_id2",
					ExpiryTime: time.Date(2015, 1, 14, 0, 0, 0, 0, time.UTC),
					Units:      1,
				},
			},
		},
	}
	err := testStruct.clearUsage("test_id3")
	if err != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	if reflect.DeepEqual(testStruct, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expStruct, testStruct)
	}
}

func TestResourceRecordUsage(t *testing.T) {
	var r1 *matchedResource
	var ru1 *utils.ResourceUsage
	var ru2 *utils.ResourceUsage
	ru1 = &utils.ResourceUsage{
		Tenant:     "cgrates.org",
		ID:         "RU1",
		ExpiryTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
		Units:      1,
	}

	ru2 = &utils.ResourceUsage{
		Tenant:     "cgrates.org",
		ID:         "RU2",
		ExpiryTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
		Units:      2,
	}

	r1 = &matchedResource{
		Resource: &utils.Resource{
			Tenant: "cgrates.org",
			ID:     "RL1",
			Usages: map[string]*utils.ResourceUsage{
				ru1.ID: ru1,
			},
			TTLIdx: []string{ru1.ID},
		},
		profile: &utils.ResourceProfile{
			Tenant:    "cgrates.org",
			ID:        "RL1",
			FilterIDs: []string{"FLTR_RES_RL1", "*ai:~*req.AnswerTime:2014-07-03T13:43:00Z|2014-07-03T13:44:00Z"},
			Weights: utils.DynamicWeights{
				{
					Weight: 100,
				}},
			Limit:        2,
			ThresholdIDs: []string{"TEST_ACTIONS"},

			UsageTTL:          time.Millisecond,
			AllocationMessage: "ALLOC",
		},
		totalUsage: utils.Float64Pointer(2),
	}

	if err := r1.recordUsage(ru2); err != nil {
		t.Error(err.Error())
	} else {
		if err := r1.recordUsage(ru1); err == nil {
			t.Error("duplicate ResourceUsage id should not be allowed")
		}
		if _, found := r1.Resource.Usages[ru2.ID]; !found {
			t.Error("ResourceUsage was not recorded")
		}
		if *r1.totalUsage != 4 {
			t.Errorf("expecting: %+v, received: %+v", 4, r1.totalUsage)
		}
	}
}

func TestResourceRemoveExpiredUnits(t *testing.T) {
	var r1 *matchedResource
	var ru1 *utils.ResourceUsage
	var ru2 *utils.ResourceUsage
	ru1 = &utils.ResourceUsage{
		Tenant:     "cgrates.org",
		ID:         "RU1",
		ExpiryTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
		Units:      1,
	}

	ru2 = &utils.ResourceUsage{
		Tenant:     "cgrates.org",
		ID:         "RU2",
		ExpiryTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
		Units:      2,
	}

	r1 = &matchedResource{
		Resource: &utils.Resource{
			Tenant: "cgrates.org",
			ID:     "RL1",
			Usages: map[string]*utils.ResourceUsage{
				ru1.ID: ru1,
			},
			TTLIdx: []string{ru1.ID},
		},
		profile: &utils.ResourceProfile{
			Tenant:    "cgrates.org",
			ID:        "RL1",
			FilterIDs: []string{"FLTR_RES_RL1", "*ai:~*req.AnswerTime:2014-07-03T13:43:00Z|2014-07-03T13:44:00Z"},
			Weights: utils.DynamicWeights{
				{
					Weight: 100,
				}},
			Limit:        2,
			ThresholdIDs: []string{"TEST_ACTIONS"},

			UsageTTL:          time.Millisecond,
			AllocationMessage: "ALLOC",
		},
		totalUsage: utils.Float64Pointer(2),
	}

	if err := r1.recordUsage(ru2); err != nil {
		t.Error(err.Error())
	} else {
		if err := r1.recordUsage(ru1); err == nil {
			t.Error("duplicate ResourceUsage id should not be allowed")
		}
		if _, found := r1.Resource.Usages[ru2.ID]; !found {
			t.Error("ResourceUsage was not recorded")
		}
		if *r1.totalUsage != 4 {
			t.Errorf("expecting: %+v, received: %+v", 4, r1.totalUsage)
		}
	}
	r1.Resource.Usages = map[string]*utils.ResourceUsage{
		ru1.ID: ru1,
	}
	*r1.totalUsage = 2

	r1.removeExpiredUnits()

	if len(r1.Resource.Usages) != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, len(r1.Resource.Usages))
	}
	if len(r1.Resource.TTLIdx) != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, len(r1.Resource.TTLIdx))
	}
	if r1.totalUsage != nil && *r1.totalUsage != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, r1.totalUsage)
	}
}
func TestResourceUsedUnits(t *testing.T) {
	ru1 := &utils.ResourceUsage{
		Tenant:     "cgrates.org",
		ID:         "RU1",
		ExpiryTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
		Units:      1,
	}

	ru2 := &utils.ResourceUsage{
		Tenant:     "cgrates.org",
		ID:         "RU2",
		ExpiryTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
		Units:      2,
	}

	r1 := &matchedResource{
		Resource: &utils.Resource{
			Tenant: "cgrates.org",
			ID:     "RL1",
			Usages: map[string]*utils.ResourceUsage{
				ru1.ID: ru1,
			},
			TTLIdx: []string{ru1.ID},
		},
		profile: &utils.ResourceProfile{
			Tenant:    "cgrates.org",
			ID:        "RL1",
			FilterIDs: []string{"FLTR_RES_RL1", "*ai:~*req.AnswerTime:2014-07-03T13:43:00Z|2014-07-03T13:44:00Z"},
			Weights: utils.DynamicWeights{
				{
					Weight: 100,
				}},
			Limit:        2,
			ThresholdIDs: []string{"TEST_ACTIONS"},

			UsageTTL:          time.Millisecond,
			AllocationMessage: "ALLOC",
		},
		totalUsage: utils.Float64Pointer(2),
	}

	if err := r1.recordUsage(ru2); err != nil {
		t.Error(err.Error())
	} else {
		if err := r1.recordUsage(ru1); err == nil {
			t.Error("duplicate ResourceUsage id should not be allowed")
		}
		if _, found := r1.Resource.Usages[ru2.ID]; !found {
			t.Error("ResourceUsage was not recorded")
		}
		if *r1.totalUsage != 4 {
			t.Errorf("expecting: %+v, received: %+v", 4, r1.totalUsage)
		}
	}
	r1.Resource.Usages = map[string]*utils.ResourceUsage{
		ru1.ID: ru1,
	}
	r1.totalUsage = nil
	if usedUnits := r1.Resource.TotalUsage(); usedUnits != 1 {
		t.Errorf("Expecting: %+v, received: %+v", 1, usedUnits)
	}
}

func TestResourceClearUsage(t *testing.T) {
	ru1 := &utils.ResourceUsage{
		Tenant:     "cgrates.org",
		ID:         "RU1",
		ExpiryTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
		Units:      1,
	}

	ru2 := &utils.ResourceUsage{
		Tenant:     "cgrates.org",
		ID:         "RU2",
		ExpiryTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
		Units:      2,
	}

	r1 := &matchedResource{
		Resource: &utils.Resource{
			Tenant: "cgrates.org",
			ID:     "RL1",
			Usages: map[string]*utils.ResourceUsage{
				ru1.ID: ru1,
			},
			TTLIdx: []string{ru1.ID},
		},
		profile: &utils.ResourceProfile{
			Tenant:    "cgrates.org",
			ID:        "RL1",
			FilterIDs: []string{"FLTR_RES_RL1", "*ai:~*req.AnswerTime:2014-07-03T13:43:00Z|2014-07-03T13:44:00Z"},
			Weights: utils.DynamicWeights{
				{
					Weight: 100,
				}},
			Limit:        2,
			ThresholdIDs: []string{"TEST_ACTIONS"},

			UsageTTL:          time.Millisecond,
			AllocationMessage: "ALLOC",
		},
		totalUsage: utils.Float64Pointer(2),
	}

	if err := r1.recordUsage(ru2); err != nil {
		t.Error(err.Error())
	} else {
		if err := r1.recordUsage(ru1); err == nil {
			t.Error("duplicate ResourceUsage id should not be allowed")
		}
		if _, found := r1.Resource.Usages[ru2.ID]; !found {
			t.Error("ResourceUsage was not recorded")
		}
		if *r1.totalUsage != 4 {
			t.Errorf("expecting: %+v, received: %+v", 4, r1.totalUsage)
		}
	}
	r2 := &matchedResource{
		Resource: &utils.Resource{
			Tenant: "cgrates.org",
			ID:     "RL2",
			// AllocationMessage: "ALLOC2",
			Usages: map[string]*utils.ResourceUsage{
				ru2.ID: ru2,
			},
		},
		profile: &utils.ResourceProfile{
			ID:        "RL2",
			FilterIDs: []string{"FLTR_RES_RL2", "*ai:~*req.AnswerTime:2014-07-03T13:43:00Z|2014-07-03T13:44:00Z"},
			Weights: utils.DynamicWeights{
				{
					Weight: 50,
				}},
			Limit:        2,
			ThresholdIDs: []string{"TEST_ACTIONS"},
			UsageTTL:     time.Millisecond,
		},
		totalUsage: utils.Float64Pointer(2),
	}

	r1.Resource.Usages = map[string]*utils.ResourceUsage{
		ru1.ID: ru1,
	}
	r1.totalUsage = nil
	r1.clearUsage(ru1.ID)
	if len(r1.Resource.Usages) != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, len(r1.Resource.Usages))
	}
	if r1.Resource.TotalUsage() != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, r1.totalUsage)
	}
	if err := r2.clearUsage(ru2.ID); err != nil {
		t.Error(err)
	} else if len(r2.Resource.Usages) != 0 {
		t.Errorf("Unexpected usages %+v", r2.Resource.Usages)
	} else if *r2.totalUsage != 0 {
		t.Errorf("Unexpected totalUsage %+v", r2.totalUsage)
	}
}
func TestResourceRecordUsages(t *testing.T) {
	ru1 := &utils.ResourceUsage{
		Tenant:     "cgrates.org",
		ID:         "RU1",
		ExpiryTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
		Units:      1,
	}

	ru2 := &utils.ResourceUsage{
		Tenant:     "cgrates.org",
		ID:         "RU2",
		ExpiryTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
		Units:      2,
	}

	r1 := &matchedResource{
		Resource: &utils.Resource{
			Tenant: "cgrates.org",
			ID:     "RL1",
			Usages: map[string]*utils.ResourceUsage{
				ru1.ID: ru1,
			},
			TTLIdx: []string{ru1.ID},
		},
		profile: &utils.ResourceProfile{
			Tenant:    "cgrates.org",
			ID:        "RL1",
			FilterIDs: []string{"FLTR_RES_RL1", "*ai:~*req.AnswerTime:2014-07-03T13:43:00Z|2014-07-03T13:44:00Z"},
			Weights: utils.DynamicWeights{
				{
					Weight: 100,
				}},
			Limit:        2,
			ThresholdIDs: []string{"TEST_ACTIONS"},

			UsageTTL:          time.Millisecond,
			AllocationMessage: "ALLOC",
		},
		totalUsage: utils.Float64Pointer(2),
	}

	if err := r1.recordUsage(ru2); err != nil {
		t.Error(err.Error())
	} else {
		if err := r1.recordUsage(ru1); err == nil {
			t.Error("duplicate ResourceUsage id should not be allowed")
		}
		if _, found := r1.Resource.Usages[ru2.ID]; !found {
			t.Error("ResourceUsage was not recorded")
		}
		if *r1.totalUsage != 4 {
			t.Errorf("expecting: %+v, received: %+v", 4, r1.totalUsage)
		}
	}
	r2 := &matchedResource{
		Resource: &utils.Resource{
			Tenant: "cgrates.org",
			ID:     "RL2",
			// AllocationMessage: "ALLOC2",
			Usages: map[string]*utils.ResourceUsage{
				ru2.ID: ru2,
			},
		},
		profile: &utils.ResourceProfile{
			ID:        "RL2",
			FilterIDs: []string{"FLTR_RES_RL2", "*ai:~*req.AnswerTime:2014-07-03T13:43:00Z|2014-07-03T13:44:00Z"},
			Weights: utils.DynamicWeights{
				{
					Weight: 50,
				}},
			Limit:        2,
			ThresholdIDs: []string{"TEST_ACTIONS"},
			UsageTTL:     time.Millisecond,
		},
		totalUsage: utils.Float64Pointer(2),
	}

	rs := Resources{r2, r1}
	r1.Resource.Usages = map[string]*utils.ResourceUsage{
		ru1.ID: ru1,
	}
	r1.Resource.Usages = map[string]*utils.ResourceUsage{
		ru1.ID: ru1,
	}
	r1.totalUsage = nil
	if err := rs.recordUsage(ru1); err == nil {
		t.Error("should get duplicated error")
	}
}
func TestResourceAllocateResource(t *testing.T) {
	ru1 := &utils.ResourceUsage{
		Tenant:     "cgrates.org",
		ID:         "RU1",
		ExpiryTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
		Units:      1,
	}

	ru2 := &utils.ResourceUsage{
		Tenant:     "cgrates.org",
		ID:         "RU2",
		ExpiryTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
		Units:      2,
	}

	r1 := &matchedResource{
		Resource: &utils.Resource{
			Tenant: "cgrates.org",
			ID:     "RL1",
			Usages: map[string]*utils.ResourceUsage{
				ru1.ID: ru1,
			},
			TTLIdx: []string{ru1.ID},
		},
		profile: &utils.ResourceProfile{
			Tenant:    "cgrates.org",
			ID:        "RL1",
			FilterIDs: []string{"FLTR_RES_RL1", "*ai:~*req.AnswerTime:2014-07-03T13:43:00Z|2014-07-03T13:44:00Z"},
			Weights: utils.DynamicWeights{
				{
					Weight: 100,
				}},
			Limit:        2,
			ThresholdIDs: []string{"TEST_ACTIONS"},

			UsageTTL:          time.Millisecond,
			AllocationMessage: "ALLOC",
		},
		totalUsage: utils.Float64Pointer(2),
	}

	if err := r1.recordUsage(ru2); err != nil {
		t.Error(err.Error())
	} else {
		if err := r1.recordUsage(ru1); err == nil {
			t.Error("duplicate ResourceUsage id should not be allowed")
		}
		if _, found := r1.Resource.Usages[ru2.ID]; !found {
			t.Error("ResourceUsage was not recorded")
		}
		if *r1.totalUsage != 4 {
			t.Errorf("expecting: %+v, received: %+v", 4, r1.totalUsage)
		}
	}
	r2 := &matchedResource{
		Resource: &utils.Resource{
			Tenant: "cgrates.org",
			ID:     "RL2",
			// AllocationMessage: "ALLOC2",
			Usages: map[string]*utils.ResourceUsage{
				ru2.ID: ru2,
			},
		},
		profile: &utils.ResourceProfile{
			ID:        "RL2",
			FilterIDs: []string{"FLTR_RES_RL2", "*ai:~*req.AnswerTime:2014-07-03T13:43:00Z|2014-07-03T13:44:00Z"},
			Weights: utils.DynamicWeights{
				{
					Weight: 50,
				}},
			Limit:        2,
			ThresholdIDs: []string{"TEST_ACTIONS"},
			UsageTTL:     time.Millisecond,
		},
		totalUsage: utils.Float64Pointer(2),
	}

	rs := Resources{r1, r2}
	r1.Resource.Usages = map[string]*utils.ResourceUsage{
		ru1.ID: ru1,
	}
	r1.Resource.Usages = map[string]*utils.ResourceUsage{
		ru1.ID: ru1,
	}
	r1.totalUsage = nil
	if err := rs.recordUsage(ru1); err == nil {
		t.Error("should get duplicated error")
	}
	rs.clearUsage(ru1.ID)
	rs.clearUsage(ru2.ID)
	ru1.ExpiryTime = time.Now().Add(time.Second)
	ru2.ExpiryTime = time.Now().Add(time.Second)
	if alcMessage, err := rs.allocateResource(ru1, false); err != nil {
		t.Error(err.Error())
	} else {
		if alcMessage != "ALLOC" {
			t.Errorf("Wrong allocation message: %v", alcMessage)
		}
	}
	if _, err := rs.allocateResource(ru2, false); err != utils.ErrResourceUnavailable {
		t.Error("Did not receive " + utils.ErrResourceUnavailable.Error() + " error")
	}
	rs[0].profile.Limit = 1
	rs[1].profile.Limit = 4
	if alcMessage, err := rs.allocateResource(ru1, false); err != nil {
		t.Error(err.Error())
	} else {
		if alcMessage != "ALLOC" {
			t.Errorf("Wrong allocation message: %v", alcMessage)
		}
	}

	if alcMessage, err := rs.allocateResource(ru2, false); err != nil {
		t.Error(err.Error())
	} else {
		if alcMessage != "RL2" {
			t.Errorf("Wrong allocation message: %v", alcMessage)
		}
	}

	ru2.Units = 0
	if _, err := rs.allocateResource(ru2, false); err != nil {
		t.Error(err)
	}
}

// TestRSCacheSetGet assurace the presence of private params in cached resource
func TestRSCacheSetGet(t *testing.T) {
	r := &utils.Resource{
		Tenant: "cgrates.org",
		ID:     "RL",
		Usages: map[string]*utils.ResourceUsage{
			"RU2": {
				Tenant:     "cgrates.org",
				ID:         "RU2",
				ExpiryTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
				Units:      2,
			},
		},
	}
	if err := engine.Cache.Set(context.TODO(), utils.CacheResources, r.TenantID(), r, nil, true, ""); err != nil {
		t.Errorf("Expecting: nil, received: %s", err)
	}
	if x, ok := engine.Cache.Get(utils.CacheResources, r.TenantID()); !ok {
		t.Error("not in cache")
	} else if x == nil {
		t.Error("nil resource")
	} else if !reflect.DeepEqual(r, x.(*utils.Resource)) {
		t.Errorf("Expecting: %+v, received: %+v", r, x)
	}
}

func TestResourceAddResourceProfile(t *testing.T) {
	var dmRES *engine.DataManager
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dmRES = engine.NewDataManager(dbCM, cfg, nil)
	cfg.ResourceSCfg().StoreInterval = 1
	cfg.ResourceSCfg().StringIndexedFields = nil
	cfg.ResourceSCfg().PrefixIndexedFields = nil
	fltrRes1 := &engine.Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_RES_1",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Resources",
				Values:  []string{"ResourceProfile1"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req.UsageInterval",
				Values:  []string{time.Second.String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Values:  []string{time.Second.String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"9.0"},
			},
		},
	}
	dmRES.SetFilter(context.Background(), fltrRes1, true)
	fltrRes2 := &engine.Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_RES_2",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Resources",
				Values:  []string{"ResourceProfile2"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req.PddInterval",
				Values:  []string{time.Second.String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Values:  []string{time.Second.String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"15.0"},
			},
		},
	}
	dmRES.SetFilter(context.Background(), fltrRes2, true)
	fltrRes3 := &engine.Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_RES_3",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Resources",
				Values:  []string{"ResourceProfilePrefix"},
			},
		},
	}
	dmRES.SetFilter(context.Background(), fltrRes3, true)
	resprf := []*utils.ResourceProfile{
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile1",
			FilterIDs:         []string{"FLTR_RES_1", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weights: utils.DynamicWeights{
				{
					Weight: 20.00,
				}},
			ThresholdIDs: []string{""},
		},
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile2", // identifier of this resource
			FilterIDs:         []string{"FLTR_RES_2", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weights: utils.DynamicWeights{
				{
					Weight: 20.00,
				}},
			ThresholdIDs: []string{""},
		},
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile3",
			FilterIDs:         []string{"FLTR_RES_3", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weights: utils.DynamicWeights{
				{
					Weight: 20.00,
				}},
			ThresholdIDs: []string{""},
		},
	}
	resourceTest := []*utils.Resource{
		{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     "ResourceProfile1",
			Usages: map[string]*utils.ResourceUsage{},
			TTLIdx: []string{},
		},
		{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     "ResourceProfile2",
			Usages: map[string]*utils.ResourceUsage{},
			TTLIdx: []string{},
		},
		{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     "ResourceProfile3",
			Usages: map[string]*utils.ResourceUsage{},
			TTLIdx: []string{},
		},
	}
	for _, resProfile := range resprf {
		dmRES.SetResourceProfile(context.TODO(), resProfile, true)
	}
	for _, res := range resourceTest {
		dmRES.SetResource(context.TODO(), res)
	}
	//Test each resourceProfile from cache
	for _, resPrf := range resprf {
		if tempRes, err := dmRES.GetResourceProfile(context.TODO(), resPrf.Tenant,
			resPrf.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(resPrf, tempRes) {
			t.Errorf("Expecting: %+v, received: %+v", resPrf, tempRes)
		}
	}
}

func newTestMatchingSetup(t *testing.T) (*ResourceS, []*utils.ResourceProfile, Resources, []*utils.CGREvent) {
	t.Helper()
	rS, dm := newTestResourceS(t)
	engine.Cache.Clear(nil)
	tenant := config.CgrConfig().GeneralCfg().DefaultTenant

	dm.SetFilter(context.Background(), &engine.Filter{
		Tenant: tenant,
		ID:     "FLTR_RES_1",
		Rules: []*engine.FilterRule{
			{Type: utils.MetaString, Element: "~*req.Resources", Values: []string{"ResourceProfile1"}},
			{Type: utils.MetaGreaterOrEqual, Element: "~*req.UsageInterval", Values: []string{time.Second.String()}},
			{Type: utils.MetaGreaterOrEqual, Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage, Values: []string{time.Second.String()}},
			{Type: utils.MetaGreaterOrEqual, Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight, Values: []string{"9.0"}},
		},
	}, true)
	dm.SetFilter(context.Background(), &engine.Filter{
		Tenant: tenant,
		ID:     "FLTR_RES_2",
		Rules: []*engine.FilterRule{
			{Type: utils.MetaString, Element: "~*req.Resources", Values: []string{"ResourceProfile2"}},
			{Type: utils.MetaGreaterOrEqual, Element: "~*req.PddInterval", Values: []string{time.Second.String()}},
			{Type: utils.MetaGreaterOrEqual, Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage, Values: []string{time.Second.String()}},
			{Type: utils.MetaGreaterOrEqual, Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight, Values: []string{"15.0"}},
		},
	}, true)
	dm.SetFilter(context.Background(), &engine.Filter{
		Tenant: tenant,
		ID:     "FLTR_RES_3",
		Rules: []*engine.FilterRule{
			{Type: utils.MetaPrefix, Element: "~*req.Resources", Values: []string{"ResourceProfilePrefix"}},
		},
	}, true)

	prfs := []*utils.ResourceProfile{
		{
			Tenant:            tenant,
			ID:                "ResourceProfile1",
			FilterIDs:         []string{"FLTR_RES_1", "*ai:*now:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10,
			AllocationMessage: "AllocationMessage",
			Weights:           utils.DynamicWeights{{Weight: 20}},
			ThresholdIDs:      []string{""},
		},
		{
			Tenant:            tenant,
			ID:                "ResourceProfile2",
			FilterIDs:         []string{"FLTR_RES_2", "*ai:*now:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10,
			AllocationMessage: "AllocationMessage",
			Weights:           utils.DynamicWeights{{Weight: 20}},
			ThresholdIDs:      []string{""},
		},
		{
			Tenant:            tenant,
			ID:                "ResourceProfile3",
			FilterIDs:         []string{"FLTR_RES_3", "*ai:*now:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10,
			AllocationMessage: "AllocationMessage",
			Weights:           utils.DynamicWeights{{Weight: 20}},
			ThresholdIDs:      []string{""},
		},
	}

	resources := Resources{
		{
			Resource: &utils.Resource{
				Tenant: tenant,
				ID:     "ResourceProfile1",
				Usages: map[string]*utils.ResourceUsage{},
				TTLIdx: []string{},
			},
			profile: prfs[0],
		},
		{
			Resource: &utils.Resource{
				Tenant: tenant,
				ID:     "ResourceProfile2",
				Usages: map[string]*utils.ResourceUsage{},
				TTLIdx: []string{},
			},
			profile: prfs[1],
		},
		{
			Resource: &utils.Resource{
				Tenant: tenant,
				ID:     "ResourceProfile3",
				Usages: map[string]*utils.ResourceUsage{},
				TTLIdx: []string{},
			},
			profile: prfs[2],
		},
	}

	events := []*utils.CGREvent{
		{
			Tenant: tenant,
			ID:     "event1",
			Event: map[string]any{
				"Resources":      "ResourceProfile1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				utils.Weight:     "20.0",
				utils.Usage:      135 * time.Second,
				utils.Cost:       123.0,
			},
		},
		{
			Tenant: tenant,
			ID:     "event2",
			Event: map[string]any{
				"Resources":      "ResourceProfile2",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				utils.Weight:     "15.0",
				utils.Usage:      45 * time.Second,
			},
		},
		{
			Tenant: tenant,
			ID:     "event3",
			Event: map[string]any{
				"Resources": "ResourceProfilePrefix",
				utils.Usage: 30 * time.Second,
			},
		},
	}

	for _, prf := range prfs {
		dm.SetResourceProfile(context.TODO(), prf, true)
	}
	for _, res := range resources {
		dm.SetResource(context.TODO(), res.Resource)
	}
	return rS, prfs, resources, events
}

func TestResourceMatchingResourcesForEvent(t *testing.T) {
	rS, _, resources, events := newTestMatchingSetup(t)
	ttl := 10 * time.Second
	for i, ev := range events {
		mres, unlock, err := rS.matchingResourcesForEvent(context.TODO(), ev.Tenant, ev,
			fmt.Sprintf("TestResourceMatchingResourcesForEvent%d", i+1), &ttl)
		if err != nil {
			t.Errorf("Event %d error: %+v", i, err)
			continue
		}
		unlock()
		if !reflect.DeepEqual(resources[i].Resource.Tenant, mres[0].Resource.Tenant) {
			t.Errorf("Event %d tenant: expecting %+v, received %+v", i, resources[i].Resource.Tenant, mres[0].Resource.Tenant)
		} else if !reflect.DeepEqual(resources[i].Resource.ID, mres[0].Resource.ID) {
			t.Errorf("Event %d ID: expecting %+v, received %+v", i, resources[i].Resource.ID, mres[0].Resource.ID)
		} else if !reflect.DeepEqual(resources[i].profile, mres[0].profile) {
			t.Errorf("Event %d profile: expecting %+v, received %+v", i, resources[i].profile, mres[0].profile)
		}
	}
}

func TestResourceUsageTTL(t *testing.T) {
	cases := []struct {
		name       string
		profileTTL time.Duration
		duration   *time.Duration
		wantTTL    *time.Duration
	}{
		{
			name:       "zero profile TTL with 10s duration",
			profileTTL: 0,
			duration:   utils.DurationPointer(10 * time.Second),
			wantTTL:    utils.DurationPointer(10 * time.Second),
		},
		{
			name:       "zero profile TTL with nil duration",
			profileTTL: 0,
			duration:   nil,
			wantTTL:    utils.DurationPointer(0),
		},
		{
			name:       "zero profile TTL with zero duration",
			profileTTL: 0,
			duration:   utils.DurationPointer(0),
			wantTTL:    nil,
		},
		{
			name:       "nonzero profile TTL with 10s duration",
			profileTTL: 5,
			duration:   utils.DurationPointer(10 * time.Second),
			wantTTL:    utils.DurationPointer(10 * time.Second),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rS, prfs, _, events := newTestMatchingSetup(t)
			prfs[0].UsageTTL = tc.profileTTL
			rS.dm.SetResourceProfile(context.TODO(), prfs[0], true)

			mres, unlock, err := rS.matchingResourcesForEvent(context.TODO(), events[0].Tenant, events[0],
				"TestResourceUsageTTL_"+tc.name, tc.duration)
			if err != nil {
				t.Fatalf("Error: %+v", err)
			}
			unlock()
			if mres[0].Resource.Tenant != prfs[0].Tenant {
				t.Errorf("Tenant: expecting %+v, received %+v", prfs[0].Tenant, mres[0].Resource.Tenant)
			}
			if mres[0].Resource.ID != prfs[0].ID {
				t.Errorf("ID: expecting %+v, received %+v", prfs[0].ID, mres[0].Resource.ID)
			}
			if !reflect.DeepEqual(mres[0].profile, prfs[0]) {
				t.Errorf("profile: expecting %+v, received %+v", prfs[0], mres[0].profile)
			}
			if !reflect.DeepEqual(tc.wantTTL, mres[0].ttl) {
				t.Errorf("TTL: expecting %+v, received %+v", tc.wantTTL, mres[0].ttl)
			}
		})
	}
}

func TestResourceMatchWithIndexFalse(t *testing.T) {
	rS, _, resources, events := newTestMatchingSetup(t)
	rS.cfg.ResourceSCfg().IndexedSelects = false
	ttl := 10 * time.Second
	for i, ev := range events {
		mres, unlock, err := rS.matchingResourcesForEvent(context.TODO(), ev.Tenant, ev,
			fmt.Sprintf("TestResourceMatchWithIndexFalse%d", i+1), &ttl)
		if err != nil {
			t.Errorf("Event %d error: %+v", i, err)
			continue
		}
		unlock()
		if !reflect.DeepEqual(resources[i].Resource.Tenant, mres[0].Resource.Tenant) {
			t.Errorf("Event %d tenant: expecting %+v, received %+v", i, resources[i].Resource.Tenant, mres[0].Resource.Tenant)
		} else if !reflect.DeepEqual(resources[i].Resource.ID, mres[0].Resource.ID) {
			t.Errorf("Event %d ID: expecting %+v, received %+v", i, resources[i].Resource.ID, mres[0].Resource.ID)
		} else if !reflect.DeepEqual(resources[i].profile, mres[0].profile) {
			t.Errorf("Event %d profile: expecting %+v, received %+v", i, resources[i].profile, mres[0].profile)
		}
	}
}

func TestResourceCaching(t *testing.T) {
	//clear the cache
	engine.Cache.Clear(nil)

	// start fresh with new dataManager
	cfg := config.NewDefaultCGRConfig()
	idb, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: idb}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)

	cfg.ResourceSCfg().StoreInterval = 1
	cfg.ResourceSCfg().StringIndexedFields = nil
	cfg.ResourceSCfg().PrefixIndexedFields = nil
	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(cfg, dm,
		fltrs, nil)

	resProf := &utils.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "ResourceProfileCached",
		FilterIDs:         []string{"*string:~*req.Account:1001", "*ai:*now:2014-07-14T14:25:00Z"},
		UsageTTL:          -1,
		Limit:             10.00,
		AllocationMessage: "AllocationMessage",
		Weights: utils.DynamicWeights{
			{
				Weight: 20.00,
			}},
		ThresholdIDs: []string{utils.MetaNone},
	}

	if err := engine.Cache.Set(context.TODO(), utils.CacheResourceProfiles, "cgrates.org:ResourceProfileCached",
		resProf, nil, true, utils.EmptyString); err != nil {
		t.Errorf("Expecting: nil, received: %s", err)
	}

	res := &utils.Resource{Tenant: resProf.Tenant,
		ID:     resProf.ID,
		Usages: make(map[string]*utils.ResourceUsage)}

	if err := engine.Cache.Set(context.TODO(), utils.CacheResources, "cgrates.org:ResourceProfileCached",
		res, nil, true, utils.EmptyString); err != nil {
		t.Errorf("Expecting: nil, received: %s", err)
	}

	resources := Resources{
		{
			Resource: res,
			profile:  resProf,
		},
	}
	if err := engine.Cache.Set(context.TODO(), utils.CacheEventResources, "TestResourceCaching", []string{resProf.ID}, nil, true, ""); err != nil {
		t.Errorf("Expecting: nil, received: %s", err)
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Event: map[string]any{
			"Account":     "1001",
			"Destination": "3002"},
	}

	mres, unlock, err := rS.matchingResourcesForEvent(context.TODO(), ev.Tenant, ev,
		"TestResourceCaching", nil)
	if err != nil {
		t.Fatal(err)
	}
	unlock()
	if !reflect.DeepEqual(resources[0].Resource.Tenant, mres[0].Resource.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", resources[0].Resource.Tenant, mres[0].Resource.Tenant)
	} else if !reflect.DeepEqual(resources[0].Resource.ID, mres[0].Resource.ID) {
		t.Errorf("Expecting: %+v, received: %+v", resources[0].Resource.ID, mres[0].Resource.ID)
	} else if !reflect.DeepEqual(resources[0].profile, mres[0].profile) {
		t.Errorf("Expecting: %+v, received: %+v", resources[0].profile, mres[0].profile)
	} else if !reflect.DeepEqual(resources[0].ttl, mres[0].ttl) {
		t.Errorf("Expecting: %+v, received: %+v", resources[0].ttl, mres[0].ttl)
	}
}

func TestResourcesRemoveExpiredUnitsResetTotalUsage(t *testing.T) {
	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 4)

	r := &matchedResource{
		Resource: &utils.Resource{
			TTLIdx: []string{"ResGroup1", "ResGroup2", "ResGroup3"},
			Usages: map[string]*utils.ResourceUsage{
				"ResGroup2": {
					Tenant:     "cgrates.org",
					ID:         "RU_2",
					Units:      11,
					ExpiryTime: time.Date(2021, 5, 3, 13, 0, 0, 0, time.UTC),
				},
				"ResGroup3": {
					Tenant: "cgrates.org",
					ID:     "RU_3",
				},
			},
		},
		totalUsage: utils.Float64Pointer(10),
	}

	exp := &matchedResource{
		Resource: &utils.Resource{
			TTLIdx: []string{"ResGroup3"},
			Usages: map[string]*utils.ResourceUsage{
				"ResGroup3": {
					Tenant: "cgrates.org",
					ID:     "RU_3",
				},
			},
		},
	}

	explog := "CGRateS <> [WARNING] resetting total usage for resourceID: , usage smaller than 0: -1.000000\n"
	r.removeExpiredUnits()

	if !reflect.DeepEqual(r, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, r)
	}

	rcvlog := buf.String()
	if rcvlog != explog {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", explog, rcvlog)
	}
}

func TestResourcesAvailable(t *testing.T) {
	r := utils.ResourceWithConfig{
		Resource: &utils.Resource{
			Usages: map[string]*utils.ResourceUsage{
				"RU_1": {
					Units: 4,
				},
				"RU_2": {
					Units: 7,
				},
			},
		},
		Config: &utils.ResourceProfile{
			Limit: 10,
		},
	}

	exp := -1.0
	rcv := r.Available()

	if rcv != exp {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestResourcesRecordUsageZeroTTL(t *testing.T) {
	r := &matchedResource{
		Resource: &utils.Resource{
			Usages: map[string]*utils.ResourceUsage{
				"RU_1": {
					Tenant: "cgrates.org",
					ID:     "RU_1",
				},
			},
		},
		ttl: utils.DurationPointer(0),
	}
	ru := &utils.ResourceUsage{
		ID: "RU_2",
	}

	err := r.recordUsage(ru)

	if err != nil {
		t.Error(err)
	}
}

func TestResourcesRecordUsageGtZeroTTL(t *testing.T) {
	r := &matchedResource{
		Resource: &utils.Resource{
			Usages: map[string]*utils.ResourceUsage{
				"RU_1": {
					Tenant: "cgrates.org",
					ID:     "RU_1",
				},
			},
			TTLIdx: []string{"RU_1"},
		},
		ttl: utils.DurationPointer(1 * time.Second),
	}
	ru := &utils.ResourceUsage{
		Tenant: "cgrates.org",
		ID:     "RU_2",
	}

	exp := &matchedResource{
		Resource: &utils.Resource{
			Usages: map[string]*utils.ResourceUsage{
				"RU_1": {
					Tenant: "cgrates.org",
					ID:     "RU_1",
				},
				"RU_2": {
					Tenant: "cgrates.org",
					ID:     "RU_2",
				},
			},
			TTLIdx: []string{"RU_1", "RU_2"},
		},
		ttl: utils.DurationPointer(1 * time.Second),
	}
	err := r.recordUsage(ru)
	exp.Resource.Usages[ru.ID].ExpiryTime = r.Resource.Usages[ru.ID].ExpiryTime

	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(r, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(r))
	}
}

type mockWriter struct {
	WriteF func(p []byte) (n int, err error)
}

func (mW *mockWriter) Write(p []byte) (n int, err error) {
	if mW.WriteF != nil {
		return mW.WriteF(p)
	}
	return 0, nil
}

func TestResourcesRecordUsageClearErr(t *testing.T) {
	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 4)

	rs := Resources{
		{
			Resource: &utils.Resource{
				Usages: map[string]*utils.ResourceUsage{
					"RU_1": {
						Tenant: "cgrates.org",
						ID:     "RU_1",
					},
					"RU_2": {
						Tenant: "cgrates.org",
						ID:     "RU_2",
					},
				},
				TTLIdx: []string{"RU_1", "RU_2"},
			},
			ttl: utils.DurationPointer(1 * time.Second),
		},
		{
			Resource: &utils.Resource{
				Usages: map[string]*utils.ResourceUsage{
					"RU_3": {
						Tenant: "cgrates.org",
						ID:     "RU_3",
					},
					"RU_4": {
						Tenant: "cgrates.org",
						ID:     "RU_4",
					},
				},
				TTLIdx: []string{"RU_3"},
			},
			ttl: utils.DurationPointer(2 * time.Second),
		},
	}

	ru := &utils.ResourceUsage{
		Tenant: "cgrates.org",
		ID:     "RU_4",
	}

	exp := Resources{
		{
			Resource: &utils.Resource{
				Usages: map[string]*utils.ResourceUsage{
					"RU_1": {
						Tenant: "cgrates.org",
						ID:     "RU_1",
					},
					"RU_2": {
						Tenant: "cgrates.org",
						ID:     "RU_2",
					},
				},
				TTLIdx: []string{"RU_1", "RU_2", "RU_4"},
			},
			ttl: utils.DurationPointer(1 * time.Second),
		},
		{
			Resource: &utils.Resource{
				Usages: map[string]*utils.ResourceUsage{
					"RU_3": {
						Tenant: "cgrates.org",
						ID:     "RU_3",
					},
					"RU_4": {
						Tenant: "cgrates.org",
						ID:     "RU_4",
					},
				},
				TTLIdx: []string{"RU_3"},
			},
			ttl: utils.DurationPointer(2 * time.Second),
		},
	}

	explog := []string{
		fmt.Sprintf("CGRateS <> [WARNING] <%s>cannot record usage, err: duplicate resource usage with id: %s:%s", utils.ResourceS, ru.Tenant, ru.ID),
		fmt.Sprintf("CGRateS <> [WARNING] <%s> cannot clear usage, err: cannot find usage record with id: %s", utils.ResourceS, ru.ID),
	}
	experr := fmt.Sprintf("duplicate resource usage with id: %s", "cgrates.org:"+ru.ID)

	utils.Logger = utils.NewStdLoggerWithWriter(&mockWriter{
		WriteF: func(p []byte) (n int, err error) {
			delete(rs[0].Resource.Usages, "RU_4")
			return buf.Write(p)
		},
	}, "", 4)

	err := rs.recordUsage(ru)

	if err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if !reflect.DeepEqual(rs, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(rs))
	}

	rcv := strings.Split(buf.String(), "\n")
	for idx, exp := range explog {
		if rcv[idx] != exp {
			t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv[idx])
		}
	}
}

func TestResourceClearUsageErr(t *testing.T) {
	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 4)
	rs := Resources{
		{
			Resource: &utils.Resource{
				Usages: map[string]*utils.ResourceUsage{
					"RU_1": {
						Tenant: "cgrates.org",
						ID:     "RU_1",
					},
					"RU_2": {
						Tenant: "cgrates.org",
						ID:     "RU_2",
					},
				},
				TTLIdx: []string{"RU_1", "RU_2"},
			},
			ttl: utils.DurationPointer(1 * time.Second),
		},
	}

	ruTntID := "cgrates.org:RU_3"

	experr := fmt.Sprintf("cannot find usage record with id: %s", ruTntID)
	explog := fmt.Sprintf("CGRateS <> [WARNING] <ResourceS>, clear ruID: %s, err: %s\n", ruTntID, experr)
	err := rs.clearUsage(ruTntID)

	if err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	rcvlog := buf.String()
	if rcvlog != explog {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", explog, rcvlog)
	}
}

func TestResourcesAllocateResourceErrRsUnavailable(t *testing.T) {
	rs := Resources{}
	ru := &utils.ResourceUsage{}

	experr := utils.ErrResourceUnavailable
	rcv, err := rs.allocateResource(ru, false)

	if err == nil || !errors.Is(err, experr) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if rcv != "" {
		t.Errorf("\nexpected empty string, got %s", rcv)
	}
}

func TestResourcesAllocateResourceDryRun(t *testing.T) {
	rs := Resources{
		{
			Resource: &utils.Resource{
				Tenant: "cgrates.org",
				ID:     "Res_1",
				Usages: map[string]*utils.ResourceUsage{
					"RU_1": {
						Tenant: "cgrates.org",
						ID:     "RU_1",
					},
					"RU_2": {
						Tenant: "cgrates.org",
						ID:     "RU_2",
					},
				},
				TTLIdx: []string{"RU_1", "RU_2"},
			},
			ttl: utils.DurationPointer(1 * time.Second),
			profile: &utils.ResourceProfile{
				Tenant: "cgrates.org",
				ID:     "ResGroup1",
			},
		},
	}

	ru := &utils.ResourceUsage{
		Tenant: "cgrates.org",
		ID:     "RU_2",
	}

	exp := "ResGroup1"
	rcv, err := rs.allocateResource(ru, true)

	if err != nil {
		t.Errorf("\nexpected nil, got %+v", err)
	}

	if rcv != exp {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestResourcesStoreResources(t *testing.T) {
	tmp := engine.Cache
	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
		engine.Cache = tmp
	}()

	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 6)
	cfg := config.NewDefaultCGRConfig()
	rS := &ResourceS{
		cfg: cfg,
		storedResources: utils.StringSet{
			"Res1": struct{}{},
		},
	}

	value := &utils.Resource{
		Tenant: "cgrates.org",
		ID:     "testResource",
	}

	engine.Cache.SetWithoutReplicate(utils.CacheResources, "Res1", value, nil, true,
		utils.NonTransactional)

	explog := fmt.Sprintf("CGRateS <> [WARNING] <%s> failed saving Resource with ID: %s, error: %s\n",
		utils.ResourceS, value.ID, utils.ErrNoDatabaseConn.Error())
	exp := &ResourceS{
		cfg: cfg,
		storedResources: utils.StringSet{
			"Res1": struct{}{},
		},
	}
	rS.storeResources(context.TODO())

	if !reflect.DeepEqual(rS, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rS)
	}

	rcvlog := buf.String()

	if rcvlog != explog {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", explog, rcvlog)
	}
}

func TestResourcesStoreResourceOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	idb, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: idb}, cfg.DbCfg())
	rS := &ResourceS{
		dm: engine.NewDataManager(dbCM, cfg, nil),
	}
	r := &utils.Resource{}

	if err = rS.storeResource(context.TODO(), r); err != nil {
		t.Errorf("\nexpected nil, received %+v", err)
	}
}

func TestResourcesStoreResourceErrCache(t *testing.T) {
	tmp := engine.Cache
	tmpLogger := utils.Logger

	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 6)
	defer func() {
		engine.Cache = tmp
		utils.Logger = tmpLogger
	}()

	dft := config.CgrConfig()
	defer config.SetCgrConfig(dft)

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{"test"}
	cfg.CacheCfg().Partitions[utils.CacheResources].Replicate = true
	cfg.RPCConns()["test"] = &config.RPCConn{Conns: []*config.RemoteHost{{}}}
	config.SetCgrConfig(cfg)
	idb, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: idb}, cfg.DbCfg())
	cM := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(dbCM, cfg, cM)
	rS := NewResourceService(cfg, dm, nil, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, cM, nil)
	r := &utils.Resource{
		Tenant: "cgrates.org",
		ID:     "RES1",
	}
	engine.Cache.Set(context.Background(), utils.CacheResources, r.TenantID(), r, nil, true, "")

	explog := `CGRateS <> [WARNING] <ResourceS> failed caching Resource with ID: cgrates.org:RES1, error: DISCONNECTED
`
	if err := rS.storeResource(context.Background(), r); err == nil ||
		err.Error() != rpcclient.ErrDisconnected.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", rpcclient.ErrDisconnected, err)
	}

	rcvlog := buf.String()

	if rcvlog != explog {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", explog, rcvlog)
	}
}
func TestResourcesAllocateResourceEmptyKey(t *testing.T) {
	rs := Resources{
		{
			Resource: &utils.Resource{
				Usages: map[string]*utils.ResourceUsage{
					"": {},
				},
			},
			profile: &utils.ResourceProfile{
				Tenant:            "cgrates.org",
				ID:                "RP_1",
				AllocationMessage: "allocation msg",
			},
		},
	}

	ru := &utils.ResourceUsage{}
	exp := "allocation msg"
	rcv, err := rs.allocateResource(ru, false)

	if err != nil {
		t.Errorf("\nexpected nil, received %+v", err)
	}

	if rcv != exp {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestResourcesProcessThresholdsNoConns(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rS := &ResourceS{
		cfg: cfg,
	}
	r := &matchedResource{
		Resource: &utils.Resource{
			Tenant: "cgrates.org",
			ID:     "RES_1",
		},
	}
	opts := map[string]any{}

	err := rS.processThresholds(context.TODO(), Resources{r}, opts)

	if err != nil {
		t.Errorf("\nexpected nil, received %+v", err)
	}
}

func TestResourcesProcessThresholdsOK(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.ResourceSCfg().Conns[utils.MetaThresholds] = []*config.DynamicConns{{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}}}
	engine.Cache = engine.NewCacheS(cfg, nil, nil, nil)

	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ThresholdSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				exp := &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     args.(*utils.CGREvent).ID,
					Event: map[string]any{
						utils.EventType:  utils.ResourceUpdate,
						utils.ResourceID: "RES_1",
						utils.Usage:      0.,
					},
					APIOpts: map[string]any{
						utils.MetaEventType:            utils.ResourceUpdate,
						utils.OptsThresholdsProfileIDs: []string{"THD_1"},
					},
				}
				if !reflect.DeepEqual(exp, args) {
					return fmt.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
						utils.ToJSON(exp), utils.ToJSON(args))
				}
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	rS := &ResourceS{
		cfg: cfg,
		cm:  engine.NewConnManager(cfg),
	}
	rS.cm.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), utils.ThresholdSv1, rpcInternal)
	r := &matchedResource{
		Resource: &utils.Resource{
			Tenant: "cgrates.org",
			ID:     "RES_1",
		},
		profile: &utils.ResourceProfile{
			Tenant:       "cgrates.org",
			ID:           "RP_1",
			ThresholdIDs: []string{"THD_1"},
		},
	}

	err := rS.processThresholds(context.TODO(), Resources{r}, nil)

	if err != nil {
		t.Errorf("\nexpected nil, received %+v", err)
	}

}

func TestResourcesProcessThresholdsCallErr(t *testing.T) {
	tmp := engine.Cache
	tmpLogger := utils.Logger
	defer func() {
		engine.Cache = tmp
		utils.Logger = tmpLogger
	}()

	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 4)

	cfg := config.NewDefaultCGRConfig()
	cfg.ResourceSCfg().Conns[utils.MetaThresholds] = []*config.DynamicConns{{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}}}
	engine.Cache = engine.NewCacheS(cfg, nil, nil, nil)

	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ThresholdSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				exp := &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     args.(*utils.CGREvent).ID,
					Event: map[string]any{
						utils.EventType:  utils.ResourceUpdate,
						utils.ResourceID: "RES_1",
						utils.Usage:      0.,
					},
					APIOpts: map[string]any{
						utils.MetaEventType:            utils.ResourceUpdate,
						utils.OptsThresholdsProfileIDs: []string{"THD_1"},
					},
				}
				if !reflect.DeepEqual(exp, args) {
					return fmt.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
						utils.ToJSON(exp), utils.ToJSON(args))
				}
				return utils.ErrExists
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	rS := &ResourceS{
		cfg: cfg,
		cm:  engine.NewConnManager(cfg),
	}
	rS.cm.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), utils.ThresholdSv1, rpcInternal)
	r := &matchedResource{
		Resource: &utils.Resource{
			Tenant: "cgrates.org",
			ID:     "RES_1",
		},
		profile: &utils.ResourceProfile{
			Tenant:       "cgrates.org",
			ID:           "RP_1",
			ThresholdIDs: []string{"THD_1"},
		},
	}

	experr := utils.ErrPartiallyExecuted
	err := rS.processThresholds(context.TODO(), Resources{r}, nil)

	if err == nil || err != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
	if !strings.Contains(buf.String(), "CGRateS <> [WARNING] <ResourceS> error: EXISTS") {
		t.Errorf("expected log warning")
	}

}

func TestResourcesProcessThresholdsThdConnMetaNone(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ResourceSCfg().Conns[utils.MetaThresholds] = []*config.DynamicConns{{ConnIDs: []string{"connID"}}}
	rS := &ResourceS{
		cfg: cfg,
	}
	r := &matchedResource{
		Resource: &utils.Resource{
			Tenant: "cgrates.org",
			ID:     "RES_1",
		},
		profile: &utils.ResourceProfile{
			ThresholdIDs: []string{utils.MetaNone},
		},
	}
	opts := map[string]any{}

	err := rS.processThresholds(context.TODO(), Resources{r}, opts)

	if err != nil {
		t.Errorf("\nexpected nil, received: %+v", err)
	}
}

func TestResourceMatchingResourcesForEventNotFoundInCache(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	db, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dmRES := engine.NewDataManager(dbCM, cfg, nil)
	cfg.ResourceSCfg().StoreInterval = 1
	cfg.ResourceSCfg().StringIndexedFields = nil
	cfg.ResourceSCfg().PrefixIndexedFields = nil
	fltrs := engine.NewFilterS(cfg, nil, dmRES)
	rS := NewResourceService(cfg, dmRES,
		fltrs, nil)

	engine.Cache.Set(context.Background(), utils.CacheEventResources, "TestResourceMatchingResourcesForEventNotFoundInCache", nil, nil, true, utils.NonTransactional)
	_, _, err := rS.matchingResourcesForEvent(context.Background(), "cgrates.org", new(utils.CGREvent),
		"TestResourceMatchingResourcesForEventNotFoundInCache", utils.DurationPointer(10*time.Second))
	if err != utils.ErrNotFound {
		t.Errorf("Error: %+v", err)
	}
}

func TestResourceMatchingResourcesForEventNotFoundInDB(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	db, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dmRES := engine.NewDataManager(dbCM, cfg, nil)
	cfg.ResourceSCfg().StoreInterval = 1
	cfg.ResourceSCfg().StringIndexedFields = nil
	cfg.ResourceSCfg().PrefixIndexedFields = nil
	fltrs := engine.NewFilterS(cfg, nil, dmRES)
	rS := NewResourceService(cfg, dmRES,
		fltrs, nil)

	engine.Cache.Set(context.Background(), utils.CacheEventResources, "TestResourceMatchingResourcesForEventNotFoundInDB", []string{"Res2"}, nil, true, utils.NonTransactional)
	_, _, err := rS.matchingResourcesForEvent(context.Background(), "cgrates.org", new(utils.CGREvent),
		"TestResourceMatchingResourcesForEventNotFoundInDB", utils.DurationPointer(10*time.Second))
	if err != utils.ErrNotFound {
		t.Errorf("Error: %+v", err)
	}
}

func TestResourceMatchingResourcesForEventLocks(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	db, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cfg.ResourceSCfg().StoreInterval = 1
	cfg.ResourceSCfg().StringIndexedFields = nil
	cfg.ResourceSCfg().PrefixIndexedFields = nil
	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(cfg, dm,
		fltrs, nil)
	engine.Cache.Clear(nil)

	prfs := make([]*utils.ResourceProfile, 0)
	ids := make([]string, 0, 10)
	for i := range 10 {
		rPrf := &utils.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                fmt.Sprintf("RES%d", i),
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weights: utils.DynamicWeights{
				{
					Weight: 20.00,
				}},
			ThresholdIDs: []string{utils.MetaNone},
		}
		dm.SetResourceProfile(context.Background(), rPrf, true)
		prfs = append(prfs, rPrf)
		ids = append(ids, rPrf.ID)
	}
	dm.RemoveResource(context.Background(), "cgrates.org", "RES1")
	engine.Cache.Set(context.Background(), utils.CacheEventResources, "TestResourceMatchingResourcesForEventLocks", ids, nil, true, utils.NonTransactional)
	_, _, err := rS.matchingResourcesForEvent(context.Background(), "cgrates.org", new(utils.CGREvent),
		"TestResourceMatchingResourcesForEventLocks", utils.DurationPointer(10*time.Second))
	if err != utils.ErrNotFound {
		t.Errorf("Error: %+v", err)
	}
}

func TestResourceMatchingResourcesForEventLocks2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	db, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cfg.ResourceSCfg().StoreInterval = 1
	cfg.ResourceSCfg().StringIndexedFields = nil
	cfg.ResourceSCfg().PrefixIndexedFields = nil
	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(cfg, dm,
		fltrs, nil)
	engine.Cache.Clear(nil)

	prfs := make([]*utils.ResourceProfile, 0)
	ids := make([]string, 0, 11)
	for i := range 10 {
		rPrf := &utils.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                fmt.Sprintf("RES%d", i),
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weights: utils.DynamicWeights{
				{
					Weight: 20.00,
				}},
			ThresholdIDs: []string{utils.MetaNone},
		}
		dm.SetResourceProfile(context.Background(), rPrf, true)
		prfs = append(prfs, rPrf)
		ids = append(ids, rPrf.ID)
	}
	rPrf := &utils.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES20",
		FilterIDs:         []string{"FLTR_RES_201"},
		UsageTTL:          10 * time.Second,
		Limit:             10.00,
		AllocationMessage: "AllocationMessage",
		Weights: utils.DynamicWeights{
			{
				Weight: 20.00,
			}},
		ThresholdIDs: []string{utils.MetaNone},
	}
	err := db.SetResourceProfileDrv(context.Background(), rPrf)
	if err != nil {
		t.Fatal(err)
	}
	prfs = append(prfs, rPrf)
	ids = append(ids, rPrf.ID)
	engine.Cache.Set(context.Background(), utils.CacheEventResources, "TestResourceMatchingResourcesForEventLocks2", ids, nil, true, utils.NonTransactional)
	_, _, err = rS.matchingResourcesForEvent(context.Background(), "cgrates.org", new(utils.CGREvent),
		"TestResourceMatchingResourcesForEventLocks2", utils.DurationPointer(10*time.Second))
	expErr := utils.ErrPrefixNotFound(rPrf.FilterIDs[0])
	if err == nil || err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s ,received: %+v", expErr, err)
	}
}

func TestResourceMatchingResourcesForEventLocks3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	prfs := make([]*utils.ResourceProfile, 0)
	engine.Cache.Clear(nil)
	db := &engine.DataDBMock{
		GetResourceProfileDrvF: func(_ *context.Context, tnt, id string) (*utils.ResourceProfile, error) {
			if id == "RES1" {
				return nil, utils.ErrNotImplemented
			}
			rPrf := &utils.ResourceProfile{
				Tenant:            "cgrates.org",
				ID:                id,
				UsageTTL:          10 * time.Second,
				Limit:             10.00,
				AllocationMessage: "AllocationMessage",
				Weights: utils.DynamicWeights{
					{
						Weight: 20.00,
					}},
				ThresholdIDs: []string{utils.MetaNone},
			}
			engine.Cache.Set(context.Background(), utils.CacheResources, rPrf.TenantID(),
				&utils.Resource{
					Tenant: rPrf.Tenant,
					ID:     rPrf.ID,
					Usages: make(map[string]*utils.ResourceUsage),
				}, nil, true, utils.NonTransactional)
			prfs = append(prfs, rPrf)
			return rPrf, nil
		},
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cfg.ResourceSCfg().StoreInterval = 1
	cfg.ResourceSCfg().StringIndexedFields = nil
	cfg.ResourceSCfg().PrefixIndexedFields = nil
	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(cfg, dm,
		fltrs, nil)

	ids := make([]string, 0, 10)
	for i := range 10 {
		ids = append(ids, fmt.Sprintf("RES%d", i))
	}
	engine.Cache.Set(context.Background(), utils.CacheEventResources, "TestResourceMatchingResourcesForEventLocks3", ids, nil, true, utils.NonTransactional)
	_, _, err := rS.matchingResourcesForEvent(context.Background(), "cgrates.org", new(utils.CGREvent),
		"TestResourceMatchingResourcesForEventLocks3", utils.DurationPointer(10*time.Second))
	if err != utils.ErrNotImplemented {
		t.Errorf("Error: %+v", err)
	}
}

func TestResourcesRunBackupStoreIntervalLessThanZero(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ResourceSCfg().StoreInterval = -1
	rS := &ResourceS{
		cfg:         cfg,
		loopStopped: make(chan struct{}, 1),
	}
	rS.runBackup(context.Background())
	select {
	case <-rS.loopStopped:
	case <-time.After(time.Second):
		t.Error("timed out waiting for loop to stop")
	}
}

func TestResourcesRunBackupStop(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ResourceSCfg().StoreInterval = 5 * time.Millisecond
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	tnt := "cgrates.org"
	resID := "Res1"
	rS := &ResourceS{
		dm: dm,
		storedResources: utils.StringSet{
			resID: struct{}{},
		},
		cfg:         cfg,
		loopStopped: make(chan struct{}, 1),
		stopBackup:  make(chan struct{}),
	}
	value := &utils.Resource{
		Tenant: tnt,
		ID:     resID,
	}
	engine.Cache.SetWithoutReplicate(utils.CacheResources, resID, value, nil, true, "")

	// Backup loop checks for the state of the stopBackup
	// channel after storing the resource. Channel can be
	// safely closed beforehand.
	close(rS.stopBackup)
	rS.runBackup(context.Background())

	want := &utils.Resource{
		Tenant: tnt,
		ID:     resID,
	}
	if got, err := rS.dm.GetResource(context.Background(), tnt, resID, true, false, ""); err != nil {
		t.Errorf("dm.GetResource(%q,%q): got unexpected err=%v", tnt, resID, err)
	} else if !reflect.DeepEqual(got, want) {
		t.Errorf("dm.GetResource(%q,%q) = %v, want %v", tnt, resID, got, want)
	}

	select {
	case <-rS.loopStopped:
	case <-time.After(time.Second):
		t.Error("timed out waiting for loop to stop")
	}
}

func TestResourcesReload(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ResourceSCfg().StoreInterval = 5 * time.Millisecond
	rS := &ResourceS{
		stopBackup:  make(chan struct{}),
		loopStopped: make(chan struct{}, 1),
		cfg:         cfg,
	}
	rS.loopStopped <- struct{}{}
	rS.Reload(context.Background())
	close(rS.stopBackup)
	select {
	case <-rS.loopStopped:
	case <-time.After(time.Second):
		t.Error("timed out waiting for loop to stop")
	}
}

func TestResourcesStartLoop(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ResourceSCfg().StoreInterval = -1
	rS := &ResourceS{
		loopStopped: make(chan struct{}),
		cfg:         cfg,
	}
	rS.StartLoop(context.Background())
	select {
	case <-rS.loopStopped:
	case <-time.After(time.Second):
		t.Error("timed out waiting for loop to stop")
	}
}

// func TestResourcesMatchingResourcesForEvent2(t *testing.T) {
// 	tmp := engine.Cache
// 	tmpC := config.CgrConfig()
// 	defer func() {
// 		engine.Cache = tmp
// 		config.SetCgrConfig(tmpC)
// 	}()

// 	engine.Cache.Clear(nil)
// 	cfg := config.NewDefaultCGRConfig()
// 	cfg.CacheCfg().ReplicationConns = []string{"test"}
// 	cfg.CacheCfg().Partitions[utils.CacheEventResources].Replicate = true
// 	cfg.RPCConns()["test"] = &config.RPCConn{Conns: []*config.RemoteHost{{}}}
// 	config.SetCgrConfig(cfg)
// 	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
// 	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
// dm := engine.NewDataManager(dbCM, cfg.CacheCfg(), nil)
// 	connMgr = engine.NewConnManager(cfg)
// 	engine.Cache = engine.NewCacheS(cfg, dm, nil,nil)

// 	fltrs := engine.NewFilterS(cfg, nil, dm)

// 	rsPrf := &utils.ResourceProfile{
// 		Tenant:            "cgrates.org",
// 		ID:                "RES1",
// 		FilterIDs:         []string{"*string:~*req.Account:1001"},
// 		ThresholdIDs:      []string{utils.MetaNone},
// 		AllocationMessage: "Approved",
// 		Weight:            10,
// 		Limit:             10,
// 		UsageTTL:          time.Minute,
// 		Stored:            true,
// 	}

// 	err := dm.SetResourceProfile(context.Background(), rsPrf, true)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	rS := NewResourceService(cfg, dm, fltrs, connMgr)
// 	ev := &utils.CGREvent{
// 		Tenant: "cgrates.org",
// 		ID:     "TestMatchingResourcesForEvent",
// 		Event: map[string]any{
// 			utils.AccountField: "1001",
// 		},
// 		APIOpts: map[string]any{},
// 	}

// 	engine.Cache.SetWithoutReplicate(utils.CacheEventResources, ev.ID, utils.StringSet{
// 		"RES1": struct{}{},
// 	}, nil, true, utils.NonTransactional)
// 	_, err = rS.matchingResourcesForEvent(context.Background(), "cgrates.org", ev, ev.ID, utils.DurationPointer(10*time.Second))
// }

func TestResourcesMatchingResourcesForEventCacheSetErr(t *testing.T) {
	tmp := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{"test"}
	cfg.CacheCfg().Partitions[utils.CacheEventResources].Replicate = true
	cfg.RPCConns()["test"] = &config.RPCConn{Conns: []*config.RemoteHost{{}}}
	config.SetCgrConfig(cfg)
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	connMgr := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	engine.Cache = engine.NewCacheS(cfg, dm, connMgr, nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)

	rS := NewResourceService(cfg, dm, fltrs, connMgr)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestMatchingResourcesForEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{},
	}

	rcv, _, err := rS.matchingResourcesForEvent(context.Background(), "cgrates.org", ev, ev.ID,
		utils.DurationPointer(10*time.Second))
	if err == nil || err.Error() != utils.ErrDisconnected.Error() {
		t.Fatalf("expected: <%+v>, \nreceived: <%+v>", utils.ErrDisconnected, err)
	}
	if rcv != nil {
		t.Errorf("expected nil, received: <%+v>", rcv)
	}
}

func TestResourcesMatchingResourcesForEventFinalCacheSetErr(t *testing.T) {
	tmp := engine.Cache
	tmpC := config.CgrConfig()
	defer func() {
		engine.Cache = tmp
		config.SetCgrConfig(tmpC)
	}()

	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().ReplicationConns = []string{"test"}
	cfg.CacheCfg().Partitions[utils.CacheEventResources].Replicate = true
	cfg.RPCConns()["test"] = &config.RPCConn{Conns: []*config.RemoteHost{{}}}
	config.SetCgrConfig(cfg)
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	connMgr := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	engine.Cache = engine.NewCacheS(cfg, dm, connMgr, nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)

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

	rS := NewResourceService(cfg, dm, fltrs, connMgr)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestMatchingResourcesForEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{},
	}
	exp := &matchedResource{
		Resource: &utils.Resource{
			Tenant: "cgrates.org",
			ID:     "RES1",
			Usages: make(map[string]*utils.ResourceUsage),
		},
		ttl:     utils.DurationPointer(10 * time.Second),
		profile: rsPrf,
	}

	rcv, unlock, err := rS.matchingResourcesForEvent(context.Background(), "cgrates.org", ev, ev.ID,
		utils.DurationPointer(10*time.Second))
	if err != nil {
		t.Fatalf("expected no error, received: %+v", err)
	}
	defer unlock()
	if !reflect.DeepEqual(rcv[0], exp) {
		t.Errorf("expected: <%+v>, received: <%+v>", exp, rcv[0])
	}
}

func TestResourceMatchingResourcesForEventWeightFromDynamicsErr(t *testing.T) {
	defer func() {
		engine.Cache = engine.NewCacheS(config.NewDefaultCGRConfig(), nil, nil, nil)
	}()

	cfg := config.NewDefaultCGRConfig()
	db, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	cfg.ResourceSCfg().StoreInterval = 1
	cfg.ResourceSCfg().StringIndexedFields = nil
	cfg.ResourceSCfg().PrefixIndexedFields = nil
	fltrs := engine.NewFilterS(cfg, nil, dm)
	rS := NewResourceService(cfg, dm,
		fltrs, nil)

	ids := make([]string, 0, 10)
	for i := range 10 {
		rPrf := &utils.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                fmt.Sprintf("RES%d", i),
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weights: utils.DynamicWeights{
				{
					FilterIDs: []string{"*stirng:~*req.Account:1001"},
					Weight:    float64(10 - i),
				}},
			Blocker:      i == 4,
			ThresholdIDs: []string{utils.MetaNone},
		}
		dm.SetResourceProfile(context.Background(), rPrf, true)
		ids = append(ids, rPrf.ID)
	}
	if err := engine.Cache.Set(context.Background(), utils.CacheEventResources, "TestResourceMatchingResourcesForEventLocksBlocker", ids, nil, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}

	expErr := "NOT_IMPLEMENTED:*stirng"
	_, _, err := rS.matchingResourcesForEvent(context.Background(), "cgrates.org", new(utils.CGREvent),
		"TestResourceMatchingResourcesForEventLocksBlocker", utils.DurationPointer(10*time.Second))
	if err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%+v>, received error <%+v>", expErr, err)
	}

}
func TestStoreMatchedResources(t *testing.T) {
	resPrf := &utils.ResourceProfile{
		Tenant: "cgrates.org",
		ID:     "RES1",
		Limit:  10,
		Stored: true,
	}

	newRSWithInterval := func(t *testing.T, interval time.Duration) (*ResourceS, *engine.DataManager) {
		t.Helper()
		cfg := config.NewDefaultCGRConfig()
		cfg.ResourceSCfg().StoreInterval = interval
		data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
		dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
		dm := engine.NewDataManager(dbCM, cfg, nil)
		rS := &ResourceS{dm: dm, cfg: cfg, storedResources: utils.NewStringSet(nil)}
		return rS, dm
	}

	t.Run("store interval zero is a no-op", func(t *testing.T) {
		rS, _ := newRSWithInterval(t, 0)
		res := Resources{{
			Resource: &utils.Resource{
				Tenant: "cgrates.org",
				ID:     "RES1",
				Usages: map[string]*utils.ResourceUsage{},
			},
			profile: resPrf,
		}}
		if err := rS.storeMatchedResources(context.Background(), res); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("positive interval queues to storedResources", func(t *testing.T) {
		rS, _ := newRSWithInterval(t, 10*time.Second)
		res := Resources{
			{
				Resource: &utils.Resource{
					Tenant: "cgrates.org",
					ID:     "RES1",
					Usages: map[string]*utils.ResourceUsage{},
				},
				profile: resPrf,
			},
			{
				Resource: &utils.Resource{
					Tenant: "cgrates.org",
					ID:     "RES2",
					Usages: map[string]*utils.ResourceUsage{},
				},
				profile: resPrf,
			},
		}
		if err := rS.storeMatchedResources(context.Background(), res); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !rS.storedResources.Has("cgrates.org:RES1") || !rS.storedResources.Has("cgrates.org:RES2") {
			t.Error("Expected both resources in storedResources set")
		}
		if rS.storedResources.Size() != 2 {
			t.Errorf("Expected 2 resources in set, got: %d", rS.storedResources.Size())
		}
	})

	t.Run("negative interval stores to DB", func(t *testing.T) {
		rS, dm := newRSWithInterval(t, -1)
		res := Resources{{
			Resource: &utils.Resource{
				Tenant: "cgrates.org",
				ID:     "RES1",
				Usages: map[string]*utils.ResourceUsage{},
			},
			profile: resPrf,
		}}
		if err := rS.storeMatchedResources(context.Background(), res); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		savedRes, err := dm.GetResource(context.Background(), "cgrates.org", "RES1", false, false, "")
		if err != nil {
			t.Errorf("Expected resource to be stored, got error: %v", err)
		}
		if savedRes == nil {
			t.Error("Expected resource to be stored in database")
		}
	})

	t.Run("Stored false skips store", func(t *testing.T) {
		rS, dm := newRSWithInterval(t, -1)
		res := Resources{{
			Resource: &utils.Resource{
				Tenant: "cgrates.org",
				ID:     "RES1",
				Usages: map[string]*utils.ResourceUsage{},
			},
			profile: &utils.ResourceProfile{Stored: false},
		}}
		if err := rS.storeMatchedResources(context.Background(), res); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err := dm.GetResource(context.Background(), "cgrates.org", "RES1", false, false, "")
		if err == nil {
			t.Error("Expected resource with Stored=false not to be stored")
		}
	})

	t.Run("stores resource with usages", func(t *testing.T) {
		rS, dm := newRSWithInterval(t, -1)
		res := Resources{{
			Resource: &utils.Resource{
				Tenant: "cgrates.org",
				ID:     "RES1",
				Usages: map[string]*utils.ResourceUsage{
					"USAGE1": {
						Tenant:     "cgrates.org",
						ID:         "USAGE1",
						ExpiryTime: time.Now().Add(time.Hour),
						Units:      5,
					},
					"USAGE2": {
						Tenant:     "cgrates.org",
						ID:         "USAGE2",
						ExpiryTime: time.Now().Add(2 * time.Hour),
						Units:      3,
					},
				},
				TTLIdx: []string{"USAGE1", "USAGE2"},
			},
			profile: resPrf,
		}}
		if err := rS.storeMatchedResources(context.Background(), res); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		savedRes, err := dm.GetResource(context.Background(), "cgrates.org", "RES1", false, false, "")
		if err != nil {
			t.Fatalf("Expected resource to be stored, got error: %v", err)
		}
		if len(savedRes.Usages) != 2 {
			t.Errorf("Expected 2 usages, got: %d", len(savedRes.Usages))
		}
		if _, ok := savedRes.Usages["USAGE1"]; !ok {
			t.Error("Expected USAGE1 to be stored")
		}
		if _, ok := savedRes.Usages["USAGE2"]; !ok {
			t.Error("Expected USAGE2 to be stored")
		}
		if len(savedRes.TTLIdx) != 2 {
			t.Errorf("Expected 2 TTL indexes, got: %d", len(savedRes.TTLIdx))
		}
	})
}
