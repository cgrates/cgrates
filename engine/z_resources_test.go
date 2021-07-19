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
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/rpcclient"

	"github.com/cgrates/cgrates/utils"
)

func TestResourceProfileTenantID(t *testing.T) {
	testStruct := ResourceProfile{
		Tenant: "test_tenant",
		ID:     "test_id",
	}
	result := testStruct.TenantID()
	expected := utils.ConcatenatedKey(testStruct.Tenant, testStruct.ID)
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expected, result)
	}
}

func TestResourceUsageTenantID(t *testing.T) {
	testStruct := ResourceUsage{
		Tenant: "test_tenant",
		ID:     "test_id",
	}
	result := testStruct.TenantID()
	expected := utils.ConcatenatedKey(testStruct.Tenant, testStruct.ID)
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expected, result)
	}
}

func TestResourceUsageisActive(t *testing.T) {
	testStruct := ResourceUsage{
		Tenant:     "test_tenant",
		ID:         "test_id",
		ExpiryTime: time.Date(2014, 1, 14, 0, 0, 0, 0, time.UTC),
	}
	result := testStruct.isActive(time.Date(2014, 1, 13, 0, 0, 0, 0, time.UTC))
	if !reflect.DeepEqual(true, result) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", true, result)
	}
}

func TestResourceUsageisActiveFalse(t *testing.T) {
	testStruct := ResourceUsage{
		Tenant:     "test_tenant",
		ID:         "test_id",
		ExpiryTime: time.Date(2014, 1, 14, 0, 0, 0, 0, time.UTC),
	}
	result := testStruct.isActive(time.Date(2014, 1, 15, 0, 0, 0, 0, time.UTC))
	if !reflect.DeepEqual(false, result) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", false, result)
	}
}

func TestResourceUsageClone(t *testing.T) {
	testStruct := &ResourceUsage{
		Tenant:     "test_tenant",
		ID:         "test_id",
		ExpiryTime: time.Date(2014, 1, 14, 0, 0, 0, 0, time.UTC),
	}
	result := testStruct.Clone()
	if !reflect.DeepEqual(testStruct, result) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", testStruct, result)
	}
	testStruct.Tenant = "test_tenant2"
	testStruct.ID = "test_id2"
	testStruct.ExpiryTime = time.Date(2015, 1, 14, 0, 0, 0, 0, time.UTC)
	if reflect.DeepEqual(testStruct.Tenant, result.Tenant) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", testStruct.Tenant, result.Tenant)
	}
	if reflect.DeepEqual(testStruct.ID, result.ID) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", testStruct.ID, result.ID)
	}
	if reflect.DeepEqual(testStruct.ExpiryTime, result.ExpiryTime) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", testStruct.ExpiryTime, result.ExpiryTime)
	}
}

func TestResourceTenantID(t *testing.T) {
	testStruct := Resource{
		Tenant: "test_tenant",
	}
	result := testStruct.TenantID()
	if reflect.DeepEqual(testStruct.Tenant, result) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", testStruct.Tenant, result)
	}
}

func TestResourceTotalUsage1(t *testing.T) {
	testStruct := Resource{
		Tenant: "test_tenant",
		ID:     "test_id",
		Usages: map[string]*ResourceUsage{
			"0": {
				Tenant:     "test_tenant2",
				ID:         "test_id2",
				ExpiryTime: time.Date(2015, 1, 14, 0, 0, 0, 0, time.UTC),
				Units:      1,
			},
			"1": {
				Tenant:     "test_tenant3",
				ID:         "test_id3",
				ExpiryTime: time.Date(2015, 1, 14, 0, 0, 0, 0, time.UTC),
				Units:      2,
			},
		},
	}
	result := testStruct.totalUsage()
	if reflect.DeepEqual(3, result) {
		t.Errorf("\nExpecting <3>,\n Received <%+v>", result)
	}
}

func TestResourceTotalUsage2(t *testing.T) {
	testStruct := Resource{
		Tenant: "test_tenant",
		ID:     "test_id",
		Usages: map[string]*ResourceUsage{
			"0": {
				Tenant:     "test_tenant2",
				ID:         "test_id2",
				ExpiryTime: time.Date(2015, 1, 14, 0, 0, 0, 0, time.UTC),
				Units:      1,
			},
			"1": {
				Tenant:     "test_tenant3",
				ID:         "test_id3",
				ExpiryTime: time.Date(2015, 1, 14, 0, 0, 0, 0, time.UTC),
				Units:      2,
			},
		},
	}
	result := testStruct.TotalUsage()
	if reflect.DeepEqual(3, result) {
		t.Errorf("\nExpecting <3>,\n Received <%+v>", result)
	}
}

