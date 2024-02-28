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
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
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
		Event: map[string]any{
			utils.Usage: 20 * time.Second,
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
		Event: map[string]any{
			utils.Usage: 20 * time.Second,
			"test1":     1,
			"test2":     2,
			"test3":     3,
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
	if err == nil || err.Error() != utils.NewErrMandatoryIeMissing("test4").Error() {
		t.Errorf("Expected %s, received %s", utils.NewErrMandatoryIeMissing("test4"), err)
	}
}

func TestCGREventFielAsString(t *testing.T) {
	//empty check
	cgrEvent := new(CGREvent)
	fldname := "test"
	_, err := cgrEvent.FieldAsString(fldname)
	if err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %s, received %s", utils.ErrNotFound, err)
	}
	//normal check
	cgrEvent = &CGREvent{
		Event: map[string]any{
			utils.Usage: 20 * time.Second,
			"test1":     1,
			"test2":     2,
			"test3":     3,
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
			utils.Usage: 20 * time.Second,
		},
	}
	seErr := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "supplierEvent1",
		Event:  make(map[string]any),
	}
	answ, err := se.FieldAsDuration(utils.Usage)
	if err != nil {
		t.Error(err)
	}
	if answ != se.Event[utils.Usage] {
		t.Errorf("Expecting: %+v, received: %+v", se.Event[utils.Usage], answ)
	}
	_, err = seErr.FieldAsDuration(utils.Usage)
	if err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestCGREventFieldAsTime(t *testing.T) {
	se := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "supplierEvent1",
		Event: map[string]any{
			utils.AnswerTime: time.Now(),
		},
	}
	seErr := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "supplierEvent1",
		Event:  make(map[string]any),
	}
	answ, err := se.FieldAsTime(utils.AnswerTime, "UTC")
	if err != nil {
		t.Error(err)
	}
	if answ != se.Event[utils.AnswerTime] {
		t.Errorf("Expecting: %+v, received: %+v", se.Event[utils.AnswerTime], answ)
	}
	_, err = seErr.FieldAsTime(utils.AnswerTime, "CET")
	if err != utils.ErrNotFound {
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

func TestCGREventFieldAsFloat64(t *testing.T) {
	se := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "supplierEvent1",
		Event: map[string]any{
			utils.AnswerTime:   time.Now(),
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

	if _, err := se.FieldAsFloat64(utils.AnswerTime); err == nil || !strings.HasPrefix(err.Error(), "cannot convert field") {
		t.Errorf("Unexpected error : %+v", err)
	}
	if _, err := se.FieldAsFloat64(utils.AccountField); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %s, received %s", utils.ErrNotFound, err)
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
		Event: map[string]any{
			utils.AnswerTime:   time.Now(),
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
	if cloned.Time == ev.Time {
		t.Errorf("Expecting: different pointer but received: %+v", cloned.Time)
	}
}

func TestCGREventconsumeRoutePaginator(t *testing.T) {
	//empty check
	var opts map[string]any
	rcv, err := GetRoutePaginatorFromOpts(opts)
	if err != nil {
		t.Error(err)
	}
	var eOut utils.Paginator
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
	opts = map[string]any{
		utils.OptsRoutesLimit:  18,
		utils.OptsRoutesOffset: 20,
	}

	eOut = utils.Paginator{
		Limit:  utils.IntPointer(18),
		Offset: utils.IntPointer(20),
	}
	rcv, err = GetRoutePaginatorFromOpts(opts)
	if err != nil {
		t.Error(err)
	}
	//check if *rouLimit and *rouOffset was deleted
	if _, has := opts[utils.OptsRoutesLimit]; has {
		t.Errorf("*rouLimit wasn't deleted")
	} else if _, has := opts[utils.OptsRoutesOffset]; has {
		t.Errorf("*rouOffset wasn't deleted")
	}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting:  %+v, received: %+v", eOut, rcv)
	}
	//check without *rouLimit, but with *rouOffset
	opts = map[string]any{
		utils.OptsRoutesOffset: 20,
	}

	eOut = utils.Paginator{
		Offset: utils.IntPointer(20),
	}
	rcv, err = GetRoutePaginatorFromOpts(opts)
	if err != nil {
		t.Error(err)
	}
	//check if *rouLimit and *rouOffset was deleted
	if _, has := opts[utils.OptsRoutesLimit]; has {
		t.Errorf("*rouLimit wasn't deleted")
	} else if _, has := opts[utils.OptsRoutesOffset]; has {
		t.Errorf("*rouOffset wasn't deleted")
	}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting:  %+v, received: %+v", eOut, rcv)
	}
	//check with notAnInt at *rouLimit
	opts = map[string]any{
		utils.OptsRoutesLimit: "Not an int",
	}
	eOut = utils.Paginator{}
	rcv, err = GetRoutePaginatorFromOpts(opts)
	if err == nil {
		t.Error("Expected error")
	}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting:  %+v, received: %+v", eOut, rcv)
	}
	//check with notAnInt at and *rouOffset
	opts = map[string]any{
		utils.OptsRoutesOffset: "Not an int",
	}
	eOut = utils.Paginator{}
	rcv, err = GetRoutePaginatorFromOpts(opts)
	if err == nil {
		t.Error("Expected error")
	}
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting:  %+v, received: %+v", eOut, rcv)
	}
}

