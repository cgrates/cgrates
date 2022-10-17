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
	"strings"
	"testing"
	"time"
)

func TestCGREventHasField(t *testing.T) {
	//empty check
	cgrEvent := new(CGREvent)
	rcv := cgrEvent.HasField("")
	if rcv {
		t.Error("Expecting: false, received: ", rcv)
	}
	//normal check
	cgrEvent = &CGREvent{
		Event: map[string]interface{}{
			Usage: 20 * time.Second,
		},
	}
	rcv = cgrEvent.HasField("Usage")
	if !rcv {
		t.Error("Expecting: true, received: ", rcv)
	}
}

func TestCGREventCheckMandatoryFields(t *testing.T) {
	//empty check
	cgrEvent := new(CGREvent)
	fldNames := []string{}
	err := cgrEvent.CheckMandatoryFields(fldNames)
	if err != nil {
		t.Error(err)
	}
	cgrEvent = &CGREvent{
		Event: map[string]interface{}{
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
		Event: map[string]interface{}{
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
		Event: map[string]interface{}{
			Usage: 20 * time.Second,
		},
	}
	seErr := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "supplierEvent1",
		Event:  make(map[string]interface{}),
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
		Event: map[string]interface{}{
			AnswerTime: time.Now(),
		},
	}
	seErr := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "supplierEvent1",
		Event:  make(map[string]interface{}),
	}
	answ, err := se.FieldAsTime(AnswerTime, "UTC")
	if err != nil {
		t.Error(err)
	}
	if answ != se.Event[AnswerTime] {
		t.Errorf("Expecting: %+v, received: %+v", se.Event[AnswerTime], answ)
	}
	answ, err = seErr.FieldAsTime(AnswerTime, "CET")
	if err != ErrNotFound {
		t.Error(err)
	}
}

func TestCGREventFieldAsString(t *testing.T) {
	se := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "supplierEvent1",
		Event: map[string]interface{}{
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

func TestCGREventFieldAsFloat64(t *testing.T) {
	se := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "supplierEvent1",
		Event: map[string]interface{}{
			AnswerTime:         time.Now(),
			"supplierprofile1": "Supplier",
			"UsageInterval":    "54.2",
			"PddInterval":      "1s",
			"Weight":           20.0,
		},
	}
	answ, err := se.FieldAsFloat64("UsageInterval")
	if err != nil {
		t.Error(err)
	}
	if answ != float64(54.2) {
		t.Errorf("Expecting: %+v, received: %+v", se.Event["UsageInterval"], answ)
	}
	answ, err = se.FieldAsFloat64("Weight")
	if err != nil {
		t.Error(err)
	}
	if answ != float64(20.0) {
		t.Errorf("Expecting: %+v, received: %+v", se.Event["Weight"], answ)
	}
	answ, err = se.FieldAsFloat64("PddInterval")
	if err == nil || err.Error() != `strconv.ParseFloat: parsing "1s": invalid syntax` {
		t.Errorf("Expected %s, received %s", `strconv.ParseFloat: parsing "1s": invalid syntax`, err)
	}
	if answ != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, answ)
	}

	if _, err := se.FieldAsFloat64(AnswerTime); err == nil || !strings.HasPrefix(err.Error(), "cannot convert field") {
		t.Errorf("Unexpected error : %+v", err)
	}
	if _, err := se.FieldAsFloat64(AccountField); err == nil || err.Error() != ErrNotFound.Error() {
		t.Errorf("Expected %s, received %s", ErrNotFound, err)
	}
	// }
}

// }
func TestCGREventTenantID(t *testing.T) {
	//empty check
	cgrEvent := new(CGREvent)
	rcv := cgrEvent.TenantID()
	eOut := ":"
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
	//normal check
	cgrEvent = &CGREvent{
		Tenant: "cgrates.org",
		ID:     "supplierEvent1",
	}
	rcv = cgrEvent.TenantID()
	eOut = "cgrates.org:supplierEvent1"
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}

}

