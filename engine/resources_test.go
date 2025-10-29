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
package engine

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
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
			Event: map[string]any{
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
			Event: map[string]any{
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
			Event: map[string]any{
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
			Event: map[string]any{},
		},
		UsageID: "test1", // ResourceUsage Identifier
		Units:   20,
	}
	argsMissingUsageID := utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "id1",
			Event:  map[string]any{},
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

// UsageTTL 0 in ResourceProfile and give 10s duration
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

// UsageTTL 5s in ResourceProfile and give nil duration
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

// UsageTTL 5s in ResourceProfile and give 0 duration
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

// UsageTTL 5s in ResourceProfile and give 10s duration
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

func TestResourcesStoreResourceError(t *testing.T) {
	Cache.Clear(nil)
	cfg, _ := config.NewDefaultCGRConfig()
	cfg.ResourceSCfg().StoreInterval = -1
	cfg.RPCConns()["test"] = &config.RPCConn{
		Conns: []*config.RemoteHost{{}},
	}
	cfg.DataDbCfg().RplConns = []string{"test"}
	dft := config.CgrConfig()
	config.SetCgrConfig(cfg)
	defer config.SetCgrConfig(dft)

	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), NewConnManager(cfg, make(map[string]chan birpc.ClientConnector)))

	rS, _ := NewResourceService(dm, cfg, NewFilterS(cfg, nil, dm), nil)

	rsPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.META_NONE},
		AllocationMessage: "Approved",
		Weight:            10,
		Limit:             10,
		UsageTTL:          time.Minute,
		Stored:            true,
	}

	err := dm.SetResourceProfile(rsPrf, true)
	if err != nil {
		t.Fatal(err)
	}
	err = dm.SetResource(&Resource{
		Tenant: "cgrates.org",
		ID:     "RES1",
		Usages: make(map[string]*ResourceUsage),
	})
	if err != nil {
		t.Fatal(err)
	}

	args := utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "EventAuthorizeResource",
			Event: map[string]any{
				"Account": "1001",
			},
		},
		UsageID:  "RU_Test",
		UsageTTL: utils.DurationPointer(time.Minute),
		Units:    5,
	}
	expErr := "dial tcp: missing address"
	cfg.DataDbCfg().Items[utils.MetaResources].Replicate = true
	var reply string
	if err := rS.V1AllocateResource(args, &reply); err == nil || err.Error() != expErr {
		t.Error(err)
	}
	cfg.DataDbCfg().Items[utils.MetaResources].Replicate = false

	if err := rS.V1AllocateResource(args, &reply); err != nil {
		t.Error(err)
	} else if reply != "Approved" {
		t.Errorf("Unexpected reply returned: %q", reply)
	}

	cfg.DataDbCfg().Items[utils.MetaResources].Replicate = true
	if err := rS.V1ReleaseResource(args, &reply); err == nil || err.Error() != expErr {
		t.Error(err)
	}
}

func TestResourceMatchingResourcesForEventNotFoundInCache(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmRES := NewDataManager(db, config.CgrConfig().CacheCfg(), nil)
	cfg.ResourceSCfg().StoreInterval = 1
	cfg.ResourceSCfg().StringIndexedFields = nil
	cfg.ResourceSCfg().PrefixIndexedFields = nil
	rS, _ := NewResourceService(dmRES, cfg,
		&FilterS{dm: dmRES, cfg: cfg}, nil)

	Cache.Set(utils.CacheEventResources, "TestResourceMatchingResourcesForEventNotFoundInCache", nil, nil, true, utils.NonTransactional)
	mres, err := rS.matchingResourcesForEvent(&utils.CGREvent{Tenant: "cgrates.org"},
		"TestResourceMatchingResourcesForEventNotFoundInCache", utils.DurationPointer(10*time.Second))
	defer mres.unlock()
	if err != utils.ErrNotFound {
		t.Errorf("Error: %+v", err)
	}
}

