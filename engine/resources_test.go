/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/
package engine

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	r1, r2               *Resource
	ru1, ru2, ru3        *ResourceUsage
	rs                   Resources
	cloneExpTimeResource time.Time
	expTimeResource      = time.Now().Add(time.Duration(20 * time.Minute))
	timeDurationExample  = time.Duration(10) * time.Second
	resService           *ResourceService
	dmRES                *DataManager
	resprf               = []*ResourceProfile{
		{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ResourceProfile1",
			FilterIDs: []string{"FLTR_RES_1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			UsageTTL:          time.Duration(10) * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
		{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ResourceProfile2", // identifier of this resource
			FilterIDs: []string{"FLTR_RES_2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			UsageTTL:          time.Duration(10) * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
		{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ResourceProfile3",
			FilterIDs: []string{"FLTR_RES_3"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			UsageTTL:          time.Duration(10) * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
	}
	resourceTest = Resources{
		{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     "ResourceProfile1",
			Usages: map[string]*ResourceUsage{},
			TTLIdx: []string{},
			rPrf:   resprf[0],
		},
		{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     "ResourceProfile2",
			Usages: map[string]*ResourceUsage{},
			TTLIdx: []string{},
			rPrf:   resprf[1],
		},
		{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     "ResourceProfile3",
			Usages: map[string]*ResourceUsage{},
			TTLIdx: []string{},
			rPrf:   resprf[2],
		},
	}
	resEvs = []*utils.CGREvent{
		{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     "event1",
			Event: map[string]interface{}{
				"Resources":      "ResourceProfile1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				utils.Weight:     "20.0",
				utils.Usage:      time.Duration(135 * time.Second),
				utils.COST:       123.0,
			},
		},
		{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     "event2",
			Event: map[string]interface{}{
				"Resources":      "ResourceProfile2",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				utils.Weight:     "15.0",
				utils.Usage:      time.Duration(45 * time.Second),
			},
		},
		{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     "event3",
			Event: map[string]interface{}{
				"Resources": "ResourceProfilePrefix",
				utils.Usage: time.Duration(30 * time.Second),
			},
		},
	}
)

func TestResourceRecordUsage1(t *testing.T) {
	ru1 = &ResourceUsage{
		Tenant:     "cgrates.org",
		ID:         "RU1",
		ExpiryTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
		Units:      1,
	}

	ru2 = &ResourceUsage{
		Tenant:     "cgrates.org",
		ID:         "RU2",
		ExpiryTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
		Units:      2,
	}

	r1 = &Resource{
		Tenant: "cgrates.org",
		ID:     "RL1",
		rPrf: &ResourceProfile{
			Tenant:    "cgrates.org",
			ID:        "RL1",
			FilterIDs: []string{"FLTR_RES_RL1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC).Add(time.Duration(1 * time.Millisecond)),
			},
			Weight:       100,
			Limit:        2,
			ThresholdIDs: []string{"TEST_ACTIONS"},

			UsageTTL:          time.Duration(1 * time.Millisecond),
			AllocationMessage: "ALLOC",
		},
		Usages: map[string]*ResourceUsage{
			ru1.ID: ru1,
		},
		TTLIdx: []string{ru1.ID},
		tUsage: utils.Float64Pointer(2),
	}

	if err := r1.recordUsage(ru2); err != nil {
		t.Error(err.Error())
	} else {
		if err := r1.recordUsage(ru1); err == nil {
			t.Error("duplicate ResourceUsage id should not be allowed")
		}
		if _, found := r1.Usages[ru2.ID]; !found {
			t.Error("ResourceUsage was not recorded")
		}
		if *r1.tUsage != 4 {
			t.Errorf("expecting: %+v, received: %+v", 4, r1.tUsage)
		}
	}
}

func TestResourceRemoveExpiredUnits(t *testing.T) {
	r1.Usages = map[string]*ResourceUsage{
		ru1.ID: ru1,
	}
	*r1.tUsage = 2

	r1.removeExpiredUnits()

	if len(r1.Usages) != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, len(r1.Usages))
	}
	if len(r1.TTLIdx) != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, len(r1.TTLIdx))
	}
	if r1.tUsage != nil && *r1.tUsage != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, r1.tUsage)
	}
}

func TestResourceUsedUnits(t *testing.T) {
	r1.Usages = map[string]*ResourceUsage{
		ru1.ID: ru1,
	}
	r1.tUsage = nil
	if usedUnits := r1.TotalUsage(); usedUnits != 1 {
		t.Errorf("Expecting: %+v, received: %+v", 1, usedUnits)
	}
}