func TestCGREventClone(t *testing.T) {
	now := time.Now()
	ev := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "supplierEvent1",
		Time:   &now,
		Event: map[string]interface{}{
			AnswerTime:         time.Now(),
			"supplierprofile1": "Supplier",
			"UsageInterval":    "54.2",
			"PddInterval":      "1s",
			"Weight":           20.0,
		},
		APIOpts: map[string]interface{}{
			"testKey": 12,
		},
	}
	cloned := ev.Clone()
	if !reflect.DeepEqual(ev, cloned) {
		t.Errorf("Expecting: %+v, received: %+v", ev, cloned)
	}
	if cloned.Time == ev.Time {
		t.Errorf("Expecting: different pointer but received: %+v", cloned.Time)
	}
}

func TestCGREventconsumeRoutePaginator(t *testing.T) {
	//empty check
	var opts map[string]interface{}
	rcv, err := GetRoutePaginatorFromOpts(opts)
	if err != nil {
		t.Error(err)
	}
	var eOut Paginator
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting:  %+v, received: %+v", eOut, rcv)
	}
	opts = nil
	rcv, err = GetRoutePaginatorFromOpts(opts)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting:  %+v, received: %+v", eOut, rcv)
	}
	//normal check
	opts = map[string]interface{}{
		OptsRoutesLimit:  18,
		OptsRoutesOffset: 20,
	}

	eOut = Paginator{
		Limit:  IntPointer(18),
		Offset: IntPointer(20),
	}
	rcv, err = GetRoutePaginatorFromOpts(opts)
	if err != nil {
		t.Error(err)
	}
	//check if *rouLimit and *rouOffset was deleted
	if _, has := opts[OptsRoutesLimit]; has {
		t.Errorf("*rouLimit wasn't deleted")
	} else if _, has := opts[OptsRoutesOffset]; has {
		t.Errorf("*rouOffset wasn't deleted")
	}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting:  %+v, received: %+v", eOut, rcv)
	}
	//check without *rouLimit, but with *rouOffset
	opts = map[string]interface{}{
		OptsRoutesOffset: 20,
	}

	eOut = Paginator{
		Offset: IntPointer(20),
	}
	rcv, err = GetRoutePaginatorFromOpts(opts)
	if err != nil {
		t.Error(err)
	}
	//check if *rouLimit and *rouOffset was deleted
	if _, has := opts[OptsRoutesLimit]; has {
		t.Errorf("*rouLimit wasn't deleted")
	} else if _, has := opts[OptsRoutesOffset]; has {
		t.Errorf("*rouOffset wasn't deleted")
	}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting:  %+v, received: %+v", eOut, rcv)
	}
	//check with notAnInt at *rouLimit
	opts = map[string]interface{}{
		OptsRoutesLimit: "Not an int",
	}
	eOut = Paginator{}
	rcv, err = GetRoutePaginatorFromOpts(opts)
	if err == nil {
		t.Error("Expected error")
	}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting:  %+v, received: %+v", eOut, rcv)
	}
	//check with notAnInt at and *rouOffset
	opts = map[string]interface{}{
		OptsRoutesOffset: "Not an int",
	}
	eOut = Paginator{}
	rcv, err = GetRoutePaginatorFromOpts(opts)
	if err == nil {
		t.Error("Expected error")
	}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting:  %+v, received: %+v", eOut, rcv)
	}
}

func TestCGREventconsumeRoutePaginatorCase1(t *testing.T) {
	opts := map[string]interface{}{}
	if _, err := GetRoutePaginatorFromOpts(opts); err != nil {
		t.Error(err)
	}
}

func TestCGREventOptAsInt64(t *testing.T) {
	ev := &CGREvent{
		APIOpts: map[string]interface{}{
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

func TestCGREventFieldAsInt64(t *testing.T) {
	se := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "supplierEvent1",
		Event: map[string]interface{}{
			AnswerTime:         time.Now(),
			"supplierprofile1": "Supplier",
			"UsageInterval":    "54",
			"PddInterval":      "1s",
			"Weight":           20,
		},
	}
	answ, err := se.FieldAsInt64("UsageInterval")
	if err != nil {
		t.Error(err)
	}
	if answ != int64(54) {
		t.Errorf("Expecting: %+v, received: %+v", se.Event["UsageInterval"], answ)
	}
	answ, err = se.FieldAsInt64("Weight")
	if err != nil {
		t.Error(err)
	}
	if answ != int64(20) {
		t.Errorf("Expecting: %+v, received: %+v", se.Event["Weight"], answ)
	}
	answ, err = se.FieldAsInt64("PddInterval")
	if err == nil || err.Error() != `strconv.ParseInt: parsing "1s": invalid syntax` {
		t.Errorf("Expected %s, received %s", `strconv.ParseInt: parsing "1s": invalid syntax`, err)
	}
	if answ != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, answ)
	}

	if _, err := se.FieldAsInt64(AnswerTime); err == nil || !strings.HasPrefix(err.Error(), "cannot convert field") {
		t.Errorf("Unexpected error : %+v", err)
	}
	if _, err := se.FieldAsInt64(AccountField); err == nil || err.Error() != ErrNotFound.Error() {
		t.Errorf("Expected %s, received %s", ErrNotFound, err)
	}
	// }
}