func TestResourceMatchingResourcesForEventNotFoundInDB(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmRES := NewDataManager(db, config.CgrConfig().CacheCfg(), nil)
	cfg.ResourceSCfg().StoreInterval = 1
	cfg.ResourceSCfg().StringIndexedFields = nil
	cfg.ResourceSCfg().PrefixIndexedFields = nil
	rS, _ := NewResourceService(dmRES, cfg,
		&FilterS{dm: dmRES, cfg: cfg}, nil)

	Cache.Set(utils.CacheEventResources, "TestResourceMatchingResourcesForEventNotFoundInDB", utils.StringMap{"Res2": true}, nil, true, utils.NonTransactional)
	mres, err := rS.matchingResourcesForEvent(&utils.CGREvent{Tenant: "cgrates.org"},
		"TestResourceMatchingResourcesForEventNotFoundInDB", utils.DurationPointer(10*time.Second))
	defer mres.unlock()
	if err != utils.ErrNotFound {
		t.Errorf("Error: %+v", err)
	}
}

func TestResourceMatchingResourcesForEventLocks(t *testing.T) {
	Cache.Clear(nil)
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, config.CgrConfig().CacheCfg(), nil)
	cfg.ResourceSCfg().StoreInterval = 1
	cfg.ResourceSCfg().StringIndexedFields = nil
	cfg.ResourceSCfg().PrefixIndexedFields = nil
	rS, _ := NewResourceService(dm, cfg,
		&FilterS{dm: dm, cfg: cfg}, nil)

	prfs := make([]*ResourceProfile, 0)
	ids := utils.StringMap{}
	for i := 0; i < 10; i++ {
		rPrf := &ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                fmt.Sprintf("RES%d", i),
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{utils.META_NONE},
		}
		dm.SetResourceProfile(rPrf, true)
		if rPrf.ID != "RES1" {
			err = dm.SetResource(&Resource{
				Tenant: "cgrates.org",
				ID:     rPrf.ID,
				Usages: make(map[string]*ResourceUsage),
			})
			if err != nil {
				t.Fatal(err)
			}
		}
		prfs = append(prfs, rPrf)
		ids[rPrf.ID] = true
	}
	Cache.Set(utils.CacheEventResources, "TestResourceMatchingResourcesForEventLocks", ids, nil, true, utils.NonTransactional)
	mres, err := rS.matchingResourcesForEvent(&utils.CGREvent{Tenant: "cgrates.org"},
		"TestResourceMatchingResourcesForEventLocks", utils.DurationPointer(10*time.Second))
	if err != utils.ErrNotFound {
		t.Errorf("Error: %+v", err)
	}
	mres.unlock()
	for _, rPrf := range prfs {
		if rPrf.isLocked() {
			t.Fatalf("Expected profile to not be locked %q", rPrf.ID)
		}
		if rPrf.ID == "RES1" {
			continue
		}
		if r, err := dm.GetResource(rPrf.Tenant, rPrf.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("error %s for <%s>", err, rPrf.ID)
		} else if r.isLocked() {
			t.Fatalf("Expected resource to not be locked %q", rPrf.ID)
		}
	}

}

