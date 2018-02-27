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
	resserv              ResourceService
	dmRES                *DataManager
	resprf               = []*ResourceProfile{
		&ResourceProfile{
			Tenant:    "cgrates.org",
			ID:        "resourcesprofile1", // identifier of this resource
			FilterIDs: []string{"filter9"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			UsageTTL:          time.Duration(10) * time.Second, // auto-expire the usage after this duration
			Limit:             10.00,                           // limit value
			AllocationMessage: "AllocationMessage",             // message returned by the winning resource on allocation
			Blocker:           false,                           // blocker flag to stop processing on filters matched
			Stored:            false,
			Weight:            20.00,        // Weight to sort the resources
			ThresholdIDs:      []string{""}, // Thresholds to check after changing Limit
		},
		&ResourceProfile{
			Tenant:    "cgrates.org",
			ID:        "resourcesprofile2", // identifier of this resource
			FilterIDs: []string{"filter10"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			UsageTTL:          time.Duration(10) * time.Second, // auto-expire the usage after this duration
			Limit:             10.00,                           // limit value
			AllocationMessage: "AllocationMessage",             // message returned by the winning resource on allocation
			Blocker:           false,                           // blocker flag to stop processing on filters matched
			Stored:            false,
			Weight:            20.00,        // Weight to sort the resources
			ThresholdIDs:      []string{""}, // Thresholds to check after changing Limit
		},
		&ResourceProfile{
			Tenant:    "cgrates.org",
			ID:        "resourcesprofile3", // identifier of this resource
			FilterIDs: []string{"preffilter5"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			UsageTTL:          time.Duration(10) * time.Second, // auto-expire the usage after this duration
			Limit:             10.00,                           // limit value
			AllocationMessage: "AllocationMessage",             // message returned by the winning resource on allocation
			Blocker:           false,                           // blocker flag to stop processing on filters matched
			Stored:            false,
			Weight:            20.00,        // Weight to sort the resources
			ThresholdIDs:      []string{""}, // Thresholds to check after changing Limit
		},
		&ResourceProfile{
			Tenant:    "cgrates.org",
			ID:        "resourcesprofile4", // identifier of this resource
			FilterIDs: []string{"defaultf5"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			UsageTTL:          time.Duration(10) * time.Second, // auto-expire the usage after this duration
			Limit:             10.00,                           // limit value
			AllocationMessage: "AllocationMessage",             // message returned by the winning resource on allocation
			Blocker:           false,                           // blocker flag to stop processing on filters matched
			Stored:            false,
			Weight:            20.00,        // Weight to sort the resources
			ThresholdIDs:      []string{""}, // Thresholds to check after changing Limit
		},
	}
	resourceTest = []*Resource{
		&Resource{
			Tenant: "cgrates.org",
			ID:     "resourcesprofile1",
			Usages: map[string]*ResourceUsage{},
			TTLIdx: []string{}, // holds ordered list of ResourceIDs based on their TTL, empty if feature is disabled
			rPrf:   resprf[0],  // for ordering purposes
		},
		&Resource{
			Tenant: "cgrates.org",
			ID:     "resourcesprofile2",
			Usages: map[string]*ResourceUsage{},
			TTLIdx: []string{}, // holds ordered list of ResourceIDs based on their TTL, empty if feature is disabled
			rPrf:   resprf[1],  // for ordering purposes
		},
		&Resource{
			Tenant: "cgrates.org",
			ID:     "resourcesprofile3",
			Usages: map[string]*ResourceUsage{},
			TTLIdx: []string{}, // holds ordered list of ResourceIDs based on their TTL, empty if feature is disabled
			rPrf:   resprf[2],  // for ordering purposes
		},
		&Resource{
			Tenant: "cgrates.org",
			ID:     "resourcesprofile4",
			Usages: map[string]*ResourceUsage{},
			TTLIdx: []string{}, // holds ordered list of ResourceIDs based on their TTL, empty if feature is disabled
			rPrf:   resprf[3],  // for ordering purposes
		},
	}
	resEvs = []*utils.CGREvent{
		&utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				"Resources":      "ResourcesProfile1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				"Weight":         "20.0",
				utils.Usage:      time.Duration(135 * time.Second),
				utils.COST:       123.0,
			}},
		&utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event2",
			Event: map[string]interface{}{
				"Resources":      "ResourcesProfile2",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				"Weight":         "21.0",
				utils.Usage:      time.Duration(45 * time.Second),
			}},
		&utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event3",
			Event: map[string]interface{}{
				"Resources": "ResourcesProfilePrefix",
				utils.Usage: time.Duration(30 * time.Second),
			}},
		&utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event3",
			Event: map[string]interface{}{
				"Weight":    "200.0",
				utils.Usage: time.Duration(65 * time.Second),
			}},
	}
)

