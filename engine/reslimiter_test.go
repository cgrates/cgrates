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

	// "reflect"
	"github.com/cgrates/cgrates/utils"
)

var (
	rl, rl2      *ResourceLimit
	ru, ru2, ru3 *ResourceUsage
	rls          ResourceLimits
)

func TestResourceLimitRecordUsage(t *testing.T) {
	ru = &ResourceUsage{
		ID:    "RU1",
		Time:  time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
		Units: 2,
	}

	ru2 = &ResourceUsage{
		ID:    "RU2",
		Time:  time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
		Units: 2,
	}

	rl = &ResourceLimit{
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
		},
		ExpiryTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
		Weight:     100,
		Limit:      2,
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				// ID            string // original csv tag
				// UniqueID      string // individual id
				ThresholdType:  utils.TRIGGER_MAX_EVENT_COUNTER,
				ThresholdValue: 2,
				// Recurrent      bool          // reset excuted flag each run
				// MinSleep       time.Duration // Minimum duration between two executions in case of recurrent triggers
				// ExpirationDate time.Time
				// ActivationDate time.Time
				Balance: &BalanceFilter{
					Type:       utils.StringPointer(utils.MONETARY),
					Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
				},
				ActionsID: "TEST_ACTIONS",
				// Weight            float64
				// ActionsID         string
				// MinQueuedItems    int // Trigger actions only if this number is hit (stats only)
				// Executed          bool
				// LastExecutionTime time.Time
			},
		},
		UsageTTL:          time.Duration(1 * time.Millisecond),
		AllocationMessage: "ALLOC",
		Usage: map[string]*ResourceUsage{
			ru.ID: ru,
		},
		TotalUsage: 2,
	}

	if err := rl.RecordUsage(ru2); err != nil {
		t.Error(err.Error())
	} else {
		if err := rl.RecordUsage(ru); err == nil {
			t.Error("Duplicate ResourceUsage id should not be allowed")
		}
		if _, found := rl.Usage[ru2.ID]; !found {
			t.Error("ResourceUsage was not recorded")
		}
		if rl.TotalUsage != 4 {
			t.Errorf("Expecting: %+v, received: %+v", 4, rl.TotalUsage)
		}
	}

}

func TestRLClearUsage(t *testing.T) {
	rl.Usage = map[string]*ResourceUsage{
		ru.ID: ru,
	}
	rl.TotalUsage = 3

	rl.ClearUsage(ru.ID)

	if len(rl.Usage) != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, len(rl.Usage))
	}
	if rl.TotalUsage != 1 {
		t.Errorf("Expecting: %+v, received: %+v", 1, rl.TotalUsage)
	}
}

func TestRLRemoveExpiredUnits(t *testing.T) {
	rl.Usage = map[string]*ResourceUsage{
		ru.ID: ru,
	}
	rl.TotalUsage = 2

	rl.removeExpiredUnits()

	if len(rl.Usage) != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, len(rl.Usage))
	}
	if rl.TotalUsage != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, rl.TotalUsage)
	}
}

func TestRLUsedUnits(t *testing.T) {
	rl.Usage = map[string]*ResourceUsage{
		ru.ID: ru,
	}
	rl.TotalUsage = 2

	usedUnits := rl.UsedUnits()

	if len(rl.Usage) != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, len(rl.Usage))
	}
	if usedUnits != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, usedUnits)
	}
}

func TestRLSort(t *testing.T) {
	rl2 = &ResourceLimit{
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
		},
		ExpiryTime: time.Date(2014, 7, 3, 13, 43, 0, 1, time.UTC),
		Weight:     50,
		Limit:      2,
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				// ID            string // original csv tag
				// UniqueID      string // individual id
				ThresholdType:  utils.TRIGGER_MAX_EVENT_COUNTER,
				ThresholdValue: 2,
				// Recurrent      bool          // reset excuted flag each run
				// MinSleep       time.Duration // Minimum duration between two executions in case of recurrent triggers
				// ExpirationDate time.Time
				// ActivationDate time.Time
				Balance: &BalanceFilter{
					Type:       utils.StringPointer(utils.MONETARY),
					Directions: utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
				},
				ActionsID: "TEST_ACTIONS",
				// Weight            float64
				// ActionsID         string
				// MinQueuedItems    int // Trigger actions only if this number is hit (stats only)
				// Executed          bool
				// LastExecutionTime time.Time
			},
		},
		UsageTTL: time.Duration(1 * time.Millisecond),
		// AllocationMessage: "ALLOC2",
		Usage: map[string]*ResourceUsage{
			ru2.ID: ru2,
		},
		TotalUsage: 2,
	}

	rls = ResourceLimits{rl2, rl}
	rls.Sort()

	if rls[0].ID != "RL1" {
		t.Error("Sort failed")
	}
}

func TestRLsClearUsage(t *testing.T) {
	rls.ClearUsage(ru2.ID)
	for _, rl := range rls {
		if len(rl.Usage) > 0 {
			t.Errorf("Expecting: %+v, received: %+v", 0, len(rl.Usage))
		}
	}
}

func TestRLsRecordUsages(t *testing.T) {
	if err := rls.RecordUsage(ru); err != nil {
		for _, rl := range rls {
			if _, found := rl.Usage[ru.ID]; found {
				t.Error("Fallback on error failed")
			}
		}
		t.Error(err.Error())
	} else {
		for _, rl := range rls {
			if _, found := rl.Usage[ru.ID]; !found {
				t.Error("ResourceUsage not found ")
			}
			if rl.TotalUsage != 2 {
				t.Errorf("Expecting: %+v, received: %+v", 2, rl.TotalUsage)
			}
		}
	}
}

func TestRLsAllocateResource(t *testing.T) {
	rls.ClearUsage(ru.ID)
	rls.ClearUsage(ru2.ID)

	rls[0].UsageTTL = time.Duration(20 * time.Second)
	rls[1].UsageTTL = time.Duration(20 * time.Second)
	ru.Time = time.Now()
	ru2.Time = time.Now()

	if alcMessage, err := rls.AllocateResource(ru, false); err != nil {
		t.Error(err.Error())
	} else {
		if alcMessage != "ALLOC" {
			t.Errorf("Wrong allocation message: %v", alcMessage)
		}
	}

	if _, err := rls.AllocateResource(ru2, false); err != utils.ErrResourceUnavailable {
		t.Error("Did not receive " + utils.ErrResourceUnavailable.Error() + " error")
	}

	rls[0].Limit = 2
	rls[1].Limit = 4

	if alcMessage, err := rls.AllocateResource(ru, true); err != nil {
		t.Error(err.Error())
	} else {
		if alcMessage != "RL2" {
			t.Errorf("Wrong allocation message: %v", alcMessage)
		}
	}

	if alcMessage, err := rls.AllocateResource(ru2, false); err != nil {
		t.Error(err.Error())
	} else {
		if alcMessage != "RL2" {
			t.Errorf("Wrong allocation message: %v", alcMessage)
		}
	}

	ru2.Units = 0
	if _, err := rls.AllocateResource(ru2, false); err == nil {
		t.Error("Duplicate ResourceUsage id should not be allowed")
	}
}