func TestResourceMatchingResourcesForEventLocks2(t *testing.T) {
	Cache.Clear(nil)
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, config.CgrConfig().CacheCfg(), nil)
	cfg.ResourceSCfg().StoreInterval = 1
	cfg.ResourceSCfg().StringIndexedFields = nil
	cfg.ResourceSCfg().PrefixIndexedFields = nil
	rS, _ := NewResourceService(dm, cfg,
		&FilterS{dm: dm, cfg: cfg}, nil)

	prfs := make([]*ResourceProfile, 0)
	ids := utils.StringMap{}
	for i := 0; i < 10; i++ {
		rPrf := &ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                fmt.Sprintf("RES%d", i),
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{utils.META_NONE},
		}
		err = dm.SetResource(&Resource{
			Tenant: "cgrates.org",
			ID:     rPrf.ID,
			Usages: make(map[string]*ResourceUsage),
		})
		if err != nil {
			t.Fatal(err)
		}
		dm.SetResourceProfile(rPrf, true)
		prfs = append(prfs, rPrf)
		ids[rPrf.ID] = true
	}
	rPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES20",
		FilterIDs:         []string{"FLTR_RES_2011"},
		UsageTTL:          10 * time.Second,
		Limit:             10.00,
		AllocationMessage: "AllocationMessage",
		Weight:            20.00,
		ThresholdIDs:      []string{utils.META_NONE},
	}
	err = db.SetResourceProfileDrv(rPrf)
	if err != nil {
		t.Fatal(err)
	}
	err = dm.SetResource(&Resource{
		Tenant: "cgrates.org",
		ID:     rPrf.ID,
		Usages: make(map[string]*ResourceUsage),
	})
	if err != nil {
		t.Fatal(err)
	}
	prfs = append(prfs, rPrf)
	ids[rPrf.ID] = true
	Cache.Set(utils.CacheEventResources, "TestResourceMatchingResourcesForEventLocks2", ids, nil, true, utils.NonTransactional)
	mres, err := rS.matchingResourcesForEvent(&utils.CGREvent{Tenant: "cgrates.org"},
		"TestResourceMatchingResourcesForEventLocks2", utils.DurationPointer(10*time.Second))
	expErr := utils.ErrPrefixNotFound(rPrf.FilterIDs[0])
	if err == nil || err.Error() != expErr.Error() {
		t.Errorf("Expected error: %s ,received: %+v", expErr, err)
	}
	defer mres.unlock()
	for _, rPrf := range prfs {
		if rPrf.isLocked() {
			t.Fatalf("Expected profile to not be locked %q", rPrf.ID)
		}
		if rPrf.ID == "RES20" {
			continue
		}
		if r, err := dm.GetResource(rPrf.Tenant, rPrf.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("error %s for <%s>", err, rPrf.ID)
		} else if r.isLocked() {
			t.Fatalf("Expected resource to not be locked %q", rPrf.ID)
		}
	}
}

func TestResourceMatchingResourcesForEventLocksBlocker(t *testing.T) {
	Cache.Clear(nil)
	cfg, _ := config.NewDefaultCGRConfig()
	dm := NewDataManager(NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items), config.CgrConfig().CacheCfg(), nil)
	cfg.ResourceSCfg().StoreInterval = 1
	cfg.ResourceSCfg().StringIndexedFields = nil
	cfg.ResourceSCfg().PrefixIndexedFields = nil
	rS, _ := NewResourceService(dm, cfg,
		&FilterS{dm: dm, cfg: cfg}, nil)

	prfs := make([]*ResourceProfile, 0)
	ids := utils.StringMap{}
	for i := 0; i < 10; i++ {
		rPrf := &ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                fmt.Sprintf("RESBL%d", i),
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            float64(10 - i),
			Blocker:           i == 4,
			ThresholdIDs:      []string{utils.META_NONE},
		}
		err = dm.SetResource(&Resource{
			Tenant: "cgrates.org",
			ID:     rPrf.ID,
			Usages: make(map[string]*ResourceUsage),
		})
		if err != nil {
			t.Fatal(err)
		}
		dm.SetResourceProfile(rPrf, true)
		prfs = append(prfs, rPrf)
		ids[rPrf.ID] = true
	}
	Cache.Set(utils.CacheEventResources, "TestResourceMatchingResourcesForEventLocksBlocker", ids, nil, true, utils.NonTransactional)
	mres, err := rS.matchingResourcesForEvent(&utils.CGREvent{Tenant: "cgrates.org"},
		"TestResourceMatchingResourcesForEventLocksBlocker", utils.DurationPointer(10*time.Second))
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	defer mres.unlock()
	if len(mres) != 5 {
		t.Fatal("Expected 6 resources")
	}
	for _, rPrf := range prfs[5:] {
		if rPrf.isLocked() {
			t.Errorf("Expected profile to not be locked %q", rPrf.ID)
		}
		if r, err := dm.GetResource(rPrf.Tenant, rPrf.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("error %s for <%s>", err, rPrf.ID)
		} else if r.isLocked() {
			t.Fatalf("Expected resource to not be locked %q", rPrf.ID)
		}
	}
	for _, rPrf := range prfs[:5] {
		if !rPrf.isLocked() {
			t.Errorf("Expected profile to be locked %q", rPrf.ID)
		}
		if r, err := dm.GetResource(rPrf.Tenant, rPrf.ID, true, false, utils.NonTransactional); err != nil {
			t.Errorf("error %s for <%s>", err, rPrf.ID)
		} else if !r.isLocked() {
			t.Fatalf("Expected resource to be locked %q", rPrf.ID)
		}
	}
}

