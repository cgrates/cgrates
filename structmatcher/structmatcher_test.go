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
package structmatcher

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func toJSON(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", " ")
	return string(b)
}

func TestStructMatcher(t *testing.T) {
	cl := &StructMatcher{}
	err := cl.Parse(`{"*or":[{"test":1},{"field":{"*gt":1}},{"best":"coco"}]}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}

	err = cl.Parse(`{"*has":["NAT","RET","EUR"]}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	err = cl.Parse(`{"Field":7, "Other":true}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	err = cl.Parse(``)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
}

func TestStructMatcherKeyValue(t *testing.T) {
	o := struct {
		Test    string
		Field   float64
		Other   bool
		ExpDate time.Time
	}{
		Test:    "test",
		Field:   6.0,
		Other:   true,
		ExpDate: time.Date(2016, 1, 19, 20, 47, 0, 0, time.UTC),
	}
	cl := &StructMatcher{}
	err := cl.Parse(`{"Test":"test"}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":6}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Other":true}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":6, "Other":true}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":7, "Other":true}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":6, "Other":false}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Other":true, "Field":{"*gt":5}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"*not":[{"Other":true, "Field":{"*gt":5}}]}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Other":true, "Field":{"*gt":7}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(``)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"ExpDate":{"*exp":true}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"ExpDate":{"*exp":false}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"*and":[{"Field":{"*gte":50}},{"Test":{"*eq":"test"}}]}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"WrongFieldName":{"*eq":1}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); check || err == nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
}

func TestStructMatcherKeyValuePointer(t *testing.T) {
	o := &struct {
		Test  string
		Field float64
		Other bool
	}{
		Test:  "test",
		Field: 6.0,
		Other: true,
	}
	cl := &StructMatcher{}
	err := cl.Parse(`{"Test":"test"}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":6}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Other":true}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
}

func TestStructMatcherOperatorValue(t *testing.T) {
	root := &operatorValue{operator: CondGT, value: 3.4}
	if check, err := root.checkStruct(3.5); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(root))
	}
	root = &operatorValue{operator: CondEQ, value: 3.4}
	if check, err := root.checkStruct(3.5); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(root))
	}
	root = &operatorValue{operator: CondEQ, value: 3.4}
	if check, err := root.checkStruct(3.4); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(root))
	}
	root = &operatorValue{operator: CondEQ, value: "zinc"}
	if check, err := root.checkStruct("zinc"); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(root))
	}
	root = &operatorValue{operator: CondHAS, value: []interface{}{"NAT", "RET", "EUR"}}
	if check, err := root.checkStruct(StringMap{"WOR": true, "EUR": true, "NAT": true, "RET": true, "ROM": true}); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", !check, err, toJSON(root))
	}
}

func TestStructMatcherKeyStruct(t *testing.T) {
	o := struct {
		Test  string
		Field float64
		Other bool
	}{
		Test:  "test",
		Field: 6.0,
		Other: true,
	}
	cl := &StructMatcher{}
	err := cl.Parse(`{"Field":{"*gt": 5}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Test":{"*gt": 5}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); check || !strings.HasPrefix(err.Error(), "INVALID_ARGUMENT") {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":{"*gte": 6}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":{"*lt": 7}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":{"*lte": 6}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":{"*eq": 6}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Test":{"*eq": "test"}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
}

func TestStructMatcherKeyStructPointer(t *testing.T) {
	o := &struct {
		Test  string
		Field float64
		Other bool
	}{
		Test:  "test",
		Field: 6.0,
		Other: true,
	}
	cl := &StructMatcher{}
	err := cl.Parse(`{"Field":{"*gt": 5}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Test":{"*gt": 5}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); check || !strings.HasPrefix(err.Error(), "INVALID_ARGUMENT") {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":{"*gte": 6}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":{"*lt": 7}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":{"*lte": 6}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":{"*eq": 6}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Test":{"*eq": "test"}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
}

func TestStructMatcherOperatorSlice(t *testing.T) {
	o := &struct {
		Test  string
		Field float64
		Other bool
	}{
		Test:  "test",
		Field: 6.0,
		Other: true,
	}
	cl := &StructMatcher{}
	err := cl.Parse(`{"*or":[{"Test":"test"},{"Field":{"*gt":5}},{"Other":true}]}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"*or":[{"Test":"test"},{"Field":{"*gt":7}},{"Other":false}]}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"*not":[{"*or":[{"Test":"test"},{"Field":{"*gt":7}},{"Other":false}]}]}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"*and":[{"Test":"test"},{"Field":{"*gt":5}},{"Other":true}]}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"*not":[{"*and":[{"Test":"test"},{"Field":{"*gt":5}},{"Other":true}]}]}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"*and":[{"Test":"test"},{"Field":{"*gt":7}},{"Other":false}]}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
}

func TestStructMatcherMixed(t *testing.T) {
	o := &struct {
		Test       string
		Field      float64
		Categories StringMap
		Other      bool
	}{
		Test:       "test",
		Field:      6.0,
		Categories: StringMap{"call": true, "data": true, "voice": true},
		Other:      true,
	}
	cl := &StructMatcher{}
	err := cl.Parse(`{"*and":[{"Test":"test"},{"Field":{"*gt":5}},{"Other":true},{"Categories":{"*has":["data", "call"]}}]}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
}

func TestStructMatcherBalanceType(t *testing.T) {
	type Balance struct {
		Value float64
	}

	o := &struct {
		BalanceType string
		Balance
	}{
		BalanceType: "*monetary",
		Balance:     Balance{Value: 10},
	}
	cl := &StructMatcher{}
	err := cl.Parse(`{"BalanceType":"*monetary","Value":10}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", !check, err, toJSON(cl.rootElement))
	}
}

func TestStructMatcherRSR(t *testing.T) {
	o := &struct {
		BalanceType string
	}{
		BalanceType: "*monetary",
	}
	cl := &StructMatcher{}
	err := cl.Parse(`{"BalanceType":{"*rsr":"^*mon"}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", !check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"BalanceType":{"*rsr":"^*min"}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", toJSON(cl.rootElement), err)
	}
	if check, err := cl.Match(o); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", !check, err, toJSON(cl.rootElement))
	}
}
