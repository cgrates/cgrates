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
package utils

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestCGREventCheckMandatoryFields(t *testing.T) {
	//empty check
	cgrEvent := new(CGREvent)
	fldNames := []string{}
	err := cgrEvent.CheckMandatoryFields(fldNames)
	if err != nil {
		t.Error(err)
	}
	cgrEvent = &CGREvent{
		Event: map[string]any{
			Usage:   20 * time.Second,
			"test1": 1,
			"test2": 2,
			"test3": 3,
		},
	}
	//normal check
	fldNames = []string{"test1", "test2"}
	err = cgrEvent.CheckMandatoryFields(fldNames)
	if err != nil {
		t.Error(err)
	}
	//MANDATORY_IE_MISSING
	fldNames = []string{"test4", "test5"}
	err = cgrEvent.CheckMandatoryFields(fldNames)
	if err == nil || err.Error() != NewErrMandatoryIeMissing("test4").Error() {
		t.Errorf("Expected %s, received %s", NewErrMandatoryIeMissing("test4"), err)
	}
}

func TestCGREventFielAsString(t *testing.T) {
	//empty check
	cgrEvent := new(CGREvent)
	fldname := "test"
	_, err := cgrEvent.FieldAsString(fldname)
	if err == nil || err.Error() != ErrNotFound.Error() {
		t.Errorf("Expected %s, received %s", ErrNotFound, err)
	}
	//normal check
	cgrEvent = &CGREvent{
		Event: map[string]any{
			Usage:   20 * time.Second,
			"test1": 1,
			"test2": 2,
			"test3": 3,
		},
	}
	fldname = "test1"
	rcv, err := cgrEvent.FieldAsString(fldname)
	if err != nil {
		t.Error(err)
	}
	if rcv != "1" {
		t.Errorf("Expected: 1, received %+q", rcv)
	}

}

func TestLibRoutesUsage(t *testing.T) {
	se := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "supplierEvent1",
		Event: map[string]any{
			Usage: 20 * time.Second,
		},
	}
	seErr := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "supplierEvent1",
		Event:  make(map[string]any),
	}
	answ, err := se.FieldAsDuration(Usage)
	if err != nil {
		t.Error(err)
	}
	if answ != se.Event[Usage] {
		t.Errorf("Expecting: %+v, received: %+v", se.Event[Usage], answ)
	}
	_, err = seErr.FieldAsDuration(Usage)
	if err != ErrNotFound {
		t.Error(err)
	}
}

func TestCGREventFieldAsTime(t *testing.T) {
	se := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "supplierEvent1",
		Event: map[string]any{
			AnswerTime: time.Now(),
		},
	}
	seErr := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "supplierEvent1",
		Event:  make(map[string]any),
	}
	answ, err := se.FieldAsTime(AnswerTime, "UTC")
	if err != nil {
		t.Error(err)
	}
	if answ != se.Event[AnswerTime] {
		t.Errorf("Expecting: %+v, received: %+v", se.Event[AnswerTime], answ)
	}
	_, err = seErr.FieldAsTime(AnswerTime, "CET")
	if err != ErrNotFound {
		t.Error(err)
	}
}

func TestCGREventFieldAsString(t *testing.T) {
	se := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "supplierEvent1",
		Event: map[string]any{
			"supplierprofile1": "Supplier",
			"UsageInterval":    time.Second,
			"PddInterval":      "1s",
			"Weight":           20.0,
		},
	}
	answ, err := se.FieldAsString("UsageInterval")
	if err != nil {
		t.Error(err)
	}
	if answ != "1s" {
		t.Errorf("Expecting: %+v, received: %+v", se.Event["UsageInterval"], answ)
	}
	answ, err = se.FieldAsString("PddInterval")
	if err != nil {
		t.Error(err)
	}
	if answ != se.Event["PddInterval"] {
		t.Errorf("Expecting: %+v, received: %+v", se.Event["PddInterval"], answ)
	}
	answ, err = se.FieldAsString("supplierprofile1")
	if err != nil {
		t.Error(err)
	}
	if answ != se.Event["supplierprofile1"] {
		t.Errorf("Expecting: %+v, received: %+v", se.Event["supplierprofile1"], answ)
	}
	answ, err = se.FieldAsString("Weight")
	if err != nil {
		t.Error(err)
	}
	if answ != "20" {
		t.Errorf("Expecting: %+v, received: %+v", se.Event["Weight"], answ)
	}
}