func TestResourceMatchingResourcesForEventLocksActivationInterval(t *testing.T) {
	Cache.Clear(nil)
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, config.CgrConfig().CacheCfg(), nil)
	cfg.ResourceSCfg().StoreInterval = 1
	cfg.ResourceSCfg().StringIndexedFields = nil
	cfg.ResourceSCfg().PrefixIndexedFields = nil
	rS, _ := NewResourceService(dm, cfg,
		&FilterS{dm: dm, cfg: cfg}, nil)

	ids := utils.StringMap{}
	for i := 0; i < 10; i++ {
		rPrf := &ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                fmt.Sprintf("RES%d", i),
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{utils.META_NONE},
		}
		err = dm.SetResource(&Resource{
			Tenant: "cgrates.org",
			ID:     rPrf.ID,
			Usages: make(map[string]*ResourceUsage),
		})
		if err != nil {
			t.Fatal(err)
		}
		dm.SetResourceProfile(rPrf, true)
		ids[rPrf.ID] = true
	}
	rPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES21",
		UsageTTL:          10 * time.Second,
		Limit:             10.00,
		AllocationMessage: "AllocationMessage",
		Weight:            20.00,
		ThresholdIDs:      []string{utils.META_NONE},
		ActivationInterval: &utils.ActivationInterval{
			ExpiryTime: time.Now().Add(-5 * time.Second),
		},
	}
	err = dm.SetResource(&Resource{
		Tenant: "cgrates.org",
		ID:     rPrf.ID,
		Usages: make(map[string]*ResourceUsage),
	})
	if err != nil {
		t.Fatal(err)
	}
	dm.SetResourceProfile(rPrf, true)
	ids[rPrf.ID] = true
	Cache.Set(utils.CacheEventResources, "TestResourceMatchingResourcesForEventLocks2", ids, nil, true, utils.NonTransactional)
	mres, err := rS.matchingResourcesForEvent(&utils.CGREvent{Tenant: "cgrates.org", Time: utils.TimePointer(time.Now())},
		"TestResourceMatchingResourcesForEventLocks2", utils.DurationPointer(10*time.Second))
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	defer mres.unlock()
	if rPrf.isLocked() {
		t.Fatalf("Expected profile to not be locked %q", rPrf.ID)
	}
	if r, err := dm.GetResource(rPrf.Tenant, rPrf.ID, true, false, utils.NonTransactional); err != nil {
		t.Errorf("error %s for <%s>", err, rPrf.ID)
	} else if r.isLocked() {
		t.Fatalf("Expected resource to not be locked %q", rPrf.ID)
	}
}