func TestResourcesRecordUsage(t *testing.T) {
	testStruct := &Resource{
		Tenant: "test_tenant",
		ID:     "test_id",
		Usages: map[string]*ResourceUsage{
			"test_id2": {
				Tenant:     "test_tenant2",
				ID:         "test_id2",
				ExpiryTime: time.Date(2015, 1, 14, 0, 0, 0, 0, time.UTC),
				Units:      1,
			},
		},
	}
	recordStruct := &ResourceUsage{
		Tenant:     "test_tenant3",
		ID:         "test_id3",
		ExpiryTime: time.Date(2016, 1, 14, 0, 0, 0, 0, time.UTC),
		Units:      1,
	}
	expStruct := Resource{
		Tenant: "test_tenant",
		ID:     "test_id",
		Usages: map[string]*ResourceUsage{
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
	testStruct := &Resource{
		Tenant: "test_tenant",
		ID:     "test_id",
		Usages: map[string]*ResourceUsage{
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
	}
	expStruct := Resource{
		Tenant: "test_tenant",
		ID:     "test_id",
		Usages: map[string]*ResourceUsage{
			"test_id2": {
				Tenant:     "test_tenant2",
				ID:         "test_id2",
				ExpiryTime: time.Date(2015, 1, 14, 0, 0, 0, 0, time.UTC),
				Units:      1,
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
	var r1 *Resource
	var ru1 *ResourceUsage
	var ru2 *ResourceUsage
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
			Tenant:       "cgrates.org",
			ID:           "RL1",
			FilterIDs:    []string{"FLTR_RES_RL1", "*ai:~*req.AnswerTime:2014-07-03T13:43:00Z|2014-07-03T13:44:00Z"},
			Weight:       100,
			Limit:        2,
			ThresholdIDs: []string{"TEST_ACTIONS"},

			UsageTTL:          time.Millisecond,
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
	var r1 *Resource
	var ru1 *ResourceUsage
	var ru2 *ResourceUsage
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
			Tenant:       "cgrates.org",
			ID:           "RL1",
			FilterIDs:    []string{"FLTR_RES_RL1", "*ai:~*req.AnswerTime:2014-07-03T13:43:00Z|2014-07-03T13:44:00Z"},
			Weight:       100,
			Limit:        2,
			ThresholdIDs: []string{"TEST_ACTIONS"},

			UsageTTL:          time.Millisecond,
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
	var r1 *Resource
	var ru1 *ResourceUsage
	var ru2 *ResourceUsage
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
			Tenant:       "cgrates.org",
			ID:           "RL1",
			FilterIDs:    []string{"FLTR_RES_RL1", "*ai:~*req.AnswerTime:2014-07-03T13:43:00Z|2014-07-03T13:44:00Z"},
			Weight:       100,
			Limit:        2,
			ThresholdIDs: []string{"TEST_ACTIONS"},

			UsageTTL:          time.Millisecond,
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
	r1.Usages = map[string]*ResourceUsage{
		ru1.ID: ru1,
	}
	r1.tUsage = nil
	if usedUnits := r1.totalUsage(); usedUnits != 1 {
		t.Errorf("Expecting: %+v, received: %+v", 1, usedUnits)
	}
}

func TestResourceSort(t *testing.T) {
	var r1 *Resource
	var r2 *Resource
	var ru1 *ResourceUsage
	var ru2 *ResourceUsage
	var rs Resources
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
			Tenant:       "cgrates.org",
			ID:           "RL1",
			FilterIDs:    []string{"FLTR_RES_RL1", "*ai:~*req.AnswerTime:2014-07-03T13:43:00Z|2014-07-03T13:44:00Z"},
			Weight:       100,
			Limit:        2,
			ThresholdIDs: []string{"TEST_ACTIONS"},

			UsageTTL:          time.Millisecond,
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
	r2 = &Resource{
		Tenant: "cgrates.org",
		ID:     "RL2",
		rPrf: &ResourceProfile{
			ID:           "RL2",
			FilterIDs:    []string{"FLTR_RES_RL2", "*ai:~*req.AnswerTime:2014-07-03T13:43:00Z|2014-07-03T13:44:00Z"},
			Weight:       50,
			Limit:        2,
			ThresholdIDs: []string{"TEST_ACTIONS"},
			UsageTTL:     time.Millisecond,
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
	var r1 *Resource
	var r2 *Resource
	var ru1 *ResourceUsage
	var ru2 *ResourceUsage
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
			Tenant:       "cgrates.org",
			ID:           "RL1",
			FilterIDs:    []string{"FLTR_RES_RL1", "*ai:~*req.AnswerTime:2014-07-03T13:43:00Z|2014-07-03T13:44:00Z"},
			Weight:       100,
			Limit:        2,
			ThresholdIDs: []string{"TEST_ACTIONS"},

			UsageTTL:          time.Millisecond,
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
	r2 = &Resource{
		Tenant: "cgrates.org",
		ID:     "RL2",
		rPrf: &ResourceProfile{
			ID:           "RL2",
			FilterIDs:    []string{"FLTR_RES_RL2", "*ai:~*req.AnswerTime:2014-07-03T13:43:00Z|2014-07-03T13:44:00Z"},
			Weight:       50,
			Limit:        2,
			ThresholdIDs: []string{"TEST_ACTIONS"},
			UsageTTL:     time.Millisecond,
		},
		// AllocationMessage: "ALLOC2",
		Usages: map[string]*ResourceUsage{
			ru2.ID: ru2,
		},
		tUsage: utils.Float64Pointer(2),
	}

	rs := Resources{r2, r1}
	rs.Sort()

	if rs[0].rPrf.ID != "RL1" {
		t.Error("Sort failed")
	}
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
func TestResourceRecordUsages(t *testing.T) {
	var r1 *Resource
	var r2 *Resource
	var ru1 *ResourceUsage
	var ru2 *ResourceUsage
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
			Tenant:       "cgrates.org",
			ID:           "RL1",
			FilterIDs:    []string{"FLTR_RES_RL1", "*ai:~*req.AnswerTime:2014-07-03T13:43:00Z|2014-07-03T13:44:00Z"},
			Weight:       100,
			Limit:        2,
			ThresholdIDs: []string{"TEST_ACTIONS"},

			UsageTTL:          time.Millisecond,
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
	r2 = &Resource{
		Tenant: "cgrates.org",
		ID:     "RL2",
		rPrf: &ResourceProfile{
			ID:           "RL2",
			FilterIDs:    []string{"FLTR_RES_RL2", "*ai:~*req.AnswerTime:2014-07-03T13:43:00Z|2014-07-03T13:44:00Z"},
			Weight:       50,
			Limit:        2,
			ThresholdIDs: []string{"TEST_ACTIONS"},
			UsageTTL:     time.Millisecond,
		},
		// AllocationMessage: "ALLOC2",
		Usages: map[string]*ResourceUsage{
			ru2.ID: ru2,
		},
		tUsage: utils.Float64Pointer(2),
	}

	rs := Resources{r2, r1}
	rs.Sort()

	if rs[0].rPrf.ID != "RL1" {
		t.Error("Sort failed")
	}
	r1.Usages = map[string]*ResourceUsage{
		ru1.ID: ru1,
	}
	r1.Usages = map[string]*ResourceUsage{
		ru1.ID: ru1,
	}
	r1.tUsage = nil
	if err := rs.recordUsage(ru1); err == nil {
		t.Error("should get duplicated error")
	}
}
func TestResourceAllocateResource(t *testing.T) {
	var r1 *Resource
	var r2 *Resource
	var ru1 *ResourceUsage
	var ru2 *ResourceUsage
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
			Tenant:       "cgrates.org",
			ID:           "RL1",
			FilterIDs:    []string{"FLTR_RES_RL1", "*ai:~*req.AnswerTime:2014-07-03T13:43:00Z|2014-07-03T13:44:00Z"},
			Weight:       100,
			Limit:        2,
			ThresholdIDs: []string{"TEST_ACTIONS"},

			UsageTTL:          time.Millisecond,
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
	r2 = &Resource{
		Tenant: "cgrates.org",
		ID:     "RL2",
		rPrf: &ResourceProfile{
			ID:           "RL2",
			FilterIDs:    []string{"FLTR_RES_RL2", "*ai:~*req.AnswerTime:2014-07-03T13:43:00Z|2014-07-03T13:44:00Z"},
			Weight:       50,
			Limit:        2,
			ThresholdIDs: []string{"TEST_ACTIONS"},
			UsageTTL:     time.Millisecond,
		},
		// AllocationMessage: "ALLOC2",
		Usages: map[string]*ResourceUsage{
			ru2.ID: ru2,
		},
		tUsage: utils.Float64Pointer(2),
	}

	rs := Resources{r2, r1}
	rs.Sort()

	if rs[0].rPrf.ID != "RL1" {
		t.Error("Sort failed")
	}
	r1.Usages = map[string]*ResourceUsage{
		ru1.ID: ru1,
	}
	r1.Usages = map[string]*ResourceUsage{
		ru1.ID: ru1,
	}
	r1.tUsage = nil
	if err := rs.recordUsage(ru1); err == nil {
		t.Error("should get duplicated error")
	}
	rs.clearUsage(ru1.ID)
	rs.clearUsage(ru2.ID)
	ru1.ExpiryTime = time.Now().Add(time.Second)
	ru2.ExpiryTime = time.Now().Add(time.Second)
	if alcMessage, err := rs.allocateResource(context.TODO(), ru1, false); err != nil {
		t.Error(err.Error())
	} else {
		if alcMessage != "ALLOC" {
			t.Errorf("Wrong allocation message: %v", alcMessage)
		}
	}
	if _, err := rs.allocateResource(context.TODO(), ru2, false); err != utils.ErrResourceUnavailable {
		t.Error("Did not receive " + utils.ErrResourceUnavailable.Error() + " error")
	}
	rs[0].rPrf.Limit = 1
	rs[1].rPrf.Limit = 4
	if alcMessage, err := rs.allocateResource(context.TODO(), ru1, false); err != nil {
		t.Error(err.Error())
	} else {
		if alcMessage != "ALLOC" {
			t.Errorf("Wrong allocation message: %v", alcMessage)
		}
	}

	if alcMessage, err := rs.allocateResource(context.TODO(), ru2, false); err != nil {
		t.Error(err.Error())
	} else {
		if alcMessage != "RL2" {
			t.Errorf("Wrong allocation message: %v", alcMessage)
		}
	}

	ru2.Units = 0
	if _, err := rs.allocateResource(context.TODO(), ru2, false); err != nil {
		t.Error(err)
	}
}

// TestRSCacheSetGet assurace the presence of private params in cached resource
func TestRSCacheSetGet(t *testing.T) {
	r := &Resource{
		Tenant: "cgrates.org",
		ID:     "RL",
		rPrf: &ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                "RL",
			FilterIDs:         []string{"FLTR_RES_RL", "*ai:~*req.AnswerTime:2014-07-03T13:43:00Z|2014-07-03T13:44:00Z"},
			AllocationMessage: "ALLOC_RL",
			Weight:            50,
			Limit:             2,
			ThresholdIDs:      []string{"TEST_ACTIONS"},
			UsageTTL:          time.Millisecond,
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
	if err := Cache.Set(context.TODO(), utils.CacheResources, r.TenantID(), r, nil, true, ""); err != nil {
		t.Errorf("Expecting: nil, received: %s", err)
	}
	if x, ok := Cache.Get(utils.CacheResources, r.TenantID()); !ok {
		t.Error("not in cache")
	} else if x == nil {
		t.Error("nil resource")
	} else if !reflect.DeepEqual(r, x.(*Resource)) {
		t.Errorf("Expecting: %+v, received: %+v", r, x)
	}
}
func TestResourceV1AuthorizeResourceMissingStruct(t *testing.T) {
	var dmRES *DataManager
	defaultCfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmRES = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	defaultCfg.ResourceSCfg().StoreInterval = 1
	defaultCfg.ResourceSCfg().StringIndexedFields = nil
	defaultCfg.ResourceSCfg().PrefixIndexedFields = nil
	resService := NewResourceService(dmRES, defaultCfg,
		&FilterS{dm: dmRES, cfg: defaultCfg}, nil)
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
	if err := resService.V1AuthorizeResources(context.TODO(), argsMissingTenant, reply); err != nil && err.Error() != "MANDATORY_IE_MISSING: [Event]" {
		t.Error(err.Error())
	}
	if err := resService.V1AuthorizeResources(context.TODO(), argsMissingUsageID, reply); err != nil && err.Error() != "MANDATORY_IE_MISSING: [Event]" {
		t.Error(err.Error())
	}
}

func TestResourceAddResourceProfile(t *testing.T) {
	var dmRES *DataManager
	defaultCfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmRES = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	defaultCfg.ResourceSCfg().StoreInterval = 1
	defaultCfg.ResourceSCfg().StringIndexedFields = nil
	defaultCfg.ResourceSCfg().PrefixIndexedFields = nil
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
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"9.0"},
			},
		},
	}
	dmRES.SetFilter(context.Background(), fltrRes1, true)
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
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"15.0"},
			},
		},
	}
	dmRES.SetFilter(context.Background(), fltrRes2, true)
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
	dmRES.SetFilter(context.Background(), fltrRes3, true)
	resprf := []*ResourceProfile{
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile1",
			FilterIDs:         []string{"FLTR_RES_1", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile2", // identifier of this resource
			FilterIDs:         []string{"FLTR_RES_2", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile3",
			FilterIDs:         []string{"FLTR_RES_3", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
	}
	resourceTest := Resources{
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

func TestResourceMatchingResourcesForEvent(t *testing.T) {
	var dmRES *DataManager
	defaultCfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmRES = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	defaultCfg.ResourceSCfg().StoreInterval = 1
	defaultCfg.ResourceSCfg().StringIndexedFields = nil
	defaultCfg.ResourceSCfg().PrefixIndexedFields = nil
	resService := NewResourceService(dmRES, defaultCfg,
		&FilterS{dm: dmRES, cfg: defaultCfg}, nil)
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
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"9.0"},
			},
		},
	}
	dmRES.SetFilter(context.Background(), fltrRes1, true)
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
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"15.0"},
			},
		},
	}
	dmRES.SetFilter(context.Background(), fltrRes2, true)
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
	dmRES.SetFilter(context.Background(), fltrRes3, true)
	resprf := []*ResourceProfile{
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile1",
			FilterIDs:         []string{"FLTR_RES_1", "*ai:*now:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile2", // identifier of this resource
			FilterIDs:         []string{"FLTR_RES_2", "*ai:*now:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile3",
			FilterIDs:         []string{"FLTR_RES_3", "*ai:*now:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
	}
	resourceTest := Resources{
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
	resEvs := []*utils.CGREvent{
		{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     "event1",
			Event: map[string]interface{}{
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
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     "event2",
			Event: map[string]interface{}{
				"Resources":      "ResourceProfile2",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				utils.Weight:     "15.0",
				utils.Usage:      45 * time.Second,
			},
		},
		{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     "event3",
			Event: map[string]interface{}{
				"Resources": "ResourceProfilePrefix",
				utils.Usage: 30 * time.Second,
			},
		},
	}
	timeDurationExample := 10 * time.Second
	for _, resProfile := range resprf {
		dmRES.SetResourceProfile(context.TODO(), resProfile, true)
	}
	for _, res := range resourceTest {
		dmRES.SetResource(context.TODO(), res)
	}
	mres, err := resService.matchingResourcesForEvent(context.TODO(), resEvs[0].Tenant, resEvs[0],
		"TestResourceMatchingResourcesForEvent1", &timeDurationExample)
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

	mres, err = resService.matchingResourcesForEvent(context.TODO(), resEvs[1].Tenant, resEvs[1],
		"TestResourceMatchingResourcesForEvent2", &timeDurationExample)
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

	mres, err = resService.matchingResourcesForEvent(context.TODO(), resEvs[2].Tenant, resEvs[2],
		"TestResourceMatchingResourcesForEvent3", &timeDurationExample)
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
}

//UsageTTL 0 in ResourceProfile and give 10s duration
func TestResourceUsageTTLCase1(t *testing.T) {
	resprf := []*ResourceProfile{
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile1",
			FilterIDs:         []string{"FLTR_RES_1", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile2", // identifier of this resource
			FilterIDs:         []string{"FLTR_RES_2", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile3",
			FilterIDs:         []string{"FLTR_RES_3", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
	}
	resourceTest := Resources{
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
	resEvs := []*utils.CGREvent{
		{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     "event1",
			Event: map[string]interface{}{
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
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     "event2",
			Event: map[string]interface{}{
				"Resources":      "ResourceProfile2",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				utils.Weight:     "15.0",
				utils.Usage:      45 * time.Second,
			},
		},
		{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     "event3",
			Event: map[string]interface{}{
				"Resources": "ResourceProfilePrefix",
				utils.Usage: 30 * time.Second,
			},
		},
	}

	timeDurationExample := 10 * time.Second

	var dmRES *DataManager
	defaultCfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmRES = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	defaultCfg.ResourceSCfg().StoreInterval = 1
	defaultCfg.ResourceSCfg().StringIndexedFields = nil
	defaultCfg.ResourceSCfg().PrefixIndexedFields = nil
	resService := NewResourceService(dmRES, defaultCfg,
		&FilterS{dm: dmRES, cfg: defaultCfg}, nil)
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
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"9.0"},
			},
		},
	}
	dmRES.SetFilter(context.Background(), fltrRes1, true)
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
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"15.0"},
			},
		},
	}
	dmRES.SetFilter(context.Background(), fltrRes2, true)
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
	dmRES.SetFilter(context.Background(), fltrRes3, true)
	resprf[0].UsageTTL = 0
	resourceTest[0].rPrf = resprf[0]
	resourceTest[0].ttl = &timeDurationExample
	if err := dmRES.SetResourceProfile(context.TODO(), resprf[0], true); err != nil {
		t.Error(err)
	}
	if err := dmRES.SetResource(context.TODO(), resourceTest[0]); err != nil {
		t.Error(err)
	}
	mres, err := resService.matchingResourcesForEvent(context.TODO(), resEvs[0].Tenant, resEvs[0],
		"TestResourceUsageTTLCase1", &timeDurationExample)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
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
	resprf := []*ResourceProfile{
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile1",
			FilterIDs:         []string{"FLTR_RES_1", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile2", // identifier of this resource
			FilterIDs:         []string{"FLTR_RES_2", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile3",
			FilterIDs:         []string{"FLTR_RES_3", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
	}
	resourceTest := Resources{
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
	resEvs := []*utils.CGREvent{
		{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     "event1",
			Event: map[string]interface{}{
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
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     "event2",
			Event: map[string]interface{}{
				"Resources":      "ResourceProfile2",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				utils.Weight:     "15.0",
				utils.Usage:      45 * time.Second,
			},
		},
		{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     "event3",
			Event: map[string]interface{}{
				"Resources": "ResourceProfilePrefix",
				utils.Usage: 30 * time.Second,
			},
		},
	}

	var dmRES *DataManager
	defaultCfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmRES = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	defaultCfg.ResourceSCfg().StoreInterval = 1
	defaultCfg.ResourceSCfg().StringIndexedFields = nil
	defaultCfg.ResourceSCfg().PrefixIndexedFields = nil
	resService := NewResourceService(dmRES, defaultCfg,
		&FilterS{dm: dmRES, cfg: defaultCfg}, nil)
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
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"9.0"},
			},
		},
	}
	dmRES.SetFilter(context.Background(), fltrRes1, true)
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
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"15.0"},
			},
		},
	}
	dmRES.SetFilter(context.Background(), fltrRes2, true)
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
	dmRES.SetFilter(context.Background(), fltrRes3, true)
	resprf[0].UsageTTL = 0
	resourceTest[0].rPrf = resprf[0]
	resourceTest[0].ttl = &resprf[0].UsageTTL
	if err := dmRES.SetResourceProfile(context.TODO(), resprf[0], true); err != nil {
		t.Error(err)
	}
	if err := dmRES.SetResource(context.TODO(), resourceTest[0]); err != nil {
		t.Error(err)
	}
	mres, err := resService.matchingResourcesForEvent(context.TODO(), resEvs[0].Tenant, resEvs[0],
		"TestResourceUsageTTLCase2", nil)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
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
	resprf := []*ResourceProfile{
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile1",
			FilterIDs:         []string{"FLTR_RES_1", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile2", // identifier of this resource
			FilterIDs:         []string{"FLTR_RES_2", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile3",
			FilterIDs:         []string{"FLTR_RES_3", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
	}
	resourceTest := Resources{
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
	resEvs := []*utils.CGREvent{
		{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     "event1",
			Event: map[string]interface{}{
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
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     "event2",
			Event: map[string]interface{}{
				"Resources":      "ResourceProfile2",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				utils.Weight:     "15.0",
				utils.Usage:      45 * time.Second,
			},
		},
		{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     "event3",
			Event: map[string]interface{}{
				"Resources": "ResourceProfilePrefix",
				utils.Usage: 30 * time.Second,
			},
		},
	}

	var dmRES *DataManager
	defaultCfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmRES = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	defaultCfg.ResourceSCfg().StoreInterval = 1
	defaultCfg.ResourceSCfg().StringIndexedFields = nil
	defaultCfg.ResourceSCfg().PrefixIndexedFields = nil
	resService := NewResourceService(dmRES, defaultCfg,
		&FilterS{dm: dmRES, cfg: defaultCfg}, nil)
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
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"9.0"},
			},
		},
	}
	dmRES.SetFilter(context.Background(), fltrRes1, true)
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
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"15.0"},
			},
		},
	}
	dmRES.SetFilter(context.Background(), fltrRes2, true)
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
	dmRES.SetFilter(context.Background(), fltrRes3, true)
	resprf[0].UsageTTL = 0
	resourceTest[0].rPrf = resprf[0]
	resourceTest[0].ttl = nil
	if err := dmRES.SetResourceProfile(context.TODO(), resprf[0], true); err != nil {
		t.Error(err)
	}
	if err := dmRES.SetResource(context.TODO(), resourceTest[0]); err != nil {
		t.Error(err)
	}
	mres, err := resService.matchingResourcesForEvent(context.TODO(), resEvs[0].Tenant, resEvs[0],
		"TestResourceUsageTTLCase3", utils.DurationPointer(0))
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
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
	resprf := []*ResourceProfile{
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile1",
			FilterIDs:         []string{"FLTR_RES_1", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile2", // identifier of this resource
			FilterIDs:         []string{"FLTR_RES_2", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile3",
			FilterIDs:         []string{"FLTR_RES_3", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
	}
	resourceTest := Resources{
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
	resEvs := []*utils.CGREvent{
		{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     "event1",
			Event: map[string]interface{}{
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
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     "event2",
			Event: map[string]interface{}{
				"Resources":      "ResourceProfile2",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				utils.Weight:     "15.0",
				utils.Usage:      45 * time.Second,
			},
		},
		{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     "event3",
			Event: map[string]interface{}{
				"Resources": "ResourceProfilePrefix",
				utils.Usage: 30 * time.Second,
			},
		},
	}

	var dmRES *DataManager
	defaultCfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmRES = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	defaultCfg.ResourceSCfg().StoreInterval = 1
	defaultCfg.ResourceSCfg().StringIndexedFields = nil
	defaultCfg.ResourceSCfg().PrefixIndexedFields = nil
	resService := NewResourceService(dmRES, defaultCfg,
		&FilterS{dm: dmRES, cfg: defaultCfg}, nil)
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
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"9.0"},
			},
		},
	}
	dmRES.SetFilter(context.Background(), fltrRes1, true)
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
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"15.0"},
			},
		},
	}
	dmRES.SetFilter(context.Background(), fltrRes2, true)
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
	timeDurationExample := 10 * time.Second
	dmRES.SetFilter(context.Background(), fltrRes3, true)
	resprf[0].UsageTTL = 5
	resourceTest[0].rPrf = resprf[0]
	resourceTest[0].ttl = &timeDurationExample
	if err := dmRES.SetResourceProfile(context.TODO(), resprf[0], true); err != nil {
		t.Error(err)
	}
	if err := dmRES.SetResource(context.TODO(), resourceTest[0]); err != nil {
		t.Error(err)
	}
	mres, err := resService.matchingResourcesForEvent(context.TODO(), resEvs[0].Tenant, resEvs[0],
		"TestResourceUsageTTLCase4", &timeDurationExample)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
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