func TestResourceSort(t *testing.T) {
	r2 = &Resource{
		Tenant: "cgrates.org",
		ID:     "RL2",
		rPrf: &ResourceProfile{
			ID:        "RL2",
			FilterIDs: []string{"FLTR_RES_RL2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
			},

			Weight:       50,
			Limit:        2,
			ThresholdIDs: []string{"TEST_ACTIONS"},
			UsageTTL:     time.Duration(1 * time.Millisecond),
		},
		// AllocationMessage: "ALLOC2",
		Usages: map[string]*ResourceUsage{
			ru2.ID: ru2,
		},
		tUsage: utils.Float64Pointer(2),
	}

	rs = Resources{r2, r1}
	rs.Sort()

	if rs[0].rPrf.ID != "RL1" {
		t.Error("Sort failed")
	}
}

func TestResourceClearUsage(t *testing.T) {
	r1.Usages = map[string]*ResourceUsage{
		ru1.ID: ru1,
	}
	r1.tUsage = nil
	r1.clearUsage(ru1.ID)
	if len(r1.Usages) != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, len(r1.Usages))
	}
	if r1.TotalUsage() != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, r1.tUsage)
	}
	if err := r2.clearUsage(ru2.ID); err != nil {
		t.Error(err)
	} else if len(r2.Usages) != 0 {
		t.Errorf("Unexpected usages %+v", r2.Usages)
	} else if *r2.tUsage != 0 {
		t.Errorf("Unexpected tUsage %+v", r2.tUsage)
	}
}

func TestResourceRecordUsages(t *testing.T) {
	r1.Usages = map[string]*ResourceUsage{
		ru1.ID: ru1,
	}
	r1.tUsage = nil
	if err := rs.recordUsage(ru1); err == nil {
		t.Error("should get duplicated error")
	}
}

func TestResourceAllocateResource(t *testing.T) {
	rs.clearUsage(ru1.ID)
	rs.clearUsage(ru2.ID)
	ru1.ExpiryTime = time.Now().Add(time.Duration(1 * time.Second))
	ru2.ExpiryTime = time.Now().Add(time.Duration(1 * time.Second))
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
	rs[0].rPrf.Limit = 1
	rs[1].rPrf.Limit = 4
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
	if *dbType == utils.MetaInternal {
		t.SkipNow()
	}
	r := &Resource{
		Tenant: "cgrates.org",
		ID:     "RL",
		rPrf: &ResourceProfile{
			Tenant:    "cgrates.org",
			ID:        "RL",
			FilterIDs: []string{"FLTR_RES_RL"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
			},
			AllocationMessage: "ALLOC_RL",
			Weight:            50,
			Limit:             2,
			ThresholdIDs:      []string{"TEST_ACTIONS"},
			UsageTTL:          time.Duration(1 * time.Millisecond),
		},
		Usages: map[string]*ResourceUsage{
			"RU2": {
				Tenant:     "cgrates.org",
				ID:         "RU2",
				ExpiryTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
				Units:      2,
			},
		},
		tUsage: utils.Float64Pointer(2),
		dirty:  utils.BoolPointer(true),
	}
	Cache.Set(utils.CacheResources, r.TenantID(), r, nil, true, "")
	if x, ok := Cache.Get(utils.CacheResources, r.TenantID()); !ok {
		t.Error("not in cache")
	} else if x == nil {
		t.Error("nil resource")
	} else if !reflect.DeepEqual(r, x.(*Resource)) {
		t.Errorf("Expecting: %+v, received: %+v", r, x)
	}
}

func TestResourcePopulateResourceService(t *testing.T) {
	defaultCfg, _ := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true, defaultCfg.DataDbCfg().Items)
	dmRES = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	defaultCfg.ResourceSCfg().StoreInterval = 1
	defaultCfg.ResourceSCfg().StringIndexedFields = nil
	defaultCfg.ResourceSCfg().PrefixIndexedFields = nil
	resService, err = NewResourceService(dmRES, defaultCfg,
		&FilterS{dm: dmRES, cfg: defaultCfg}, nil)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
}

func TestResourceV1AuthorizeResourceMissingStruct(t *testing.T) {
	var reply *string
	argsMissingTenant := utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			ID:    "id1",
			Event: map[string]interface{}{},
		},
		UsageID: "test1", // ResourceUsage Identifier
		Units:   20,
	}
	argsMissingUsageID := utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "id1",
			Event:  map[string]interface{}{},
		},
		Units: 20,
	}
	if err := resService.V1AuthorizeResources(argsMissingTenant, reply); err.Error() != "MANDATORY_IE_MISSING: [Tenant Event]" {
		t.Error(err.Error())
	}
	if err := resService.V1AuthorizeResources(argsMissingUsageID, reply); err.Error() != "MANDATORY_IE_MISSING: [Event]" {
		t.Error(err.Error())
	}
}