func TestCGREventClone(t *testing.T) {
	ev := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "supplierEvent1",
		Event: map[string]any{
			AnswerTime:         time.Now(),
			"supplierprofile1": "Supplier",
			"UsageInterval":    "54.2",
			"PddInterval":      "1s",
			"Weight":           20.0,
		},
		APIOpts: map[string]any{
			"testKey": 12,
		},
	}
	cloned := ev.Clone()
	if !reflect.DeepEqual(ev, cloned) {
		t.Errorf("Expecting: %+v, received: %+v", ev, cloned)
	}
}

func TestCGREventOptAsInt64(t *testing.T) {
	ev := &CGREvent{
		APIOpts: map[string]any{
			"testKey": "13",
		},
	}

	received, err := ev.OptAsInt64("testKey")
	if err != nil {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", nil, err)
	}
	expected := int64(13)

	if received != expected {
		t.Errorf("\nExpected: %q, \nReceived: %q", expected, received)
	}
	errExpect := ErrNotFound
	if _, err = ev.OptAsInt64("nonExistingKey"); err == nil || err != errExpect {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", errExpect, err)
	}
}

func TestCGREventAsDataProvider(t *testing.T) {
	ev := &CGREvent{
		APIOpts: map[string]any{
			"testKey1": 13,
			"testKey2": "testString1",
		},
		Event: map[string]any{
			"testKey1": 30,
			"testKey2": "testString2",
		},
	}

	expected := MapStorage{
		MetaOpts: ev.APIOpts,
		MetaReq:  ev.Event,
	}

	received := ev.AsDataProvider()

	if !reflect.DeepEqual(expected, received) {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, received)
	}
}

