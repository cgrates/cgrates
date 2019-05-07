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
package utils

import (
	"reflect"
	"testing"
	"time"
)

func TestNewDTCSFromRPKey(t *testing.T) {
	rpKey := "*out:tenant12:call:dan12"
	if dtcs, err := NewDTCSFromRPKey(rpKey); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dtcs, &DirectionTenantCategorySubject{"*out", "tenant12", "call", "dan12"}) {
		t.Error("Received: ", dtcs)
	}
}

func TestPaginatorPaginateStringSlice(t *testing.T) {
	eOut := []string{"1", "2", "3", "4"}
	pgnt := new(Paginator)
	if rcv := pgnt.PaginateStringSlice([]string{"1", "2", "3", "4"}); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
	eOut = []string{"1", "2", "3"}
	pgnt.Limit = IntPointer(3)
	if rcv := pgnt.PaginateStringSlice([]string{"1", "2", "3", "4"}); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
	eOut = []string{"2", "3", "4"}
	pgnt.Offset = IntPointer(1)
	if rcv := pgnt.PaginateStringSlice([]string{"1", "2", "3", "4"}); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
	eOut = []string{}
	pgnt.Offset = IntPointer(4)
	if rcv := pgnt.PaginateStringSlice([]string{"1", "2", "3", "4"}); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
	eOut = []string{"3"}
	pgnt.Offset = IntPointer(2)
	pgnt.Limit = IntPointer(1)
	if rcv := pgnt.PaginateStringSlice([]string{"1", "2", "3", "4"}); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
}

