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

package sessions

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var attrs = &engine.AttrSProcessEventReply{
	MatchedProfiles: []string{"ATTR_ACNT_1001"},
	AlteredFields:   []string{"OfficeGroup"},
	CGREvent: &utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "TestSSv1ItAuth",
		Context: utils.StringPointer(utils.MetaSessionS),
		Event: map[string]interface{}{
			utils.CGRID:       "5668666d6b8e44eb949042f25ce0796ec3592ff9",
			utils.Tenant:      "cgrates.org",
			utils.Category:    "call",
			utils.ToR:         utils.VOICE,
			utils.Account:     "1001",
			utils.Subject:     "ANY2CNT",
			utils.Destination: "1002",
			"OfficeGroup":     "Marketing",
			utils.OriginID:    "TestSSv1It1",
			utils.RequestType: utils.META_PREPAID,
			utils.SetupTime:   "2018-01-07T17:00:00Z",
			utils.Usage:       300000000000.0,
		},
	},
}

func TestSessionsNewV1AuthorizeArgs(t *testing.T) {
	cgrEv := utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "Event",
		Context: utils.StringPointer(utils.MetaSessionS),
		Event: map[string]interface{}{
			utils.Account:     "1001",
			utils.Destination: "1002",
		},
	}
	expected := &V1AuthorizeArgs{
		AuthorizeResources: true,
		GetAttributes:      true,
		CGREvent:           cgrEv,
	}
	rply := NewV1AuthorizeArgs(true, true, false, false, false, false, false, false, cgrEv)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = &V1AuthorizeArgs{
		GetAttributes:         true,
		AuthorizeResources:    false,
		GetMaxUsage:           true,
		ProcessThresholds:     false,
		ProcessStats:          true,
		GetSuppliers:          false,
		SuppliersIgnoreErrors: true,
		SuppliersMaxCost:      utils.MetaSuppliersEventCost,
		CGREvent:              cgrEv,
	}
	rply = NewV1AuthorizeArgs(true, false, true, false, true, false, true, true, cgrEv)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v,\n received: %+v", expected, rply)
	}
}

func TestSessionsNewV1UpdateSessionArgs(t *testing.T) {
	cgrEv := utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "Event",
		Context: utils.StringPointer(utils.MetaSessionS),
		Event: map[string]interface{}{
			utils.Account:     "1001",
			utils.Destination: "1002",
		},
	}
	expected := &V1UpdateSessionArgs{
		GetAttributes: true,
		UpdateSession: true,
		CGREvent:      cgrEv,
	}
	rply := NewV1UpdateSessionArgs(true, true, cgrEv)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = &V1UpdateSessionArgs{
		GetAttributes: false,
		UpdateSession: true,
		CGREvent:      cgrEv,
	}
	rply = NewV1UpdateSessionArgs(false, true, cgrEv)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestSessionsNewV1TerminateSessionArgs(t *testing.T) {
	cgrEv := utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "Event",
		Context: utils.StringPointer(utils.MetaSessionS),
		Event: map[string]interface{}{
			utils.Account:     "1001",
			utils.Destination: "1002",
		},
	}
	expected := &V1TerminateSessionArgs{
		TerminateSession:  true,
		ProcessThresholds: true,
		CGREvent:          cgrEv,
	}
	rply := NewV1TerminateSessionArgs(true, false, true, false, cgrEv)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = &V1TerminateSessionArgs{
		CGREvent: cgrEv,
	}
	rply = NewV1TerminateSessionArgs(false, false, false, false, cgrEv)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestSessionsNewV1ProcessEventArgs(t *testing.T) {
	cgrEv := utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "Event",
		Context: utils.StringPointer(utils.MetaSessionS),
		Event: map[string]interface{}{
			utils.Account:     "1001",
			utils.Destination: "1002",
		},
	}
	expected := &V1ProcessEventArgs{
		AllocateResources: true,
		Debit:             true,
		GetAttributes:     true,
		CGREvent:          cgrEv,
	}
	rply := NewV1ProcessEventArgs(true, true, true, false, false, cgrEv)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = &V1ProcessEventArgs{
		AllocateResources: true,
		GetAttributes:     true,
		CGREvent:          cgrEv,
	}
	rply = NewV1ProcessEventArgs(true, false, true, false, false, cgrEv)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestSessionsNewV1InitSessionArgs(t *testing.T) {
	cgrEv := utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "Event",
		Context: utils.StringPointer(utils.MetaSessionS),
		Event: map[string]interface{}{
			utils.Account:     "1001",
			utils.Destination: "1002",
		},
	}
	expected := &V1InitSessionArgs{
		GetAttributes:     true,
		AllocateResources: true,
		InitSession:       true,
		ProcessThresholds: true,
		ProcessStats:      true,
		CGREvent:          cgrEv,
	}
	rply := NewV1InitSessionArgs(true, true, true, true, true, cgrEv)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = &V1InitSessionArgs{
		GetAttributes:     true,
		AllocateResources: false,
		InitSession:       true,
		ProcessThresholds: false,
		ProcessStats:      true,
		CGREvent:          cgrEv,
	}
	rply = NewV1InitSessionArgs(true, false, true, false, true, cgrEv)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestV1AuthorizeReplyAsNavigableMap(t *testing.T) {
	splrs := &engine.SortedSuppliers{
		ProfileID: "SPL_ACNT_1001",
		Sorting:   utils.MetaWeight,
		SortedSuppliers: []*engine.SortedSupplier{
			{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					"Weight": 20.0,
				},
			},
			{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					"Weight": 10.0,
				},
			},
		},
	}
	thIDs := &[]string{"THD_RES_1", "THD_STATS_1", "THD_STATS_2", "THD_CDRS_1"}
	statIDs := &[]string{"Stats2", "Stats1", "Stats3"}
	v1AuthRpl := new(V1AuthorizeReply)
	expected := config.NewNavigableMap(map[string]interface{}{})
	if rply, _ := v1AuthRpl.AsNavigableMap(nil); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1AuthRpl.Attributes = attrs
	expected.Set([]string{utils.CapAttributes},
		map[string]interface{}{"OfficeGroup": "Marketing"},
		false, false)
	if rply, _ := v1AuthRpl.AsNavigableMap(nil); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1AuthRpl.MaxUsage = utils.DurationPointer(5 * time.Minute)
	expected.Set([]string{utils.CapMaxUsage},
		5*time.Minute, false, false)
	if rply, _ := v1AuthRpl.AsNavigableMap(nil); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1AuthRpl = &V1AuthorizeReply{
		Attributes:         attrs,
		ResourceAllocation: utils.StringPointer("ResGr1"),
		MaxUsage:           utils.DurationPointer(5 * time.Minute),
		Suppliers:          splrs,
		ThresholdIDs:       thIDs,
		StatQueueIDs:       statIDs,
	}
	expected = config.NewNavigableMap(map[string]interface{}{
		utils.CapAttributes:         map[string]interface{}{"OfficeGroup": "Marketing"},
		utils.CapResourceAllocation: "ResGr1",
		utils.CapMaxUsage:           5 * time.Minute,
		utils.CapSuppliers:          splrs.Digest(),
		utils.CapThresholds:         *thIDs,
		utils.CapStatQueues:         *statIDs,
	})
	if rply, _ := v1AuthRpl.AsNavigableMap(nil); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
}