func TestCGREventOptAsStringEmpty(t *testing.T) {
	ev := &CGREvent{}

	expstr := ""
	experr := ErrNotFound
	received, err := ev.OptAsString("testString")

	if !reflect.DeepEqual(received, expstr) {
		t.Errorf("\nExpected: %q, \nReceived: %q", expstr, received)
	}
	if err == nil || err != experr {
		t.Errorf("\nExpected: %q, \nReceived: %q", experr, err)
	}
}

func TestCGREventOptAsString(t *testing.T) {
	ev := &CGREvent{
		APIOpts: map[string]interface{}{
			"testKey": 13,
		},
	}

	received, err := ev.OptAsString("testKey")
	if err != nil {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", nil, err)
	}
	expected := "13"

	if received != expected {
		t.Errorf("\nExpected: %q, \nReceived: %q", expected, received)
	}
}

func TestCGREventOptAsDurationEmpty(t *testing.T) {
	ev := &CGREvent{}

	var expdur time.Duration
	experr := ErrNotFound
	received, err := ev.OptAsDuration("testString")

	if !reflect.DeepEqual(received, expdur) {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expdur, received)
	}
	if err == nil || err != experr {
		t.Errorf("\nExpected: %q, \nReceived: %q", experr, err)
	}
}

func TestCGREventOptAsDuration(t *testing.T) {
	ev := &CGREvent{
		APIOpts: map[string]interface{}{
			"testKey": 30,
		},
	}

	received, err := ev.OptAsDuration("testKey")
	if err != nil {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", nil, err)
	}
	expected := 30 * time.Nanosecond

	if received != expected {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, received)
	}
}

func TestCGREventAsDataProvider(t *testing.T) {
	ev := &CGREvent{
		APIOpts: map[string]interface{}{
			"testKey1": 13,
			"testKey2": "testString1",
		},
		Event: map[string]interface{}{
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
	eEv := map[string]interface{}{
		"FirstLevel2.SecondLevel2.Field2":        "Value2",
		"FirstLevel.SecondLevel.ThirdLevel.Fld1": "Val1",
		"FirstLevel2.Field3":                     "Value3",
		"FirstLevel2.Field5":                     "Value5",
		"FirstLevel2.Field6":                     "Value6",
		"Field4":                                 "Val4",
	}
	if cgrEv := NMAsCGREvent(nM, "cgrates.org",
		NestingSep, MapStorage{}); cgrEv.Tenant != "cgrates.org" ||
		cgrEv.Time == nil ||
		!reflect.DeepEqual(eEv, cgrEv.Event) {
		t.Errorf("expecting: %+v, \nreceived: %+v", ToJSON(eEv), ToJSON(cgrEv.Event))
	}
}

func TestCGREventRPCClone(t *testing.T) {
	att := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "234",
		clnb:   false,
		Time:   TimePointer(time.Date(2013, 12, 30, 15, 0, 1, 0, time.UTC)),
		Event: map[string]interface{}{
			"key": "value",
		},
		APIOpts: map[string]interface{}{
			"rand": "we",
		},
	}

	if rec, _ := att.RPCClone(); !reflect.DeepEqual(rec, att) {
		t.Errorf("expected %v and received %v ", ToJSON(att), ToJSON(rec))

	}

	att.SetCloneable(true)
	if att.clnb != true {
		t.Error("expected true")
	}
	rec, _ := att.RPCClone()
	att.SetCloneable(false)
	if !reflect.DeepEqual(rec, att) {
		t.Errorf("expected %v and received %v ", att, rec)

	}

}
