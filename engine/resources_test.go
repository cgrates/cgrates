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

	"github.com/cgrates/cgrates/cache"
	"github.com/cgrates/cgrates/utils"
)

var (
	r1, r2        *Resource
	ru1, ru2, ru3 *ResourceUsage
	rs            Resources
)

func TestRSRecordUsage(t *testing.T) {
	ru1 = &ResourceUsage{
		ID:         "RU1",
		ExpiryTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
		Units:      2,
	}

	ru2 = &ResourceUsage{
		ID:         "RU2",
		ExpiryTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
		Units:      2,
	}

	r1 = &Resource{
		ID: "RL1",
		rPrf: &ResourceProfile{
			ID: "RL1",
			Filters: []*RequestFilter{
				&RequestFilter{
					Type:      MetaString,
					FieldName: "Account",
					Values:    []string{"1001", "1002"},
				},
				&RequestFilter{
					Type:      MetaRSRFields,
					Values:    []string{"Subject(~^1.*1$)", "Destination(1002)"},
					rsrFields: utils.ParseRSRFieldsMustCompile("Subject(~^1.*1$);Destination(1002)", utils.INFIELD_SEP),
				},
			},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC).Add(time.Duration(1 * time.Millisecond)),
			},
			Weight:     100,
			Limit:      2,
			Thresholds: []string{"TEST_ACTIONS"},

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
			t.Error("Duplicate ResourceUsage id should not be allowed")
		}
		if _, found := r1.Usages[ru2.ID]; !found {
			t.Error("ResourceUsage was not recorded")
		}
		if *r1.tUsage != 4 {
			t.Errorf("Expecting: %+v, received: %+v", 4, r1.tUsage)
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
	if *r1.tUsage != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, r1.tUsage)
	}
}

func TestRSUsedUnits(t *testing.T) {
	r1.Usages = map[string]*ResourceUsage{
		ru1.ID: ru1,
	}
	r1.tUsage = nil
	if usedUnits := r1.totalUsage(); usedUnits != 2 {
		t.Errorf("Expecting: %+v, received: %+v", 2, usedUnits)
	}
}

func TestRSRsort(t *testing.T) {
	r2 = &Resource{
		ID: "RL2",
		rPrf: &ResourceProfile{
			ID: "RL2",
			Filters: []*RequestFilter{
				&RequestFilter{
					Type:      MetaString,
					FieldName: "Account",
					Values:    []string{"1001", "1002"},
				},
				&RequestFilter{
					Type:      MetaRSRFields,
					Values:    []string{"Subject(~^1.*1$)", "Destination(1002)"},
					rsrFields: utils.ParseRSRFieldsMustCompile("Subject(~^1.*1$);Destination(1002)", utils.INFIELD_SEP),
				},
			},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
			},

			Weight:     50,
			Limit:      2,
			Thresholds: []string{"TEST_ACTIONS"},
			UsageTTL:   time.Duration(1 * time.Millisecond),
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

	rs[0].rPrf.UsageTTL = time.Duration(20 * time.Second)
	rs[1].rPrf.UsageTTL = time.Duration(20 * time.Second)
	//ru1.ExpiryTime = time.Now()
	//ru2.Time = time.Now()

	if alcMessage, err := rs.AllocateResource(ru1, false); err != nil {
		t.Error(err.Error())
	} else {
		if alcMessage != "ALLOC" {
			t.Errorf("Wrong allocation message: %v", alcMessage)
		}
	}

	if _, err := rs.AllocateResource(ru2, false); err != utils.ErrResourceUnavailable {
		t.Error("Did not receive " + utils.ErrResourceUnavailable.Error() + " error")
	}

	rs[0].rPrf.Limit = 2
	rs[1].rPrf.Limit = 4

	if alcMessage, err := rs.AllocateResource(ru1, true); err != nil {
		t.Error(err.Error())
	} else {
		if alcMessage != "RL2" {
			t.Errorf("Wrong allocation message: %v", alcMessage)
		}
	}

	if alcMessage, err := rs.AllocateResource(ru2, false); err != nil {
		t.Error(err.Error())
	} else {
		if alcMessage != "RL2" {
			t.Errorf("Wrong allocation message: %v", alcMessage)
		}
	}

	ru2.Units = 0
	if _, err := rs.AllocateResource(ru2, false); err == nil {
		t.Error("Duplicate ResourceUsage id should not be allowed")
	}
}

// TestRSCacheSetGet assurace the presence of private params in cached resource
func TestRSCacheSetGet(t *testing.T) {
	r := &Resource{
		ID: "RL",
		rPrf: &ResourceProfile{
			ID: "RL",
			Filters: []*RequestFilter{
				&RequestFilter{
					Type:      MetaString,
					FieldName: "Account",
					Values:    []string{"1001", "1002"},
				},
				&RequestFilter{
					Type:      MetaRSRFields,
					Values:    []string{"Subject(~^1.*1$)", "Destination(1002)"},
					rsrFields: utils.ParseRSRFieldsMustCompile("Subject(~^1.*1$);Destination(1002)", utils.INFIELD_SEP),
				},
			},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
			},
			AllocationMessage: "ALLOC_RL",
			Weight:            50,
			Limit:             2,
			Thresholds:        []string{"TEST_ACTIONS"},
			UsageTTL:          time.Duration(1 * time.Millisecond),
		},
		Usages: map[string]*ResourceUsage{
			"RU2": &ResourceUsage{
				ID:         "RU2",
				ExpiryTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
				Units:      2,
			},
		},
		tUsage: utils.Float64Pointer(2),
		dirty:  utils.BoolPointer(true),
	}
	key := utils.ResourcesPrefix + r.ID
	cache.Set(key, r, true, "")
	if x, ok := cache.Get(key); !ok {
		t.Error("not in cache")
	} else if x == nil {
		t.Error("nil resource")
	} else if !reflect.DeepEqual(r, x.(*Resource)) {
		t.Errorf("Expecting: +v, received: %+v", r, x)
	}
}