func TestResourceResIDsMp(t *testing.T) {
	resprf := []*ResourceProfile{
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile1",
			FilterIDs:         []string{"FLTR_RES_1", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile2", // identifier of this resource
			FilterIDs:         []string{"FLTR_RES_2", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile3",
			FilterIDs:         []string{"FLTR_RES_3", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
	}
	resourceTest := Resources{
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
	expected := utils.StringSet{
		"ResourceProfile1": struct{}{},
		"ResourceProfile2": struct{}{},
		"ResourceProfile3": struct{}{},
	}
	if rcv := resourceTest.resIDsMp(); !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
}
func TestResourceTenantIDs(t *testing.T) {
	resprf := []*ResourceProfile{
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile1",
			FilterIDs:         []string{"FLTR_RES_1", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile2", // identifier of this resource
			FilterIDs:         []string{"FLTR_RES_2", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile3",
			FilterIDs:         []string{"FLTR_RES_3", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
	}
	resourceTest := Resources{
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
	expected := []string{
		"cgrates.org:ResourceProfile1",
		"cgrates.org:ResourceProfile2",
		"cgrates.org:ResourceProfile3",
	}
	if rcv := resourceTest.tenantIDs(); !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
}

func TestResourceIDs(t *testing.T) {
	resprf := []*ResourceProfile{
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile1",
			FilterIDs:         []string{"FLTR_RES_1", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile2", // identifier of this resource
			FilterIDs:         []string{"FLTR_RES_2", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile3",
			FilterIDs:         []string{"FLTR_RES_3", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
	}
	resourceTest := Resources{
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
	expected := []string{
		"ResourceProfile1",
		"ResourceProfile2",
		"ResourceProfile3",
	}
	if rcv := resourceTest.IDs(); !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
}

func TestResourceMatchWithIndexFalse(t *testing.T) {
	var dmRES *DataManager
	defaultCfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmRES = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	defaultCfg.ResourceSCfg().StoreInterval = 1
	defaultCfg.ResourceSCfg().StringIndexedFields = nil
	defaultCfg.ResourceSCfg().PrefixIndexedFields = nil
	resService := NewResourceService(dmRES, defaultCfg,
		&FilterS{dm: dmRES, cfg: defaultCfg}, nil)
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
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"9.0"},
			},
		},
	}
	dmRES.SetFilter(context.Background(), fltrRes1, true)
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
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"15.0"},
			},
		},
	}
	dmRES.SetFilter(context.Background(), fltrRes2, true)
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
	dmRES.SetFilter(context.Background(), fltrRes3, true)
	resprf := []*ResourceProfile{
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile1",
			FilterIDs:         []string{"FLTR_RES_1", "*ai:*now:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile2", // identifier of this resource
			FilterIDs:         []string{"FLTR_RES_2", "*ai:*now:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile3",
			FilterIDs:         []string{"FLTR_RES_3", "*ai:*now:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
	}
	resourceTest := Resources{
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
	resEvs := []*utils.CGREvent{
		{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     "event1",
			Event: map[string]interface{}{
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
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     "event2",
			Event: map[string]interface{}{
				"Resources":      "ResourceProfile2",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				"PddInterval":    "1s",
				utils.Weight:     "15.0",
				utils.Usage:      45 * time.Second,
			},
		},
		{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     "event3",
			Event: map[string]interface{}{
				"Resources": "ResourceProfilePrefix",
				utils.Usage: 30 * time.Second,
			},
		},
	}
	timeDurationExample := 10 * time.Second
	for _, resProfile := range resprf {
		dmRES.SetResourceProfile(context.TODO(), resProfile, true)
	}
	for _, res := range resourceTest {
		dmRES.SetResource(context.TODO(), res)
	}
	resService.cgrcfg.ResourceSCfg().IndexedSelects = false
	mres, err := resService.matchingResourcesForEvent(context.TODO(), resEvs[0].Tenant, resEvs[0],
		"TestResourceMatchWithIndexFalse1", &timeDurationExample)
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

	mres, err = resService.matchingResourcesForEvent(context.TODO(), resEvs[1].Tenant, resEvs[1],
		"TestResourceMatchWithIndexFalse2", &timeDurationExample)
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

	mres, err = resService.matchingResourcesForEvent(context.TODO(), resEvs[2].Tenant, resEvs[2],
		"TestResourceMatchWithIndexFalse3", &timeDurationExample)
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
}

func TestResourceCaching(t *testing.T) {
	var dmRES *DataManager
	defaultCfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dmRES = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	defaultCfg.ResourceSCfg().StoreInterval = 1
	defaultCfg.ResourceSCfg().StringIndexedFields = nil
	defaultCfg.ResourceSCfg().PrefixIndexedFields = nil
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
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"9.0"},
			},
		},
	}
	dmRES.SetFilter(context.Background(), fltrRes1, true)
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
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Weight,
				Values:  []string{"15.0"},
			},
		},
	}
	dmRES.SetFilter(context.Background(), fltrRes2, true)
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
	dmRES.SetFilter(context.Background(), fltrRes3, true)
	resprf := []*ResourceProfile{
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile1",
			FilterIDs:         []string{"FLTR_RES_1", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile2", // identifier of this resource
			FilterIDs:         []string{"FLTR_RES_2", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
		{
			Tenant:            config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:                "ResourceProfile3",
			FilterIDs:         []string{"FLTR_RES_3", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			UsageTTL:          10 * time.Second,
			Limit:             10.00,
			AllocationMessage: "AllocationMessage",
			Weight:            20.00,
			ThresholdIDs:      []string{""},
		},
	}
	resourceTest := Resources{
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

	for _, resProfile := range resprf {
		dmRES.SetResourceProfile(context.TODO(), resProfile, true)
	}
	for _, res := range resourceTest {
		dmRES.SetResource(context.TODO(), res)
	}
	//clear the cache
	Cache.Clear(nil)
	// start fresh with new dataManager

	dmRES = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)

	defaultCfg.ResourceSCfg().StoreInterval = 1
	defaultCfg.ResourceSCfg().StringIndexedFields = nil
	defaultCfg.ResourceSCfg().PrefixIndexedFields = nil
	resService := NewResourceService(dmRES, defaultCfg,
		&FilterS{dm: dmRES, cfg: defaultCfg}, nil)

	resProf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "ResourceProfileCached",
		FilterIDs:         []string{"*string:~*req.Account:1001", "*ai:*now:2014-07-14T14:25:00Z"},
		UsageTTL:          -1,
		Limit:             10.00,
		AllocationMessage: "AllocationMessage",
		Weight:            20.00,
		ThresholdIDs:      []string{utils.MetaNone},
	}

	if err := Cache.Set(context.TODO(), utils.CacheResourceProfiles, "cgrates.org:ResourceProfileCached",
		resProf, nil, cacheCommit(utils.EmptyString), utils.EmptyString); err != nil {
		t.Errorf("Expecting: nil, received: %s", err)
	}

	res := &Resource{Tenant: resProf.Tenant,
		ID:     resProf.ID,
		Usages: make(map[string]*ResourceUsage)}

	if err := Cache.Set(context.TODO(), utils.CacheResources, "cgrates.org:ResourceProfileCached",
		res, nil, cacheCommit(utils.EmptyString), utils.EmptyString); err != nil {
		t.Errorf("Expecting: nil, received: %s", err)
	}

	resources := Resources{res}
	if err := Cache.Set(context.TODO(), utils.CacheEventResources, "TestResourceCaching", resources.resIDsMp(), nil, true, ""); err != nil {
		t.Errorf("Expecting: nil, received: %s", err)
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Event: map[string]interface{}{
			"Account":     "1001",
			"Destination": "3002"},
	}

	mres, err := resService.matchingResourcesForEvent(context.TODO(), ev.Tenant, ev,
		"TestResourceCaching", nil)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
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

func TestResourcesRemoveExpiredUnitsResetTotalUsage(t *testing.T) {
	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	r := &Resource{
		TTLIdx: []string{"ResGroup1", "ResGroup2", "ResGroup3"},
		Usages: map[string]*ResourceUsage{
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
		tUsage: utils.Float64Pointer(10),
	}

	exp := &Resource{
		TTLIdx: []string{"ResGroup3"},
		Usages: map[string]*ResourceUsage{
			"ResGroup3": {
				Tenant: "cgrates.org",
				ID:     "RU_3",
			},
		},
	}

	explog := "CGRateS <> [WARNING] resetting total usage for resourceID: , usage smaller than 0: -1.000000\n"
	r.removeExpiredUnits()

	if !reflect.DeepEqual(r, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, r)
	}

	rcvlog := buf.String()[20:]
	if rcvlog != explog {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", explog, rcvlog)
	}

	utils.Logger.SetLogLevel(0)
}

func TestResourcesAvailable(t *testing.T) {
	r := ResourceWithConfig{
		Resource: &Resource{
			Usages: map[string]*ResourceUsage{
				"RU_1": {
					Units: 4,
				},
				"RU_2": {
					Units: 7,
				},
			},
		},
		Config: &ResourceProfile{
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
	r := &Resource{
		Usages: map[string]*ResourceUsage{
			"RU_1": {
				Tenant: "cgrates.org",
				ID:     "RU_1",
			},
		},
		ttl: utils.DurationPointer(0),
	}
	ru := &ResourceUsage{
		ID: "RU_2",
	}

	err := r.recordUsage(ru)

	if err != nil {
		t.Error(err)
	}
}

func TestResourcesRecordUsageGtZeroTTL(t *testing.T) {
	r := &Resource{
		Usages: map[string]*ResourceUsage{
			"RU_1": {
				Tenant: "cgrates.org",
				ID:     "RU_1",
			},
		},
		TTLIdx: []string{"RU_1"},
		ttl:    utils.DurationPointer(1 * time.Second),
	}
	ru := &ResourceUsage{
		Tenant: "cgrates.org",
		ID:     "RU_2",
	}

	exp := &Resource{
		Usages: map[string]*ResourceUsage{
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
		ttl:    utils.DurationPointer(1 * time.Second),
	}
	err := r.recordUsage(ru)
	exp.Usages[ru.ID].ExpiryTime = r.Usages[ru.ID].ExpiryTime

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
	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer

	rs := Resources{
		{
			Usages: map[string]*ResourceUsage{
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
			ttl:    utils.DurationPointer(1 * time.Second),
		},
		{
			Usages: map[string]*ResourceUsage{
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
			ttl:    utils.DurationPointer(2 * time.Second),
		},
	}

	ru := &ResourceUsage{
		Tenant: "cgrates.org",
		ID:     "RU_4",
	}

	exp := Resources{
		{
			Usages: map[string]*ResourceUsage{
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
			ttl:    utils.DurationPointer(1 * time.Second),
		},
		{
			Usages: map[string]*ResourceUsage{
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
			ttl:    utils.DurationPointer(2 * time.Second),
		},
	}

	explog := []string{
		fmt.Sprintf("CGRateS <> [WARNING] <%s>cannot record usage, err: duplicate resource usage with id: %s:%s", utils.ResourceS, ru.Tenant, ru.ID),
		fmt.Sprintf("CGRateS <> [WARNING] <%s> cannot clear usage, err: cannot find usage record with id: %s", utils.ResourceS, ru.ID),
	}
	experr := fmt.Sprintf("duplicate resource usage with id: %s", "cgrates.org:"+ru.ID)

	defer log.SetOutput(os.Stderr)
	log.SetOutput(&mockWriter{
		WriteF: func(p []byte) (n int, err error) {
			delete(rs[0].Usages, "RU_4")
			return buf.Write(p)
		},
	})

	err := rs.recordUsage(ru)

	if err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if !reflect.DeepEqual(rs, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(rs))
	}

	rcv := strings.Split(buf.String(), "\n")
	for idx, exp := range explog {
		rcv[idx] = rcv[idx][20:]
		if rcv[idx] != exp {
			t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv[idx])
		}
	}

	utils.Logger.SetLogLevel(0)
}

func TestResourceAllocateResourceOtherDB(t *testing.T) {
	rProf := &ResourceProfile{
		Tenant:       "cgrates.org",
		ID:           "RL_DB",
		FilterIDs:    []string{"*string:~*opts.Resource:RL_DB"},
		Weight:       100,
		Limit:        2,
		ThresholdIDs: []string{utils.MetaNone},
		UsageTTL:     -time.Nanosecond,
	}

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	dm := NewDataManager(NewInternalDB(nil, nil, true), cfg.CacheCfg(), nil)
	fltS := NewFilterS(cfg, nil, dm)
	rs := NewResourceService(dm, cfg, fltS, nil)
	if err := dm.SetResourceProfile(context.TODO(), rProf, true); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetResource(context.TODO(), &Resource{
		Tenant: "cgrates.org",
		ID:     "RL_DB",
		Usages: map[string]*ResourceUsage{
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
	if err := rs.V1AllocateResources(context.TODO(), utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "ef0f554",
			Event:   map[string]interface{}{"": ""},
			APIOpts: map[string]interface{}{"Resource": "RL_DB"},
		},
		UsageID: "56156434-2e44-4f16-a766-086f10b413cd",
		Units:   1,
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != exp {
		t.Errorf("Expected: %q, received: %q", exp, reply)
	}

}

func TestResourceClearUsageErr(t *testing.T) {
	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	rs := Resources{
		{
			Usages: map[string]*ResourceUsage{
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
			ttl:    utils.DurationPointer(1 * time.Second),
		},
	}

	ruTntID := "cgrates.org:RU_3"

	experr := fmt.Sprintf("cannot find usage record with id: %s", ruTntID)
	explog := fmt.Sprintf("CGRateS <> [WARNING] <ResourceS>, clear ruID: %s, err: %s\n", ruTntID, experr)
	err := rs.clearUsage(ruTntID)

	if err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	rcvlog := buf.String()[20:]
	if rcvlog != explog {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", explog, rcvlog)
	}

	utils.Logger.SetLogLevel(0)
}

func TestResourcesAllocateResourceErrRsUnavailable(t *testing.T) {
	rs := Resources{}
	ru := &ResourceUsage{}

	experr := utils.ErrResourceUnavailable
	rcv, err := rs.allocateResource(context.TODO(), ru, false)

	if err == nil || !errors.Is(err, experr) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if rcv != "" {
		t.Errorf("\nexpected empty string, got %s", rcv)
	}
}

func TestResourcesAllocateResourceEmptyConfiguration(t *testing.T) {
	rs := Resources{
		{
			Usages: map[string]*ResourceUsage{
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
			ttl:    utils.DurationPointer(1 * time.Second),
			Tenant: "cgrates.org",
			ID:     "Res_1",
		},
	}

	ru := &ResourceUsage{
		Tenant: "cgrates.org",
		ID:     "RU_2",
	}

	experr := fmt.Sprintf("empty configuration for resourceID: %s", rs[0].TenantID())
	rcv, err := rs.allocateResource(context.TODO(), ru, false)

	if err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if rcv != "" {
		t.Errorf("\nexpected empty string, got %s", rcv)
	}
}

func TestResourcesAllocateResourceDryRun(t *testing.T) {
	rs := Resources{
		{
			Usages: map[string]*ResourceUsage{
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
			ttl:    utils.DurationPointer(1 * time.Second),
			Tenant: "cgrates.org",
			ID:     "Res_1",
			rPrf: &ResourceProfile{
				Tenant: "cgrates.org",
				ID:     "ResGroup1",
			},
		},
	}

	ru := &ResourceUsage{
		Tenant: "cgrates.org",
		ID:     "RU_2",
	}

	exp := "ResGroup1"
	rcv, err := rs.allocateResource(context.TODO(), ru, true)

	if err != nil {
		t.Errorf("\nexpected nil, got %+v", err)
	}

	if rcv != exp {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestResourcesShutdown(t *testing.T) {
	utils.Logger.SetLogLevel(6)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	rS := &ResourceService{
		storedResources: utils.StringSet{
			"Res1": struct{}{},
		},
		stopBackup: make(chan struct{}),
	}

	expLogs := []string{
		fmt.Sprintf("CGRateS <> [INFO] <%s> service shutdown initialized",
			utils.ResourceS),
		fmt.Sprintf("CGRateS <> [WARNING] <%s> failed retrieving from cache resource with ID: %s",
			utils.ResourceS, "Res1"),
		fmt.Sprintf("CGRateS <> [INFO] <%s> service shutdown complete",
			utils.ResourceS),
	}
	exp := utils.StringSet{}
	rS.Shutdown(context.TODO())

	if !reflect.DeepEqual(rS.storedResources, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
			exp, rS.storedResources)
	}

	rcvLogs := strings.Split(buf.String(), "\n")
	rcvLogs = rcvLogs[:len(rcvLogs)-1]

	for idx, rcvLog := range rcvLogs {
		rcvLog := rcvLog[20:]
		if rcvLog != expLogs[idx] {
			t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
				expLogs[idx], rcvLog)
		}
	}

	utils.Logger.SetLogLevel(0)
}

func TestResourcesStoreResources(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	utils.Logger.SetLogLevel(6)
	utils.Logger.SetSyslog(nil)
	defer func() {
		utils.Logger.SetLogLevel(0)
	}()

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	rS := &ResourceService{
		storedResources: utils.StringSet{
			"Res1": struct{}{},
		},
	}

	value := &Resource{
		dirty:  utils.BoolPointer(true),
		Tenant: "cgrates.org",
		ID:     "testResource",
	}

	Cache.SetWithoutReplicate(utils.CacheResources, "Res1", value, nil, true,
		utils.NonTransactional)

	explog := fmt.Sprintf("CGRateS <> [WARNING] <%s> failed saving Resource with ID: %s, error: %s\n",
		utils.ResourceS, value.ID, utils.ErrNoDatabaseConn.Error())
	exp := &ResourceService{
		storedResources: utils.StringSet{
			"Res1": struct{}{},
		},
	}
	rS.storeResources(context.TODO())

	if !reflect.DeepEqual(rS, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rS)
	}

	rcvlog := buf.String()[20:]

	if rcvlog != explog {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", explog, rcvlog)
	}
}

func TestResourcesStoreResourceNotDirty(t *testing.T) {
	rS := &ResourceService{}
	r := &Resource{
		dirty: utils.BoolPointer(false),
	}

	err := rS.StoreResource(context.TODO(), r)

	if err != nil {
		t.Errorf("\nexpected nil, received %+v", err)
	}
}

func TestResourcesStoreResourceOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rS := &ResourceService{
		dm: NewDataManager(NewInternalDB(nil, nil, true), cfg.CacheCfg(), nil),
	}
	r := &Resource{
		dirty: utils.BoolPointer(true),
	}

	err := rS.StoreResource(context.TODO(), r)

	if err != nil {
		t.Errorf("\nexpected nil, received %+v", err)
	}

	if *r.dirty != false {
		t.Errorf("\nexpected false, received %+v", *r.dirty)
	}
}

func TestResourcesStoreResourceErrCache(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	utils.Logger.SetLogLevel(6)
	utils.Logger.SetSyslog(nil)
	defer func() {
		utils.Logger.SetLogLevel(0)
	}()

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	cfg := config.NewDefaultCGRConfig()
	dm := NewDataManager(&DataDBMock{
		SetResourceDrvF: func(ctx *context.Context, r *Resource) error {
			return nil
		}}, cfg.CacheCfg(), nil)
	rS := NewResourceService(dm, cfg, nil, nil)
	Cache = NewCacheS(cfg, dm, nil)
	r := &Resource{
		Tenant: "cgrates.org",
		ID:     "RES1",
		dirty:  utils.BoolPointer(true),
	}
	Cache.Set(context.Background(), utils.CacheResources, r.TenantID(), r, nil, true, "")

	explog := `CGRateS <> [WARNING] <ResourceS> failed caching Resource with ID: cgrates.org:RES1, error: NOT_IMPLEMENTED
`
	if err := rS.StoreResource(context.Background(), r); err == nil ||
		err.Error() != utils.ErrNotImplemented.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotImplemented, err)
	}

	if *r.dirty != true {
		t.Errorf("\nexpected true, received %+v", *r.dirty)
	}

	rcvlog := buf.String()[20:]

	if rcvlog != explog {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", explog, rcvlog)
	}
}

func TestResourcesAllocateResourceEmptyKey(t *testing.T) {
	rs := Resources{
		{
			Usages: map[string]*ResourceUsage{
				"": {},
			},
			rPrf: &ResourceProfile{
				Tenant:            "cgrates.org",
				ID:                "RP_1",
				AllocationMessage: "allocation msg",
			},
		},
	}

	ru := &ResourceUsage{}
	exp := "allocation msg"
	rcv, err := rs.allocateResource(context.TODO(), ru, false)

	if err != nil {
		t.Errorf("\nexpected nil, received %+v", err)
	}

	if rcv != exp {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestResourcesProcessThresholdsNoConns(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	rS := &ResourceService{
		cgrcfg: cfg,
	}
	r := &Resource{
		Tenant: "cgrates.org",
		ID:     "RES_1",
	}
	opts := map[string]interface{}{}

	err := rS.processThresholds(context.TODO(), r, opts)

	if err != nil {
		t.Errorf("\nexpected nil, received %+v", err)
	}
}

func TestResourcesProcessThresholdsOK(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.ResourceSCfg().ThresholdSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}
	Cache = NewCacheS(cfg, nil, nil)

	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.ThresholdSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				exp := &ThresholdsArgsProcessEvent{
					ThresholdIDs: []string{"THD_1"},
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						ID:     args.(*ThresholdsArgsProcessEvent).CGREvent.ID,
						Event: map[string]interface{}{
							utils.EventType:  utils.ResourceUpdate,
							utils.ResourceID: "RES_1",
							utils.Usage:      0.,
						},
						APIOpts: map[string]interface{}{
							utils.MetaEventType: utils.ResourceUpdate,
						},
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
	rS := &ResourceService{
		cgrcfg: cfg,
		connMgr: NewConnManager(cfg, map[string]chan birpc.ClientConnector{
			utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds): rpcInternal,
		}),
	}
	r := &Resource{
		Tenant: "cgrates.org",
		ID:     "RES_1",
		rPrf: &ResourceProfile{
			Tenant:       "cgrates.org",
			ID:           "RP_1",
			ThresholdIDs: []string{"THD_1"},
		},
	}

	err := rS.processThresholds(context.TODO(), r, nil)

	if err != nil {
		t.Errorf("\nexpected nil, received %+v", err)
	}

}

func TestResourcesProcessThresholdsCallErr(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)
	defer func() {
		utils.Logger.SetLogLevel(0)
	}()

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.ResourceSCfg().ThresholdSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}
	Cache = NewCacheS(cfg, nil, nil)

	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.ThresholdSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				exp := &ThresholdsArgsProcessEvent{
					ThresholdIDs: []string{"THD_1"},
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						ID:     args.(*ThresholdsArgsProcessEvent).CGREvent.ID,
						Event: map[string]interface{}{
							utils.EventType:  utils.ResourceUpdate,
							utils.ResourceID: "RES_1",
							utils.Usage:      0.,
						},
						APIOpts: map[string]interface{}{
							utils.MetaEventType: utils.ResourceUpdate,
						},
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
	rS := &ResourceService{
		cgrcfg: cfg,
		connMgr: NewConnManager(cfg, map[string]chan birpc.ClientConnector{
			utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds): rpcInternal,
		}),
	}
	r := &Resource{
		Tenant: "cgrates.org",
		ID:     "RES_1",
		rPrf: &ResourceProfile{
			Tenant:       "cgrates.org",
			ID:           "RP_1",
			ThresholdIDs: []string{"THD_1"},
		},
	}

	experr := utils.ErrExists
	err := rS.processThresholds(context.TODO(), r, nil)

	if err == nil || err != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
	if !strings.Contains(buf.String(), "CGRateS <> [WARNING] <ResourceS> error: EXISTS") {
		t.Errorf("expected log warning")
	}

}

func TestResourcesProcessThresholdsThdConnMetaNone(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ResourceSCfg().ThresholdSConns = []string{"connID"}
	rS := &ResourceService{
		cgrcfg: cfg,
	}
	r := &Resource{
		Tenant: "cgrates.org",
		ID:     "RES_1",
		rPrf: &ResourceProfile{
			ThresholdIDs: []string{utils.MetaNone},
		},
	}
	opts := map[string]interface{}{}

	err := rS.processThresholds(context.TODO(), r, opts)

	if err != nil {
		t.Errorf("\nexpected nil, received: %+v", err)
	}
}

type ccMock struct {
	calls map[string]func(ctx *context.Context, args interface{}, reply interface{}) error
}

func (ccM *ccMock) Call(ctx *context.Context, serviceMethod string, args interface{}, reply interface{}) (err error) {
	if call, has := ccM.calls[serviceMethod]; !has {
		return rpcclient.ErrUnsupporteServiceMethod
	} else {
		return call(ctx, args, reply)
	}
}
func TestResourcesUpdateResource(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dm := NewDataManager(NewInternalDB(nil, nil, true), cfg.CacheCfg(), nil)
	Cache.Clear(nil)
	res := &ResourceProfile{
		Tenant:   "cgrates.org",
		ID:       "RES1",
		UsageTTL: 0,
		Limit:    10,
		Stored:   true,
	}
	r := &Resource{
		Tenant: res.Tenant,
		ID:     res.ID,
		Usages: map[string]*ResourceUsage{
			"jkbdfgs": {
				Tenant:     res.Tenant,
				ID:         "jkbdfgs",
				ExpiryTime: time.Now(),
				Units:      5,
			},
		},
		TTLIdx: []string{"jkbdfgs"},
	}
	expR := &Resource{
		Tenant: res.Tenant,
		ID:     res.ID,
		Usages: make(map[string]*ResourceUsage),
	}
	if err := dm.SetResourceProfile(context.Background(), res, true); err != nil {
		t.Fatal(err)
	}

	if r, err := dm.GetResource(context.Background(), res.Tenant, res.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(r, expR) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expR), utils.ToJSON(r))
	}

	if err := dm.RemoveResource(context.Background(), r.Tenant, r.ID); err != nil {
		t.Fatal(err)
	}

	if err := dm.SetResourceProfile(context.Background(), res, true); err != nil {
		t.Fatal(err)
	}

	if r, err := dm.GetResource(context.Background(), res.Tenant, res.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(r, expR) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expR), utils.ToJSON(r))
	}

	if err := dm.SetResource(context.Background(), r); err != nil {
		t.Fatal(err)
	}

	res = &ResourceProfile{
		Tenant:   "cgrates.org",
		ID:       "RES1",
		UsageTTL: 0,
		Limit:    5,
		Stored:   true,
	}
	if err := dm.SetResourceProfile(context.Background(), res, true); err != nil {
		t.Fatal(err)
	}
	if r, err := dm.GetResource(context.Background(), res.Tenant, res.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(r, expR) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expR), utils.ToJSON(r))
	}

	if err := dm.SetResource(context.Background(), r); err != nil {
		t.Fatal(err)
	}

	res = &ResourceProfile{
		Tenant:   "cgrates.org",
		ID:       "RES1",
		UsageTTL: 10,
		Limit:    5,
		Stored:   true,
	}
	if err := dm.SetResourceProfile(context.Background(), res, true); err != nil {
		t.Fatal(err)
	}
	if r, err := dm.GetResource(context.Background(), res.Tenant, res.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(r, expR) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expR), utils.ToJSON(r))
	}

	if err := dm.SetResource(context.Background(), r); err != nil {
		t.Fatal(err)
	}

	res = &ResourceProfile{
		Tenant:   "cgrates.org",
		ID:       "RES1",
		UsageTTL: 10,
		Limit:    5,
		Stored:   false,
	}
	if err := dm.SetResourceProfile(context.Background(), res, true); err != nil {
		t.Fatal(err)
	}
	if r, err := dm.GetResource(context.Background(), res.Tenant, res.ID, false, false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(r, expR) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expR), utils.ToJSON(r))
	}

	if err := dm.RemoveResourceProfile(context.Background(), res.Tenant, res.ID, true); err != nil {
		t.Fatal(err)
	}

	if _, err := dm.GetResource(context.Background(), res.Tenant, res.ID, false, false, utils.NonTransactional); err != utils.ErrNotFound {
		t.Fatal(err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestResourcesV1ResourcesForEventOK(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	rsPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weight:            10,
		Limit:             10,
		UsageTTL:          time.Minute,
	}
	rs := &Resource{
		rPrf:   rsPrf,
		Tenant: "cgrates.org",
		ID:     "RES1",
		Usages: map[string]*ResourceUsage{
			"RU1": {
				Tenant: "cgrates.org",
				ID:     "RU1",
				Units:  10,
			},
		},
		dirty:  utils.BoolPointer(false),
		tUsage: utils.Float64Pointer(10),
		ttl:    utils.DurationPointer(time.Minute),
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

	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			ID: "ResourcesForEventTest",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
		UsageID: "RU_TEST1",
	}

	exp := Resources{
		{
			rPrf:   rsPrf,
			Tenant: "cgrates.org",
			ID:     "RES1",
			Usages: map[string]*ResourceUsage{
				"RU1": {
					Tenant: "cgrates.org",
					ID:     "RU1",
					Units:  10,
				},
			},
			dirty:  utils.BoolPointer(false),
			tUsage: utils.Float64Pointer(10),
			ttl:    utils.DurationPointer(time.Minute),
			TTLIdx: []string{},
		},
	}
	var reply Resources
	if err := rS.V1ResourcesForEvent(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp), utils.ToJSON(reply))
	}
}

func TestResourcesV1ResourcesForEventNotFound(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	rsPrf := &ResourceProfile{
		Tenant:       "cgrates.org",
		ID:           "RES1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		ThresholdIDs: []string{utils.MetaNone},
		Weight:       10,
		Limit:        10,
		UsageTTL:     time.Minute,
	}
	rs := &Resource{
		rPrf:   rsPrf,
		Tenant: "cgrates.org",
		ID:     "RES1",
		Usages: map[string]*ResourceUsage{
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

	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ResourcesForEventTest",
			Event: map[string]interface{}{
				utils.AccountField: "1002",
			},
		},
		UsageID: "RU_TEST1",
	}

	var reply Resources
	if err := rS.V1ResourcesForEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestResourcesV1ResourcesForEventMissingParameters(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	rsPrf := &ResourceProfile{
		Tenant:       "cgrates.org",
		ID:           "RES1",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		ThresholdIDs: []string{utils.MetaNone},
		Weight:       10,
		Limit:        10,
		UsageTTL:     time.Minute,
	}
	rs := &Resource{
		rPrf:   rsPrf,
		Tenant: "cgrates.org",
		ID:     "RES1",
		Usages: map[string]*ResourceUsage{
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

	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := utils.ArgRSv1ResourceUsage{
		UsageID: "RU_TEST1",
	}

	experr := `MANDATORY_IE_MISSING: [Event]`
	var reply Resources
	if err := rS.V1ResourcesForEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	args = utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
		UsageID: "RU_TEST2",
	}

	experr = `MANDATORY_IE_MISSING: [ID]`
	if err := rS.V1ResourcesForEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	args = utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ResourcesForEventTest",
		},
		UsageID: "RU_TEST3",
	}

	experr = `MANDATORY_IE_MISSING: [Event]`
	if err := rS.V1ResourcesForEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	args = utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ResourcesForEventTest",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
	}

	experr = `MANDATORY_IE_MISSING: [UsageID]`
	if err := rS.V1ResourcesForEvent(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestResourcesV1ResourcesForEventCacheReplyExists(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	config.SetCgrConfig(cfg)
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	cacheKey := utils.ConcatenatedKey(utils.ResourceSv1GetResourcesForEvent,
		utils.ConcatenatedKey("cgrates.org", "ResourcesForEventTest"))
	rsPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weight:            10,
		Limit:             10,
		UsageTTL:          time.Minute,
	}
	rs := &Resource{
		rPrf:   rsPrf,
		Tenant: "cgrates.org",
		ID:     "RES1",
		Usages: map[string]*ResourceUsage{
			"RU1": {
				Tenant: "cgrates.org",
				ID:     "RU1",
				Units:  10,
			},
		},
		dirty:  utils.BoolPointer(false),
		tUsage: utils.Float64Pointer(10),
		ttl:    utils.DurationPointer(time.Minute),
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

	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			ID: "ResourcesForEventTest",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
		UsageID: "RU_TEST1",
	}

	cacheReply := Resources{
		{
			rPrf:   rsPrf,
			Tenant: "cgrates.org",
			ID:     "RES1",
			Usages: map[string]*ResourceUsage{
				"RU1": {
					Tenant: "cgrates.org",
					ID:     "RU1",
					Units:  10,
				},
			},
			dirty:  utils.BoolPointer(false),
			tUsage: utils.Float64Pointer(10),
			ttl:    utils.DurationPointer(time.Minute),
			TTLIdx: []string{},
		},
	}
	Cache.Set(context.Background(), utils.CacheRPCResponses, cacheKey,
		&utils.CachedRPCResponse{Result: &cacheReply, Error: nil},
		nil, true, utils.NonTransactional)
	var reply Resources
	if err := rS.V1ResourcesForEvent(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, cacheReply) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(cacheReply), utils.ToJSON(reply))
	}

	config.SetCgrConfig(config.NewDefaultCGRConfig())
}

