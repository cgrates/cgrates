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
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

var (
	r, r2        *Resource
	ru, ru2, ru3 *ResourceUsage
	rs           Resources
)

func TestResourceLimitRecordUsage(t *testing.T) {
	ru = &ResourceUsage{
		ID:         "RU1",
		ExpiryTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
		Units:      2,
	}

	ru2 = &ResourceUsage{
		ID:         "RU2",
		ExpiryTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
		Units:      2,
	}

	r = &Resource{
		rCfg: &ResourceCfg{
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
		usages: map[string]*ResourceUsage{
			ru.ID: ru,
		},
		ttlUsages: []*ResourceUsage{ru},
		tUsage:    utils.Float64Pointer(2),
	}

	if err := r.recordUsage(ru2); err != nil {
		t.Error(err.Error())
	} else {
		if err := r.recordUsage(ru); err == nil {
			t.Error("Duplicate ResourceUsage id should not be allowed")
		}
		if _, found := r.usages[ru2.ID]; !found {
			t.Error("ResourceUsage was not recorded")
		}
		if *r.tUsage != 4 {
			t.Errorf("Expecting: %+v, received: %+v", 4, r.tUsage)
		}
	}

}

func TestRLClearUsage(t *testing.T) {
	r.usages = map[string]*ResourceUsage{
		ru.ID: ru,
	}
	*r.tUsage = 3
	r.clearUsage(ru.ID)
	if len(r.usages) != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, len(r.usages))
	}
	if *r.tUsage != 1 {
		t.Errorf("Expecting: %+v, received: %+v", 1, r.tUsage)
	}
}

func TestRLRemoveExpiredUnits(t *testing.T) {
	r.usages = map[string]*ResourceUsage{
		ru.ID: ru,
	}
	*r.tUsage = 2

	r.removeExpiredUnits()

	if len(r.usages) != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, len(r.usages))
	}
	if len(r.ttlUsages) != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, len(r.ttlUsages))
	}
	if *r.tUsage != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, r.tUsage)
	}
}

func TestRLUsedUnits(t *testing.T) {
	r.usages = map[string]*ResourceUsage{
		ru.ID: ru,
	}
	*r.tUsage = 2
	if usedUnits := r.totalUsage(); usedUnits != 2 {
		t.Errorf("Expecting: %+v, received: %+v", 2, usedUnits)
	}
}

func TestRsort(t *testing.T) {
	r2 = &Resource{
		rCfg: &ResourceCfg{
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
		usages: map[string]*ResourceUsage{
			ru2.ID: ru2,
		},
		tUsage: utils.Float64Pointer(2),
	}

	rs = Resources{r2, r}
	rs.Sort()

	if rs[0].rCfg.ID != "RL1" {
		t.Error("Sort failed")
	}
}

func TestRsClearUsage(t *testing.T) {
	if err := r2.clearUsage(ru2.ID); err != nil {
		t.Error(err)
	} else if len(r2.usages) != 0 {
		t.Errorf("Unexpected usages %+v", r2.usages)
	} else if *r2.tUsage != 0 {
		t.Errorf("Unexpected tUsage %+v", r2.tUsage)
	}
}

func TestRsRecordUsages(t *testing.T) {
	if err := rs.recordUsage(ru); err == nil {
		t.Error("should get duplicated error")
	}
}

func TestRsAllocateResource(t *testing.T) {
	rs.clearUsage(ru.ID)
	rs.clearUsage(ru2.ID)

	rs[0].rCfg.UsageTTL = time.Duration(20 * time.Second)
	rs[1].rCfg.UsageTTL = time.Duration(20 * time.Second)
	//ru.ExpiryTime = time.Now()
	//ru2.Time = time.Now()

	if alcMessage, err := rs.AllocateResource(ru, false); err != nil {
		t.Error(err.Error())
	} else {
		if alcMessage != "ALLOC" {
			t.Errorf("Wrong allocation message: %v", alcMessage)
		}
	}

	if _, err := rs.AllocateResource(ru2, false); err != utils.ErrResourceUnavailable {
		t.Error("Did not receive " + utils.ErrResourceUnavailable.Error() + " error")
	}

	rs[0].rCfg.Limit = 2
	rs[1].rCfg.Limit = 4

	if alcMessage, err := rs.AllocateResource(ru, true); err != nil {
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