func TestResourceAddFilters(t *testing.T) {
	fltrRes1 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_RES_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Resources",
				Values:  []string{"ResourceProfile1"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req.UsageInterval",
				Values:  []string{(1 * time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Values:  []string{(1 * time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"9.0"},
			},
		},
	}
	dmRES.SetFilter(fltrRes1)
	fltrRes2 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_RES_2",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Resources",
				Values:  []string{"ResourceProfile2"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req.PddInterval",
				Values:  []string{(1 * time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Values:  []string{(1 * time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"15.0"},
			},
		},
	}
	dmRES.SetFilter(fltrRes2)
	fltrRes3 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_RES_3",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Resources",
				Values:  []string{"ResourceProfilePrefix"},
			},
		},
	}
	dmRES.SetFilter(fltrRes3)
}

func TestResourceAddResourceProfile(t *testing.T) {
	for _, resProfile := range resprf {
		dmRES.SetResourceProfile(resProfile, true)
	}
	for _, res := range resourceTest {
		dmRES.SetResource(res)
	}
	//Test each resourceProfile from cache
	for _, resPrf := range resprf {
		if tempRes, err := dmRES.GetResourceProfile(resPrf.Tenant,
			resPrf.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(resPrf, tempRes) {
			t.Errorf("Expecting: %+v, received: %+v", resPrf, tempRes)
		}
	}
}

func TestResourceMatchingResourcesForEvent(t *testing.T) {
	mres, err := resService.matchingResourcesForEvent(resEvs[0],
		"TestResourceMatchingResourcesForEvent1", &timeDurationExample)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	mres.unlock()
	if !reflect.DeepEqual(resourceTest[0].Tenant, mres[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[0].Tenant, mres[0].Tenant)
	} else if !reflect.DeepEqual(resourceTest[0].ID, mres[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[0].ID, mres[0].ID)
	} else if !reflect.DeepEqual(resourceTest[0].rPrf, mres[0].rPrf) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[0].rPrf, mres[0].rPrf)
	}

	mres, err = resService.matchingResourcesForEvent(resEvs[1],
		"TestResourceMatchingResourcesForEvent2", &timeDurationExample)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	mres.unlock()
	if !reflect.DeepEqual(resourceTest[1].Tenant, mres[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[1].Tenant, mres[0].Tenant)
	} else if !reflect.DeepEqual(resourceTest[1].ID, mres[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[1].ID, mres[0].ID)
	} else if !reflect.DeepEqual(resourceTest[1].rPrf, mres[0].rPrf) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[1].rPrf, mres[0].rPrf)
	}

	mres, err = resService.matchingResourcesForEvent(resEvs[2],
		"TestResourceMatchingResourcesForEvent3", &timeDurationExample)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	mres.unlock()
	if !reflect.DeepEqual(resourceTest[2].Tenant, mres[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[2].Tenant, mres[0].Tenant)
	} else if !reflect.DeepEqual(resourceTest[2].ID, mres[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[2].ID, mres[0].ID)
	} else if !reflect.DeepEqual(resourceTest[2].rPrf, mres[0].rPrf) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[2].rPrf, mres[0].rPrf)
	}
}

//UsageTTL 0 in ResourceProfile and give 10s duration
func TestResourceUsageTTLCase1(t *testing.T) {
	Cache.Clear(nil)
	resprf[0].UsageTTL = time.Duration(0)
	resourceTest[0].rPrf = resprf[0]
	resourceTest[0].ttl = &timeDurationExample
	if err := dmRES.SetResourceProfile(resprf[0], true); err != nil {
		t.Error(err)
	}
	if err := dmRES.SetResource(resourceTest[0]); err != nil {
		t.Error(err)
	}
	mres, err := resService.matchingResourcesForEvent(resEvs[0],
		"TestResourceUsageTTLCase1", &timeDurationExample)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	mres.unlock()
	if !reflect.DeepEqual(resourceTest[0].Tenant, mres[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[0].Tenant, mres[0].Tenant)
	} else if !reflect.DeepEqual(resourceTest[0].ID, mres[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[0].ID, mres[0].ID)
	} else if !reflect.DeepEqual(resourceTest[0].rPrf, mres[0].rPrf) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[0].rPrf, mres[0].rPrf)
	} else if !reflect.DeepEqual(resourceTest[0].ttl, mres[0].ttl) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[0].ttl, mres[0].ttl)
	}
}