func TestResourcesV1ResourcesForEventCacheReplySet(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	config.SetCgrConfig(cfg)
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	cacheKey := utils.ConcatenatedKey(utils.ResourceSv1GetResourcesForEvent,
		utils.ConcatenatedKey("cgrates.org", "ResourcesForEventTest"))
	rsPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weight:            10,
		Limit:             10,
		UsageTTL:          time.Minute,
	}
	rs := &Resource{
		rPrf:   rsPrf,
		Tenant: "cgrates.org",
		ID:     "RES1",
		Usages: map[string]*ResourceUsage{
			"RU1": {
				Tenant: "cgrates.org",
				ID:     "RU1",
				Units:  10,
			},
		},
		dirty:  utils.BoolPointer(false),
		tUsage: utils.Float64Pointer(10),
		ttl:    utils.DurationPointer(time.Minute),
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

	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			ID: "ResourcesForEventTest",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
		UsageID: "RU_TEST1",
	}

	exp := Resources{
		{
			rPrf:   rsPrf,
			Tenant: "cgrates.org",
			ID:     "RES1",
			Usages: map[string]*ResourceUsage{
				"RU1": {
					Tenant: "cgrates.org",
					ID:     "RU1",
					Units:  10,
				},
			},
			dirty:  utils.BoolPointer(false),
			tUsage: utils.Float64Pointer(10),
			ttl:    utils.DurationPointer(time.Minute),
			TTLIdx: []string{},
		},
	}
	var reply Resources
	if err := rS.V1ResourcesForEvent(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp), utils.ToJSON(reply))
	}

	if itm, has := Cache.Get(utils.CacheRPCResponses, cacheKey); has {
		resp := itm.(*utils.CachedRPCResponse)
		if !reflect.DeepEqual(*resp.Result.(*Resources), exp) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, *resp.Result.(*Resources))
		}
	}

	config.SetCgrConfig(config.NewDefaultCGRConfig())
}