func TestV1InitSessionReplyAsNavigableMap(t *testing.T) {
	thIDs := &[]string{"THD_RES_1", "THD_STATS_1", "THD_STATS_2", "THD_CDRS_1"}
	statIDs := &[]string{"Stats2", "Stats1", "Stats3"}
	v1InitRpl := new(V1InitSessionReply)
	expected := config.NewNavigableMap(map[string]interface{}{})
	if rply, _ := v1InitRpl.AsNavigableMap(nil); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1InitRpl.Attributes = attrs
	expected.Set([]string{utils.CapAttributes},
		map[string]interface{}{"OfficeGroup": "Marketing"},
		false, false)
	if rply, _ := v1InitRpl.AsNavigableMap(nil); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1InitRpl.MaxUsage = utils.DurationPointer(5 * time.Minute)
	expected.Set([]string{utils.CapMaxUsage},
		5*time.Minute, false, false)
	if rply, _ := v1InitRpl.AsNavigableMap(nil); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1InitRpl = &V1InitSessionReply{
		Attributes:         attrs,
		ResourceAllocation: utils.StringPointer("ResGr1"),
		MaxUsage:           utils.DurationPointer(5 * time.Minute),
		ThresholdIDs:       thIDs,
		StatQueueIDs:       statIDs,
	}
	expected = config.NewNavigableMap(map[string]interface{}{
		utils.CapAttributes:         map[string]interface{}{"OfficeGroup": "Marketing"},
		utils.CapResourceAllocation: "ResGr1",
		utils.CapMaxUsage:           5 * time.Minute,
		utils.CapThresholds:         *thIDs,
		utils.CapStatQueues:         *statIDs,
	})
	if rply, _ := v1InitRpl.AsNavigableMap(nil); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
}

func TestV1UpdateSessionReplyAsNavigableMap(t *testing.T) {
	v1UpdtRpl := new(V1UpdateSessionReply)
	expected := config.NewNavigableMap(map[string]interface{}{})
	if rply, _ := v1UpdtRpl.AsNavigableMap(nil); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1UpdtRpl.Attributes = attrs
	expected.Set([]string{utils.CapAttributes},
		map[string]interface{}{"OfficeGroup": "Marketing"},
		false, false)
	if rply, _ := v1UpdtRpl.AsNavigableMap(nil); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1UpdtRpl.MaxUsage = utils.DurationPointer(5 * time.Minute)
	expected.Set([]string{utils.CapMaxUsage},
		5*time.Minute, false, false)
	if rply, _ := v1UpdtRpl.AsNavigableMap(nil); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
}
func TestV1ProcessEventReplyAsNavigableMap(t *testing.T) {
	v1PrcEvRpl := new(V1ProcessEventReply)
	expected := config.NewNavigableMap(map[string]interface{}{})
	if rply, _ := v1PrcEvRpl.AsNavigableMap(nil); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1PrcEvRpl.Attributes = attrs
	expected.Set([]string{utils.CapAttributes},
		map[string]interface{}{"OfficeGroup": "Marketing"},
		false, false)
	if rply, _ := v1PrcEvRpl.AsNavigableMap(nil); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1PrcEvRpl.MaxUsage = utils.DurationPointer(5 * time.Minute)
	expected.Set([]string{utils.CapMaxUsage},
		5*time.Minute, false, false)
	if rply, _ := v1PrcEvRpl.AsNavigableMap(nil); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
	v1PrcEvRpl.ResourceAllocation = utils.StringPointer("ResGr1")
	expected.Set([]string{utils.CapResourceAllocation},
		"ResGr1", false, false)
	if rply, _ := v1PrcEvRpl.AsNavigableMap(nil); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting \n%+v\n, received: \n%+v", expected, rply)
	}
}