func TestResourceForEvent(t *testing.T) {
	Cache.Clear(nil)
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	rS, _ := NewResourceService(dm, cfg,
		NewFilterS(cfg, nil, dm), nil)

	args := utils.ArgRSv1ResourceUsage{
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e51",
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]any{
				"Resource":    "Resource1",
				"Account":     "1002",
				"Subject":     "1001",
				"Destination": "1002"},
		},
		Units: 1,
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("res12345"),
		},
	}
	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "RS_FLT",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Destination",
				Values:  []string{"1002", "1003"},
			},
		},
	}
	dm.SetFilter(fltr)
	fltr2 := &Filter{
		Tenant: "cgrates.org",
		ID:     "RS_FLT2",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Resource",
				Values:  []string{"Resource1"},
			},
		},
	}
	dm.SetFilter(fltr2)
	var reply Resources

	rsP := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES20",
		FilterIDs:         []string{"RS_FLT"},
		UsageTTL:          10 * time.Second,
		Limit:             10.00,
		AllocationMessage: "AllocationMessage",
		Weight:            20.00,
		ThresholdIDs:      []string{utils.META_NONE},
	}

	dm.SetResourceProfile(rsP, true)
	dm.SetResource(&Resource{Tenant: "cgrates.org",
		ID: rsP.ID})

	rsP2 := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES21",
		UsageTTL:          10 * time.Second,
		Limit:             10.00,
		FilterIDs:         []string{"RS_FLT2"},
		AllocationMessage: "AllocationMessage",
		Weight:            20.00,
		ThresholdIDs:      []string{utils.META_NONE},
		ActivationInterval: &utils.ActivationInterval{
			ExpiryTime: time.Now().Add(-5 * time.Second),
		},
	}
	dm.SetResourceProfile(rsP2, true)
	dm.SetResource(&Resource{Tenant: "cgrates.org",
		ID: rsP2.ID})

	if err := rS.V1ResourcesForEvent(args, &reply); err != nil {
		t.Error(err)
	}
}

func TestResourcesRelease(t *testing.T) {
	Cache.Clear(nil)
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	rS, _ := NewResourceService(dm, cfg,
		NewFilterS(cfg, nil, dm), nil)

	args := utils.ArgRSv1ResourceUsage{
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e51",
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]any{
				"Resource":    "Resource1",
				"Account":     "1002",
				"Subject":     "1001",
				"Destination": "1002"},
		},
		Units: 1,
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("res12345"),
		},
	}
	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "RS_FLT",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Destination",
				Values:  []string{"1002", "1003"},
			},
		},
	}
	dm.SetFilter(fltr)
	rsP := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES20",
		FilterIDs:         []string{"RS_FLT"},
		UsageTTL:          10 * time.Second,
		Limit:             10.00,
		AllocationMessage: "AllocationMessage",
		Weight:            20.00,
		ThresholdIDs:      []string{utils.META_NONE},
	}
	dm.SetResourceProfile(rsP, true)
	dm.SetResource(&Resource{
		Tenant: "cgrates.org",
		ID:     rsP.ID,
		Usages: map[string]*ResourceUsage{
			"651a8db2-4f67-4cf8-b622-169e8a482e51": {
				Tenant: "cgrates.org",
				ID:     "651a8db2-4f67-4cf8-b622-169e8a482e21",
				Units:  2,
			},
		},
	})
	var reply string
	if err := rS.V1ReleaseResource(args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK,Received %v", reply)
	}
}

func TestResourceAuthorizeResources22(t *testing.T) {
	Cache.Clear(nil)
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, config.CgrConfig().CacheCfg(), nil)
	cfg.ResourceSCfg().StoreInterval = 1
	cfg.ResourceSCfg().StringIndexedFields = nil
	cfg.ResourceSCfg().PrefixIndexedFields = nil
	rS, _ := NewResourceService(dm, cfg,
		NewFilterS(cfg, nil, dm), nil)
	args := utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]any{
				utils.Account:     "1001",
				utils.Destination: "1002",
				"Resource":        "Resource1",
			},
		},

		UsageID: utils.UUIDSha1Prefix(),
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("res12345"),
		},
	}
	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_RES_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Resource",
				Values:  []string{"Resource1"},
			},
		}}
	dm.SetFilter(fltr)

	resource := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES20",
		FilterIDs:         []string{"FLTR_RES_1"},
		UsageTTL:          10 * time.Second,
		Limit:             10.00,
		AllocationMessage: "AllocationMessage",
		Weight:            20.00,
		ThresholdIDs:      []string{utils.META_NONE},
	}
	dm.SetResourceProfile(resource, true)
	dm.SetResource(&Resource{Tenant: "cgrates.org", ID: "RES20"})
	var reply string
	if err := rS.V1AuthorizeResources(args, &reply); err != nil {
		t.Error(err)
	}
	var replyS Resource
	if err := rS.V1GetResource(&utils.TenantID{Tenant: "cgrates.org", ID: "RES20"}, &replyS); err != nil {
		t.Error(err)
	}

}