func TestResourcesV1GetResourceOK(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)

	rsPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weight:            10,
		Limit:             10,
		UsageTTL:          time.Minute,
	}
	rs := &Resource{
		rPrf:   rsPrf,
		Tenant: "cgrates.org",
		ID:     "RES1",
		Usages: map[string]*ResourceUsage{
			"RU1": {
				Tenant: "cgrates.org",
				ID:     "RU1",
				Units:  10,
			},
		},
		dirty:  utils.BoolPointer(false),
		tUsage: utils.Float64Pointer(10),
		ttl:    utils.DurationPointer(time.Minute),
		TTLIdx: []string{},
	}
	err := dm.SetResource(context.Background(), rs)
	if err != nil {
		t.Error(err)
	}

	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	exp := Resource{
		rPrf:   rsPrf,
		Tenant: "cgrates.org",
		ID:     "RES1",
		Usages: map[string]*ResourceUsage{
			"RU1": {
				Tenant: "cgrates.org",
				ID:     "RU1",
				Units:  10,
			},
		},
		dirty:  utils.BoolPointer(false),
		tUsage: utils.Float64Pointer(10),
		ttl:    utils.DurationPointer(time.Minute),
		TTLIdx: []string{},
	}

	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "RES1",
		},
	}
	var reply Resource
	if err := rS.V1GetResource(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp), utils.ToJSON(reply))
	}
}