//UsageTTL 5s in ResourceProfile and give nil duration
func TestResourceUsageTTLCase2(t *testing.T) {
	Cache.Clear(nil)
	resprf[0].UsageTTL = time.Duration(0)
	resourceTest[0].rPrf = resprf[0]
	resourceTest[0].ttl = &resprf[0].UsageTTL
	if err := dmRES.SetResourceProfile(resprf[0], true); err != nil {
		t.Error(err)
	}
	if err := dmRES.SetResource(resourceTest[0]); err != nil {
		t.Error(err)
	}
	mres, err := resService.matchingResourcesForEvent(resEvs[0],
		"TestResourceUsageTTLCase2", nil)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	mres.unlock()
	if !reflect.DeepEqual(resourceTest[0].Tenant, mres[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[0].Tenant, mres[0].Tenant)
	} else if !reflect.DeepEqual(resourceTest[0].ID, mres[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[0].ID, mres[0].ID)
	} else if !reflect.DeepEqual(resourceTest[0].rPrf, mres[0].rPrf) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[0].rPrf, mres[0].rPrf)
	} else if !reflect.DeepEqual(resourceTest[0].ttl, mres[0].ttl) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[0].ttl, mres[0].ttl)
	}
}

//UsageTTL 5s in ResourceProfile and give 0 duration
func TestResourceUsageTTLCase3(t *testing.T) {
	Cache.Clear(nil)
	resprf[0].UsageTTL = time.Duration(0)
	resourceTest[0].rPrf = resprf[0]
	resourceTest[0].ttl = nil
	if err := dmRES.SetResourceProfile(resprf[0], true); err != nil {
		t.Error(err)
	}
	if err := dmRES.SetResource(resourceTest[0]); err != nil {
		t.Error(err)
	}
	mres, err := resService.matchingResourcesForEvent(resEvs[0],
		"TestResourceUsageTTLCase3", utils.DurationPointer(time.Duration(0)))
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	mres.unlock()
	if !reflect.DeepEqual(resourceTest[0].Tenant, mres[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[0].Tenant, mres[0].Tenant)
	} else if !reflect.DeepEqual(resourceTest[0].ID, mres[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[0].ID, mres[0].ID)
	} else if !reflect.DeepEqual(resourceTest[0].rPrf, mres[0].rPrf) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[0].rPrf, mres[0].rPrf)
	} else if !reflect.DeepEqual(resourceTest[0].ttl, mres[0].ttl) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[0].ttl, mres[0].ttl)
	}
}

//UsageTTL 5s in ResourceProfile and give 10s duration
func TestResourceUsageTTLCase4(t *testing.T) {
	Cache.Clear(nil)
	resprf[0].UsageTTL = time.Duration(5)
	resourceTest[0].rPrf = resprf[0]
	resourceTest[0].ttl = &timeDurationExample
	if err := dmRES.SetResourceProfile(resprf[0], true); err != nil {
		t.Error(err)
	}
	if err := dmRES.SetResource(resourceTest[0]); err != nil {
		t.Error(err)
	}
	mres, err := resService.matchingResourcesForEvent(resEvs[0],
		"TestResourceUsageTTLCase4", &timeDurationExample)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	mres.unlock()
	if !reflect.DeepEqual(resourceTest[0].Tenant, mres[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[0].Tenant, mres[0].Tenant)
	} else if !reflect.DeepEqual(resourceTest[0].ID, mres[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[0].ID, mres[0].ID)
	} else if !reflect.DeepEqual(resourceTest[0].rPrf, mres[0].rPrf) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[0].rPrf, mres[0].rPrf)
	} else if !reflect.DeepEqual(resourceTest[0].ttl, mres[0].ttl) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[0].ttl, mres[0].ttl)
	}
}