func TestResourceService(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	cfg.ResourceSCfg().StoreInterval = 4 * time.Millisecond
	tmpCache := Cache
	defer func() {
		Cache = tmpCache
	}()
	Cache.Clear(nil)
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	rS, err := NewResourceService(dm, cfg, nil, nil)
	if err != nil {
		t.Error(err)
	}
	rS.storedResources = utils.StringMap{
		"R1": true,
	}
	Cache.Set(utils.CacheResources, "R1", &Resource{Tenant: "cgrates.org", ID: "R1", Usages: map[string]*ResourceUsage{
		"RU2": {
			Tenant:     "cgrates.org",
			ID:         "RU2",
			ExpiryTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
			Units:      2,
		},
	}}, []string{}, true, utils.NonTransactional)
	go func() {
		rS.loopStoped <- struct{}{}
	}()

	rS.Reload()
}

func TestRSProcessThreshold(t *testing.T) {
	Cache.Clear(nil)
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, config.CgrConfig().CacheCfg(), nil)
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(ctx *context.Context, serviceMethod string, _, _ any) error {
		if serviceMethod == utils.ThresholdSv1ProcessEvent {
			return nil
		}
		return utils.ErrNotImplemented
	})
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds): clientConn})
	rS, _ := NewResourceService(dm, cfg,
		NewFilterS(cfg, nil, dm), connMgr)
	cfg.ResourceSCfg().ThresholdSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}
	rs := Resources{
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
	if err := rS.processThresholds(rs, nil); err != nil {
		t.Error(err)
	}
}

func TestResourcesLockUnlock(t *testing.T) {
	rp := ResourceProfile{
		Tenant: "",
		ID:     "",
	}

	rp.lock("")

	if rp.lkID == "" {
		t.Error("didn't lock")
	}

	rp.lkID = ""

	rp.unlock()
}

func TestResourcesLockUnlock2(t *testing.T) {
	rp := Resource{
		Tenant: "",
		ID:     "",
	}

	rp.lock("")

	if rp.lkID == "" {
		t.Error("didn't lock")
	}

	rp.lkID = ""

	rp.unlock()
}

func TestResourcesRecordUsage(t *testing.T) {
	tm := 0 * time.Millisecond
	r := Resource{
		ttl: &tm,
	}

	ru := ResourceUsage{}

	err := r.recordUsage(&ru)
	if err != nil {
		t.Error(err)
	}
}

func TestResourcesrecordUsage(t *testing.T) {
	r := Resource{
		Usages: map[string]*ResourceUsage{"tes": {ID: "tes"}},
	}
	r2 := Resource{
		Usages: map[string]*ResourceUsage{"test": {ID: "test"}},
	}
	r3 := Resource{
		Usages: map[string]*ResourceUsage{"test": {ID: "test"}},
	}

	rs := Resources{&r, &r2, &r3}
	ru := ResourceUsage{ID: "test"}

	err := rs.recordUsage(&ru)
	if err.Error() != "duplicate resource usage with id: :test" {
		t.Error(err)
	}
}

func TestResourcesClearUsage(t *testing.T) {
	tm := 1 * time.Millisecond
	r := Resource{
		ttl: &tm,
	}
	rs := Resources{&r}

	err := rs.clearUsage("test")

	if err.Error() != "cannot find usage record with id: test" {
		t.Error(err)
	}
}

func TestResourcesAllocateReource(t *testing.T) {
	rs := Resources{}

	rcv, err := rs.allocateResource(nil, false)

	if err.Error() != utils.ErrResourceUnavailable.Error() {
		t.Error(err)
	}

	if rcv != "" {
		t.Errorf("expected %s\n, recived %s\n", rcv, "")
	}

	r := Resource{}
	rs2 := Resources{&r}
	ru := ResourceUsage{
		ID: "test",
	}

	rcv, err = rs2.allocateResource(&ru, true)

	if err.Error() != "empty configuration for resourceID: :" {
		t.Error(err)
	}

	if rcv != "" {
		t.Errorf("expected %s\n, recived %s\n", rcv, "")
	}
}