func TestResourcesV1GetResourceNotFound(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)

	rsPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weight:            10,
		Limit:             10,
		UsageTTL:          time.Minute,
	}
	rs := &Resource{
		rPrf:   rsPrf,
		Tenant: "cgrates.org",
		ID:     "RES1",
		Usages: map[string]*ResourceUsage{
			"RU1": {
				Tenant: "cgrates.org",
				ID:     "RU1",
				Units:  10,
			},
		},
		dirty:  utils.BoolPointer(false),
		tUsage: utils.Float64Pointer(10),
		ttl:    utils.DurationPointer(time.Minute),
		TTLIdx: []string{},
	}
	err := dm.SetResource(context.Background(), rs)
	if err != nil {
		t.Error(err)
	}

	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "RES2",
		},
	}
	var reply Resource
	if err := rS.V1GetResource(context.Background(), args, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestResourcesV1GetResourceMissingParameters(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)

	rsPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weight:            10,
		Limit:             10,
		UsageTTL:          time.Minute,
	}
	rs := &Resource{
		rPrf:   rsPrf,
		Tenant: "cgrates.org",
		ID:     "RES1",
		Usages: map[string]*ResourceUsage{
			"RU1": {
				Tenant: "cgrates.org",
				ID:     "RU1",
				Units:  10,
			},
		},
		dirty:  utils.BoolPointer(false),
		tUsage: utils.Float64Pointer(10),
		ttl:    utils.DurationPointer(time.Minute),
		TTLIdx: []string{},
	}
	err := dm.SetResource(context.Background(), rs)
	if err != nil {
		t.Error(err)
	}

	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{},
	}

	experr := `MANDATORY_IE_MISSING: [ID]`
	var reply Resource
	if err := rS.V1GetResource(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestResourcesV1GetResourceWithConfigOK(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)

	rsPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weight:            10,
		Limit:             10,
		UsageTTL:          time.Minute,
	}
	rs := &Resource{
		rPrf:   rsPrf,
		Tenant: "cgrates.org",
		ID:     "RES1",
		Usages: map[string]*ResourceUsage{
			"RU1": {
				Tenant: "cgrates.org",
				ID:     "RU1",
				Units:  10,
			},
		},
		dirty:  utils.BoolPointer(false),
		tUsage: utils.Float64Pointer(10),
		ttl:    utils.DurationPointer(time.Minute),
		TTLIdx: []string{},
	}
	err := dm.SetResource(context.Background(), rs)
	if err != nil {
		t.Error(err)
	}

	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	exp := ResourceWithConfig{
		Resource: &Resource{
			rPrf:   rsPrf,
			Tenant: "cgrates.org",
			ID:     "RES1",
			Usages: map[string]*ResourceUsage{
				"RU1": {
					Tenant: "cgrates.org",
					ID:     "RU1",
					Units:  10,
				},
			},
			dirty:  utils.BoolPointer(false),
			tUsage: utils.Float64Pointer(10),
			ttl:    utils.DurationPointer(time.Minute),
			TTLIdx: []string{},
		},
		Config: rsPrf,
	}

	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "RES1",
		},
	}
	var reply ResourceWithConfig
	if err := rS.V1GetResourceWithConfig(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp), utils.ToJSON(reply))
	}
}

func TestResourcesV1GetResourceWithConfigNilrPrfOK(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)

	rsPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weight:            10,
		Limit:             10,
		UsageTTL:          time.Minute,
	}
	err := dm.SetResourceProfile(context.Background(), rsPrf, true)
	if err != nil {
		t.Error(err)
	}
	rs := &Resource{
		Tenant: "cgrates.org",
		ID:     "RES1",
		Usages: map[string]*ResourceUsage{
			"RU1": {
				Tenant: "cgrates.org",
				ID:     "RU1",
				Units:  10,
			},
		},
		dirty:  utils.BoolPointer(false),
		tUsage: utils.Float64Pointer(10),
		ttl:    utils.DurationPointer(time.Minute),
		TTLIdx: []string{},
	}
	err = dm.SetResource(context.Background(), rs)
	if err != nil {
		t.Error(err)
	}

	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	exp := ResourceWithConfig{
		Resource: &Resource{
			rPrf:   rsPrf,
			Tenant: "cgrates.org",
			ID:     "RES1",
			Usages: map[string]*ResourceUsage{
				"RU1": {
					Tenant: "cgrates.org",
					ID:     "RU1",
					Units:  10,
				},
			},
			dirty:  utils.BoolPointer(false),
			tUsage: utils.Float64Pointer(10),
			ttl:    utils.DurationPointer(time.Minute),
			TTLIdx: []string{},
		},
		Config: rsPrf,
	}

	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "RES1",
		},
	}
	var reply ResourceWithConfig
	if err := rS.V1GetResourceWithConfig(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp), utils.ToJSON(reply))
	}
}