func TestRSRecordUsage1(t *testing.T) {
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

func TestRSRemoveExpiredUnits(t *testing.T) {
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

func TestRSUsedUnits(t *testing.T) {
	r1.Usages = map[string]*ResourceUsage{
		ru1.ID: ru1,
	}
	r1.tUsage = nil
	if usedUnits := r1.totalUsage(); usedUnits != 1 {
		t.Errorf("Expecting: %+v, received: %+v", 1, usedUnits)
	}
}

func TestRSRsort(t *testing.T) {
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

func TestRSClearUsage(t *testing.T) {
	r1.Usages = map[string]*ResourceUsage{
		ru1.ID: ru1,
	}
	r1.tUsage = nil
	r1.clearUsage(ru1.ID)
	if len(r1.Usages) != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, len(r1.Usages))
	}
	if r1.totalUsage() != 0 {
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

func TestRSRecordUsages(t *testing.T) {
	r1.Usages = map[string]*ResourceUsage{
		ru1.ID: ru1,
	}
	r1.tUsage = nil
	if err := rs.recordUsage(ru1); err == nil {
		t.Error("should get duplicated error")
	}
}

func TestRSAllocateResource(t *testing.T) {
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
	if alcMessage, err := rs.allocateResource(ru1, true); err != nil {
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
			"RU2": &ResourceUsage{
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

func TestV1AuthorizeResourceMissingStruct(t *testing.T) {
	data, _ := NewMapStorage()
	dmresmiss := NewDataManager(data)

	rserv := &ResourceService{
		dm:                  dmresmiss,
		filterS:             &FilterS{dm: dmresmiss},
		stringIndexedFields: &[]string{}, // speed up query on indexes
	}
	var reply *string
	argsMissingTenant := utils.ArgRSv1ResourceUsage{
		CGREvent: utils.CGREvent{
			ID:    "id1",
			Event: map[string]interface{}{},
		},
		UsageID: "test1", // ResourceUsage Identifier
		Units:   20,
	}
	argsMissingUsageID := utils.ArgRSv1ResourceUsage{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "id1",
			Event:  map[string]interface{}{},
		},
		Units: 20,
	}
	if err := rserv.V1AuthorizeResources(argsMissingTenant, reply); err.Error() != "MANDATORY_IE_MISSING: [Tenant]" {
		t.Error(err.Error())
	}
	if err := rserv.V1AuthorizeResources(argsMissingUsageID, reply); err.Error() != "MANDATORY_IE_MISSING: [UsageID]" {
		t.Error(err.Error())
	}
}

func TestRSPopulateResourceService(t *testing.T) {
	data, _ := NewMapStorage()
	dmRES = NewDataManager(data)
	var filters1 []*FilterRule
	var filters2 []*FilterRule
	var preffilter []*FilterRule
	var defaultf []*FilterRule
	second := 1 * time.Second
	resserv = ResourceService{
		dm:      dmRES,
		filterS: &FilterS{dm: dmRES},
	}
	ref := NewFilterIndexer(dmRES, utils.ResourceProfilesPrefix, "cgrates.org")
	//filter1
	x, err := NewFilterRule(MetaString, "Resources", []string{"ResourcesProfile1"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters1 = append(filters1, x)
	x, err = NewFilterRule(MetaGreaterOrEqual, "UsageInterval", []string{second.String()})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters1 = append(filters1, x)
	x, err = NewFilterRule(MetaGreaterOrEqual, utils.Usage, []string{second.String()})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters1 = append(filters1, x)
	x, err = NewFilterRule(MetaGreaterOrEqual, "Weight", []string{"9.0"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters1 = append(filters1, x)
	filter9 := &Filter{Tenant: config.CgrConfig().DefaultTenant, ID: "filter9", Rules: filters1}
	dmRES.SetFilter(filter9)
	ref.IndexTPFilter(FilterToTPFilter(filter9), "resourcesprofile1")
	//filter2
	x, err = NewFilterRule(MetaString, "Resources", []string{"ResourcesProfile2"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters2 = append(filters2, x)
	x, err = NewFilterRule(MetaGreaterOrEqual, "PddInterval", []string{second.String()})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters2 = append(filters2, x)
	x, err = NewFilterRule(MetaGreaterOrEqual, utils.Usage, []string{second.String()})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters2 = append(filters2, x)
	x, err = NewFilterRule(MetaGreaterOrEqual, "Weight", []string{"15.0"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters2 = append(filters2, x)
	filter10 := &Filter{Tenant: config.CgrConfig().DefaultTenant, ID: "filter10", Rules: filters2}
	dmRES.SetFilter(filter10)
	ref.IndexTPFilter(FilterToTPFilter(filter10), "resourcesprofile2")
	//prefix filter
	x, err = NewFilterRule(MetaPrefix, "Resources", []string{"ResourcesProfilePrefix"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	preffilter = append(preffilter, x)
	x, err = NewFilterRule(MetaGreaterOrEqual, utils.Usage, []string{second.String()})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	preffilter = append(preffilter, x)
	preffilter5 := &Filter{Tenant: config.CgrConfig().DefaultTenant, ID: "preffilter5", Rules: preffilter}
	dmRES.SetFilter(preffilter5)
	ref.IndexTPFilter(FilterToTPFilter(preffilter5), "resourcesprofile3")
	//default filter
	x, err = NewFilterRule(MetaGreaterOrEqual, "Weight", []string{"200.00"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	defaultf = append(defaultf, x)
	x, err = NewFilterRule(MetaGreaterOrEqual, utils.Usage, []string{second.String()})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	defaultf = append(defaultf, x)
	defaultf5 := &Filter{Tenant: config.CgrConfig().DefaultTenant, ID: "defaultf5", Rules: defaultf}
	dmRES.SetFilter(defaultf5)
	ref.IndexTPFilter(FilterToTPFilter(defaultf5), "resourcesprofile4")
	for _, res := range resourceTest {
		dmRES.SetResource(res)
	}
	for _, resp := range resprf {
		dmRES.SetResourceProfile(resp, false)
	}
	err = ref.StoreIndexes(true, utils.NonTransactional)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
}

func TestCachedResourcesForEvent(t *testing.T) {
	data, _ := NewMapStorage()
	dmRES = NewDataManager(data)
	resS := ResourceService{
		dm:      dmRES,
		filterS: &FilterS{dm: dmRES},
	}
	args := &utils.ArgRSv1ResourceUsage{
		CGREvent: *resEvs[0],
		UsageID:  "IDF",
		Units:    10.0,
	}
	val := []*utils.TenantID{
		&utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "RL",
		},
	}
	resources := []*Resource{
		&Resource{
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
				"RU2": &ResourceUsage{
					Tenant:     "cgrates.org",
					ID:         "RU2",
					ExpiryTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
					Units:      2,
				},
			},
			tUsage: utils.Float64Pointer(2),
			dirty:  utils.BoolPointer(true),
		},
	}
	Cache.Set(utils.CacheResources, resources[0].TenantID(),
		resources[0], nil, true, "")
	Cache.Set(utils.CacheEventResources, args.TenantID(),
		val, nil, true, "")
	rcv := resS.cachedResourcesForEvent(args.TenantID())
	if !reflect.DeepEqual(resources[0], rcv[0]) {
		t.Errorf("Expecting: %+v, received: %+v",
			utils.ToJSON(resources[0]), utils.ToJSON(rcv[0]))
	}
}

func TestRSmatchingResourcesForEvent(t *testing.T) {
	mres, err := resserv.matchingResourcesForEvent(resEvs[0], &timeDurationExample)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(resourceTest[0].Tenant, mres[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[0].Tenant, mres[0].Tenant)
	} else if !reflect.DeepEqual(resourceTest[0].ID, mres[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[0].ID, mres[0].ID)
	} else if !reflect.DeepEqual(resourceTest[0].rPrf, mres[0].rPrf) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[0].rPrf, mres[0].rPrf)
	}
	mres, err = resserv.matchingResourcesForEvent(resEvs[1], &timeDurationExample)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(resourceTest[1].Tenant, mres[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[1].Tenant, mres[0].Tenant)
	} else if !reflect.DeepEqual(resourceTest[1].ID, mres[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[1].ID, mres[0].ID)
	} else if !reflect.DeepEqual(resourceTest[1].rPrf, mres[0].rPrf) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[1].rPrf, mres[0].rPrf)
	}
	mres, err = resserv.matchingResourcesForEvent(resEvs[2], &timeDurationExample)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(resourceTest[2].Tenant, mres[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[2].Tenant, mres[0].Tenant)
	} else if !reflect.DeepEqual(resourceTest[2].ID, mres[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[2].ID, mres[0].ID)
	} else if !reflect.DeepEqual(resourceTest[2].rPrf, mres[0].rPrf) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[2].rPrf, mres[0].rPrf)
	}
	mres, err = resserv.matchingResourcesForEvent(resEvs[3], &timeDurationExample)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(resourceTest[3].Tenant, mres[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[3].Tenant, mres[0].Tenant)
	} else if !reflect.DeepEqual(resourceTest[3].ID, mres[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[3].ID, mres[0].ID)
	} else if !reflect.DeepEqual(resourceTest[3].rPrf, mres[0].rPrf) {
		t.Errorf("Expecting: %+v, received: %+v", resourceTest[3].rPrf, mres[0].rPrf)
	}
}

//UsageTTL 0 in ResourceProfile and give 10s duration
func TestRSUsageTTLCase1(t *testing.T) {
	resPrf := &ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "resourcesprofile1",
		FilterIDs: []string{"filter9"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		UsageTTL:          time.Duration(0),
		Limit:             10.00,
		AllocationMessage: "AllocationMessage",
		Blocker:           false,
		Stored:            false,
		Weight:            20.00,
		ThresholdIDs:      []string{""},
	}
	res := &Resource{
		Tenant: "cgrates.org",
		ID:     "resourcesprofile1",
		Usages: map[string]*ResourceUsage{},
		TTLIdx: []string{},
		rPrf:   resPrf,
		ttl:    &timeDurationExample,
	}
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]interface{}{
			"Resources":      "ResourcesProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			"Weight":         "20.0",
			utils.Usage:      time.Duration(135 * time.Second),
			utils.COST:       123.0,
		}}
	if err := dmRES.SetResourceProfile(resPrf, false); err != nil {
		t.Error(err)
	}
	if err := dmRES.SetResource(res); err != nil {
		t.Error(err)
	}
	mres, err := resserv.matchingResourcesForEvent(ev, &timeDurationExample)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(res.Tenant, mres[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", res.Tenant, mres[0].Tenant)
	} else if !reflect.DeepEqual(res.ID, mres[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", res.ID, mres[0].ID)
	} else if !reflect.DeepEqual(res.rPrf, mres[0].rPrf) {
		t.Errorf("Expecting: %+v, received: %+v", res.rPrf, mres[0].rPrf)
	} else if !reflect.DeepEqual(res.ttl, mres[0].ttl) {
		t.Errorf("Expecting: %+v, received: %+v", res.ttl, mres[0].ttl)
	}
}

//UsageTTL 5s in ResourceProfile and give nil duration
func TestRSUsageTTLCase2(t *testing.T) {
	resPrf := &ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "resourcesprofile2",
		FilterIDs: []string{"filter10"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		UsageTTL:          time.Duration(5) * time.Second, // auto-expire the usage after this duration
		Limit:             10.00,                          // limit value
		AllocationMessage: "AllocationMessage",            // message returned by the winning resource on allocation
		Blocker:           false,                          // blocker flag to stop processing on filters matched
		Stored:            false,
		Weight:            20.00,        // Weight to sort the resources
		ThresholdIDs:      []string{""}, // Thresholds to check after changing Limit
	}
	res := &Resource{
		Tenant: "cgrates.org",
		ID:     "resourcesprofile2",
		Usages: map[string]*ResourceUsage{},
		TTLIdx: []string{},
		rPrf:   resPrf,
		ttl:    &resPrf.UsageTTL,
	}
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]interface{}{
			"Resources":      "ResourcesProfile2",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			"Weight":         "20.0",
			utils.Usage:      time.Duration(135 * time.Second),
			utils.COST:       123.0,
		}}
	if err := dmRES.SetResource(res); err != nil {
		t.Error(err)
	}
	if err := dmRES.SetResourceProfile(resPrf, false); err != nil {
		t.Error(err)
	}
	mres, err := resserv.matchingResourcesForEvent(ev, nil)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(res.Tenant, mres[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", res.Tenant, mres[0].Tenant)
	} else if !reflect.DeepEqual(res.ID, mres[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", res.ID, mres[0].ID)
	} else if !reflect.DeepEqual(res.rPrf, mres[0].rPrf) {
		t.Errorf("Expecting: %+v, received: %+v", res.rPrf, mres[0].rPrf)
	} else if !reflect.DeepEqual(res.ttl, mres[0].ttl) {
		t.Errorf("Expecting: %+v, received: %+v", res.ttl, mres[0].ttl)
	}
}

//UsageTTL 5s in ResourceProfile and give 0 duration
func TestRSUsageTTLCase3(t *testing.T) {
	resPrf := &ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "resourcesprofile3", // identifier of this resource
		FilterIDs: []string{"preffilter5"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		UsageTTL:          time.Duration(5) * time.Second, // auto-expire the usage after this duration
		Limit:             10.00,                          // limit value
		AllocationMessage: "AllocationMessage",            // message returned by the winning resource on allocation
		Blocker:           false,                          // blocker flag to stop processing on filters matched
		Stored:            false,
		Weight:            20.00,        // Weight to sort the resources
		ThresholdIDs:      []string{""}, // Thresholds to check after changing Limit
	}
	res := &Resource{
		Tenant: "cgrates.org",
		ID:     "resourcesprofile3",
		Usages: map[string]*ResourceUsage{},
		TTLIdx: []string{},
		rPrf:   resPrf,
		ttl:    nil,
	}
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event3",
		Event: map[string]interface{}{
			"Resources": "ResourcesProfilePrefix",
			utils.Usage: time.Duration(30 * time.Second),
		}}
	if err := dmRES.SetResource(res); err != nil {
		t.Error(err)
	}
	if err := dmRES.SetResourceProfile(resPrf, false); err != nil {
		t.Error(err)
	}
	mres, err := resserv.matchingResourcesForEvent(ev, utils.DurationPointer(time.Duration(0)))
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(res.Tenant, mres[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", res.Tenant, mres[0].Tenant)
	} else if !reflect.DeepEqual(res.ID, mres[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", res.ID, mres[0].ID)
	} else if !reflect.DeepEqual(res.rPrf, mres[0].rPrf) {
		t.Errorf("Expecting: %+v, received: %+v", res.rPrf, mres[0].rPrf)
	} else if !reflect.DeepEqual(res.ttl, mres[0].ttl) {
		t.Errorf("Expecting: %+v, received: %+v", res.ttl, mres[0].ttl)
	}
}

//UsageTTL 5s in ResourceProfile and give 10s duration
func TestRSUsageTTLCase4(t *testing.T) {
	resPrf := &ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "resourcesprofile4", // identifier of this resource
		FilterIDs: []string{"defaultf5"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		UsageTTL:          time.Duration(5) * time.Second, // auto-expire the usage after this duration
		Limit:             10.00,                          // limit value
		AllocationMessage: "AllocationMessage",            // message returned by the winning resource on allocation
		Blocker:           false,                          // blocker flag to stop processing on filters matched
		Stored:            false,
		Weight:            20.00,        // Weight to sort the resources
		ThresholdIDs:      []string{""}, // Thresholds to check after changing Limit
	}
	res := &Resource{
		Tenant: "cgrates.org",
		ID:     "resourcesprofile4",
		Usages: map[string]*ResourceUsage{},
		TTLIdx: []string{},
		rPrf:   resPrf,
		ttl:    &timeDurationExample,
	}
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event3",
		Event: map[string]interface{}{
			"Weight":    "200.0",
			utils.Usage: time.Duration(65 * time.Second),
		}}
	if err := dmRES.SetResource(res); err != nil {
		t.Error(err)
	}
	if err := dmRES.SetResourceProfile(resPrf, false); err != nil {
		t.Error(err)
	}
	mres, err := resserv.matchingResourcesForEvent(ev, &timeDurationExample)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(res.Tenant, mres[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", res.Tenant, mres[0].Tenant)
	} else if !reflect.DeepEqual(res.ID, mres[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", res.ID, mres[0].ID)
	} else if !reflect.DeepEqual(res.rPrf, mres[0].rPrf) {
		t.Errorf("Expecting: %+v, received: %+v", res.rPrf, mres[0].rPrf)
	} else if !reflect.DeepEqual(res.ttl, mres[0].ttl) {
		t.Errorf("Expecting: %+v, received: %+v", res.ttl, mres[0].ttl)
	}
}