func TestCGREventconsumeRoutePaginatorCase1(t *testing.T) {
	opts := map[string]any{}
	if _, err := GetRoutePaginatorFromOpts(opts); err != nil {
		t.Error(err)
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
	errExpect := utils.ErrNotFound
	if _, err = ev.OptAsInt64("nonExistingKey"); err == nil || err != errExpect {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", errExpect, err)
	}
}

func TestCGREventFieldAsInt64(t *testing.T) {
	se := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "supplierEvent1",
		Event: map[string]any{
			utils.AnswerTime:   time.Now(),
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

	if _, err := se.FieldAsInt64(utils.AnswerTime); err == nil || !strings.HasPrefix(err.Error(), "cannot convert field") {
		t.Errorf("Unexpected error : %+v", err)
	}
	if _, err := se.FieldAsInt64(utils.AccountField); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %s, received %s", utils.ErrNotFound, err)
	}
	// }
}

func TestCGREventOptAsStringEmpty(t *testing.T) {
	ev := &CGREvent{}

	expstr := ""
	experr := utils.ErrNotFound
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
		APIOpts: map[string]any{
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
	experr := utils.ErrNotFound
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
		APIOpts: map[string]any{
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
		APIOpts: map[string]any{
			"testKey1": 13,
			"testKey2": "testString1",
		},
		Event: map[string]any{
			"testKey1": 30,
			"testKey2": "testString2",
		},
	}

	expected := utils.MapStorage{
		utils.MetaOpts: ev.APIOpts,
		utils.MetaReq:  ev.Event,
	}

	received := ev.AsDataProvider()

	if !reflect.DeepEqual(expected, received) {
		t.Errorf("\nExpected: <%+v>, \nReceived: <%+v>", expected, received)
	}
}

func TestNMAsCGREvent(t *testing.T) {
	if cgrEv := NMAsCGREvent(nil, "cgrates.org",
		utils.NestingSep, nil); cgrEv != nil {
		t.Errorf("expecting: %+v, \nreceived: %+v", utils.ToJSON(nil), utils.ToJSON(cgrEv.Event))
	}

	nM := utils.NewOrderedNavigableMap()
	if cgrEv := NMAsCGREvent(nM, "cgrates.org",
		utils.NestingSep, nil); cgrEv != nil {
		t.Errorf("expecting: %+v, \nreceived: %+v", utils.ToJSON(nil), utils.ToJSON(cgrEv.Event))
	}

	path := []string{"FirstLevel", "SecondLevel", "ThirdLevel", "Fld1"}
	if err := nM.SetAsSlice(&utils.FullPath{Path: strings.Join(path, utils.NestingSep), PathSlice: path}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{
			Data: "Val1",
		}}}); err != nil {
		t.Error(err)
	}

	path = []string{"FirstLevel2", "SecondLevel2", "Field2"}
	if err := nM.SetAsSlice(&utils.FullPath{Path: strings.Join(path, utils.NestingSep), PathSlice: path}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{
			Data:        "attrVal1",
			AttributeID: "attribute1",
		}},
		{Type: utils.NMDataType, Value: &utils.DataLeaf{
			Data: "Value2",
		}}}); err != nil {
		t.Error(err)
	}

	path = []string{"FirstLevel2", "Field3"}
	if err := nM.SetAsSlice(&utils.FullPath{Path: strings.Join(path, utils.NestingSep), PathSlice: path}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{
			Data: "Value3",
		}}}); err != nil {
		t.Error(err)
	}

	path = []string{"FirstLevel2", "Field5"}
	if err := nM.SetAsSlice(&utils.FullPath{Path: strings.Join(path, utils.NestingSep), PathSlice: path}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{
			Data: "Value5",
		}},
		{Type: utils.NMDataType, Value: &utils.DataLeaf{
			Data:        "attrVal5",
			AttributeID: "attribute5",
		}}}); err != nil {
		t.Error(err)
	}

	path = []string{"FirstLevel2", "Field6"}
	if err := nM.SetAsSlice(&utils.FullPath{Path: strings.Join(path, utils.NestingSep), PathSlice: path}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{
			Data:      "Value6",
			NewBranch: true,
		}},
		{Type: utils.NMDataType, Value: &utils.DataLeaf{
			Data:        "attrVal6",
			AttributeID: "attribute6",
		}}}); err != nil {
		t.Error(err)
	}

	path = []string{"Field4"}
	if err := nM.SetAsSlice(&utils.FullPath{Path: strings.Join(path, utils.NestingSep), PathSlice: path}, []*utils.DataNode{
		{Type: utils.NMDataType, Value: &utils.DataLeaf{
			Data: "Val4",
		}},
		{Type: utils.NMDataType, Value: &utils.DataLeaf{
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
		utils.NestingSep, utils.MapStorage{}); cgrEv.Tenant != "cgrates.org" ||
		cgrEv.Time == nil ||
		!reflect.DeepEqual(eEv, cgrEv.Event) {
		t.Errorf("expecting: %+v, \nreceived: %+v", utils.ToJSON(eEv), utils.ToJSON(cgrEv.Event))
	}
}

func TestCGREventRPCClone(t *testing.T) {
	att := &CGREvent{
		Tenant: "cgrates.org",
		ID:     "234",
		clnb:   false,
		Time:   utils.TimePointer(time.Date(2013, 12, 30, 15, 0, 1, 0, time.UTC)),
		Event: map[string]any{
			"key": "value",
		},
		APIOpts: map[string]any{
			"rand": "we",
		},
	}

	if rec, _ := att.RPCClone(); !reflect.DeepEqual(rec, att) {
		t.Errorf("expected %v and received %v ", utils.ToJSON(att), utils.ToJSON(rec))

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

func TestOptionsGetFloat64Opts(t *testing.T) {

	// option key not populated, retrieve default value
	ev := &CGREvent{
		APIOpts: make(map[string]any),
	}
	if rcv, err := GetFloat64Opts(ev, 1.2, "optionName"); err != nil {
		t.Error(err)
	} else if rcv != 1.2 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 1.2, rcv)
	}

	// option key populated with valid input, get its value
	ev.APIOpts["optionName"] = 0.11
	if rcv, err := GetFloat64Opts(ev, 1.2, "optionName"); err != nil {
		t.Error(err)
	} else if rcv != 0.11 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 0.11, rcv)
	}

	// option key populated with invalid input, receive error
	expectedErr := `strconv.ParseFloat: parsing "invalid": invalid syntax`
	ev.APIOpts["optionName"] = "invalid"
	if _, err := GetFloat64Opts(ev, 1.2, "optionName"); err == nil ||
		err.Error() != expectedErr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedErr, err)
	}
}