func TestResourcesV1GetResourceWithConfigNilrPrfProfileNotFound(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)

	rsPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES2",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weight:            10,
		Limit:             10,
		UsageTTL:          time.Minute,
	}
	err := dm.SetResourceProfile(context.Background(), rsPrf, true)
	if err != nil {
		t.Error(err)
	}
	rs := &Resource{
		Tenant: "cgrates.org",
		ID:     "RES1",
		Usages: map[string]*ResourceUsage{
			"RU1": {
				Tenant: "cgrates.org",
				ID:     "RU1",
				Units:  10,
			},
		},
		dirty:  utils.BoolPointer(false),
		tUsage: utils.Float64Pointer(10),
		ttl:    utils.DurationPointer(time.Minute),
		TTLIdx: []string{},
	}
	err = dm.SetResource(context.Background(), rs)
	if err != nil {
		t.Error(err)
	}

	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "RES1",
		},
	}
	var reply ResourceWithConfig
	if err := rS.V1GetResourceWithConfig(context.Background(), args, &reply); err == nil ||
		err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestResourcesV1GetResourceWithConfigResourceNotFound(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)

	rsPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES2",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weight:            10,
		Limit:             10,
		UsageTTL:          time.Minute,
	}
	rs := &Resource{
		rPrf:   rsPrf,
		Tenant: "cgrates.org",
		ID:     "RES2",
		Usages: map[string]*ResourceUsage{
			"RU1": {
				Tenant: "cgrates.org",
				ID:     "RU1",
				Units:  10,
			},
		},
		dirty:  utils.BoolPointer(false),
		tUsage: utils.Float64Pointer(10),
		ttl:    utils.DurationPointer(time.Minute),
		TTLIdx: []string{},
	}
	err := dm.SetResource(context.Background(), rs)
	if err != nil {
		t.Error(err)
	}

	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "RES1",
		},
	}
	var reply ResourceWithConfig
	if err := rS.V1GetResourceWithConfig(context.Background(), args, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestResourcesV1GetResourceWithConfigMissingParameters(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)

	rsPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weight:            10,
		Limit:             10,
		UsageTTL:          time.Minute,
	}
	rs := &Resource{
		rPrf:   rsPrf,
		Tenant: "cgrates.org",
		ID:     "RES1",
		Usages: map[string]*ResourceUsage{
			"RU1": {
				Tenant: "cgrates.org",
				ID:     "RU1",
				Units:  10,
			},
		},
		dirty:  utils.BoolPointer(false),
		tUsage: utils.Float64Pointer(10),
		ttl:    utils.DurationPointer(time.Minute),
		TTLIdx: []string{},
	}
	err := dm.SetResource(context.Background(), rs)
	if err != nil {
		t.Error(err)
	}

	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	experr := `MANDATORY_IE_MISSING: [ID]`
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{},
	}
	var reply ResourceWithConfig
	if err := rS.V1GetResourceWithConfig(context.Background(), args, &reply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestResourcesV1AuthorizeResourcesOK(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)

	rsPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weight:            10,
		Limit:             10,
		UsageTTL:          time.Minute,
	}

	err := dm.SetResourceProfile(context.Background(), rsPrf, true)
	if err != nil {
		t.Error(err)
	}

	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			ID: "EventAuthorizeResource",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
		UsageID:  "RU_Test",
		UsageTTL: utils.DurationPointer(time.Minute),
		Units:    5,
	}
	var reply string

	if err := rS.V1AuthorizeResources(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != "Approved" {
		t.Errorf("Unexpected reply returned: %q", reply)
	}
}

func TestResourcesV1AuthorizeResourcesNotAuthorized(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)

	rsPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weight:            10,
		Limit:             0,
		UsageTTL:          time.Minute,
	}

	err := dm.SetResourceProfile(context.Background(), rsPrf, true)
	if err != nil {
		t.Error(err)
	}

	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "EventAuthorizeResource",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
		UsageID:  "RU_Test",
		UsageTTL: utils.DurationPointer(time.Minute),
		Units:    5,
	}
	var reply string

	if err := rS.V1AuthorizeResources(context.Background(), args, &reply); err == nil ||
		err != utils.ErrResourceUnauthorized {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrResourceUnauthorized, err)
	}
}

func TestResourcesV1AuthorizeResourcesNoMatch(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)

	rsPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weight:            10,
		Limit:             10,
		UsageTTL:          time.Minute,
	}

	err := dm.SetResourceProfile(context.Background(), rsPrf, true)
	if err != nil {
		t.Error(err)
	}

	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "EventAuthorizeResource",
			Event: map[string]interface{}{
				utils.AccountField: "1002",
			},
		},
		UsageID:  "RU_Test",
		UsageTTL: utils.DurationPointer(time.Minute),
		Units:    5,
	}
	var reply string

	if err := rS.V1AuthorizeResources(context.Background(), args, &reply); err == nil ||
		err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestResourcesV1AuthorizeResourcesNilCGREvent(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)

	rsPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weight:            10,
		Limit:             10,
		UsageTTL:          time.Minute,
	}

	err := dm.SetResourceProfile(context.Background(), rsPrf, true)
	if err != nil {
		t.Error(err)
	}

	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := utils.ArgRSv1ResourceUsage{
		UsageID:  "RU_Test",
		UsageTTL: utils.DurationPointer(time.Minute),
		Units:    5,
	}
	experr := `MANDATORY_IE_MISSING: [Event]`
	var reply string

	if err := rS.V1AuthorizeResources(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestResourcesV1AuthorizeResourcesMissingUsageID(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)

	rsPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weight:            10,
		Limit:             10,
		UsageTTL:          time.Minute,
	}

	err := dm.SetResourceProfile(context.Background(), rsPrf, true)
	if err != nil {
		t.Error(err)
	}

	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			ID: "EventAuthorizeResource",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
		UsageTTL: utils.DurationPointer(time.Minute),
		Units:    5,
	}
	experr := `MANDATORY_IE_MISSING: [UsageID]`
	var reply string

	if err := rS.V1AuthorizeResources(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestResourcesV1AuthorizeResourcesCacheReplyExists(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	config.SetCgrConfig(cfg)
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	cacheKey := utils.ConcatenatedKey(utils.ResourceSv1AuthorizeResources,
		utils.ConcatenatedKey("cgrates.org", "EventAuthorizeResource"))

	rsPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weight:            10,
		Limit:             10,
		UsageTTL:          time.Minute,
	}
	rs := &Resource{
		rPrf:   rsPrf,
		Tenant: "cgrates.org",
		ID:     "RES1",
		Usages: map[string]*ResourceUsage{
			"RU1": {
				Tenant: "cgrates.org",
				ID:     "RU1",
				Units:  10,
			},
		},
		dirty:  utils.BoolPointer(false),
		tUsage: utils.Float64Pointer(10),
		ttl:    utils.DurationPointer(time.Minute),
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

	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			ID: "EventAuthorizeResource",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
		UsageID:  "RU_Test",
		UsageTTL: utils.DurationPointer(time.Minute),
		Units:    5,
	}

	cacheReply := "Approved"
	Cache.Set(context.Background(), utils.CacheRPCResponses, cacheKey,
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
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	config.SetCgrConfig(cfg)
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	cacheKey := utils.ConcatenatedKey(utils.ResourceSv1AuthorizeResources,
		utils.ConcatenatedKey("cgrates.org", "EventAuthorizeResource"))

	rsPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weight:            10,
		Limit:             -1,
		UsageTTL:          time.Minute,
	}
	rs := &Resource{
		rPrf:   rsPrf,
		Tenant: "cgrates.org",
		ID:     "RES1",
		Usages: map[string]*ResourceUsage{
			"RU1": {
				Tenant: "cgrates.org",
				ID:     "RU1",
				Units:  4,
			},
		},
		dirty:  utils.BoolPointer(false),
		tUsage: utils.Float64Pointer(10),
		ttl:    utils.DurationPointer(time.Minute),
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

	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			ID: "EventAuthorizeResource",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
		UsageID:  "RU_Test",
		UsageTTL: utils.DurationPointer(time.Minute),
		Units:    2,
	}

	var reply string
	if err := rS.V1AuthorizeResources(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != "Approved" {
		t.Errorf("Unexpected reply returned: %q", reply)
	}

	if itm, has := Cache.Get(utils.CacheRPCResponses, cacheKey); has {
		resp := itm.(*utils.CachedRPCResponse)
		if *resp.Result.(*string) != "Approved" {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				"Approved", *resp.Result.(*string))
		}
	}

	config.SetCgrConfig(config.NewDefaultCGRConfig())
}

func TestResourcesV1AllocateResourcesOK(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)

	rsPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weight:            10,
		Limit:             10,
		UsageTTL:          time.Minute,
	}

	err := dm.SetResourceProfile(context.Background(), rsPrf, true)
	if err != nil {
		t.Error(err)
	}

	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			ID: "EventAuthorizeResource",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
		UsageID:  "RU_Test",
		UsageTTL: utils.DurationPointer(time.Minute),
		Units:    5,
	}
	var reply string

	if err := rS.V1AllocateResources(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != "Approved" {
		t.Errorf("Unexpected reply returned: %q", reply)
	}
}