func TestResourceMatchWithIndexFalse(t *testing.T) {
	Cache.Clear(nil)
	resService.cgrcfg.ResourceSCfg().IndexedSelects = false
	mres, err := resService.matchingResourcesForEvent(resEvs[0],
		"TestResourceMatchWithIndexFalse1", &timeDurationExample)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	mres.unlock()
	if !reflect.DeepEqual(resourceTest[0].Tenant, mres[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[0].Tenant, mres[0].Tenant)
	} else if !reflect.DeepEqual(resourceTest[0].ID, mres[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[0].ID, mres[0].ID)
	} else if !reflect.DeepEqual(resourceTest[0].rPrf, mres[0].rPrf) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[0].rPrf, mres[0].rPrf)
	}

	mres, err = resService.matchingResourcesForEvent(resEvs[1],
		"TestResourceMatchWithIndexFalse2", &timeDurationExample)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	mres.unlock()
	if !reflect.DeepEqual(resourceTest[1].Tenant, mres[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[1].Tenant, mres[0].Tenant)
	} else if !reflect.DeepEqual(resourceTest[1].ID, mres[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[1].ID, mres[0].ID)
	} else if !reflect.DeepEqual(resourceTest[1].rPrf, mres[0].rPrf) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[1].rPrf, mres[0].rPrf)
	}

	mres, err = resService.matchingResourcesForEvent(resEvs[2],
		"TestResourceMatchWithIndexFalse3", &timeDurationExample)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	mres.unlock()
	if !reflect.DeepEqual(resourceTest[2].Tenant, mres[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[2].Tenant, mres[0].Tenant)
	} else if !reflect.DeepEqual(resourceTest[2].ID, mres[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[2].ID, mres[0].ID)
	} else if !reflect.DeepEqual(resourceTest[2].rPrf, mres[0].rPrf) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[2].rPrf, mres[0].rPrf)
	}
}

func TestResourceResIDsMp(t *testing.T) {
	expected := utils.StringMap{
		"ResourceProfile1": true,
		"ResourceProfile2": true,
		"ResourceProfile3": true,
	}
	if rcv := resourceTest.resIDsMp(); !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
}

func TestResourceTenatIDs(t *testing.T) {
	expected := []string{
		"cgrates.org:ResourceProfile1",
		"cgrates.org:ResourceProfile2",
		"cgrates.org:ResourceProfile3",
	}
	if rcv := resourceTest.tenatIDs(); !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
}

func TestResourceIDs(t *testing.T) {
	expected := []string{
		"ResourceProfile1",
		"ResourceProfile2",
		"ResourceProfile3",
	}
	if rcv := resourceTest.IDs(); !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
}

func TestResourceCaching(t *testing.T) {
	if *dbType == utils.MetaInternal {
		t.SkipNow()
	}
	//clear the cache
	Cache.Clear(nil)
	// start fresh with new dataManager
	defaultCfg, _ := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true, defaultCfg.DataDbCfg().Items)
	dmRES = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)

	defaultCfg.ResourceSCfg().StoreInterval = 1
	defaultCfg.ResourceSCfg().StringIndexedFields = nil
	defaultCfg.ResourceSCfg().PrefixIndexedFields = nil
	resService, err = NewResourceService(dmRES, defaultCfg,
		&FilterS{dm: dmRES, cfg: defaultCfg}, nil)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}

	resProf := &ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "ResourceProfileCached",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		UsageTTL:          time.Duration(-1),
		Limit:             10.00,
		AllocationMessage: "AllocationMessage",
		Weight:            20.00,
		ThresholdIDs:      []string{utils.META_NONE},
	}

	Cache.Set(utils.CacheResourceProfiles, "cgrates.org:ResourceProfileCached",
		resProf, nil, cacheCommit(utils.EmptyString), utils.EmptyString)

	res := &Resource{Tenant: resProf.Tenant,
		ID:     resProf.ID,
		Usages: make(map[string]*ResourceUsage)}

	Cache.Set(utils.CacheResources, "cgrates.org:ResourceProfileCached",
		res, nil, cacheCommit(utils.EmptyString), utils.EmptyString)

	resources := Resources{res}
	Cache.Set(utils.CacheEventResources, "TestResourceCaching", resources.resIDsMp(), nil, true, "")

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Event: map[string]interface{}{
			"Account":     "1001",
			"Destination": "3002"},
	}

	mres, err := resService.matchingResourcesForEvent(ev,
		"TestResourceCaching", nil)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	mres.unlock()
	if !reflect.DeepEqual(resources[0].Tenant, mres[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", resources[0].Tenant, mres[0].Tenant)
	} else if !reflect.DeepEqual(resources[0].ID, mres[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", resources[0].ID, mres[0].ID)
	} else if !reflect.DeepEqual(resources[0].rPrf, mres[0].rPrf) {
		t.Errorf("Expecting: %+v, received: %+v", resources[0].rPrf, mres[0].rPrf)
	} else if !reflect.DeepEqual(resources[0].ttl, mres[0].ttl) {
		t.Errorf("Expecting: %+v, received: %+v", resources[0].ttl, mres[0].ttl)
	}
}