func TestNMAsCGREvent(t *testing.T) {
	if cgrEv := NMAsCGREvent(nil, "cgrates.org",
		NestingSep, nil); cgrEv != nil {
		t.Errorf("expecting: %+v, \nreceived: %+v", ToJSON(nil), ToJSON(cgrEv.Event))
	}

	nM := NewOrderedNavigableMap()
	if cgrEv := NMAsCGREvent(nM, "cgrates.org",
		NestingSep, nil); cgrEv != nil {
		t.Errorf("expecting: %+v, \nreceived: %+v", ToJSON(nil), ToJSON(cgrEv.Event))
	}

	path := []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"}
	if err := nM.SetAsSlice(&FullPath{Path: strings.Join(path, NestingSep), PathSlice: path}, []*DataNode{
		{Type: NMDataType, Value: &DataLeaf{
			Data: "Val1",
		}}}); err != nil {
		t.Error(err)
	}

	path = []string{"FirstLevel2", "SecondLevel2", "Field2"}
	if err := nM.SetAsSlice(&FullPath{Path: strings.Join(path, NestingSep), PathSlice: path}, []*DataNode{
		{Type: NMDataType, Value: &DataLeaf{
			Data:        "attrVal1",
			AttributeID: "attribute1",
		}},
		{Type: NMDataType, Value: &DataLeaf{
			Data: "Value2",
		}}}); err != nil {
		t.Error(err)
	}

	path = []string{"FirstLevel2", "Field3"}
	if err := nM.SetAsSlice(&FullPath{Path: strings.Join(path, NestingSep), PathSlice: path}, []*DataNode{
		{Type: NMDataType, Value: &DataLeaf{
			Data: "Value3",
		}}}); err != nil {
		t.Error(err)
	}

	path = []string{"FirstLevel2", "Field5"}
	if err := nM.SetAsSlice(&FullPath{Path: strings.Join(path, NestingSep), PathSlice: path}, []*DataNode{
		{Type: NMDataType, Value: &DataLeaf{
			Data: "Value5",
		}},
		{Type: NMDataType, Value: &DataLeaf{
			Data:        "attrVal5",
			AttributeID: "attribute5",
		}}}); err != nil {
		t.Error(err)
	}

	path = []string{"FirstLevel2", "Field6"}
	if err := nM.SetAsSlice(&FullPath{Path: strings.Join(path, NestingSep), PathSlice: path}, []*DataNode{
		{Type: NMDataType, Value: &DataLeaf{
			Data:      "Value6",
			NewBranch: true,
		}},
		{Type: NMDataType, Value: &DataLeaf{
			Data:        "attrVal6",
			AttributeID: "attribute6",
		}}}); err != nil {
		t.Error(err)
	}

	path = []string{"Field4"}
	if err := nM.SetAsSlice(&FullPath{Path: strings.Join(path, NestingSep), PathSlice: path}, []*DataNode{
		{Type: NMDataType, Value: &DataLeaf{
			Data: "Val4",
		}},
		{Type: NMDataType, Value: &DataLeaf{
			Data:        "attrVal2",
			AttributeID: "attribute2",
		}}}); err != nil {
		t.Error(err)
	}
	eEv := map[string]any{
		"FirstLevel2.SecondLevel2.Field2":        "Value2",
		"FirstLevel.SecondLevel.ThirdLevel.Fld1": "Val1",
		"FirstLevel2.Field3":                     "Value3",
		"FirstLevel2.Field5":                     "Value5",
		"FirstLevel2.Field6":                     "Value6",
		"Field4":                                 "Val4",
	}
	if cgrEv := NMAsCGREvent(nM, "cgrates.org",
		NestingSep, MapStorage{}); cgrEv.Tenant != "cgrates.org" ||
		!reflect.DeepEqual(eEv, cgrEv.Event) {
		t.Errorf("expecting: %+v, \nreceived: %+v", ToJSON(eEv), ToJSON(cgrEv.Event))
	}
}

func TestCGREventSetCloneable(t *testing.T) {
	cgrEv := &CGREvent{
		clnb: false,
	}

	cgrEv.SetCloneable(true)
	if !cgrEv.clnb {
		t.Error("Expected clnb to be set to true")
	}
}

func TestCGREventRPCClone(t *testing.T) {
	cgrEv := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			AnswerTime:         nil,
			"supplierprofile1": "Supplier",
			"UsageInterval":    "54.2",
			"PddInterval":      "1s",
			"Weight":           20.0,
		},
		APIOpts: map[string]any{
			"testKey": 12,
		},
		clnb: false, //first make it non clonable
	}

	rcv, err := cgrEv.RPCClone()
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(cgrEv, rcv) {
		t.Errorf("Expected %v \n but received \n %v", cgrEv, rcv)
	}

	cgrEv.clnb = true
	rcv, err = cgrEv.RPCClone()
	if err != nil {
		t.Error(err)
	}
	exp := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]any{
			AnswerTime:         nil,
			"supplierprofile1": "Supplier",
			"UsageInterval":    "54.2",
			"PddInterval":      "1s",
			"Weight":           20.0,
		},
		APIOpts: map[string]any{
			"testKey": 12,
		},
		clnb: false,
	}
	if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("Expected %v \n but received \n %v", exp, rcv.(*CGREvent))
	}
}

func TestHasField(t *testing.T) {
	ev := &CGREvent{
		Event: map[string]any{
			"supplierprofile1": "Supplier",
		},
	}

	exp := true

	if rcv := ev.HasField("supplierprofile1"); !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n received %v", exp, rcv)

	}
}