func TestResourcesV1AllocateResourcesNoMatch(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)

	rsPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weight:            10,
		Limit:             10,
		UsageTTL:          time.Minute,
	}

	err := dm.SetResourceProfile(context.Background(), rsPrf, true)
	if err != nil {
		t.Error(err)
	}

	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			ID: "EventAuthorizeResource",
			Event: map[string]interface{}{
				utils.AccountField: "1002",
			},
		},
		UsageID:  "RU_Test",
		UsageTTL: utils.DurationPointer(time.Minute),
		Units:    5,
	}
	var reply string

	if err := rS.V1AllocateResources(context.Background(), args, &reply); err == nil ||
		err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestResourcesV1AllocateResourcesMissingParameters(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)

	rsPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weight:            10,
		Limit:             10,
		UsageTTL:          time.Minute,
	}

	err := dm.SetResourceProfile(context.Background(), rsPrf, true)
	if err != nil {
		t.Error(err)
	}

	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			ID: "EventAuthorizeResource",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
		UsageTTL: utils.DurationPointer(time.Minute),
		Units:    5,
	}
	var reply string

	experr := `MANDATORY_IE_MISSING: [UsageID]`
	if err := rS.V1AllocateResources(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	args = utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			ID: "EventAuthorizeResource",
		},
		UsageID:  "RU_Test",
		UsageTTL: utils.DurationPointer(time.Minute),
		Units:    5,
	}

	experr = `MANDATORY_IE_MISSING: [Event]`
	if err := rS.V1AllocateResources(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	args = utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
		UsageID:  "RU_Test",
		UsageTTL: utils.DurationPointer(time.Minute),
		Units:    5,
	}

	experr = `MANDATORY_IE_MISSING: [ID]`
	if err := rS.V1AllocateResources(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	args = utils.ArgRSv1ResourceUsage{
		UsageID:  "RU_Test",
		UsageTTL: utils.DurationPointer(time.Minute),
		Units:    5,
	}

	experr = `MANDATORY_IE_MISSING: [Event]`
	if err := rS.V1AllocateResources(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestResourcesV1AllocateResourcesCacheReplyExists(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	config.SetCgrConfig(cfg)
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	cacheKey := utils.ConcatenatedKey(utils.ResourceSv1AllocateResources,
		utils.ConcatenatedKey("cgrates.org", "EventAllocateResource"))

	rsPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weight:            10,
		Limit:             -1,
		UsageTTL:          time.Minute,
	}
	rs := &Resource{
		rPrf:   rsPrf,
		Tenant: "cgrates.org",
		ID:     "RES1",
		Usages: map[string]*ResourceUsage{
			"RU1": {
				Tenant: "cgrates.org",
				ID:     "RU1",
				Units:  10,
			},
		},
		dirty:  utils.BoolPointer(false),
		tUsage: utils.Float64Pointer(10),
		ttl:    utils.DurationPointer(time.Minute),
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

	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			ID: "EventAllocateResource",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
		UsageID:  "RU_Test",
		UsageTTL: utils.DurationPointer(time.Minute),
		Units:    5,
	}

	cacheReply := "cacheApproved"
	Cache.Set(context.Background(), utils.CacheRPCResponses, cacheKey,
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
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	config.SetCgrConfig(cfg)
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	cacheKey := utils.ConcatenatedKey(utils.ResourceSv1AllocateResources,
		utils.ConcatenatedKey("cgrates.org", "EventAllocateResource"))

	rsPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weight:            10,
		Limit:             -1,
		UsageTTL:          time.Minute,
	}
	rs := &Resource{
		rPrf:   rsPrf,
		Tenant: "cgrates.org",
		ID:     "RES1",
		Usages: map[string]*ResourceUsage{
			"RU1": {
				Tenant: "cgrates.org",
				ID:     "RU1",
				Units:  4,
			},
		},
		dirty:  utils.BoolPointer(false),
		tUsage: utils.Float64Pointer(10),
		ttl:    utils.DurationPointer(time.Minute),
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

	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			ID: "EventAllocateResource",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
		UsageID:  "RU_Test",
		UsageTTL: utils.DurationPointer(time.Minute),
		Units:    2,
	}

	var reply string
	if err := rS.V1AllocateResources(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != "Approved" {
		t.Errorf("Unexpected reply returned: %q", reply)
	}

	if itm, has := Cache.Get(utils.CacheRPCResponses, cacheKey); has {
		resp := itm.(*utils.CachedRPCResponse)
		if *resp.Result.(*string) != "Approved" {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				"Approved", *resp.Result.(*string))
		}
	}

	config.SetCgrConfig(config.NewDefaultCGRConfig())
}

func TestResourcesV1AllocateResourcesResAllocErr(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)

	rsPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "",
		Weight:            10,
		Limit:             -1,
		UsageTTL:          time.Minute,
	}
	err := dm.SetResourceProfile(context.Background(), rsPrf, true)
	if err != nil {
		t.Error(err)
	}

	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			ID: "EventAuthorizeResource",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
		UsageID:  "RU_Test",
		UsageTTL: utils.DurationPointer(time.Minute),
		Units:    5,
	}
	var reply string

	if err := rS.V1AllocateResources(context.Background(), args, &reply); err == nil ||
		err != utils.ErrResourceUnavailable {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrResourceUnavailable, err)
	}
}

func TestResourcesV1AllocateResourcesProcessThErr(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.ResourceSCfg().StoreInterval = 2
	cfg.ResourceSCfg().ThresholdSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)

	rsPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)},
		AllocationMessage: "Approved",
		Weight:            10,
		Limit:             -1,
		UsageTTL:          time.Minute,
	}
	rs := &Resource{
		rPrf:   rsPrf,
		Tenant: "cgrates.org",
		ID:     "RES1",
		Usages: map[string]*ResourceUsage{
			"RU1": {
				Tenant: "cgrates.org",
				ID:     "RU1",
				Units:  10,
			},
		},
		dirty:  utils.BoolPointer(false),
		tUsage: utils.Float64Pointer(10),
		ttl:    utils.DurationPointer(time.Minute),
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
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.ThresholdSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				return utils.ErrExists
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	cM := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds): rpcInternal,
	})
	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, cM)

	args := utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			ID: "EventAuthorizeResource",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
		UsageID:  "RU_Test",
		UsageTTL: utils.DurationPointer(time.Minute),
		Units:    5,
	}
	var reply string

	if err := rS.V1AllocateResources(context.Background(), args, &reply); err == nil ||
		err != utils.ErrExists {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrExists, err)
	}
	dm.DataDB().Flush(utils.EmptyString)
}

func TestResourcesV1ReleaseResourcesOK(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)

	rsPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weight:            10,
		Limit:             10,
		UsageTTL:          time.Minute,
	}

	err := dm.SetResourceProfile(context.Background(), rsPrf, true)
	if err != nil {
		t.Error(err)
	}

	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			ID: "EventAuthorizeResource",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
		UsageID:  "RU_Test",
		UsageTTL: utils.DurationPointer(time.Minute),
		Units:    5,
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
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)

	rsPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weight:            10,
		Limit:             10,
		UsageTTL:          0,
	}

	err := dm.SetResourceProfile(context.Background(), rsPrf, true)
	if err != nil {
		t.Error(err)
	}

	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			ID: "EventAuthorizeResource",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
		UsageID:  "RU_Test",
		UsageTTL: utils.DurationPointer(time.Minute),
		Units:    5,
	}
	var reply string
	if err := rS.V1AllocateResources(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != "Approved" {
		t.Errorf("Unexpected reply returned: %q", reply)
	}

	args = utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			ID: "EventAuthorizeResource",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
		UsageID:  "RU_Test2",
		UsageTTL: utils.DurationPointer(time.Minute),
		Units:    5,
	}

	experr := `cannot find usage record with id: RU_Test2`
	if err := rS.V1ReleaseResources(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestResourcesV1ReleaseResourcesNoMatch(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)

	rsPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weight:            10,
		Limit:             10,
		UsageTTL:          time.Minute,
	}

	err := dm.SetResourceProfile(context.Background(), rsPrf, true)
	if err != nil {
		t.Error(err)
	}

	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			ID: "EventAuthorizeResource",
			Event: map[string]interface{}{
				utils.AccountField: "1002",
			},
		},
		UsageID:  "RU_Test",
		UsageTTL: utils.DurationPointer(time.Minute),
		Units:    5,
	}
	var reply string

	if err := rS.V1ReleaseResources(context.Background(), args, &reply); err == nil ||
		err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestResourcesV1ReleaseResourcesMissingParameters(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)

	rsPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weight:            10,
		Limit:             10,
		UsageTTL:          time.Minute,
	}

	err := dm.SetResourceProfile(context.Background(), rsPrf, true)
	if err != nil {
		t.Error(err)
	}

	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			ID: "EventAuthorizeResource",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
		UsageTTL: utils.DurationPointer(time.Minute),
		Units:    5,
	}
	var reply string

	experr := `MANDATORY_IE_MISSING: [UsageID]`
	if err := rS.V1ReleaseResources(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	args = utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			ID: "EventAuthorizeResource",
		},
		UsageID:  "RU_Test",
		UsageTTL: utils.DurationPointer(time.Minute),
		Units:    5,
	}

	experr = `MANDATORY_IE_MISSING: [Event]`
	if err := rS.V1ReleaseResources(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	args = utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
		UsageID:  "RU_Test",
		UsageTTL: utils.DurationPointer(time.Minute),
		Units:    5,
	}

	experr = `MANDATORY_IE_MISSING: [ID]`
	if err := rS.V1ReleaseResources(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	args = utils.ArgRSv1ResourceUsage{
		UsageID:  "RU_Test",
		UsageTTL: utils.DurationPointer(time.Minute),
		Units:    5,
	}

	experr = `MANDATORY_IE_MISSING: [Event]`
	if err := rS.V1ReleaseResources(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestResourcesV1ReleaseResourcesCacheReplyExists(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	config.SetCgrConfig(cfg)
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	cacheKey := utils.ConcatenatedKey(utils.ResourceSv1ReleaseResources,
		utils.ConcatenatedKey("cgrates.org", "EventReleaseResource"))

	rsPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weight:            10,
		Limit:             -1,
		UsageTTL:          time.Minute,
	}
	rs := &Resource{
		rPrf:   rsPrf,
		Tenant: "cgrates.org",
		ID:     "RES1",
		Usages: map[string]*ResourceUsage{
			"RU1": {
				Tenant: "cgrates.org",
				ID:     "RU1",
				Units:  10,
			},
		},
		dirty:  utils.BoolPointer(false),
		tUsage: utils.Float64Pointer(10),
		ttl:    utils.DurationPointer(time.Minute),
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

	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "EventReleaseResource",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
		UsageID:  "RU_Test",
		UsageTTL: utils.DurationPointer(time.Minute),
		Units:    5,
	}
	cacheReply := "cacheReply"
	Cache.Set(context.Background(), utils.CacheRPCResponses, cacheKey,
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
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	config.SetCgrConfig(cfg)
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	cacheKey := utils.ConcatenatedKey(utils.ResourceSv1ReleaseResources,
		utils.ConcatenatedKey("cgrates.org", "EventReleaseResource"))

	rsPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.MetaNone},
		AllocationMessage: "Approved",
		Weight:            10,
		Limit:             -1,
		UsageTTL:          time.Minute,
	}
	rs := &Resource{
		rPrf:   rsPrf,
		Tenant: "cgrates.org",
		ID:     "RES1",
		Usages: map[string]*ResourceUsage{
			"RU1": {
				Tenant: "cgrates.org",
				ID:     "RU1",
				Units:  4,
			},
		},
		dirty:  utils.BoolPointer(false),
		tUsage: utils.Float64Pointer(10),
		ttl:    utils.DurationPointer(time.Minute),
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

	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, nil)

	args := utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			ID: "EventReleaseResource",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
		UsageID:  "RU_Test",
		UsageTTL: utils.DurationPointer(time.Minute),
		Units:    2,
	}

	var reply string
	experr := `cannot find usage record with id: RU_Test`
	if err := rS.V1ReleaseResources(context.Background(), args, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if itm, has := Cache.Get(utils.CacheRPCResponses, cacheKey); has {
		resp := itm.(*utils.CachedRPCResponse)
		if *resp.Result.(*string) != "" {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				"", *resp.Result.(*string))
		}
	}

	config.SetCgrConfig(config.NewDefaultCGRConfig())
}

func TestResourcesV1ReleaseResourcesProcessThErr(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.ResourceSCfg().StoreInterval = 2
	cfg.ResourceSCfg().ThresholdSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.ThresholdSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				return utils.ErrExists
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	cM := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds): rpcInternal,
	})

	rsPrf := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES1",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		ThresholdIDs:      []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)},
		AllocationMessage: "Approved",
		Weight:            10,
		Limit:             -1,
		UsageTTL:          time.Minute,
	}
	rs := &Resource{
		rPrf:   rsPrf,
		Tenant: "cgrates.org",
		ID:     "RES1",
		Usages: map[string]*ResourceUsage{
			"RU_Test": {
				Tenant: "cgrates.org",
				ID:     "RU_Test",
				Units:  4,
			},
		},
		dirty:  utils.BoolPointer(false),
		tUsage: utils.Float64Pointer(10),
		ttl:    utils.DurationPointer(time.Minute),
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

	fltrs := NewFilterS(cfg, nil, dm)
	rS := NewResourceService(dm, cfg, fltrs, cM)

	args := utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			ID: "EventAuthorizeResource",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
		UsageID:  "RU_Test",
		UsageTTL: utils.DurationPointer(time.Minute),
		Units:    5,
	}
	var reply string
	var resources Resources
	resources = append(resources, rs)
	if _, err := resources.allocateResource(context.Background(), &ResourceUsage{
		Tenant: "cgrates.org",
		ID:     "RU_ID",
		Units:  1}, true); err != nil {
		t.Error(err)
	}

	if err := rS.V1ReleaseResources(context.Background(), args, &reply); err == nil ||
		err != utils.ErrExists {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrExists, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}