func TestOptionsGetDurationOpts(t *testing.T) {

	// option key not populated, retrieve default value
	ev := &CGREvent{
		APIOpts: make(map[string]any),
	}
	if rcv, err := GetDurationOpts(ev, time.Minute, "optionName"); err != nil {
		t.Error(err)
	} else if rcv != time.Minute {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", time.Minute, rcv)
	}

	// option key populated with valid input, get its value
	ev.APIOpts["optionName"] = 2 * time.Second
	if rcv, err := GetDurationOpts(ev, time.Minute, "optionName"); err != nil {
		t.Error(err)
	} else if rcv != 2*time.Second {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 2*time.Second, rcv)
	}
	ev.APIOpts["optionName"] = 600
	if rcv, err := GetDurationOpts(ev, time.Minute, "optionName"); err != nil {
		t.Error(err)
	} else if rcv != 600 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 600, rcv)
	}
	ev.APIOpts["optionName"] = "2m0s"
	if rcv, err := GetDurationOpts(ev, time.Minute, "optionName"); err != nil {
		t.Error(err)
	} else if rcv != 2*time.Minute {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 2*time.Minute, rcv)
	}

	// option key populated with invalid input, receive error
	expectedErr := `time: invalid duration "invalid"`
	ev.APIOpts["optionName"] = "invalid"
	if _, err := GetDurationOpts(ev, time.Minute, "optionName"); err == nil ||
		err.Error() != expectedErr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedErr, err)
	}
}