func TestAppendToSMCostFilter(t *testing.T) {
	var err error
	smfltr := new(SMCostFilter)
	expected := &SMCostFilter{
		CGRIDs: []string{"CGRID1", "CGRID2"},
	}
	if smfltr, err = AppendToSMCostFilter(smfltr, "*string", DynamicDataPrefix+CGRID, []string{"CGRID1", "CGRID2"}, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}
	expected.NotCGRIDs = []string{"CGRID3", "CGRID4"}
	if smfltr, err = AppendToSMCostFilter(smfltr, "*notstring", DynamicDataPrefix+CGRID, []string{"CGRID3", "CGRID4"}, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}

	expected.RunIDs = []string{"RunID1", "RunID2"}
	if smfltr, err = AppendToSMCostFilter(smfltr, "*string", DynamicDataPrefix+RunID, []string{"RunID1", "RunID2"}, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}
	expected.NotRunIDs = []string{"RunID3", "RunID4"}
	if smfltr, err = AppendToSMCostFilter(smfltr, "*notstring", DynamicDataPrefix+RunID, []string{"RunID3", "RunID4"}, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}

	expected.OriginHosts = []string{"OriginHost1", "OriginHost2"}
	if smfltr, err = AppendToSMCostFilter(smfltr, "*string", DynamicDataPrefix+OriginHost, []string{"OriginHost1", "OriginHost2"}, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}
	expected.NotOriginHosts = []string{"OriginHost3", "OriginHost4"}
	if smfltr, err = AppendToSMCostFilter(smfltr, "*notstring", DynamicDataPrefix+OriginHost, []string{"OriginHost3", "OriginHost4"}, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}

	expected.OriginIDs = []string{"OriginID1", "OriginID2"}
	if smfltr, err = AppendToSMCostFilter(smfltr, "*string", DynamicDataPrefix+OriginID, []string{"OriginID1", "OriginID2"}, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}
	expected.NotOriginIDs = []string{"OriginID3", "OriginID4"}
	if smfltr, err = AppendToSMCostFilter(smfltr, "*notstring", DynamicDataPrefix+OriginID, []string{"OriginID3", "OriginID4"}, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}

	expected.CostSources = []string{"CostSource1", "CostSource2"}
	if smfltr, err = AppendToSMCostFilter(smfltr, "*string", DynamicDataPrefix+CostSource, []string{"CostSource1", "CostSource2"}, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}
	expected.NotCostSources = []string{"CostSource3", "CostSource4"}
	if smfltr, err = AppendToSMCostFilter(smfltr, "*notstring", DynamicDataPrefix+CostSource, []string{"CostSource3", "CostSource4"}, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}

	if smfltr, err = AppendToSMCostFilter(smfltr, "*prefix", DynamicDataPrefix+CGRID, []string{"CGRID1", "CGRID2"}, ""); err == nil || err.Error() != "FilterType: \"*prefix\" not supported for FieldName: \"~CGRID\"" {
		t.Errorf("Expected error: FilterType: \"*prefix\" not supported for FieldName: \"~CGRID\" ,received %v", err)
	}
	if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}
	if smfltr, err = AppendToSMCostFilter(smfltr, "*string", CGRID, []string{"CGRID1", "CGRID2"}, ""); err == nil || err.Error() != "FieldName: \"CGRID\" not supported" {
		t.Errorf("Expected error: FieldName: \"CGRID\" not supported ,received %v", err)
	}
	if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}
	if smfltr, err = AppendToSMCostFilter(smfltr, "*prefix", CGRID, []string{"CGRID1", "CGRID2"}, ""); err == nil || err.Error() != "FieldName: \"CGRID\" not supported" {
		t.Errorf("Expected error: FieldName: \"CGRID\" not supported ,received %v", err)
	}
	if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))

	}
	expected.Usage.Min = DurationPointer(time.Second)
	if smfltr, err = AppendToSMCostFilter(smfltr, "*gte", DynamicDataPrefix+Usage, []string{"1s", "2s"}, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}
	expected.Usage.Max = DurationPointer(3 * time.Second)
	if smfltr, err = AppendToSMCostFilter(smfltr, "*lt", DynamicDataPrefix+Usage, []string{"3s", "4s"}, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}

	if smfltr, err = AppendToSMCostFilter(smfltr, "*gte", DynamicDataPrefix+Usage, []string{"one second"}, ""); err == nil || err.Error() != "Error when converting field: \"*gte\"  value: \"~Usage\" in time.Duration " {
		t.Errorf("Expected error: Error when converting field: \"*gte\"  value: \"~Usage\" in time.Duration ,received %v", err)
	}
	if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}

	now := time.Now().UTC().Round(time.Second)
	strNow := now.Format("2006-01-02T15:04:05")

	expected.CreatedAt.Begin = &now
	if smfltr, err = AppendToSMCostFilter(smfltr, "*gte", DynamicDataPrefix+CreatedAt, []string{strNow}, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}

	expected.CreatedAt.End = &now
	if smfltr, err = AppendToSMCostFilter(smfltr, "*lt", DynamicDataPrefix+CreatedAt, []string{strNow}, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}

	if smfltr, err = AppendToSMCostFilter(smfltr, "*gte", DynamicDataPrefix+CreatedAt, []string{time.Now().String()}, ""); err == nil || err.Error() != "Error when converting field: \"*gte\"  value: \"~CreatedAt\" in time.Time " {
		t.Errorf("Expected error: Error when converting field: \"*gte\"  value: \"~CreatedAt\" in time.Time ,received %v", err)
	}
	if !reflect.DeepEqual(smfltr, expected) {
		t.Errorf("Expected: %s ,received: %s ", ToJSON(expected), ToJSON(smfltr))
	}
}

func TestInitAttrReloadCache(t *testing.T) {
	var expected AttrReloadCache
	expected.DestinationIDs = &[]string{}
	expected.ReverseDestinationIDs = &[]string{}
	expected.RatingPlanIDs = &[]string{}
	expected.RatingProfileIDs = &[]string{}
	expected.ActionIDs = &[]string{}
	expected.ActionPlanIDs = &[]string{}
	expected.AccountActionPlanIDs = &[]string{}
	expected.ActionTriggerIDs = &[]string{}
	expected.SharedGroupIDs = &[]string{}
	expected.ResourceProfileIDs = &[]string{}
	expected.ResourceIDs = &[]string{}
	expected.StatsQueueIDs = &[]string{}
	expected.StatsQueueProfileIDs = &[]string{}
	expected.ThresholdIDs = &[]string{}
	expected.ThresholdProfileIDs = &[]string{}
	expected.FilterIDs = &[]string{}
	expected.SupplierProfileIDs = &[]string{}
	expected.AttributeProfileIDs = &[]string{}
	expected.ChargerProfileIDs = &[]string{}
	expected.DispatcherProfileIDs = &[]string{}
	expected.DispatcherHostIDs = &[]string{}
	expected.DispatcherRoutesIDs = &[]string{}

	if rcv := InitAttrReloadCache(); !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
}