func TestOptionsGetStringOpts(t *testing.T) {

	// option key not populated, retrieve default value
	ev := &CGREvent{
		APIOpts: make(map[string]any),
	}
	if rcv := GetStringOpts(ev, "default", "optionName"); rcv != "default" {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", "default", rcv)
	}

	// option key populated, get its value
	ev.APIOpts["optionName"] = "optionValue"
	if rcv := GetStringOpts(ev, "default", "optionName"); rcv != "optionValue" {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", "optionValue", rcv)
	}
	ev.APIOpts["optionName"] = false
	if rcv := GetStringOpts(ev, "default", "optionName"); rcv != "false" {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", "false", rcv)
	}
	ev.APIOpts["optionName"] = 5 * time.Minute
	if rcv := GetStringOpts(ev, "default", "optionName"); rcv != "5m0s" {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", "5m0s", rcv)
	}
	ev.APIOpts["optionName"] = 12.34
	if rcv := GetStringOpts(ev, "default", "optionName"); rcv != "12.34" {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", "12.34", rcv)
	}
}

func TestOptionsGetStringSliceOpts(t *testing.T) {

	// option key not populated, retrieve default value
	ev := &CGREvent{
		APIOpts: make(map[string]any),
	}
	defaultValue := []string{"default"}
	if rcv, err := GetStringSliceOpts(ev, defaultValue, "optionName"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, defaultValue) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", defaultValue, rcv)
	}

	// option key populated with valid input, get its value
	optValue := []string{"optValue"}
	ev.APIOpts["optionName"] = optValue
	if rcv, err := GetStringSliceOpts(ev, defaultValue, "optionName"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, optValue) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", optValue, rcv)
	}
	optValue2 := []string{"true", "false"}
	ev.APIOpts["optionName"] = []bool{true, false}
	if rcv, err := GetStringSliceOpts(ev, defaultValue, "optionName"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, optValue2) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", optValue2, rcv)
	}
	optValue3 := []string{"12.34", "3"}
	ev.APIOpts["optionName"] = []float64{12.34, 3}
	if rcv, err := GetStringSliceOpts(ev, defaultValue, "optionName"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, optValue3) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", optValue3, rcv)
	}

	// option key populated with invalid input, receive error
	expectedErr := `cannot convert field: string to []string`
	ev.APIOpts["optionName"] = "invalid"
	if _, err := GetStringSliceOpts(ev, defaultValue, "optionName"); err == nil ||
		err.Error() != expectedErr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedErr, err)
	}
}

func TestOptionsGetIntOpts(t *testing.T) {

	// option key not populated, retrieve default value
	ev := &CGREvent{
		APIOpts: make(map[string]any),
	}
	if rcv, err := GetIntOpts(ev, 5, "optionName"); err != nil {
		t.Error(err)
	} else if rcv != 5 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 5, rcv)
	}

	// option key populated with valid input, get its value
	ev.APIOpts["optionName"] = 12
	if rcv, err := GetIntOpts(ev, 5, "optionName"); err != nil {
		t.Error(err)
	} else if rcv != 12 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 12, rcv)
	}
	ev.APIOpts["optionName"] = 12.7
	if rcv, err := GetIntOpts(ev, 5, "optionName"); err != nil {
		t.Error(err)
	} else if rcv != 12 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 12, rcv)
	}

	// option key populated with invalid input, receive error
	expectedErr := `strconv.ParseInt: parsing "invalid": invalid syntax`
	ev.APIOpts["optionName"] = "invalid"
	if _, err := GetIntOpts(ev, 5, "optionName"); err == nil ||
		err.Error() != expectedErr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedErr, err)
	}
}

func TestOptionsGetBoolOpts(t *testing.T) {

	// option key not populated, retrieve default value
	ev := &CGREvent{
		APIOpts: make(map[string]any),
	}
	if rcv, err := GetBoolOpts(ev, false, "optionName"); err != nil {
		t.Error(err)
	} else if rcv != false {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", false, rcv)
	}

	// option key populated with valid input, get its value
	ev.APIOpts["optionName"] = true
	if rcv, err := GetBoolOpts(ev, false, "optionName"); err != nil {
		t.Error(err)
	} else if rcv != true {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", true, rcv)
	}
	ev.APIOpts["optionName"] = 5
	if rcv, err := GetBoolOpts(ev, false, "optionName"); err != nil {
		t.Error(err)
	} else if rcv != true {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", true, rcv)
	}
	ev.APIOpts["optionName"] = "true"
	if rcv, err := GetBoolOpts(ev, false, "optionName"); err != nil {
		t.Error(err)
	} else if rcv != true {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", true, rcv)
	}

	// option key populated with invalid input, receive error
	expectedErr := `strconv.ParseBool: parsing "invalid": invalid syntax`
	ev.APIOpts["optionName"] = "invalid"
	if _, err := GetBoolOpts(ev, false, "optionName"); err == nil ||
		err.Error() != expectedErr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedErr, err)
	}
}

func TestOptionsGetInterfaceOpts(t *testing.T) {

	// option key not populated, retrieve default value
	ev := &CGREvent{
		APIOpts: make(map[string]any),
	}
	if rcv := GetInterfaceOpts(ev, "default", "optionName"); rcv != "default" {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", "default", rcv)
	}

	// option key populated, get its value
	ev.APIOpts["optionName"] = 0.11
	if rcv := GetInterfaceOpts(ev, "default", "optionName"); rcv != 0.11 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 0.11, rcv)
	}
	ev.APIOpts["optionName"] = true
	if rcv := GetInterfaceOpts(ev, "default", "optionName"); rcv != true {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", true, rcv)
	}
	ev.APIOpts["optionName"] = 5 * time.Minute
	if rcv := GetInterfaceOpts(ev, "default", "optionName"); rcv != 5*time.Minute {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 5*time.Minute, rcv)
	}
}

func TestOptionsGetIntPointerOpts(t *testing.T) {
	defaultValue := utils.IntPointer(5)

	// option key not populated, retrieve default value
	ev := &CGREvent{
		APIOpts: make(map[string]any),
	}
	if rcv, err := GetIntPointerOpts(ev, defaultValue, "optionName"); err != nil {
		t.Error(err)
	} else if rcv == nil || rcv != defaultValue {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", defaultValue, rcv)
	}

	// option key populated with valid input, get its value
	ev.APIOpts["optionName"] = 12
	if rcv, err := GetIntPointerOpts(ev, defaultValue, "optionName"); err != nil {
		t.Error(err)
	} else if rcv == nil || *rcv != 12 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 12, rcv)
	}
	ev.APIOpts["optionName"] = 12.7
	if rcv, err := GetIntPointerOpts(ev, defaultValue, "optionName"); err != nil {
		t.Error(err)
	} else if rcv == nil || *rcv != 12 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 12, rcv)
	}

	// option key populated with invalid input, receive error
	expectedErr := `strconv.ParseInt: parsing "invalid": invalid syntax`
	ev.APIOpts["optionName"] = "invalid"
	if _, err := GetIntPointerOpts(ev, defaultValue, "optionName"); err == nil ||
		err.Error() != expectedErr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedErr, err)
	}
}

func TestOptionsGetDurationPointerOpts(t *testing.T) {
	defaultValue := utils.DurationPointer(time.Minute)
	// option key not populated, retrieve default value
	ev := &CGREvent{
		APIOpts: make(map[string]any),
	}
	if rcv, err := GetDurationPointerOpts(ev, defaultValue, "optionName"); err != nil {
		t.Error(err)
	} else if rcv == nil || rcv != defaultValue {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", defaultValue, rcv)
	}

	// option key populated with valid input, get its value
	ev.APIOpts["optionName"] = 2 * time.Second
	if rcv, err := GetDurationPointerOpts(ev, defaultValue, "optionName"); err != nil {
		t.Error(err)
	} else if rcv == nil || *rcv != 2*time.Second {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 2*time.Second, rcv)
	}
	ev.APIOpts["optionName"] = 600
	if rcv, err := GetDurationPointerOpts(ev, defaultValue, "optionName"); err != nil {
		t.Error(err)
	} else if rcv == nil || *rcv != 600 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 600, rcv)
	}
	ev.APIOpts["optionName"] = "2m0s"
	if rcv, err := GetDurationPointerOpts(ev, defaultValue, "optionName"); err != nil {
		t.Error(err)
	} else if rcv == nil || *rcv != 2*time.Minute {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 2*time.Minute, rcv)
	}

	// option key populated with invalid input, receive error
	expectedErr := `time: invalid duration "invalid"`
	ev.APIOpts["optionName"] = "invalid"
	if _, err := GetDurationPointerOpts(ev, defaultValue, "optionName"); err == nil ||
		err.Error() != expectedErr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedErr, err)
	}
}

func TestOptionsGetDecimalBigOpts(t *testing.T) {
	defaultValue := decimal.New(1234, 3)

	// option key not populated, retrieve default value
	ev := &CGREvent{
		APIOpts: make(map[string]any),
	}
	if rcv, err := GetDecimalBigOpts(ev, defaultValue, "optionName"); err != nil {
		t.Error(err)
	} else if rcv.Cmp(defaultValue) != 0 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", defaultValue, rcv)
	}

	// option key populated with valid input, get its value
	optValue := decimal.New(15, 1)
	ev.APIOpts["optionName"] = 1.5
	if rcv, err := GetDecimalBigOpts(ev, defaultValue, "optionName"); err != nil {
		t.Error(err)
	} else if rcv.Cmp(optValue) != 0 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", optValue, rcv)
	}
	ev.APIOpts["optionName"] = "1.5"
	if rcv, err := GetDecimalBigOpts(ev, defaultValue, "optionName"); err != nil {
		t.Error(err)
	} else if rcv.Cmp(optValue) != 0 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", optValue, rcv)
	}

	// option key populated with invalid input, receive error
	expectedErr := `can't convert <invalid> to decimal`
	ev.APIOpts["optionName"] = "invalid"
	if _, err := GetDecimalBigOpts(ev, defaultValue, "optionName"); err == nil ||
		err.Error() != expectedErr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedErr, err)
	}
}

func TestCGREventUnmarshalJSON(t *testing.T) {
	var testEC = &EventCost{
		CGRID:     "164b0422fdc6a5117031b427439482c6a4f90e41",
		RunID:     utils.MetaDefault,
		StartTime: time.Date(2017, 1, 9, 16, 18, 21, 0, time.UTC),
		AccountSummary: &AccountSummary{
			Tenant: "cgrates.org",
			ID:     "dan",
			BalanceSummaries: []*BalanceSummary{
				{
					UUID:     "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
					ID:       "BALANCE_1",
					Type:     utils.MetaMonetary,
					Value:    50,
					Initial:  60,
					Disabled: false,
				},
			},
			AllowNegative: false,
			Disabled:      false,
		},
	}
	cdBytes, err := json.Marshal(testEC)
	if err != nil {
		t.Fatal(err)
	}

	cgrEv := CGREvent{
		Tenant: "cgrates.org",
		ID:     "ev1",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}

	t.Run("UnmarshalFromMap", func(t *testing.T) {
		var cdMap map[string]any
		if err = json.Unmarshal(cdBytes, &cdMap); err != nil {
			t.Fatal(err)
		}

		cgrEv.Event[utils.CostDetails] = cdMap

		cgrEvBytes, err := json.Marshal(cgrEv)
		if err != nil {
			t.Fatal(err)
		}

		var rcvCGREv CGREvent
		if err = json.Unmarshal(cgrEvBytes, &rcvCGREv); err != nil {
			t.Fatal(err)
		}

		expectedType := "*engine.EventCost"
		if cdType := fmt.Sprintf("%T", rcvCGREv.Event[utils.CostDetails]); cdType != expectedType {
			t.Fatalf("expected type to be %v, received %v", expectedType, cdType)
		}

		cgrEv.Event[utils.CostDetails] = testEC
		if !reflect.DeepEqual(rcvCGREv, cgrEv) {
			t.Errorf("expected: %v,\nreceived: %v",
				utils.ToJSON(cgrEv), utils.ToJSON(rcvCGREv))
		}
	})
	t.Run("UnmarshalFromString", func(t *testing.T) {
		cdStringBytes, err := json.Marshal(string(cdBytes))
		if err != nil {
			t.Fatal(err)
		}
		var cdString string
		if err = json.Unmarshal(cdStringBytes, &cdString); err != nil {
			t.Fatal(err)
		}

		cgrEv.Event[utils.CostDetails] = cdString

		cgrEvBytes, err := json.Marshal(cgrEv)
		if err != nil {
			t.Fatal(err)
		}

		var rcvCGREv CGREvent
		if err = json.Unmarshal(cgrEvBytes, &rcvCGREv); err != nil {
			t.Fatal(err)
		}

		expectedType := "*engine.EventCost"
		if cdType := fmt.Sprintf("%T", rcvCGREv.Event[utils.CostDetails]); cdType != expectedType {
			t.Fatalf("expected type to be %v, received %v", expectedType, cdType)
		}

		cgrEv.Event[utils.CostDetails] = testEC
		if !reflect.DeepEqual(rcvCGREv, cgrEv) {
			t.Errorf("expected: %v,\nreceived: %v",
				utils.ToJSON(cgrEv), utils.ToJSON(rcvCGREv))
		}
	})
}
